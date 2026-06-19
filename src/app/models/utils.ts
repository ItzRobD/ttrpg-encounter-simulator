import { Actor, isCharacter, isMonster } from './combatants';

export function getAC(actor: Actor): string {
  // New Actor model AC
  const ac = actor.ac;
  if (ac !== undefined && ac !== null && ac !== 0) {
    return ac.toString();
  }

  if (isCharacter(actor)) {
    let baseAC = 10;
    const armor = actor.equipment?.armor;
    if (armor) {
      baseAC = armor.ac;
      if (armor.dexBonus) {
        const dexterity = actor.abilities?.abilityScores?.dexterity || 10;
        const dexMod = Math.floor((dexterity - 10) / 2);
        const maxBonus = armor.maxBonus ? 2 : 10;
        baseAC += Math.min(dexMod, maxBonus);
      }
    } else {
      // Unarmored AC = 10 + Dex
      const dexterity = actor.abilities?.abilityScores?.dexterity || 10;
      const dexMod = Math.floor((dexterity - 10) / 2);
      baseAC += dexMod;
    }

    if (actor.equipment?.hasShieldEquipped) {
      if (actor.equipment.shield) {
        baseAC += actor.equipment.shield.ac;
      } else {
        baseAC += 2; // Default shield bonus
      }
    }

    return baseAC.toString();
  } else if (isMonster(actor)) {
    return actor.ac?.toString() || '??';
  }

  console.log('Unable to get AC for ', actor)
  return '??'
}
