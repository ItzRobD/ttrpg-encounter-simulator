package core

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

type AbilityScores struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}
