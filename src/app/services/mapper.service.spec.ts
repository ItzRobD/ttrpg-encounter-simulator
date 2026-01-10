import { TestBed } from '@angular/core/testing';
import { MapperService } from './mapper.service';
import { TimelineNode } from '../models/simulation.model';

describe('MapperService', () => {
  let service: MapperService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(MapperService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('mapKeys', () => {
    it('should convert PascalCase to camelCase and handle acronyms', () => {
      const raw = {
        ID: '123',
        SequenceID: 'seq-1',
        ParentID: 'parent-1',
        AC: 15,
        HP: 20,
        MaxHP: 30,
        ArmorClass: 18,
        Data: {
          ActorName: 'Bob',
          TargetID: '456'
        }
      };

      const mapped = service.mapKeys(raw) as any;

      expect(mapped.id).toBe('123');
      expect(mapped.sequenceId).toBe('seq-1');
      expect(mapped.parentId).toBe('parent-1');
      expect(mapped.ac).toBe(15);
      expect(mapped.hp).toBe(20);
      expect(mapped.maxHp).toBe(30);
      expect(mapped.armorClass).toBe(18);
      expect(mapped.data.actorName).toBe('Bob');
      expect(mapped.data.targetId).toBe('456');
    });

    it('should handle arrays recursively', () => {
      const raw = [
        { ID: '1', Name: 'A' },
        { ID: '2', Name: 'B' }
      ];
      const mapped = service.mapKeys(raw) as any[];
      expect(mapped[0].id).toBe('1');
      expect(mapped[0].name).toBe('A');
      expect(mapped[1].id).toBe('2');
      expect(mapped[1].name).toBe('B');
    });
  });

  describe('mapSimulationLog', () => {
    it('should build a hierarchical tree by Round and Turn', () => {
      const rawLog = [
        {
          ID: 'init-1',
          Round: 0,
          Type: 'initiative',
          Data: { Actor: { Name: 'Bob' } }
        },
        {
          ID: 'choice-1',
          SequenceID: 'turn-1',
          ParentID: 'turn-1',
          Round: 1,
          Type: 'choice',
          Data: { Actor: { Name: 'Acolyte' } }
        },
        {
          ID: 'attack-1',
          SequenceID: 'turn-1',
          ParentID: 'choice-1',
          Round: 1,
          Type: 'attack',
          Data: { Actor: { Name: 'Acolyte' } }
        }
      ];

      const tree = service.mapSimulationLog(rawLog);

      // 1. Check Rounds
      expect(tree.length).toBe(2); // Round 0 and Round 1
      expect(tree[0].data.id).toBe('round-0');
      expect(tree[1].data.id).toBe('round-1');

      // 2. Check Round 0 (Initiative)
      expect(tree[0].children?.length).toBe(1);
      expect(tree[0].children?.[0].data.id).toBe('init-1');

      // 3. Check Round 1 (Turn)
      expect(tree[1].children?.length).toBe(1); // One turn node
      const turnNode = tree[1].children?.[0] as TimelineNode;
      expect(turnNode.data.id).toBe('turn-1');

      // 4. Check Action Hierarchy within Turn
      expect(turnNode.children?.length).toBe(1); // Top-level choice
      const choiceNode = turnNode.children?.[0] as TimelineNode;
      expect(choiceNode.data.id).toBe('choice-1');

      expect(choiceNode.children?.length).toBe(1); // Nested attack
      expect(choiceNode.children?.[0].data.id).toBe('attack-1');
    });

    it('should handle deep nesting beyond turn level', () => {
      const rawLog = [
        { ID: 'round-1-turn-1', SequenceID: 'turn-1', ParentID: 'turn-1', Round: 1, Type: 'choice', Data: {} },
        { ID: 'attack-1', SequenceID: 'turn-1', ParentID: 'round-1-turn-1', Round: 1, Type: 'attack', Data: {} },
        { ID: 'damage-1', SequenceID: 'turn-1', ParentID: 'attack-1', Round: 1, Type: 'damageroll', Data: {} }
      ];

      const tree = service.mapSimulationLog(rawLog);
      const turnNode = tree[0].children?.[0] as TimelineNode;
      const choiceNode = turnNode.children?.[0] as TimelineNode;
      const attackNode = choiceNode.children?.[0] as TimelineNode;
      const damageNode = attackNode.children?.[0] as TimelineNode;

      expect(damageNode.data.id).toBe('damage-1');
    });
  });
});
