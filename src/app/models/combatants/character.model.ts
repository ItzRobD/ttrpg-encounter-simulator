import {AbilityScoreProficiency, AbilityScores, Class, EntityState, Equipment, Race, SpecialAbilities, Spellcasting} from '../core';
import {Entity, EntitySummary} from './entity.model';

export interface Character extends Entity {
  race: Race;
  class: Class;
  level: number;
  classId?: number;
  raceId?: number;
  equipment?: Equipment;
  specialAbilities?: SpecialAbilities;
}

export interface CharacterSummary extends EntitySummary {
  race: Race;
  class: Class;
  level: number;
  classId?: number;
  raceId?: number;
  isSpellcaster: boolean;
}
