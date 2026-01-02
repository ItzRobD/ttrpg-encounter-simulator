package core

import "strings"

type SpellPriority int

const (
	SPNoPreference SpellPriority = iota
	SPHighestLevel
	SPLowestLevel
	SPCantrip // Prioritizes highest value
	SPRandomCantrip
	SPRandomLeveledSpell
	SPAreaOfEffect // Prioritizes highest value
	SPHighestDamage
)

func (sp SpellPriority) String() string {
	switch sp {
	case SPNoPreference:
		return "no preference"
	case SPHighestLevel:
		return "highest level"
	case SPLowestLevel:
		return "lowest level"
	case SPCantrip:
		return "highest cantrip"
	case SPRandomCantrip:
		return "random cantrip"
	case SPRandomLeveledSpell:
		return "random leveled spell"
	case SPAreaOfEffect:
		return "area of effect"
	case SPHighestDamage:
		return "highest damage"
	default:
		return "unknown spell priority"
	}
}

func NewSpellPriority(s string) SpellPriority {
	switch s {
	case "no preference":
		return SPNoPreference
	case "highest level":
		return SPHighestLevel
	case "lowest level":
		return SPLowestLevel
	case "highest cantrip":
		return SPCantrip
	case "random cantrip":
		return SPRandomCantrip
	case "random leveled spell":
		return SPRandomLeveledSpell
	case "area of effect":
		return SPAreaOfEffect
	case "highest damage":
		return SPHighestDamage
	default:
		return SPNoPreference
	}
}

// TargetPriority defines prioritization logic for targeting entities in combat simulations.
// Healer|Spellcaster secondarily prioritizes most damaged (effective healing)
type TargetPriority int

const (
	NoPriority TargetPriority = iota
	PrioritizeLowestHealth
	PrioritizeMostDamaged
	PrioritizeLeastDamaged
	PrioritizeSpellcaster
	PrioritizeHealer
	PrioritizeHighestLevel
	PrioritizeLowestLevel
	PrioritizeHighestMaxHP
	PrioritizeLowestMaxHP
	PrioritizeUnconscious
)

func (p TargetPriority) String() string {
	switch p {
	case NoPriority:
		return "no priority"
	case PrioritizeLowestHealth:
		return "lowest health"
	case PrioritizeMostDamaged:
		return "most damaged"
	case PrioritizeSpellcaster:
		return "spellcaster"
	case PrioritizeHealer:
		return "healer"
	case PrioritizeHighestLevel:
		return "highest CR"
	case PrioritizeLowestLevel:
		return "lowest CR"
	case PrioritizeHighestMaxHP:
		return "highest max HP"
	case PrioritizeLowestMaxHP:
		return "lowest max HP"
	default:
		return "unknown prioritization"
	}
}

func NewPrioritization(s string) TargetPriority {
	switch strings.ToLower(s) {
	case "no priority":
		return NoPriority
	case "lowest health":
		return PrioritizeLowestHealth
	case "most damaged":
		return PrioritizeMostDamaged
	case "spellcaster":
		return PrioritizeSpellcaster
	case "healer":
		return PrioritizeHealer
	case "highest CR":
		return PrioritizeHighestLevel
	case "lowest CR":
		return PrioritizeLowestLevel
	case "highest max HP":
		return PrioritizeHighestMaxHP
	case "lowest max HP":
		return PrioritizeLowestMaxHP
	default:
		return NoPriority
	}
}

type ActionPreference int

const (
	APNoPreference ActionPreference = iota
	APPreferMelee
	APPreferRanged
	APPreferSpells
)

func (p ActionPreference) String() string {
	switch p {
	case APNoPreference:
		return "no preference"
	case APPreferMelee:
		return "prefer melee"
	case APPreferRanged:
		return "prefer ranged"
	case APPreferSpells:
		return "prefer spells"
	default:
		return "unknown action preference"
	}
}

func NewActionPreference(s string) ActionPreference {
	switch strings.ToLower(s) {
	case "no preference":
		return APNoPreference
	case "prefer melee":
		return APPreferMelee
	case "prefer ranged":
		return APPreferRanged
	case "prefer spells":
		return APPreferSpells
	default:
		return APNoPreference
	}
}

type ActionType string

const (
	ATNoAction                 ActionType = "no action"
	ATDamage                              = "damage"
	ATMelee                               = "melee attack"
	ATRanged                              = "ranged attack"
	ATSpell                               = "spell attack"
	ATHeal                                = "healing"
	ATUnarmed                             = "unarmed attack"
	ATMonsterHeal                         = "monster heal"
	ATMonsterDamage                       = "monster damage"
	ATMonsterAction                       = "monster action"
	ATMonsterMultiattack                  = "monster multiattack"
	ATLegendaryAction                     = "legendary action"
	ATMonsterSpecial                      = "monster special ability"
	ATLairAction                          = "lair action"
	ATOffhand                             = "offhand attack"
	ATDragonbornBreathWeapon              = "dragonborn breath weapon"
	ATMonsterDeathEffect                  = "monster death effect"
	ATMonsterRetaliatoryEffect            = "monster retaliatory effect"
)

func (a ActionType) String() string {
	return string(a)
}

func NewActionType(s string) ActionType {
	switch strings.ToLower(s) {
	case "no action":
		return ATNoAction
	case "melee attack":
		return ATMelee
	case "ranged attack":
		return ATRanged
	case "spell attack":
		return ATSpell
	case "healing":
		return ATHeal
	case "unarmed attack":
		return ATUnarmed
	case "lair action":
		return ATLairAction
	default:
		return ATNoAction
	}
}

type VersatileWeaponPreference int

const (
	VWPNoPreference VersatileWeaponPreference = iota
	VWPPreferVersatile
	VWPPreferNonVersatile
)

func (p VersatileWeaponPreference) String() string {
	switch p {
	case VWPNoPreference:
		return "no preference"
	case VWPPreferVersatile:
		return "prefer versatile"
	case VWPPreferNonVersatile:
		return "prefer non-versatile"
	default:
		return "unknown melee preference"
	}
}

func NewMeleePreference(s string) VersatileWeaponPreference {
	switch strings.ToLower(s) {
	case "no preference":
		return VWPNoPreference
	case "prefer versatile":
		return VWPPreferVersatile
	case "prefer non-versatile":
		return VWPPreferNonVersatile
	default:
		return VWPNoPreference
	}
}

func GetActionFromPreference(pref ActionPreference) ActionType {
	switch pref {
	case APPreferMelee:
		return ATMelee
	case APPreferRanged:
		return ATRanged
	case APPreferSpells:
		return ATSpell
	default:
		return ATNoAction
	}
}

type WeaponSlot string

const (
	WSPrimary   WeaponSlot = "primary"
	WSSecondary WeaponSlot = "secondary"
	WSRanged    WeaponSlot = "ranged"
)

func (ws WeaponSlot) String() string {
	return string(ws)
}

func NewWeaponSlot(s string) WeaponSlot {
	switch strings.ToLower(s) {
	case "primary":
		return WSPrimary
	case "secondary":
		return WSSecondary
	case "ranged":
		return WSRanged
	default:
		return WSPrimary
	}
}
