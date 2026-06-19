package core

import "strings"

type Decision string

const (
	DecisionAttack    Decision = "attack"
	DecisionHeal      Decision = "heal"
	DecisionLegendary Decision = "legendary"
)

type ActionPreference string

const (
	APAttack       ActionPreference = "attack"        // Prefers melee attack, heals if emergency need
	APRanged       ActionPreference = "ranged_attack" // Prefers ranged attack, heals if emergency need
	APHeal         ActionPreference = "heal"          // Prefers healing if allies under threshold
	APSpell        ActionPreference = "spell"         // Prefers casting spells, heals if emergency need
	APSlayer       ActionPreference = "slayer"        // Prefers attack, ignores allies in need of healing
	APSlayerRanged ActionPreference = "slayer_ranged" // Prefers ranged attack, ignores allies in need of healing
	APSlayerSpell  ActionPreference = "slayer_spell"  // Prefers casting spells, ignores allies in need of healing
)

func (ap *ActionPreference) IsSlayer() bool {
	return *ap == APSlayer || *ap == APSlayerSpell || *ap == APSlayerRanged
}

func (ap *ActionPreference) IsHealer() bool {
	return *ap == APHeal
}

type TargetPriority string

const (
	PriorityNone          TargetPriority = "none"
	PriorityLowestHP      TargetPriority = "lowest_hp"
	PriorityMostDamaged   TargetPriority = "most_damaged"
	PriorityLeastDamaged  TargetPriority = "least_damaged"
	PrioritySpellcaster   TargetPriority = "spellcaster"
	PriorityHealer        TargetPriority = "healer"
	PriorityHighestScaler TargetPriority = "highest_scaler"
	PriorityLowestScaler  TargetPriority = "lowest_scaler"
	PriorityHighestMaxHP  TargetPriority = "highest_max_hp"
	PriorityLowestMaxHP   TargetPriority = "lowest_max_hp"
)

func MakeTargetPriority(s string) TargetPriority {
	switch strings.ToLower(s) {
	case "lowest_hp", "lowest hp":
		return PriorityLowestHP
	case "most_damaged", "most damaged":
		return PriorityMostDamaged
	case "spellcaster":
		return PrioritySpellcaster
	case "healer":
		return PriorityHealer
	case "highest_scaler", "highest scaler":
		return PriorityHighestScaler
	case "lowest_scaler", "lowest scaler":
		return PriorityLowestScaler
	default:
		return PriorityNone
	}
}

type SpellPriority string

const (
	SPNoPreference       SpellPriority = "no preference"
	SPHighestLevel       SpellPriority = "highest level"
	SPLowestLevel        SpellPriority = "lowest level"
	SPCantrip            SpellPriority = "highest cantrip"
	SPRandomCantrip      SpellPriority = "random cantrip"
	SPRandomLeveledSpell SpellPriority = "random leveled spell"
	SPAreaOfEffect       SpellPriority = "area of effect"
	SPHighestDamage      SpellPriority = "highest damage"
)

func MakeSpellPriority(s string) SpellPriority {
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
