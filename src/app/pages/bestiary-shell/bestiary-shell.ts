import { Component, ChangeDetectionStrategy, inject, OnInit, signal } from '@angular/core';
import { ButtonModule } from 'primeng/button';
import { TooltipModule } from 'primeng/tooltip';
import { IconFieldModule } from 'primeng/iconfield';
import { InputIconModule } from 'primeng/inputicon';
import { InputTextModule } from 'primeng/inputtext';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { ConfirmationService, MessageService } from 'primeng/api';
import { SharedTable } from '../../components/shared-table/shared-table.component';
import { MonsterService } from '../../services/monster.service';
import { EntityCard } from '../../components/entity-card/entity-card';
import { MonsterEditorComponent } from '../../components/editors/monster-editor/monster-editor';
import { Monster } from '../../models';

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
    MonsterEditorComponent
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
}
