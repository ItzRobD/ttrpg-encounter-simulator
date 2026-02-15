import {computed, inject, Injectable, signal, WritableSignal} from '@angular/core';
import { ActorState, EventType, SimulationLog, Actor, ActorSummary, Condition, ActorStateSnapshot } from '../models';
import {CombatantService} from './combatant.service';
import {SimulationStateService} from './simulation-state.service';

@Injectable({
  providedIn: 'root',
})
export class TimelineService {
  private readonly combatantService = inject(CombatantService);
  private readonly stateService = inject(SimulationStateService);

  readonly selectedSimulationLog = this.stateService.selectedSimulationLog;

  readonly activeEvent = computed(() => {
    const log = this.selectedSimulationLog();
    const index = this._currentTimelineIndex();
    const events = log?.events;
    if (!events || index < 0 || index >= events.length) {
      return null;
    }
    return events[index];
  });

  setSelectedSimulationLog(log: SimulationLog | null): void {
    this.stateService.setSelectedSimulationLog(log);
    this._scrubberIndex.set(0);
    this._currentTimelineIndex.set(0);
  }

  private readonly _currentTimelineIndex: WritableSignal<number> = signal(0);
  readonly currentEventIndex = this._currentTimelineIndex.asReadonly();

  /**
   * The list of events that the user can scrub through (Rounds, Turns, Choices).
   */
  readonly scrubbableEvents = computed(() => {
    const log = this.selectedSimulationLog();
    const events = log?.events;
    if (!events) return [];
    return events.filter(e =>
      e.type === EventType.Round ||
      e.type === EventType.Turn ||
      e.type === EventType.Choice ||
      e.type === EventType.Unconscious ||
      e.type === EventType.Death ||
      e.type === EventType.HPModified
    );
  });

  /**
   * The index within the scrubbableEvents list.
   */
  private readonly _scrubberIndex = signal(0);

  get currentTimelineIndex(): number {
    return this._scrubberIndex();
  }

  set currentTimelineIndex(index: number) {
    this._scrubberIndex.set(index);

    // Sync the internal event index
    const scrubbable = this.scrubbableEvents();
    const event = scrubbable[index];
    if (event) {
      const log = this.selectedSimulationLog();
      const events = log?.events;
      if (events) {
        // If it's a round, find the first event of that round
        if (event.type === EventType.Round) {
            const firstInRound = events.findIndex(e => e.round === event.round);
            if (firstInRound !== -1) {
                this._currentTimelineIndex.set(firstInRound);
                return;
            }
        }

        const fullIndex = events.findIndex(e => e.id === event.id);
        if (fullIndex !== -1) {
          this._currentTimelineIndex.set(fullIndex);
        }
      }
    }
  }

  incrementTimelineIndex(): void {
    if (this._scrubberIndex() < this.scrubbableEvents().length - 1) {
      this.currentTimelineIndex = this._scrubberIndex() + 1;
    }
  }

  decrementTimelineIndex(): void {
    if (this._scrubberIndex() > 0) {
      this.currentTimelineIndex = this._scrubberIndex() - 1;
    }
  }

  readonly projectedState = computed(() => {
    const log = this.selectedSimulationLog();
    const simResult = this.stateService.simulationResult();

    if (!log) {
      return new Map<number, ActorState>();
    }

    const events = log.events;
    if (!events) {
      return new Map<number, ActorState>();
    }
    const index = this._currentTimelineIndex();

    const stateMap = new Map<number, ActorState>();

    console.log(`[TimelineService] computing projectedState at index ${index}`);

    // Map backend instanceId to frontend instanceId
    const backendToFrontendIdMap = new Map<number, number>();
    const initialStates = log.actorInitialStates;
    const logActorConfigs = log.actorConfigs || [];

    if (initialStates) {
      console.log(`[TimelineService] Raw actorInitialStates keys:`, Object.keys(initialStates));
      console.log(`[TimelineService] Initializing states from actorInitialStates`, initialStates);
      Object.entries(initialStates).forEach(([id, state]: [string, ActorStateSnapshot]) => {
        const instanceId = Number(id);
        if (!isNaN(instanceId)) {
          // Robust mapping for HP values which might be camelCase or snake_case
          const currentHp = state.currentHp;
          const maxHp = (state as any).maxHp ?? (state as any).max_hp ?? (state as any).maxHP;
          const tempHp = state.tempHp;

          // Attempt to find original frontend instanceId by name
          const actorConfig = (logActorConfigs as Actor[]).find((c: Actor) => c.instanceId === instanceId);
          let targetInstanceId = instanceId;

          if (actorConfig) {
            const combatant = this.combatantService.combatants().find(c => c.name === actorConfig.name);
            if (combatant) {
              targetInstanceId = combatant.instanceId;
              backendToFrontendIdMap.set(instanceId, targetInstanceId);
              console.log(`[TimelineService] Mapping backend instanceId ${instanceId} to frontend instanceId ${targetInstanceId} for actor ${actorConfig.name}`);
            }
          }

          stateMap.set(targetInstanceId, {
            ...state,
            // Ensure properties match ActorState interface
            currentHp: Number(currentHp ?? 0),
            maxHp: Number(maxHp ?? 0),
            tempHp: Number(tempHp ?? 0),
            hitDie: (state as any).hitDie ?? 10,
            conditions: (state as any).conditions || {},
            deathSaves: (state as any).deathSaves || { successes: 0, failures: 0 },
            resistances: (state as any).resistances || {},
            isStable: (state as any).isStable ?? (state as any).stable ?? true,
            isDead: (state as any).isDead ?? (state.healthState === 'dead'),
            initiative: (state as any).initiative ?? 0,
            isProjected: true
          } as unknown as ActorState);
          console.log(`[TimelineService] stateMap.set(${targetInstanceId}, HP: ${stateMap.get(targetInstanceId)?.currentHp})`, stateMap.get(targetInstanceId));
        }
      });
    }

    // 2. Replay events up to the current index
    console.log(`[TimelineService] Replaying up to index ${index}. Target Event ID: ${events[index]?.id}`);

    for (let i = 0; i <= index && i < events.length; i++) {
      const event = events[i];
      const data = event.data || {};

      // If the event has actorStates (snapshot), use it to update the stateMap
      if (data.actorStates) {
        Object.entries(data.actorStates).forEach(([id, state]: [string, ActorStateSnapshot]) => {
          const backendId = Number(id);
          const targetInstanceId = backendToFrontendIdMap.get(backendId) || backendId;
          const currentState = stateMap.get(targetInstanceId);

          if (currentState) {
            // Map snapshot state to ActorState interface
            const currentHp = state.currentHp;
            const tempHp = state.tempHp;

            stateMap.set(targetInstanceId, {
              ...currentState,
              currentHp: Number(currentHp ?? currentState.currentHp),
              tempHp: Number(tempHp ?? currentState.tempHp),
              conditions: (state as any).conditions || currentState.conditions,
              isDead: state.healthState === 'dead',
              isStable: (state as any).isStable ?? (state as any).stable ?? currentState.isStable
            });
          }
        });

        // If we have a snapshot, we might be able to skip detailed processing for this event
        // unless it's an event that changes something NOT in the snapshot (unlikely currently)
        if (event.type !== EventType.HPModified && event.type !== EventType.Initiative) {
            continue;
        }
      }

      switch (event.type) {
        case EventType.IntermissionHealing: {
            if (data.healing) {
                Object.entries(data.healing).forEach(([id, value]) => {
                    const backendId = Number(id);
                    const targetInstanceId = backendToFrontendIdMap.get(backendId) || backendId;
                    const currentState = stateMap.get(targetInstanceId);
                    if (currentState) {
                        stateMap.set(targetInstanceId, {
                            ...currentState,
                            currentHp: Math.min(currentState.maxHp, currentState.currentHp + Number(value))
                        });
                    }
                });
            }
            break;
        }
        case EventType.HPModified: {
          const rawTargetId = data.target?.instanceId || event.actor?.instanceId || data.targetId;
          const targetId = (rawTargetId !== undefined) ? (backendToFrontendIdMap.get(Number(rawTargetId)) || Number(rawTargetId)) : undefined;

          // Try to get newHp from multiple possible locations
          const newHp = (data.result?.newHp !== undefined && data.result?.newHp !== null) ? data.result.newHp : data.finalHp;

          const newTempHp = (data.result?.newTempHp !== undefined && data.result?.newTempHp !== null) ? data.result.newTempHp : data.finalTempHp;

          const modificationValue = (data.result?.modificationValue !== undefined) ? data.result.modificationValue :
                                    (data.value !== undefined) ? data.value : 0;

          console.log(`[TimelineService] HPModified event detail:`, {
            id: event.id,
            targetId,
            newHp,
            newTempHp,
            modificationValue,
            dataResult: data.result,
            data
          });

          if (targetId !== undefined) {
            const instanceId = Number(targetId);
            const currentState = stateMap.get(instanceId);
            if (currentState) {
              const maxHp = currentState.maxHp || 1;
              const update: Partial<ActorState> = {};

              if (newHp !== undefined && newHp !== null) {
                update.currentHp = Math.max(0, Math.min(Number(newHp), maxHp));
                console.log(`[TimelineService] HPModified: Actor ${instanceId}, Applied New HP: ${update.currentHp}/${maxHp}`);

                // If HP > 0, they might not be unconscious/dead anymore
                if (update.currentHp > 0) {
                    update.isDead = false;
                    update.isStable = true;
                    if (currentState.conditions) {
                        update.conditions = { ...currentState.conditions, [Condition.Unconscious]: false };
                    }
                }
              } else if (modificationValue !== 0) {
                // If newHp is missing, but modificationValue is present, calculate it
                const isHeal = data.result?.didHealHp ?? (Number(modificationValue) < 0);
                const change = Math.abs(Number(modificationValue));
                const oldHp = currentState.currentHp;
                update.currentHp = isHeal
                  ? Math.min(maxHp, oldHp + change)
                  : Math.max(0, oldHp - change);

                console.log(`[TimelineService] HPModified: Actor ${instanceId}, Calculated New HP: ${update.currentHp}/${maxHp} (from ${oldHp} ${isHeal ? '+' : '-'} ${change})`);

                if (update.currentHp > 0) {
                    update.isDead = false;
                    update.isStable = true;
                    if (currentState.conditions) {
                        update.conditions = { ...currentState.conditions, [Condition.Unconscious]: false };
                    }
                }
              }

              if (newTempHp !== undefined && newTempHp !== null) {
                update.tempHp = Math.max(0, Number(newTempHp));
              }

              const newState = {
                ...currentState,
                ...update
              };
              stateMap.set(instanceId, newState);
            } else {
              console.warn(`[TimelineService] HPModified: Actor ${instanceId} not found in stateMap. Current stateMap keys:`, Array.from(stateMap.keys()));
            }
          }
          break;
        }
        case EventType.Initiative: {
          const rawActorId = data.actor?.instanceId || event.actor?.instanceId || data.actorId;
          const actorId = (rawActorId !== undefined) ? (backendToFrontendIdMap.get(Number(rawActorId)) || Number(rawActorId)) : undefined;
          const initiative = data.roll?.total ?? data.value;

          console.log(`[TimelineService] Initiative event:`, { id: event.id, actorId, initiative, data });

          if (actorId !== undefined && initiative !== undefined) {
            const instanceId = Number(actorId);
            const currentState = stateMap.get(instanceId);
            if (currentState) {
              const newState = {
                ...currentState,
                initiative: Number(initiative)
              };
              console.log(`[TimelineService] Updating Actor ${instanceId} state (Initiative)`, newState);
              stateMap.set(instanceId, newState);
            }
          }
          break;
        }
        case EventType.Unconscious: {
          const rawTargetId = data.target?.instanceId || event.actor?.instanceId || data.targetId;
          const targetId = (rawTargetId !== undefined) ? (backendToFrontendIdMap.get(Number(rawTargetId)) || Number(rawTargetId)) : undefined;
          if (targetId !== undefined) {
            const instanceId = Number(targetId);
            const currentState = stateMap.get(instanceId);
            if (currentState) {
              const newState = {
                ...currentState,
                isDead: false,
                isStable: false,
                currentHp: (data.finalHp !== undefined && data.finalHp !== null) ? Number(data.finalHp) : currentState.currentHp,
                conditions: {
                  ...currentState.conditions,
                  [Condition.Unconscious]: true
                }
              };
              stateMap.set(instanceId, newState);
            }
          }
          break;
        }
        case EventType.Death: {
          const rawTargetId = data.target?.instanceId || event.actor?.instanceId || data.targetId;
          const targetId = (rawTargetId !== undefined) ? (backendToFrontendIdMap.get(Number(rawTargetId)) || Number(rawTargetId)) : undefined;
          if (targetId !== undefined) {
            const instanceId = Number(targetId);
            const currentState = stateMap.get(instanceId);
            if (currentState) {
              const newState = {
                ...currentState,
                isDead: true,
                isStable: false,
                currentHp: (data.finalHp !== undefined && data.finalHp !== null) ? Number(data.finalHp) : 0,
                conditions: {
                  ...currentState.conditions,
                  [Condition.Unconscious]: true
                }
              };
              stateMap.set(instanceId, newState);
            }
          }
          break;
        }
        case EventType.Victory: {
          // Stable status might change at victory or end of combat if we care
          break;
        }
      }
    }

    // 3. Log final state map
    console.log(`[TimelineService] Final stateMap keys:`, Array.from(stateMap.keys()));
    console.log(`[TimelineService] Final stateMap size: ${stateMap.size}. Summary:`,
      Array.from(stateMap.entries()).map(([id, state]) => {
        // Try to find name in combatantService since IDs are now mapped to frontend IDs
        const name = this.combatantService.combatants().find(c => c.instanceId === id)?.name || 'Unknown';
        return `${name} (ID ${id}): HP ${state.currentHp}/${state.maxHp}`;
      })
    );
    return stateMap;
  });
}
