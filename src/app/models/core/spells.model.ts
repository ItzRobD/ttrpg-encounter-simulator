import {DamageType} from './combat.model';

export type SpellType = 'damage' | 'healing' | 'other';

export type CastingTime = 'action' | 'bonus action' | 'reaction' | 'instant' | 'minute' | 'hour';

export type SaveSuccessEffect = 'half' | 'none' | 'other';

export enum Ability {
  Strength = "strength",
  Dexterity = "dexterity",
  Constitution = "constitution",
  Intelligence = "intelligence",
  Wisdom = "wisdom",
  Charisma = "charisma"
}

export enum LevelType {
  Slot = "slot",
  Character = "character"
}

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
  ability: Ability | string;
  onSuccess: SaveSuccessEffect | string;
}

export interface SpellSummary {
  id: number | string;
  name: string;
  isCustom?: boolean;
  isConcentration: boolean;
  isRitual?: boolean;
  level: number;
  spellType: SpellType;
  isAOE: boolean;
  isTouch?: boolean;
  hasDC: boolean;
}

// TODO: Add dc save ability and on success
export interface Spell {
  id: number | string;
  name: string;
  isCustom?: boolean;
  description: string;
  isConcentration: boolean;
  castingTime: CastingTime;
  isRitual: boolean;
  level: number;
  spellType: SpellType;
  isAOE: boolean;
  isTouch?: boolean;
  hasDC: boolean;
  isAutoHit: boolean;
  levelType: LevelType;
  spellDC: SpellDC;
  isInnate?: boolean;
  maxCastsPerDay?: number;
  formulas?: Record<number, SpellFormula>;
}

export interface SpellFormula {
  castLevel: number;
  numberOfDice: number;
  die: number;
  amountToAdd: number;
  useSpellMod: boolean;
  damageType: DamageType;
  averageValue: number;
}

export interface Spellcasting {
  casterType: CasterType;
  casterLevel: number;
  spellSlots: SpellSlots;
  spells: Spell[];
  spellSaveDC: number;
  spellAttackBonus: number;
}
