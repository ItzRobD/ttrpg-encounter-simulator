package races

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"strings"
)

type RaceID uint8

const (
	_ RaceID = iota
	Dwarf
	Halfling
	Human
	Dragonborn
	Gnome
	HalfElf
	HalfOrc
	Tiefling
)

func (r RaceID) Int() uint8 { return uint8(r) }

func (r RaceID) String() string {
	switch r {
	case Dwarf:
		return "Dwarf"
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

type Race struct {
	ID                 RaceID
	Name               string
	DragonbornFeatures *DragonbornFeatures
	Resistances        core.DamageResistances
	SavingThrowAdv     RacialSavingThrowAdvantage
}

type DragonbornFeatures struct {
	AncestryColor DragonbornColor
	DamageType    core.DamageType
	NumberOfDice  int
	Die           core.DiceType
}

type RaceQueryParams struct {
	Name            string
	ID              RaceID
	Level           uint8
	DragonbornColor *DragonbornColor
}

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

type RacialSavingThrowAdvantage struct {
	Abilities                  map[core.Ability]core.AdvantageType
	DamageTypes                map[core.DamageType]core.AdvantageType
	AdvantageOnlyAgainstSpells bool
}

func NewRacialSavingThrowAdvantage() RacialSavingThrowAdvantage {
	return RacialSavingThrowAdvantage{
		Abilities: map[core.Ability]core.AdvantageType{
			core.AbilityStrength:     core.RollNormal,
			core.AbilityDexterity:    core.RollNormal,
			core.AbilityConstitution: core.RollNormal,
			core.AbilityIntelligence: core.RollNormal,
			core.AbilityWisdom:       core.RollNormal,
			core.AbilityCharisma:     core.RollNormal,
		},
		DamageTypes: map[core.DamageType]core.AdvantageType{
			core.DamageAcid:        core.RollNormal,
			core.DamageCold:        core.RollNormal,
			core.DamageFire:        core.RollNormal,
			core.DamageForce:       core.RollNormal,
			core.DamageLightning:   core.RollNormal,
			core.DamageNecrotic:    core.RollNormal,
			core.DamagePoison:      core.RollNormal,
			core.DamagePsychic:     core.RollNormal,
			core.DamageRadiant:     core.RollNormal,
			core.DamageThunder:     core.RollNormal,
			core.DamageSlashing:    core.RollNormal,
			core.DamageBludgeoning: core.RollNormal,
			core.DamagePiercing:    core.RollNormal,
		},
		AdvantageOnlyAgainstSpells: false,
	}
}

func (rst *RacialSavingThrowAdvantage) SetAdvantageAbility(ability core.Ability, advantageType core.AdvantageType) {
	if rst.Abilities == nil {
		rst.Abilities = map[core.Ability]core.AdvantageType{}
	}
	rst.Abilities[ability] = advantageType
}

func (rst RacialSavingThrowAdvantage) GetAdvantageAbility(ability core.Ability) core.AdvantageType {
	return rst.Abilities[ability]
}

func (rst *RacialSavingThrowAdvantage) SetAdvantageDamageType(dt core.DamageType, advantageType core.AdvantageType) {
	if rst.DamageTypes == nil {
		rst.DamageTypes = map[core.DamageType]core.AdvantageType{}
	}
	rst.DamageTypes[dt] = advantageType
}

func (rst RacialSavingThrowAdvantage) GetAdvantageDamageType(dt core.DamageType) core.AdvantageType {
	return rst.DamageTypes[dt]
}

func (rst *RacialSavingThrowAdvantage) SetAdvantageOnlyAgainstSpells(advantage bool) {
	rst.AdvantageOnlyAgainstSpells = advantage
}
