import { Component, ChangeDetectionStrategy, inject, OnInit, signal } from '@angular/core';
import { ButtonModule } from 'primeng/button';
import { TooltipModule } from 'primeng/tooltip';
import { IconFieldModule } from 'primeng/iconfield';
import { InputIconModule } from 'primeng/inputicon';
import { InputTextModule } from 'primeng/inputtext';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { ToastModule } from 'primeng/toast';
import { ConfirmationService, MessageService } from 'primeng/api';
import { SharedTable } from '../../components/shared-table/shared-table.component';
import { MonsterService } from '../../services/monster.service';
import { EntityCard } from '../../components/entity-card/entity-card';
import { MonsterEditorComponent } from '../../components/editors/monster-editor/monster-editor';
import { Monster, Entity } from '../../models';
import { CombatantService } from '../../services/combatant.service';

@Component({
  selector: 'app-bestiary-shell',
  imports: [
    ButtonModule,
    TooltipModule,
    IconFieldModule,
    InputIconModule,
    InputTextModule,
    ConfirmDialogModule,
    SharedTable,
    EntityCard,
    MonsterEditorComponent,
    ToastModule
  ],
  providers: [ConfirmationService, MessageService],
  templateUrl: './bestiary-shell.html',
  styles: [
    `
      :host {
        display: block;
      }
    `,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class BestiaryShell implements OnInit {
  public readonly monsterService = inject(MonsterService);
  private readonly confirmationService = inject(ConfirmationService);
  private readonly combatantService = inject(CombatantService);
  private readonly messageService = inject(MessageService);
  public readonly searchTerm = signal('');
  public readonly isEditorVisible = signal(false);
  public readonly monsterToEdit = signal<Monster | null>(null);

  ngOnInit(): void {
    this.monsterService.getSummaries().subscribe();
  }

  onCreateMonster(): void {
    this.monsterToEdit.set(null);
    this.isEditorVisible.set(true);
  }

  onEditMonster(monster: Monster): void {
    this.monsterToEdit.set(monster);
    this.isEditorVisible.set(true);
  }

  onDeleteMonster(monster: Monster): void {
    this.confirmationService.confirm({
      message: `Are you sure you want to delete ${monster.name}? This action cannot be undone.`,
      header: 'Confirm Deletion',
      icon: 'pi pi-exclamation-triangle',
      acceptButtonStyleClass: 'p-button-danger',
      accept: () => {
        this.monsterService.deleteMonster(monster.id).subscribe();
      }
    });
  }

  onSearch(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.searchTerm.set(target.value);
  }

  onClearSearch(): void {
    this.searchTerm.set('');
  }

  onAddToSimulator(monster: Entity): void {
    const success = this.combatantService.addToSimulator(monster);
    if (success) {
      this.messageService.add({
        severity: 'success',
        summary: 'Added to Simulator',
        detail: `${monster.name} has been added to the encounter simulator.`,
        life: 3000
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
}
