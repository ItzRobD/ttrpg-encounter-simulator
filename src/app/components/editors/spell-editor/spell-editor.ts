import { Component, effect, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AbstractControl, FormArray, FormBuilder, FormGroup, ReactiveFormsModule, ValidationErrors, ValidatorFn, Validators } from '@angular/forms';
import { DialogModule } from 'primeng/dialog';
import { ButtonModule } from 'primeng/button';
import { InputTextModule } from 'primeng/inputtext';
import { TextareaModule } from 'primeng/textarea';
import { InputNumberModule } from 'primeng/inputnumber';
import { SelectModule } from 'primeng/select';
import { CheckboxModule } from 'primeng/checkbox';
import { MessageModule } from 'primeng/message';
import { ToggleButtonModule } from 'primeng/togglebutton';
import { SelectButtonModule } from 'primeng/selectbutton';
import { TooltipModule } from 'primeng/tooltip';
import { BaseEditorDirective } from '../base-editor.directive';
import { Spell, CastingTime, SpellType, LevelType, DamageType, SpellFormula, DiceType, Ability, SaveSuccessEffect } from '../../../models';
import { CustomEntityType } from '../../../services/custom-content.service';
import { FluidModule } from 'primeng/fluid';

@Component({
  selector: 'app-spell-editor',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    DialogModule,
    ButtonModule,
    InputTextModule,
    TextareaModule,
    InputNumberModule,
    SelectModule,
    CheckboxModule,
    MessageModule,
    FluidModule,
    ToggleButtonModule,
    SelectButtonModule,
    TooltipModule
  ],
  templateUrl: './spell-editor.html'
})
export class SpellEditorComponent extends BaseEditorDirective<Spell> implements OnInit {
  private readonly fb = inject(FormBuilder);

  protected readonly diceOptions = [
    { label: 'd4', value: DiceType.D4 },
    { label: 'd6', value: DiceType.D6 },
    { label: 'd8', value: DiceType.D8 },
    { label: 'd10', value: DiceType.D10 },
    { label: 'd12', value: DiceType.D12 },
    { label: 'd20', value: DiceType.D20 }
  ];

  private duplicateLevelsValidator(): ValidatorFn {
    return (control: AbstractControl): ValidationErrors | null => {
      const formArray = control as FormArray;
      const levels = formArray.controls.map(c => c.get('castLevel')?.value);

      const duplicates = levels.filter((level, index) => levels.indexOf(level) !== index);

      formArray.controls.forEach(c => {
        const levelControl = c.get('castLevel');
        if (levelControl) {
          const isDuplicate = duplicates.includes(levelControl.value);
          const currentErrors = levelControl.errors || {};
          if (isDuplicate) {
            levelControl.setErrors({ ...currentErrors, duplicate: true });
          } else {
            const { duplicate, ...remainingErrors } = currentErrors;
            levelControl.setErrors(Object.keys(remainingErrors).length ? remainingErrors : null);
          }
        }
      });

      return duplicates.length > 0 ? { duplicateLevels: true } : null;
    };
  }

  public spellForm: FormGroup = this.fb.group({
    id: [null],
    name: ['', [Validators.required]],
    description: ['', [Validators.required]],
    level: [0, [Validators.required, Validators.min(0), Validators.max(9)]],
    spellType: ['damage', [Validators.required]],
    castingTime: ['action', [Validators.required]],
    isConcentration: [false],
    isRitual: [false],
    isAOE: [false],
    isTouch: [false],
    hasDC: [false],
    spellDC: this.fb.group({
      ability: [Ability.Dexterity],
      onSuccess: ['half']
    }),
    isAutoHit: [false],
    levelType: [LevelType.Slot],
    formulas: this.fb.array([], [this.duplicateLevelsValidator()])
  });

  protected readonly spellTypes: { label: string, value: SpellType }[] = [
    { label: 'Damage', value: 'damage' },
    { label: 'Healing', value: 'healing' },
    { label: 'Other', value: 'other' }
  ];

  protected readonly castingTimes: { label: string, value: CastingTime }[] = [
    { label: 'Action', value: 'action' },
    { label: 'Bonus Action', value: 'bonus action' },
    { label: 'Reaction', value: 'reaction' },
    { label: 'Minute', value: 'minute' },
    { label: 'Hour', value: 'hour' }
  ];

  protected readonly damageTypes = Object.values(DamageType).map(t => ({ label: t.charAt(0).toUpperCase() + t.slice(1), value: t }));

  protected readonly abilities = Object.values(Ability).map(a => ({ label: a.charAt(0).toUpperCase() + a.slice(1), value: a }));

  protected readonly successEffects: { label: string, value: SaveSuccessEffect }[] = [
    { label: 'Half Damage', value: 'half' },
    { label: 'No Damage (None)', value: 'none' },
    { label: 'Other', value: 'other' }
  ];

  protected readonly levelTypes = [
    { label: 'Slot Level', value: LevelType.Slot },
    { label: 'Character Level', value: LevelType.Character }
  ];

  get formulas(): FormArray {
    return this.spellForm.get('formulas') as FormArray;
  }

  constructor() {
    super();

    // Watch for itemToEdit changes to populate form
    effect(() => {
      const item = this.itemToEdit();
      if (item) {
        // Disable levelType change watcher temporarily if needed,
        // but here we just want to ensure we don't trigger clear unnecessarily
        this.formulas.clear({ emitEvent: false });
        if (item.formulas) {
          Object.entries(item.formulas).forEach(([level, formula]) => {
            this.addFormula(Number(level), formula);
          });
        }
        this.spellForm.patchValue(item, { emitEvent: false });
      } else {
        this.spellForm.reset({
          level: 0,
          spellType: 'damage',
          castingTime: 'action',
          levelType: LevelType.Slot,
          isConcentration: false,
          isRitual: false,
          isAOE: false,
          isTouch: false,
          hasDC: false,
          isAutoHit: false
        });
        this.formulas.clear();
      }
    });
  }

  ngOnInit(): void {
  }

  protected override getEntityType(): CustomEntityType {
    return 'spells';
  }

  addFormula(level?: number, formula?: SpellFormula): void {
    const defaultLevel = level !== undefined ? level : (this.spellForm.get('level')?.value || 0);
    const levelType = this.spellForm.get('levelType')?.value;
    const minLevel = 1;
    const maxLevel = levelType === LevelType.Slot ? 9 : 20;

    const f = this.fb.group({
      castLevel: [formula?.castLevel ?? defaultLevel, [Validators.required, Validators.min(minLevel), Validators.max(maxLevel)]],
      numberOfDice: [formula?.numberOfDice ?? 1, [Validators.required, Validators.min(0)]],
      die: [formula?.die ?? DiceType.D6, [Validators.required]],
      amountToAdd: [formula?.amountToAdd ?? 0],
      damageType: [formula?.damageType ?? DamageType.Fire],
      useSpellMod: [formula?.useSpellMod ?? false]
    });
    this.formulas.push(f);
  }

  removeFormula(index: number): void {
    this.formulas.removeAt(index);
  }

  onSave(): void {
    if (this.spellForm.invalid) {
      this.spellForm.markAllAsTouched();
      return;
    }

    const rawValue = this.spellForm.getRawValue();

    // Map formulas array back to Record<number, SpellFormula>
    const formulasRecord: Record<number, SpellFormula> = {};
    rawValue.formulas.forEach((f: any) => {
      const avg = (f.numberOfDice * (f.die + 1) / 2) + f.amountToAdd;
      formulasRecord[f.castLevel] = {
        ...f,
        averageValue: Number(avg.toFixed(1))
      };
    });

    const entity: Spell = {
      ...rawValue,
      formulas: formulasRecord
    };

    // Ensure numeric values are correct
    entity.level = Number(entity.level);

    this.saveEntity(entity);
  }
}
