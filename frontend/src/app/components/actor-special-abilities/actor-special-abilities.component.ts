import { Component, computed, input, ChangeDetectionStrategy } from '@angular/core';
import { Actor } from '../../models';

@Component({
  selector: 'app-actor-special-abilities',
  imports: [],
  templateUrl: './actor-special-abilities.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './actor-special-abilities.component.scss',
})
export class ActorSpecialAbilities {
  public readonly actor = input.required<Actor>();

  protected readonly features = computed(() => {
    const a = this.actor();
    return a.features || [];
  });
}
