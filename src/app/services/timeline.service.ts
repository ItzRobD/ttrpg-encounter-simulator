import {computed, inject, Injectable, signal, WritableSignal} from '@angular/core';
import {EntityState, EventType, SimulationLog} from '../models';
import {CombatantService} from './combatant.service';

@Injectable({
  providedIn: 'root',
})
export class TimelineService {
  private readonly combatantService = inject(CombatantService);
  private readonly _selectedSimulationLog = signal<SimulationLog | null>(null);
  readonly selectedSimulationLog = this._selectedSimulationLog.asReadonly();

  readonly activeEvent = computed(() => {
    const log = this._selectedSimulationLog();
    const index = this._currentTimelineIndex();
    if (!log || index < 0 || index >= log.events.length) {
      return null;
    }
    return log.events[index];
  });

  setSelectedSimulationLog(log: SimulationLog | null): void {
    this._selectedSimulationLog.set(log);
    this._scrubberIndex.set(0);
    this._currentTimelineIndex.set(0);
  }

  private readonly _currentTimelineIndex: WritableSignal<number> = signal(0);

  /**
   * The list of events that the user can scrub through (Rounds, Turns, Choices).
   */
  readonly scrubbableEvents = computed(() => {
    const log = this._selectedSimulationLog();
    if (!log) return [];
    return log.events.filter(e =>
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
    if (scrubbable[index]) {
      const log = this._selectedSimulationLog();
      if (log) {
        const fullIndex = log.events.findIndex(e => e.id === scrubbable[index].id);
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
    const log = this._selectedSimulationLog();
    if (!log) {
      return new Map<number, EntityState>();
    }

    const events = log.events;
    const index = this._currentTimelineIndex();

    const stateMap = new Map<number, EntityState>();

    // 1. Initialize with starting states from the log's entities
    log.entities.forEach(c => {
      stateMap.set(c.instanceId, { ...c.state });
    });

    // 2. Replay events up to the current index
    for (let i = 0; i <= index && i < events.length; i++) {
      const event = events[i];
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
