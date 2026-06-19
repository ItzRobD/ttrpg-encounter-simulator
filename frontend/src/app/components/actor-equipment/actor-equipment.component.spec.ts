import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActorEquipment } from './actor-equipment.component';
import { Character, Class, Race, WeaponSlot, DiceType, DamageType } from '../../models';
import { EquipmentService } from '../../services/equipment.service';
import { signal } from '@angular/core';

describe('ActorEquipment', () => {
  let component: ActorEquipment;
  let fixture: ComponentFixture<ActorEquipment>;
  let mockEquipmentService: any;

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
      primarySlot: [{ weaponId: 101, isProficient: true, modifiers: { attackBonus: 0, damageBonus: 0, isMagic: false } } as any],
      secondarySlot: [],
      rangedSlot: [{ weaponId: 102, isProficient: true, modifiers: { attackBonus: 0, damageBonus: 0, isMagic: false } } as any]
    },
    state: {} as any
  };

  beforeEach(async () => {
    mockEquipmentService = {
      summaries: signal([
        { id: 101, name: 'Longsword' },
        { id: 102, name: 'Longbow' }
      ])
    };

    await TestBed.configureTestingModule({
      imports: [ActorEquipment],
      providers: [
        { provide: EquipmentService, useValue: mockEquipmentService }
      ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ActorEquipment);
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
