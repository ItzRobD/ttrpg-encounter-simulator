import { Injectable, signal, computed, effect, inject } from '@angular/core';
import {
  ActorSummary,
  Race,
  Class,
  ResistanceType,
  WeaponSlot,
  DiceType,
  DamageType,
  MonsterSize,
  MonsterType,
  Condition,
  CasterType,
  LevelType,
  isCharacter,
  isMonster,
  Actor,
  SimulationResult,
  IndividualResult,
  EncounterResult,
  SimulationEvent,
  CombatantReference
} from '../models';
import { environment } from '../../environments/environment';
import { LocalStorageService } from './local-storage.service';

@Injectable({
  providedIn: 'root',
})
export class CombatantService {
  private readonly localStorage = inject(LocalStorageService);
  private readonly PARTY_STORAGE_KEY = 'dnd5e_active_party';
  private readonly ENCOUNTERS_STORAGE_KEY = 'dnd5e_active_encounters';

  private readonly _party = signal<Actor[]>(
    this.localStorage.getItem<Actor[]>(this.PARTY_STORAGE_KEY) || []
  );

  private readonly _encounters = signal<Actor[][]>(
    this.localStorage.getItem<Actor[][]>(this.ENCOUNTERS_STORAGE_KEY) || [[]]
  );

  private readonly _activeEncounterIndex = signal<number>(0);

  // Public readonly signals
  readonly party = this._party.asReadonly();
  readonly encounters = this._encounters.asReadonly();
  readonly activeEncounterIndex = this._activeEncounterIndex.asReadonly();

  readonly activeEncounter = computed(() => this._encounters()[this._activeEncounterIndex()] || []);

  readonly combatants = computed(() => [...this._party(), ...this.activeEncounter()]);

  readonly count = computed(() => this.combatants().length);

  constructor() {
    // Automatically persist to localStorage whenever the list changes
    effect(() => {
      this.localStorage.setItem(this.PARTY_STORAGE_KEY, this._party());
    });
    effect(() => {
      this.localStorage.setItem(this.ENCOUNTERS_STORAGE_KEY, this._encounters());
    });
  }

  // Helper to get the next unique instanceId
  private getNextInstanceId(): number {
    const allCombatants = [...this._party(), ...this._encounters().flat()];
    if (allCombatants.length === 0) return 1;
    return Math.max(...allCombatants.map(a => a.instanceId)) + 1;
  }

  setActiveEncounter(index: string | number | undefined): void {
    if (index === undefined) return;
    const idx = typeof index === 'string' ? parseInt(index, 10) : index;
    if (idx >= 0 && idx < this._encounters().length) {
      this._activeEncounterIndex.set(idx);
    }
  }

  addEncounter(): void {
    this._encounters.update(e => [...e, []]);
    this._activeEncounterIndex.set(this._encounters().length - 1);
  }

  removeEncounter(index: number): void {
    this._encounters.update(e => {
      if (e.length <= 1) return [[]];
      const newList = e.filter((_, i) => i !== index);
      return newList;
    });

    // Adjust active index if needed
    if (this._activeEncounterIndex() >= this._encounters().length) {
      this._activeEncounterIndex.set(this._encounters().length - 1);
    }
  }

  /**
   * Adds a new actor to the party or current encounter if limits allow.
   */
  addCombatant(actor: Actor): boolean {
    const isChar = isCharacter(actor);
    const currentCharacters = this._party().length;

    // Check that we're below the max actors
    if (this.count() >= environment.limits.maxTotal) return false;

    // Check if we're already at the max number of characters
    if (isChar && currentCharacters >= environment.limits.maxCharacters) return false;

    const newCombatant = {
      ...actor,
      instanceId: this.getNextInstanceId()
    };

    if (isChar) {
      this._party.update(list => [...list, newCombatant]);
    } else {
      this._encounters.update(encs => {
        const newEncs = [...encs];
        newEncs[this._activeEncounterIndex()] = [...newEncs[this._activeEncounterIndex()], newCombatant];
        return newEncs;
      });
    }
    return true;
  }

  /**
   * Adds an actor to the simulator from the library (characters or bestiary).
   * This method clones the actor and initializes it for combat simulation.
   */
  addToSimulator(actor: Actor): boolean {
    const isChar = isCharacter(actor);

    // Create a fresh copy with combat-ready state
    const combatActor: Actor = {
      ...actor,
      instanceId: 0, // Will be set by addCombatant
      state: {
        ...actor.state,
        // Initialize HP if not already set
        currentHp: actor.state?.currentHp || actor.hpConfig?.hpAverage || actor.hpConfig?.value || 1,
        maxHp: actor.state?.maxHp || actor.hpConfig?.hpAverage || actor.hpConfig?.value || 1,
        tempHp: 0,
        initiative: 0,
        isStable: true,
        isDead: false,
        conditions: Object.values(Condition).reduce((acc, curr) => ({ ...acc, [curr]: false }), {} as Record<Condition, boolean>),
        deathSaves: { successes: 0, failures: 0 },
        resistances: actor.state?.resistances || actor.resistances || Object.values(DamageType).reduce((acc, curr) => ({ ...acc, [curr]: ResistanceType.None }), {} as Record<DamageType, ResistanceType>),
      }
    };

    return this.addCombatant(combatActor);
  }

  /**
   * Removes an actor by its unique instanceId.
   */
  removeCombatant(instanceId: number): void {
    if (this._party().some(a => a.instanceId === instanceId)) {
      this._party.update(list => list.filter(a => a.instanceId !== instanceId));
    } else {
      this._encounters.update(encs => encs.map(e => e.filter(a => a.instanceId !== instanceId)));
    }
  }

  /**
   * Clears the current active encounter.
   */
  clearActiveEncounter(): void {
    this._encounters.update(encs => {
      const newEncs = [...encs];
      newEncs[this._activeEncounterIndex()] = [];
      return newEncs;
    });
  }

  /**
   * Clears everything.
   */
  clearAll(): void {
    this._party.set([]);
    this._encounters.set([[]]);
    this._activeEncounterIndex.set(0);
  }

  /**
   * Updates a specific combatant.
   */
  updateCombatant(instanceId: number, updates: Partial<Actor>): void {
    if (this._party().some(a => a.instanceId === instanceId)) {
      this._party.update(list =>
        list.map(a => a.instanceId === instanceId ? { ...a, ...updates } : a)
      );
    } else {
      this._encounters.update(encs =>
        encs.map(e => e.map(a => a.instanceId === instanceId ? { ...a, ...updates } : a))
      );
    }
  }

  /**
   * Reorders a combatant in the list.
   */
  reorderCombatant(fromIndex: number, toIndex: number): void {
    // Note: Reordering is currently only supporting reordering within the same list (party or active encounter)
    // and depends on how the UI presents the list.
    // If it's a flat list of all combatants:
    const all = [...this.combatants()];
    const [movedItem] = all.splice(fromIndex, 1);
    all.splice(toIndex, 0, movedItem);

    // Distribute back to party and active encounter
    this._party.set(all.filter(a => isCharacter(a)));
    this._encounters.update(encs => {
      const newEncs = [...encs];
      newEncs[this._activeEncounterIndex()] = all.filter(a => isMonster(a));
      return newEncs;
    });
  }

  /**
   * Sorts the combatants by initiative (descending).
   */
  sortByInitiative(): void {
    this._party.update(list => [...list].sort((a, b) => b.state.initiative - a.state.initiative));
    this._encounters.update(encs => encs.map(e => [...e].sort((a, b) => b.state.initiative - a.state.initiative)));
  }

  /**
   * Loads combatants from a simulation result.
   */
  loadFromSimulation(result: SimulationResult): void {
    console.log('[CombatantService] loadFromSimulation called with:', result);

    // actorConfigs now includes BOTH characters and monsters
    const allActorConfigs = result.actorConfigs || [];

    // Extract characters from actorConfigs
    const characters = allActorConfigs.filter(a => isCharacter(a)).map(c => ({
      ...c,
      state: {
        ...c.state,
        currentHp: c.hpConfig?.hpAverage || c.hpConfig?.value || c.state?.currentHp || 0,
        maxHp: c.hpConfig?.hpAverage || c.hpConfig?.value || c.state?.maxHp || 0,
      }
    }));

    this._party.set(characters);

    // Extract encounters from the first individual result
    if (result.individualResults && result.individualResults.length > 0) {
      const firstRun: IndividualResult = result.individualResults[0];
      const encounters: Actor[][] = [];

      firstRun.encounterResults.forEach((er: EncounterResult, erIndex: number) => {
        const monstersInEncounter: Actor[] = [];
        const seenInEncounter = new Set<number>();

        // 1. First, check logs for actors present in this encounter
        er.logs.forEach((event: SimulationEvent) => {
          const checkActor = (ref?: CombatantReference) => {
            if (ref?.instanceId && !seenInEncounter.has(ref.instanceId)) {
                seenInEncounter.add(ref.instanceId);
                const config = allActorConfigs.find(a => a.instanceId === ref.instanceId);
                if (config && isMonster(config)) {
                    monstersInEncounter.push({
                      ...config,
                      state: {
                        ...config.state,
                        currentHp: config.hpConfig?.hpAverage || config.hpConfig?.value || config.state?.currentHp || 0,
                        maxHp: config.hpConfig?.hpAverage || config.hpConfig?.value || config.state?.maxHp || 0,
                      }
                    });
                }
            }
          };

          checkActor(event.actor);
          if (event.data?.target) checkActor(event.data.target);
          if (event.data?.actorId) checkActor({ instanceId: Number(event.data.actorId), name: '', type: '' });
          if (event.data?.targetId) checkActor({ instanceId: Number(event.data.targetId), name: '', type: '' });
          if (event.data?.actorStates) {
              Object.keys(event.data.actorStates).forEach(id => {
                  checkActor({ instanceId: Number(id), name: '', type: '' });
              });
          }
        });

        // 2. Also check encounter level initialState if logs were sparse
        if (er.initialState) {
            Object.keys(er.initialState).forEach(id => {
                const instanceId = Number(id);
                if (!seenInEncounter.has(instanceId)) {
                    const config = allActorConfigs.find(a => a.instanceId === instanceId);
                    if (config && isMonster(config)) {
                        seenInEncounter.add(instanceId);
                        monstersInEncounter.push({
                          ...config,
                          state: {
                            ...config.state,
                            currentHp: config.hpConfig?.hpAverage || config.hpConfig?.value || config.state?.currentHp || 0,
                            maxHp: config.hpConfig?.hpAverage || config.hpConfig?.value || config.state?.maxHp || 0,
                          }
                        });
                    }
                }
            });
        }

        encounters.push(monstersInEncounter);
      });

      if (encounters.length > 0) {
        this._encounters.set(encounters);
      } else {
        this._encounters.set([[]]);
      }
    }

    this._activeEncounterIndex.set(0);
  }
  }
