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
  }

  private readonly _currentTimelineIndex: WritableSignal<number> = signal(0);

  get currentTimelineIndex(): number {
    return this._currentTimelineIndex();
  }

  set currentTimelineIndex(index: number) {
    this._currentTimelineIndex.set(index);
  }

  incrementTimelineIndex(): void {
    this._currentTimelineIndex.update(i => i + 1);
  }

  decrementTimelineIndex(): void {
    this._currentTimelineIndex.update(i => i - 1);
  }

  readonly projectedState = computed(() => {
    const log = this._selectedSimulationLog();
    if (!log) {
      return new Map<number, EntityState>();
    }

    const events = log.events;
    const index = this._currentTimelineIndex();

    const stateMap = new Map<number, EntityState>();

    // 1. Initialize with starting states from CombatantService
    this.combatantService.combatants().forEach(c => {
      stateMap.set(c.instanceId, { ...c.state });
    });

    // 2. Replay events up to the current index
    for (let i = 0; i <= index && i < events.length; i++) {
      const event = events[i];
      switch (event.type) {
        // TODO: Conditions, death saves
        case EventType.HPModified: {
          const targetID = event.data.target?.instanceId;
          const newHP = event.data.finalHP;
          const newTempHP = event.data.finalTempHP;

          if (targetID !== undefined && newHP !== undefined) {
            const currentState = stateMap.get(targetID);
            if (currentState) {
              stateMap.set(targetID, {
                ...currentState,
                currentHP: newHP,
                tempHP: newTempHP != undefined ? newTempHP : currentState.tempHP
              });
            }
          }
          break;
        }
        case EventType.Initiative: {
          const actorID = event.data.actor?.instanceId;
          const initiative = event.data.roll?.total;

          if (actorID !== undefined && initiative !== undefined) {
            const currentState = stateMap.get(actorID);
            if (currentState) {
              stateMap.set(actorID, {
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
