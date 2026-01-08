export enum ActionType {
  ATDamage = 'damage',
  ATHeal = 'heal',
  // TODO: There may be additional action types to add in the backend. I don't think it handles more
}
export interface TargetFactorWeights {
  highThreat: number;
  targetPotency: number;
  targetHitability: number;
  vengeance: number;
  lowHP: number;
  casterPriority: number;
  concentrationBreak: number;
  elitePriority: number;
  emergencyHeal: number;
}
export interface UtilityWeights {
  actionWeights: { [key in ActionType]?: number };
  targetFactorWeights: TargetFactorWeights;
  resourceExpenditureWeight: number;
}
