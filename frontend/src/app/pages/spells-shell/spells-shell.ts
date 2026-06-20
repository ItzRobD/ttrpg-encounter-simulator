import { Component, inject, OnInit, signal, ChangeDetectionStrategy } from '@angular/core';
import {Button} from 'primeng/button';
import {Tooltip} from 'primeng/tooltip';
import {TabsModule} from 'primeng/tabs';
import {LibraryPage} from '../../components/library-page/library-page';
import {SpellCard} from '../../components/spell-card/spell-card';
import {SharedTable} from '../../components/shared-table/shared-table.component';
import {SpellsService} from '../../services/spells.service';
import {SpellEditorComponent} from '../../components/editors/spell-editor/spell-editor';
import {Spell} from '../../models';

@Component({
  selector: 'app-spells-shell',
  imports: [
    Button,
    Tooltip,
    TabsModule,
    LibraryPage,
    SpellCard,
    SharedTable,
    SpellEditorComponent,
  ],
  templateUrl: './spells-shell.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
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

}
