import { Component, computed, input } from '@angular/core';
import { CardModule } from 'primeng/card';
import { Spellcasting } from '../../models';

@Component({
  selector: 'app-entity-spellcasting',
  standalone: true,
  imports: [CardModule],
  templateUrl: './entity-spellcasting.html',
  styleUrl: './entity-spellcasting.css',
})
export class EntitySpellcasting {
  public readonly spellcasting = input.required<Spellcasting>();

  protected readonly groupedSpells = computed(() => {
    const sc = this.spellcasting();
    console.log('Rendering spellcasting data:', sc);
    const groups: { [level: number]: any[] } = {};
    sc.spells.forEach((spell) => {
      if (!groups[spell.level]) {
        groups[spell.level] = [];
      }
      groups[spell.level].push(spell);
    });
    return Object.entries(groups)
      .map(([level, spells]) => ({
        level: Number(level),
        spells: spells.sort((a, b) => a.name.localeCompare(b.name)),
      }))
      .sort((a, b) => a.level - b.level);
  });
}
