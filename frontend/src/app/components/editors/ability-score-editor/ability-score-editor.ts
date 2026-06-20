import { Component, input, ChangeDetectionStrategy } from '@angular/core';
import { FormGroup, ReactiveFormsModule } from '@angular/forms';
import { InputNumberModule } from 'primeng/inputnumber';
import { TooltipModule } from 'primeng/tooltip';
import { getModifier, formatModifier } from '../../../shared/utils/dnd-utils';

@Component({
  selector: 'app-ability-score-editor',
  imports: [ReactiveFormsModule, InputNumberModule, TooltipModule],
  templateUrl: './ability-score-editor.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
  styleUrl: './ability-score-editor.scss',
})
export class AbilityScoreEditorComponent {
  public readonly group = input.required<FormGroup>(); // Expects group with 'abilityScores' and 'proficiencies' sub-groups

  protected readonly abilities = [
    { name: 'Strength', key: 'strength', short: 'STR' },
    { name: 'Dexterity', key: 'dexterity', short: 'DEX' },
    { name: 'Constitution', key: 'constitution', short: 'CON' },
    { name: 'Intelligence', key: 'intelligence', short: 'INT' },
    { name: 'Wisdom', key: 'wisdom', short: 'WIS' },
    { name: 'Charisma', key: 'charisma', short: 'CHA' }
  ];

  getModifier(score: number): string {
    return formatModifier(getModifier(score));
  }

  toggleProficiency(ability: string): void {
    const profGroup = this.group().get('proficiencies') as FormGroup;
    if (profGroup) {
      const current = profGroup.get(ability)?.value;
      profGroup.get(ability)?.setValue(!current);
      profGroup.get(ability)?.markAsTouched();
    }
  }

  isProficient(ability: string): boolean {
    return this.group().get('proficiencies')?.get(ability)?.value || false;
  }
}
