import {
  Component,
  inject,
  ChangeDetectionStrategy,
  input,
  computed,
} from '@angular/core';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { toSignal } from '@angular/core/rxjs-interop';
import { map } from 'rxjs/operators';
import { TableModule } from 'primeng/table';
import { TooltipModule } from 'primeng/tooltip';
import { MessageModule } from 'primeng/message';
import { MonsterService } from '../../services/monster.service';
import { CommonModule } from '@angular/common';
import { CharacterService } from '../../services/character.service';
import { CrFormatPipe } from '../../pipes/cr-format.pipe';
import { EntityService } from '../../services/entity.service.interface';
import { Entity, EntitySummary, MonsterSummary, CharacterSummary } from '../../models';

@Component({
  selector: 'app-entity-table',
  standalone: true,
  imports: [TableModule, TooltipModule, MessageModule, CommonModule, CrFormatPipe],
  templateUrl: './entity-table.component.html',
  styleUrl: './entity-table.component.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class EntityTable {
  private readonly monsterService = inject(MonsterService);
  private readonly characterService = inject(CharacterService);
  private readonly breakpointObserver = inject(BreakpointObserver);

  public readonly mode = input<'monster' | 'character'>('monster');
  public readonly searchTerm = input('');

  public readonly activeService = computed<EntityService<Entity, EntitySummary>>(() =>
    this.mode() === 'monster' ? this.monsterService : this.characterService
  );

  protected readonly colWidths = {
    desktop: {
      name: { width: '35%', minWidth: '15rem' },
      type: { width: '25%', minWidth: '10rem' },
      size: { width: '2rem', minWidth: '2rem' },
      cr: { width: '3rem', minWidth: '3rem' },
      ac: { width: '3rem', minWidth: '3rem' },
      level: { width: '3rem', minWidth: '3rem' },
      class: { width: '20%', minWidth: '10rem' },
      race: { width: '20%', minWidth: '10rem' },
      status: { width: '0', minWidth: '0' },
    },
    mobile: {
      name: { width: '40%', minWidth: '12rem' },
      type: { width: '30%', minWidth: '8rem' },
      size: { width: '0', minWidth: '0' },
      cr: { width: '5rem', minWidth: '5rem' },
      ac: { width: '0', minWidth: '0' },
      level: { width: '5rem', minWidth: '5rem' },
      class: { width: '15%', minWidth: '12rem' },
      race: { width: '15%', minWidth: '12rem' },
      status: { width: '4rem', minWidth: '4rem' },
    }
  };

  public readonly isMobile = toSignal(
    this.breakpointObserver.observe(Breakpoints.Handset).pipe(map((result) => result.matches)),
    { initialValue: false }
  );

  public readonly filteredEntities = computed(() => {
    const entities = this.activeService().summaries();
    const term = this.searchTerm().toLowerCase().trim();

    if (!term) {
      return entities;
    }

    return entities.filter((e: EntitySummary) => {
      const basicMatch = e.name.toLowerCase().includes(term);

      if (this.mode() === 'monster') {
        const m = e as MonsterSummary;
        return (
          basicMatch ||
          m.type?.toLowerCase().includes(term) ||
          m.size?.toLowerCase().includes(term) ||
          m.cr?.toString().includes(term) ||
          m.ac?.toString().includes(term) ||
          (m.isLegendary && 'legendary'.includes(term)) ||
          (m.isSpellcaster && 'spellcaster'.includes(term))
        );
      } else {
        const c = e as CharacterSummary;
        return (
          basicMatch ||
          c.race?.toLowerCase().includes(term) ||
          c.class?.toLowerCase().includes(term) ||
          c.level?.toString().includes(term)
        );
      }
    });
  });

  public readonly selectedSummary = computed(() => {
    const selected = this.activeService().selectedEntity();
    if (!selected) return null;
    return this.activeService().summaries().find((s: EntitySummary) => s.id === selected.id) || null;
  });

  onRowSelect(event: any): void {
    const summary = event.data as EntitySummary;
    this.activeService().selectEntityByID(summary.id.toString()).subscribe();
  }

  onRowUnselect(): void {
    this.activeService().selectEntity(null);
  }
}
