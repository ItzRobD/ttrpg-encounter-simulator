import {computed, inject, Injectable, signal, WritableSignal} from '@angular/core';
import { ActorState, EventType, SimulationLog, Actor, ActorSummary } from '../models';
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

    // 1. Initialize with starting states
    // Try to use the initialState from the selected log (run-specific) or from the overall simulation result
    const initialState = log.initialState || simResult?.initialState;
    if (initialState) {
      Object.entries(initialState).forEach(([id, data]: [string, any]) => {
        const instanceId = parseInt(id, 10);
        if (data.state) {
          // Deep clone the state and provide defaults for missing required fields
          const baseState = {
            conditions: {},
            deathSaves: { successes: 0, failures: 0 },
            resistances: {},
            tempHp: 0,
            initiative: 0,
            isStable: true,
            isDead: false,
            ...data.state
          };
          stateMap.set(instanceId, baseState);
        }
      });
    } else if (log.actors) {
      // Fallback to log.actors (backward compatibility)
      log.actors.forEach(c => {
        stateMap.set(c.instanceId, { ...c.state });
      });
    }

    // 2. Replay events up to the current index
    for (let i = 0; i <= index && i < events.length; i++) {
      const event = events[i];
      if (!event.data) continue;

      switch (event.type) {
        // TODO: Conditions, death saves
        case EventType.HPModified: {
          const targetId = event.data.target?.instanceId;
          const newHp = event.data.finalHp;
          const newTempHp = event.data.finalTempHp;

          if (targetId !== undefined && newHp !== undefined) {
            const currentState = stateMap.get(targetId);
            if (currentState) {
              const maxHp = currentState.maxHp || 1;
              stateMap.set(targetId, {
                ...currentState,
                currentHp: Math.max(0, Math.min(newHp, maxHp)),
                tempHp: newTempHp != undefined ? Math.max(0, newTempHp) : currentState.tempHp
              });
            }
          }
          break;
        }
        case EventType.Initiative: {
          const actorId = event.data.actor?.instanceId;
          const initiative = event.data.roll?.total;

          if (actorId !== undefined && initiative !== undefined) {
            const currentState = stateMap.get(actorId);
            if (currentState) {
              stateMap.set(actorId, {
                ...currentState,
                initiative: initiative
              });
            }
          }
          break;
        }
      }
    }

    return stateMap;
  });
}
