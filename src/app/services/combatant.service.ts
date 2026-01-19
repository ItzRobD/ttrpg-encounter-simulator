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

  /**
   * Seeds the encounter with specific dummy data matching the timeline_output.json.
   */
  seedTimelineDummyData(): void {
    this.clearEncounter();

    // Bob (Character)
    const bob: Character = {
      id: 100,
      instanceId: 0,
      name: 'Bob',
      raceId: 4, // Human
      race: Race.Human,
      classId: 5, // Fighter
      class: Class.Fighter,
      level: 1,
      asConfig: {
        abilityScores: {
          strength: 10,
          dexterity: 10,
          constitution: 10,
          intelligence: 10,
          wisdom: 10,
          charisma: 10,
        },
        proficiencies: {
          strength: false,
          dexterity: false,
          constitution: false,
          intelligence: false,
          wisdom: false,
          charisma: false,
        },
      },
      state: {
        currentHp: 150,
        maxHp: 150,
        tempHp: 0,
        hitDie: 10,
        conditions: Object.values(Condition).reduce((acc, curr) => ({ ...acc, [curr]: false }), {} as Record<Condition, boolean>),
        deathSaves: { successes: 0, failures: 0 },
        resistances: Object.values(DamageType).reduce((acc, curr) => ({ ...acc, [curr]: ResistanceType.None }), {} as Record<DamageType, ResistanceType>),
        isStable: true,
        isDead: false,
        initiative: 0,
      },
      equipment: {
        armorId: 1, // Plate
        hasShieldEquipped: false,
        primarySlot: [
          {
            weaponId: 101, // Longsword
            isProficient: true,
            modifiers: {
              isMagic: false,
              isSilvered: false,
              isAdamantine: false,
              isColdForgedIron: false,
              attackBonus: 0,
              damageBonus: 0,
            },
          },
        ],
        secondarySlot: [],
        rangedSlot: []
      },
      hp: {
        hpSetMethod: 0,
        value: 150,
        hpAverage: 150,
        numberOfDice: 15,
        hitDie: 10,
        amountToAdd: 0,
        modifier: 0,
      },
      specialAbilities: {} as SpecialAbilities,
    };

    // Acolyte (Monster)
    const acolyte: Monster = {
      id: 200,
      instanceId: 1,
      name: 'Acolyte',
      size: MonsterSize.Medium,
      type: MonsterType.Humanoid,
      cr: 0.25,
      proficiencyBonus: 2,
      isLegendary: false,
      isSpellcaster: true,
      isInnateSpellcaster: false,
      specialAbilities: {} as SpecialAbilities,
      ac: 10,
      hp: {
        hpSetMethod: 0,
        value: 0,
        hpAverage: 9,
        numberOfDice: 2,
        hitDie: 8,
        amountToAdd: 0,
        modifier: 0,
      },
      monsterActions: {
        actions: [
          {
            actionId: 1,
            name: 'Club',
            rechargeValue: 0,
            hasDC: false,
            index: 0,
            attackBonus: 2,
            damageBlocks: [
              {
                numberOfDice: 1,
                die: DiceType.D4,
                amountToAdd: 0,
                damageType: DamageType.Bludgeoning,
              }
            ],
          },
        ],
        multiattacks: [],
        legendaryActions: [],
        rechargeActions: {},
      },
      asConfig: {
        abilityScores: {
          strength: 10,
          dexterity: 10,
          constitution: 10,
          intelligence: 10,
          wisdom: 14,
          charisma: 11,
        },
        proficiencies: {
          strength: false,
          dexterity: false,
          constitution: false,
          intelligence: false,
          wisdom: false,
          charisma: false,
        },
      },
      state: {
        currentHp: 11,
        maxHp: 11,
        tempHp: 0,
        hitDie: 8,
        conditions: Object.values(Condition).reduce((acc, curr) => ({ ...acc, [curr]: false }), {} as Record<Condition, boolean>),
        deathSaves: { successes: 0, failures: 0 },
        resistances: Object.values(DamageType).reduce((acc, curr) => ({ ...acc, [curr]: ResistanceType.None }), {} as Record<DamageType, ResistanceType>),
        isStable: true,
        isDead: false,
        initiative: 0,
      },
      spellcasting: {
        casterType: CasterType.Full,
        casterLevel: 1,
        spellSaveDC: 12,
        spellAttackBonus: 4,
        spellSlots: {
          1: { current: 3, max: 3 }
        },
        spells: [
          {
            id: 1,
            name: 'Light',
            level: 0,
            description: '',
            isConcentration: false,
            castingTime: 'action' as any,
            isRitual: false,
            spellType: 'healing' as any,
            isAOE: false,
            hasDC: false,
            isAutoHit: true,
            levelType: LevelType.Slot,
            spellDC: { ability: '', onSuccess: '' },
          },
          {
            id: 2,
            name: 'Sacred Flame',
            level: 0,
            description: '',
            isConcentration: false,
            castingTime: 'action' as any,
            isRitual: false,
            spellType: 'damage' as any,
            isAOE: false,
            hasDC: true,
            isAutoHit: false,
            levelType: LevelType.Slot,
            spellDC: { ability: '', onSuccess: '' },
          },
          {
            id: 3,
            name: 'Thaumaturgy',
            level: 0,
            description: '',
            isConcentration: false,
            castingTime: 'action' as any,
            isRitual: false,
            spellType: 'healing' as any,
            isAOE: false,
            hasDC: false,
            isAutoHit: true,
            levelType: LevelType.Slot,
            spellDC: { ability: '', onSuccess: '' },
          },
          {
            id: 4,
            name: 'Bless',
            level: 1,
            description: '',
            isConcentration: true,
            castingTime: 'action' as any,
            isRitual: false,
            spellType: 'healing' as any,
            isAOE: true,
            hasDC: false,
            isAutoHit: true,
            levelType: LevelType.Slot,
            spellDC: { ability: '', onSuccess: '' },
          },
          {
            id: 5,
            name: 'Cure Wounds',
            level: 1,
            description: '',
            isConcentration: false,
            castingTime: 'action' as any,
            isRitual: false,
            spellType: 'healing' as any,
            isAOE: false,
            hasDC: false,
            isAutoHit: true,
            levelType: LevelType.Slot,
            spellDC: { ability: '', onSuccess: '' },
          },
          {
            id: 6,
            name: 'Sanctuary',
            level: 1,
            description: '',
            isConcentration: false,
            castingTime: 'bonus action' as any,
            isRitual: false,
            spellType: 'healing' as any,
            isAOE: false,
            hasDC: true,
            isAutoHit: false,
            levelType: LevelType.Slot,
            spellDC: { ability: '', onSuccess: '' },
          },
        ],
      }
    };

    console.log('Seeding Acolyte with spells (ID 200):', acolyte);
    const entities = [bob, acolyte];
    this._combatants.set(entities);
  }

  /**
   * Seeds the encounter with dummy data for testing.
   */
  seedDummyData(): void {
    const dummyCharacter: Character = {
      id: 1,
      instanceId: 0,
      name: 'Valerius the Brave',
      raceId: 4, // Human
      race: Race.Human,
      classId: 7, // Paladin
      class: Class.Paladin,
      level: 5,
      asConfig: {
        abilityScores: {
          strength: 16,
          dexterity: 12,
          constitution: 14,
          intelligence: 10,
          wisdom: 12,
          charisma: 14,
        },
        proficiencies: {
          strength: true,
          dexterity: false,
          constitution: true,
          intelligence: false,
          wisdom: false,
          charisma: false,
        },
      },
      ac: 20,
      state: {
        currentHp: 40,
        maxHp: 63,
        tempHp: 12,
        hitDie: 10,
        conditions: Object.values(Condition).reduce((acc, curr) => ({ ...acc, [curr]: curr === Condition.Charmed || curr === Condition.Poisoned || curr === Condition.Prone }), {} as Record<Condition, boolean>),
        deathSaves: { successes: 2, failures: 1 },
        resistances: {
          ...Object.values(DamageType).reduce((acc, curr) => ({ ...acc, [curr]: ResistanceType.None }), {} as Record<DamageType, ResistanceType>),
          acid: ResistanceType.Immune,
          bludgeoning: ResistanceType.Resistant,
          fire: ResistanceType.Vulnerable,
          piercing: ResistanceType.Resistant,
          slashing: ResistanceType.Resistant,
        },
        isStable: true,
        isDead: false,
        initiative: 14,
      },
      equipment: {
        armorId: 1, // Plate
        hasShieldEquipped: true,
        primarySlot: [
          {
            weaponId: 101, // Longsword
            isProficient: true,
            modifiers: {
              isMagic: true,
              isSilvered: false,
              isAdamantine: false,
              isColdForgedIron: false,
              attackBonus: 1,
              damageBonus: 1,
            },
          },
        ],
        secondarySlot: [
          {
            weaponId: 102, // Shortsword
            isProficient: true,
            modifiers: {
              isMagic: false,
              isSilvered: false,
              isAdamantine: false,
              isColdForgedIron: false,
              attackBonus: 0,
              damageBonus: 0,
            },
          }
        ],
        rangedSlot: [
          {
            weaponId: 103, // Longbow
            isProficient: true,
            modifiers: {
              isMagic: false,
              isSilvered: false,
              isAdamantine: false,
              isColdForgedIron: false,
              attackBonus: 0,
              damageBonus: 0,
            },
          }
        ]
      },
      hp: {
        hpSetMethod: 0,
        value: 63,
        hpAverage: 63,
        numberOfDice: 9,
        hitDie: 10,
        amountToAdd: 0,
        modifier: 0,
      },
    };

    const dummyMonster: Monster = {
      id: 2,
      instanceId: 0,
      name: 'Ancient Red Dragon',
      size: MonsterSize.Gargantuan,
      type: MonsterType.Dragon,
      cr: 24,
      proficiencyBonus: 7,
      isInnateSpellcaster: false,
      isLegendary: true,
      isSpellcaster: true,
      ac: 19,
      hp: {
        hpSetMethod: 0,
        value: 0,
        hpAverage: 546,
        numberOfDice: 28,
        hitDie: 20,
        amountToAdd: 252,
        modifier: 9,
      },
      asConfig: {
        abilityScores: {
          strength: 30,
          dexterity: 10,
          constitution: 29,
          intelligence: 18,
          wisdom: 15,
          charisma: 23,
        },
        proficiencies: {
          strength: false,
          dexterity: true,
          constitution: true,
          intelligence: false,
          wisdom: true,
          charisma: true,
        },
      },
      state: {
        currentHp: 546,
        maxHp: 546,
        tempHp: 0,
        hitDie: 20,
        conditions: {
          blinded: false,
          charmed: false,
          deafened: false,
          frightened: true,
          grappled: false,
          incapacitated: false,
          invisible: false,
          paralyzed: false,
          petrified: false,
          poisoned: false,
          prone: false,
          restrained: false,
          stunned: false,
          unconscious: false,
        } as any,
        deathSaves: { successes: 0, failures: 0 },
        resistances: {
          acid: ResistanceType.None,
          bludgeoning: ResistanceType.None,
          cold: ResistanceType.None,
          fire: ResistanceType.Immune,
          force: ResistanceType.Resistant,
          lightning: ResistanceType.None,
          necrotic: ResistanceType.None,
          piercing: ResistanceType.None,
          poison: ResistanceType.None,
          psychic: ResistanceType.None,
          radiant: ResistanceType.None,
          slashing: ResistanceType.None,
          thunder: ResistanceType.None,
        } as any,
        isStable: false,
        isDead: false,
        initiative: 10,
      },
      specialAbilities: {
        assassinate: false,
        berserkThreshold: 0,
        bloodFrenzy: true,
        consumeLifeDie: DiceType.D0,
        corrosiveFormNumDice: 0,
        deathBurstNumDice: 3,
        deathBurstDamageType: DamageType.Fire,
        deathBurstDC: 12,
        deathThroesNumDice: 0,
        deathThroesDC: 0,
        divineEminenceNumDice: 0,
        evasion: false,
        fireAuraNumDice: 0,
        fireForm: false,
        gibbering: false,
        gnomeCunning: false,
        heatedBodyNumDice: 0,
        legendaryResistanceCount: 3,
        lightningAbsorption: false,
        limitedMagicImmunityLevel: 0,
        magicResistance: true,
        magicWeapons: true,
        martialAdvantageNumDice: 0,
        packTactics: true,
        reckless: false,
        reflectiveCarapace: false,
        regenerationValue: 20,
        relentlessThreshold: 0,
        sneakAttackNumDice: 0,
        undeadFortitude: false,
      },
      monsterActions: {
        actions: [
          {
            actionId: 1,
            name: 'Bite',
            attackBonus: 17,
            index: 0,
            rechargeValue: 0,
            hasDC: false,
            damageBlocks: [
              {
                numberOfDice: 2,
                die: DiceType.D10,
                amountToAdd: 10,
                damageType: DamageType.Piercing,
              }
            ],
          },
          {
            actionId: 2,
            name: 'Claw',
            attackBonus: 17,
            index: 1,
            rechargeValue: 0,
            hasDC: false,
            damageBlocks: [
              {
                numberOfDice: 2,
                die: DiceType.D6,
                amountToAdd: 10,
                damageType: DamageType.Slashing,
              }
            ],
          },
          {
            actionId: 3,
            name: 'Fire Breath',
            attackBonus: 0,
            index: 2,
            rechargeValue: 5,
            hasDC: true,
            dc: 24,
            dcAbility: 'constitution',
            damageBlocks: [
              {
                numberOfDice: 26,
                die: DiceType.D6,
                amountToAdd: 0,
                damageType: DamageType.Fire,
              }
            ],
          },
        ],
        multiattacks: [
          [
            { actionId: 1, count: 1 },
            { actionId: 2, count: 2 },
          ],
        ],
        legendaryActions: [
          {
            actionId: 4,
            name: 'Detect',
            attackBonus: 0,
            index: 0,
            rechargeValue: 0,
            hasDC: false,
            description: 'The dragon makes a Wisdom (Perception) check.',
            cost: 1,
            damageBlocks: [
              {
                numberOfDice: 0,
                die: DiceType.D0,
                amountToAdd: 0,
                damageType: DamageType.Bludgeoning,
              }
            ],
          },
          {
            actionId: 6,
            name: 'Tail Attack',
            attackBonus: 17,
            index: 1,
            rechargeValue: 0,
            hasDC: false,
            cost: 1,
            damageBlocks: [
              {
                numberOfDice: 2,
                die: DiceType.D8,
                amountToAdd: 10,
                damageType: DamageType.Bludgeoning,
              }
            ],
          },
          {
            actionId: 6,
            name: 'Wing Attack',
            attackBonus: 17,
            index: 2,
            rechargeValue: 0,
            hasDC: false,
            cost: 2,
            description:
              'The dragon beats its wings. Each creature within 15 feet of the dragon must succeed on a DC 25 Dexterity saving throw or take 17 (2d6 + 10) bludgeoning damage and be knocked prone.',
            damageBlocks: [
              {
                numberOfDice: 2,
                die: DiceType.D6,
                amountToAdd: 10,
                damageType: DamageType.Bludgeoning,
              }
            ],
          },
        ],
        rechargeActions: { 3: 5 },
      },
    };

    this.addCombatant(dummyCharacter);
    this.addCombatant(dummyMonster);
  }
}
