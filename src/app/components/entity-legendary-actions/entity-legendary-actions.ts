import { Component, input } from '@angular/core';
import { DiceType, Monster } from '../../models';
import { CardModule } from 'primeng/card';
import { formatMonsterAction } from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-entity-legendary-actions',
  standalone: true,
  imports: [CardModule],
  templateUrl: './entity-legendary-actions.html',
  styleUrl: './entity-legendary-actions.css',
})
export class EntityLegendaryActions {
  public readonly monster = input.required<Monster>();

  protected readonly DiceType = DiceType;
  protected readonly formatMonsterAction = formatMonsterAction;
}
