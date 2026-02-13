import { Injectable } from '@angular/core';
import { Actor, ActorSummary, Race, Class, SimulationEvent, TimelineNode, EventType } from '../models';

@Injectable({
  providedIn: 'root',
})
export class MapperService {
  private readonly raceMap: Record<Race, number> = {
    [Race.Dwarf]: 1,
    [Race.Dragonborn]: 2,
    [Race.Elf]: 3,
    [Race.Halfling]: 4,
    [Race.Human]: 5,
    [Race.Gnome]: 6,
    [Race.HalfElf]: 7,
    [Race.HalfOrc]: 8,
    [Race.Tiefling]: 9
  };

  private readonly classMap: Record<Class, number> = {
    [Class.Artificer]: 1,
    [Class.Barbarian]: 2,
    [Class.Bard]: 3,
    [Class.Cleric]: 4,
    [Class.Druid]: 5,
    [Class.Fighter]: 6,
    [Class.Monk]: 7,
    [Class.Paladin]: 8,
    [Class.Ranger]: 9,
    [Class.Rogue]: 10,
    [Class.Sorcerer]: 11,
    [Class.Warlock]: 12,
    [Class.Wizard]: 13
  };

  public getRaceName(raceId: number): string {
    const entry = Object.entries(this.raceMap).find(([_, id]) => id === raceId);
    return entry ? entry[0] : 'Unknown Race';
  }

  public getClassName(classId: number): string {
    const entry = Object.entries(this.classMap).find(([_, id]) => id === classId);
    return entry ? entry[0] : 'Unknown Class';
  }

  public getRaceId(race: Race): number {
    return this.raceMap[race];
  }

  public getClassId(clazz: Class): number {
    return this.classMap[clazz];
  }

  /**
   * Transforms a raw gzipped JSON log into a hierarchical tree for the UI.
   * Expects the events array.
   */
  mapSimulationLog(events: SimulationEvent[]): TimelineNode[] {
    return this.buildTree(events);
  }

  /**
   * Recursively converts object keys from snake_case to camelCase.
   * Handles response envelopes (data, details) and common D&D abbreviations.
   */
  mapKeys(obj: unknown): unknown {
    if (Array.isArray(obj)) {
      return obj.map((v) => this.mapKeys(v));
    } else if (obj !== null && typeof obj === 'object') {
      const record = obj as Record<string, unknown>;

      // Resolve unified response envelopes: { data: { ... } } or { data: { data: { ... } } }
      let current = record;
      while (current && typeof current === 'object' && !Array.isArray(current)) {
        if (current['data'] !== undefined && current['id'] === undefined && current['ID'] === undefined && current['type'] === undefined && current['name'] === undefined && current['spell_type'] === undefined) {
          current = current['data'] as Record<string, unknown>;
        } else if (current['details'] !== undefined && typeof current['details'] === 'object' && current['id'] === undefined && current['ID'] === undefined && current['type'] === undefined) {
          const details = current['details'] as Record<string, unknown>;
          if (details['data'] !== undefined && details['id'] === undefined && details['type'] === undefined) {
            current = details['data'] as Record<string, unknown>;
          } else {
            break;
          }
        } else {
          break;
        }
      }

      if (current !== record) {
        return this.mapKeys(current);
      }

      const result: Record<string, unknown> = {};
      Object.keys(current).forEach((key) => {
        const camelKey = this.toCamelCase(key);
        result[camelKey] = this.mapKeys(current[key]);
      });

      return result;
    }
    return obj;
  }

  /**
   * Recursively converts object keys from camelCase to snake_case for the backend.
   */
  serializeKeys<T>(obj: T): any {
    if (Array.isArray(obj)) {
      return obj.map((v) => this.serializeKeys(v));
    } else if (obj !== null && typeof obj === 'object') {
      const result: Record<string, any> = {};
      const typedObj = obj as Record<string, any>;

      Object.keys(typedObj).forEach((key) => {
        // Hydrated objects that shouldn't be sent back as full objects if IDs are present
        if (key === 'armor' && typedObj['armorId']) return;
        if (key === 'shield' && typedObj['shieldId']) return;
        if (key === 'spellcasting') return;
        if (key === 'hp') return;
        if (key === 'asConfig') return;
        if (key === 'equipment' && typedObj['equipmentConfigs']) return;
        if (key === 'monsterActions') return;
        if (key === 'state') return;

        const snakeKey = this.camelToSnake(key);
        result[snakeKey] = this.serializeKeys(typedObj[key]);
      });
      return result;
    }
    return obj;
  }

  private camelToSnake(key: string): string {
    // Special mappings for D&D acronyms and specific fields
    const specialMappings: Record<string, string> = {
      'ID': 'id',
      'id': 'ID',
      'spellSaveDC': 'spell_save_dc',
      'spellAttackBonus': 'spell_attack_bonus',
      'instanceId': 'instance_id',
      'classId': 'class_id',
      'raceId': 'race_id',
      'armorId': 'armor_id',
      'shieldId': 'shield_id',
      'weaponId': 'weapon_id',
      'knownSpellIDs': 'known_spell_ids',
      'equipmentConfigs': 'equipment_configs',
      'spellcasterMetadata': 'spellcaster_metadata',
      'hpConfig': 'hp_config',
      'numberOfDice': 'number_of_dice',
      'isConcentration': 'is_concentration',
      'castingTime': 'casting_time',
      'spellType': 'spell_type',
      'isRitual': 'is_ritual',
      'isTouch': 'is_touch',
      'isAOE': 'is_aoe'
    };

    if (specialMappings[key]) return specialMappings[key];

    let tempKey = key;
    tempKey = tempKey.replace(/DC/g, 'Dc');
    tempKey = tempKey.replace(/HP/g, 'Hp');
    tempKey = tempKey.replace(/AC/g, 'Ac');
    tempKey = tempKey.replace(/AI/g, 'Ai');
    tempKey = tempKey.replace(/AOE/g, 'Aoe');

    let snakeKey = tempKey.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`);
    if (snakeKey.startsWith('_')) {
      snakeKey = snakeKey.substring(1);
    }
    return snakeKey;
  }

  /**
   * Robust snake_case/PascalCase to camelCase converter.
   */
  private toCamelCase(str: string): string {
    // Exact match overrides
    const overrides: Record<string, string> = {
      'ID': 'id',
      'AC': 'ac',
      'HP': 'hp',
      'CR': 'cr',
      'DC': 'dc',
      'ability_scores': 'abilityScores',
      'ability_score_prof': 'abilityScoreProficiency',
      'dice_count': 'numberOfDice',
      'number_of_dice': 'numberOfDice',
      'hp_method': 'hpSetMethod',
      'hp_set_method': 'hpSetMethod',
      'class_id': 'classId',
      'race_id': 'raceId',
      'armor_id': 'armorId',
      'shield_id': 'shieldId',
      'weapon_id': 'weaponId',
      'instance_id': 'instanceId',
      'monster_actions': 'monsterActions',
      'as_config': 'asConfig',
      'is_custom': 'isCustom',
      'spell_dc': 'spellDC',
      'casting_time': 'castingTime',
      'spell_type': 'spellType',
      'is_concentration': 'isConcentration',
      'is_ritual': 'isRitual',
      'is_touch': 'isTouch',
      'is_aoe': 'isAOE',
      'on_success': 'onSuccess',
      'actor_type': 'actorType',
      'hp_config': 'hpConfig',
      'equipment_configs': 'equipmentConfigs',
      'known_spell_ids': 'knownSpellIDs',
      'custom_equipment': 'customEquipment',
      'custom_spells': 'customSpells',
      'spell_actions': 'spellActions',
      'spellcasting_config': 'spellcasting',
      'action_preference': 'actionPreference',
      'secondary_action_preference': 'secondaryActionPreference',
      'target_priority': 'targetPriority',
      'secondary_target_priority': 'secondaryTargetPriority',
      'spellcaster_metadata': 'spellcasterMetadata',
      'is_spellcaster': 'isSpellcaster',
      'is_innate_caster': 'isInnateCaster',
      'spellcasting_ability': 'spellcastingAbility',
      'spellcasting_level': 'spellcastingLevel',
      'is_legendary': 'isLegendary',
      'max_legendary_actions': 'maxLegendaryActions',
      'average_offensive_value': 'averageOffensiveValue',
      'highest_offensive_value': 'highestOffensiveValue',
      'hit_dice': 'hitDice',
      'amount_to_add': 'amountToAdd',
      'use_spellmod': 'useSpellMod',
      'level_type': 'levelType',
      'minimum_str': 'minimumStr',
      'damage_blocks': 'damageBlocks',
      'dice_block': 'damageBlocks',
      'leveled_spells': 'leveledSpells',
      'innate_spells': 'innateSpells',
      'save_dc': 'saveDC',
      'attack_modifier': 'attackModifier',
      'is_silvered': 'isSilvered',
      'is_adamantine': 'isAdamantine',
      'is_cold_forged_iron': 'isColdForgedIron',
      'is_versatile': 'isVersatile',
      'is_finesse': 'isFinesse',
      'is_ranged': 'isRanged',
      'is_heavy': 'isHeavy',
      'is_two_handed': 'isTwoHanded',
      'is_light': 'isLight',
      'is_thrown': 'isThrown',
      'is_only_ranged': 'isOnlyRanged',
      'attack_bonus': 'attackBonus',
      'damage_bonus': 'damageBonus',
      'dex_bonus': 'dexBonus',
      'max_bonus': 'maxBonus',
      'is_innate': 'isInnate',
      'max_casts_per_day': 'maxCastsPerDay'
    };

    if (overrides[str]) return overrides[str];

    // Standard snake_case to camelCase
    let result = str.toLowerCase().replace(/(_[a-z])/g, (g) => g.toUpperCase().replace('_', ''));

    // Handle PascalCase or already camelCase but with D&D acronyms
    result = result.replace(/Ac(?=[A-Z]|$)/g, 'AC');
    result = result.replace(/Hp(?=[A-Z]|$)/g, 'HP');
    result = result.replace(/Dc(?=[A-Z]|$)/g, 'DC');

    // Ensure first character is lowercase
    return result.charAt(0).toLowerCase() + result.slice(1);
  }

  /**
   * Builds a hierarchical tree from a flat list of events.
   */
  private buildTree(events: SimulationEvent[]): TimelineNode[] {
    const roundsMap = new Map<number, TimelineNode>();
    const turnsMap = new Map<string, TimelineNode>();
    const nodesMap = new Map<string, TimelineNode>();

    const filteredEvents = events.filter(event => {
      return !(event.type === EventType.DamageModified && !event.data.wasModified);
    });

    filteredEvents.forEach((event) => {
      nodesMap.set(event.id, { data: event, children: [] });
    });

    const rootNodes: TimelineNode[] = [];

    filteredEvents.forEach((event) => {
      const node = nodesMap.get(event.id)!;

      if ((event.type === EventType.Death || event.type === EventType.Victory) && !event.parentId) {
        rootNodes.push(node);
        return;
      }

      if (!roundsMap.has(event.round)) {
        const roundNode: TimelineNode = {
          data: {
            round: event.round,
            type: EventType.Round,
            id: `round-${event.round}`,
            data: { note: `Round ${event.round}` },
          },
          children: [],
          expanded: false,
        };
        roundsMap.set(event.round, roundNode);
        rootNodes.push(roundNode);
      }
      const roundNode = roundsMap.get(event.round)!;

      if (event.sequenceId) {
        if (!turnsMap.has(event.sequenceId)) {
          const turnNode: TimelineNode = {
            data: {
              round: event.round,
              type: EventType.Turn,
              id: event.sequenceId,
              sequenceId: event.sequenceId,
              data: {
                actor: event.data.actor,
                note: `Turn: ${event.data.actor?.name || 'Unknown'}`,
              },
            },
            children: [],
          };
          turnsMap.set(event.sequenceId, turnNode);
          roundNode.children?.push(turnNode);
        }
        const turnNode = turnsMap.get(event.sequenceId)!;

        if (event.parentId && event.parentId !== event.sequenceId && nodesMap.has(event.parentId)) {
          nodesMap.get(event.parentId)!.children?.push(node);
        } else {
          turnNode.children?.push(node);
        }
      } else {
        if (event.parentId && nodesMap.has(event.parentId)) {
          nodesMap.get(event.parentId)!.children?.push(node);
        } else {
          roundNode.children?.push(node);
        }
      }
    });

    return rootNodes;
  }
}
