import { Actor, ResistanceType, Condition} from './combatants';
import {SimulationOptions} from './simoptions.model';
import {CharacterConfig} from './configs/character-config.model';
import {MonsterConfig} from './configs/monster-config.model';

export interface CombatantReference {
  name: string;
  instanceId: number;
  type: 'character' | 'monster' | '';
}

export enum AdvantageType {
  Disadvantage = -1,
  Normal = 0,
  Advantage = 1
}

export interface RollResult {
  advantage: AdvantageType;
  dice: number;
  finalRollValue: number;
  finalRolls: number[];
  isCritical: boolean;
  isNaturalOne: boolean;
  modifier: number;
  numberOfDice: number;
  originalRolls: number[];
  rerollEvents: unknown;
  rollType: string;
  total: number;
  isSuccess?: boolean;
}

export interface DiceRoll {
  numberOfDice: number;
  dice: number;
  modifier: number;
  total: number;
  results: number[];
  success?: boolean;
  targetValue?: number;
  advantage?: AdvantageType;
}

export interface EventData {
  actor?: CombatantReference;
  target?: CombatantReference;
  roll?: RollResult;
  targetId?: number;
  actorId?: number;
  choiceType?: string;
  choice?: string | null;
  decision?: string;
  actionName?: string;
  result?: {
    modificationValue: number;
    didHealHp: boolean;
    didHealTempHp: boolean;
    didHpDamage: boolean;
    didTempDamage: boolean;
    newHp: number;
    newTempHp: number;
    originalHp: number;
    originalTempHp: number;
    tempHpUsed: number;
  };
  scores?: {
    utilityScore: number;
    factors: { [key: string]: number } | null;
    topReasons: string[] | null;
  };
  finalHp?: number;
  finalTempHp?: number;
  value?: number;
  attackType?: string;
  diceRoll?: DiceRoll;
  wasModified?: boolean;
  resistanceType?: string;
  resistanceBroken?: boolean;
  note?: string;
  targetAc?: number;
  dc?: number;
  saveSuccess?: boolean;
  isHit?: boolean;
  winner?: string;
  rounds?: number;
  actorStates?: Record<string, ActorStateSnapshot>;
  healing?: Record<string, number>;
}

export interface ActorStateSnapshot {
  currentHp: number;
  tempHp: number;
  conditions: Condition[] | null;
  healthState: string;
}


export enum EventType {
  Initiative = 'initiative',
  Choice = 'choice',
  Attack = 'attack',
  DamageRoll = 'damageroll',
  SavingThrow = 'savingthrow',
  HPModified = 'hp_modified',
  DamageModified = 'damagemodified',
  Round = 'round',
  Turn = 'turn',
  Death = 'death',
  Unconscious = 'unconscious',
  Victory = 'victory',
  Equipment = 'equipment',
  DecisionStart = 'decision_start',
  ActionStart = 'action_start',
  Resolution = 'resolution',
  AttackRoll = 'attack_roll',
  IntermissionHealing = 'intermission_healing',
  CombatStart = 'combat_start',
}

export interface SimulationEvent {
  round: number;
  id: string;
  type: EventType;
  actor?: CombatantReference;
  data?: EventData;
  timestamp?: string;
  sequenceId?: string; // Groups events into a "Turn"
  parentId?: string;   // Defines hierarchy (e.g. Attack -> Damage)
  actorStates?: Record<string, ActorStateSnapshot>;
  healing?: Record<string, number>;
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
  actors: Actor[];
  events: SimulationEvent[];
  initialState?: Record<string, any>;
  actorInitialStates?: Record<string, ActorStateSnapshot>;
  actorConfigs?: Actor[];
}

export interface EncounterResult {
  encounterName: string;
  victoryStatus: string;
  rounds: number;
  seed: { seed1: number; seed2: number };
  logs: SimulationEvent[];
  initialState?: Record<string, ActorStateSnapshot>;
}

export interface IndividualResult {
  runId: number;
  victoryStatus: string;
  totalRounds: number;
  seed: { seed1: number; seed2: number };
  encounterResults: EncounterResult[];
  initialState?: Record<string, any>;
  actorInitialStates?: Record<string, ActorStateSnapshot>;
  actorConfigs?: Actor[];
}

export interface SimulationPerformance {
  executionTimeMs: number;
  executionTimeHuman: string;
  memoryAllocatedMb: number;
  peakGoroutines: number;
}

export interface SimulationResult {
  totalRuns: number;
  characterVictories: number;
  monsterVictories: number;
  otherVictories: number;
  averageRounds: number;
  winRatePercentage: number;
  individualResults: IndividualResult[];
  performance?: SimulationPerformance;
  initialState?: Record<string, any>;
  actorConfigs?: Actor[];
  // For UI compatibility, we'll map the logs from the first run or flatten them
  logs: SimulationLog[];
  count: number;
}

export interface SimulationResponse {
  createdAt: string;
  results: SimulationResult;
  simulationId: string;
  status: string;
  updatedAt: string;
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
  simulationId: string;
  status: SimulationStatus;
  createdAt?: string;
  updatedAt?: string;
  error?: string;
}

export interface IntermissionConfig {
  maxShortRests: number;
  shortRestHealThreshold: number;
  postRestHealThreshold: number;
}

export interface SimulationEncounterConfig {
  name: string;
  monsterIds: number[];
  monsterConfigs: Actor[];
}

export interface SimulationPayload {
  base_options: SimulationOptions;
  character_configs: Actor[];
  encounters: SimulationEncounterConfig[];
  intermission: IntermissionConfig;
  number_of_runs: number;
  max_rounds: number;
  include_logs: boolean;
}
