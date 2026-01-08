import { Action, DiceType, MultiattackOption } from '../../models/core';

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
  const parts = options.map(opt => {
    const action = allActions.find(a => a.actionId === opt.actionId);
    const actionName = action ? action.name : 'Unknown Action';
    return `${actionName} (${opt.count})`;
  });

  return `Multiattack: The monster makes ${parts.join(', ')}.`;
}
