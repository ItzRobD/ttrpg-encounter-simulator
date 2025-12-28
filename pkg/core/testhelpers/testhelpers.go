package testhelpers

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"math/rand/v2"
)

type entityStub struct{}

// Only three methods really matter for EquipmentManager behavior
// Implement them on a small embedding type below.
func (entityStub) IsDead() bool                              { panic("not used") }
func (entityStub) IsUnconscious() bool                       { panic("not used") }
func (entityStub) GetHPStatus() core.HPStatus                { panic("not used") }
func (entityStub) GetName() string                           { return "Test" }
func (entityStub) GetAC() int                                { return 10 }
func (entityStub) GetEventListener() func(event interface{}) { return nil }
func (entityStub) SetEventListener(func(event interface{}))  {}
func (entityStub) GetLevel() float64                         { return 1 }
func (entityStub) GetHitDie() core.DiceType                  { return core.D8 }
func (entityStub) GetCasterLevel() int                       { return 0 }
func (entityStub) MakeSavingThrow(core.Ability, int, bool, core.DamageType) (core.RollResult, error) {
	panic("not used")
}
func (entityStub) GetSpellSaveDC(*core.Ability) (int, error)                     { panic("not used") }
func (entityStub) GetAbilityScores() core.AbilityScores                          { return core.AbilityScores{} }
func (entityStub) GetAbilityScore(core.Ability) int                              { return 10 }
func (entityStub) GetAbilityScoreModifier(core.Ability) (int, error)             { return 0, nil }
func (entityStub) GetSavingThrowBonus(core.Ability) (int, error)                 { return 0, nil }
func (entityStub) IsCharacter() bool                                             { return true }
func (entityStub) IsMonster() bool                                               { return false }
func (entityStub) GetIsLegendary() bool                                          { return false }
func (entityStub) GetHPConfig() core.HPConfig                                    { return core.HPConfig{} }
func (entityStub) GetState() interface{}                                         { return nil }
func (entityStub) RollInitiative() (int, error)                                  { return 0, nil }
func (entityStub) InitializeHP() error                                           { return nil }
func (entityStub) IsSpellcaster() bool                                           { return false }
func (entityStub) IsHealer() bool                                                { return false }
func (entityStub) GetTargetPriority() core.TargetPriority                        { return 0 }
func (entityStub) SetTargetPriority(core.TargetPriority)                         {}
func (entityStub) ChooseSpellByHealingEfficiency(int) (*core.SpellChoice, error) { return nil, nil }
func (entityStub) ChooseDamageSpellByPriority(core.SpellPriority) (*core.SpellChoice, error) {
	return nil, nil
}
func (entityStub) GetHealingSpellCount() int                                     { return 0 }
func (entityStub) GetDamageSpellCount() int                                      { return 0 }
func (entityStub) GetRNG() *rand.Rand                                            { return nil }
func (entityStub) GetAIRequest(int, core.AIRequestType) (*core.AIRequest, error) { return nil, nil }
func (entityStub) ExecuteAIRequest(*core.AIRequest) (*core.ActionOutcome, error) { return nil, nil }
func (entityStub) UpdateAICombatContext(*core.CombatContext) error               { return nil }
func (entityStub) ModifyHP(int, bool, bool) (core.HPModificationResult, error)   { panic("not used") }
func (entityStub) RefreshLegendaryActions()                                      {}
func (entityStub) CanTakeActions() bool                                          { return true }
func (entityStub) ProcessTurn(int, core.TurnType) (*core.TurnResult, *core.AIRequest, error) {
	return nil, nil, nil
}
func (entityStub) GetConditions() core.EntityConditions { return core.EntityConditions{} }
func (entityStub) GetType() string                      { return "Humanoid" }
func (entityStub) IsConcentrating() bool                { return false }
func (entityStub) BreakConcentration()                  {}
func (entityStub) SetConcentrating(bool, string)        {}

// Keep only what EquipmentManager needs here
type EmEntity struct {
	entityStub
	lvl     float64
	as      core.AbilityScores
	Monster bool
	classID uint8
}

func NewEmEntity(lvl float64, as core.AbilityScores, class *uint8) EmEntity {
	if class == nil {
		return EmEntity{lvl: lvl, as: as, Monster: false, classID: 0}
	} else {
		return EmEntity{lvl: lvl, as: as, Monster: false, classID: *class}
	}
}

func (e EmEntity) GetLevel() float64                    { return e.lvl }
func (e EmEntity) GetAbilityScores() core.AbilityScores { return e.as }
func (e EmEntity) GetAbilityScoreModifier(ability core.Ability) (int, error) {
	score := e.as.GetScore(ability)
	return (score - 10) / 2, nil
}
func (e EmEntity) GetClassID() uint8 { return e.classID }

// Identity overrides
func (e EmEntity) IsMonster() bool   { return e.Monster }
func (e EmEntity) IsCharacter() bool { return !e.Monster }

// NewEmMonster returns an entity that reports IsMonster() == true.
func NewEmMonster(lvl float64, as core.AbilityScores) EmEntity {
	return EmEntity{lvl: lvl, as: as, Monster: true}
}
