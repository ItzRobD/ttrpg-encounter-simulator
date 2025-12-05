package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"math/rand/v2"
	"testing"
)

// target stub that only controls AC
type targetStub struct{ core.Entity }

func (t targetStub) GetAC() int      { return 0 }
func (t targetStub) GetName() string { return "Target" }

func newTestMonster(t *testing.T) *Monster {
	t.Helper()
	m := &Monster{
		MonsterBase: MonsterBase{
			Name:             "TestMonster",
			AC:               10,
			ProficiencyBonus: 2,
			AbilityScores:    core.AbilityScores{Strength: 16},
		},
		RNG: rand.New(rand.NewPCG(1, 2)),
	}
	// Roll manager
	m.RollManager = roll_manager.NewRollManager(m, roll_manager.RerollAbilities{})
	// Entity state
	esm, err := entity_state_manager.NewEntityStateManager(m, entity_state_manager.EntityStateConfig{
		CurrentHP:  10,
		MaxHP:      10,
		Conditions: core.NewEntityConditions(),
	})
	if err != nil {
		t.Fatalf("NewEntityStateManager: %v", err)
	}
	m.EntityState = esm
	return m
}

func basicConfig() MAMConfig {
	act := Action{ActionID: 1, Name: "Claw", NumberOfDice: 1, Die: core.D8, AmountToAdd: 2, AttackBonus: 5, DamageType: core.DamageSlashing}
	// Multiattack: two claws
	ma := map[int][]Multiattack{0: {{ActionID: 1, Count: 2}}}
	// Legendary mirrors action
	leg := map[int]LegendaryAction{100: {Cost: 1, Action: act}}
	return MAMConfig{
		Actions:          map[int]Action{1: act},
		Multiattacks:     ma,
		LegendaryActions: leg,
		SpecialAbilities: nil,
	}
}

func TestPrecomputeAndGetAttackData(t *testing.T) {
	m := newTestMonster(t)
	cfg := basicConfig()
	mam := NewMonsterActionManager(m, m.RollManager, &cfg)

	// Precomputed maps should be populated
	if len(mam.ActionAttackData) != 1 {
		t.Fatalf("ActionAttackData size=%d want 1", len(mam.ActionAttackData))
	}
	if len(mam.LegendaryAttackData) != 1 {
		t.Fatalf("LegendaryAttackData size=%d want 1", len(mam.LegendaryAttackData))
	}
	if len(mam.MultiAttackData) != 1 {
		t.Fatalf("MultiAttackData size=%d want 1", len(mam.MultiAttackData))
	}

	// GetAttackDataFromIndex for each type
	if got := mam.GetAttackDataFromIndex(1, core.ATMonsterAction); len(got) != 1 {
		t.Errorf("action attack data len=%d want 1", len(got))
	}
	if got := mam.GetAttackDataFromIndex(0, core.ATMonsterMultiattack); len(got) != 2 {
		t.Errorf("multiattack data len=%d want 2", len(got))
	}
	if got := mam.GetAttackDataFromIndex(100, core.ATLegendaryAction); len(got) != 1 {
		t.Errorf("legendary data len=%d want 1", len(got))
	}
}

func TestProcessAttackRequest_ActionMultiLegendary(t *testing.T) {
	m := newTestMonster(t)
	cfg := basicConfig()
	mam := NewMonsterActionManager(m, m.RollManager, &cfg)

	tgt := targetStub{Entity: m} // AC overridden to 0 via stub method

	// Action
	ad := mam.GetAttackDataFromIndex(1, core.ATMonsterAction)
	req := &core.AttackRequest{AttackData: ad, AttackOptions: core.AttackOptions{}, Target: tgt}
	res, err := mam.ProcessAttackRequest(req)
	if err != nil || len(res) != 1 {
		t.Fatalf("action process err=%v len=%d", err, len(res))
	}
	if !res[0].IsHit || res[0].DamageRoll.GetTotal() <= 0 {
		t.Errorf("expected hit with damage > 0: %+v", res[0])
	}

	// Multiattack (two claws)
	mad := mam.GetAttackDataFromIndex(0, core.ATMonsterMultiattack)
	mreq := &core.AttackRequest{AttackData: mad, AttackOptions: core.AttackOptions{}, Target: tgt}
	mres, err := mam.ProcessAttackRequest(mreq)
	if err != nil || len(mres) != 2 {
		t.Fatalf("multiattack err=%v len=%d", err, len(mres))
	}

	// Legendary
	lad := mam.GetAttackDataFromIndex(100, core.ATLegendaryAction)
	lreq := &core.AttackRequest{AttackData: lad, AttackOptions: core.AttackOptions{}, Target: tgt}
	lres, err := mam.ProcessAttackRequest(lreq)
	if err != nil || len(lres) != 1 {
		t.Fatalf("legendary err=%v len=%d", err, len(lres))
	}
}

func TestRechargeFlow_WithInitializedMap(t *testing.T) {
	// Note: MonsterActionManager.InitializeActions writes to mam.RechargeActions.
	// The struct does not initialize this map, so we must initialize it in tests
	// to avoid a panic. This indicates a likely bug in production initialization.
	m := newTestMonster(t)
	// Action with RechargeValue 1 (always succeeds on d6)
	act := Action{ActionID: 2, Name: "Roar", RechargeValue: 1}
	cfg := MAMConfig{Actions: map[int]Action{2: act}}

	mam := NewMonsterActionManager(m, m.RollManager, nil)
	mam.RechargeActions = make(map[int]uint8) // prevent nil map write in InitializeActions
	mam.InitializeActions(&cfg)

	// Expend then roll recharge
	mam.ExpendRechargeAction(2)
	if mam.GetRechargeActionStatus()[2] {
		t.Fatalf("expected action 2 to be expended")
	}
	mam.RollRechargeActions()
	if !mam.GetRechargeActionStatus()[2] {
		t.Errorf("expected action 2 to be recharged after roll")
	}
}
