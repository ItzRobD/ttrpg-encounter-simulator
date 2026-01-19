import { Injectable } from '@angular/core';
import { Entity, EventType, SimulationEvent, SimulationLog, TimelineNode, Race, Class, WeaponSlotData } from '../models';

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
   * Recursively converts object keys from PascalCase to camelCase.
   * Special handling for Monster data from backend.
   */
  mapKeys(obj: unknown): unknown {
    if (Array.isArray(obj)) {
      return obj.map((v) => this.mapKeys(v));
    } else if (obj !== null && typeof obj === 'object') {
      const record = obj as Record<string, unknown>;

      // Resolve unified response envelopes: { data: { ... } } or { data: { data: { ... } } }
      // including { data: { details: { data: { ... } } } }
      let current = record;
      while (current && typeof current === 'object' && !Array.isArray(current)) {
        if (current['data'] !== undefined) {
          current = current['data'] as Record<string, unknown>;
        } else if (current['details'] !== undefined && typeof current['details'] === 'object') {
          const details = current['details'] as Record<string, unknown>;
          if (details['data'] !== undefined) {
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
      if (record['spell_type'] !== undefined || record['casting_time'] !== undefined || record['spell_dc'] !== undefined || record['formulas'] !== undefined || record['spellType'] !== undefined || record['is_concentration'] !== undefined || record['isConcentration'] !== undefined) {
        return this.mapSpellResponse(record as Record<string, unknown>);
      }

      // Handle weapons and armor directly (they might not have as_config)
      if (record['damage_blocks'] && record['properties'] && record['modifiers']) {
        // This looks like a Weapon
        const result: Record<string, unknown> = {};
        Object.keys(record).forEach((key) => {
          const camelKey = this.toCamelCase(key);
          result[camelKey] = this.mapKeys(record[key]);
        });
        return result;
      }

      if (record['ac'] !== undefined && record['minimum_strength'] !== undefined) {
        // This looks like Armor
        const result: Record<string, unknown> = {};
        Object.keys(record).forEach((key) => {
          const camelKey = this.toCamelCase(key);
          result[camelKey] = this.mapKeys(record[key]);
        });
        return result;
      }

      // Special case for dictionary of objects (e.g., { "119": { ... } })
      const entries = Object.entries(record);
      if (entries.length > 0 && typeof entries[0][1] === 'object' && entries[0][1] !== null) {
        // Check if the first value looks like a spell or entity
        const firstVal = entries[0][1] as Record<string, unknown>;
        if (firstVal['spell_type'] !== undefined || firstVal['casting_time'] !== undefined || firstVal['spell_dc'] !== undefined || firstVal['formulas'] !== undefined || firstVal['spellType'] !== undefined || firstVal['is_concentration'] !== undefined || firstVal['isConcentration'] !== undefined) {
          // If there's only one entry, it might be a single spell wrapped in its ID
          if (entries.length === 1) {
            return this.mapSpellResponse(firstVal);
          }

          const result: Record<string, unknown> = {};
          entries.forEach(([k, v]) => {
            result[k] = this.mapSpellResponse(v as Record<string, unknown>);
          });
          return result;
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
  private mapCharacterResponse(response: Record<string, unknown>): Record<string, unknown> {
    const result: Record<string, unknown> = {};

    Object.keys(response).forEach((key) => {
      const camelKey = this.toCamelCase(key);
      result[camelKey] = this.mapKeys(response[key]);
    });

    // Ensure raceId and classId are present and numeric
    if (response['race_id'] !== undefined) {
      result['raceId'] = Number(response['race_id']);
    } else if (response['race'] !== undefined && typeof response['race'] === 'string') {
      result['raceId'] = this.raceMap[response['race'] as Race] ?? 0;
    }

    if (response['class_id'] !== undefined) {
      result['classId'] = Number(response['class_id']);
    } else if (response['class'] !== undefined && typeof response['class'] === 'string') {
      result['classId'] = this.classMap[response['class'] as Class] ?? 0;
    }

    // Explicitly ensure AC and HP are mapped if they exist in the response
    if (response['ac'] !== undefined) result['ac'] = response['ac'];
    if (response['hp'] !== undefined) result['hp'] = this.mapKeys(response['hp']);

    // Ensure armorId and weaponIds are present in the summary if this is a summary
    if (response['armor_id'] !== undefined) result['armorId'] = response['armor_id'];
    if (response['weapon_ids'] !== undefined) result['weaponIds'] = response['weapon_ids'];

    // Map as_config directly to asConfig
    if (response['as_config']) {
      result['asConfig'] = this.mapKeys(response['as_config']);
    }

    // 1. Map Spellcasting if present
    const rawSpellcasting = response['spellcasting'] as Record<string, unknown> | undefined;
    const knownSpells = response['known_spells'] as (string | number)[] | undefined;

    if (rawSpellcasting || knownSpells) {
      const spellIds: (string | number)[] = knownSpells || (rawSpellcasting ? (rawSpellcasting['spell_ids'] || rawSpellcasting['known_spells'] || []) as (string | number)[] : []);

      let casterType = 'none';
      let casterLevel = 0;
      let spellSlots: Record<number, { current: number; max: number }> = {};
      let spellSaveDC = 0;
      let spellAttackBonus = 0;

      if (rawSpellcasting) {
        casterType = ((rawSpellcasting['caster_type'] as string) || 'none').toLowerCase();
        casterLevel = (rawSpellcasting['casting_level'] as number) || 0;
        spellSaveDC = (rawSpellcasting['save_dc'] as number) || 0;
        spellAttackBonus = (rawSpellcasting['attack_modifier'] as number) || 0;

        const rawSlots = rawSpellcasting['spell_slots'] as Record<string, number> | undefined;
        if (rawSlots) {
          Object.entries(rawSlots).forEach(([level, max]) => {
            spellSlots[Number(level)] = { current: max, max: max };
          });
        }
      }

      result['spellcasting'] = {
        casterType,
        casterLevel,
        spellSlots,
        spellIds,
        spellSaveDC,
        spellAttackBonus
      };
    }

    // 2. Map Equipment
    const rawEquipment = response['equipment'] as Record<string, unknown> | undefined;
    if (rawEquipment) {
      const mapWeaponSlot = (slotData: unknown[]): WeaponSlotData[] => {
        if (!Array.isArray(slotData)) return [];
        return slotData.map(item => {
          const itemRecord = item as Record<string, unknown>;
          return {
            weaponId: itemRecord['weapon_id'] as string | number || itemRecord['id'] as string | number,
            isProficient: !!itemRecord['is_proficient'],
            modifiers: this.mapKeys(itemRecord['modifiers'] || {}) as any
          };
        });
      };

      result['equipment'] = {
        armorId: rawEquipment['armor_id'],
        shieldId: rawEquipment['shield_id'],
        hasShieldEquipped: !!rawEquipment['has_shield_equipped'],
        primarySlot: mapWeaponSlot(rawEquipment['primary_slot'] as unknown[]),
        secondarySlot: mapWeaponSlot(rawEquipment['secondary_slot'] as unknown[]),
        rangedSlot: mapWeaponSlot(rawEquipment['ranged_slot'] as unknown[])
      };

      if (rawEquipment['armor_data']) {
        (result['equipment'] as any).armor = this.mapKeys(rawEquipment['armor_data']);
      }
      if (rawEquipment['shield_data']) {
        (result['equipment'] as any).shield = this.mapKeys(rawEquipment['shield_data']);
      }
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
      if (key === 'formulas' || key === 'spell_dc' || key === 'casting_time' || key === 'description' || key === 'level') return;
      const camelKey = this.toCamelCase(key);
      result[camelKey] = this.mapKeys(response[key]);
    });

    // Handle level explicitly - sometimes it comes as a string from some backends or partials
    if (response['level'] !== undefined) {
      result['level'] = Number(response['level']);
    }

    // Handle description explicitly to avoid any weird recursive mapping if it contains underscores (unlikely but safe)
    if (response['description'] !== undefined) {
      result['description'] = response['description'];
    }

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
    } else if (response['spellDC'] && typeof response['spellDC'] === 'object') {
      // Already camelCased but might need internal mapping
      result['spellDC'] = this.mapKeys(response['spellDC']);
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
      if (['actions', 'legendary_actions', 'multiattacks', 'resistances', 'spellcasting', 'hp', 'details'].includes(key)) {
        return;
      }
      const camelKey = this.toCamelCase(key);
      result[camelKey] = this.mapKeys(response[key]);
    });

    // Handle details nesting if present directly in the object
    if (response['details'] && typeof response['details'] === 'object') {
       const details = response['details'] as Record<string, unknown>;
       if (details['data']) {
          return this.mapMonsterResponse(details['data'] as Record<string, unknown>);
       }
    }

    // Map as_config directly to asConfig
    if (response['as_config']) {
      result['asConfig'] = this.mapKeys(response['as_config']);
    }

    // Map HP
    if (response['hp']) {
      result['hp'] = this.mapKeys(response['hp']);
    }

    // 1. Map Actions
    const rawActions = (response['actions'] || []) as unknown[];
    const actions = Array.isArray(rawActions) ? rawActions.map((a) => {
      const mappedAction = this.mapKeys(a) as Record<string, unknown>;
      // Map damage_blocks if present to damageBlocks
      if ((a as Record<string, unknown>)['damage_blocks']) {
        mappedAction['damageBlocks'] = this.mapKeys((a as Record<string, unknown>)['damage_blocks']);
      }
      return mappedAction;
    }) : [];

    // 2. Map Multiattacks
    const rawMultiattacks = (response['multiattacks'] || []) as unknown[];
    const multiattacks = Array.isArray(rawMultiattacks) ? rawMultiattacks.map((m) => this.mapKeys(m)) : [];

    // 3. Map Legendary Actions
    const rawLegendary = (response['legendary_actions'] || []) as unknown[];
    const legendaryActions = Array.isArray(rawLegendary) ? rawLegendary.map((la) => {
      const mapped = this.mapKeys(la) as Record<string, unknown>;
      if (mapped['action']) {
        const action = this.mapKeys(mapped['action']) as Record<string, unknown>;
        return { ...action, cost: mapped['cost'] };
      }
      return mapped;
    }) : [];

    // 4. Map Resistances
    const resistances: Record<string, string> = {};
    const rawResistances = (response['resistances'] || {}) as Record<string, { resistance?: string } | string>;
    Object.entries(rawResistances).forEach(([type, data]) => {
      if (typeof data === 'string') {
        resistances[type] = data.toLowerCase();
      } else {
        resistances[type] = (data.resistance || 'none').toLowerCase();
      }
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
   * Recursively converts object keys from camelCase to snake_case for the backend.
   */
  serializeKeys<T>(obj: T): any {
    if (Array.isArray(obj)) {
      return obj.map((v) => this.serializeKeys(v));
    } else if (obj !== null && typeof obj === 'object') {
      const result: Record<string, any> = {};
      const typedObj = obj as Record<string, any>;

      // SPECIAL HANDLING: If this is an object that contains 'spellcasting',
      // we want to replace 'spellcasting' with 'known_spells' (array of IDs)
      if (typedObj['spellcasting'] !== undefined && typeof typedObj['spellcasting'] === 'object') {
        const sc = typedObj['spellcasting'];
        const spellIds = sc.spellIds || (Array.isArray(sc.spells) ? sc.spells.map((s: any) => s.id) : []);

        Object.keys(typedObj).forEach((key) => {
          if (key === 'spellcasting' || key === 'armor' || key === 'shield') return; // Stripping hydrated data
          const snakeKey = this.camelToSnake(key);
          result[snakeKey] = this.serializeKeys(typedObj[key]);
        });

        result['known_spells'] = spellIds;
        return result;
      }

      Object.keys(obj).forEach((key) => {
        if (key === 'armor' && typedObj['armorId']) return; // Skip hydrated armor if armorId is present
        if (key === 'shield' && typedObj['shieldId']) return; // Skip hydrated shield if shieldId is present
        const snakeKey = this.camelToSnake(key);
        result[snakeKey] = this.serializeKeys(typedObj[key]);
      });
      return result;
    }
    return obj;
  }

  private camelToSnake(key: string): string {
    // Special mappings for D&D acronyms to prevent over-segmentation
    if (key === 'spellSaveDC') return 'spell_save_dc';
    if (key === 'spellAttackBonus') return 'spell_attack_bonus';
    if (key === 'useHPAverageMonster') return 'use_hp_average_monster';
    if (key === 'useHPAverageCharacter') return 'use_hp_average_character';
    if (key === 'aoeHitsAllEnemies') return 'aoe_hits_all_enemies';
    if (key === 'useWeightedAI') return 'use_weighted_ai';
    if (key === 'debugAI') return 'debug_ai';
    if (key === 'instanceId') return 'instance_id';
    if (key === 'classId') return 'class_id';
    if (key === 'raceId') return 'race_id';
    if (key === 'armorId') return 'armor_id';
    if (key === 'shieldId') return 'shield_id';
    if (key === 'weaponId') return 'weapon_id';

    // General conversion: camelCase -> snake_case
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
    if (str === 'damage_blocks') return 'damageBlocks';

    // SpellFormula fields
    if (str === 'cast_level') return 'castLevel';
    if (str === 'number_of_dice') return 'numberOfDice';
    if (str === 'amount_to_add' || str === 'modifier') return 'amountToAdd';
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
