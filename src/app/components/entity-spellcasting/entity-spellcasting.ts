import { Component, computed, input } from '@angular/core';
import { CardModule } from 'primeng/card';
import { TooltipModule } from 'primeng/tooltip';
import { Spell, Spellcasting } from '../../models';

interface SpellGroup {
  level: number;
  label: string;
  spells: Spell[];
  isInnate: boolean;
}

@Component({
  selector: 'app-entity-spellcasting',
  standalone: true,
  imports: [CardModule, TooltipModule],
  templateUrl: './entity-spellcasting.html',
  styleUrl: './entity-spellcasting.css',
})
export class EntitySpellcasting {
  public readonly spellcasting = input.required<Spellcasting>();
  public readonly hideState = input<boolean>(false);

  protected readonly groupedSpells = computed(() => {
    const sc = this.spellcasting();
    console.log('Rendering spellcasting data:', sc);
    const leveledGroups: { [level: number]: Spell[] } = {};
    const innateGroups: { [usage: string]: Spell[] } = {};

    sc.spells.forEach((spell) => {
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
        spells: spells.sort((a, b) => a.name.localeCompare(b.name)),
        isInnate: false
      }))
      .sort((a, b) => a.level - b.level);

    const innate: SpellGroup[] = Object.entries(innateGroups)
      .map(([usage, spells]) => ({
        level: -1,
        label: usage,
        spells: spells.sort((a, b) => a.name.localeCompare(b.name)),
        isInnate: true
      }))
      .sort((a, b) => {
        // Sort innate by frequency: "At Will" first, then descending by count?
        // Let's just do a simple string sort or specific order
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
