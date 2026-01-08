export enum Condition {
  Blinded = "blinded",
  Charmed = "charmed",
  Deafened = "deafened",
  Frightened = "frightened",
  Grappled = "grappled",
  Incapacitated = "incapacitated",
  Invisible = "invisible",
  Paralyzed = "paralyzed",
  Petrified = "petrified",
  Poisoned = "poisoned",
  Prone = "prone",
  Restrained = "restrained",
  Stunned = "stunned",
  Unconscious = "unconscious"
}

export type Conditions = {
  [condition in Condition]: boolean;
};
