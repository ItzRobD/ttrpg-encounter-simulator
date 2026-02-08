import { Component, computed, input } from '@angular/core';
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
    const a = this.actor();
    if (this.hideState()) {
      return {
        conditions: {},
        resistances: a?.state?.resistances || {},
        deathSaves: { successes: 0, failures: 0 },
        isStable: false,
        isDead: false,
      } as ActorState;
    }
    return this.projectedState() || a?.state;
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
