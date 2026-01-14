export type SpellType = 'damage' | 'healing' | 'other';

export type CastingTime = 'action' | 'bonus action' | 'reaction' | 'instant' | 'minute' | 'hour';

export enum CasterType {
  Full = "full",
  Half = "half",
  Third = "third",
  Warlock = "warlock",
  None = "none"
}

export interface SpellSlots {
  [level: number]: {
    current: number;
    max: number;
  };
}

export interface SpellDC {
  ability: string;
  onSuccess: string;
}

export interface Spell {
  id: number;
  name: string;
  description: string;
  isConcentration: boolean;
  castingTime: CastingTime;
  isRitual: boolean;
  level: number;
  spellType: SpellType;
  isAOE: boolean;
  hasDC: boolean;
  isAutoHit: boolean;
  levelType: string;
  spellDC: SpellDC;
  isInnate?: boolean;
  maxCastsPerDay?: number;
}

export interface Spellcasting {
  casterType: CasterType;
  casterLevel: number;
  spellSlots: SpellSlots;
  spells: Spell[];
  spellSaveDC: number;
  spellAttackBonus: number;
}
