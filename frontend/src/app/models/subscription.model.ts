export enum UserTier {
  Free = 'free',
  Premium = 'premium',
  Pro = 'pro'
}

export interface UsageStats {
  current: number;
  max: number;
}

export interface UserLimits {
  userId?: string;
  tier: UserTier | string;
  characters: UsageStats;
  equipment: UsageStats;
  monsters: UsageStats;
  spells: UsageStats;
}
