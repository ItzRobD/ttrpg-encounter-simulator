import {
  AbilityScoreProficiency,
  AbilityScores,
  Class, DragonbornColor,
  EntityState,
  Equipment,
  Race,
  SpecialAbilities,
  Spellcasting
} from '../core';
import {Entity, EntitySummary} from './entity.model';

export interface Character extends Entity {
  raceId: number;
  race?: string; // Virtual property for UI display
  dragonbornColor?: DragonbornColor;
  classId: number;
  class?: string; // Virtual property for UI display
  level: number;
  equipment?: Equipment;
  specialAbilities?: SpecialAbilities;
}

export interface CharacterSummary extends EntitySummary {
  raceId: number;
  race?: string; // Virtual property for UI display
  dragonbornColor?: DragonbornColor;
  classId: number;
  class?: string; // Virtual property for UI display
  level: number;
  isSpellcaster: boolean;
  armorId?: number | string;
  armorName?: string;
  weaponIds?: (number | string)[];
  weapons?: string[];
}
