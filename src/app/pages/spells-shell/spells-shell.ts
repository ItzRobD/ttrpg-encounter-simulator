import { Component, inject, OnInit, signal} from '@angular/core';
import {Button} from 'primeng/button';
import {SpellCard} from '../../components/spell-card/spell-card';
import {IconField} from 'primeng/iconfield';
import {InputIcon} from 'primeng/inputicon';
import {InputText} from 'primeng/inputtext';
import {SharedTable} from '../../components/shared-table/shared-table.component';
import {Tooltip} from 'primeng/tooltip';
import {SpellsService} from '../../services/spells.service';

@Component({
  selector: 'app-spells-shell',
  imports: [
    Button,
    SpellCard,
    IconField,
    InputIcon,
    InputText,
    SharedTable,
    Tooltip
  ],
  templateUrl: './spells-shell.html',
  styles: [`
    :host {
      display: block;
    }
  `]
})
export class SpellsShell implements OnInit {
  public readonly spellsService = inject(SpellsService);
  public readonly searchTerm = signal('');

  ngOnInit(): void {
    this.spellsService.getSummaries().subscribe();
  }

  onSearch(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.searchTerm.set(target.value);
    }

    onClearSearch(): void {
      this.searchTerm.set('');
    }
    }
