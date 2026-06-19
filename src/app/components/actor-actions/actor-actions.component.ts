import { Component, computed, input, ChangeDetectionStrategy } from '@angular/core';
import { Actor, Action } from '../../models';
import { CardModule } from 'primeng/card';
import { formatMonsterAction, formatMultiattack } from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-actor-actions',
  standalone: true,
  imports: [CardModule],
  templateUrl: './actor-actions.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './actor-actions.component.css',
})
export class ActorActions {
  public readonly actor = input.required<Actor>();

  protected readonly actions = computed(() => {
    return (this.actor().actions || []).filter(a => a.actionType === 'monster_action');
  });

  protected readonly multiattacks = computed(() => {
    return (this.actor().actions || []).filter(a => a.actionType === 'monster_multiattack')
      .map(a => ({
        name: a.name,
        multiattack: a.multiattack || []
      }));
  });

  protected readonly allActionsForLookup = computed(() => {
    return this.actor().actions || [];
  });

  protected readonly formatMonsterAction = formatMonsterAction;
  protected readonly formatMultiattack = formatMultiattack;

  getRechargeValue(action: Action): number | undefined {
    return action.rechargeValue;
  }
}
