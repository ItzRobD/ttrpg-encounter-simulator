import { DamageType } from './combat.model';
import { DiceType } from './stats.model';

export enum WeaponSlot {
  Primary = "Primary",
  Secondary = "Secondary",
  Ranged = "Ranged"
}

export interface WeaponProperties {
  isVersatile: boolean;
  isFinesse: boolean;
  isRanged: boolean;
  isHeavy: boolean;
  isTwoHanded: boolean;
  isLight: boolean;
  isThrown: boolean;
  isOnlyRanged: boolean;
}

export interface WeaponModifiers {
  isMagic: boolean;
  isSilvered: boolean;
  isAdamantine: boolean;
  isColdForgedIron: boolean;
  attackBonus: number;
  damageBonus: number;
}

export interface Weapon {
  id?: number;
  name: string;
  numberOfDice: number;
  die: DiceType;
  damageType: DamageType;
  properties: WeaponProperties;
  modifiers: WeaponModifiers;
  isProficient?: boolean;
}

export interface Armor {
  id: number;
  name: string;
  ac: number;
  minimumStrength: number;
}

export interface Equipment {
  armor?: Armor;
  shield?: Armor;
  hasShieldEquipped: boolean;
  weapons: {
    [slot in WeaponSlot]?: Weapon[];
  };
}
