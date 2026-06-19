import { UtilityWeights } from './utility-weights.model';

export type ActionPreference = 'melee' | 'ranged' | 'spell' | 'balanced' | string;
export type TargetPriority = 'lowest_hp' | 'highest_threat' | 'spellcaster' | 'no_preference' | string;

export interface BehaviorConfig {
  actionPreference: ActionPreference;
  secondaryActionPreference?: ActionPreference;
  targetPriority: TargetPriority;
  secondaryTargetPriority?: TargetPriority;
  weights?: UtilityWeights;
}
