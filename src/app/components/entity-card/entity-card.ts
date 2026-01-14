import { Component, computed, effect, input, signal } from '@angular/core';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import {
  Character,
  Class,
  Monster,
  MonsterSize,
  MonsterType,
  Race,
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
  formatDice,
  formatModifier,
  getAbilityScoreEntries,
  getModifier,
  getSpecialAbilityNames,
} from '../../shared/utils/dnd-utils';
import { TitleCasePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { EntityStats } from '../entity-stats/entity-stats';
import { EntitySpecialAbilities } from '../entity-special-abilities/entity-special-abilities';
import { EntityAttacks } from '../entity-attacks/entity-attacks';
import { EntityEquipment } from '../entity-equipment/entity-equipment';
import { EntityLegendaryActions } from '../entity-legendary-actions/entity-legendary-actions';
import { EntitySpellcasting } from '../entity-spellcasting/entity-spellcasting';
import {CrFormatPipe} from '../../pipes/cr-format.pipe';

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
    EntityStats,
    EntitySpecialAbilities,
    EntityAttacks,
    EntityEquipment,
    EntityLegendaryActions,
    EntitySpellcasting,
    CrFormatPipe,
  ],
  templateUrl: './entity-card.html',
  styleUrl: './entity-card.css',
})
export class EntityCard {
  protected readonly DiceType = DiceType;
  public readonly gradientStop = input<string>('50%');
  public readonly entity = input.required<Entity>();
  public readonly projectedState = input<EntityState>();
  public readonly hideState = input<boolean>(false);
  public readonly startExpanded = input<boolean>(false);

  constructor() {
    effect(() => {
      if (this.startExpanded()) {
        this.mainActivePanel.set('0');
        this.activePanels.set(['0', '1', '2', '3', '4']);
      }
    });
  }

  protected readonly displayState = computed(() => {
    if (this.hideState()) {
      const e = this.entity();
      return {
        conditions: {},
        resistances: e.state?.resistances || {},
        deathSaves: { successes: 0, failures: 0 },
        isStable: false,
        isDead: false,
      } as EntityState;
    }
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

  protected readonly hiddenStateHpString = computed(() => {
    if (!this.hideState()) return '';

    const entity = this.entity();
    if (!entity.hp) {
      return '';
    }

    if (this.isMonster(entity)) {
      return `${entity.hp.hpAverage} ( ${formatDice(entity.hp.numberOfDice, entity.hp.hitDie, entity.hp.amountToAdd)} )`;
    }

    if (this.isCharacter(entity)) {
      return `${entity.hp.value}`;
    }

    return '';
  });

  protected readonly activeConditions = computed(() => {
    if (this.hideState()) return [];
    const state = this.displayState();
    return Object.entries(state.conditions)
      .filter(([_, active]) => active)
      .map(([condition]) => condition)
      .sort();
  });
  protected readonly hpPercent = computed(() => {
    if (this.hideState()) return 0;
    const state = this.displayState();
    if (!state || !state.maxHp) return 0;
    return Math.min(100, Math.floor((state.currentHp / state.maxHp) * 100));
  });
  protected readonly tempHpPercent = computed(() => {
    if (this.hideState()) return 0;
    const state = this.displayState();
    if (!state || state.tempHp <= 0 || !state.maxHp) return 0;
    // We calculate temp HP relative to Max HP to see how much of the bar it should occupy
    return Math.floor((state.tempHp / state.maxHp) * 100);
  });
  protected readonly hpColor = computed(() => {
    if (this.hideState()) return '';
    const percent = this.hpPercent();
    if (percent >= 50) return '#22c55e'; // Green 500
    if (percent >= 25) return '#eab308'; // Yellow 500
    return '#ef4444'; // Red 500
  });

  protected readonly specialAbilityNames = computed(() => {
    const e = this.entity();
    if (!e) return [];
    const abilities = (e as any).specialAbilities;
    if (!abilities) return [];
    return getSpecialAbilityNames(abilities);
  });

  protected readonly actionNames = computed(() => {
    const e = this.entity();
    if (!e) return [];
    if (this.isCharacter(e)) {
      if (!e.equipment?.weapons) return [];
      return Object.entries(e.equipment.weapons)
        .filter(([_, weapon]) => !!weapon)
        .map(([_, weapon]) => weapon!.name);
    } else {
      const monster = e as Monster;
      if (!monster.monsterActions) return [];
      const names = (monster.monsterActions.actions || []).map((a: any) => a.name);
      if (monster.monsterActions.multiattacks && monster.monsterActions.multiattacks.length > 0) {
        names.unshift('Multiattack');
      }
      return names;
    }
  });

  protected readonly legendaryActionNames = computed(() => {
    const e = this.entity();
    if (!e || 'class' in e) return [];
    const monster = e as Monster;
    if (!monster.monsterActions) return [];
    return (monster.monsterActions.legendaryActions || []).map((a: any) => a.name);
  });

  protected readonly activePanels = signal<string | number | string[] | number[] | null | undefined>(
    undefined,
  );

  protected readonly mainActivePanel = signal<string | number | string[] | number[] | null | undefined>(
    undefined,
  );

  private isPanelExpanded(panelValue: string | number): boolean {
    const active = this.activePanels();
    if (!active) return false;
    if (Array.isArray(active)) {
      return (active as (string | number)[]).includes(panelValue.toString()) ||
             (active as (string | number)[]).includes(Number(panelValue));
    }
    return active.toString() === panelValue.toString();
  }

  protected readonly statsHeaderText = computed(() => {
    const e = this.entity();
    if (!e || !e.asConfig?.abilityScores) return { label: 'Statistics', list: '' };

    const isExpanded = this.isPanelExpanded('0');
    if (isExpanded) {
      return { label: 'Statistics', list: '' };
    }

    const order = getAbilityScoreEntries(e.asConfig.abilityScores);
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

  protected readonly spellcastingHeaderText = computed(() => {
    const e = this.entity();
    if (!e) return { label: '', list: '' };
    const sc = e.spellcasting;
    if (!sc || !sc.spells) return { label: '', list: '' };
    const label = 'Spells';
    const names = sc.spells.map((s) => s.name);
    const isExpanded = this.isPanelExpanded('4');
    if (isExpanded) {
      return { label, list: '' };
    }
    return { label, list: ' ' + names.join(' | ') };
  })

  protected readonly getModifier = getModifier;
  protected readonly getAbilityScoreEntries = getAbilityScoreEntries;
  protected readonly formatModifier = formatModifier;
  protected readonly formatDice = formatDice;
}
