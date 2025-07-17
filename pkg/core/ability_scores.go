package core

import "strings"

type Ability string

const (
	AbilityStrength     Ability = "strength"
	AbilityDexterity    Ability = "dexterity"
	AbilityConstitution Ability = "constitution"
	AbilityIntelligence Ability = "intelligence"
	AbilityWisdom       Ability = "wisdom"
	AbilityCharisma     Ability = "charisma"
	AbilityNone         Ability = ""
)

func (a Ability) String() string {
	return string(a)
}

func MakeAbility(s string) Ability {
	switch strings.ToLower(s) {
	case "str", "strength":
		return AbilityStrength
	case "dex", "dexterity":
		return AbilityDexterity
	case "con", "constitution":
		return AbilityConstitution
	case "int", "intelligence":
		return AbilityIntelligence
	case "wis", "wisdom":
		return AbilityWisdom
	case "cha", "charisma":
		return AbilityCharisma
	default:
		return AbilityNone
	}
}

type AbilityScores struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

type AbilityScoresProficiencies struct {
	Strength     bool `json:"strength"`
	Dexterity    bool `json:"dexterity"`
	Constitution bool `json:"constitution"`
	Intelligence bool `json:"intelligence"`
	Wisdom       bool `json:"wisdom"`
	Charisma     bool `json:"charisma"`
}

type SaveProficiencies struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}
