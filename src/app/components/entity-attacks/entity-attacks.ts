import { Component, input } from '@angular/core';
import { Monster } from '../../models';
import { CardModule } from 'primeng/card';
import { formatMonsterAction, formatMultiattack } from '../../shared/utils/dnd-utils';

@Component({
  selector: 'app-entity-attacks',
  standalone: true,
  imports: [CardModule],
  templateUrl: './entity-attacks.html',
  styleUrl: './entity-attacks.css',
})
export class EntityAttacks {
  public readonly monster = input.required<Monster>();

  protected readonly formatMonsterAction = formatMonsterAction;
  protected readonly formatMultiattack = formatMultiattack;
}
