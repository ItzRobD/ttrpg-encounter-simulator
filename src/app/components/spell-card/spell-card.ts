import { Component, input, computed } from '@angular/core';
import { CommonModule, TitleCasePipe } from '@angular/common';
import { CardModule } from 'primeng/card';
import { TagModule } from 'primeng/tag';
import { Spell } from '../../models';

@Component({
  selector: 'app-spell-card',
  standalone: true,
  imports: [CardModule, TitleCasePipe, CommonModule, TagModule],
  templateUrl: './spell-card.html',
  styleUrl: './spell-card.css',
})
export class SpellCard {
  public readonly item = input.required<Spell>();

  protected readonly levelDisplay = computed(() => {
    const spell = this.item();
    if (spell.level === undefined) return 'Unknown Level';
    if (spell.level === 0) return 'Cantrip';

    const suffix = (l: number) => {
      if (l >= 11 && l <= 13) return 'th';
      switch (l % 10) {
        case 1: return 'st';
        case 2: return 'nd';
        case 3: return 'rd';
        default: return 'th';
      }
    };
    return `${spell.level}${suffix(spell.level)}-level`;
  });

  protected readonly spellTags = computed(() => {
    const spell = this.item();
    const tags: { label: string, severity: "success" | "secondary" | "info" | "warn" | "danger" | "contrast" | null | undefined }[] = [];

    if (spell.isConcentration) {
      tags.push({ label: 'Concentration', severity: 'warn' });
    }
    if (spell.isRitual) {
      tags.push({ label: 'Ritual', severity: 'info' });
    }

    return tags;
  });

  protected readonly sortedFormulas = computed(() => {
    const formulas = this.item().formulas;
    if (!formulas) return [];

    return Object.entries(formulas)
      .map(([key, value]) => ({ level: Number(key), formula: value }))
      .sort((a, b) => a.level - b.level);
  });
}
