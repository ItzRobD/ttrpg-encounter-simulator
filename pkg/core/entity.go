package core

import (
	"math/rand/v2"
)

type Entity interface {
	GetClassID() uint8 // Monsters == 0, Characters >= 1
	IsDead() bool
	IsUnconscious() bool
	GetHPStatus() HPStatus
	GetName() string
	GetAC() int
	GetEventListener() func(event interface{})
	SetEventListener(listener func(event interface{}))
	GetLevel() float64
	GetHitDie() DiceType
	GetCasterLevel() int
	MakeSavingThrow(ability Ability, targetValue int, isSpell bool, damageType DamageType) (RollResult, error)
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
	ModifyHP(value int, isTemp bool, tempStacking bool, allowMassiveDamage bool) (HPModificationResult, error)
	RefreshLegendaryActions()
	CanTakeActions() bool
	ProcessTurn(actorID int, turnType TurnType) (*TurnResult, *AIRequest, error)
	GetConditions() EntityConditions
	GetType() string
	IsConcentrating() bool
	BreakConcentration()
	SetConcentrating(val bool, spellName string)
}

type Combatant struct {
	Entity     Entity
	Info       *CombatantInfo
	Initiative int
	IsLair     bool
}

// NewCombatant creates a new Combatant with the specified Entity and initiative, defaulting CanAct to true.
func NewCombatant(entity Entity, info *CombatantInfo, initiative int) *Combatant {
	return &Combatant{entity, info, initiative, false}
}

func NewCombatantWithInfo(entity Entity) *Combatant {
	// Create combatant with nil info initially (circular dependency)
	combatant := &Combatant{
		Entity:     entity,
		Info:       nil,
		Initiative: 0,
		IsLair:     false,
	}

	// Create and attach CombatantInfo
	combatant.Info = NewCombatantInfo(combatant)

	// Initialize state from entity
	combatant.Info.UpdateState()

	return combatant
}

func (c *Combatant) GetInitiative() int {
	return c.Initiative
}

func (c *Combatant) GetEntity() Entity {
	return c.Entity
}
