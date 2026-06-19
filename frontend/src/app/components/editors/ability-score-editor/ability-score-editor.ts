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
  styles: [`
    .ability-grid {
      display: grid;
      grid-template-columns: repeat(6, 1fr);
      gap: 0.25rem;
    }
    @media (max-width: 768px) {
      .ability-grid {
        grid-template-columns: repeat(3, 1fr);
      }
    }
    @media (max-width: 480px) {
      .ability-grid {
        grid-template-columns: repeat(2, 1fr);
      }
    }
    .ability-card {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 0.25rem;
      border-radius: 4px;
      border: 1px solid var(--surface-border);
      background: var(--surface-card);
      text-align: center;
      transition: border-color 0.2s;
    }
    .ability-card:hover {
      border-color: var(--primary-color);
    }
    .prof-toggle {
      cursor: pointer;
      font-size: 0.65rem;
      transition: transform 0.1s;
    }
    .prof-toggle:hover {
      transform: scale(1.2);
    }
    .mod-text {
      font-size: 0.75rem;
      color: var(--text-color-secondary);
      margin-top: -0.25rem;
      margin-bottom: 0.25rem;
    }
    .score-input-container {
      width: 100%;
    }
    .ability-label {
      font-size: 0.7rem;
      font-weight: bold;
      text-transform: uppercase;
      color: var(--text-color-secondary);
    }
    :host ::ng-deep .p-inputnumber {
      width: 100%;
    }
    :host ::ng-deep .p-inputnumber-input {
      text-align: center !important;
      padding: 0 !important;
      border: none;
      background: transparent;
      font-size: 1.15rem;
      font-weight: 700;
      width: 100%;
      height: 2rem;
      display: block;
      margin: 0 auto;
      box-sizing: border-box;
    }
    :host ::ng-deep .p-inputnumber-input:focus {
      box-shadow: none;
    }
  `]
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
