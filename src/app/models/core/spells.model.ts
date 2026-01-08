export type SpellType = 'damage' | 'healing';

export type CastingTime = 'action' | 'bonus action' | 'reaction' | 'instant';

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

export interface Spell {
  id: number;
  name: string;
  description: string;
  isConcentration: boolean;
  castingTime: CastingTime;
  level: number;
  spellType: SpellType;
  isAOE: boolean;
  hasDC: boolean;
  isAutoHit: boolean;
  isInnate?: boolean;
}

export interface Spellcasting {
  casterType: CasterType;
  casterLevel: number;
  spellSlots: SpellSlots;
  spells: Spell[];
  spellSaveDC: number;
  spellAttackBonus: number;
}
