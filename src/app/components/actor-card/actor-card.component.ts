import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { MapperService } from '../../services/mapper.service';
import {
  DiceType,
  WeaponSlotData,
  ActorState, getAC, isCharacter, isMonster,
  Actor
} from '../../models';
import { Tag } from 'primeng/tag';
import { Accordion, AccordionContent, AccordionHeader, AccordionPanel } from 'primeng/accordion';
import {
  formatDice,
  formatModifier,
  getAbilityScoreEntries,
  getModifier,
} from '../../shared/utils/dnd-utils';
import { TitleCasePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActorStats } from '../actor-stats/actor-stats.component';
import { ActorSpecialAbilities } from '../actor-special-abilities/actor-special-abilities.component';
import { ActorActions } from '../actor-actions/actor-actions.component';
import { ActorEquipment } from '../actor-equipment/actor-equipment.component';
import { ActorLegendaryActions } from '../actor-legendary-actions/actor-legendary-actions.component';
import { ActorSpellcasting } from '../actor-spellcasting/actor-spellcasting.component';
import {CrFormatPipe} from '../../pipes/cr-format.pipe';
import { EquipmentService } from '../../services/equipment.service';

@Component({
  selector: 'app-actor-card',
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
    ActorStats,
    ActorSpecialAbilities,
    ActorActions,
    ActorEquipment,
    ActorLegendaryActions,
    ActorSpellcasting,
    CrFormatPipe,
  ],
  templateUrl: './actor-card.component.html',
  styleUrl: './actor-card.component.css',
})
export class ActorCard {
  private readonly equipmentService = inject(EquipmentService);
  protected readonly mapperService = inject(MapperService);
  protected readonly DiceType = DiceType;
  public readonly gradientStop = input<string>('50%');
  public readonly actor = input.required<Actor>();
  public readonly projectedState = input<ActorState>();
  public readonly hideState = input<boolean>(false);
  public readonly startExpanded = input<boolean>(false);
  public readonly showDelete = input<boolean>(false);

  public readonly delete = output<number>();

  constructor() {
    effect(() => {
      if (this.startExpanded()) {
        this.mainActivePanel.set('0');
        this.activePanels.set(['0', '1', '2', '3', '4']);
      }
    });
  }

  protected readonly displayState = computed(() => {
    const projected = this.projectedState();
    if (this.hideState()) {
      const a = this.actor();
      return {
        conditions: {},
        resistances: a.state?.resistances || {},
        deathSaves: { successes: 0, failures: 0 },
        isStable: false,
        isDead: false,
        currentHp: 0,
        maxHp: 0,
        tempHp: 0,
        hitDie: 0,
        initiative: 0
      } as ActorState;
    }
    const state = projected || this.actor().state;

    // Fallback to ensure we never return undefined
    if (!state || state.maxHp === 0 || isNaN(state.maxHp)) {
      const a = this.actor();
      const fallbackMaxHp = state?.maxHp || a.state?.maxHp || a.hpConfig?.hpAverage || a.hpConfig?.value || a.ac || (a as any).AC || 0;
      return {
        ...(state || a.state),
        conditions: state?.conditions || a.state?.conditions || {},
        resistances: state?.resistances || a.state?.resistances || {},
        deathSaves: state?.deathSaves || a.state?.deathSaves || { successes: 0, failures: 0 },
        isStable: state?.isStable ?? a.state?.isStable ?? true,
        isDead: state?.isDead ?? a.state?.isDead ?? false,
        currentHp: state?.currentHp ?? a.state?.currentHp ?? 0,
        maxHp: fallbackMaxHp,
        tempHp: state?.tempHp ?? a.state?.tempHp ?? 0,
      } as ActorState;
    }

    return state;
  });

  isCharacter(a: Actor): boolean {
    return isCharacter(a);
  }

  isMonster(a: Actor): boolean {
    return isMonster(a);
  }

  protected readonly hiddenStateHpString = computed(() => {
    if (!this.hideState()) return '';

    const actor = this.actor();
    if (!actor.hpConfig) {
      return '';
    }

    if (this.isMonster(actor)) {
      const hp = actor.hpConfig;
      return `${hp.hpAverage} ( ${formatDice(hp.numberOfDice || 0, (hp.hitDie || 0) as DiceType, hp.amountToAdd)} )`;
    }

    if (this.isCharacter(actor)) {
      const hp = actor.hpConfig;
      return `${hp.value}`;
    }

    return '';
  });

  protected readonly activeConditions = computed(() => {
    if (this.hideState()) return [];
    const state = this.displayState();
    return Object.entries(state.conditions || {})
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
    if (!state || (state.tempHp || 0) <= 0 || !state.maxHp) return 0;
    // We calculate temp HP relative to Max HP to see how much of the bar it should occupy
    return Math.floor(((state.tempHp || 0) / state.maxHp) * 100);
  });
  protected readonly hpColor = computed(() => {
    if (this.hideState()) return '';
    const percent = this.hpPercent();
    if (percent >= 50) return '#22c55e'; // Green 500
    if (percent >= 25) return '#eab308'; // Yellow 500
    return '#ef4444'; // Red 500
  });

  protected readonly specialAbilityNames = computed(() => {
    const a = this.actor();
    return (a.features || []).map(f => f.name);
  });

  protected readonly actionNames = computed(() => {
    const a = this.actor();
    return (a.actions || [])
      .filter(a => (a as any).actionType === 'monster_action' || (a as any).action_type === 'monster_action' || (a as any).actionType === 'monster_multiattack' || (a as any).action_type === 'monster_multiattack')
      .map(act => act.name);
  });

  protected readonly weaponNames = computed(() => {
    const a = this.actor();
    if (!this.isCharacter(a)) return [];
    const character = a;
    const eq = character.equipment;
    if (!eq) return [];

    const weaponIds = [
      ...(eq.primarySlot || []).map((w: WeaponSlotData) => w.weaponId),
      ...(eq.secondarySlot || []).map((w: WeaponSlotData) => w.weaponId),
      ...(eq.rangedSlot || []).map((w: WeaponSlotData) => w.weaponId),
    ];

    return weaponIds.map(id => {
      const summary = this.equipmentService.summaries().find((s) => s.id.toString() === id.toString() && s.type === 'Weapon');
      return summary ? summary.name : `Weapon #${id}`;
    });
  });

  protected readonly legendaryActionNames = computed(() => {
    const a = this.actor();
    return (a.actions || [])
      .filter(act => (act as any).actionType === 'monster_legendary' || (act as any).action_type === 'monster_legendary')
      .map(act => act.name);
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
    const a = this.actor();
    const abilities = a.abilities;
    if (!a || !abilities?.abilityScores) return { label: 'Statistics', list: '' };

    const isExpanded = this.isPanelExpanded('0');
    if (isExpanded) {
      return { label: 'Statistics', list: '' };
    }

    const order = getAbilityScoreEntries(abilities.abilityScores);
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
    const names = this.actionNames();
    const isExpanded = this.isPanelExpanded('1');
    const label = 'Actions';

    if (isExpanded || names.length === 0) {
      return { label, list: '' };
    }
    return { label, list: ' ' + names.join(' | ') };
  });

  protected readonly equipmentHeaderText = computed(() => {
    const names = this.weaponNames();
    const isExpanded = this.isPanelExpanded('5');
    const label = 'Equipment';

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
    const a = this.actor();
    if (!a) return { label: '', list: '' };
    const sc = a.spellcasting;
    if (!sc) return { label: '', list: '' };

    let spells: any[] = [];
    if (sc.leveledSpells || sc.innateSpells) {
      spells = [...(sc.leveledSpells || []), ...(sc.innateSpells || [])];
    } else if ((sc as any).spells) {
      spells = (sc as any).spells;
    }

    if (spells.length === 0) return { label: '', list: '' };

    const label = 'Spells';
    const names = spells.map((s) => s.name);
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
  protected readonly getAC = getAC;
}
