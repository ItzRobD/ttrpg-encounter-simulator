package core

import "strings"

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

type MonsterSize string

const (
	MonsterSizeTiny       MonsterSize = "Tiny"
	MonsterSizeSmall      MonsterSize = "Small"
	MonsterSizeMedium     MonsterSize = "Medium"
	MonsterSizeLarge      MonsterSize = "Large"
	MonsterSizeHuge       MonsterSize = "Huge"
	MonsterSizeGargantuan MonsterSize = "Gargantuan"
	MonsterSizeSwarm      MonsterSize = "Swarm"
	MonsterSizeUnknown    MonsterSize = "Unknown"
)

func MakeMonsterSize(s string) MonsterSize {
	switch strings.ToLower(s) {
	case "tiny":
		return MonsterSizeTiny
	case "small":
		return MonsterSizeSmall
	case "medium":
		return MonsterSizeMedium
	case "large":
		return MonsterSizeLarge
	case "huge":
		return MonsterSizeHuge
	case "gargantuan":
		return MonsterSizeGargantuan
	case "swarm":
		return MonsterSizeSwarm
	default:
		return MonsterSizeUnknown
	}
}
