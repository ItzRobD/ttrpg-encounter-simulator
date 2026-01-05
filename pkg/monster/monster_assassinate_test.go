package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"math/rand/v2"
	"testing"
)

func TestMonster_Assassinate(t *testing.T) {
	m := newSeededMonster(t)
	m.SpecialAbilities.Assassinate = true

	// Configure a simple action
	act := Action{ActionID: 1, Name: "Dagger", NumberOfDice: 1, Die: core.D4, AttackBonus: 5, DamageType: core.DamagePiercing}
	cfg := MAMConfig{Actions: map[int]Action{1: act}}
	m.ActionManager = NewMonsterActionManager(m, m.RollManager, &cfg)
	m.AI = NewMonsterAI(m, nil)

	target := &assassinateTargetStub{}
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ctx := core.NewCombatContext(simOptions)
	m.AI.UpdateCombatContext(ctx)

	// 1. Target hasn't taken a turn - should have advantage
	target.hasTakenTurn = false
	req, _ := m.createAttackRequest(target, 1, core.ATMonsterAction, simOptions)
	if req.AttackOptions.Advantage != core.RollAdvantage {
		t.Errorf("Expected advantage against target who hasn't taken a turn, got %v", req.AttackOptions.Advantage)
	}

	// 2. Target has taken a turn - should NOT have advantage
	target.hasTakenTurn = true
	req, _ = m.createAttackRequest(target, 1, core.ATMonsterAction, simOptions)
	if req.AttackOptions.Advantage != core.RollNormal {
		t.Errorf("Expected normal roll against target who has taken a turn, got %v", req.AttackOptions.Advantage)
	}
}

type assassinateTargetStub struct {
	hasTakenTurn bool
}

func (s *assassinateTargetStub) GetEntityType() core.EntityType {
	//TODO implement me
	panic("implement me")
}

func (s *assassinateTargetStub) GetHasTakenTurnInCombat() bool {
	return s.hasTakenTurn
}

func (s *assassinateTargetStub) GetAC() int {
	return 10
}

func (s *assassinateTargetStub) IsDead() bool {
	return false
}

func (s *assassinateTargetStub) IsUnconscious() bool {
	return false
}

func (s *assassinateTargetStub) GetConditions() core.EntityConditions {
	return core.NewEntityConditions()
}

func (s *assassinateTargetStub) IsMonster() bool {
	return true
}

func (s *assassinateTargetStub) IsCharacter() bool {
	return false
}

func (s *assassinateTargetStub) GetName() string {
	return "Target"
}

func (s *assassinateTargetStub) GetClassID() uint8                          { return 0 }
func (s *assassinateTargetStub) GetHPStatus() core.HPStatus                 { return core.NewHPStatusStub() }
func (s *assassinateTargetStub) GetEventListener() func(event interface{})  { return nil }
func (s *assassinateTargetStub) SetEventListener(f func(event interface{})) {}
func (s *assassinateTargetStub) GetLevel() float64                          { return 1 }
func (s *assassinateTargetStub) GetHitDie() core.DiceType                   { return core.D8 }
func (s *assassinateTargetStub) GetCasterLevel() int                        { return 0 }
func (s *assassinateTargetStub) MakeSavingThrow(core.Ability, int, bool, core.DamageType, *core.SimulationOptions) (core.RollResult, error) {
	return nil, nil
}
func (s *assassinateTargetStub) GetSpellSaveDC(*core.Ability) (int, error)         { return 0, nil }
func (s *assassinateTargetStub) GetAbilityScores() core.AbilityScores              { return core.AbilityScores{} }
func (s *assassinateTargetStub) GetAbilityScore(core.Ability) int                  { return 10 }
func (s *assassinateTargetStub) GetAbilityScoreModifier(core.Ability) (int, error) { return 0, nil }
func (s *assassinateTargetStub) GetSavingThrowBonus(core.Ability) (int, error)     { return 0, nil }
func (s *assassinateTargetStub) GetIsLegendary() bool                              { return false }
func (s *assassinateTargetStub) GetHPConfig() core.HPConfig                        { return core.HPConfig{} }
func (s *assassinateTargetStub) GetState() interface{}                             { return nil }
func (s *assassinateTargetStub) RollInitiative() (int, error)                      { return 0, nil }
func (s *assassinateTargetStub) InitializeHP() error                               { return nil }
func (s *assassinateTargetStub) IsSpellcaster() bool                               { return false }
func (s *assassinateTargetStub) IsHealer() bool                                    { return false }
func (s *assassinateTargetStub) GetTargetPriority() core.TargetPriority            { return 0 }
func (s *assassinateTargetStub) SetTargetPriority(core.TargetPriority)             {}
func (s *assassinateTargetStub) ChooseSpellByHealingEfficiency(int) (*core.SpellChoice, error) {
	return nil, nil
}
func (s *assassinateTargetStub) ChooseDamageSpellByPriority(core.SpellPriority) (*core.SpellChoice, error) {
	return nil, nil
}
func (s *assassinateTargetStub) GetHealingSpellCount() int { return 0 }
func (s *assassinateTargetStub) GetDamageSpellCount() int  { return 0 }
func (s *assassinateTargetStub) GetRNG() *rand.Rand        { return nil }
func (s *assassinateTargetStub) GetAIRequest(int, core.AIRequestType) (*core.AIRequest, error) {
	return nil, nil
}
func (s *assassinateTargetStub) ExecuteAIRequest(*core.AIRequest) (*core.ActionOutcome, error) {
	return nil, nil
}
func (s *assassinateTargetStub) UpdateAICombatContext(*core.CombatContext) error { return nil }
func (s *assassinateTargetStub) ModifyHP(int, bool, bool, bool, core.DamageType, bool) (core.HPModificationResult, error) {
	return nil, nil
}
func (s *assassinateTargetStub) RefreshLegendaryActions() {}
func (s *assassinateTargetStub) CanTakeActions() bool     { return true }
func (s *assassinateTargetStub) ProcessTurn(int, core.TurnType) (*core.TurnResult, *core.AIRequest, error) {
	return nil, nil, nil
}
func (s *assassinateTargetStub) GetType() string               { return "Monster" }
func (s *assassinateTargetStub) IsConcentrating() bool         { return false }
func (s *assassinateTargetStub) BreakConcentration()           {}
func (s *assassinateTargetStub) SetConcentrating(bool, string) {}
func (s *assassinateTargetStub) Regenerate()                   {}
func (s *assassinateTargetStub) GetID() int                    { return 0 }
func (s *assassinateTargetStub) GetInstanceID() int            { return 0 }
func (s *assassinateTargetStub) SetInstanceID(int)             {}
func (s *assassinateTargetStub) GetAttackBonus() int           { return 0 }
