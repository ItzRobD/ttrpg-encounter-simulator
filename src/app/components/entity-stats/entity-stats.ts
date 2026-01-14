import { Component, computed, input } from '@angular/core';
import { TitleCasePipe } from '@angular/common';
import {
  Entity,
  EntityState,
  ResistanceType,
} from '../../models';
import {
  formatModifier,
  getAbilityScoreEntries,
  getDamageTypesByResistance,
} from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-entity-stats',
  standalone: true,
  imports: [TitleCasePipe],
  templateUrl: './entity-stats.html',
  styleUrl: './entity-stats.css',
})
export class EntityStats {
  public readonly entity = input.required<Entity>();
  public readonly projectedState = input<EntityState>();
  public readonly gradientStop = input<string>('50%');
  public readonly hideState = input<boolean>(false);

  protected readonly displayState = computed(() => {
    const e = this.entity();
    if (this.hideState()) {
      return {
        conditions: {},
        resistances: e?.state?.resistances || {},
        deathSaves: { successes: 0, failures: 0 },
        isStable: false,
        isDead: false,
      } as EntityState;
    }
    return this.projectedState() || e?.state;
  });

  protected readonly immunities = computed(() => {
    const e = this.entity();
    const state = this.displayState();
    if (!e || !state || !state.resistances) return [];
    return getDamageTypesByResistance(state.resistances, ResistanceType.Immune);
  });

  protected readonly resistances = computed(() => {
    const e = this.entity();
    const state = this.displayState();
    if (!e || !state || !state.resistances) return [];
    return getDamageTypesByResistance(state.resistances, ResistanceType.Resistant);
  });

  protected readonly vulnerabilities = computed(() => {
    const e = this.entity();
    const state = this.displayState();
    if (!e || !state || !state.resistances) return [];
    return getDamageTypesByResistance(state.resistances, ResistanceType.Vulnerable);
  });

  protected readonly showResistances = computed(() => {
    return this.resistances().length > 0 || this.vulnerabilities().length > 0 || this.immunities().length > 0;
  });

  isCharacter(e: Entity): boolean {
    return 'class' in e;
  }

  protected readonly getAbilityScoreEntries = getAbilityScoreEntries;
  protected readonly formatModifier = formatModifier;
}
