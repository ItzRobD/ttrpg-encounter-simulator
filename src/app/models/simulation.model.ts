import { Entity, ResistanceType} from './combatants';
import {SimulationOptions} from './simoptions.model';
import {CharacterConfig} from './configs/character-config.model';
import {MonsterConfig} from './configs/monster-config.model';

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
  originalRolls?: number[];
  rerollEvents?: unknown;
  wasRerolled?: boolean;
  name?: string;
}

export interface DiceRoll {
  numberOfDice: number;
  die: number;
  modifier: number;
  total: number;
  results: number[];
  success?: boolean;
  targetValue?: number;
}

export interface EventData {
  actor?: CombatantReference;
  target?: CombatantReference;
  roll?: RollResult;
  diceRoll?: DiceRoll; // Used in 'attack' and 'savingthrow' types
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
  originalValue?: number;
  finalValue?: number;
  wasModified?: boolean;
  resistanceType?: string;
  resistanceBroken?: boolean;
  note?: string;
  attackType?: string;
  winner?: string;
  rounds?: number;
  name?: string;
  numberOfDice?: number;
  die?: string;
  damageType?: string;
  attackBonus?: number;
  damageBonus?: number;
  isRanged?: boolean;
  properties?: string[];
  modifiers?: string[];
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
  Victory = 'victory',
  Equipment = 'equipment',
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

export interface IndividualResult {
  runId: number;
  victoryStatus: string;
  rounds: number;
  seed: { seed1: number; seed2: number };
  logs: SimulationEvent[];
}

export interface SimulationResult {
  totalRuns: number;
  characterVictories: number;
  monsterVictories: number;
  otherVictories: number;
  averageRounds: number;
  winRatePercentage: number;
  individualResults: IndividualResult[];
  // For UI compatibility, we'll map the logs from the first run or flatten them
  logs: SimulationLog[];
  count: number;
}

export interface SimulationRequest {
  payload: SimulationPayload;
}

export enum SimulationStatus {
  Pending = 'pending',
  Running = 'running',
  Completed = 'completed',
  Failed = 'failed'
}

export interface SimulationStatusResponse {
  simulation_id: string;
  status: SimulationStatus;
  created_at?: string;
  updated_at?: string;
  error?: string;
}

export interface SimulationPayload {
  base_options: SimulationOptions;
  character_configs: Entity[];
  monster_ids: number[];
  monster_configs: MonsterConfig[];
  lair_config: any;
  number_of_runs: number;
  max_rounds: number;
  include_logs: boolean;
}
