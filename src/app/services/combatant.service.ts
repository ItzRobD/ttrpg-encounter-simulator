import { Injectable, signal, computed, effect, inject } from '@angular/core';
import {
  Entity,
  Monster,
  Character,
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
  SpecialAbilities,
  LevelType,
  isCharacter,
  isMonster
} from '../models';
import { environment } from '../../environments/environment';
import { LocalStorageService } from './local-storage.service';

@Injectable({
  providedIn: 'root',
})
export class CombatantService {
  private readonly localStorage = inject(LocalStorageService);
  private readonly ENCOUNTER_STORAGE_KEY = 'dnd5e_active_encounter';

  private readonly _combatants = signal<Entity[]>(
    this.localStorage.getItem<Entity[]>(this.ENCOUNTER_STORAGE_KEY) || []
  );

  // Public readonly signals
  readonly combatants = this._combatants.asReadonly();

  readonly count = computed(() => this._combatants().length);

  readonly monsters = computed(() =>
    this._combatants().filter((e): e is Monster => isMonster(e))
  );

  readonly characters = computed(() =>
    this._combatants().filter((e): e is Character => isCharacter(e))
  );

  constructor() {
    // Automatically persist to localStorage whenever the list changes
    effect(() => {
      this.localStorage.setItem(this.ENCOUNTER_STORAGE_KEY, this._combatants());
    });
  }

  // Helper to get the next unique instanceId
  private getNextInstanceId(): number {
    const current = this._combatants();
    if (current.length === 0) return 1;
    return Math.max(...current.map(e => e.instanceId)) + 1;
  }

  /**
   * Adds a new entity to the encounter if limits allow.
   * Fluid limits:
   * 1. Overall total (maxTotal) is the primary constraint.
   * 2. Characters are hard-capped at maxCharacters.
   * 3. Monsters can fill any remaining capacity up to maxTotal.
   *    (e.g., if maxTotal is 23 and characters are 0, you can have 23 monsters).
   */
  addCombatant(entity: Entity): boolean {
    const isChar = isCharacter(entity);
    const currentCharacters = this.characters().length;

    // Check that we're below the max entities
    if (this.count() >= environment.limits.maxTotal) return false;

    // Check if we're already at the max number of characters
    if (isChar && currentCharacters >= environment.limits.maxCharacters) return false;

    // Note: No monster-specific cap is checked here to allow them to fill
    // the remaining capacity as requested (fluid behavior).

    const newCombatant = {
      ...entity,
      instanceId: this.getNextInstanceId()
    };

    this._combatants.update(list => [...list, newCombatant]);
    return true;
  }

  /**
   * Adds an entity to the simulator from the library (characters or bestiary).
   * This method clones the entity and initializes it for combat simulation.
   */
  addToSimulator(entity: Entity): boolean {
    const isChar = isCharacter(entity);

    // Create a fresh copy with combat-ready state
    const combatEntity: Entity = {
      ...entity,
      instanceId: 0, // Will be set by addCombatant
      state: {
        ...entity.state,
        // Initialize HP if not already set
        currentHp: entity.state?.currentHp || entity.hp?.hpAverage || entity.hp?.value || 1,
        maxHp: entity.state?.maxHp || entity.hp?.hpAverage || entity.hp?.value || 1,
        tempHp: 0,
        initiative: 0,
        isStable: true,
        isDead: false,
        conditions: Object.values(Condition).reduce((acc, curr) => ({ ...acc, [curr]: false }), {} as Record<Condition, boolean>),
        deathSaves: { successes: 0, failures: 0 },
        resistances: entity.state?.resistances || Object.values(DamageType).reduce((acc, curr) => ({ ...acc, [curr]: ResistanceType.None }), {} as Record<DamageType, ResistanceType>),
      }
    };

    // Re-ensure type-specific properties are present if they might have been lost or to help type guards
    if (isChar && isCharacter(combatEntity)) {
      const charEntity = combatEntity as Character;
      const originalChar = entity as Character;
      if (!charEntity.class && originalChar.class) {
        charEntity.class = originalChar.class;
      }
      if (!charEntity.classId && originalChar.classId) {
        charEntity.classId = originalChar.classId;
      }
      if (!charEntity.level && originalChar.level) {
        charEntity.level = originalChar.level;
      }
    }

    return this.addCombatant(combatEntity);
  }

  /**
   * Removes an entity from the encounter by its unique instanceId.
   */
  removeCombatant(instanceId: number): void {
    this._combatants.update(list => list.filter(e => e.instanceId !== instanceId));
  }

  /**
   * Clears all combatants from the encounter.
   */
  clearEncounter(): void {
    this._combatants.set([]);
  }

  /**
   * Updates a specific combatant.
   */
  updateCombatant(instanceId: number, updates: Partial<Entity>): void {
    this._combatants.update(list =>
      list.map(e => e.instanceId === instanceId ? { ...e, ...updates } : e)
    );
  }

  /**
   * Reorders a combatant in the list.
   */
  reorderCombatant(fromIndex: number, toIndex: number): void {
    this._combatants.update(list => {
      const newList = [...list];
      const [movedItem] = newList.splice(fromIndex, 1);
      newList.splice(toIndex, 0, movedItem);
      return newList;
    });
  }

  /**
   * Sorts the combatants by initiative (descending).
   */
  sortByInitiative(): void {
    this._combatants.update(list => [...list].sort((a, b) => b.state.initiative - a.state.initiative));
  }
  }
