package core

import "strings"

type RaceID int

const (
	_ RaceID = iota
	Dwarf
	Elf
	Halfling
	Human
	Dragonborn
	Gnome
	HalfElf
	HalfOrc
	Tiefling
)

func (r RaceID) Int() int { return int(r) }

type DragonbornColor string

const (
	DragonbornBlue   DragonbornColor = "Blue"
	DragonbornBlack  DragonbornColor = "Black"
	DragonbornBrass  DragonbornColor = "Brass"
	DragonbornBronze DragonbornColor = "Bronze"
	DragonbornCopper DragonbornColor = "Copper"
	DragonbornGold   DragonbornColor = "Gold"
	DragonbornGreen  DragonbornColor = "Green"
	DragonbornRed    DragonbornColor = "Red"
	DragonbornSilver DragonbornColor = "Silver"
	DragonbornWhite  DragonbornColor = "White"
)

func (r RaceID) String() string {
	switch r {
	case Dwarf:
		return "Dwarf"
	case Elf:
		return "Elf"
	case Halfling:
		return "Halfling"
	case Human:
		return "Human"
	case Dragonborn:
		return "Dragonborn"
	case Gnome:
		return "Gnome"
	case HalfElf:
		return "Half-Elf"
	case HalfOrc:
		return "Half-Orc"
	case Tiefling:
		return "Tiefling"
	default:
		return "unknown race"
	}
}

func MakeRaceID(s string) RaceID {
	switch strings.ToLower(s) {
	case "dwarf":
		return Dwarf
	case "elf":
		return Elf
	case "halfling":
		return Halfling
	case "human":
		return Human
	case "dragonborn":
		return Dragonborn
	case "gnome":
		return Gnome
	case "half-elf":
		return HalfElf
	case "halfelf":
		return HalfElf
	case "half-orc":
		return HalfOrc
	case "halforc":
		return HalfOrc
	case "tiefling":
		return Tiefling
	default:
		return 0
	}
}
