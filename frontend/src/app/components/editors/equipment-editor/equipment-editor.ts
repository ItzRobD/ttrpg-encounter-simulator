import { Component, effect, OnInit, inject, signal, ChangeDetectionStrategy } from '@angular/core';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { DialogModule } from 'primeng/dialog';
import { ButtonModule } from 'primeng/button';
import { SplitButtonModule } from 'primeng/splitbutton';
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
import { CustomActorType } from '../../../services/custom-content.service';

/** Editor form value for a weapon (flat dice fields, distinct from Weapon's damageBlocks). */
interface WeaponFormValue {
  id: string | number | null;
  name: string;
  numberOfDice: number;
  die: DiceType;
  damageType: DamageType;
  properties: {
    isVersatile: boolean; isFinesse: boolean; isRanged: boolean; isHeavy: boolean;
    isTwoHanded: boolean; isLight: boolean; isThrown: boolean; isOnlyRanged: boolean;
  };
  modifiers: {
    isMagic: boolean; isSilvered: boolean; isAdamantine: boolean; isColdForgedIron: boolean;
    attackBonus: number; damageBonus: number;
  };
}

/** Editor form value for armor. */
interface ArmorFormValue {
  id: string | number | null;
  name: string;
  ac: number;
  dexBonus: boolean;
  maxBonus: boolean;
  minimumStrength: number;
  modifier: number;
}

@Component({
  selector: 'app-equipment-editor',
  imports: [
    ReactiveFormsModule,
    DialogModule,
    ButtonModule,
    SplitButtonModule,
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
  styleUrl: './equipment-editor.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class EquipmentEditorComponent extends BaseEditorDirective<EquipmentItem> implements OnInit {
  private readonly fb = inject(FormBuilder);

  public readonly activeTab = signal<'weapon' | 'armor'>('weapon');

  public weaponForm = this.fb.group({
    id: this.fb.control<string | number | null>(null),
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

  public armorForm = this.fb.group({
    id: this.fb.control<string | number | null>(null),
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
        // Armor is the only variant with an `ac` field; everything else is a weapon.
        if ('ac' in item) {
          this.activeTab.set('armor');
          this.armorForm.patchValue(item as unknown as Partial<ArmorFormValue>);
        } else {
          this.activeTab.set('weapon');
          // Boundary: external EquipmentItem patched into the (differently-shaped) form.
          this.weaponForm.patchValue(item as unknown as Partial<WeaponFormValue>);
        }
      } else {
        this.resetForms();
      }
    });
  }

  ngOnInit(): void {}

  protected override getActorType(): CustomActorType {
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
        // Boundary: form value persisted as a Weapon.
        this.saveActor(this.weaponForm.getRawValue() as unknown as Weapon);
      }
    } else {
      if (this.armorForm.valid) {
        this.saveActor(this.armorForm.getRawValue() as unknown as Armor);
      }
    }
  }

  onTabChange(value: string | number | undefined): void {
    if (value === 'weapon' || value === 'armor') {
      this.activeTab.set(value);
    }
  }
}
