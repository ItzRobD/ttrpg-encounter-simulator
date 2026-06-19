package core

import "strings"

type ClassID int

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

func (c ClassID) Int() int { return int(c) }

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
