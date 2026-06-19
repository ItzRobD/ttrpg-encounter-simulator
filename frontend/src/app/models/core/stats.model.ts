export enum DiceType {
  D0 = 0,
  D4 = 4,
  D6 = 6,
  D8 = 8,
  D10 = 10,
  D12 = 12,
  D20 = 20,
  D100 = 100
}

export interface AbilityScores {
  strength: number;
  dexterity: number;
  constitution: number;
  intelligence: number;
  wisdom: number;
  charisma: number;
}

export interface AbilityScoreProficiency {
  strength: boolean;
  dexterity: boolean;
  constitution: boolean;
  intelligence: boolean;
  wisdom: boolean;
  charisma: boolean;
  [key: string]: boolean;
}

export interface AsConfig {
  abilityScores: AbilityScores;
  proficiencies: AbilityScoreProficiency;
}

export interface Abilities {
  abilityScores: AbilityScores;
  proficiencies: AbilityScoreProficiency;
}

export interface DeathSaves {
  successes: number;
  failures: number;
}
