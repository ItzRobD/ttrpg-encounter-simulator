import { Entity } from './entity.model';
import {MonsterActions, MonsterSize, MonsterType, SpecialAbilities, Spellcasting} from '../core';

export interface Monster extends Entity {
  size: MonsterSize;
  type: MonsterType;
  cr: number;
  proficiencyBonus: number;
  isLegendary: boolean;
  isSpellcaster: boolean;
  isInnateSpellcaster: boolean;
  specialAbilities: SpecialAbilities;
  monsterActions: MonsterActions;
  spellcasting?: Spellcasting;
}
