package core

type Entity interface {
	IsUnconscious() bool                       // Added C|M
	GetHPStatus() HPStatus                     // Added C|M
	GetName() string                           // Added C|M
	GetAC() int                                // Added C|M
	GetEventListener() func(event interface{}) // Added C|M
	SetEventListener(listener func(event interface{}))
	GetLevel() float64                                                    // Added C|M
	GetHitDie() DiceType                                                  // Added C|M
	GetCasterLevel() int                                                  // Added C|M
	MakeSavingThrow(ability Ability, targetValue int) (RollResult, error) // Added C|M
	GetSpellSaveDC(ability *Ability) (int, error)                         // Added C|M
	GetAbilityScores() AbilityScores                                      // Added C|M
	GetAbilityScore(ability Ability) int                                  // Added C|M
	GetAbilityScoreModifier(ability Ability) (int, error)                 // Added C|M
	GetSavingThrowBonus(ability Ability) (int, error)                     // Added C|M
	IsCharacter() bool                                                    // Added C|M
	IsMonster() bool                                                      // Added C|M
	GetHPConfig() HPConfig
	GetState() interface{}
	RollInitiative() (int, error)
	InitializeHP() error
	IsSpellcaster() bool
	IsHealer() bool
	GetTargetPriority() TargetPriority
	SetTargetPriority(priority TargetPriority)
}

type Combatant struct {
	Entity     Entity
	Initiative int
}

func NewCombatant(entity Entity, initiative int) Combatant {
	return Combatant{entity, initiative}
}

func (c Combatant) GetInitiative() int {
	return c.Initiative
}

func (c Combatant) GetEntity() Entity {
	return c.Entity
}

func (c Combatant) CanAct() bool {
	return true
}
