import { ComponentFixture, TestBed } from '@angular/core/testing';
import { EntitySpecialAbilities } from './entity-special-abilities';
import { Monster, MonsterType, MonsterSize, DamageType, DiceType } from '../../models';

describe('EntitySpecialAbilities', () => {
  let component: EntitySpecialAbilities;
  let fixture: ComponentFixture<EntitySpecialAbilities>;

  const emptySpecialAbilities = {
    assassinate: false, berserkThreshold: 0, bloodFrenzy: false, consumeLifeDie: DiceType.D0,
    corrosiveFormNumDice: 0, deathBurstNumDice: 0, deathBurstDamageType: DamageType.Acid,
    deathBurstDC: 0, deathThroesNumDice: 0, deathThroesDC: 0, divineEminenceNumDice: 0,
    evasion: false, fireAuraNumDice: 0, fireForm: false, gibbering: false,
    gnomeCunning: false, heatedBodyNumDice: 0, legendaryResistanceCount: 0,
    lightningAbsorption: false, limitedMagicImmunityLevel: 0, magicResistance: false,
    magicWeapons: false, martialAdvantageNumDice: 0, packTactics: false,
    reckless: false, reflectiveCarapace: false, regenerationValue: 0,
    relentlessThreshold: 0, sneakAttackNumDice: 0, undeadFortitude: false
  };

  const mockMonster: Monster = {
    id: 1,
    instanceId: 1,
    name: 'Goblin',
    cr: 0.25,
    type: MonsterType.Humanoid,
    size: MonsterSize.Small,
    proficiencyBonus: 2,
    isLegendary: false,
    isSpellcaster: false,
    isInnateSpellcaster: false,
    ac: 15,
    hp: {
      hpSetMethod: 0,
      value: 0,
      hpAverage: 7,
      numberOfDice: 2,
      hitDie: 6,
      amountToAdd: 0,
      modifier: 0,
    },
    abilityScores: {} as any,
    abilityScoreProficiency: {} as any,
    specialAbilities: {
      ...emptySpecialAbilities,
      packTactics: true
    },
    monsterActions: {
      actions: [],
      multiattacks: [],
      legendaryActions: [],
      rechargeActions: {}
    },
    state: {} as any
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EntitySpecialAbilities]
    })
    .compileComponents();

    fixture = TestBed.createComponent(EntitySpecialAbilities);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.componentRef.setInput('entity', mockMonster);
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should render special abilities', () => {
    fixture.componentRef.setInput('entity', mockMonster);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Pack Tactics');
  });

  it('should render nothing if entity has no special abilities', () => {
    const monsterNoAbilities = { ...mockMonster, specialAbilities: emptySpecialAbilities };
    fixture.componentRef.setInput('entity', monsterNoAbilities);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    const grid = compiled.querySelector('.grid');
    expect(grid?.children.length).toBe(0);
  });
});
