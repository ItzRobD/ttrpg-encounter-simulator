import {
  Component,
  inject,
  ChangeDetectionStrategy,
  input,
  computed,
  Signal,
  output,
} from '@angular/core';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { toSignal } from '@angular/core/rxjs-interop';
import { map, take } from 'rxjs/operators';
import { Observable } from 'rxjs';
import { TableModule } from 'primeng/table';
import { TooltipModule } from 'primeng/tooltip';
import { MessageModule } from 'primeng/message';
import { ButtonModule } from 'primeng/button';
import { MonsterService } from '../../services/monster.service';
import { CommonModule } from '@angular/common';
import { CharacterService } from '../../services/character.service';
import { EquipmentService } from '../../services/equipment.service';
import { CrFormatPipe } from '../../pipes/cr-format.pipe';
import { EntityService } from '../../services/entity.service.interface';
import { MapperService } from '../../services/mapper.service';
import {
  Entity,
  EntitySummary,
  MonsterSummary,
  CharacterSummary,
  EquipmentSummary,
  Monster,
  Character,
  EquipmentItem, SpellSummary, Spell
} from '../../models';
import {SpellsService} from '../../services/spells.service';

type SupportedService =
  | EntityService<Monster, MonsterSummary>
  | EntityService<Character, CharacterSummary>
  | (Omit<EquipmentService, 'selectedItem' | 'selectItem'> & {
      selectedEntity: Signal<EquipmentItem | null>;
      selectEntity: (item: EquipmentItem | null) => void;
      selectEntityByID: (id: string) => Observable<EquipmentItem>;
    })
  | (Omit<SpellsService, 'selectedSpell' | 'selectSpell'> & {
      selectedEntity: Signal<Spell | null>;
      selectEntity: (spell: Spell | null) => void;
      selectEntityByID: (id: string) => Observable<Spell>;
    });

@Component({
  selector: 'app-shared-table',
  imports: [TableModule, TooltipModule, MessageModule, CommonModule, CrFormatPipe, ButtonModule],
  templateUrl: './shared-table.component.html',
  styleUrl: './shared-table.component.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class SharedTable {
  private readonly monsterService = inject(MonsterService);
  private readonly characterService = inject(CharacterService);
  private readonly equipmentService = inject(EquipmentService);
  private readonly spellsService = inject(SpellsService);
  private readonly breakpointObserver = inject(BreakpointObserver);
  protected readonly mapperService = inject(MapperService);

  public readonly mode = input<'monster' | 'character' | 'equipment' | 'spells'>('monster');
  public readonly searchTerm = input('');
  public readonly categoryFilter = input<string>('all');
  public readonly showAddToSimulator = input(false);

  public readonly addToSimulator = output<Entity>();

  public readonly activeService = computed<SupportedService>(() => {
    switch (this.mode()) {
      case 'monster': return this.monsterService;
      case 'character': return this.characterService;
      case 'equipment': {
        const service = this.equipmentService;
        return {
          ...service,
          selectedEntity: service.selectedItem,
          selectEntity: service.selectItem.bind(service),
          selectEntityByID: (id: string) => service.selectItemByID(id, 'Weapon') // Defaulting to Weapon for now
        } as unknown as SupportedService;
      }
      case 'spells': {
        const service = this.spellsService;
        return {
          ...service,
          selectedEntity: service.selectedSpell,
          selectEntity: service.selectSpell.bind(service),
          selectEntityByID: (id: string) => service.selectSpellByID(id)
        } as unknown as SupportedService;
      }
    }
  });

  protected readonly colWidths = {
    desktop: {
      name: { width: '35%', minWidth: '15rem' },
      type: { width: '15%', minWidth: '10rem' },
      size: { width: '2rem', minWidth: '2rem' },
      cr: { width: '3rem', minWidth: '3rem' },
      ac: { width: '3rem', minWidth: '3rem' },
      level: { width: '3rem', minWidth: '3rem' },
      class: { width: '20%', minWidth: '10rem' },
      race: { width: '20%', minWidth: '10rem' },
      detail: { width: '25%', minWidth: '10rem' },
      status: { width: '0', minWidth: '0' },
      actions: { width: '4rem', minWidth: '4rem' },
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
      detail: { width: '25%', minWidth: '10rem' },
      status: { width: '4rem', minWidth: '4rem' },
      actions: { width: '4rem', minWidth: '4rem' },
    }
  };

  public readonly isMobile = toSignal(
    this.breakpointObserver.observe(Breakpoints.Handset).pipe(map((result) => result.matches)),
    { initialValue: false }
  );

  public readonly filteredItems = computed(() => {
    let summaries = this.activeService().summaries();
    const term = this.searchTerm().toLowerCase().trim();
    const category = this.categoryFilter().toLowerCase();

    let items: (MonsterSummary | CharacterSummary | EquipmentSummary | SpellSummary)[] = summaries;

    // 1. Category Filtering
    if (category !== 'all') {
      items = items.filter((i: any) => {
        if (this.mode() === 'monster') {
          const m = i as MonsterSummary;
          if (category === 'srd') return !m.isCustom;
          if (category === 'custom') return !!m.isCustom;
        } else if (this.mode() === 'equipment') {
          const eq = i as EquipmentSummary;
          if (category === 'armor') return eq.type === 'Armor' || eq.type === 'Shield';
          if (category === 'weapons') return eq.type === 'Weapon';
        } else if (this.mode() === 'spells') {
          const s = i as SpellSummary;
          if (category === 'damage') return s.spellType === 'damage';
          if (category === 'healing') return s.spellType === 'healing';
          if (category === 'utility') return s.spellType === 'other';
        }
        return true;
      });
    }

    // 2. Search Term Filtering
    if (!term) {
      return items;
    }

    return (items as (MonsterSummary | CharacterSummary | EquipmentSummary | SpellSummary)[]).filter((i) => {
      const basicMatch = i.name.toLowerCase().includes(term);

      if (this.mode() === 'monster') {
        const m = i as MonsterSummary;
        return (
          basicMatch ||
          m.type?.toLowerCase().includes(term) ||
          m.size?.toLowerCase().includes(term) ||
          m.cr?.toString().includes(term) ||
          m.ac?.toString().includes(term) ||
          (m.isLegendary && 'legendary'.includes(term)) ||
          (m.isSpellcaster && 'spellcaster'.includes(term))
        );
      } else if (this.mode() === 'character') {
        const c = i as CharacterSummary;
        return (
          basicMatch ||
          c.race?.toLowerCase().includes(term) ||
          c.class?.toLowerCase().includes(term) ||
          c.level?.toString().includes(term)
        );
      } else if (this.mode() === 'equipment') {
        const eq = i as EquipmentSummary;
        const propertyMatch = eq.properties ? (
          (eq.properties.isVersatile && 'versatile'.includes(term)) ||
          (eq.properties.isFinesse && 'finesse'.includes(term)) ||
          (eq.properties.isHeavy && 'heavy'.includes(term)) ||
          (eq.properties.isLight && 'light'.includes(term)) ||
          (eq.properties.isTwoHanded && 'two-handed'.includes(term)) ||
          (eq.properties.isThrown && 'thrown'.includes(term)) ||
          (eq.properties.isRanged && 'ranged'.includes(term))
        ) : false;
        return (
          basicMatch ||
          eq.type?.toLowerCase().includes(term) ||
          eq.detail?.toLowerCase().includes(term) ||
          propertyMatch
        );
      } else if (this.mode() === 'spells') {
        const spell = i as SpellSummary;
        return (
          basicMatch ||
          spell.spellType?.toLowerCase().includes(term) ||
          spell.level?.toString().includes(term) ||
          (spell.isConcentration && 'concentration'.includes(term)) ||
          (spell.isRitual && 'ritual'.includes(term)) ||
          (spell.isAOE && ('aoe'.includes(term) || 'area of effect'.includes(term))) ||
          (spell.hasDC && ('dc'.includes(term) || 'save'.includes(term)))
        );
      } else {
        return false;
      }
    });
  });

  public readonly selectedSummary = computed(() => {
    const service = this.activeService();
    const selected = service.selectedEntity();
    if (!selected) return null;
    return service.summaries().find((s) => (s as { id: number }).id === (selected as { id: number }).id) || null;
  });

  onRowSelect(event: { data?: MonsterSummary | CharacterSummary | EquipmentSummary | SpellSummary | (MonsterSummary | CharacterSummary | EquipmentSummary | SpellSummary)[] }): void {
    const summary = event.data;
    if (!summary || Array.isArray(summary)) return;

    const id = (summary as { id?: number }).id;
    if (id === undefined || id === null) {
      console.error('Cannot select row: summary.id is undefined or null', summary);
      return;
    }

    const service = this.activeService();
    const mode = this.mode();

    if (mode === 'monster' || mode === 'character') {
      (service as EntityService<Entity, EntitySummary>).selectEntityByID(id.toString()).pipe(take(1)).subscribe();
    } else if (mode === 'equipment') {
      const eqSummary = summary as EquipmentSummary;
      this.equipmentService.selectItemByID(id.toString(), eqSummary.type).pipe(take(1)).subscribe();
    } else if (mode === 'spells') {
      this.spellsService.selectSpellByID(id.toString()).pipe(take(1)).subscribe();
    }
  }

  onRowUnselect(): void {
    this.activeService().selectEntity(null);
  }

  onAddToSimulatorClick(event: Event, summary: MonsterSummary | CharacterSummary): void {
    event.stopPropagation();

    const service = this.activeService();
    const id = (summary as { id?: number }).id;
    if (id === undefined || id === null) return;

    const mode = this.mode();
    if (mode === 'monster' || mode === 'character') {
      const entityService = service as EntityService<Entity, EntitySummary>;
      entityService.selectEntityByID(id.toString()).pipe(take(1)).subscribe({
        next: () => {
          const entity = entityService.selectedEntity();
          if (entity) this.addToSimulator.emit(entity);
        }
      });
    }
  }
}
