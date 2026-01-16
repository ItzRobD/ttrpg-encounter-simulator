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
  id?: number | string;
  name: string;
  isCustom?: boolean;
  numberOfDice: number;
  die: DiceType;
  damageType: DamageType;
  properties: WeaponProperties;
  modifiers: WeaponModifiers;
  isProficient?: boolean;
}

export interface Armor {
  id?: number | string;
  name: string;
  isCustom?: boolean;
  ac: number;
  dexBonus: number;
  maxBonus: number;
  minimumStrength: number;
  modifier: number;
}

export type EquipmentItem = Weapon | Armor;

export interface EquipmentSummary {
  id: number | string;
  name: string;
  isCustom?: boolean;
  type: 'Weapon' | 'Armor' | 'Shield';
  detail: string; // e.g., "1d8 Slashing" or "AC 18"
  properties?: {
    isVersatile?: boolean;
    isFinesse?: boolean;
    isHeavy?: boolean;
    isLight?: boolean;
    isTwoHanded?: boolean;
    isThrown?: boolean;
    isRanged?: boolean;
  };
}

export interface Equipment {
  armor?: Armor;
  shield?: Armor;
  hasShieldEquipped: boolean;
  weapons: {
    [slot in WeaponSlot]?: Weapon[];
  };
}
