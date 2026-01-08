import {AbilityScoreProficiency, AbilityScores, EntityState, Spellcasting} from '../core';

export interface Entity {
  id: number;
  instanceId: number;
  name: string;
  abilityScores: AbilityScores;
  abilityScoreProficiency: AbilityScoreProficiency;
  state: EntityState;
  spellcasting?: Spellcasting;
}
