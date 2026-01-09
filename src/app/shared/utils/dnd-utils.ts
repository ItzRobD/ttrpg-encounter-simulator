import { AbilityScores, Action, DiceType, MultiattackOption } from '../../models';

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
export function formatMultiattack(options: MultiattackOption[], allActions: Action[]): string {
  const parts = options.map((opt) => {
    const action = allActions.find((a) => a.actionId === opt.actionId);
    const actionName = action ? action.name : 'Unknown Action';
    return `${actionName} (${opt.count})`;
  });

  return `Multiattack: The monster makes ${parts.join(', ')}.`;
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
