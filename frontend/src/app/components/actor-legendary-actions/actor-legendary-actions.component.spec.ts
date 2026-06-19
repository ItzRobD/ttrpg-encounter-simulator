import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActorLegendaryActions } from './actor-legendary-actions.component';
import { Monster, MonsterType, MonsterSize, DamageType, DiceType } from '../../models';

describe('ActorLegendaryActions', () => {
  let component: ActorLegendaryActions;
  let fixture: ComponentFixture<ActorLegendaryActions>;

  const mockMonster: Monster = {
    id: 1,
    instanceId: 1,
    name: 'Adult Red Dragon',
    cr: 17,
    type: MonsterType.Dragon,
    size: MonsterSize.Huge,
    proficiencyBonus: 6,
    isLegendary: true,
    isSpellcaster: false,
    isInnateSpellcaster: false,
    ac: 19,
    hp: {
      hpSetMethod: 0,
      value: 0,
      hpAverage: 256,
      numberOfDice: 17,
      hitDie: 12,
      amountToAdd: 119,
      modifier: 7,
    },
    abilityScores: {} as any,
    abilityScoreProficiency: {} as any,
    specialAbilities: {} as any,
    monsterActions: {
      actions: [],
      multiattacks: [],
      legendaryActions: [
        {
          actionId: 10, name: 'Detect', description: 'The dragon makes a Wisdom (Perception) check.',
          rechargeValue: 0, hasDC: false, index: 0, numberOfDice: 0, die: DiceType.D0, amountToAdd: 0, attackBonus: 0, damageType: DamageType.Acid,
          cost: 1
        },
        {
          actionId: 11, name: 'Tail Attack', description: 'The dragon makes a tail attack.',
          rechargeValue: 0, hasDC: false, index: 1, numberOfDice: 1, die: DiceType.D8, amountToAdd: 0, attackBonus: 0, damageType: DamageType.Bludgeoning,
          cost: 1
        }
      ],
      rechargeActions: {}
    },
    state: {} as any
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ActorLegendaryActions]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ActorLegendaryActions);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.componentRef.setInput('monster', mockMonster);
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should render legendary actions', () => {
    fixture.componentRef.setInput('monster', mockMonster);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Detect');
    expect(compiled.textContent).toContain('Tail Attack');
  });
});
