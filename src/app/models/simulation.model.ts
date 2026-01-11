import { Entity } from './combatants';

export interface CombatantReference {
  name: string;
  instanceId: number;
  type: 'character' | 'monster' | '';
}

export interface RollResult {
  round: number;
  timestamp: string;
  id: string;
  rollType: string;
  numberOfDice: number;
  die: string;
  finalRollValue: number;
  finalRolls: number[];
  modifier: number;
  total: number;
  advantage: 'Normal' | 'Advantage' | 'Disadvantage';
  isCritical: boolean;
  isNaturalOne: boolean;
  isSuccess: boolean;
  targetValue: number;
}

export interface EventData {
  actor?: CombatantReference;
  target?: CombatantReference;
  roll?: RollResult;
  diceRoll?: any; // Used in 'attack' and 'savingthrow' types
  choiceType?: string;
  choice?: string | null;
  scores?: {
    utilityScore: number;
    factors: { [key: string]: number } | null;
    topReasons: string[] | null;
  };
  value?: number;
  originalHp?: number;
  finalHp?: number;
  originalTempHp?: number;
  finalTempHp?: number;
  note?: string;
}


export enum EventType {
  Initiative = 'initiative',
  Choice = 'choice',
  Attack = 'attack',
  DamageRoll = 'damageroll',
  SavingThrow = 'savingthrow',
  HPModified = 'hpmodified',
  DamageModified = 'damagemodified',
  Round = 'round',
  Turn = 'turn',
  Death = 'death',
  Unconscious = 'unconscious',
}

export interface SimulationEvent {
  round: number;
  id: string;
  type: EventType;
  data: EventData;
  timestamp?: string;
  sequenceId?: string; // Groups events into a "Turn"
  parentId?: string;   // Defines hierarchy (e.g. Attack -> Damage)
}

/**
 * Processed structure for the TreeTable/Timeline
 */
export interface TimelineNode {
  data: SimulationEvent;
  children?: TimelineNode[];
  expanded?: boolean;
}

export interface SimulationLog {
  entities: Entity[];
  events: SimulationEvent[];
}

export interface SimulationResult {
  logs: SimulationLog[];
  count: number;
}
