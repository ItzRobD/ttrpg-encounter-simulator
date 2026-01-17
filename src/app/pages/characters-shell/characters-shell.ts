import { Component, inject, OnInit, signal, ChangeDetectionStrategy} from '@angular/core';
import {SharedTable} from "../../components/shared-table/shared-table.component";
import {Button} from "primeng/button";
import {EntityCard} from "../../components/entity-card/entity-card";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {InputText} from "primeng/inputtext";
import {Tooltip} from "primeng/tooltip";
import {ConfirmationService, MessageService} from 'primeng/api';
import {ConfirmDialogModule} from 'primeng/confirmdialog';
import {CharacterService} from '../../services/character.service';
import {Character, CharacterSummary} from '../../models';
import {CharacterEditorComponent} from '../../components/editors/character-editor/character-editor';

@Component({
  selector: 'app-characters-shell',
  imports: [
      SharedTable,
      Button,
      EntityCard,
      IconField,
      InputIcon,
      InputText,
      Tooltip,
      ConfirmDialogModule,
      CharacterEditorComponent
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

  public readonly searchTerm = signal('');
  public readonly isEditorVisible = signal(false);
  public readonly characterToEdit = signal<Character | null>(null);

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

  onEditCharacter(character: Character): void {
    this.characterToEdit.set(character);
    this.isEditorVisible.set(true);
  }

  onDeleteCharacter(character: Character | CharacterSummary): void {
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
}
