import {
  Action,
  DamageResistances,
  DamageType,
  DiceType,
  MultiattackOption,
  ResistanceType,
  Weapon,
  EquipmentItem,
  Actor,
  Feature,
  AbilityScores
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
    shortName: getAbilityScoreShortName(key as string),
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
  if (action.damageBlocks && action.damageBlocks.length > 0) {
    action.damageBlocks.forEach((comp, index) => {
      const diceStr = formatDice(comp.numberOfDice, comp.die, comp.amountToAdd);
      const avgDmg = Math.floor(comp.numberOfDice * (comp.die / 2 + 0.5) + comp.amountToAdd);
      const part = `${avgDmg} (${diceStr}) ${comp.damageType} damage`;
      parts.push(index === 0 ? part : `plus ${part}`);
    });
  } else if ((action as any).numberOfDice) {
    // Fallback for legacy data
    const diceStr = formatDice(
      (action as any).numberOfDice,
      (action as any).die,
      (action as any).amountToAdd
    );
    const avgDmg = Math.floor(
      (action as any).numberOfDice * ((action as any).die / 2 + 0.5) + (action as any).amountToAdd
    );
    parts.push(`${avgDmg} (${diceStr}) ${(action as any).damageType} damage`);
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
export function formatMultiattack(actorName: string, options: MultiattackOption[], allActions: Action[]): string {
  if (options.length === 0) return '';

  const parts = options.map((opt) => {
    const action = allActions.find((a) => {
      const aId = (a.actionId || (a as any).id || (a as any).ID)?.toString();
      return aId === opt.actionId?.toString();
    });
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

  return `Multiattack: The ${actorName} attacks ${joinedParts}.`;
}

/**
 * Formats a monster action's full stat line.
 */
export function formatMonsterAction(action: Action): string {
  const toHit = action.attackBonus;
  const parts: string[] = [];

  if (action.damageBlocks && action.damageBlocks.length > 0) {
    action.damageBlocks.forEach((comp, index) => {
      const diceStr = formatDice(comp.numberOfDice, comp.die, comp.amountToAdd);
      const avgDmg = Math.floor((comp.numberOfDice * (comp.die / 2 + 0.5)) + comp.amountToAdd);
      const part = `${avgDmg} (${diceStr}) ${comp.damageType} damage`;
      parts.push(index === 0 ? part : `plus ${part}`);
    });
  }

  let base = '';
  if (toHit > 0) {
    base = `${formatModifier(toHit)} to hit. `;
  }

  if (parts.length > 0) {
    base += `Hit: ${parts.join(' ')}.`;
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
export function getWeaponAbilityModifier(actor: Actor, weapon: Weapon): number {
  const abilities = actor.abilities || (actor as any).asConfig;
  if (!abilities?.abilityScores) return 0;
  const str = getModifier(abilities.abilityScores.strength);
  const dex = getModifier(abilities.abilityScores.dexterity);

  if (weapon.properties.isRanged || weapon.properties.isOnlyRanged) return dex;
  if (weapon.properties.isFinesse) return Math.max(str, dex);
  return str;
}

/**
 * Formats a weapon's full stat line.
 */
export function formatWeaponData(actor: Actor, weapon: Weapon): string {
  const levelOrCr = 'level' in actor ? (actor as any).level : (actor as any).metadata?.cr || 1;
  const proficiency = getProficiencyBonus(levelOrCr);
  const abilityMod = getWeaponAbilityModifier(actor, weapon);

  // To Hit: Ability Mod + Proficiency (if proficient) + Weapon Magic Bonus
  const isProficient = weapon.isProficient !== undefined ? weapon.isProficient : true;
  const toHit = abilityMod + (isProficient ? proficiency : 0) + (weapon.modifiers?.attackBonus || 0);

  // Damage Dice String: e.g., "1d8+3"
  let diceStr = '';
  let avgDmg = 0;

  if (weapon.damageBlocks && weapon.damageBlocks.length > 0) {
    const components = weapon.damageBlocks.map((c, i) => {
      // For the first component, we add the ability modifier and magic damage bonus
      const bonus = i === 0 ? abilityMod + (weapon.modifiers?.damageBonus || 0) : 0;
      const totalBonus = (c.modifier || 0) + bonus;
      avgDmg += Math.floor(c.numberOfDice * (c.die / 2 + 0.5) + totalBonus);
      return formatDice(c.numberOfDice, c.die, totalBonus) + (c.damageType ? ` ${c.damageType}` : '');
    });
    diceStr = components.join(' + ');
  }

  return `${weapon.name}. ${formatModifier(toHit)} to hit. Damage: ${avgDmg} (${diceStr}).`.replace(/\s+/g, ' ');
}

/**
 * Returns a detail string for an equipment item based on its type.
 */
export function getEquipmentDetail(item: EquipmentItem): string {
  const inner = (item as any).weapon || (item as any).armor || item;

  // Weapon damage handling: rely exclusively on damageBlocks
  if ('damageBlocks' in inner && Array.isArray(inner.damageBlocks)) {
    if (inner.damageBlocks.length === 0) return 'No damage details';

    return inner.damageBlocks
      .map((c: any) => {
        const dicePart = formatDice(c.numberOfDice, c.die, c.modifier || 0);
        const typePart = c.damageType ? ` ${c.damageType}` : '';
        return `${dicePart}${typePart}`;
      })
      .join(' + ');
  }

  // Armor AC handling
  if ('ac' in inner) {
    let detail = `AC ${inner.ac}`;
    if (inner.modifier && inner.modifier !== 0) {
      detail += ` (${formatModifier(inner.modifier)})`;
    }
    return detail;
  }

  return 'No detail available';
}

/**
 * Returns an array of Title Case names of active special abilities.
 */
export function getSpecialAbilityNames(features: Feature[] | undefined): string[] {
  if (!features) return [];
  return features.map(f => f.name);
}

/**
 * Formats monster special abilities into an array of readable strings.
 */
export function getFormattedSpecialAbilities(features: Feature[] | undefined): string[] {
  if (!features) return [];
  return features.map(f => `${f.name}: ${f.description}`);
}

/**
 * Formats a camelCase string to Title Case (e.g., "magicResistance" -> "Magic Resistance").
 */
function formatCamelCase(str: string): string {
  const result = str.replace(/([A-Z])/g, ' $1');
  return result.charAt(0).toUpperCase() + result.slice(1);
}
