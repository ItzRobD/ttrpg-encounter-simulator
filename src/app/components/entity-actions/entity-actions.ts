import { Component, input } from '@angular/core';
import { Monster } from '../../models';
import { CardModule } from 'primeng/card';
import { formatMonsterAction, formatMultiattack } from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-entity-actions',
  standalone: true,
  imports: [CardModule],
  templateUrl: './entity-actions.html',
  styleUrl: './entity-actions.css',
})
export class EntityActions {
  public readonly monster = input.required<Monster>();

  protected readonly formatMonsterAction = formatMonsterAction;
  protected readonly formatMultiattack = formatMultiattack;
}
