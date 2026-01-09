import { Component, computed, signal } from '@angular/core';
import { CardModule } from 'primeng/card';
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
} from '../../models';
import { Tag } from 'primeng/tag';
import { Accordion, AccordionContent, AccordionHeader, AccordionPanel } from 'primeng/accordion';
import {
  formatModifier,
  getAbilityScoreEntries,
  getAbilityScoreShortName,
  getAbilityScoresOrder,
  getModifier,
} from '../../shared/utils/dnd-utils';
import { KeyValuePipe } from '@angular/common';
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
    KeyValuePipe,
  ],
  templateUrl: './entity-card.html',
  styleUrl: './entity-card.css',
})
export class EntityCard {
  protected readonly entity = signal<Character | Monster | undefined>(undefined);
  protected readonly activeConditions = computed(() => {
    const e = this.entity();
    if (!e) return [];
    return Object.entries(e.state.conditions)
      .filter(([_, active]) => active)
      .map(([condition]) => condition)
      .sort();
  });
  protected expanded = false;
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
        currentHP: 45,
        maxHP: 50,
        tempHP: 5,
        hitDie: 10,
        conditions: {
          blinded: false,
          charmed: false,
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
        deathSaves: { successes: 0, failures: 0 },
        resistances: {
          acid: ResistanceType.None,
          bludgeoning: ResistanceType.None,
          cold: ResistanceType.None,
          fire: ResistanceType.Resistant,
          force: ResistanceType.None,
          lightning: ResistanceType.None,
          necrotic: ResistanceType.None,
          piercing: ResistanceType.None,
          poison: ResistanceType.None,
          psychic: ResistanceType.None,
          radiant: ResistanceType.None,
          slashing: ResistanceType.None,
          thunder: ResistanceType.None,
        },
        isStable: true,
        isDead: false,
        initiative: 14,
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
          frightened: false,
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
          force: ResistanceType.None,
          lightning: ResistanceType.None,
          necrotic: ResistanceType.None,
          piercing: ResistanceType.None,
          poison: ResistanceType.None,
          psychic: ResistanceType.None,
          radiant: ResistanceType.None,
          slashing: ResistanceType.None,
          thunder: ResistanceType.None,
        },
        isStable: true,
        isDead: false,
        initiative: 10,
      },
      specialAbilities: {
        assassinate: false,
        berserkThreshold: 0,
        bloodFrenzy: false,
        consumeLifeDie: DiceType.D0,
        corrosiveFormNumDice: 0,
        deathBurstNumDice: 0,
        deathBurstDamageType: 'fire' as any,
        deathBurstDC: 0,
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
        magicResistance: false,
        magicWeapons: true,
        martialAdvantageNumDice: 0,
        packTactics: false,
        reckless: false,
        reflectiveCarapace: false,
        regenerationValue: 0,
        relentlessThreshold: 0,
        sneakAttackNumDice: 0,
        undeadFortitude: false,
      },
      monsterActions: { actions: [], multiattacks: [], legendaryActions: [], rechargeActions: {} },
    };
    // this.entity.set(dummyMonster);
    this.entity.set(dummyCharacter);
  }

  protected readonly getModifier = getModifier;
  protected readonly getAbilityScoresOrder = getAbilityScoresOrder;
  protected readonly getAbilityScoreShortName = getAbilityScoreShortName;
  protected readonly getAbilityScoreEntries = getAbilityScoreEntries;
  protected readonly formatModifier = formatModifier;
}
