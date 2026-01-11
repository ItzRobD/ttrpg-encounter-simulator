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
  Entity,
  EntityState,
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
  public readonly entity = input.required<Entity>();
  public readonly projectedState = input<EntityState>();

  protected readonly displayState = computed(() => {
    return this.projectedState() || this.entity().state;
  });

  isCharacter(e: Entity): e is Character {
    return 'class' in e;
  }

  isMonster(e: Entity): e is Monster {
    return !('class' in e);
  }

  asCharacter(e: Entity): Character {
    return e as Character;
  }

  asMonster(e: Entity): Monster {
    return e as Monster;
  }

  protected readonly activeConditions = computed(() => {
    const state = this.displayState();
    return Object.entries(state.conditions)
      .filter(([_, active]) => active)
      .map(([condition]) => condition)
      .sort();
  });
  protected readonly hpPercent = computed(() => {
    const state = this.displayState();
    if (!state || !state.maxHp) return 0;
    return Math.min(100, Math.floor((state.currentHp / state.maxHp) * 100));
  });
  protected readonly tempHpPercent = computed(() => {
    const state = this.displayState();
    if (!state || state.tempHp <= 0 || !state.maxHp) return 0;
    // We calculate temp HP relative to Max HP to see how much of the bar it should occupy
    return Math.floor((state.tempHp / state.maxHp) * 100);
  });
  protected readonly hpColor = computed(() => {
    const percent = this.hpPercent();
    if (percent >= 50) return '#22c55e'; // Green 500
    if (percent >= 25) return '#eab308'; // Yellow 500
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
      const char = e as Character;
      return Object.entries(char.equipment.weapons)
        .filter(([_, weapon]) => !!weapon)
        .map(([_, weapon]) => weapon!.name);
    } else {
      const monster = e as Monster;
      const names = monster.monsterActions.actions.map((a: any) => a.name);
      if (monster.monsterActions.multiattacks.length > 0) {
        names.unshift('Multiattack');
      }
      return names;
    }
  });

  protected readonly legendaryActionNames = computed(() => {
    const e = this.entity();
    if (!e || 'class' in e) return [];
    const monster = e as Monster;
    return monster.monsterActions.legendaryActions.map((a: any) => a.name);
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

  protected readonly statsHeaderText = computed(() => {
    const e = this.entity();
    if (!e) return { label: 'Statistics', list: '' };

    const isExpanded = this.isPanelExpanded('0');
    if (isExpanded) {
      return { label: 'Statistics', list: '' };
    }

    const order = getAbilityScoreEntries(e.abilityScores);
    const list = ' ' + order.map((entry) => `${entry.shortName.toUpperCase()} ${entry.value}`).join(' | ');

    return { label: 'Statistics', list };
  });

  protected readonly saHeaderText = computed(() => {
    const names = this.specialAbilityNames();
    const isExpanded = this.isPanelExpanded('2');

    if (isExpanded || names.length === 0) {
      return { label: 'Special Abilities', list: '' };
    }
    return { label: 'Special Abilities', list: ' [' + names.join('] [') + ']' };
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
    return { label, list: ' ' + names.join(' | ') };
  });

  protected readonly legendaryActionsHeaderText = computed(() => {
    const names = this.legendaryActionNames();
    const isExpanded = this.isPanelExpanded('3');

    if (isExpanded || names.length === 0) {
      return { label: 'Legendary Actions', list: '' };
    }
    return { label: 'Legendary Actions', list: ' ' + names.join(' | ') };
  });

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
