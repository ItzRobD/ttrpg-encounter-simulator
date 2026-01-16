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

      // New unified response structure: { data: { ... } } or { data: { data: { ... } } }
      if (record['data'] !== undefined) {
        let data = record['data'];

        // Handle double nesting { data: { data: { ... } } }
        if (data && typeof data === 'object' && !Array.isArray(data) && (data as Record<string, unknown>)['data'] !== undefined) {
          data = (data as Record<string, unknown>)['data'] as Record<string, unknown>;
        }

        if (data && typeof data === 'object' && !Array.isArray(data)) {
          const dataRecord = data as Record<string, unknown>;
          // Distinguish between Monster and Character within the "data" envelope
          if (dataRecord['actions'] && dataRecord['as_config']) {
            return this.mapMonsterResponse(dataRecord);
          }

          if (dataRecord['class_id'] !== undefined && dataRecord['race_id'] !== undefined) {
            return this.mapCharacterResponse(dataRecord);
          }

          if (dataRecord['spell_type'] !== undefined || dataRecord['casting_time'] !== undefined) {
            return this.mapSpellResponse(dataRecord);
          }

          // Handle dictionary of objects (e.g., { "119": { ... } })
          const values = Object.values(dataRecord);
          if (values.length > 0 && typeof values[0] === 'object' && values[0] !== null) {
            const firstVal = values[0] as Record<string, unknown>;
            if (firstVal['spell_type'] !== undefined || firstVal['casting_time'] !== undefined) {
              return this.mapSpellResponse(firstVal);
            }
          }
        }

        // If it's just raw data (Weapon, Armor, or any other object/array) inside the envelope
        return this.mapKeys(data);
      }

      // Handle entities directly (from seeds or timelines)
      if (record['as_config'] && typeof record['as_config'] === 'object') {
        const data = record as Record<string, unknown>;
        if (data['monster_actions'] || data['actions']) {
          return this.mapMonsterResponse(data);
        }
        if (data['class_id'] !== undefined || data['class']) {
          return this.mapCharacterResponse(data);
        }
      }

      // Handle spells directly
      if (record['spell_type'] !== undefined || record['casting_time'] !== undefined) {
        return this.mapSpellResponse(record as Record<string, unknown>);
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
  private mapCharacterResponse(response: Record<string, unknown>): Record<string, unknown> {
    const result: Record<string, unknown> = {};

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
    const rawSpellcasting = response['spellcasting'] as Record<string, unknown> | undefined;
    if (rawSpellcasting) {
      const slots: Record<number, { current: number; max: number }> = {};
      const rawSlots = rawSpellcasting['spell_slots'] as Record<string, number> | undefined;
      if (rawSlots) {
        Object.entries(rawSlots).forEach(([level, max]) => {
          slots[Number(level)] = { current: max, max: max };
        });
      }

      const rawSpells = (rawSpellcasting['known_spells'] || rawSpellcasting['spells'] || []) as unknown[];
      const spells = Array.isArray(rawSpells) ? rawSpells.map((s) => this.mapKeys(s)) : [];

      result['spellcasting'] = {
        casterType: ((rawSpellcasting['caster_type'] as string) || 'none').toLowerCase(),
        casterLevel: (rawSpellcasting['casting_level'] as number) || 0,
        spellSlots: slots,
        spells: spells,
        spellSaveDC: (rawSpellcasting['save_dc'] as number) || 0,
        spellAttackBonus: (rawSpellcasting['attack_modifier'] as number) || 0
      };
    }

    // 2. Map Equipment
    const rawEquipment = response['equipment'] as Record<string, unknown> | undefined;
    if (rawEquipment) {
      const weapons: Record<string, unknown[]> = {};

      const mapWeaponSlot = (slotData: unknown[]) => {
        if (!Array.isArray(slotData)) return [];
        return slotData.map(item => {
          const itemRecord = item as Record<string, unknown>;
          const weapon = this.mapKeys(itemRecord['weapon_data'] || itemRecord) as Record<string, unknown>;
          return {
            ...weapon,
            isProficient: !!itemRecord['is_proficient']
          };
        });
      };

      if (rawEquipment['primary_slot']) {
        weapons['Primary'] = mapWeaponSlot(rawEquipment['primary_slot'] as unknown[]);
      }
      if (rawEquipment['secondary_slot']) {
        weapons['Secondary'] = mapWeaponSlot(rawEquipment['secondary_slot'] as unknown[]);
      }
      if (rawEquipment['ranged_slot']) {
        weapons['Ranged'] = mapWeaponSlot(rawEquipment['ranged_slot'] as unknown[]);
      }

      result['equipment'] = {
        armor: rawEquipment['armor_data'] ? this.mapKeys(rawEquipment['armor_data']) : (rawEquipment['armor'] ? this.mapKeys(rawEquipment['armor']) : undefined),
        shield: rawEquipment['shield_data'] ? this.mapKeys(rawEquipment['shield_data']) : (rawEquipment['shield'] ? this.mapKeys(rawEquipment['shield']) : undefined),
        hasShieldEquipped: !!rawEquipment['has_shield_equipped'],
        weapons: weapons
      };
    }

    // 3. Map Resistances
    const resistances: Record<string, string> = {};
    const rawResistances = (response['resistances'] || {}) as Record<string, { resistance?: string }>;
    Object.entries(rawResistances).forEach(([type, data]) => {
      resistances[type] = (data.resistance || 'none').toLowerCase();
    });

    // 4. Map Special Abilities
    if (response['special_abilities']) {
      result['specialAbilities'] = this.mapKeys(response['special_abilities']);
    }

    // 5. Ensure state exists
    result['state'] = {
      ...(result['state'] as Record<string, unknown>),
      resistances: resistances,
      conditions: {},
      deathSaves: { successes: 0, failures: 0 },
      isStable: false,
      isDead: false
    };

    return result;
  }

  /**
   * Specifically handles the mapping of the Spell response from the backend.
   */
  private mapSpellResponse(response: Record<string, unknown>): Record<string, unknown> {
    const result: Record<string, unknown> = {};

    Object.keys(response).forEach((key) => {
      if (key === 'formulas' || key === 'spell_dc' || key === 'casting_time') return;
      const camelKey = this.toCamelCase(key);
      result[camelKey] = this.mapKeys(response[key]);
    });

    // Handle Formulas - converting map keys to numbers and values to camelCase
    if (response['formulas'] && typeof response['formulas'] === 'object') {
      const rawFormulas = response['formulas'] as Record<string, unknown>;
      const formulas: Record<number, unknown> = {};

      Object.entries(rawFormulas).forEach(([key, value]) => {
        formulas[Number(key)] = this.mapKeys(value);
      });
      result['formulas'] = formulas;
    }

    // Handle spell_dc explicitly
    if (response['spell_dc'] && typeof response['spell_dc'] === 'object') {
      result['spellDC'] = this.mapKeys(response['spell_dc']);
    }

    // Handle casting_time explicitly
    if (response['casting_time'] !== undefined) {
      result['castingTime'] = response['casting_time'];
    }

    // Explicitly map range if present (it's often missing or differently named in backend)
    if (response['range'] !== undefined) {
      result['range'] = response['range'];
    }

    return result;
  }

  /**
   * Specifically handles the mapping of the complex Monster response from the backend.
   */
  private mapMonsterResponse(response: Record<string, unknown>): Record<string, unknown> {
    const result: Record<string, unknown> = {};

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
    const rawActions = (response['actions'] || []) as unknown[];
    const actions = Array.isArray(rawActions) ? rawActions.map((a) => this.mapKeys(a)) : [];

    // 2. Map Multiattacks
    const rawMultiattacks = (response['multiattacks'] || []) as unknown[];
    const multiattacks = Array.isArray(rawMultiattacks) ? rawMultiattacks.map((m) => this.mapKeys(m)) : [];

    // 3. Map Legendary Actions
    const rawLegendary = (response['legendary_actions'] || []) as unknown[];
    const legendaryActions = Array.isArray(rawLegendary) ? rawLegendary.map((la) => {
      const mapped = this.mapKeys(la) as Record<string, unknown>;
      if (mapped['action']) {
        return { ...(mapped['action'] as Record<string, unknown>), cost: mapped['cost'] };
      }
      return mapped;
    }) : [];

    // 4. Map Resistances
    const resistances: Record<string, string> = {};
    const rawResistances = (response['resistances'] || {}) as Record<string, { resistance?: string }>;
    Object.entries(rawResistances).forEach(([type, data]) => {
      resistances[type] = (data.resistance || 'none').toLowerCase();
    });

    // 5. Assemble the monsterActions structure
    const rechargeActions: Record<number, number> = {};
    actions.forEach((a) => {
      const action = a as Record<string, unknown>;
      if ((action['rechargeValue'] as number) > 0) {
        rechargeActions[action['actionId'] as number] = action['rechargeValue'] as number;
      }
    });

    result['monsterActions'] = {
      actions: actions,
      multiattacks: multiattacks,
      legendaryActions: legendaryActions,
      rechargeActions: rechargeActions
    };

    // 6. Map Spellcasting
    const rawSpellcasting = response['spellcasting'] as Record<string, unknown> | undefined;
    if (rawSpellcasting) {
      const spells = ((rawSpellcasting['leveled_spells'] || rawSpellcasting['known_spells'] || []) as unknown[]).map((s) => this.mapKeys(s));
      const innateSpells = ((rawSpellcasting['innate_spells'] || []) as unknown[]).map((is) => {
        const innateRecord = is as Record<string, unknown>;
        const spell = this.mapKeys(innateRecord['Spell'] || innateRecord['spell'] || innateRecord) as Record<string, unknown>;
        return {
          ...spell,
          isInnate: true,
          maxCastsPerDay: innateRecord['MaxCastsPerDay'] !== undefined ? innateRecord['MaxCastsPerDay'] : (innateRecord['max_casts_per_day'] !== undefined ? innateRecord['max_casts_per_day'] : -1)
        };
      });

      const slots: Record<number, { current: number; max: number }> = {};
      if (rawSpellcasting['spell_slots']) {
        Object.entries(rawSpellcasting['spell_slots'] as Record<string, number>).forEach(([level, max]) => {
          slots[Number(level)] = { current: max, max: max };
        });
      }

      result['spellcasting'] = {
        casterType: ((rawSpellcasting['caster_type'] as string) || 'full').toLowerCase(),
        casterLevel: (rawSpellcasting['casting_level'] as number) || 0,
        spellSlots: slots,
        spells: [...spells, ...innateSpells],
        spellSaveDC: (rawSpellcasting['save_dc'] as number) || 0,
        spellAttackBonus: (rawSpellcasting['attack_modifier'] as number) || 0
      };
    } else if (response['spellcasting_config']) {
      // Fallback for older monster format if any
      const oldSpellcasting = response['spellcasting_config'] as Record<string, unknown>;
      const spells = ((oldSpellcasting['LeveledSpells'] || []) as unknown[]).map((s) => this.mapKeys(s));
      const innateSpells = ((oldSpellcasting['InnateSpells'] || []) as unknown[]).map((is) => {
        const innateRecord = is as Record<string, unknown>;
        const spell = this.mapKeys(innateRecord['Spell'] || innateRecord['spell'] || {}) as Record<string, unknown>;
        return {
          ...spell,
          isInnate: true,
          maxCastsPerDay: innateRecord['MaxCastsPerDay'] !== undefined ? innateRecord['MaxCastsPerDay'] : (innateRecord['max_casts_per_day'] !== undefined ? innateRecord['max_casts_per_day'] : -1)
        };
      });
      const slots: Record<number, { current: number; max: number }> = {};
      if (oldSpellcasting['SpellSlots']) {
        Object.entries(oldSpellcasting['SpellSlots'] as Record<string, number>).forEach(([level, max]) => {
          slots[Number(level)] = { current: max, max: max };
        });
      }
      result['spellcasting'] = {
        casterType: 'full',
        casterLevel: (oldSpellcasting['CastingLevel'] as number) || 0,
        spellSlots: slots,
        spells: [...spells, ...innateSpells],
        spellSaveDC: (oldSpellcasting['SaveDC'] as number) || 0,
        spellAttackBonus: (oldSpellcasting['AttackModifier'] as number) || 0
      };
    }

    // 7. Ensure state exists
    result['state'] = {
      ...(result['state'] as Record<string, unknown>),
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

    // Handle snake_case specifically for spell_dc and other dnd fields
    if (str === 'on_success') return 'onSuccess';
    if (str === 'spell_dc') return 'spellDC';
    if (str === 'casting_time') return 'castingTime';
    if (str === 'spell_type') return 'spellType';
    if (str === 'level_type') return 'levelType';
    if (str === 'is_concentration') return 'isConcentration';
    if (str === 'is_ritual') return 'isRitual';
    if (str === 'is_touch') return 'isTouch';
    if (str === 'is_aoe') return 'isAOE';
    if (str === 'has_dc') return 'hasDC';
    if (str === 'is_auto_hit') return 'isAutoHit';

    // SpellFormula fields
    if (str === 'cast_level') return 'castLevel';
    if (str === 'number_of_dice') return 'numberOfDice';
    if (str === 'amount_to_add') return 'amountToAdd';
    if (str === 'use_spellmod') return 'useSpellMod';
    if (str === 'damage_type') return 'damageType';
    if (str === 'average_value') return 'averageValue';

    // Handle PascalCase explicitly for SpellFormulas (Fallback for older backend versions)
    if (str === 'CastLevel') return 'castLevel';
    if (str === 'NumberOfDice') return 'numberOfDice';
    if (str === 'Die') return 'die';
    if (str === 'AmountToAdd') return 'amountToAdd';
    if (str === 'UseSpellmod') return 'useSpellMod';
    if (str === 'DamageType') return 'damageType';
    if (str === 'AverageValue') return 'averageValue';

    let result = str;

    // Handle snake_case
    if (result.includes('_')) {
      result = result.replace(/_([a-z0-9])/g, (g) => g[1].toUpperCase());
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
    const finalResult = result.charAt(0).toLowerCase() + result.slice(1);

    // Final check for common dnd fields that might have slipped through
    if (finalResult === 'castingTime') return 'castingTime';
    if (finalResult === 'spellType') return 'spellType';
    if (finalResult === 'levelType') return 'levelType';

    return finalResult;
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
