import { Component, effect, OnInit, inject, signal, ChangeDetectionStrategy, DestroyRef } from '@angular/core';
import { toSignal, takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { AbstractControl, FormArray, FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { DialogModule } from 'primeng/dialog';
import { ButtonModule } from 'primeng/button';
import { SplitButtonModule } from 'primeng/splitbutton';
import { InputTextModule } from 'primeng/inputtext';
import { InputNumberModule } from 'primeng/inputnumber';
import { TextareaModule } from 'primeng/textarea';
import { SelectModule } from 'primeng/select';
import { CheckboxModule } from 'primeng/checkbox';
import { MessageModule } from 'primeng/message';
import { TooltipModule } from 'primeng/tooltip';
import { TabsModule } from 'primeng/tabs';
import { FluidModule } from 'primeng/fluid';
import { BaseEditorDirective } from '../base-editor.directive';
import { MonsterSize, MonsterType, DiceType, DamageType, CasterType, Ability, ResistanceType, Actor, Action, DamageComponent, Conditions, ActorState, DamageResistances } from '../../../models';
import { CustomActorType } from '../../../services/custom-content.service';
import { AbilityScoreEditorComponent } from '../ability-score-editor/ability-score-editor';
import { SpellcastingEditorComponent } from '../spellcasting-editor/spellcasting-editor';
import { getProficiencyBonus } from '../../../shared/utils/dnd-utils';
import { environment } from '../../../../environments/environment';
import { MapperService } from '../../../services/mapper.service';

/** Loose shape of a single damage block as stored in the editor's form data. */
interface DamageBlockFormData extends Partial<DamageComponent> {}

/**
 * Shape of action/legendary-action data fed into the editor's FormArrays. This
 * is the editor's flat form representation (damageBlocks, plus `cost` for
 * legendary actions) rather than the backend `core.Action` shape.
 */
interface ActionFormData {
  actionId?: string | number;
  name?: string;
  attackBonus?: number;
  rechargeValue?: number;
  cost?: number;
  hasDC?: boolean;
  dc?: number;
  dcAbility?: Ability;
  dcOnSuccess?: string;
  description?: string;
  damageBlocks?: DamageBlockFormData[];
}

/**
 * A monster being edited may arrive with its actions either grouped under a
 * legacy `monsterActions` object or flat on `actions`/`legendaryActions`.
 * The backend itself only sends `actions` (see pkg/actor/actor_config.go); the
 * grouped form is a local/legacy shape the editor still tolerates.
 */
interface EditableMonster {
  monsterActions?: { actions?: ActionFormData[]; legendaryActions?: ActionFormData[] };
  actions?: ActionFormData[];
  legendaryActions?: ActionFormData[];
}

@Component({
  selector: 'app-monster-editor',
  imports: [
    ReactiveFormsModule,
    DialogModule,
    ButtonModule,
    SplitButtonModule,
    InputTextModule,
    InputNumberModule,
    TextareaModule,
    SelectModule,
    CheckboxModule,
    MessageModule,
    TooltipModule,
    TabsModule,
    FluidModule,
    AbilityScoreEditorComponent,
    SpellcastingEditorComponent
  ],
  templateUrl: './monster-editor.html',
  styleUrl: './monster-editor.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class MonsterEditorComponent extends BaseEditorDirective<Actor> implements OnInit {
  private readonly fb = inject(FormBuilder);
  protected readonly mapperService = inject(MapperService);
  private readonly destroyRef = inject(DestroyRef);

  public monsterForm: FormGroup = this.fb.group({
    id: [null],
    name: ['', [Validators.required]],
    size: [MonsterSize.Medium, [Validators.required]],
    type: [MonsterType.Humanoid, [Validators.required]],
    cr: [1, [Validators.required]],
    ac: [10, [Validators.required, Validators.min(0), Validators.max(30)]],
    proficiencyBonus: [2],
    hp: this.fb.group({
      hpSetMethod: [0], // SetValue
      value: [10, [Validators.required, Validators.min(1)]],
      hpAverage: [10],
      numberOfDice: [1],
      hitDie: [8],
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
    isLegendary: [false],
    isSpellcaster: [false],
    isInnateSpellcaster: [false],
    spellcasting: this.fb.group({
      casterType: [CasterType.None],
      ability: [Ability.Intelligence],
      casterLevel: [1],
      spellSaveDC: [10],
      spellAttackBonus: [0],
      spellSlots: this.fb.group(this.createEmptySpellSlots()),
      spells: [[]]
    }),
    specialAbilities: this.fb.group({
      assassinate: [false],
      berserkThreshold: [0],
      bloodFrenzy: [false],
      consumeLifeDie: [DiceType.D0],
      corrosiveFormNumDice: [0],
      deathBurstNumDice: [0],
      deathBurstDamageType: [DamageType.Acid],
      deathBurstDC: [10],
      deathThroesNumDice: [0],
      deathThroesDC: [10],
      divineEminenceNumDice: [0],
      evasion: [false],
      fireAuraNumDice: [0],
      fireForm: [false],
      gibbering: [false],
      gnomeCunning: [false],
      heatedBodyNumDice: [0],
      legendaryResistanceCount: [0],
      lightningAbsorption: [false],
      limitedMagicImmunityLevel: [0],
      magicResistance: [false],
      magicWeapons: [false],
      martialAdvantageNumDice: [0],
      packTactics: [false],
      reckless: [false],
      reflectiveCarapace: [false],
      regenerationValue: [0],
      relentlessThreshold: [0],
      sneakAttackNumDice: [0],
      undeadFortitude: [false]
    }),
    behavior: this.fb.group({
      actionPreference: ['balanced'],
      targetPriority: ['no_preference'],
      secondaryActionPreference: ['balanced'],
      secondaryTargetPriority: ['no_preference']
    }),
    monsterActions: this.fb.group({
      actions: this.fb.array([]),
      multiattacks: [[]],
      legendaryActions: this.fb.array([]),
      rechargeActions: [{}]
    }),
    state: this.fb.group({
      currentHp: [10],
      maxHp: [10],
      tempHp: [0],
      hitDie: [8],
      conditions: [{}],
      deathSaves: this.fb.group({ successes: [0], failures: [0] }),
      resistances: this.fb.array([]),
      isStable: [true],
      isDead: [false],
      initiative: [0]
    })
  });

  protected readonly sizes = Object.values(MonsterSize).map(s => ({ label: s, value: s }));
  protected readonly types = Object.values(MonsterType).map(t => ({ label: t, value: t }));
  protected readonly crOptions = [
    { label: '0', value: 0 },
    { label: '1/8', value: 0.125 },
    { label: '1/4', value: 0.25 },
    { label: '1/2', value: 0.5 },
    ...Array.from({ length: 30 }, (_, i) => ({ label: (i + 1).toString(), value: i + 1 }))
  ];
  protected readonly environment = environment;
  protected readonly diceOptions = [
    { label: 'd4', value: DiceType.D4 },
    { label: 'd6', value: DiceType.D6 },
    { label: 'd8', value: DiceType.D8 },
    { label: 'd10', value: DiceType.D10 },
    { label: 'd12', value: DiceType.D12 },
    { label: 'd20', value: DiceType.D20 }
  ];
  protected readonly damageTypes = Object.values(DamageType).map(t => ({ label: t.charAt(0).toUpperCase() + t.slice(1), value: t }));
  protected readonly resistanceTypes = Object.values(ResistanceType)
    .filter(rt => rt !== ResistanceType.None)
    .map(rt => ({ label: rt.charAt(0).toUpperCase() + rt.slice(1), value: rt }));
  protected readonly abilities = Object.values(Ability).map(a => ({ label: a.charAt(0).toUpperCase() + a.slice(1), value: a }));

  protected readonly actionPreferences = [
    { label: 'Melee', value: 'melee' },
    { label: 'Ranged', value: 'ranged' },
    { label: 'Spell', value: 'spell' },
    { label: 'Balanced', value: 'balanced' }
  ];

  protected readonly targetPriorities = [
    { label: 'Lowest HP', value: 'lowest_hp' },
    { label: 'Highest Threat', value: 'highest_threat' },
    { label: 'Spellcaster', value: 'spellcaster' },
    { label: 'No Preference', value: 'no_preference' }
  ];

  get actions(): FormArray {
    return this.monsterForm.get('monsterActions.actions') as FormArray;
  }

  get legendaryActions(): FormArray {
    return this.monsterForm.get('monsterActions.legendaryActions') as FormArray;
  }

  get resistancesArray(): FormArray {
    return this.monsterForm.get('state.resistances') as FormArray;
  }

  get asConfigGroup(): FormGroup {
    return this.monsterForm.get('asConfig') as FormGroup;
  }

  get spellcastingGroup(): FormGroup {
    return this.monsterForm.get('spellcasting') as FormGroup;
  }

  get behaviorGroup(): FormGroup {
    return this.monsterForm.get('behavior') as FormGroup;
  }

  get specialAbilitiesGroup(): FormGroup {
    return this.monsterForm.get('specialAbilities') as FormGroup;
  }

  public hpDisplay = signal<string>('10 (1d8 + 0)');

  constructor() {
    super();
    effect(() => {
      const item = this.itemToEdit();
      if (item) {
        // Clear dynamic arrays
        this.actions.clear({ emitEvent: false });
        this.legendaryActions.clear({ emitEvent: false });
        this.resistancesArray.clear({ emitEvent: false });

        const editable = item as Actor & EditableMonster;
        const monsterActions = editable.monsterActions ?? {
          actions: editable.actions ?? [],
          legendaryActions: editable.legendaryActions ?? []
        };

        // Add actions
        if (monsterActions.actions) {
          monsterActions.actions.forEach((a: ActionFormData) => this.addAction(a));
        }
        if (monsterActions.legendaryActions) {
          monsterActions.legendaryActions.forEach((la: ActionFormData) => this.addLegendaryAction(la));
        }

        // Handle resistances array
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

        // Initialize spell slots if missing
        if (item.spellcasting && !item.spellcasting.spellSlots) {
          item.spellcasting.spellSlots = {};
        }

        this.monsterForm.patchValue(item, { emitEvent: false });
        this.updateHPDisplay();
      } else {
        this.resetForm();
      }
    });

    effect(() => {
      const hpGroup = this.monsterForm.get('hp');
      const conCtrl = this.monsterForm.get('asConfig.abilityScores.constitution');
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
    this.monsterForm.get('cr')?.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(cr => {
      if (cr !== null) {
        this.monsterForm.patchValue({ proficiencyBonus: getProficiencyBonus(cr) }, { emitEvent: false });
        this.calculateSpellcasting();
      }
    });

    this.monsterForm.get('hp')?.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      this.updateHPDisplay();
    });

    this.monsterForm.get('asConfig.abilityScores')?.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      this.calculateSpellcasting();
      this.updateHPModifier();
    });

    this.monsterForm.get('asConfig.proficiencies')?.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      this.calculateSpellcasting();
    });

    this.monsterForm.get('spellcasting.ability')?.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      this.calculateSpellcasting();
    });

    this.monsterForm.get('isSpellcaster')?.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      this.calculateSpellcasting();
    });

    this.monsterForm.get('isInnateSpellcaster')?.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      this.calculateSpellcasting();
    });

    this.monsterForm.get('spellcasting.casterLevel')?.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      this.calculateSpellcasting();
    });

    this.monsterForm.get('proficiencyBonus')?.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => {
      this.calculateSpellcasting();
    });
  }

  private updateHPModifier(): void {
    const constitution = this.monsterForm.get('asConfig.abilityScores.constitution')?.value || 10;
    const conMod = Math.floor((constitution - 10) / 2);
    const numDice = this.monsterForm.get('hp.numberOfDice')?.value || 1;
    this.monsterForm.get('hp.modifier')?.setValue(conMod * numDice, { emitEvent: false });
    this.updateHPDisplay();
  }

  private updateHPDisplay(): void {
    const hpGroup = this.monsterForm.get('hp');
    if (hpGroup) {
      const numDice = hpGroup.get('numberOfDice')?.value || 0;
      const hitDie = hpGroup.get('hitDie')?.value || 0;
      const amountToAdd = hpGroup.get('amountToAdd')?.value || 0;
      const constitution = this.monsterForm.get('asConfig.abilityScores.constitution')?.value || 10;
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
    const proficiency = this.monsterForm.get('proficiencyBonus')?.value || 2;
    const spellAbility = this.monsterForm.get('spellcasting.ability')?.value as Ability;
    const scores = this.monsterForm.get('asConfig.abilityScores')?.value;
    const proficiencies = this.monsterForm.get('asConfig.proficiencies')?.value;

    if (spellAbility && scores) {
      const score = scores[spellAbility] || 10;
      const mod = Math.floor((score - 10) / 2);
      const isProficient = proficiencies ? proficiencies[spellAbility] : false;
      const appliedProficiency = isProficient ? proficiency : 0;

      const isSpellcaster = this.monsterForm.get('isSpellcaster')?.value;
      const isInnate = this.monsterForm.get('isInnateSpellcaster')?.value;

      this.monsterForm.get('spellcasting')?.patchValue({
        casterType: isSpellcaster ? CasterType.Full : (isInnate ? CasterType.None : CasterType.None),
        spellSaveDC: 8 + appliedProficiency + mod,
        spellAttackBonus: appliedProficiency + mod
      }, { emitEvent: false });
    }
  }

  protected override getActorType(): CustomActorType {
    return 'monsters';
  }

  private createEmptySpellSlots(): Record<number, FormGroup> {
    const slots: Record<number, FormGroup> = {};
    for (let i = 1; i <= 9; i++) {
      slots[i] = this.fb.group({ current: [0], max: [0] });
    }
    return slots;
  }

  private resetForm() {
    this.actions.clear();
    this.legendaryActions.clear();
    this.resistancesArray.clear();
    this.monsterForm.reset({
      size: MonsterSize.Medium,
      type: MonsterType.Humanoid,
      cr: 1,
      ac: 10,
      proficiencyBonus: 2,
      hp: { hpSetMethod: 0, value: 10, hitDie: 8, numberOfDice: 1, amountToAdd: 0 },
      isLegendary: false,
      isSpellcaster: false,
      isInnateSpellcaster: false,
      asConfig: {
        abilityScores: { strength: 10, dexterity: 10, constitution: 10, intelligence: 10, wisdom: 10, charisma: 10 },
        proficiencies: { strength: false, dexterity: false, constitution: false, intelligence: false, wisdom: false, charisma: false }
      },
      spellcasting: { casterType: CasterType.None, ability: Ability.Intelligence, casterLevel: 1, spellSaveDC: 10, spellAttackBonus: 0, spells: [] },
      specialAbilities: {
        assassinate: false,
        berserkThreshold: 0,
        bloodFrenzy: false,
        consumeLifeDie: DiceType.D0,
        corrosiveFormNumDice: 0,
        deathBurstNumDice: 0,
        deathBurstDamageType: DamageType.Acid,
        deathBurstDC: 10,
        deathThroesNumDice: 0,
        deathThroesDC: 10,
        divineEminenceNumDice: 0,
        evasion: false,
        fireAuraNumDice: 0,
        fireForm: false,
        gibbering: false,
        gnomeCunning: false,
        heatedBodyNumDice: 0,
        legendaryResistanceCount: 0,
        lightningAbsorption: false,
        limitedMagicImmunityLevel: 0,
        magicResistance: false,
        magicWeapons: false,
        martialAdvantageNumDice: 0,
        packTactics: false,
        reckless: false,
        reflectiveCarapace: false,
        regenerationValue: 0,
        relentlessThreshold: 0,
        sneakAttackNumDice: 0,
        undeadFortitude: false
      },
      state: {
        currentHp: 10,
        maxHp: 10,
        tempHp: 0,
        hitDie: 8,
        conditions: {},
        deathSaves: { successes: 0, failures: 0 },
        isStable: true,
        isDead: false,
        initiative: 0
      }
    });
  }

  addAction(data: ActionFormData | null = null): void {
    const actionGroup = this.fb.group({
      actionId: [data?.actionId || Math.floor(Math.random() * 1000000)],
      name: [data?.name || '', [Validators.required]],
      attackBonus: [data?.attackBonus || 0],
      rechargeValue: [data?.rechargeValue || 0],
      hasDC: [data?.hasDC || false],
      dc: [data?.dc || 10],
      dcAbility: [data?.dcAbility || Ability.Dexterity],
      dcOnSuccess: [data?.dcOnSuccess || 'half'],
      description: [data?.description || ''],
      damageBlocks: this.fb.array([])
    });

    const components = actionGroup.get('damageBlocks') as FormArray;
    if (data?.damageBlocks && Array.isArray(data.damageBlocks)) {
      data.damageBlocks.forEach((c: DamageBlockFormData) => this.addDamageComponent(components, c));
    } else if (!data) {
      // Default component for new actions
      this.addDamageComponent(components);
    }

    this.actions.push(actionGroup);
  }

  addDamageComponent(array: FormArray, data: DamageBlockFormData | null = null): void {
    array.push(this.fb.group({
      numberOfDice: [data?.numberOfDice || 1],
      die: [data?.die || DiceType.D6],
      amountToAdd: [data?.amountToAdd || 0],
      damageType: [data?.damageType || DamageType.Bludgeoning]
    }));
  }

  removeDamageComponent(array: FormArray, index: number): void {
    array.removeAt(index);
  }

  removeAction(index: number): void {
    this.actions.removeAt(index);
  }

  addLegendaryAction(data: ActionFormData | null = null): void {
    const actionGroup = this.fb.group({
      actionId: [data?.actionId || Math.floor(Math.random() * 1000000)],
      name: [data?.name || '', [Validators.required]],
      cost: [data?.cost || 1, [Validators.min(1), Validators.max(3)]],
      attackBonus: [data?.attackBonus || 0],
      hasDC: [data?.hasDC || false],
      dc: [data?.dc || 10],
      description: [data?.description || ''],
      damageBlocks: this.fb.array([])
    });

    const components = actionGroup.get('damageBlocks') as FormArray;
    if (data?.damageBlocks && Array.isArray(data.damageBlocks)) {
      data.damageBlocks.forEach((c: DamageBlockFormData) => this.addDamageComponent(components, c));
    } else if (!data) {
      // Default component for new actions
      this.addDamageComponent(components);
    }

    this.legendaryActions.push(actionGroup);
  }

  getDamageBlocks(action: AbstractControl): FormArray {
    return action.get('damageBlocks') as FormArray;
  }

  removeLegendaryAction(index: number): void {
    this.legendaryActions.removeAt(index);
  }

  onSave(): void {
    if (this.monsterForm.valid) {
      const monster = this.prepareMonsterObject();
      this.saveActor(monster);
    } else {
      this.monsterForm.markAllAsTouched();
    }
  }

  private prepareMonsterObject(): Actor {
    const rawValues = this.monsterForm.getRawValue();

    // Create a clean Actor object
    const monster: Partial<Actor> = {
      id: rawValues.id,
      name: rawValues.name,
      actorType: 'monster',
      side: 'monster',
      ac: rawValues.ac,
      abilities: rawValues.asConfig,
      hpConfig: {
        ...rawValues.hp,
        value: rawValues.hp.value,
        modifier: rawValues.hp.modifier,
        hitDice: { [rawValues.hp.hitDie.toString()]: rawValues.hp.numberOfDice }
      },
      metadata: {
        cr: rawValues.cr,
        level: rawValues.cr, // Monsters often use CR as level for some calcs
        size: rawValues.size,
        type: rawValues.type,
        isLegendary: rawValues.isLegendary,
        maxLegendaryActions: rawValues.isLegendary ? 3 : 0, // Default to 3 if legendary
        spellcasterMetadata: {
          isSpellcaster: rawValues.isSpellcaster,
          isInnateCaster: rawValues.isInnateSpellcaster,
          spellcastingAbility: rawValues.spellcasting.ability,
          spellcastingLevel: rawValues.spellcasting.casterLevel
        }
      },
      behavior: rawValues.behavior,
      actions: [],
      features: [],
      isCustom: true
    };

    // Map actions
    if (rawValues.monsterActions?.actions) {
      monster.actions?.push(...rawValues.monsterActions.actions);
    }
    if (rawValues.monsterActions?.legendaryActions) {
      rawValues.monsterActions.legendaryActions.forEach((la: ActionFormData) => {
        monster.actions?.push({
          ...la,
          isLegendary: true
        } as Action);
      });
    }

    // Add multiattacks if any
    if (rawValues.monsterActions?.multiattacks?.length > 0) {
      monster.actions?.push({
        name: 'Multiattack',
        description: 'The monster makes multiple attacks.',
        multiattack: rawValues.monsterActions.multiattacks
      } as Action);
    }

    // Map special abilities to features
    const specialAbilities = rawValues.specialAbilities;
    Object.entries(specialAbilities).forEach(([key, value]) => {
      if (value === true || (typeof value === 'number' && value > 0)) {
        monster.features?.push({
          id: key,
          name: key.replace(/([A-Z])/g, ' $1').replace(/^./, str => str.toUpperCase()),
          description: `Has the ${key} trait.`,
          data: { value }
        });
      }
    });

    // Special handling for Consume Life Die
    if (specialAbilities.consumeLifeDie && specialAbilities.consumeLifeDie !== DiceType.D0) {
      let feature = monster.features?.find(f => f.id === 'consumeLife');
      if (!feature) {
        feature = { id: 'consumeLife', name: 'Consume Life', description: 'Has the Consume Life trait.', data: {} };
        monster.features?.push(feature);
      }
      if (feature.data) feature.data['die'] = specialAbilities.consumeLifeDie;
    }

    // Special handling for Death Burst/Throes
    if (specialAbilities.deathBurstNumDice > 0) {
      monster.features?.push({
        id: 'deathBurst',
        name: 'Death Burst',
        description: `Death Burst for ${specialAbilities.deathBurstNumDice} dice.`,
        data: { numDice: specialAbilities.deathBurstNumDice, damageType: specialAbilities.deathBurstDamageType, dc: specialAbilities.deathBurstDC }
      });
    }
    if (specialAbilities.deathThroesNumDice > 0) {
      monster.features?.push({
        id: 'deathThroes',
        name: 'Death Throes',
        description: `Death Throes for ${specialAbilities.deathThroesNumDice} dice.`,
        data: { numDice: specialAbilities.deathThroesNumDice, dc: specialAbilities.deathThroesDC }
      });
    }

    // Resistances. The form stores an array of {damageType, resistanceType};
    // DamageResistances is a full keyed map, so the partial object is narrowed.
    const resistanceObj: Partial<Record<DamageType, ResistanceType>> = {};
    if (rawValues.state?.resistances && Array.isArray(rawValues.state.resistances)) {
      rawValues.state.resistances.forEach((res: { damageType: DamageType; resistanceType: ResistanceType }) => {
        resistanceObj[res.damageType] = res.resistanceType;
      });
    }
    monster.resistances = resistanceObj as DamageResistances;

    // Spells
    if (monster.metadata?.spellcasterMetadata?.isSpellcaster || monster.metadata?.spellcasterMetadata?.isInnateCaster) {
      monster.knownSpellIDs = (rawValues.spellcasting.spells || []).map((s: { id: string | number }) => Number(s.id));
    }
    monster.customSpells = [];

    // Cleanup legacy fields not part of the Actor shape.
    const deletable = monster as Record<string, unknown>;
    delete deletable['spellcasting'];
    delete deletable['monsterActions'];
    delete deletable['hp'];
    delete deletable['asConfig'];

    // Final state
    if (!this.itemToEdit()) {
      monster.state = {
        currentHp: rawValues.hp.value,
        maxHp: rawValues.hp.value,
        tempHp: 0,
        hitDie: rawValues.hp.hitDie,
        conditions: {} as Conditions,
        deathSaves: { successes: 0, failures: 0 },
        resistances: resistanceObj as DamageResistances,
        isStable: true,
        isDead: false,
        initiative: 0
      };
    } else {
      monster.state = {
        ...this.itemToEdit()?.state,
        resistances: resistanceObj as DamageResistances
      } as ActorState;
    }

    return monster as Actor;
  }

  /**
   * Exports the current monster as a JSON file for testing.
   */
  exportToJSON(): void {
    const monster = this.prepareMonsterObject();
    const exportObj = { ...monster };
    delete (exportObj as Record<string, unknown>)['state'];

    const serializedMonster = this.mapperService.serializeKeys(exportObj);
    const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(serializedMonster, null, 2));
    const downloadAnchorNode = document.createElement('a');
    downloadAnchorNode.setAttribute("href", dataStr);
    downloadAnchorNode.setAttribute("download", `${monster.name || 'monster'}_v2.json`);
    document.body.appendChild(downloadAnchorNode);
    downloadAnchorNode.click();
    downloadAnchorNode.remove();
  }
}
