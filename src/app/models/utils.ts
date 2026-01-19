import { Entity, Character, Monster } from './combatants';

export function isCharacter(entity: Entity | null | undefined): entity is Character {
  return !!entity && 'classId' in entity && 'level' in entity;
}

export function isMonster(entity: Entity | null | undefined): entity is Monster {
  return !!entity && 'cr' in entity && !('classId' in entity);
}
