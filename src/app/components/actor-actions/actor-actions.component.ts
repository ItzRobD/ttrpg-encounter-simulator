import { Component, computed, input } from '@angular/core';
import { Actor, Action } from '../../models';
import { CardModule } from 'primeng/card';
import { formatMonsterAction, formatMultiattack } from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-actor-actions',
  standalone: true,
  imports: [CardModule],
  templateUrl: './actor-actions.component.html',
  styleUrl: './actor-actions.component.css',
})
export class ActorActions {
  public readonly actor = input.required<Actor>();

  protected readonly actions = computed(() => {
    return (this.actor().actions || []).filter(a => (a as any).actionType === 'monster_action' || (a as any).action_type === 'monster_action');
  });

  protected readonly multiattacks = computed(() => {
    return (this.actor().actions || []).filter(a => (a as any).actionType === 'monster_multiattack' || (a as any).action_type === 'monster_multiattack')
      .map(a => ({
        name: a.name,
        multiattack: (a as any).multiattack || []
      }));
  });

  protected readonly allActionsForLookup = computed(() => {
    return this.actor().actions || [];
  });

  protected readonly formatMonsterAction = formatMonsterAction;
  protected readonly formatMultiattack = formatMultiattack;

  getRechargeValue(action: Action): number | undefined {
    return (action as any).rechargeValue;
  }
}
