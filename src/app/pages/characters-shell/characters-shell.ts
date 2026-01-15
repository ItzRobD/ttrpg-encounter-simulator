import { Component, inject, OnInit, signal} from '@angular/core';
import {SharedTable} from "../../components/shared-table/shared-table.component";
import {Button} from "primeng/button";
import {EntityCard} from "../../components/entity-card/entity-card";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {InputText} from "primeng/inputtext";
import {Tooltip} from "primeng/tooltip";
import {CharacterService} from '../../services/character.service';

@Component({
  selector: 'app-characters-shell',
  standalone: true,
  imports: [
      SharedTable,
      Button,
      EntityCard,
      IconField,
      InputIcon,
      InputText,
      Tooltip
  ],
  templateUrl: './characters-shell.html',
  styles: [
    `
      :host {
        display: block;
      }
    `,
  ],
})
export class CharactersShell implements OnInit {
  public readonly characterService = inject(CharacterService);
  public readonly searchTerm = signal('');

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
}
