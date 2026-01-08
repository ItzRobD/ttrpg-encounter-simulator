import { DamageType } from './combat.model';
import { DiceType } from './stats.model';

export interface SpecialAbilities {
  assassinate: boolean;
  berserkThreshold: number;
  bloodFrenzy: boolean;
  consumeLifeDie: DiceType;
  corrosiveFormNumDice: number;
  deathBurstNumDice: number;
  deathBurstDamageType: DamageType;
  deathBurstDC: number;
  deathThroesNumDice: number;
  deathThroesDC: number;
  divineEminenceNumDice: number;
  evasion: boolean;
  fireAuraNumDice: number;
  fireForm: boolean;
  gibbering: boolean;
  gnomeCunning: boolean;
  heatedBodyNumDice: number;
  legendaryResistanceCount: number;
  lightningAbsorption: boolean;
  limitedMagicImmunityLevel: number;
  magicResistance: boolean;
  magicWeapons: boolean;
  martialAdvantageNumDice: number;
  packTactics: boolean;
  reckless: boolean;
  reflectiveCarapace: boolean;
  regenerationValue: number;
  relentlessThreshold: number;
  sneakAttackNumDice: number;
  undeadFortitude: boolean;
}
