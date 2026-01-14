import {AbilityScoreProficiency, AbilityScores, Class, EntityState, Equipment, Race, Spellcasting} from '../core';
import {Entity, EntitySummary} from './entity.model';

export interface Character extends Entity {
  race: Race;
  class: Class;
  level: number;
  equipment?: Equipment;
}

export interface CharacterSummary extends EntitySummary {
  race: Race;
  class: Class;
  level: number;
  isSpellcaster: boolean;
}
