import { Component, computed, input, ChangeDetectionStrategy } from '@angular/core';
import { TitleCasePipe } from '@angular/common';
import {
  ActorState,
  ResistanceType,
  Actor
} from '../../models';
import {
  formatModifier,
  getAbilityScoreEntries,
  getDamageTypesByResistance,
} from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-actor-stats',
  standalone: true,
  imports: [TitleCasePipe],
  templateUrl: './actor-stats.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './actor-stats.component.css',
})
export class ActorStats {
  public readonly actor = input.required<Actor>();
  public readonly projectedState = input<ActorState>();
  public readonly gradientStop = input<string>('50%');
  public readonly hideState = input<boolean>(false);

  protected readonly abilities = computed(() => {
    return this.actor().abilities;
  });

  protected readonly displayState = computed(() => {
    const projected = this.projectedState();
    if (this.hideState()) {
      const a = this.actor();
      return {
        conditions: {},
        resistances: a?.state?.resistances || {},
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

    if (projected) {
      return projected;
    }

    // If no projected state, use the initial actor state
    const a = this.actor();
    const state = a?.state;
    const maxHpFromConfig = a?.hpConfig?.value || a?.hpConfig?.hpAverage || 0;
    const maxHp = Math.max(1, Number(state?.maxHp || maxHpFromConfig || 1));

    return {
      ...(state || {}),
      currentHp: Number(state?.currentHp ?? 0),
      maxHp: maxHp,
      tempHp: Number(state?.tempHp ?? 0),
      hitDie: state?.hitDie ?? a?.hpConfig?.hitDie ?? 10,
      conditions: state?.conditions || {},
      deathSaves: state?.deathSaves || { successes: 0, failures: 0 },
      resistances: state?.resistances || {},
      isStable: state?.isStable ?? true,
      isDead: state?.isDead ?? false,
      initiative: state?.initiative ?? 0,
    } as ActorState;
  });

  protected readonly immunities = computed(() => {
    const state = this.displayState();
    if (!state || !state.resistances) return [];
    return getDamageTypesByResistance(state.resistances, ResistanceType.Immune);
  });

  protected readonly resistances = computed(() => {
    const state = this.displayState();
    if (!state || !state.resistances) return [];
    return getDamageTypesByResistance(state.resistances, ResistanceType.Resistant);
  });

  protected readonly vulnerabilities = computed(() => {
    const state = this.displayState();
    if (!state || !state.resistances) return [];
    return getDamageTypesByResistance(state.resistances, ResistanceType.Vulnerable);
  });

  protected readonly showResistances = computed(() => {
    return this.resistances().length > 0 || this.vulnerabilities().length > 0 || this.immunities().length > 0;
  });

  isCharacter(a: Actor): boolean {
    return a.actorType === 'character';
  }

  protected readonly getAbilityScoreEntries = getAbilityScoreEntries;
  protected readonly formatModifier = formatModifier;
}
