import { Component, computed, input } from '@angular/core';
import { Actor } from '../../models';

@Component({
  selector: 'app-actor-special-abilities',
  standalone: true,
  imports: [],
  templateUrl: './actor-special-abilities.component.html',
  styleUrl: './actor-special-abilities.component.css',
})
export class ActorSpecialAbilities {
  public readonly actor = input.required<Actor>();

  protected readonly features = computed(() => {
    const a = this.actor();
    return a.features || [];
  });
}
