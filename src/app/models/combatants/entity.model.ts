import {AsConfig, EntityState, Spellcasting} from '../core';

export interface Entity {
  id: number;
  instanceId: number;
  name: string;
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
  id: number;
  name: string;
}
