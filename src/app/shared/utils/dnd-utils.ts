import {
  AbilityScores,
  Action,
  DamageResistances,
  DamageType,
  DiceType, Entity,
  MultiattackOption,
  ResistanceType,
  SpecialAbilities,
  Weapon,
  EquipmentItem
} from '../../models';

/**
 * Returns an array of AbilityScoreEntry objects for display purposes.
 * @param scores
 */
export function getAbilityScoreEntries(scores: AbilityScores): AbilityScoreEntry[] {
  if (!scores) return [];
  const order: (keyof AbilityScores)[] = [
    'strength',
    'dexterity',
    'constitution',
    'intelligence',
    'wisdom',
    'charisma',
  ];

  if (!scores) {
    return order.map((key) => ({
      key,
      shortName: getAbilityScoreShortName(key),
      value: 10,
      modifier: 0,
    }));
  }

  return order.map((key) => ({
    key,
    shortName: getAbilityScoreShortName(key),
    value: scores[key],
    modifier: getModifier(scores[key]),
  }));
}

/**
 * Returns an array of damage types that match a specific resistance type.
 */
export function getDamageTypesByResistance(
  resistances: DamageResistances,
  type: ResistanceType,
): DamageType[] {
  return Object.entries(resistances)
    .filter(([_, value]) => value === type)
    .map(([key]) => key as DamageType);
}

/**
 * Formats a dice roll into a string (e.g., "1d8+3").
 */
export function formatDice(count: number, die: DiceType, bonus: number = 0): string {
  if (die === DiceType.D0) return bonus.toString();

  const bonusStr = bonus !== 0 ? (bonus > 0 ? `+${bonus}` : `${bonus}`) : '';
  return `${count}d${die}${bonusStr}`;
}

/**
 * Generates a standard D&D action description if one isn't provided.
 */
export function generateActionDescription(action: Action): string {
  if (action.description) return action.description;

  const parts: string[] = [];
  if (action.damageComponents && action.damageComponents.length > 0) {
    action.damageComponents.forEach((comp, index) => {
      const diceStr = formatDice(comp.numberOfDice, comp.die, comp.amountToAdd);
      const avgDmg = Math.floor((comp.numberOfDice * (comp.die / 2 + 0.5)) + comp.amountToAdd);
      const part = `${avgDmg} (${diceStr}) ${comp.damageType} damage`;
      parts.push(index === 0 ? part : `plus ${part}`);
    });
  }
  const damageStr = parts.join(' ');

  if (action.hasDC) {
    return `Each target must make a DC ${action.dc} ${action.dcAbility} saving throw, taking ${damageStr} on a failed save.`;
  }

  const hitBonus = action.attackBonus >= 0 ? `+${action.attackBonus}` : `${action.attackBonus}`;
  return `Weapon Attack: ${hitBonus} to hit. Hit: (${damageStr}) damage.`;
}

/**
 * Formats a multiattack routine into a readable string.
 */
export function formatMultiattack(entityName: string, options: MultiattackOption[], allActions: Action[]): string {
  if (options.length === 0) return '';

  const parts = options.map((opt) => {
    const action = allActions.find((a) => a.actionId === opt.actionId);
    const actionName = action ? action.name : 'Unknown Action';
    return `${opt.count} times with ${actionName}`;
  });

  let joinedParts: string;
  if (parts.length === 1) {
    joinedParts = parts[0];
  } else if (parts.length === 2) {
    joinedParts = parts.join(' and ');
  } else {
    joinedParts = parts.slice(0, -1).join(', ') + ', and ' + parts.slice(-1);
  }

  return `Multiattack: The ${entityName} attacks ${joinedParts}.`;
}

/**
 * Formats a monster action's full stat line.
 */
export function formatMonsterAction(action: Action): string {
  const toHit = action.attackBonus;
  const parts: string[] = [];

  if (action.damageComponents && action.damageComponents.length > 0) {
    action.damageComponents.forEach((comp, index) => {
      const diceStr = formatDice(comp.numberOfDice, comp.die, comp.amountToAdd);
      const avgDmg = Math.floor((comp.numberOfDice * (comp.die / 2 + 0.5)) + comp.amountToAdd);
      const part = `${avgDmg} (${diceStr}) ${comp.damageType} damage`;
      parts.push(index === 0 ? part : `plus ${part}`);
    });
  }

  let base = `${formatModifier(toHit)} to hit.`;
  if (parts.length > 0) {
    base += ` Hit: ${parts.join(' ')}.`;
  }

  if (action.hasDC) {
    base += ` DC ${action.dc} ${getAbilityScoreShortName(action.dcAbility || '').toUpperCase()} save.`;
  }

  return base;
}

/**
 * Calculates the modifier for an ability score (e.g., 14 -> +2).
 */
export function getModifier(score: number): number {
  return Math.floor((score - 10) / 2);
}

/**
 * Formats a modifier for display (e.g., 2 -> "+2", -1 -> "-1").
 */
export function formatModifier(value: number): string {
  return value >= 0 ? `+${value}` : `${value}`;
}

/**
 * Calculates the proficiency bonus based on level/CR.
 */
export function getProficiencyBonus(levelOrCr: number): number {
  if (levelOrCr < 5) return 2;
  if (levelOrCr < 9) return 3;
  if (levelOrCr < 13) return 4;
  if (levelOrCr < 17) return 5;
  if (levelOrCr < 21) return 6;
  if (levelOrCr < 25) return 7;
  if (levelOrCr < 29) return 8;
  return 9;
}

/**
 * Formats an ability score with its modifier (e.g., 14 -> "14 (+2)").
 */
export function formatAbilityScore(score: number): string {
  return `${score} (${formatModifier(getModifier(score))})`;
}

/**
 * Formats ability scores into a readable string (e.g., "Strength: 14 (+2), Dexterity: 12 (+1)").
 * @param scores
 */
export function formatAbilityScores(scores: AbilityScores): string {
  return Object.entries(scores)
    .map(([key, value]) => `${key}: ${formatAbilityScore(value)}`)
    .join(', ');
}

/**
 * Returns the order of ability scores for display purposes (e.g., [14, 12, 10, 8, 6, 4]).
 * Order is Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma.
 * @param scores
 */
export function getAbilityScoresOrder(scores: AbilityScores): number[] {
  if (!scores) return [10, 10, 10, 10, 10, 10];
  return [
    scores.strength || 10,
    scores.dexterity || 10,
    scores.constitution || 10,
    scores.intelligence || 10,
    scores.wisdom || 10,
    scores.charisma || 10,
  ];
}

/**
 * Returns the short name for an ability score (e.g., 'Str', 'Dex', 'Con', etc.).
 * @param ability
 */
export function getAbilityScoreShortName(ability: string): string {
  switch (ability) {
    case 'strength':
      return 'Str';
    case 'dexterity':
      return 'Dex';
    case 'constitution':
      return 'Con';
    case 'intelligence':
      return 'Int';
    case 'wisdom':
      return 'Wis';
    case 'charisma':
      return 'Cha';
    default:
      return '';
  }
}

export interface AbilityScoreEntry {
  key: keyof AbilityScores;
  shortName: string;
  value: number;
  modifier: number;
}

/**
 * Returns the relevant ability modifier for a weapon based on its properties.
 */
export function getWeaponAbilityModifier(e: Entity, weapon: Weapon): number {
  if (!e.asConfig?.abilityScores) return 0;
  const str = getModifier(e.asConfig.abilityScores.strength);
  const dex = getModifier(e.asConfig.abilityScores.dexterity);

  if (weapon.properties.isRanged || weapon.properties.isOnlyRanged) return dex;
  if (weapon.properties.isFinesse) return Math.max(str, dex);
  return str;
}

/**
 * Formats a weapon's full stat line.
 */
export function formatWeaponData(e: Entity, weapon: Weapon): string {
  const levelOrCr = 'level' in e ? (e as { level: number }).level : (e as unknown as { cr: number }).cr || 1;
  const proficiency = getProficiencyBonus(levelOrCr);
  const abilityMod = getWeaponAbilityModifier(e, weapon);

  // To Hit: Ability Mod + Proficiency (if proficient) + Weapon Magic Bonus
  const isProficient = weapon.isProficient !== undefined ? weapon.isProficient : true;
  const toHit = abilityMod + (isProficient ? proficiency : 0) + weapon.modifiers.attackBonus;

  // Damage Dice String: e.g., "1d8+3"
  // Note: In D&D, damage bonus usually includes the Ability Mod + Magic Bonus
  const totalDmgBonus = abilityMod + weapon.modifiers.damageBonus;
  const diceStr = formatDice(weapon.numberOfDice, weapon.die, totalDmgBonus);

  // Average Damage: (DieAvg * Count) + Bonus
  const avgDmg = Math.floor((weapon.numberOfDice * (weapon.die / 2 + 0.5)) + totalDmgBonus);

  return `${weapon.name}. ${formatModifier(toHit)} to hit. Damage: ${avgDmg} (${diceStr}) ${weapon.damageType} damage.`;
}

/**
 * Returns a detail string for an equipment item based on its type.
 */
export function getEquipmentDetail(item: EquipmentItem): string {
  if ('damageType' in item) {
    // Weapon
    return `${formatDice(item.numberOfDice, item.die, 0)} ${item.damageType}`;
  } else {
    // Armor
    return `AC ${item.ac}`;
  }
}

/**
 * Returns an array of Title Case names of active special abilities.
 */
export function getSpecialAbilityNames(abilities: SpecialAbilities | undefined): string[] {
  const names: string[] = [];
  if (!abilities) return names;

  const entries = Object.entries(abilities);
  for (const [key, value] of entries) {
    if (value === false || value === 0 || value === 'D0' || value === DiceType.D0) continue;

    // These are the specific keys that have numeric/complex values
    const complexKeys = [
      'legendaryResistanceCount',
      'divineEminenceNumDice',
      'martialAdvantageNumDice',
      'sneakAttackNumDice',
      'regenerationValue',
      'relentlessThreshold',
      'berserkThreshold',
      'limitedMagicImmunityLevel',
      'deathBurstNumDice',
      'deathThroesNumDice',
      'fireAuraNumDice',
      'heatedBodyNumDice',
      'corrosiveFormNumDice',
      'consumeLifeDie'
    ];

    if (complexKeys.includes(key)) {
      // Map back to a readable name
      let name = '';
      switch (key) {
        case 'legendaryResistanceCount': name = 'Legendary Resistance'; break;
        case 'divineEminenceNumDice': name = 'Divine Eminence'; break;
        case 'martialAdvantageNumDice': name = 'Martial Advantage'; break;
        case 'sneakAttackNumDice': name = 'Sneak Attack'; break;
        case 'regenerationValue': name = 'Regeneration'; break;
        case 'relentlessThreshold': name = 'Relentless'; break;
        case 'berserkThreshold': name = 'Berserk'; break;
        case 'limitedMagicImmunityLevel': name = 'Limited Magic Immunity'; break;
        case 'deathBurstNumDice': name = 'Death Burst'; break;
        case 'deathThroesNumDice': name = 'Death Throes'; break;
        case 'fireAuraNumDice': name = 'Fire Aura'; break;
        case 'heatedBodyNumDice': name = 'Heated Body'; break;
        case 'corrosiveFormNumDice': name = 'Corrosive Form'; break;
        case 'consumeLifeDie': name = 'Consume Life'; break;
      }
      if (name) names.push(name);
    } else if (value === true) {
      names.push(formatCamelCase(key));
    }
  }
  return names;
}

/**
 * Formats monster special abilities into an array of readable strings.
 */
export function getFormattedSpecialAbilities(abilities: SpecialAbilities | undefined): string[] {
  const formatted: string[] = [];

  if (!abilities) return formatted;

  const entries = Object.entries(abilities);

  for (const [key, value] of entries) {
    // Skip if value is false, 0, or D0
    if (value === false || value === 0 || value === 'D0' || value === DiceType.D0) continue;

    switch (key) {
      case 'legendaryResistanceCount':
        formatted.push(`Legendary Resistance: ${value} Uses`);
        break;
      case 'divineEminenceNumDice':
        formatted.push(`Divine Eminence: ${value}d6`);
        break;
      case 'martialAdvantageNumDice':
        formatted.push(`Martial Advantage: ${value}d6`);
        break;
      case 'sneakAttackNumDice':
        formatted.push(`Sneak Attack: ${value}d6`);
        break;
      case 'regenerationValue':
        formatted.push(`Regeneration: ${value} HP`);
        break;
      case 'relentlessThreshold':
        formatted.push(`Relentless: <=${value} HP`);
        break;
      case 'berserkThreshold':
        formatted.push(`Berserk: ${value} HP`);
        break;
      case 'limitedMagicImmunityLevel':
        formatted.push(`Limited Magic Immunity: Level ${value} or lower`);
        break;
      case 'deathBurstNumDice':
        formatted.push(`Death Burst (DC ${abilities.deathBurstDC}): ${value}d8 ${abilities.deathBurstDamageType || ''}`);
        break;
      case 'deathThroesNumDice':
        formatted.push(`Death Throes (DC ${abilities.deathThroesDC}): ${value}d6`);
        break;
      case 'fireAuraNumDice':
        formatted.push(`Fire Aura: ${value}d6`);
        break;
      case 'heatedBodyNumDice':
        formatted.push(`Heated Body: ${value}d10`);
        break;
      case 'corrosiveFormNumDice':
        formatted.push(`Corrosive Form: ${value}d8`);
        break;
      case 'consumeLifeDie':
        formatted.push(`Consume Life: 3d${value}`);
        break;
      default:
        // Handle boolean traits
        if (value === true) {
          formatted.push(formatCamelCase(key));
        }
        break;
    }
  }

  return formatted;
}

/**
 * Formats a camelCase string to Title Case (e.g., "magicResistance" -> "Magic Resistance").
 */
function formatCamelCase(str: string): string {
  const result = str.replace(/([A-Z])/g, ' $1');
  return result.charAt(0).toUpperCase() + result.slice(1);
}
