import { ChangeDetectionStrategy, Component, inject, input, model, effect } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DialogModule } from 'primeng/dialog';
import { ButtonModule } from 'primeng/button';
import { SharedTable } from '../shared-table/shared-table.component';
import { CombatantService } from '../../services/combatant.service';
import { MessageService } from 'primeng/api';
import { Entity } from '../../models';
import { ToastModule } from 'primeng/toast';
import { MonsterService } from '../../services/monster.service';
import { CharacterService } from '../../services/character.service';

@Component({
  selector: 'app-entity-selector-dialog',
  standalone: true,
  imports: [
    CommonModule,
    DialogModule,
    ButtonModule,
    SharedTable,
    ToastModule
  ],
  providers: [MessageService],
  templateUrl: './entity-selector-dialog.html',
  styles: `
    :host ::ng-deep .p-dialog-content {
      padding: 0;
    }
  `,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class EntitySelectorDialog {
  private readonly combatantService = inject(CombatantService);
  private readonly messageService = inject(MessageService);
  private readonly monsterService = inject(MonsterService);
  private readonly characterService = inject(CharacterService);

  public readonly visible = model.required<boolean>();
  public readonly mode = input.required<'monster' | 'character'>();

  constructor() {
    effect(() => {
      if (this.visible()) {
        const currentMode = this.mode();
        if (currentMode === 'monster') {
          this.monsterService.getSummaries().subscribe();
        } else if (currentMode === 'character') {
          this.characterService.getSummaries().subscribe();
        }
      }
    });
  }

  onAddToSimulator(entity: Entity): void {
    const success = this.combatantService.addToSimulator(entity);
    if (success) {
      this.messageService.add({
        severity: 'success',
        summary: 'Added to Simulator',
        detail: `${entity.name} has been added to the encounter.`,
        life: 2000
      });
    } else {
      this.messageService.add({
        severity: 'warn',
        summary: 'Cannot Add',
        detail: 'The encounter is at maximum capacity.',
        life: 3000
      });
    }
  }

  onHide(): void {
    this.visible.set(false);
  }
}
