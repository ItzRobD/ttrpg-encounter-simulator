import { Component, inject, OnInit, signal, ChangeDetectionStrategy} from '@angular/core';
import {SharedTable} from "../../components/shared-table/shared-table.component";
import {Button} from "primeng/button";
import {ActorCard} from "../../components/actor-card/actor-card.component";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {InputText} from "primeng/inputtext";
import {Tooltip} from "primeng/tooltip";
import {ConfirmationService, MessageService} from 'primeng/api';
import {ConfirmDialogModule} from 'primeng/confirmdialog';
import {ToastModule} from 'primeng/toast';
import {CharacterService} from '../../services/character.service';
import { Actor, ActorSummary } from '../../models';
import {CharacterEditorComponent} from '../../components/editors/character-editor/character-editor';
import {CombatantService} from '../../services/combatant.service';

@Component({
  selector: 'app-characters-shell',
  imports: [
      SharedTable,
      Button,
      ActorCard,
      IconField,
      InputIcon,
      InputText,
      Tooltip,
      ConfirmDialogModule,
      CharacterEditorComponent,
      ToastModule
  ],
  providers: [ConfirmationService, MessageService],
  templateUrl: './characters-shell.html',
  styles: [
    `
      :host {
        display: block;
      }
    `,
  ],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class CharactersShell implements OnInit {
  public readonly characterService = inject(CharacterService);
  private readonly confirmationService = inject(ConfirmationService);
  private readonly combatantService = inject(CombatantService);
  private readonly messageService = inject(MessageService);

  public readonly searchTerm = signal('');
  public readonly isEditorVisible = signal(false);
  public readonly characterToEdit = signal<Actor | null>(null);

  ngOnInit(): void {
    this.characterService.getSummaries().subscribe();
  }

  onSearch(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.searchTerm.set(target.value);
  }

  onClearSearch(): void {
    this.searchTerm.set('');
  }

  onCreateCharacter(): void {
    this.characterToEdit.set(null);
    this.isEditorVisible.set(true);
  }

  onEditCharacter(character: Actor): void {
    this.characterToEdit.set(character);
    this.isEditorVisible.set(true);
  }

  onDeleteCharacter(character: Actor | ActorSummary): void {
    this.confirmationService.confirm({
      message: `Are you sure you want to delete ${character.name}? This action cannot be undone.`,
      header: 'Confirm Deletion',
      icon: 'pi pi-exclamation-triangle',
      acceptButtonStyleClass: 'p-button-danger',
      accept: () => {
        this.characterService.deleteCharacter(character.id).subscribe();
      }
    });
  }

  onAddToSimulator(character: Actor): void {
    const success = this.combatantService.addToSimulator(character);
    if (success) {
      this.messageService.add({
        severity: 'success',
        summary: 'Added to Simulator',
        detail: `${character.name} has been added to the encounter simulator.`,
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
