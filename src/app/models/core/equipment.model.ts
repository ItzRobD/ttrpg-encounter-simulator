import { DamageType } from './combat.model';

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

export interface DamageBlock {
  numberOfDice: number;
  die: number;
  damageType: DamageType;
  modifier: number;
}

export interface Weapon {
  id?: number | string;
  name: string;
  isCustom?: boolean;
  type?: string;
  damageBlocks: DamageBlock[];
  properties: WeaponProperties;
  modifiers: WeaponModifiers;
  isProficient?: boolean;
}

export interface Armor {
  id?: number | string;
  name: string;
  isCustom?: boolean;
  type?: string;
  ac: number;
  dexBonus: boolean;
  maxBonus: boolean;
  minimumStr: number;
  modifier: number;
}

export type EquipmentItem = Weapon | Armor;

export interface WeaponSlotData {
  weaponId: number | string;
  isProficient: boolean;
  modifiers: WeaponModifiers;
}

export interface Equipment {
  armorId?: number | string;
  armor?: Armor;
  shieldId?: number | string;
  shield?: Armor;
  hasShieldEquipped: boolean;
  primarySlot?: WeaponSlotData[];
  secondarySlot?: WeaponSlotData[];
  rangedSlot?: WeaponSlotData[];
}

export interface EquipmentSummary {
  id: number | string;
  name: string;
  type: 'Weapon' | 'Armor' | 'Shield';
  isCustom?: boolean;
  detail?: string;
  properties?: Partial<WeaponProperties>;
}
