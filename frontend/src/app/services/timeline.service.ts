import {computed, inject, Injectable, signal, WritableSignal} from '@angular/core';
import { ActorState, EventType, EventData, SimulationLog, Actor, ActorSummary, Condition, ActorStateSnapshot, ActorInitialStateSnapshot, isCharacter } from '../models';
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
      e.type === EventType.CombatStart ||
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
    console.log(`[TimelineService] Recomputing projectedState at index ${index}`);

    const stateMap = new Map<number, ActorState>();

    // Map backend instanceId to frontend instanceId
    const backendToFrontendIdMap = new Map<number, number>();
    const initialStates = log.actorInitialStates;
    const logActorConfigs = log.actorConfigs || [];

    if (initialStates) {
      // console.log('[TimelineService] Initializing stateMap from actorInitialStates:', initialStates);
      Object.entries(initialStates).forEach(([id, state]: [string, ActorInitialStateSnapshot]) => {
        const instanceId = Number(id);
        if (!isNaN(instanceId)) {
          const rawCurrentHp = state.currentHp;
          const rawMaxHp = state.maxHp;
          const rawTempHp = state.tempHp;

          let maxHp = rawMaxHp;
          const currentHpVal = Number(rawCurrentHp ?? 0);
          const tempHpVal = Number(rawTempHp ?? 0);

          // Attempt to find original frontend instanceId
          const actorConfig = (logActorConfigs as Actor[]).find((c: Actor) => c.instanceId === instanceId);
          let targetInstanceId = instanceId;

          // Try to find matching combatant in current session
          const combatants = this.combatantService.combatants();
          const combatant = combatants.find(c => c.instanceId === instanceId);

          if (!maxHp && actorConfig) {
            maxHp = actorConfig.hpConfig?.value || actorConfig.hpConfig?.hpAverage || actorConfig.state?.maxHp;
          }
          if (!maxHp && combatant) {
            maxHp = combatant.hpConfig?.value || combatant.hpConfig?.hpAverage || combatant.state?.maxHp;
          }

          if (combatant) {
            targetInstanceId = combatant.instanceId;
            backendToFrontendIdMap.set(instanceId, targetInstanceId);
            // If the session combatant has a valid maxHp, use it as the source of truth
            if (combatant.state?.maxHp > 0) {
              maxHp = combatant.state.maxHp;
            }
          } else if (actorConfig) {
            targetInstanceId = actorConfig.instanceId || instanceId;
            backendToFrontendIdMap.set(instanceId, targetInstanceId);
            // Also trust actorConfig maxHp
            if (actorConfig.hpConfig?.value || actorConfig.hpConfig?.hpAverage) {
              maxHp = actorConfig.hpConfig?.value || actorConfig.hpConfig?.hpAverage;
            } else if (actorConfig.state?.maxHp > 0) {
              maxHp = actorConfig.state.maxHp;
            }
          } else {
             // Try name-based matching for characters as a last resort.
             // (actorConfig is undefined in this branch; name comes from the combatant.)
             const configName = 'Unknown';
             const nameMatch = combatants.find((c: Actor) => c.name === configName && isCharacter(c));
             if (nameMatch) {
                 targetInstanceId = nameMatch.instanceId;
                 backendToFrontendIdMap.set(instanceId, targetInstanceId);
                 console.log(`[TimelineService] Name-based ID match for ${nameMatch.name}: mapped backend ${instanceId} to frontend ${targetInstanceId}`);
             } else {
                 backendToFrontendIdMap.set(instanceId, instanceId);
             }
          }

          const finalMaxHp = Number(maxHp ?? currentHpVal ?? 0);
          const actorName = actorConfig?.name || combatant?.name || 'Unknown';
          // console.log(`[TimelineService] Init Actor ${id} (mapped to ${targetInstanceId}): HP ${currentHpVal}/${finalMaxHp}, Name: ${actorName}`);

          // The backend initial-state snapshot only carries currentHp/maxHp/
          // tempHp/conditions/healthState. The remaining ActorState fields are
          // not part of the snapshot, so they start from sensible defaults.
          stateMap.set(targetInstanceId, {
            ...state,
            currentHp: currentHpVal,
            maxHp: finalMaxHp,
            tempHp: tempHpVal,
            hitDie: 10,
            conditions: state.conditions ?? {},
            deathSaves: { successes: 0, failures: 0 },
            resistances: {},
            isStable: true,
            isDead: state.healthState === 'dead',
            initiative: 0,
            isProjected: true
          } as unknown as ActorState);
        }
      });
    } else {
      // Fallback: Initialize from current combatants if actorInitialStates is missing
      this.combatantService.combatants().forEach(c => {
        stateMap.set(c.instanceId, {
          ...c.state,
          isProjected: true
        });
      });
    }

    // 2. Replay events up to the current index
    for (let i = 0; i <= index && i < events.length; i++) {
      const event = events[i];
      const data: EventData = event.data || {};
      const actorStates = event.actorStates || data.actorStates;

      if (actorStates) {
        Object.entries(actorStates).forEach(([id, state]: [string, ActorStateSnapshot]) => {
          const backendId = Number(id);
          const targetInstanceId = backendToFrontendIdMap.get(backendId) || backendId;
          const currentState = stateMap.get(targetInstanceId);

          if (currentState) {
            // Per-event snapshots (events.ActorSnapshot) carry no maxHp, so we
            // keep the existing maxHp and only update the fields they provide.
            const currentHp = state.currentHp;
            const tempHp = state.tempHp;
            const conditions = state.conditions;

            const finalCurrentHp = currentHp !== undefined ? Number(currentHp) : currentState.currentHp;
            const finalMaxHp = currentState.maxHp;

            if (currentState.isProjected) {
                const actorName = (logActorConfigs as Actor[]).find((c: Actor) => c.instanceId === backendId)?.name || 'Unknown';
                if (actorName.includes('Henry') || actorName.includes('Acolyte')) {
                    // console.log(`[TimelineService] Snapshot Update for ${actorName} (${targetInstanceId}): HP ${finalCurrentHp}/${finalMaxHp}. Event: ${event.type}`);
                }
            }

            stateMap.set(targetInstanceId, {
              ...currentState,
              currentHp: finalCurrentHp,
              maxHp: finalMaxHp,
              tempHp: Number(tempHp ?? currentState.tempHp),
              conditions: conditions || currentState.conditions,
              isDead: state.healthState?.toLowerCase() === 'dead',
              isStable: currentState.isStable
            } as unknown as ActorState);
          } else {
             // Fallback: if currentState is missing, we might have a new monster joining
             const actorConfig = (logActorConfigs as Actor[]).find((c: Actor) => c.instanceId === backendId);
             if (actorConfig) {
                 // Per-event snapshots have no maxHp, so derive it from the config.
                 const currentHp = state.currentHp;
                 const finalMaxHp = Math.max(1, Number(actorConfig.hpConfig?.hpAverage ?? actorConfig.state?.maxHp ?? currentHp ?? 1));

                 stateMap.set(targetInstanceId, {
                    ...actorConfig.state,
                    currentHp: Number(currentHp ?? finalMaxHp),
                    maxHp: finalMaxHp,
                    tempHp: Number(state.tempHp ?? 0),
                    conditions: state.conditions || {},
                    isDead: state.healthState?.toLowerCase() === 'dead',
                    isStable: true,
                    isProjected: true
                 } as unknown as ActorState);
             }
          }
        });

        // If we have a snapshot, it is the absolute source of truth for all actors included.
        // We skip manual processing for this event.
        continue;
      }

      switch (event.type) {
        case EventType.IntermissionHealing:
        case EventType.HPModified: {
          const rawActorId = data.actor?.instanceId || event.actor?.instanceId || data.actorId;
          // hp_modified carries the new HP inside `result` (backend
          // HPModificationResult). maxHp is never changed by an HP event.
          const newHp = data.result?.newHp;
          const newTempHp = data.result?.newTempHp;

          const actorId = (rawActorId !== undefined) ? (backendToFrontendIdMap.get(Number(rawActorId)) || Number(rawActorId)) : undefined;

          if (actorId !== undefined) {
            const instanceId = Number(actorId);
            const currentState = stateMap.get(instanceId);
            if (currentState) {
              const finalCurrentHp = newHp !== undefined ? Number(newHp) : currentState.currentHp;
              const finalMaxHp = currentState.maxHp;
              const finalTempHp = newTempHp !== undefined ? Number(newTempHp) : currentState.tempHp;

              const actorName = (logActorConfigs as Actor[]).find((c: Actor) => c.instanceId === Number(rawActorId))?.name || 'Unknown';
              if (actorName.includes('Henry') || actorName.includes('Acolyte')) {
                  // console.log(`[TimelineService] HP Update Event for ${actorName} (${instanceId}): HP ${finalCurrentHp}/${finalMaxHp}`);
              }

              stateMap.set(instanceId, {
                ...currentState,
                currentHp: finalCurrentHp,
                maxHp: finalMaxHp,
                tempHp: finalTempHp,
              });
            }
          }
          break;
        }
        case EventType.Initiative: {
          const rawActorId = data.actor?.instanceId || event.actor?.instanceId || data.actorId;
          const actorId = (rawActorId !== undefined) ? (backendToFrontendIdMap.get(Number(rawActorId)) || Number(rawActorId)) : undefined;
          const initiative = data.roll?.total ?? data.value;

          if (actorId !== undefined && initiative !== undefined) {
            const instanceId = Number(actorId);
            const currentState = stateMap.get(instanceId);
            if (currentState) {
              const newState = {
                ...currentState,
                initiative: Number(initiative)
              };
              stateMap.set(instanceId, newState);
            }
          }
          break;
        }
        case EventType.Unconscious: {
          const rawActorId = data.actor?.instanceId || event.actor?.instanceId || data.actorId;
          const actorId = (rawActorId !== undefined) ? (backendToFrontendIdMap.get(Number(rawActorId)) || Number(rawActorId)) : undefined;
          if (actorId !== undefined) {
            const instanceId = Number(actorId);
            const currentState = stateMap.get(instanceId);
            if (currentState) {
              const actorName = (logActorConfigs as Actor[]).find((c: Actor) => c.instanceId === Number(rawActorId))?.name || 'Unknown';
              if (actorName.includes('Henry') || actorName.includes('Acolyte')) {
                  // console.log(`[TimelineService] Unconscious Event for ${actorName} (${instanceId})`);
              }
              stateMap.set(instanceId, {
                ...currentState,
                conditions: { ...currentState.conditions, [Condition.Unconscious]: true }
              });
            }
          }
          break;
        }
        case EventType.Death: {
          const rawActorId = data.actor?.instanceId || event.actor?.instanceId || data.actorId;
          const actorId = (rawActorId !== undefined) ? (backendToFrontendIdMap.get(Number(rawActorId)) || Number(rawActorId)) : undefined;
          if (actorId !== undefined) {
            const instanceId = Number(actorId);
            const currentState = stateMap.get(instanceId);
            if (currentState) {
              const actorName = (logActorConfigs as Actor[]).find((c: Actor) => c.instanceId === Number(rawActorId))?.name || 'Unknown';
              if (actorName.includes('Henry') || actorName.includes('Acolyte')) {
                  // console.log(`[TimelineService] Death Event for ${actorName} (${instanceId})`);
              }
              stateMap.set(instanceId, {
                ...currentState,
                isDead: true,
                currentHp: 0
              });
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

    // console.log('[TimelineService] final stateMap Summary:');
    stateMap.forEach((state, id) => {
        // console.log(`  Actor ${id}: HP ${state.currentHp}/${state.maxHp}, Dead: ${state.isDead}`);
    });

    // 3. Return state map
    return stateMap;
  });
}
