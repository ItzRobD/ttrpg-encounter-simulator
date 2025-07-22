package core

type Entity interface {
	IsUnconscious() bool                                                  // Added C
	GetHPStatus() HPStatus                                                // Added C TODO: Fix type
	GetName() string                                                      // Added C
	GetAC() int                                                           // Added C
	GetEventListener() func(event interface{})                            // Added C
	GetLevel() interface{}                                                // Added C
	GetHitDie() DiceType                                                  // Added C
	GetCasterLevel() uint8                                                // Added C
	MakeSavingThrow(ability Ability, targetValue int) (RollResult, error) // Added C
	GetSpellSaveDC(ability Ability) int                                   // Added C
	GetAbilityScores() AbilityScores                                      // Added C
	GetAbilityScore(ability Ability) int                                  // Added C
	GetAbilityScoreModifier(ability Ability) (int, error)                 // Added C
	GetSavingThrowBonus(ability Ability) (int, error)                     // Added C
	IsCharacter() bool                                                    // Added C
	IsMonster() bool                                                      // Added C
}

type Combatant struct {
	InitiativeScore int
	Entity          Entity
}
