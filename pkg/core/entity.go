package core

type Entity interface {
	IsUnconscious() bool                                                  // Added C|M
	GetHPStatus() HPStatus                                                // Added C|M
	GetName() string                                                      // Added C|M
	GetAC() int                                                           // Added C|M
	GetEventListener() func(event interface{})                            // Added C|M
	GetLevel() interface{}                                                // Added C|M
	GetHitDie() DiceType                                                  // Added C|M
	GetCasterLevel() int                                                  // Added C|M
	MakeSavingThrow(ability Ability, targetValue int) (RollResult, error) // Added C
	GetSpellSaveDC(ability *Ability) (int, error)                         // Added C|M
	GetAbilityScores() AbilityScores                                      // Added C|M
	GetAbilityScore(ability Ability) int                                  // Added C|M
	GetAbilityScoreModifier(ability Ability) (int, error)                 // Added C|M
	GetSavingThrowBonus(ability Ability) (int, error)                     // Added C|M
	IsCharacter() bool                                                    // Added C|M
	IsMonster() bool                                                      // Added C|M
}

type Combatant struct {
	InitiativeScore int
	Entity          Entity
}
