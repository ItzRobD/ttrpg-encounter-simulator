package shared

type AbilityScore string

const (
	AbilityStrength     AbilityScore = "strength"
	AbilityDexterity    AbilityScore = "dexterity"
	AbilityConstitution AbilityScore = "constitution"
	AbilityIntelligence AbilityScore = "intelligence"
	AbilityWisdom       AbilityScore = "wisdom"
	AbilityCharisma     AbilityScore = "charisma"
)

type AbilityScores struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}
