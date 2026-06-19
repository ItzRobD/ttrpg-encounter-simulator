import {
  AbilityScores,
  AbilityScoreProficiency,
  Race,
  Class,
  DamageResistances,
  DiceType,
} from '../core';
import { UtilityWeights } from './utility-weights.model';
export enum HPMethod {
  SetValue = 0,
  Roll = 1,
  Average = 2,
}
export interface Seed {
  seed1: number;
  seed2: number;
}
export interface CharacterConfig {
  name: string;
  classId: Class;
  level: number;
  raceId: Race;
  utilityWeights: UtilityWeights;
  asConfig: { abilityScores: AbilityScores; proficiencies: AbilityScoreProficiency };
  hpMethod: HPMethod;
  hpValue: number;
  seed: Seed;
  resistances: DamageResistances;
  knownSpells: string[];
}
