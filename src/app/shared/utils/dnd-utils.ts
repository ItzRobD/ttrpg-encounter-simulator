import {
  AbilityScores,
  Action,
  DamageResistances,
  DamageType,
  DiceType, Entity,
  MultiattackOption,
  ResistanceType,
  Weapon
} from '../../models';

/**
 * Returns an array of AbilityScoreEntry objects for display purposes.
 * @param scores
 */
export function getAbilityScoreEntries(scores: AbilityScores): AbilityScoreEntry[] {
  const order: (keyof AbilityScores)[] = [
    'strength',
    'dexterity',
    'constitution',
    'intelligence',
    'wisdom',
    'charisma',
  ];
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

  const diceStr = formatDice(action.numberOfDice, action.die, action.amountToAdd);

  if (action.hasDC) {
    return `Each target must make a DC ${action.dc} ${action.dcAbility} saving throw, taking ${diceStr} ${action.damageType} damage on a failed save.`;
  }

  const hitBonus = action.attackBonus >= 0 ? `+${action.attackBonus}` : `${action.attackBonus}`;
  return `Weapon Attack: ${hitBonus} to hit. Hit: (${diceStr}) ${action.damageType} damage.`;
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

  let joinedParts = '';
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
  const diceStr = formatDice(action.numberOfDice, action.die, action.amountToAdd);
  const avgDmg = Math.floor((action.numberOfDice * (action.die / 2 + 0.5)) + action.amountToAdd);

  return `${formatModifier(toHit)} to hit. Damage: ${avgDmg} (${diceStr}) ${action.damageType} damage.`;
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
export function getProficiencyBonus(level: number): number {
  return Math.ceil(level / 4) + 1;
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
  return [
    scores.strength,
    scores.dexterity,
    scores.constitution,
    scores.intelligence,
    scores.wisdom,
    scores.charisma,
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
  const str = getModifier(e.abilityScores.strength);
  const dex = getModifier(e.abilityScores.dexterity);

  if (weapon.properties.isRanged || weapon.properties.isOnlyRanged) return dex;
  if (weapon.properties.isFinesse) return Math.max(str, dex);
  return str;
}

/**
 * Formats a weapon's full stat line.
 */
export function formatWeaponData(e: Entity, weapon: Weapon): string {
  const proficiency = getProficiencyBonus('level' in e ? (e as any).level : (e as any).cr || 1);
  const abilityMod = getWeaponAbilityModifier(e, weapon);

  // To Hit: Ability Mod + Proficiency (if proficient) + Weapon Magic Bonus
  // Note: Assuming character is proficient with their equipped weapons for simplicity here
  const toHit = abilityMod + proficiency + weapon.modifiers.attackBonus;

  // Damage Dice String: e.g., "1d8+3"
  // Note: In D&D, damage bonus usually includes the Ability Mod + Magic Bonus
  const totalDmgBonus = abilityMod + weapon.modifiers.damageBonus;
  const diceStr = formatDice(weapon.numberOfDice, weapon.die, totalDmgBonus);

  // Average Damage: (DieAvg * Count) + Bonus
  const avgDmg = Math.floor((weapon.numberOfDice * (weapon.die / 2 + 0.5)) + totalDmgBonus);

  return `${weapon.name}. ${formatModifier(toHit)} to hit. Damage: ${avgDmg} (${diceStr}) ${weapon.damageType} damage.`;
}
