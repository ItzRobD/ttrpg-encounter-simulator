import { ComponentFixture, TestBed } from '@angular/core/testing';
import { EntityEquipment } from './entity-equipment';
import { Character, Class, Race, WeaponSlot, DiceType, DamageType } from '../../models';

describe('EntityEquipment', () => {
  let component: EntityEquipment;
  let fixture: ComponentFixture<EntityEquipment>;

  const mockProperties = {
    isVersatile: false, isFinesse: false, isRanged: false, isHeavy: false,
    isTwoHanded: false, isLight: false, isThrown: false, isOnlyRanged: false
  };

  const mockModifiers = {
    isMagic: false, isSilvered: false, isAdamantine: false, isColdForgedIron: false,
    attackBonus: 0, damageBonus: 0
  };

  const mockCharacter: Character = {
    id: 1,
    instanceId: 1,
    name: 'Fighter',
    class: Class.Fighter,
    race: Race.Human,
    level: 1,
    abilityScores: {} as any,
    abilityScoreProficiency: {} as any,
    equipment: {
      hasShieldEquipped: false,
      weapons: {
        [WeaponSlot.Primary]: {
          name: 'Longsword', numberOfDice: 1, die: DiceType.D8,
          damageType: DamageType.Slashing, properties: mockProperties, modifiers: mockModifiers
        },
        [WeaponSlot.Secondary]: undefined,
        [WeaponSlot.Ranged]: {
          name: 'Longbow', numberOfDice: 1, die: DiceType.D8,
          damageType: DamageType.Piercing, properties: { ...mockProperties, isRanged: true, isOnlyRanged: true },
          modifiers: mockModifiers
        }
      }
    },
    state: {} as any
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EntityEquipment]
    })
    .compileComponents();

    fixture = TestBed.createComponent(EntityEquipment);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.componentRef.setInput('character', mockCharacter);
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should render equipped weapons', () => {
    fixture.componentRef.setInput('character', mockCharacter);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Longsword');
    expect(compiled.textContent).toContain('Longbow');
  });
});
