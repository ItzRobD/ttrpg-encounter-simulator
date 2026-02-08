import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActorStats } from './actor-stats.component';
import { Entity, Race, Class, DamageType, ResistanceType, DiceType } from '../../models';

describe('ActorStats', () => {
  let component: ActorStats;
  let fixture: ComponentFixture<ActorStats>;

  const baseState = {
    currentHp: 7,
    maxHp: 7,
    tempHp: 0,
    hitDie: 6,
    conditions: {
      blinded: false, charmed: false, deafened: false, exhaustion: 0, frightened: false,
      grappled: false, incapacitated: false, invisible: false, paralyzed: false,
      petrified: false, poisoned: false, prone: false, restrained: false,
      stunned: false, unconscious: false
    } as any,
    deathSaves: { successes: 0, failures: 0 },
    resistances: Object.values(DamageType).reduce((acc, type) => {
      acc[type] = ResistanceType.None;
      return acc;
    }, {} as any),
    isStable: true,
    isDead: false,
    initiative: 0
  };

  const mockMonster: Entity = {
    id: 1,
    instanceId: 1,
    name: 'Goblin',
    abilityScores: { strength: 8, dexterity: 14, constitution: 10, intelligence: 10, wisdom: 8, charisma: 8 },
    abilityScoreProficiency: { strength: false, dexterity: true, constitution: false, intelligence: false, wisdom: false, charisma: false },
    state: {
      ...baseState,
      resistances: { ...baseState.resistances, [DamageType.Fire]: ResistanceType.Resistant }
    }
  };

  const mockCharacter: Entity = {
    ...mockMonster,
    name: 'Fighter',
    class: Class.Fighter,
    race: Race.Human,
    level: 1,
    state: {
      ...baseState,
      deathSaves: { successes: 2, failures: 1 },
      resistances: {
        ...baseState.resistances,
        [DamageType.Cold]: ResistanceType.Immune,
        [DamageType.Acid]: ResistanceType.Vulnerable
      },
    }
  } as any;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ActorStats]
    })
    .compileComponents();

    fixture = TestBed.createComponent(ActorStats);
    component = fixture.componentInstance;
  });

  it('should create', () => {
    fixture.componentRef.setInput('entity', mockMonster);
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });

  it('should render ability scores and modifiers', () => {
    fixture.componentRef.setInput('entity', mockMonster);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Str');
    expect(compiled.textContent).toContain('8');
    expect(compiled.textContent).toContain('(-1)');

    expect(compiled.textContent).toContain('Dex');
    expect(compiled.textContent).toContain('14');
    expect(compiled.textContent).toContain('(+2)');
  });

  it('should highlight proficient ability scores', () => {
    fixture.componentRef.setInput('entity', mockMonster);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    const dexProficiencyIcon = compiled.querySelector('.pi-star-fill');
    expect(dexProficiencyIcon).toBeTruthy();
  });

  it('should render death saves for characters', () => {
    fixture.componentRef.setInput('entity', mockCharacter);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Death Saves');
  });

  it('should NOT render death saves for monsters', () => {
    fixture.componentRef.setInput('entity', mockMonster);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).not.toContain('Death Saves');
  });

  it('should render resistances, immunities, and vulnerabilities', () => {
    fixture.componentRef.setInput('entity', mockCharacter);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Immunities:');
    expect(compiled.textContent).toContain('Cold');
    expect(compiled.textContent).toContain('Vulnerabilities:');
    expect(compiled.textContent).toContain('Acid');
  });

  it('should use projectedState if provided', () => {
    fixture.componentRef.setInput('entity', mockMonster);
    const projectedState = {
      ...mockMonster.state,
      resistances: { ...baseState.resistances, [DamageType.Lightning]: ResistanceType.Resistant }
    };
    fixture.componentRef.setInput('projectedState', projectedState);
    fixture.detectChanges();

    const compiled = fixture.nativeElement as HTMLElement;
    expect(compiled.textContent).toContain('Resistances:');
    expect(compiled.textContent).toContain('Lightning');
    expect(compiled.textContent).not.toContain('Fire');
  });
});
