package core

import (
	"math/rand/v2"
)

type Entity interface {
	IsUnconscious() bool
	GetHPStatus() HPStatus
	GetName() string
	GetAC() int
	GetEventListener() func(event interface{})
	SetEventListener(listener func(event interface{}))
	GetLevel() float64
	GetHitDie() DiceType
	GetCasterLevel() int
	MakeSavingThrow(ability Ability, targetValue int) (RollResult, error)
	GetSpellSaveDC(ability *Ability) (int, error)
	GetAbilityScores() AbilityScores
	GetAbilityScore(ability Ability) int
	GetAbilityScoreModifier(ability Ability) (int, error)
	GetSavingThrowBonus(ability Ability) (int, error)
	IsCharacter() bool
	IsMonster() bool
	GetIsLegendary() bool
	GetHPConfig() HPConfig
	GetState() interface{}
	RollInitiative() (int, error)
	InitializeHP() error
	IsSpellcaster() bool
	IsHealer() bool
	GetTargetPriority() TargetPriority
	SetTargetPriority(priority TargetPriority)
	ChooseSpellByHealingEfficiency(targetValue int) (*SpellChoice, error)
	ChooseDamageSpellByPriority(p SpellPriority) (*SpellChoice, error)
	GetHealingSpellCount() int
	GetDamageSpellCount() int
	GetRNG() *rand.Rand
	GetAIRequest(actorID int, t AIRequestType) (*AIRequest, error)
	ExecuteAIRequest(req *AIRequest) (*ActionOutcome, error)
	UpdateAICombatContext(ctx *CombatContext) error
	ModifyHP(value int, isTemp bool, tempStacking bool) (HPModificationResult, error)
	RefreshLegendaryActions()
}

type Combatant struct {
	Entity     Entity
	Initiative int
	CanAct     bool
}

// NewCombatant creates a new Combatant with the specified Entity and initiative, defaulting CanAct to true.
func NewCombatant(entity Entity, initiative int) *Combatant {
	return &Combatant{entity, initiative, true}
}

func (c *Combatant) GetInitiative() int {
	return c.Initiative
}

func (c *Combatant) GetEntity() Entity {
	return c.Entity
}

func (c *Combatant) GetCanAct() bool {
	return c.CanAct
}

func (c *Combatant) SetCanAct(b bool) { c.CanAct = b }
