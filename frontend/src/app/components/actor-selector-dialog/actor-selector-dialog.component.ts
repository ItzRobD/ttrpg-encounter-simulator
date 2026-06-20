import { ChangeDetectionStrategy, Component, inject, input, model, effect } from '@angular/core';
import { DialogModule } from 'primeng/dialog';
import { SharedTable } from '../shared-table/shared-table.component';
import { CombatantService } from '../../services/combatant.service';
import { MessageService } from 'primeng/api';
import { Actor } from '../../models';
import { ToastModule } from 'primeng/toast';
import { MonsterService } from '../../services/monster.service';
import { CharacterService } from '../../services/character.service';

@Component({
  selector: 'app-actor-selector-dialog',
  imports: [
    DialogModule,
    SharedTable,
    ToastModule
  ],
  providers: [MessageService],
  templateUrl: './actor-selector-dialog.component.html',
  styleUrl: './actor-selector-dialog.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ActorSelectorDialog {
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

  onAddToSimulator(actor: Actor): void {
    const success = this.combatantService.addToSimulator(actor);
    if (success) {
      this.messageService.add({
        severity: 'success',
        summary: 'Added to Simulator',
        detail: `${actor.name} has been added to the encounter.`,
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
