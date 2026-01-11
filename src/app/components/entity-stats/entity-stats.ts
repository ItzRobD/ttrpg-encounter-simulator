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

  protected readonly displayState = computed(() => {
    return this.projectedState() || this.entity().state;
  });

  protected readonly immunities = computed(() => {
    const e = this.entity();
    if (!e) return [];
    return getDamageTypesByResistance(this.displayState().resistances, ResistanceType.Immune);
  });

  protected readonly resistances = computed(() => {
    const e = this.entity();
    if (!e) return [];
    return getDamageTypesByResistance(this.displayState().resistances, ResistanceType.Resistant);
  });

  protected readonly vulnerabilities = computed(() => {
    const e = this.entity();
    if (!e) return [];
    return getDamageTypesByResistance(this.displayState().resistances, ResistanceType.Vulnerable);
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
