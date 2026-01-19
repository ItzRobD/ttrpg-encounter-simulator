import { Component, inject, OnInit, signal} from '@angular/core';
import {Button} from 'primeng/button';
import {SpellCard} from '../../components/spell-card/spell-card';
import {IconField} from 'primeng/iconfield';
import {InputIcon} from 'primeng/inputicon';
import {InputText} from 'primeng/inputtext';
import {SharedTable} from '../../components/shared-table/shared-table.component';
import {Tooltip} from 'primeng/tooltip';
import {SpellsService} from '../../services/spells.service';
import {SpellEditorComponent} from '../../components/editors/spell-editor/spell-editor';
import {Spell} from '../../models';

import {TabsModule} from 'primeng/tabs';

@Component({
  selector: 'app-spells-shell',
  standalone: true,
  imports: [
    Button,
    SpellCard,
    IconField,
    InputIcon,
    InputText,
    SharedTable,
    Tooltip,
    SpellEditorComponent,
    TabsModule
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
  public readonly isEditorVisible = signal(false);
  public readonly spellToEdit = signal<Spell | null>(null);
  public readonly activeTab = signal('all');

  onTabChange(event: string | number | undefined): void {
    if (typeof event === 'string') {
      this.activeTab.set(event);
    }
  }

  ngOnInit(): void {
    this.spellsService.getSummaries().subscribe();
  }

  onCreateSpell(): void {
    this.spellToEdit.set(null);
    this.isEditorVisible.set(true);
  }

  onEditSpell(spell: Spell): void {
    this.spellToEdit.set(spell);
    this.isEditorVisible.set(true);
  }

  onDeleteSpell(spell: Spell): void {
    if (confirm(`Are you sure you want to delete the custom spell "${spell.name}"?`)) {
      this.spellsService.deleteSpell(spell.id).subscribe();
    }
  }

  onSearch(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.searchTerm.set(target.value);
    }

    onClearSearch(): void {
      this.searchTerm.set('');
    }
    }
