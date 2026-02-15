import { Component, computed, input } from '@angular/core';
import { CardModule } from 'primeng/card';
import { TooltipModule } from 'primeng/tooltip';
import { Spell, MonsterSpellcastingConfig, Spellcasting } from '../../models';

interface SpellGroup {
  level: number;
  label: string;
  spells: Spell[];
  isInnate: boolean;
}

@Component({
  selector: 'app-actor-spellcasting',
  standalone: true,
  imports: [CardModule, TooltipModule],
  templateUrl: './actor-spellcasting.component.html',
  styleUrl: './actor-spellcasting.component.css',
})
export class ActorSpellcasting {
  public readonly spellcasting = input<any | undefined>(undefined);
  public readonly hideState = input<boolean>(false);

  protected readonly data = computed(() => {
    const sc = this.spellcasting();

    if (!sc) return null;

    // Both structures (MonsterSpellcastingConfig and legacy Spellcasting) are now passed via a.spellcasting
    let spells = sc.spells || [];

    if (sc.leveledSpells) {
      spells = [...spells, ...sc.leveledSpells];
    }

    if (sc.innateSpells) {
      // Innate spells from backend are wrapped: { spell: {...}, max_casts_per_day: -1 }
      const flattenedInnate = sc.innateSpells.map((wrapper: any) => {
        if (wrapper.spell) {
          return {
            ...wrapper.spell,
            isInnate: true,
            maxCastsPerDay: wrapper.maxCastsPerDay
          };
        }
        return wrapper;
      });
      spells = [...spells, ...flattenedInnate];
    }

    // Normalize spell slots if they are just numbers (backend format)
    const rawSlots = sc.spellSlots || {};
    const normalizedSlots: any = {};
    Object.entries(rawSlots).forEach(([level, value]) => {
      if (typeof value === 'number') {
        normalizedSlots[level] = { current: value, max: value };
      } else {
        normalizedSlots[level] = value;
      }
    });

    return {
      spellSaveDC: sc.saveDC || sc.spellSaveDC,
      spellAttackBonus: sc.attackModifier || sc.spellAttackBonus,
      spellSlots: normalizedSlots,
      spells: spells
    };
  });

  protected readonly groupedSpells = computed(() => {
    const d = this.data();
    if (!d) return [];

    const leveledGroups: { [level: number]: Spell[] } = {};
    const innateGroups: { [usage: string]: Spell[] } = {};

    const spells = d.spells || [];

    spells.forEach((spell: any) => {
      if (spell.isInnate) {
        const usage = this.getInnateUsageLabel(spell.maxCastsPerDay);
        if (!innateGroups[usage]) {
          innateGroups[usage] = [];
        }
        innateGroups[usage].push(spell);
      } else {
        if (!leveledGroups[spell.level]) {
          leveledGroups[spell.level] = [];
        }
        leveledGroups[spell.level].push(spell);
      }
    });

    const leveled: SpellGroup[] = Object.entries(leveledGroups)
      .map(([level, spells]) => ({
        level: Number(level),
        label: Number(level) === 0 ? 'Cantrips' : 'Level ' + level,
        spells: spells.sort((a, b) => (a.name || '').localeCompare(b.name || '')),
        isInnate: false
      }))
      .sort((a, b) => a.level - b.level);

    const innate: SpellGroup[] = Object.entries(innateGroups)
      .map(([usage, spells]) => ({
        level: -1,
        label: usage,
        spells: spells.sort((a, b) => (a.name || '').localeCompare(b.name || '')),
        isInnate: true
      }))
      .sort((a, b) => {
        if (a.label === 'At Will') return -1;
        if (b.label === 'At Will') return 1;
        return b.label.localeCompare(a.label);
      });

    return [...innate, ...leveled];
  });

  private getInnateUsageLabel(maxCasts: number | undefined): string {
    if (maxCasts === undefined || maxCasts === -1) return 'At Will';
    return `${maxCasts}x/day`;
  }
}
