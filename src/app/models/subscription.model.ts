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
  tier: UserTier;
  usage: {
    monsters: UsageStats;
    characters: UsageStats;
    spells: UsageStats;
    equipment: UsageStats;
  };
}
