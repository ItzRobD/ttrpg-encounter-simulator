import { Component, effect, OnInit, inject, signal, computed, ChangeDetectionStrategy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { toSignal } from '@angular/core/rxjs-interop';
import { FormArray, FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
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
import { Character, Class, Race, Ability, CasterType, DragonbornColor, DamageType, ResistanceType, AbilityScores, Weapon, Armor, WeaponSlot, Equipment } from '../../../models';
import { CustomEntityType } from '../../../services/custom-content.service';
import { AbilityScoreEditorComponent } from '../ability-score-editor/ability-score-editor';
import { SpellcastingEditorComponent } from '../spellcasting-editor/spellcasting-editor';
import { getProficiencyBonus } from '../../../shared/utils/dnd-utils';
import { EquipmentService } from '../../../services/equipment.service';
import {MapperService} from '../../../services/mapper.service';

@Component({
  selector: 'app-character-editor',
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
    FluidModule,
    AbilityScoreEditorComponent,
    SpellcastingEditorComponent
  ],
  templateUrl: './character-editor.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class CharacterEditorComponent extends BaseEditorDirective<Character> implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly equipmentService = inject(EquipmentService);
  protected readonly mapperService = inject(MapperService);
  protected readonly Race = Race;
  protected readonly Number = Number;

  public hpDisplay = signal<string>('10 (1d10 + 0)');

  protected readonly equipmentSummaries = this.equipmentService.summaries;

  protected readonly armorOptions = computed(() => {
    return this.equipmentSummaries().filter(e => e.type === 'Armor');
  });

  protected readonly shieldOptions = computed(() => {
    return this.equipmentSummaries().filter(e => e.type === 'Shield');
  });

  protected readonly weaponOptions = computed(() => {
    return this.equipmentSummaries().filter(e => e.type === 'Weapon');
  });

  protected readonly rangedWeaponOptions = computed(() => {
    return this.weaponOptions().filter(w => w.properties?.isRanged);
  });

  public characterForm: FormGroup = this.fb.group({
    id: [null],
    name: ['', [Validators.required]],
    raceId: [this.mapperService.getRaceId(Race.Human), [Validators.required]],
    dragonbornColor: [null],
    classId: [this.mapperService.getClassId(Class.Fighter), [Validators.required]],
    level: [1, [Validators.required, Validators.min(1), Validators.max(20)]],
    proficiencyBonus: [2],
    hp: this.fb.group({
      hpSetMethod: [0], // SetValue
      value: [10, [Validators.required, Validators.min(1)]],
      hpAverage: [10],
      numberOfDice: [1],
      hitDie: [10],
      amountToAdd: [0],
      modifier: [0]
    }),
    asConfig: this.fb.group({
      abilityScores: this.fb.group({
        strength: [10, [Validators.required, Validators.min(1), Validators.max(30)]],
        dexterity: [10, [Validators.required, Validators.min(1), Validators.max(30)]],
        constitution: [10, [Validators.required, Validators.min(1), Validators.max(30)]],
        intelligence: [10, [Validators.required, Validators.min(1), Validators.max(30)]],
        wisdom: [10, [Validators.required, Validators.min(1), Validators.max(30)]],
        charisma: [10, [Validators.required, Validators.min(1), Validators.max(30)]]
      }),
      proficiencies: this.fb.group({
        strength: [false],
        dexterity: [false],
        constitution: [false],
        intelligence: [false],
        wisdom: [false],
        charisma: [false]
      })
    }),
    equipment: this.fb.group({
      armorId: [null],
      hasShieldEquipped: [false],
      primaryWeaponId: [null],
      secondaryWeaponId: [null],
      rangedWeaponId: [null]
    }),
    spellcasting: this.fb.group({
      casterType: [CasterType.None],
      ability: [Ability.Intelligence],
      casterLevel: [1],
      spellSaveDC: [10],
      spellAttackBonus: [0],
      spellSlots: this.fb.group(this.createEmptySpellSlots()),
      spells: [[]]
    }),
    state: this.fb.group({
      currentHp: [10],
      maxHp: [10],
      tempHp: [0],
      hitDie: [10],
      conditions: [{}],
      deathSaves: this.fb.group({ successes: [0], failures: [0] }),
      resistances: this.fb.array([]),
      isStable: [true],
      isDead: [false],
      initiative: [0]
    })
  });

  private classValue = toSignal(this.characterForm.get('classId')!.valueChanges, { initialValue: this.characterForm.get('classId')?.value });

  public isSpellcaster = computed(() => {
    const classId = this.classValue();
    if (classId === null || classId === undefined) return false;
    const spellcasterIndices = [2, 3, 4, 7, 8, 10, 11, 12];
    return spellcasterIndices.includes(Number(classId));
  });

  protected readonly races = Object.values(Race).map(r => ({ label: r, value: this.mapperService.getRaceId(r) }));
  protected readonly classes = Object.values(Class)
    .filter(c => c !== Class.Artificer)
    .map(c => ({ label: c, value: this.mapperService.getClassId(c) }));
  protected readonly dragonbornColors = Object.values(DragonbornColor).map(c => ({ label: c, value: c }));
  protected readonly damageTypes = Object.values(DamageType).map(dt => ({ label: dt.charAt(0).toUpperCase() + dt.slice(1), value: dt }));
  protected readonly resistanceTypes = Object.values(ResistanceType)
    .filter(rt => rt !== ResistanceType.None)
    .map(rt => ({ label: rt.charAt(0).toUpperCase() + rt.slice(1), value: rt }));

  get asConfigGroup(): FormGroup {
    return this.characterForm.get('asConfig') as FormGroup;
  }

  get spellcastingGroup(): FormGroup {
    return this.characterForm.get('spellcasting') as FormGroup;
  }

  get resistancesArray(): FormArray {
    return this.characterForm.get('state.resistances') as FormArray;
  }

  get isSaving(): boolean {
    return this.loading();
  }

  constructor() {
    super();
    this.equipmentService.getSummaries().subscribe();
    effect(() => {
      const item = this.itemToEdit();
      if (item) {
        // Initialize spell slots if missing
        if (item.spellcasting && !item.spellcasting.spellSlots) {
          item.spellcasting.spellSlots = {};
        }

        // Handle resistances array
        this.resistancesArray.clear();
        if (item.state?.resistances) {
          Object.entries(item.state.resistances).forEach(([type, resistance]) => {
            if (resistance && resistance !== ResistanceType.None) {
              this.resistancesArray.push(this.fb.group({
                damageType: [type],
                resistanceType: [resistance]
              }));
            }
          });
        }

        this.characterForm.patchValue({
          ...item
        }, { emitEvent: false });

        // Patch equipment specifically if it's in the character model format
        if (item.equipment) {
          const eq = item.equipment;
          this.characterForm.get('equipment')?.patchValue({
            armorId: eq.armorId,
            hasShieldEquipped: eq.hasShieldEquipped,
            primaryWeaponId: eq.primarySlot?.[0]?.weaponId,
            secondaryWeaponId: eq.secondarySlot?.[0]?.weaponId,
            rangedWeaponId: eq.rangedSlot?.[0]?.weaponId
          }, { emitEvent: false });
        }

        if (item.hp) {
          this.updateHPDisplay();
        }
      } else {
        this.resetForm();
      }
    });

    effect(() => {
      const hpGroup = this.characterForm.get('hp');
      const conCtrl = this.characterForm.get('asConfig.abilityScores.constitution');
      if (hpGroup && conCtrl) {
        const sub1 = hpGroup.valueChanges.subscribe(() => this.updateHPDisplay());
        const sub2 = conCtrl.valueChanges.subscribe(() => this.updateHPDisplay());
        return () => {
          sub1.unsubscribe();
          sub2.unsubscribe();
        };
      }
      return;
    });
  }

  public addResistance(): void {
    this.resistancesArray.push(this.fb.group({
      damageType: [DamageType.Acid],
      resistanceType: [ResistanceType.Resistant]
    }));
  }

  public removeResistance(index: number): void {
    this.resistancesArray.removeAt(index);
  }

  ngOnInit(): void {
    // Watch for level changes to update proficiency bonus
    this.characterForm.get('level')?.valueChanges.subscribe(level => {
      if (level !== null) {
        const pb = Math.floor((level - 1) / 4) + 2;
        this.characterForm.get('proficiencyBonus')?.setValue(pb, { emitEvent: false });
        this.calculateSpellcasting();
      }
    });

    // Watch for spellcasting changes
    this.characterForm.get('asConfig.abilityScores')?.valueChanges.subscribe(() => {
      this.calculateSpellcasting();
      this.updateHPModifier();
    });

    this.characterForm.get('asConfig.proficiencies')?.valueChanges.subscribe(() => {
      this.calculateSpellcasting();
    });

    this.characterForm.get('hp.numberOfDice')?.valueChanges.subscribe(() => this.updateHPModifier());
    this.characterForm.get('hp.hitDie')?.valueChanges.subscribe(() => this.updateHPDisplay());
    this.characterForm.get('hp.amountToAdd')?.valueChanges.subscribe(() => this.updateHPDisplay());

    this.characterForm.get('spellcasting.ability')?.valueChanges.subscribe(() => {
      this.calculateSpellcasting();
    });

    this.characterForm.get('class')?.valueChanges.subscribe(className => {
      // Logic below handles everything needed for class changes
    });

    this.characterForm.get('spellcasting.casterLevel')?.valueChanges.subscribe(() => {
      this.calculateSpellcasting();
    });

    this.characterForm.get('proficiencyBonus')?.valueChanges.subscribe(() => {
      this.calculateSpellcasting();
    });

    this.characterForm.get('raceId')?.valueChanges.subscribe(raceId => {
      const dragonbornId = this.mapperService.getRaceId(Race.Dragonborn);
      if (Number(raceId) !== dragonbornId) {
        this.characterForm.get('dragonbornColor')?.setValue(null, { emitEvent: false });
      }
    });

    this.characterForm.get('classId')?.valueChanges.subscribe(classId => {
      if (classId !== null && classId !== undefined) {
        this.calculateSpellcasting();
      }
    });

    // Watch for HP value changes to update state
    this.characterForm.get('hp.value')?.valueChanges.subscribe(val => {
      this.characterForm.get('state.currentHp')?.setValue(val, { emitEvent: false });
      this.characterForm.get('state.maxHp')?.setValue(val, { emitEvent: false });
    });
  }

  private updateHPModifier(): void {
    const constitution = this.characterForm.get('asConfig.abilityScores.constitution')?.value || 10;
    const conMod = Math.floor((constitution - 10) / 2);
    const numDice = this.characterForm.get('hp.numberOfDice')?.value || 1;
    this.characterForm.get('hp.modifier')?.setValue(conMod * numDice, { emitEvent: false });
    this.updateHPDisplay();
  }

  private updateHPDisplay(): void {
    const hpGroup = this.characterForm.get('hp');
    if (hpGroup) {
      const numDice = hpGroup.get('numberOfDice')?.value || 0;
      const hitDie = hpGroup.get('hitDie')?.value || 0;
      const amountToAdd = hpGroup.get('amountToAdd')?.value || 0;
      const constitution = this.characterForm.get('asConfig.abilityScores.constitution')?.value || 10;
      const conMod = Math.floor((constitution - 10) / 2);

      const totalModifier = (conMod * numDice) + amountToAdd;
      const average = Math.floor(numDice * (hitDie / 2 + 0.5)) + totalModifier;

      hpGroup.get('value')?.setValue(average, { emitEvent: false });
      hpGroup.get('hpAverage')?.setValue(average, { emitEvent: false });
      hpGroup.get('modifier')?.setValue(conMod * numDice, { emitEvent: false });

      const modSign = totalModifier >= 0 ? '+' : '-';
      const absMod = Math.abs(totalModifier);
      this.hpDisplay.set(`${average} (${numDice}d${hitDie} ${modSign} ${absMod})`);
    }
  }

  private calculateSpellcasting(): void {
    const level = this.characterForm.get('level')?.value || 1;
    const proficiency = this.characterForm.get('proficiencyBonus')?.value || 2;
    const spellAbility = this.characterForm.get('spellcasting.ability')?.value as Ability;
    const scores = this.characterForm.get('asConfig.abilityScores')?.value;
    const proficiencies = this.characterForm.get('asConfig.proficiencies')?.value;

    if (spellAbility && scores) {
      const scoreKey = spellAbility.toLowerCase() as keyof AbilityScores;
      const score = scores[scoreKey] || 10;
      const mod = Math.floor((score - 10) / 2);
      const isProficient = proficiencies ? proficiencies[scoreKey] : false;
      const appliedProficiency = isProficient ? proficiency : 0;

      const isSpellcaster = this.isSpellcaster();

      this.characterForm.get('spellcasting')?.patchValue({
        casterLevel: level,
        casterType: isSpellcaster ? CasterType.Full : CasterType.None,
        spellSaveDC: 8 + appliedProficiency + mod,
        spellAttackBonus: appliedProficiency + mod
      }, { emitEvent: false });
    }
  }

  protected override getEntityType(): CustomEntityType {
    return 'characters';
  }

  private createEmptySpellSlots(): Record<number, FormGroup> {
    const slots: Record<number, FormGroup> = {};
    for (let i = 1; i <= 9; i++) {
      slots[i] = this.fb.group({ current: [0], max: [0] });
    }
    return slots;
  }

  private resetForm() {
    this.resistancesArray.clear();
    this.characterForm.reset({
      raceId: this.mapperService.getRaceId(Race.Human),
      dragonbornColor: null,
      classId: this.mapperService.getClassId(Class.Fighter),
      level: 1,
      proficiencyBonus: 2,
      hp: {
        hpSetMethod: 0,
        value: 10,
        hpAverage: 10,
        numberOfDice: 1,
        hitDie: 10,
        amountToAdd: 0,
        modifier: 0
      },
      asConfig: {
        abilityScores: { strength: 10, dexterity: 10, constitution: 10, intelligence: 10, wisdom: 10, charisma: 10 },
        proficiencies: { strength: false, dexterity: false, constitution: false, intelligence: false, wisdom: false, charisma: false }
      },
      spellcasting: {
        casterType: CasterType.None,
        ability: Ability.Intelligence,
        casterLevel: 1,
        spellSaveDC: 10,
        spellAttackBonus: 0,
        spellSlots: this.createEmptySpellSlots(),
        spells: []
      },
      state: {
        currentHp: 10,
        maxHp: 10,
        tempHp: 0,
        hitDie: 10,
        conditions: {},
        deathSaves: { successes: 0, failures: 0 },
        isStable: true,
        isDead: false,
        initiative: 0
      }
    });
  }

  onSave(): void {
    if (this.characterForm.valid) {
      const character = this.prepareCharacterObject();
      this.saveEntity(character);
    } else {
      this.characterForm.markAllAsTouched();
    }
  }

  /**
   * Helper to prepare the character object from form values.
   * Useful for both saving and testing/exporting.
   */
  private prepareCharacterObject(): Character {
    const rawValues = this.characterForm.getRawValue();
    const character = { ...rawValues };

    // Ensure isCustom is set
    character.isCustom = true;

    // Ensure AC is removed (calculated from equipment/stats)
    delete character.ac;

    // Ensure dragonbornColor is removed if not Dragonborn
    const dragonbornId = this.mapperService.getRaceId(Race.Dragonborn);
    if (Number(character.raceId) !== dragonbornId) {
      delete character.dragonbornColor;
    }

    // Ensure isSpellcaster is removed if it somehow got in
    delete character.isSpellcaster;

    // Convert resistances array to object
    const resistanceObj: Record<string, ResistanceType> = {};
    if (character.state.resistances && Array.isArray(character.state.resistances)) {
      character.state.resistances.forEach((res: { damageType: string; resistanceType: ResistanceType }) => {
        resistanceObj[res.damageType] = res.resistanceType;
      });
    }
    character.state.resistances = resistanceObj;

    // Transform equipment form values to Equipment model
    const eqForm = rawValues.equipment;
    const equipment: Equipment = {
      armorId: eqForm.armorId || undefined,
      hasShieldEquipped: eqForm.hasShieldEquipped || false,
      primarySlot: eqForm.primaryWeaponId ? [{ weaponId: eqForm.primaryWeaponId, isProficient: true, modifiers: { attackBonus: 0, damageBonus: 0, isMagic: false, isSilvered: false, isAdamantine: false, isColdForgedIron: false } }] : [],
      secondarySlot: eqForm.secondaryWeaponId ? [{ weaponId: eqForm.secondaryWeaponId, isProficient: true, modifiers: { attackBonus: 0, damageBonus: 0, isMagic: false, isSilvered: false, isAdamantine: false, isColdForgedIron: false } }] : [],
      rangedSlot: eqForm.rangedWeaponId ? [{ weaponId: eqForm.rangedWeaponId, isProficient: true, modifiers: { attackBonus: 0, damageBonus: 0, isMagic: false, isSilvered: false, isAdamantine: false, isColdForgedIron: false } }] : [],
    };
    character.equipment = equipment;

    if (!this.itemToEdit()) {
      character.state.currentHp = character.hp.value;
      character.state.maxHp = character.hp.value;
      character.state.tempHp = 0;
      character.state.hitDie = character.hp.hitDie;
      character.state.conditions = {};
      character.state.deathSaves = { successes: 0, failures: 0 };
      character.state.isStable = true;
      character.state.isDead = false;
      character.state.initiative = 0;
    } else {
      // Update state if HP changed
      character.state.maxHp = character.hp.value;
      if (character.state.currentHp > character.state.maxHp) {
        character.state.currentHp = character.state.maxHp;
      }
      character.state.hitDie = character.hp.hitDie;
    }

    return character;
  }

  /**
   * Exports the current character as a JSON file for testing.
   */
  exportToJson(): void {
    const character = this.prepareCharacterObject();
    const serializedCharacter = this.mapperService.serializeKeys(character);
    const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(serializedCharacter, null, 2));
    const downloadAnchorNode = document.createElement('a');
    downloadAnchorNode.setAttribute("href", dataStr);
    downloadAnchorNode.setAttribute("download", `${character.name || 'character'}.json`);
    document.body.appendChild(downloadAnchorNode);
    downloadAnchorNode.click();
    downloadAnchorNode.remove();
  }
}
