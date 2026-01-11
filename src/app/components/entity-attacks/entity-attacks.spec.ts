import { ComponentFixture, TestBed } from '@angular/core/testing';
import { EntityAttacks } from './entity-attacks';
import { Monster, MonsterType, MonsterSize, DamageType, DiceType } from '../../models';

describe('EntityAttacks', () => {
  let component: EntityAttacks;
  let fixture: ComponentFixture<EntityAttacks>;

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
    abilityScores: {} as any,
    abilityScoreProficiency: {} as any,
    specialAbilities: {} as any,
    monsterActions: {
      actions: [
        {
          actionId: 1, name: 'Scimitar',
          description: 'Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) slashing damage.',
          rechargeValue: 0, hasDC: false, index: 0, numberOfDice: 1, die: DiceType.D6, amountToAdd: 2, attackBonus: 4, damageType: DamageType.Slashing
        }
      ],
      multiattacks: [
        [{ actionId: 1, count: 2 }]
      ],
      legendaryActions: [],
      rechargeActions: {}
    },
    state: {} as any
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EntityAttacks]
    })
    .compileComponents();

    fixture = TestBed.createComponent(EntityAttacks);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.componentRef.setInput('monster', mockMonster);
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should render monster actions', () => {
    fixture.componentRef.setInput('monster', mockMonster);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Scimitar');
  });

  it('should render multiattacks', () => {
    fixture.componentRef.setInput('monster', mockMonster);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Multiattack');
    // Multiattack rendering depends on formatMultiattack utility, but it should show "2x Scimitar" or similar
    expect(compiled.textContent).toContain('Scimitar');
  });
});
