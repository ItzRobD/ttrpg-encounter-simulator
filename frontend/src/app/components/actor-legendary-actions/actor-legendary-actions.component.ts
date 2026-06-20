import { Component, computed, input, ChangeDetectionStrategy } from '@angular/core';
import { Actor } from '../../models';
import { CardModule } from 'primeng/card';
import { formatMonsterAction } from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-actor-legendary-actions',
  imports: [CardModule],
  templateUrl: './actor-legendary-actions.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './actor-legendary-actions.component.scss',
})
export class ActorLegendaryActions {
  public readonly actor = input.required<Actor>();

  protected readonly legendaryActions = computed(() => {
    return (this.actor().actions || []).filter(a => a.actionType === 'monster_legendary');
  });

  protected readonly formatMonsterAction = formatMonsterAction;
}
