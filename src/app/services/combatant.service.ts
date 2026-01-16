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
    this._combatants().filter((e): e is Monster => !('class' in e))
  );

  readonly characters = computed(() =>
    this._combatants().filter((e): e is Character => 'class' in e)
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
    const isCharacter = 'class' in entity;
    const currentCharacters = this.characters().length;

    // Check that we're below the max entities
    if (this.count() >= environment.limits.maxTotal) return false;

    // Check if we're already at the max number of characters
    if (isCharacter && currentCharacters >= environment.limits.maxCharacters) return false;

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
      race: Race.Human,
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
        armor: { id: 0, name: 'None', ac: 10, minimumStrength: 0, dexBonus: 10, maxBonus: 10, modifier: 0 },
        shield: undefined,
        hasShieldEquipped: false,
        weapons: {
          [WeaponSlot.Primary]: [
            {
              name: 'Longsword',
              numberOfDice: 1,
              die: DiceType.D10,
              damageType: DamageType.Slashing,
              properties: {
                isVersatile: true,
                isFinesse: false,
                isRanged: false,
                isHeavy: false,
                isLight: false,
                isTwoHanded: false,
                isThrown: false,
                isOnlyRanged: false,
              },
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
        },
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
            numberOfDice: 1,
            die: DiceType.D4,
            amountToAdd: 0,
            attackBonus: 2,
            damageType: DamageType.Bludgeoning,
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
      race: Race.Human,
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
        armor: { id: 1, name: 'Plate', ac: 18, minimumStrength: 15, dexBonus: 0, maxBonus: 0, modifier: 0 },
        shield: { id: 2, name: 'Shield', ac: 2, minimumStrength: 0, dexBonus: 0, maxBonus: 0, modifier: 0 },
        hasShieldEquipped: true,
        weapons: {
          [WeaponSlot.Primary]: [
            {
              name: 'Longsword',
              numberOfDice: 1,
              die: DiceType.D8,
              damageType: DamageType.Slashing,
              properties: {
                isVersatile: true,
                isFinesse: false,
                isRanged: false,
                isHeavy: false,
                isTwoHanded: false,
                isLight: false,
                isThrown: false,
                isOnlyRanged: false,
              },
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
          [WeaponSlot.Secondary]: [
            {
              name: 'Shortsword',
              numberOfDice: 1,
              die: DiceType.D6,
              damageType: DamageType.Piercing,
              properties: {
                isVersatile: false,
                isFinesse: true,
                isRanged: false,
                isHeavy: false,
                isTwoHanded: false,
                isLight: true,
                isThrown: false,
                isOnlyRanged: false,
              },
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
          [WeaponSlot.Ranged]: [
            {
              name: 'Longbow',
              numberOfDice: 1,
              die: DiceType.D8,
              damageType: DamageType.Piercing,
              properties: {
                isVersatile: false,
                isFinesse: false,
                isRanged: true,
                isHeavy: true,
                isTwoHanded: true,
                isLight: false,
                isThrown: false,
                isOnlyRanged: true,
              },
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
        },
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
            numberOfDice: 2,
            die: DiceType.D10,
            amountToAdd: 10,
            attackBonus: 17,
            damageType: DamageType.Piercing,
            index: 0,
            rechargeValue: 0,
            hasDC: false,
          },
          {
            actionId: 2,
            name: 'Claw',
            numberOfDice: 2,
            die: DiceType.D6,
            amountToAdd: 10,
            attackBonus: 17,
            damageType: DamageType.Slashing,
            index: 1,
            rechargeValue: 0,
            hasDC: false,
          },
          {
            actionId: 3,
            name: 'Fire Breath',
            numberOfDice: 26,
            die: DiceType.D6,
            amountToAdd: 0,
            attackBonus: 0,
            damageType: DamageType.Fire,
            index: 2,
            rechargeValue: 5,
            hasDC: true,
            dc: 24,
            dcAbility: 'Constitution',
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
            numberOfDice: 0,
            die: DiceType.D0,
            amountToAdd: 0,
            attackBonus: 0,
            damageType: DamageType.Bludgeoning,
            index: 0,
            rechargeValue: 0,
            hasDC: false,
            description: 'The dragon makes a Wisdom (Perception) check.',
            cost: 1,
          },
          {
            actionId: 6,
            name: 'Tail Attack',
            numberOfDice: 2,
            die: DiceType.D8,
            amountToAdd: 10,
            attackBonus: 17,
            damageType: DamageType.Bludgeoning,
            index: 1,
            rechargeValue: 0,
            hasDC: false,
            cost: 1,
          },
          {
            actionId: 6,
            name: 'Wing Attack',
            numberOfDice: 2,
            die: DiceType.D6,
            amountToAdd: 10,
            attackBonus: 17,
            damageType: DamageType.Bludgeoning,
            index: 2,
            rechargeValue: 0,
            hasDC: false,
            cost: 2,
            description:
              'The dragon beats its wings. Each creature within 15 feet of the dragon must succeed on a DC 25 Dexterity saving throw or take 17 (2d6 + 10) bludgeoning damage and be knocked prone.',
          },
        ],
        rechargeActions: { 3: 5 },
      },
    };

    this.addCombatant(dummyCharacter);
    this.addCombatant(dummyMonster);
  }
}
