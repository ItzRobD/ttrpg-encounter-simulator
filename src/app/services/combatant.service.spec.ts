import { TestBed } from '@angular/core/testing';
import { CombatantService } from './combatant.service';
import { Entity, Race, Class } from '../models';
import { LocalStorageService } from './local-storage.service';
import { environment } from '../../environments/environment';
import { vi } from 'vitest';

describe('CombatantService', () => {
  let service: CombatantService;
  let localStorageSpy: any;

  const mockMonster: Entity = {
    id: 1,
    instanceId: 0,
    name: 'Goblin',
    abilityScores: { strength: 8, dexterity: 14, constitution: 10, intelligence: 10, wisdom: 8, charisma: 8 },
    abilityScoreProficiency: { strength: false, dexterity: false, constitution: false, intelligence: false, wisdom: false, charisma: false },
    state: { currentHp: 7, maxHp: 7, tempHp: 0, hitDie: 6, conditions: {} as any, deathSaves: { successes: 0, failures: 0 }, resistances: {} as any, isStable: true, isDead: false, initiative: 0 }
  };

  const mockCharacter = {
    ...mockMonster,
    name: 'Fighter',
    class: Class.Fighter,
    race: Race.Human,
    level: 1
  } as any;

  beforeEach(() => {
    localStorageSpy = {
      getItem: vi.fn().mockReturnValue(null),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn()
    };

    TestBed.configureTestingModule({
      providers: [
        CombatantService,
        { provide: LocalStorageService, useValue: localStorageSpy }
      ]
    });
    service = TestBed.inject(CombatantService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should start with an empty list if localStorage is empty', () => {
    expect(service.combatants().length).toBe(0);
    expect(service.count()).toBe(0);
  });

  it('should add a combatant and assign a unique instanceId', () => {
    const success = service.addCombatant(mockMonster);
    expect(success).toBe(true);
    expect(service.combatants().length).toBe(1);
    expect(service.combatants()[0].instanceId).toBe(1);

    service.addCombatant(mockMonster);
    expect(service.combatants().length).toBe(2);
    expect(service.combatants()[1].instanceId).toBe(2);
  });

  it('should respect limits and allow fluid monster counts', () => {
    // maxTotal is 23. maxCharacters is 8.
    // If we have 0 characters, we should be able to add 23 monsters.
    const maxTotal = environment.limits.maxTotal;

    for (let i = 0; i < maxTotal; i++) {
      expect(service.addCombatant(mockMonster)).toBe(true);
    }

    expect(service.count()).toBe(maxTotal);
    expect(service.monsters().length).toBe(maxTotal); // All 23 slots filled by monsters

    // Cannot add more
    expect(service.addCombatant(mockMonster)).toBe(false);
  });

  it('should respect character specific limits', () => {
    const maxChars = environment.limits.maxCharacters;
    for (let i = 0; i < maxChars; i++) {
      service.addCombatant(mockCharacter);
    }
    expect(service.characters().length).toBe(maxChars);
    const added = service.addCombatant(mockCharacter);
    expect(added).toBe(false);
  });

  it('should remove a combatant by instanceId', () => {
    service.addCombatant(mockMonster); // instanceId 1
    service.addCombatant(mockMonster); // instanceId 2

    service.removeCombatant(1);
    expect(service.combatants().length).toBe(1);
    expect(service.combatants()[0].instanceId).toBe(2);
  });

  it('should filter monsters and characters correctly', () => {
    service.addCombatant(mockMonster);
    service.addCombatant(mockCharacter);

    expect(service.monsters().length).toBe(1);
    expect(service.characters().length).toBe(1);
  });

  it('should reorder combatants', () => {
    service.addCombatant({ ...mockMonster, name: 'First' });
    service.addCombatant({ ...mockMonster, name: 'Second' });

    service.reorderCombatant(0, 1);
    expect(service.combatants()[0].name).toBe('Second');
    expect(service.combatants()[1].name).toBe('First');
  });

  it('should sort by initiative descending', () => {
    service.addCombatant(mockMonster);
    service.addCombatant(mockMonster);

    const id1 = service.combatants()[0].instanceId;
    const id2 = service.combatants()[1].instanceId;

    service.updateCombatant(id1, { state: { ...mockMonster.state, initiative: 10 } });
    service.updateCombatant(id2, { state: { ...mockMonster.state, initiative: 20 } });

    service.sortByInitiative();

    expect(service.combatants()[0].instanceId).toBe(id2);
    expect(service.combatants()[1].instanceId).toBe(id1);
  });

  it('should clear the encounter', () => {
    service.addCombatant(mockMonster);
    service.clearEncounter();
    expect(service.combatants().length).toBe(0);
  });
});
