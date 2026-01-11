import { Component, computed, input } from '@angular/core';
import { Entity } from '../../models';
import { getFormattedSpecialAbilities } from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-entity-special-abilities',
  standalone: true,
  imports: [],
  templateUrl: './entity-special-abilities.html',
  styleUrl: './entity-special-abilities.css',
})
export class EntitySpecialAbilities {
  public readonly entity = input.required<Entity>();

  protected readonly specialAbilities = computed(() => {
    const e = this.entity();
    if (!e || !('specialAbilities' in e)) return [];
    return getFormattedSpecialAbilities(e.specialAbilities);
  });
}
