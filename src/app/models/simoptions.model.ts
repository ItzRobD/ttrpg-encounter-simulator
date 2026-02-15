export type HPVisibilityMode = 'visible' | 'hidden' | 'percentage';

export type MultiattackFollowUpPolicy = 'aggressive' | 'random' | 'smart';
export type ActionSelectionPolicy = 'weighted' | 'random' | 'highest_damage' | 'utility';

export interface SimulationOptions {
  seed: { seed1: number; seed2: number };
  useHPAverageMonster: boolean;
  useHPAverageCharacter: boolean;
  canMonstersCrit: boolean;
  canCharactersCrit: boolean;
  hasIncreasedCrits: boolean;
  useImprovedCritical: boolean;
  charactersAlwaysUpcast: boolean;
  monstersAlwaysUpcast: boolean;
  allowCharacterHeals: boolean;
  allowMonsterHeals: boolean;
  aoeHitsAllEnemies: boolean;
  characterHealThresholdPct: number;
  monsterHealThresholdPct: number;
  characterEmergencyThresholdPct: number;
  monsterEmergencyThresholdPct: number;
  limitedLegendaryActions: boolean;
  allowLairActions: boolean;
  allowDragonbornBreathAttack: boolean;
  enableClassFeatures: boolean;
  enableRacialFeatures: boolean;
  barbarianAlwaysRecklessAttack: boolean;
  paladinAlwaysSmite: boolean;
  paladinUseHighestSmiteSlot: boolean;
  useMassiveDamage: boolean;
  enableSpecialAbilities: boolean;
  monsterDeathEffectsHitAllies: boolean;
  alwaysUseSneakAttack: boolean;

  // Premium AI Options
  useWeightedAI: boolean;
  debugAI: boolean;
  hpVisibilityMode: HPVisibilityMode;
  enableMonsterNoise: boolean;
  monsterNoiseWeight: number;

  includeStateSnapshots: boolean;
  maxLoggedRuns: number;

  aoeTargetThreshold?: number;
  multiattackPolicy?: MultiattackFollowUpPolicy;
  actionSelectionPolicy?: ActionSelectionPolicy;
  disableMonsterTurns?: boolean;
  disableCharacterTurns?: boolean;
  disableLairTurns?: boolean;
}

export interface SimulationConfig {
  numberOfRuns: number;
  maxRounds: number;
  includeLogs: boolean;
}
