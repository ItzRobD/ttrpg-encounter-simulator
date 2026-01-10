import { Component, computed, input, signal } from '@angular/core';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import {
  Character,
  Class,
  Monster,
  MonsterSize,
  MonsterType,
  Race,
  ResistanceType,
  DiceType,
  AbilityScores,
  WeaponSlot,
  DamageType,
} from '../../models';
import { Tag } from 'primeng/tag';
import { Accordion, AccordionContent, AccordionHeader, AccordionPanel } from 'primeng/accordion';
import {
  formatModifier,
  formatWeaponData,
  formatMonsterAction,
  formatMultiattack,
  getAbilityScoreEntries,
  getAbilityScoreShortName,
  getAbilityScoresOrder,
  getDamageTypesByResistance,
  getModifier,
  getFormattedSpecialAbilities,
  getSpecialAbilityNames,
} from '../../shared/utils/dnd-utils';
import { TitleCasePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
@Component({
  selector: 'app-entity-card',
  standalone: true,
  imports: [
    CardModule,
    Tag,
    Accordion,
    AccordionPanel,
    AccordionHeader,
    AccordionContent,
    TitleCasePipe,
    ButtonModule,
    FormsModule,
  ],
  templateUrl: './entity-card.html',
  styleUrl: './entity-card.css',
})
export class EntityCard {
  protected readonly DiceType = DiceType;
  public readonly gradientStop = input<string>('50%');
  protected readonly entity = signal<Character | Monster | undefined>(undefined);
  protected readonly activeConditions = computed(() => {
    const e = this.entity();
    if (!e) return [];
    return Object.entries(e.state.conditions)
      .filter(([_, active]) => active)
      .map(([condition]) => condition)
      .sort();
  });
  protected readonly hpPercent = computed(() => {
    const e = this.entity();
    if (!e) return 0;
    return Math.min(100, Math.floor((e.state.currentHP / e.state.maxHP) * 100));
  });
  protected readonly tempHpPercent = computed(() => {
    const e = this.entity();
    if (!e || e.state.tempHP <= 0) return 0;
    // We calculate temp HP relative to Max HP to see how much of the bar it should occupy
    return Math.floor((e.state.tempHP / e.state.maxHP) * 100);
  });
  protected readonly hpColor = computed(() => {
    const percent = this.hpPercent();
    if (percent > 50) return '#22c55e'; // Green 500
    if (percent > 20) return '#eab308'; // Yellow 500
    return '#ef4444'; // Red 500
  });

  protected readonly immunities = computed(() => {
    const e = this.entity();
    if (!e) return [];
    return getDamageTypesByResistance(e.state.resistances, ResistanceType.Immune);
  });

  protected readonly resistances = computed(() => {
    const e = this.entity();
    if (!e) return [];
    return getDamageTypesByResistance(e.state.resistances, ResistanceType.Resistant);
  });

  protected readonly vulnerabilities = computed(() => {
    const e = this.entity();
    if (!e) return [];
    return getDamageTypesByResistance(e.state.resistances, ResistanceType.Vulnerable);
  });

  protected readonly specialAbilities = computed(() => {
    const e = this.entity();
    if (!e || !('specialAbilities' in e)) return [];
    return getFormattedSpecialAbilities(e.specialAbilities);
  });

  protected readonly specialAbilityNames = computed(() => {
    const e = this.entity();
    if (!e || !('specialAbilities' in e)) return [];
    return getSpecialAbilityNames(e.specialAbilities);
  });

  protected readonly actionNames = computed(() => {
    const e = this.entity();
    if (!e) return [];
    if ('class' in e) {
      return Object.entries(e.equipment.weapons)
        .filter(([_, weapon]) => !!weapon)
        .map(([_, weapon]) => weapon!.name);
    } else {
      const names = e.monsterActions.actions.map((a) => a.name);
      if (e.monsterActions.multiattacks.length > 0) {
        names.unshift('Multiattack');
      }
      return names;
    }
  });

  protected readonly legendaryActionNames = computed(() => {
    const e = this.entity();
    if (!e || 'class' in e) return [];
    return e.monsterActions.legendaryActions.map((a) => a.name);
  });

  protected readonly saActiveValue = signal<string | number | string[] | number[] | null | undefined>(
    undefined,
  );

  private isPanelExpanded(panelValue: string | number): boolean {
    const active = this.saActiveValue();
    if (!active) return false;
    if (Array.isArray(active)) {
      return (active as (string | number)[]).includes(panelValue.toString()) ||
             (active as (string | number)[]).includes(Number(panelValue));
    }
    return active.toString() === panelValue.toString();
  }

  protected readonly saHeaderText = computed(() => {
    const names = this.specialAbilityNames();
    const isExpanded = this.isPanelExpanded('2');

    if (isExpanded || names.length === 0) {
      return { label: 'Special Abilities', list: '' };
    }
    return { label: 'Special Abilities', list: ' ' + names.join(', ') };
  });

  protected readonly actionsHeaderText = computed(() => {
    const e = this.entity();
    if (!e) return { label: '', list: '' };
    const label = 'class' in e ? 'Equipment' : 'Actions';
    const names = this.actionNames();
    const isExpanded = this.isPanelExpanded('1');

    if (isExpanded || names.length === 0) {
      return { label, list: '' };
    }
    return { label, list: ' ' + names.join(', ') };
  });

  protected readonly legendaryActionsHeaderText = computed(() => {
    const names = this.legendaryActionNames();
    const isExpanded = this.isPanelExpanded('3');

    if (isExpanded || names.length === 0) {
      return { label: 'Legendary Actions', list: '' };
    }
    return { label: 'Legendary Actions', list: ' ' + names.join(', ') };
  });

  constructor() {
    const dummyCharacter: Character = {
      id: 1,
      instanceId: 101,
      name: 'Valerius the Brave',
      race: Race.Human,
      class: Class.Paladin,
      level: 5,
      abilityScores: {
        strength: 16,
        dexterity: 12,
        constitution: 14,
        intelligence: 10,
        wisdom: 12,
        charisma: 14,
      },
      abilityScoreProficiency: {
        strength: true,
        dexterity: false,
        constitution: true,
        intelligence: false,
        wisdom: false,
        charisma: false,
      },
      state: {
        currentHP: 40,
        maxHP: 63,
        tempHP: 12,
        hitDie: 10,
        conditions: {
          blinded: false,
          charmed: true,
          deafened: false,
          frightened: false,
          grappled: false,
          incapacitated: false,
          invisible: false,
          paralyzed: false,
          petrified: false,
          poisoned: true,
          prone: true,
          restrained: false,
          stunned: false,
          unconscious: false,
        },
        deathSaves: { successes: 2, failures: 1 },
        resistances: {
          acid: ResistanceType.Immune,
          bludgeoning: ResistanceType.Resistant,
          cold: ResistanceType.None,
          fire: ResistanceType.Vulnerable,
          force: ResistanceType.None,
          lightning: ResistanceType.None,
          necrotic: ResistanceType.None,
          piercing: ResistanceType.Resistant,
          poison: ResistanceType.None,
          psychic: ResistanceType.None,
          radiant: ResistanceType.None,
          slashing: ResistanceType.Resistant,
          thunder: ResistanceType.None,
        },
        isStable: true,
        isDead: false,
        initiative: 14,
      },
      equipment: {
        armor: { id: 1, name: 'Plate', ac: 18, minimumStrength: 15 },
        shield: { id: 2, name: 'Shield', ac: 2, minimumStrength: 0 },
        hasShieldEquipped: true,
        weapons: {
          [WeaponSlot.Primary]: {
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
          [WeaponSlot.Secondary]: {
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
          [WeaponSlot.Ranged]: {
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
        },
      },
    };
    const dummyMonster: Monster = {
      id: 2,
      instanceId: 201,
      name: 'Ancient Red Dragon',
      size: MonsterSize.Gargantuan,
      type: MonsterType.Dragon,
      cr: 24,
      proficiencyBonus: 7,
      isInnateSpellcaster: false,
      isLegendary: true,
      isSpellcaster: true,
      abilityScores: {
        strength: 30,
        dexterity: 10,
        constitution: 29,
        intelligence: 18,
        wisdom: 15,
        charisma: 23,
      },
      abilityScoreProficiency: {
        strength: false,
        dexterity: true,
        constitution: true,
        intelligence: false,
        wisdom: true,
        charisma: true,
      },
      state: {
        currentHP: 546,
        maxHP: 546,
        tempHP: 0,
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
        },
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
        },
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
            description: 'The dragon makes a Wisdom (Perception) check.'
          },
          {
            actionId: 5,
            name: 'Tail Attack',
            numberOfDice: 2,
            die: DiceType.D8,
            amountToAdd: 10,
            attackBonus: 17,
            damageType: DamageType.Bludgeoning,
            index: 1,
            rechargeValue: 0,
            hasDC: false,
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
            description: 'The dragon beats its wings. Each creature within 15 feet of the dragon must succeed on a DC 25 Dexterity saving throw or take 17 (2d6 + 10) bludgeoning damage and be knocked prone.'
          }
        ],
        rechargeActions: { 3: 5 },
      },
    };
    // this.entity.set(dummyMonster);
    this.entity.set(dummyCharacter);
  }

  protected readonly getModifier = getModifier;
  protected readonly getAbilityScoresOrder = getAbilityScoresOrder;
  protected readonly getAbilityScoreShortName = getAbilityScoreShortName;
  protected readonly getAbilityScoreEntries = getAbilityScoreEntries;
  protected readonly formatModifier = formatModifier;
  protected readonly getDamageTypesByResistance = getDamageTypesByResistance;
  protected readonly ResistanceType = ResistanceType;
  protected readonly formatWeaponData = formatWeaponData;
  protected readonly formatMonsterAction = formatMonsterAction;
  protected readonly formatMultiattack = formatMultiattack;
  protected readonly weaponSlots: WeaponSlot[] = [
    WeaponSlot.Primary,
    WeaponSlot.Secondary,
    WeaponSlot.Ranged,
  ];
}
