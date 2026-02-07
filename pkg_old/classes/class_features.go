package classes

// Class feature level requirements
const (
	BarbarianRageLevel              = 2
	BarbarianRecklessAttackLevel    = 2
	BarbarianDangerSenseLevel       = 2
	BarbarianFeralInstinctLevel     = 7
	BarbarianBrutalCriticalLevel    = 9
	BarbarianRelentlessRageLevel    = 11
	FighterSecondWindLevel          = 1
	FighterIndomitableLevel         = 9
	MonkDeflectMissilesLevel        = 3
	MonkEvasionLevel                = 7
	PaladinDivineSmiteLevel         = 2
	PaladinImprovedDivineSmiteLevel = 11
	PaladinLayOnHandsPoolModifier   = 5
	RogueUncannyDodgeLevel          = 5
	RogueEvasionLevel               = 7
	RogueSlipperyMindLevel          = 15
	RogueElusiveLevel               = 18
)

type ClassFeatures struct {
	BarbarianFeatures *BarbarianFeatures
	FighterFeatures   *FighterFeatures
	MonkFeatures      *MonkFeatures
	PaladinFeatures   *PaladinFeatures
	RogueFeatures     *RogueFeatures
}

type BarbarianFeatures struct {
	HasRage                bool // Adv on str saves, apply extra damage, apply physical resistance
	HasRecklessAttack      bool // Outgoing advantage, incoming advantage
	HasDangerSense         bool // Adv on dex saves
	HasFeralInstinct       bool // Adv on initiative
	HasBrutalCritical      bool // extra damage dice on crit
	HasRelentlessRage      bool // When dropping to 0hp, DC (10 * times used) con save -> 1 hp, +5 dc each time
	RageDamage             int
	NumberOfBrutalCritDice int
}

type FighterFeatures struct {
	HasSecondWind   bool // regain 1d10 + level hp
	HasIndomitable  bool // Reroll {uses} number of saves
	IndomitableUses int
}

type MonkFeatures struct {
	HasDeflectMissiles bool // Reduce ranged damage by 1d10 + dex mod + level
	HasEvasion         bool // dex save = no damage
}

type PaladinFeatures struct {
	LayOnHandsPool         int  // 5 * Paladin level
	HasDivineSmite         bool // (2 * slot level)d8 radiant damage + 1d8 if fiend/undead
	HasImprovedDivineSmite bool // extra 1d8 radiant damage to every attack
}

// RogueFeatures defines features specific to a rogue character, including sneak attack and assassinate capabilities.
type RogueFeatures struct {
	HasSneakAttack       bool // Apply extra damage if advantage or (no disadvantage && ally within 5ft)
	NumOfSneakAttackDice int
	HasUncannyDodge      bool // reaction to halve damage
	HasEvasion           bool // dex save = no damage
	HasSlipperyMind      bool // Adv on wis saves
	HasElusive           bool // no incoming advantage while not incapacitated
}

// SetupFeatures initializes and updates class-specific features for a given class id and level.
func (f *ClassFeatures) SetupFeatures(classID ClassID, level uint8) {
	switch classID {
	case Barbarian:
		if f.BarbarianFeatures == nil {
			f.BarbarianFeatures = &BarbarianFeatures{}
		}
		f.BarbarianFeatures.HasRage = level >= BarbarianRageLevel
		f.BarbarianFeatures.HasDangerSense = level >= BarbarianDangerSenseLevel
		f.BarbarianFeatures.HasFeralInstinct = level >= BarbarianFeralInstinctLevel
		f.BarbarianFeatures.HasBrutalCritical = level >= BarbarianBrutalCriticalLevel
		f.BarbarianFeatures.HasRelentlessRage = level >= BarbarianRelentlessRageLevel
		f.BarbarianFeatures.HasRecklessAttack = level >= BarbarianRecklessAttackLevel
	case Fighter:
		if f.FighterFeatures == nil {
			f.FighterFeatures = &FighterFeatures{}
		}
		f.FighterFeatures.HasSecondWind = level >= FighterSecondWindLevel
		f.FighterFeatures.HasIndomitable = level >= FighterIndomitableLevel
	case Monk:
		if f.MonkFeatures == nil {
			f.MonkFeatures = &MonkFeatures{}
		}
		f.MonkFeatures.HasDeflectMissiles = level >= MonkDeflectMissilesLevel
		f.MonkFeatures.HasEvasion = level >= MonkEvasionLevel
	case Paladin:
		if f.PaladinFeatures == nil {
			f.PaladinFeatures = &PaladinFeatures{}
		}
		f.PaladinFeatures.LayOnHandsPool = PaladinLayOnHandsPoolModifier * int(level)
		f.PaladinFeatures.HasDivineSmite = level >= PaladinDivineSmiteLevel
		f.PaladinFeatures.HasImprovedDivineSmite = level >= PaladinImprovedDivineSmiteLevel
	case Rogue:
		if f.RogueFeatures == nil {
			f.RogueFeatures = &RogueFeatures{}
		}
		f.RogueFeatures.HasElusive = level >= RogueElusiveLevel
		f.RogueFeatures.HasEvasion = level >= RogueEvasionLevel
		f.RogueFeatures.HasSlipperyMind = level >= RogueSlipperyMindLevel
		f.RogueFeatures.HasUncannyDodge = level >= RogueUncannyDodgeLevel
	default:
		break
	}
}
