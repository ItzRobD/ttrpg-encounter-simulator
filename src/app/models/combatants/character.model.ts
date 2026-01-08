import {AbilityScoreProficiency, AbilityScores, Class, EntityState, Race, Spellcasting} from '../core';
import {Entity} from './entity.model';

export interface Character extends Entity {
  race: Race;
  class: Class;
  level: number;
}
