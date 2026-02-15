import { Abilities, Metadata, Feature, Spell, InnateSpell, SpellSlots, Ability, DamageResistances, Equipment, DamageType, DiceType, ActorState, Spellcasting } from '../core';
import { BehaviorConfig } from '../configs/behavior-config.model';

export type ActorType = 'character' | 'monster' | string;
export type Side = 'character' | 'monster' | string;

export interface DamageComponent {
  numberOfDice: number;
  die: DiceType;
  amountToAdd: number;
  damageType: DamageType;
}

export interface MultiattackOption {
  actionId: string | number;
  count: number;
}

export interface Action {
  actionId?: string | number;
  name: string;
  description?: string;
  rechargeValue?: number;
  hasDC?: boolean;
  attackBonus: number;
  damageBlocks: DamageComponent[];
  dcAbility?: string;
  dcOnSuccess?: string;
  dc?: number;
  multiattack?: MultiattackOption[];
}

export interface MonsterSpellcastingConfig {
  monsterId: number;
  castingLevel: number;
  ability: Ability;
  attackModifier: number;
  saveDC: number;
  leveledSpells: Spell[];
  innateSpells: InnateSpell[];
  spellSlots: SpellSlots;
}

export interface EquipmentConfig {
  id: string;
  type: 'armor' | 'weapon' | 'shield';
  slot: 'armor' | 'primary' | 'secondary' | 'ranged';
}

export interface HPConfig {
  hpSetMethod: number;
  value: number;
  hpAverage: number;
  numberOfDice: number;
  hitDie?: number;
  hitDice?: Record<number, number>;
  amountToAdd: number;
  modifier?: number;
}

export interface Actor {
  id: number | string;
  instanceId: number;
  name: string;
  isCustom?: boolean;
  state: ActorState;
  spellcasting?: Spellcasting;
  hp?: HPConfig;
  side: Side;
  actorType: ActorType;
  abilities: Abilities;
  hpConfig: HPConfig;
  metadata: Metadata;
  ac?: number;
  equipment?: Equipment;
  equipmentConfigs?: EquipmentConfig[];
  knownSpellIDs?: number[];
  customEquipment?: unknown[];
  customSpells?: Spell[];
  actions?: Action[];
  spellActions?: Action[];
  resistances?: DamageResistances;
  features?: Feature[];
  behavior?: BehaviorConfig;
}

export interface ActorConfig extends Actor {}

export interface ActorSummary {
  id: number | string;
  name: string;
  isCustom?: boolean;
  cr?: number;
  level?: number;
  type?: string;
  class?: string;
  race?: string;
  ac?: number;
  isLegendary?: boolean;
  isSpellcaster?: boolean;
  isInnateCaster?: boolean;
}
