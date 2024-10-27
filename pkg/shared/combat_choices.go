package shared

type SpellPriority int

const (
	SPNoPreference SpellPriority = iota
	SPHighestLevel
	SPLowestLevel
	SPCantrip
	SPAreaOfEffect
)

type Prioritization int

const (
	NoPriority Prioritization = iota
	PrioritizeLowestHealth
	PrioritizeMostDamaged
	PrioritizeSpellcasting
	PrioritizeHealer
	PrioritizeHighestCR
	PrioritizeLowestCR
	PrioritizeHighestMaxHP
	PrioritizeLowestMaxHP
)

type ActionPreference int

const (
	APNoPreference ActionPreference = iota
	APPreferMelee
	APPreferRanged
	APPreferSpells
)

type ActionType string

const (
	ATNoAction ActionType = "no action"
	ATMelee               = "melee attack"
	ATRanged              = "ranged attack"
	ATSpell               = "spell attack"
	ATHeal                = "healing"
)

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
