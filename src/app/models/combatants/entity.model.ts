import {AsConfig, EntityState, Spellcasting} from '../core';

export interface Entity {
  id: number | string;
  instanceId: number;
  name: string;
  isCustom?: boolean;
  asConfig: AsConfig;
  state: EntityState;
  spellcasting?: Spellcasting;
  ac?: number;
  hp?: {
    hpSetMethod: number;
    value: number;
    hpAverage: number;
    numberOfDice: number;
    hitDie: number;
    amountToAdd: number;
    modifier: number;
  };
}

export interface EntitySummary {
  id: number | string;
  name: string;
  isCustom?: boolean;
}
