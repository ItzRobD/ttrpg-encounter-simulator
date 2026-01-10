import {AbilityScoreProficiency, AbilityScores, Class, EntityState, Equipment, Race, Spellcasting} from '../core';
import {Entity} from './entity.model';

export interface Character extends Entity {
  race: Race;
  class: Class;
  level: number;

  equipment: Equipment;
}
