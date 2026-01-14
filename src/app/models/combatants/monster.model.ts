import { Entity, EntitySummary } from './entity.model';
import { MonsterActions, MonsterSize, MonsterType, SpecialAbilities, Spellcasting } from '../core';

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
}

export interface MonsterSummary extends EntitySummary {
  cr: number;
  type: MonsterType;
  size: MonsterSize;
  ac: number;
  isLegendary: boolean;
  isSpellcaster: boolean;
}
