package core

type Entity interface {
	ModifyHP(value int)
	IsUnconscious() bool
	GetCurrentHP() int
	GetCurrentHPPct() int
	GetMaxHP() int
	GetName() string
	GetAC() int
	GetEventListener() func(event interface{})
	GetLevel() interface{}
	GetCasterLevel() int
	MakeSavingThrow(ability Ability, targetValue int) (RollResult, error)
	GetSpellSaveDC(ability Ability) int
	GetAbilityScores() AbilityScores
	GetAbilityScore(ability Ability) int
	GetAbilityScoreModifier(ability Ability) (int, error)
	GetSavingThrowBonus(ability Ability) (int, error)
	IsCharacter() bool
	IsMonster() bool
}

type Combatant struct {
	InitiativeScore int
	Entity          Entity
}
