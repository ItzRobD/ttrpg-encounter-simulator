package core

import (
	"math"
	"strings"
)

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

func (as *AbilityScores) Get(ability Ability) int {
	switch ability {
	case AbilityStrength:
		return as.Strength
	case AbilityDexterity:
		return as.Dexterity
	case AbilityConstitution:
		return as.Constitution
	case AbilityIntelligence:
		return as.Intelligence
	case AbilityWisdom:
		return as.Wisdom
	case AbilityCharisma:
		return as.Charisma
	default:
		return 0
	}
}

func NewAbilityScores(strength, dexterity, constitution, intelligence, wisdom, charisma int) AbilityScores {
	return AbilityScores{
		Strength:     strength,
		Dexterity:    dexterity,
		Constitution: constitution,
		Intelligence: intelligence,
		Wisdom:       wisdom,
		Charisma:     charisma,
	}
}

type AbilityScoresProficiencies struct {
	Strength     bool `json:"strength"`
	Dexterity    bool `json:"dexterity"`
	Constitution bool `json:"constitution"`
	Intelligence bool `json:"intelligence"`
	Wisdom       bool `json:"wisdom"`
	Charisma     bool `json:"charisma"`
}

func (as *AbilityScoresProficiencies) Get(ability Ability) bool {
	switch ability {
	case AbilityStrength:
		return as.Strength
	case AbilityDexterity:
		return as.Dexterity
	case AbilityConstitution:
		return as.Constitution
	case AbilityIntelligence:
		return as.Intelligence
	case AbilityWisdom:
		return as.Wisdom
	case AbilityCharisma:
		return as.Charisma
	default:
		return false
	}
}

type SaveProficiencies struct {
	Strength     bool `json:"strength"`
	Dexterity    bool `json:"dexterity"`
	Constitution bool `json:"constitution"`
	Intelligence bool `json:"intelligence"`
	Wisdom       bool `json:"wisdom"`
	Charisma     bool `json:"charisma"`
}

func (as *SaveProficiencies) Get(ability Ability) bool {
	switch ability {
	case AbilityStrength:
		return as.Strength
	case AbilityDexterity:
		return as.Dexterity
	case AbilityConstitution:
		return as.Constitution
	case AbilityIntelligence:
		return as.Intelligence
	case AbilityWisdom:
		return as.Wisdom
	case AbilityCharisma:
		return as.Charisma
	default:
		return false
	}
}

func NewAbilityScoresProficiencies(strength, dexterity, constitution, intelligence, wisdom, charisma bool) AbilityScoresProficiencies {
	return AbilityScoresProficiencies{
		Strength:     strength,
		Dexterity:    dexterity,
		Constitution: constitution,
		Intelligence: intelligence,
		Wisdom:       wisdom,
		Charisma:     charisma,
	}
}

type Abilities struct {
	AbilityScores AbilityScores              `json:"ability_scores"`
	Proficiencies AbilityScoresProficiencies `json:"proficiencies"`
}

func (a *Abilities) GetScore(ability Ability) int {
	return a.AbilityScores.Get(ability)
}

func (a *Abilities) GetIsProficientInAbility(ability Ability) bool {
	return a.Proficiencies.Get(ability)
}

func (a *Abilities) GetAbilityModifier(ability Ability) int {
	score := a.AbilityScores.Get(ability)
	return int(math.Floor(float64(score-10) / 2.0))
}
