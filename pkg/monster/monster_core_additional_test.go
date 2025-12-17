package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"testing"
)

// Test that ExecuteAIRequest produces damage effects for a normal action
func TestExecuteAIRequest_ActionProducesDamageEffects(t *testing.T) {
	m := newTestMonster(t)
	cfg := basicConfig()
	m.ActionManager = NewMonsterActionManager(m, m.RollManager, &cfg)

	tgt := targetStub{Entity: m} // AC=0 via stub

	req := &core.AIRequest{ // minimal fields used by ExecuteAIRequest
		ActionType:  core.ATMonsterAction,
		ActionIndex: 1,
		Target:      tgt,
		TargetID:    0,
		ActorID:     1,
	}

	out, err := m.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest(action) error: %v", err)
	}
	if out.ActionType != core.ATMonsterAction {
		t.Fatalf("unexpected action type in outcome: %v", out.ActionType)
	}
	if len(out.Effects) != 1 {
		t.Fatalf("expected 1 damage effect, got %d", len(out.Effects))
	}
	if out.Effects[0].Type != core.EffectDamage {
		t.Errorf("expected EffectDamage, got %v", out.Effects[0].Type)
	}
	if out.Effects[0].Value <= 0 {
		t.Errorf("expected positive damage value, got %d", out.Effects[0].Value)
	}
}

// Test multiattack and legendary flows via ExecuteAIRequest, and LAP decrement on legendary
func TestExecuteAIRequest_MultiattackAndLegendary(t *testing.T) {
	m := newTestMonster(t)
	cfg := basicConfig()
	m.ActionManager = NewMonsterActionManager(m, m.RollManager, &cfg)

	// Seed some legendary points
	m.EntityStateManager.LegendaryActionPointsMax = 3
	m.EntityStateManager.LegendaryActionPoints = 3

	tgt := targetStub{Entity: m}

	// Multiattack (two claws)
	mreq := &core.AIRequest{
		ActionType:  core.ATMonsterMultiattack,
		ActionIndex: 0,
		Target:      tgt,
		TargetID:    0,
		ActorID:     1,
	}
	mout, err := m.ExecuteAIRequest(mreq)
	if err != nil {
		t.Fatalf("ExecuteAIRequest(multiattack) error: %v", err)
	}
	if len(mout.Effects) != 2 {
		t.Fatalf("expected 2 damage effects for multiattack, got %d", len(mout.Effects))
	}

	// Legendary action (cost 1 in basicConfig)
	before := m.EntityStateManager.LegendaryActionPoints
	lreq := &core.AIRequest{
		ActionType:  core.ATLegendaryAction,
		ActionIndex: 100,
		Target:      tgt,
		TargetID:    0,
		ActorID:     1,
	}
	lout, err := m.ExecuteAIRequest(lreq)
	if err != nil {
		t.Fatalf("ExecuteAIRequest(legendary) error: %v", err)
	}
	if len(lout.Effects) != 1 {
		t.Fatalf("expected 1 damage effect for legendary, got %d", len(lout.Effects))
	}
	if m.EntityStateManager.LegendaryActionPoints != before-1 {
		t.Errorf("legendary points not decremented: before=%d after=%d", before, m.EntityStateManager.LegendaryActionPoints)
	}
}

// Test GetAIRequest invalid type error path
func TestGetAIRequest_InvalidType_ReturnsError(t *testing.T) {
	m := newTestMonster(t)
	// Pass an invalid AIRequestType value
	_, err := m.GetAIRequest(1, core.AIRequestType(255))
	if err == nil {
		t.Fatalf("expected error for invalid AI request type, got nil")
	}
}

// Test setHP variants through InitializeHP and verify state changes
func TestSetHP_Variants_ValueAverageRoll(t *testing.T) {
	m := newTestMonster(t)

	// Collect dice roll events
	var eventsSeen int
	m.SetEventListener(func(e interface{}) { eventsSeen++ })

	// Value
	m.HP = core.HPConfig{HPSetMethod: core.HPSetValue, Value: 15, HitDie: core.D8}
	if err := m.InitializeHP(); err != nil {
		t.Fatalf("InitializeHP value: %v", err)
	}
	if m.EntityStateManager.GetCurrentHP() != 15 || m.EntityStateManager.GetMaxHP() != 15 {
		t.Errorf("SetHP value failed: hp=%d max=%d", m.EntityStateManager.GetCurrentHP(), m.EntityStateManager.GetMaxHP())
	}

	// Average
	m.HP = core.HPConfig{HPSetMethod: core.HPSetAverage, HPAverage: 12, HitDie: core.D8}
	if err := m.InitializeHP(); err != nil {
		t.Fatalf("InitializeHP average: %v", err)
	}
	if m.EntityStateManager.GetCurrentHP() != 12 || m.EntityStateManager.GetMaxHP() != 12 {
		t.Errorf("SetHP average failed: hp=%d max=%d", m.EntityStateManager.GetCurrentHP(), m.EntityStateManager.GetMaxHP())
	}

	// Roll (deterministic due to RNG seed) - use a valid dice config
	m.HP = core.HPConfig{HPSetMethod: core.HPSetRoll, NumberOfDice: 2, HitDie: core.D8}
	if err := m.InitializeHP(); err != nil {
		t.Fatalf("InitializeHP roll: %v", err)
	}
	if m.EntityStateManager.GetCurrentHP() <= 0 || m.EntityStateManager.GetMaxHP() <= 0 {
		t.Errorf("SetHP roll failed, non-positive HP: hp=%d max=%d", m.EntityStateManager.GetCurrentHP(), m.EntityStateManager.GetMaxHP())
	}

	if eventsSeen == 0 {
		t.Errorf("expected at least one dice roll event to be emitted")
	}
}

// Test that modifying HP to <= 0 kills monsters immediately (per ESM behavior)
func TestModifyHP_KillsMonsterAtZeroOrBelow(t *testing.T) {
	m := newTestMonster(t)
	// Ensure non-zero HP
	m.EntityStateManager.ResetHP()
	if _, err := m.ModifyHP(-999, false, false); err != nil {
		t.Fatalf("ModifyHP error: %v", err)
	}
	if !m.IsDead() {
		t.Errorf("monster should be dead after massive damage")
	}
}
