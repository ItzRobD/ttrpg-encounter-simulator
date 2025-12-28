package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"strings"
)

type MonsterConfig struct {
	Base               MonsterBase
	Actions            map[int]Action
	Multiattacks       map[int][]Multiattack
	LegendaryActions   map[int]LegendaryAction
	SpecialAbilities   []SpecialAbility
	Resistances        core.DamageResistances
	DamageBreakers     []core.ResistBreaker
	spellcastingConfig MonsterSpellcastingConfig
	HPSetMethod        core.HPSetMethod
	Seed               core.Seed
}

type MonsterSpellcastingConfig struct {
	MonsterID      int
	CastingLevel   int
	Ability        core.Ability
	AttackModifier int
	SaveDC         int
	LeveledSpells  []spells.Spell
	InnateSpells   []spells.InnateSpell
	SpellSlots     spells.SpellSlots
}

type MonsterType string

const (
	MTAberration  = "Aberration"
	MTBeast       = "Beast"
	MTCelestial   = "Celestial"
	MTConstruct   = "Construct"
	MTDragon      = "Dragon"
	MTElemental   = "Elemental"
	MTFey         = "Fey"
	MTFiend       = "Fiend"
	MTGiant       = "Giant"
	MTHumanoid    = "Humanoid"
	MTMonstrosity = "Monstrosity"
	MTOoze        = "Ooze"
	MTPlant       = "Plant"
	MTUndead      = "Undead"
	MTUnknown     = "Unknown"
)

func (mt MonsterType) String() string {
	return string(mt)
}

func MakeMonsterType(s string) MonsterType {
	switch strings.ToLower(s) {
	case "aberration":
		return MTAberration
	case "beast":
		return MTBeast
	case "celestial":
		return MTCelestial
	case "construct":
		return MTConstruct
	case "dragon":
		return MTDragon
	case "elemental":
		return MTElemental
	case "fey":
		return MTFey
	case "fiend":
		return MTFiend
	case "giant":
		return MTGiant
	case "humanoid":
		return MTHumanoid
	case "monstrosity":
		return MTMonstrosity
	case "ooze":
		return MTOoze
	case "plant":
		return MTPlant
	case "undead":
		return MTUndead
	default:
		return MTUnknown
	}
}
