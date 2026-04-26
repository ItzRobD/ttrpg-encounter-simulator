import {computed, inject, Injectable, signal, WritableSignal} from '@angular/core';
import { ActorState, EventType, SimulationLog, Actor, ActorSummary, Condition, ActorStateSnapshot, isCharacter } from '../models';
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
      Object.entries(initialStates).forEach(([id, state]: [string, ActorStateSnapshot]) => {
        const instanceId = Number(id);
        if (!isNaN(instanceId)) {
          const rawCurrentHp = state.currentHp ?? (state as any).current_hp;
          const rawMaxHp = (state as any).maxHp ?? (state as any).max_hp ?? (state as any).maxHP ?? (state as any).MaxHP;
          const rawTempHp = state.tempHp ?? (state as any).temp_hp;

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
             // Try name-based matching for characters as a last resort
             const configName = (actorConfig as any)?.name || 'Unknown';
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

          stateMap.set(targetInstanceId, {
            ...state,
            // Ensure properties match ActorState interface
            currentHp: currentHpVal,
            maxHp: finalMaxHp,
            tempHp: tempHpVal,
            hitDie: (state as any).hitDie ?? (state as any).hit_die ?? 10,
            conditions: (state as any).conditions || {},
            deathSaves: (state as any).deathSaves || { successes: 0, failures: 0 },
            resistances: (state as any).resistances || {},
            isStable: (state as any).isStable ?? (state as any).stable ?? true,
            isDead: (state as any).isDead ?? (state.healthState === 'dead' || (state as any).health_state === 'dead'),
            initiative: (state as any).initiative ?? 0,
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
      const data = event.data || {};
      const actorStates = event.actorStates || data.actorStates;

      if (actorStates) {
        Object.entries(actorStates).forEach(([id, state]: [string, ActorStateSnapshot]) => {
          const backendId = Number(id);
          const targetInstanceId = backendToFrontendIdMap.get(backendId) || backendId;
          const currentState = stateMap.get(targetInstanceId);

          if (currentState) {
            // Map snapshot state to ActorState interface
            const currentHp = state.currentHp ?? (state as any).current_hp ?? (state as any).CurrentHP ?? (state as any).CurrentHp;
            const maxHpSnap = (state as any).maxHp ?? (state as any).max_hp ?? (state as any).MaxHP ?? (state as any).MaxHp;
            const tempHp = state.tempHp ?? (state as any).temp_hp ?? (state as any).TempHP ?? (state as any).TempHp;
            const conditions = (state as any).conditions || state.conditions;

            const finalCurrentHp = currentHp !== undefined ? Number(currentHp) : currentState.currentHp;
            // IMPORTANT: If maxHp in snapshot is 0 or missing, DO NOT overwrite the existing valid maxHp.
            let finalMaxHp = currentState.maxHp;
            if (maxHpSnap !== undefined && Number(maxHpSnap) > 0) {
                finalMaxHp = Number(maxHpSnap);
            }

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
              isDead: (state.healthState?.toLowerCase() === 'dead' || (state as any).health_state?.toLowerCase() === 'dead' || (state as any).HealthState?.toLowerCase() === 'dead'),
              isStable: (state as any).isStable ?? (state as any).stable ?? (state as any).IsStable ?? currentState.isStable
            });
          } else {
             // Fallback: if currentState is missing, we might have a new monster joining
             const actorConfig = (logActorConfigs as Actor[]).find((c: Actor) => c.instanceId === backendId);
             if (actorConfig) {
                 const currentHp = state.currentHp ?? (state as any).current_hp ?? (state as any).CurrentHP ?? (state as any).CurrentHp;
                 const maxHpSnap = (state as any).maxHp ?? (state as any).max_hp ?? (state as any).MaxHP ?? (state as any).MaxHp;
                 const finalMaxHp = Math.max(1, Number(maxHpSnap ?? actorConfig.hpConfig?.hpAverage ?? actorConfig.state?.maxHp ?? currentHp ?? 1));

                 stateMap.set(targetInstanceId, {
                    ...actorConfig.state,
                    currentHp: Number(currentHp ?? finalMaxHp),
                    maxHp: finalMaxHp,
                    tempHp: Number(state.tempHp ?? (state as any).temp_hp ?? 0),
                    conditions: (state as any).conditions || state.conditions || {},
                    isDead: (state.healthState?.toLowerCase() === 'dead' || (state as any).health_state?.toLowerCase() === 'dead'),
                    isStable: (state as any).isStable ?? (state as any).stable ?? true,
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
          const edata = data as any;
          const currentHp = edata.currentHp ?? edata.current_hp ?? edata.CurrentHP ?? edata.CurrentHp;
          const maxHp = edata.maxHp ?? edata.max_hp ?? edata.MaxHP ?? edata.MaxHp;
          const tempHp = edata.tempHp ?? edata.temp_hp ?? edata.TempHP ?? edata.TempHp;

          const actorId = (rawActorId !== undefined) ? (backendToFrontendIdMap.get(Number(rawActorId)) || Number(rawActorId)) : undefined;

          if (actorId !== undefined) {
            const instanceId = Number(actorId);
            const currentState = stateMap.get(instanceId);
            if (currentState) {
              const finalCurrentHp = currentHp !== undefined ? Number(currentHp) : currentState.currentHp;
              // IMPORTANT: If maxHp in event is 0 or missing, DO NOT overwrite the existing valid maxHp.
              let finalMaxHp = currentState.maxHp;
              if (maxHp !== undefined && Number(maxHp) > 0) {
                  finalMaxHp = Number(maxHp);
              }
              const finalTempHp = tempHp !== undefined ? Number(tempHp) : currentState.tempHp;

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
