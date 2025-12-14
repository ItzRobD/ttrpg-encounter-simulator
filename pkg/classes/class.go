package classes

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"strings"
)

type ClassID uint8

const (
	_ ClassID = iota
	Artificer
	Barbarian
	Bard
	Cleric
	Druid
	Fighter
	Monk
	Paladin
	Ranger
	Rogue
	Sorcerer
	Warlock
	Wizard
)

func (c ClassID) Int() uint8 { return uint8(c) }

func (c ClassID) String() string {
	switch c {
	case Artificer:
		return "Artificer"
	case Barbarian:
		return "Barbarian"
	case Bard:
		return "Bard"
	case Cleric:
		return "Cleric"
	case Druid:
		return "Druid"
	case Fighter:
		return "Fighter"
	case Monk:
		return "Monk"
	case Paladin:
		return "Paladin"
	case Ranger:
		return "Ranger"
	case Rogue:
		return "Rogue"
	case Sorcerer:
		return "Sorcerer"
	case Warlock:
		return "Warlock"
	case Wizard:
		return "Wizard"
	default:
		return "unknown class"
	}
}

func MakeClassID(s string) ClassID {
	switch strings.ToLower(s) {
	case "artificer":
		return Artificer
	case "barbarian":
		return Barbarian
	case "bard":
		return Bard
	case "cleric":
		return Cleric
	case "druid":
		return Druid
	case "fighter":
		return Fighter
	case "monk":
		return Monk
	case "paladin":
		return Paladin
	case "ranger":
		return Ranger
	case "rogue":
		return Rogue
	case "sorcerer":
		return Sorcerer
	case "warlock":
		return Warlock
	case "wizard":
		return Wizard
	default:
		return 0
	}
}

// Class represents a character class with its attributes like ID, name, hit die, and spellcasting modifier.
type Class struct {
	ID                   ClassID
	Name                 string
	HitDie               core.DiceType
	SpellcastingMod      core.Ability
	SneakAttackDiceCount int
	AttackCount          int
}

// ClassQueryParams defines parameters for querying a class, including its name and ID.
type ClassQueryParams struct {
	Name  string
	ID    ClassID
	Level uint8
}
