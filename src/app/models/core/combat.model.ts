import { Conditions } from './conditions.model';
import { DeathSaves } from './stats.model';

export enum DamageType {
  Acid = "acid",
  Bludgeoning = "bludgeoning",
  Cold = "cold",
  Fire = "fire",
  Force = "force",
  Lightning = "lightning",
  Necrotic = "necrotic",
  Piercing = "piercing",
  Poison = "poison",
  Psychic = "psychic",
  Radiant = "radiant",
  Slashing = "slashing",
  Thunder = "thunder"
}

export enum ResistanceType {
  None = "none",
  Resistant = "resistant",
  Vulnerable = "vulnerable",
  Immune = "immune"
}

export type DamageResistances = {
  [dt in DamageType]: ResistanceType;
};

export interface EntityState {
  currentHP: number;
  maxHP: number;
  tempHP: number;
  hitDie: number;
  conditions: Conditions;
  deathSaves: DeathSaves;
  resistances: DamageResistances;
  isStable: boolean;
  isDead: boolean;
  initiative: number;
}
