package entity_configuration

type CharacterSpecificFeatures struct {
	Feats CharacterFeats
}

// CharacterFeats represents the set of optional feats that a character may possess in the game.
type CharacterFeats struct {
	TwoWeaponFighting bool // Add damage to offhand
	GreatWeaponMaster bool // Power Attack
	Sharpshooter      bool // Power Attack
	XBowExpert        bool // Bonus action hand crossbow attack
	ShieldMaster      bool // Better dex saves with a shield
	WarCaster         bool // Adv on conc saves
	DualWielder       bool // +1 AC While dual wielding, can use non light weapons
	// Crusher/Slasher/Piercer
	//Defensive duelist - reaction to add ac
	HeavyArmorMaster bool // If heavy armor, non magic phys damage reduced by 3
}

type ClassFeatureSet struct {
	// TODO: Implement other class features ie rage
	RogueFeatures *RogueFeatures
}

// RogueFeatures defines features specific to a rogue character, including sneak attack and assassinate capabilities.
type RogueFeatures struct {
	HasSneakAttack       bool
	HasAssassinate       bool
	NumOfSneakAttackDice int
}
