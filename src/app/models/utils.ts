import { Entity, Character, Monster } from './combatants';

export function isCharacter(entity: Entity | null | undefined): entity is Character {
  return !!entity && 'classId' in entity && 'level' in entity;
}

export function isMonster(entity: Entity | null | undefined): entity is Monster {
  return !!entity && 'cr' in entity && !('classId' in entity);
}

export function getAC(entity: Entity): string {
  if (isCharacter(entity)) {
    let baseAC = 10;
    if (entity.equipment?.armor) {
      baseAC = entity.equipment.armor.ac;
      if (entity.equipment.armor.dexBonus) {
        const dexMod = Math.floor(((entity.asConfig?.abilityScores?.dexterity || 10) - 10) / 2);
        const maxBonus = entity.equipment.armor.maxBonus ? 2 : 10; // Max 2 for medium, no limit for light (effectively)
        baseAC += Math.min(dexMod, maxBonus);
      }
    } else {
      // Unarmored AC = 10 + Dex
      const dexMod = Math.floor(((entity.asConfig?.abilityScores?.dexterity || 10) - 10) / 2);
      baseAC += dexMod;
    }

    if (entity.equipment?.hasShieldEquipped) {
      if (entity.equipment.shield) {
        baseAC += entity.equipment.shield.ac;
      } else {
        baseAC += 2; // Default shield bonus
      }
    }

    return baseAC.toString();
  } else if (isMonster(entity)) {
    return entity.ac?.toString() || '??';
  }

  console.log('Unable to get AC for ', entity)
  return '??'
}
