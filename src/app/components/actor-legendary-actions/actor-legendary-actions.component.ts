import { Component, computed, input } from '@angular/core';
import { Actor } from '../../models';
import { CardModule } from 'primeng/card';
import { formatMonsterAction } from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-actor-legendary-actions',
  standalone: true,
  imports: [CardModule],
  templateUrl: './actor-legendary-actions.component.html',
  styleUrl: './actor-legendary-actions.component.css',
})
export class ActorLegendaryActions {
  public readonly actor = input.required<Actor>();

  protected readonly legendaryActions = computed(() => {
    return (this.actor().actions || []).filter(a => (a as any).actionType === 'monster_legendary' || (a as any).action_type === 'monster_legendary');
  });

  protected readonly formatMonsterAction = formatMonsterAction;
}
