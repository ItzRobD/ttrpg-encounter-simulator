import { Injectable } from '@angular/core';
import { Entity, EventType, SimulationEvent, SimulationLog, TimelineNode } from '../models';

@Injectable({
  providedIn: 'root',
})
export class MapperService {
  /**
   * Transforms a raw gzipped JSON log into a hierarchical tree for the UI.
   * Expects the events array.
   */
  mapSimulationLog(events: SimulationEvent[]): TimelineNode[] {
    return this.buildTree(events);
  }

  /**
   * Recursively converts object keys from PascalCase to camelCase.
   * Special handling for Monster data from backend.
   */
  mapKeys(obj: unknown): unknown {
    if (Array.isArray(obj)) {
      return obj.map((v) => this.mapKeys(v));
    } else if (obj !== null && typeof obj === 'object') {
      const record = obj as Record<string, unknown>;

      // New unified response structure: { data: { ... } }
      if (record['data'] && typeof record['data'] === 'object' && !Array.isArray(record['data'])) {
        const data = record['data'] as Record<string, any>;

        // Distinguish between Monster and Character within the "data" envelope
        if (data['actions'] && data['as_config']) {
          return this.mapMonsterResponse(data);
        }

        if (data['class_id'] !== undefined && data['race_id'] !== undefined) {
          return this.mapCharacterResponse(data);
        }
      }

      // Handle entities directly (from seeds or timelines)
      if (record['as_config'] && typeof record['as_config'] === 'object') {
        const data = record as Record<string, any>;
        if (data['monster_actions'] || data['actions']) {
          return this.mapMonsterResponse(data);
        }
        if (data['class_id'] !== undefined || data['class']) {
          return this.mapCharacterResponse(data);
        }
      }

      // Fallback for objects already inside the mappers or simpler objects
      const result: Record<string, unknown> = {};
      Object.keys(record).forEach((key) => {
        const camelKey = this.toCamelCase(key);
        result[camelKey] = this.mapKeys(record[key]);
      });

      return result;
    }
    return obj;
  }

  /**
   * Specifically handles the mapping of the Character response from the backend.
   */
  private mapCharacterResponse(response: Record<string, any>): Record<string, any> {
    const result: Record<string, any> = {};

    Object.keys(response).forEach((key) => {
      const camelKey = this.toCamelCase(key);
      result[camelKey] = this.mapKeys(response[key]);
    });

    // Handle class_id and race_id explicitly to ensure both snake_case and camelCase keys are available
    // for type guards and mapping logic that might look for either.
    if (response['class_id'] !== undefined) {
      result['class_id'] = response['class_id'];
      result['classId'] = response['class_id'];
    }
    if (response['race_id'] !== undefined) {
      result['race_id'] = response['race_id'];
      result['raceId'] = response['race_id'];
    }

    // Explicitly ensure AC and HP are mapped if they exist in the response
    if (response['ac'] !== undefined) result['ac'] = response['ac'];
    if (response['hp'] !== undefined) result['hp'] = this.mapKeys(response['hp']);

    // Map as_config directly to asConfig
    if (response['as_config']) {
      result['asConfig'] = this.mapKeys(response['as_config']);
    }

    // 1. Map Spellcasting if present
    const rawSpellcasting = response['spellcasting'];
    if (rawSpellcasting) {
      const slots: Record<number, { current: number; max: number }> = {};
      const rawSlots = rawSpellcasting.spell_slots;
      if (rawSlots) {
        Object.entries(rawSlots).forEach(([level, max]: [string, any]) => {
          slots[Number(level)] = { current: max, max: max };
        });
      }

      const rawSpells = rawSpellcasting.known_spells || rawSpellcasting.spells || [];
      const spells = Array.isArray(rawSpells) ? rawSpells.map((s: any) => this.mapKeys(s)) : [];

      result['spellcasting'] = {
        casterType: (rawSpellcasting.caster_type || 'none').toLowerCase(),
        casterLevel: rawSpellcasting.casting_level || 0,
        spellSlots: slots,
        spells: spells,
        spellSaveDC: rawSpellcasting.save_dc || 0,
        spellAttackBonus: rawSpellcasting.attack_modifier || 0
      };
    }

    // 2. Map Equipment
    const rawEquipment = response['equipment'];
    if (rawEquipment) {
      const weapons: Record<string, any[]> = {};

      const mapWeaponSlot = (slotData: any[]) => {
        if (!Array.isArray(slotData)) return [];
        return slotData.map(item => {
          const weapon = this.mapKeys(item.weapon_data || item) as any;
          return {
            ...weapon,
            isProficient: !!item.is_proficient
          };
        });
      };

      if (rawEquipment.primary_slot) {
        weapons['Primary'] = mapWeaponSlot(rawEquipment.primary_slot);
      }
      if (rawEquipment.secondary_slot) {
        weapons['Secondary'] = mapWeaponSlot(rawEquipment.secondary_slot);
      }
      if (rawEquipment.ranged_slot) {
        weapons['Ranged'] = mapWeaponSlot(rawEquipment.ranged_slot);
      }

      result['equipment'] = {
        armor: rawEquipment.armor_data ? this.mapKeys(rawEquipment.armor_data) : (rawEquipment.armor ? this.mapKeys(rawEquipment.armor) : undefined),
        shield: rawEquipment.shield_data ? this.mapKeys(rawEquipment.shield_data) : (rawEquipment.shield ? this.mapKeys(rawEquipment.shield) : undefined),
        hasShieldEquipped: !!rawEquipment.has_shield_equipped,
        weapons: weapons
      };
    }

    // 3. Map Resistances
    const resistances: Record<string, string> = {};
    const rawResistances = response['resistances'] || {};
    Object.entries(rawResistances).forEach(([type, data]: [string, any]) => {
      resistances[type] = (data.resistance || 'none').toLowerCase();
    });

    // 4. Map Special Abilities
    if (response['special_abilities']) {
      result['specialAbilities'] = this.mapKeys(response['special_abilities']);
    }

    // 5. Ensure state exists
    result['state'] = {
      ...result['state'],
      resistances: resistances,
      conditions: {},
      deathSaves: { successes: 0, failures: 0 },
      isStable: false,
      isDead: false
    };

    return result;
  }

  /**
   * Specifically handles the mapping of the complex Monster response from the backend.
   */
  private mapMonsterResponse(response: Record<string, any>): Record<string, any> {
    const result: Record<string, any> = {};

    Object.keys(response).forEach((key) => {
      // Avoid recursive mapping of large arrays/objects we handle manually below
      if (['actions', 'legendary_actions', 'multiattacks', 'resistances', 'spellcasting'].includes(key)) {
        return;
      }
      const camelKey = this.toCamelCase(key);
      result[camelKey] = this.mapKeys(response[key]);
    });

    // Map as_config directly to asConfig
    if (response['as_config']) {
      result['asConfig'] = this.mapKeys(response['as_config']);
    }

    // 1. Map Actions
    const rawActions = response['actions'] || [];
    const actions = Array.isArray(rawActions) ? rawActions.map((a: any) => this.mapKeys(a)) : [];

    // 2. Map Multiattacks
    const rawMultiattacks = response['multiattacks'] || [];
    const multiattacks = Array.isArray(rawMultiattacks) ? rawMultiattacks.map((m: any) => this.mapKeys(m)) : [];

    // 3. Map Legendary Actions
    const rawLegendary = response['legendary_actions'] || [];
    const legendaryActions = Array.isArray(rawLegendary) ? rawLegendary.map((la: any) => {
      const mapped = this.mapKeys(la) as any;
      if (mapped.action) {
        return { ...mapped.action, cost: mapped.cost };
      }
      return mapped;
    }) : [];

    // 4. Map Resistances
    const resistances: Record<string, string> = {};
    const rawResistances = response['resistances'] || {};
    Object.entries(rawResistances).forEach(([type, data]: [string, any]) => {
      resistances[type] = (data.resistance || 'none').toLowerCase();
    });

    // 5. Assemble the monsterActions structure
    const rechargeActions: Record<number, number> = {};
    actions.forEach((a: any) => {
      if (a.rechargeValue > 0) {
        rechargeActions[a.actionId] = a.rechargeValue;
      }
    });

    result['monsterActions'] = {
      actions: actions,
      multiattacks: multiattacks,
      legendaryActions: legendaryActions,
      rechargeActions: rechargeActions
    };

    // 6. Map Spellcasting
    const rawSpellcasting = response['spellcasting'];
    if (rawSpellcasting) {
      const spells = (rawSpellcasting.leveled_spells || rawSpellcasting.known_spells || []).map((s: any) => this.mapKeys(s));
      const innateSpells = (rawSpellcasting.innate_spells || []).map((is: any) => {
        const spell = this.mapKeys(is.Spell || is.spell || is) as any;
        return {
          ...spell,
          isInnate: true,
          maxCastsPerDay: is.MaxCastsPerDay !== undefined ? is.MaxCastsPerDay : (is.max_casts_per_day !== undefined ? is.max_casts_per_day : -1)
        };
      });

      const slots: Record<number, { current: number; max: number }> = {};
      if (rawSpellcasting.spell_slots) {
        Object.entries(rawSpellcasting.spell_slots).forEach(([level, max]: [string, any]) => {
          slots[Number(level)] = { current: max, max: max };
        });
      }

      result['spellcasting'] = {
        casterType: (rawSpellcasting.caster_type || 'full').toLowerCase(),
        casterLevel: rawSpellcasting.casting_level || 0,
        spellSlots: slots,
        spells: [...spells, ...innateSpells],
        spellSaveDC: rawSpellcasting.save_dc || 0,
        spellAttackBonus: rawSpellcasting.attack_modifier || 0
      };
    } else if (response['spellcasting_config']) {
      // Fallback for older monster format if any
      const oldSpellcasting = response['spellcasting_config'];
      const spells = (oldSpellcasting.LeveledSpells || []).map((s: any) => this.mapKeys(s));
      const innateSpells = (oldSpellcasting.InnateSpells || []).map((is: any) => {
        const spell = this.mapKeys(is.Spell || is.spell || {}) as any;
        return {
          ...spell,
          isInnate: true,
          maxCastsPerDay: is.MaxCastsPerDay !== undefined ? is.MaxCastsPerDay : (is.max_casts_per_day !== undefined ? is.max_casts_per_day : -1)
        };
      });
      const slots: Record<number, { current: number; max: number }> = {};
      if (oldSpellcasting.SpellSlots) {
        Object.entries(oldSpellcasting.SpellSlots).forEach(([level, max]: [string, any]) => {
          slots[Number(level)] = { current: max, max: max };
        });
      }
      result['spellcasting'] = {
        casterType: 'full',
        casterLevel: oldSpellcasting.CastingLevel || 0,
        spellSlots: slots,
        spells: [...spells, ...innateSpells],
        spellSaveDC: oldSpellcasting.SaveDC || 0,
        spellAttackBonus: oldSpellcasting.AttackModifier || 0
      };
    }

    // 7. Ensure state exists
    result['state'] = {
      ...result['state'],
      resistances: resistances,
      conditions: {},
      deathSaves: { successes: 0, failures: 0 },
      isStable: false,
      isDead: false
    };

    return result;
  }

  /**
   * Robust PascalCase to camelCase converter that handles dnd specific abbreviations.
   */
  private toCamelCase(str: string): string {
    if (str === 'ID') return 'id';
    if (str === 'AC') return 'ac';
    if (str === 'HP') return 'hp';
    if (str === 'CR') return 'cr';
    if (str === 'AbilityScores' || str === 'ability_scores') return 'abilityScores';
    if (str === 'AbilityScoreProf' || str === 'ability_score_prof') return 'abilityScoreProficiency';
    if (str === 'DC') return 'dc';
    if (str === 'dice_count') return 'numberOfDice';
    if (str === 'hp_method') return 'hpSetMethod';

    let result = str;

    // Handle snake_case
    if (result.includes('_')) {
      result = result.replace(/_([a-z])/g, (g) => g[1].toUpperCase());
    }

    // Handle common acronyms at the end of strings (e.g., SequenceID -> SequenceId, MaxHP -> MaxHp)
    if (result === 'OriginalHP') return 'originalHp';
    if (result === 'FinalHP') return 'finalHp';
    if (result === 'OriginalTempHP') return 'originalTempHp';
    if (result === 'FinalTempHP') return 'finalTempHp';
    if (result === 'OriginalDamage') return 'originalDamage';
    if (result === 'FinalDamage') return 'finalDamage';

    if (result.endsWith('ID')) {
      result = result.slice(0, -2) + 'Id';
    } else if (result.endsWith('HP')) {
      result = result.slice(0, -2) + 'Hp';
    } else if (result.endsWith('AC')) {
      result = result.slice(0, -2) + 'Ac';
    }

    // Handle acronyms at the start of strings (e.g., HPValue -> hpValue)
    if (result.startsWith('HP')) {
      result = 'hp' + result.slice(2);
    } else if (result.startsWith('AC')) {
      result = 'ac' + result.slice(2);
    } else if (result.startsWith('DC')) {
      result = 'dc' + result.slice(2);
    } else if (result.endsWith('DC')) {
      result = result.slice(0, -2) + 'DC';
    }

    // Lowercase the first character for standard camelCase
    return result.charAt(0).toLowerCase() + result.slice(1);
  }

  /**
   * Builds a hierarchical tree from a flat list of events.
   * Structure: Round -> Turn (SequenceID) -> Event -> Sub-events (ParentID)
   */
  private buildTree(events: SimulationEvent[]): TimelineNode[] {
    const roundsMap = new Map<number, TimelineNode>();
    const turnsMap = new Map<string, TimelineNode>();
    const nodesMap = new Map<string, TimelineNode>();

    // 1. Pre-create nodes for all actual events to facilitate hierarchy lookup
    const filteredEvents = events.filter(event => {
      return !(event.type === EventType.DamageModified && !event.data.wasModified);

    });

    filteredEvents.forEach((event) => {
      nodesMap.set(event.id, { data: event, children: [] });
    });

    const rootNodes: TimelineNode[] = [];

    filteredEvents.forEach((event) => {
      const node = nodesMap.get(event.id)!;

      // 0. SPECIAL CASE: Death events with no parent are treated as root nodes
      if (event.type === EventType.Death && !event.parentId) {
        rootNodes.push(node);
        return;
      }

      // 0. SPECIAL CASE: Victory events with no parent are treated as root nodes
      if (event.type === EventType.Victory && !event.parentId) {
        rootNodes.push(node);
        return;
      }

      // 2. Ensure Round node exists
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

      // 3. Handle Turns (grouped by sequenceId)
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

        // 4. Attach to hierarchy
        // If it has a parentId, and that parent is NOT the turn itself, nest it
        if (
          event.parentId &&
          event.parentId !== event.sequenceId &&
          nodesMap.has(event.parentId)
        ) {
          nodesMap.get(event.parentId)!.children?.push(node);
        } else {
          // Top-level action within a turn
          turnNode.children?.push(node);
        }
      } else {
        // Events without a sequence (like Initiative) go directly under the round
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
