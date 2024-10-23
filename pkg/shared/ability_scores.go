package shared

const (
	AbilityStrength     = "strength"
	AbilityDexterity    = "dexterity"
	AbilityConstitution = "constitution"
	AbilityIntelligence = "intelligence"
	AbilityWisdom       = "wisdom"
	AbilityCharisma     = "charisma"
)

type AbilityScores struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}
