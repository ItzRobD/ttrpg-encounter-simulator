package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"math/rand/v2"
	"testing"
)

// Reuse target stub with AC=0 for guaranteed hits
type targetZeroAC struct{ core.Entity }

func (t targetZeroAC) GetEntityType() core.EntityType {
	//TODO implement me
	panic("implement me")
}

func (t targetZeroAC) Regenerate() {
	//TODO implement me
	panic("implement me")
}

func (t targetZeroAC) GetAC() int      { return 0 }
func (t targetZeroAC) GetName() string { return "Target" }

// helper to build a minimal monster with seeded RNG and ESM
func newSeededMonster(t *testing.T) *Monster {
	t.Helper()
	m := &Monster{
		MonsterBase: MonsterBase{
			Name:             "Seeded",
			AC:               10,
			ProficiencyBonus: 2,
			AbilityScores:    core.AbilityScores{Dexterity: 14},
		},
		RNG: rand.New(rand.NewPCG(3, 4)),
	}
	m.RollManager = roll_manager.NewRollManager(m, roll_manager.RerollAbilities{})
	esm, err := entity_state_manager.NewEntityStateManager(m, entity_state_manager.EntityStateConfig{
		CurrentHP:  10,
		MaxHP:      10,
		Conditions: core.NewEntityConditions(),
	})
	if err != nil {
		t.Fatalf("NewEntityStateManager: %v", err)
	}
	m.EntityStateManager = esm
	return m
}

func TestRollInitiative_WritesState(t *testing.T) {
	m := newSeededMonster(t)
	total, err := m.RollInitiative()
	if err != nil {
		t.Fatalf("RollInitiative error: %v", err)
	}
	if total != m.EntityStateManager.GetInitiative() {
		t.Errorf("initiative state mismatch: got=%d state=%d", total, m.EntityStateManager.GetInitiative())
	}
}

func TestLegendaryAction_NoPoints_RemainsZero(t *testing.T) {
	m := newSeededMonster(t)
	// Configure a simple legendary action
	act := Action{ActionID: 1, Name: "Claw", NumberOfDice: 1, Die: core.D6, AmountToAdd: 1, AttackBonus: 5, DamageType: core.DamageSlashing}
	cfg := MAMConfig{LegendaryActions: map[int]LegendaryAction{100: {Cost: 1, Action: act}}}
	m.ActionManager = NewMonsterActionManager(m, m.RollManager, &cfg)

	// Set LAP to 0 and execute a legendary action
	m.EntityStateManager.SetLegendaryActionPointsMax(0)
	m.EntityStateManager.SetLegendaryActionPoints(0)

	tgt := targetZeroAC{Entity: m}
	req := &core.AIRequest{ActionType: core.ATLegendaryAction, ActionIndex: 100, Target: tgt, ActorID: 1}
	_, err := m.ExecuteAIRequest(req)
	if err != nil {
		// ExecuteAIRequest does not currently surface LAP spend errors; ensure it still does not crash
		t.Fatalf("ExecuteAIRequest legendary error: %v", err)
	}
	if m.EntityStateManager.GetLegendaryActionPoints() != 0 {
		t.Errorf("legendary points should remain 0, got %d", m.EntityStateManager.GetLegendaryActionPoints())
	}
}

func TestSetHP_InvalidMethod_ReturnsError(t *testing.T) {
	m := newSeededMonster(t)
	if err := m.SetHP(core.HPSetMethod(255), 5); err == nil {
		t.Fatalf("expected error for invalid HP set method")
	}
}

func TestTargetPriority_RoundTrip(t *testing.T) {
	m := newSeededMonster(t)
	m.SetTargetPriority(core.PrioritizeHighestLevel)
	if got := m.GetTargetPriority(); got != core.PrioritizeHighestLevel {
		t.Errorf("target priority mismatch: got=%v want=%v", got, core.PrioritizeHighestLevel)
	}
}

func TestSetEventListener_EmitsOnInitializeHP(t *testing.T) {
	m := newSeededMonster(t)
	var seen int
	m.SetEventListener(func(e interface{}) { seen++ })
	// Value path emits a dice roll event
	m.HP = core.HPConfig{HPSetMethod: core.HPSetValue, Value: 9, HitDie: core.D8}
	if err := m.InitializeHP(); err != nil {
		t.Fatalf("InitializeHP: %v", err)
	}
	if seen == 0 {
		t.Errorf("expected at least one event to be emitted on HP init (value)")
	}
}

func TestCreateAttackRequest_InvalidType_Error(t *testing.T) {
	m := newSeededMonster(t)
	// invalid action type should error
	if _, err := m.createAttackRequest(m, 0, core.ATNoAction, &core.SimulationOptions{}); err == nil {
		t.Fatalf("expected error for invalid monster action type")
	}
}
