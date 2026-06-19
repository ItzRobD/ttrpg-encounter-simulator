import { Ability } from './spells.model';
import { MonsterSize, MonsterType, DragonbornColor } from './definitions.model';
import { DamageType } from './combat.model';

export interface SpellcasterMetadata {
  isSpellcaster?: boolean;
  isInnateCaster?: boolean;
  spellcastingAbility?: Ability | string;
  spellcastingLevel?: number;
}

export interface Metadata {
  level: number;
  cr: number;

  // Character
  classId?: number;
  raceId?: number;
  dragonbornColor?: DragonbornColor;
  dragonbornDamageType?: DamageType;

  // Monster
  size?: MonsterSize;
  type?: MonsterType;
  isLegendary?: boolean;
  maxLegendaryActions?: number;
  spellcasterMetadata?: SpellcasterMetadata;

  // Precalculated Stats
  averageOffensiveValue?: number;
  highestOffensiveValue?: number;
}
