export type HPVisibilityMode = 'visible' | 'hidden' | 'percentage';

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
}
