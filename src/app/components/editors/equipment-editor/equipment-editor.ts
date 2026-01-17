import { Component, effect, OnInit, inject, signal, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { DialogModule } from 'primeng/dialog';
import { ButtonModule } from 'primeng/button';
import { InputTextModule } from 'primeng/inputtext';
import { InputNumberModule } from 'primeng/inputnumber';
import { SelectModule } from 'primeng/select';
import { CheckboxModule } from 'primeng/checkbox';
import { MessageModule } from 'primeng/message';
import { TooltipModule } from 'primeng/tooltip';
import { TabsModule } from 'primeng/tabs';
import { FluidModule } from 'primeng/fluid';
import { BaseEditorDirective } from '../base-editor.directive';
import { Weapon, Armor, EquipmentItem, DiceType, DamageType } from '../../../models';
import { CustomEntityType } from '../../../services/custom-content.service';

@Component({
  selector: 'app-equipment-editor',
  imports: [
    CommonModule,
    ReactiveFormsModule,
    DialogModule,
    ButtonModule,
    InputTextModule,
    InputNumberModule,
    SelectModule,
    CheckboxModule,
    MessageModule,
    TooltipModule,
    TabsModule,
    FluidModule
  ],
  templateUrl: './equipment-editor.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class EquipmentEditorComponent extends BaseEditorDirective<EquipmentItem> implements OnInit {
  private readonly fb = inject(FormBuilder);

  public readonly activeTab = signal<'weapon' | 'armor'>('weapon');

  public weaponForm: FormGroup = this.fb.group({
    id: [null],
    name: ['', [Validators.required]],
    numberOfDice: [1, [Validators.required, Validators.min(0), Validators.max(20)]],
    die: [DiceType.D6, [Validators.required]],
    damageType: [DamageType.Slashing, [Validators.required]],
    properties: this.fb.group({
      isVersatile: [false],
      isFinesse: [false],
      isRanged: [false],
      isHeavy: [false],
      isTwoHanded: [false],
      isLight: [false],
      isThrown: [false],
      isOnlyRanged: [false]
    }),
    modifiers: this.fb.group({
      isMagic: [false],
      isSilvered: [false],
      isAdamantine: [false],
      isColdForgedIron: [false],
      attackBonus: [0, [Validators.min(-20), Validators.max(20)]],
      damageBonus: [0, [Validators.min(-20), Validators.max(20)]]
    })
  });

  public armorForm: FormGroup = this.fb.group({
    id: [null],
    name: ['', [Validators.required]],
    ac: [10, [Validators.required, Validators.min(0), Validators.max(30)]],
    dexBonus: [false],
    maxBonus: [false],
    minimumStrength: [0, [Validators.required, Validators.min(0), Validators.max(30)]],
    modifier: [0, [Validators.required, Validators.min(-20), Validators.max(20)]]
  });

  protected readonly diceOptions = [
    { label: 'd4', value: DiceType.D4 },
    { label: 'd6', value: DiceType.D6 },
    { label: 'd8', value: DiceType.D8 },
    { label: 'd10', value: DiceType.D10 },
    { label: 'd12', value: DiceType.D12 },
    { label: 'd20', value: DiceType.D20 }
  ];

  protected readonly damageTypes = Object.values(DamageType).map(t => ({
    label: t.charAt(0).toUpperCase() + t.slice(1),
    value: t
  }));

  constructor() {
    super();

    effect(() => {
      const item = this.itemToEdit();
      if (item) {
        if ('die' in item) {
          this.activeTab.set('weapon');
          this.weaponForm.patchValue(item);
        } else {
          this.activeTab.set('armor');
          this.armorForm.patchValue(item);
        }
      } else {
        this.resetForms();
      }
    });
  }

  ngOnInit(): void {}

  protected override getEntityType(): CustomEntityType {
    return 'equipment';
  }

  private resetForms() {
    this.weaponForm.reset({
      numberOfDice: 1,
      die: DiceType.D6,
      damageType: DamageType.Slashing,
      properties: {
        isVersatile: false,
        isFinesse: false,
        isRanged: false,
        isHeavy: false,
        isTwoHanded: false,
        isLight: false,
        isThrown: false,
        isOnlyRanged: false
      },
      modifiers: {
        isMagic: false,
        isSilvered: false,
        isAdamantine: false,
        isColdForgedIron: false,
        attackBonus: 0,
        damageBonus: 0
      }
    });
    this.armorForm.reset({
      ac: 10,
      dexBonus: false,
      maxBonus: false,
      minimumStrength: 0,
      modifier: 0
    });
  }

  onSave(): void {
    if (this.activeTab() === 'weapon') {
      if (this.weaponForm.valid) {
        this.saveEntity(this.weaponForm.value);
      }
    } else {
      if (this.armorForm.valid) {
        this.saveEntity(this.armorForm.value);
      }
    }
  }

  onTabChange(value: string | number | undefined): void {
    if (value === 'weapon' || value === 'armor') {
      this.activeTab.set(value);
    }
  }
}
