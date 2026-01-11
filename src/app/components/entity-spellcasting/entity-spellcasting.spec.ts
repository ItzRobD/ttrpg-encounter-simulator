import { ComponentFixture, TestBed } from '@angular/core/testing';
import { EntitySpellcasting } from './entity-spellcasting';
import { Spellcasting, CasterType } from '../../models';

describe('EntitySpellcasting', () => {
  let component: EntitySpellcasting;
  let fixture: ComponentFixture<EntitySpellcasting>;

  const mockSpellcasting: Spellcasting = {
    casterType: CasterType.Full,
    casterLevel: 5,
    spellSaveDC: 15,
    spellAttackBonus: 7,
    spellSlots: {
      1: { current: 4, max: 4 },
      2: { current: 3, max: 3 },
      3: { current: 2, max: 2 },
    },
    spells: [
      {
        id: 1, name: 'Fireball', level: 3, description: 'Boom', isConcentration: false,
        castingTime: 'action', spellType: 'damage', isAOE: true, hasDC: true, isAutoHit: false
      },
      {
        id: 2, name: 'Cure Wounds', level: 1, description: 'Heal', isConcentration: false,
        castingTime: 'action', spellType: 'healing', isAOE: false, hasDC: false, isAutoHit: true
      },
      {
        id: 3, name: 'Shield', level: 1, description: 'Defense', isConcentration: false,
        castingTime: 'reaction', spellType: 'healing', isAOE: false, hasDC: false, isAutoHit: true
      },
      {
        id: 4, name: 'Light', level: 0, description: 'Glow', isConcentration: false,
        castingTime: 'action', spellType: 'healing', isAOE: false, hasDC: false, isAutoHit: true
      },
    ],
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EntitySpellcasting]
    })
    .compileComponents();

    fixture = TestBed.createComponent(EntitySpellcasting);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.componentRef.setInput('spellcasting', mockSpellcasting);
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should render spell DC and attack bonus', () => {
    fixture.componentRef.setInput('spellcasting', mockSpellcasting);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Save DC');
    expect(compiled.textContent).toContain('15');
    expect(compiled.textContent).toContain('Attack Bonus');
    expect(compiled.textContent).toContain('+7');
  });

  it('should group spells by level', () => {
    fixture.componentRef.setInput('spellcasting', mockSpellcasting);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Cantrips');
    expect(compiled.textContent).toContain('Level 1');
    expect(compiled.textContent).toContain('Level 3');
    expect(compiled.textContent).toContain('Fireball');
    expect(compiled.textContent).toContain('Cure Wounds');
    expect(compiled.textContent).toContain('Light');
  });
});
