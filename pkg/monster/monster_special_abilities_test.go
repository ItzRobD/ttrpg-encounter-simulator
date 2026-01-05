package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"testing"
)

func TestMonster_MartialAdvantage(t *testing.T) {
	m := newTestMonster(t)
	m.SpecialAbilities.MartialAdvantageNumDice = 2
	cfg := basicConfig()
	m.ActionManager = NewMonsterActionManager(m, m.RollManager, &cfg)

	tgt := targetStub{Entity: m}

	// Case 1: No allies (ConsciousMonsterCount = 1)
	ctx := core.NewCombatContext(&core.SimulationOptions{EnableSpecialAbilities: true})
	ctx.ConsciousMonsterCount = 1
	m.AI.UpdateCombatContext(ctx)

	req := &core.AIRequest{
		ActionType:  core.ATMonsterAction,
		ActionIndex: 1,
		Target:      tgt,
		TargetID:    0,
		ActorID:     1,
		SimOptions:  ctx.Opt(),
	}

	out, err := m.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest error: %v", err)
	}

	// Should only have base damage effect
	if len(out.Effects) != 1 {
		t.Errorf("expected 1 effect without allies, got %d", len(out.Effects))
	}

	// Case 2: With ally (ConsciousMonsterCount = 2)
	ctx.ConsciousMonsterCount = 2
	m.AI.UpdateCombatContext(ctx)

	out, err = m.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest error: %v", err)
	}

	// Should have base damage + Martial Advantage damage
	if len(out.Effects) != 2 {
		t.Errorf("expected 2 effects with allies, got %d", len(out.Effects))
	}

	if out.Effects[1].Type != core.EffectDamage || out.Effects[1].DamageType != core.DamageSlashing {
		t.Errorf("unexpected Martial Advantage effect: %+v", out.Effects[1])
	}

	// Case 3: 1/turn limit
	out, err = m.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest error: %v", err)
	}

	// Should NOT have Martial Advantage damage on second hit in the same turn
	if len(out.Effects) != 1 {
		t.Errorf("expected 1 effect on second hit (1/turn limit), got %d", len(out.Effects))
	}
}

func TestMonster_MartialAdvantage_ImprovedCritical(t *testing.T) {
	m := newTestMonster(t)
	m.SpecialAbilities.MartialAdvantageNumDice = 2
	cfg := basicConfig()
	m.ActionManager = NewMonsterActionManager(m, m.RollManager, &cfg)

	// Enable special abilities AND Improved Criticals
	ctx := core.NewCombatContext(&core.SimulationOptions{
		EnableSpecialAbilities: true,
		UseImprovedCriticals:   true,
	})
	ctx.ConsciousMonsterCount = 2
	m.AI.UpdateCombatContext(ctx)

	// We need to force a critical hit.
	// In ExecuteAIRequest, results are from ProcessAttackRequest which calls RollAttack.
	// Since we use a seeded RNG, we might need a lot of luck or a way to mock.
	// However, we can just call resolveMartialAdvantage directly for the test.

	// Normal critical: should double dice (4d6)
	ctx.Options.UseImprovedCriticals = false
	eff := m.resolveMartialAdvantage(true, ctx.Opt())
	// Min: 4, Max: 24.
	if eff.Value < 4 || eff.Value > 24 {
		t.Errorf("Normal crit: expected value between 4 and 24, got %d", eff.Value)
	}

	// Improved critical: should roll 2d6 and add 12 (2 * 6)
	// Actually RollExtraMaxDice(2, D6) rolls 2d6 and adds 2 * 6.
	// Min: 2 + 12 = 14, Max: 12 + 12 = 24.
	ctx.Options.UseImprovedCriticals = true
	m.EntityStateManager.SetHasUsedMartialAdvantage(false)
	eff = m.resolveMartialAdvantage(true, ctx.Opt())
	if eff.Value < 14 || eff.Value > 24 {
		t.Errorf("Improved crit: expected value between 14 and 24, got %d", eff.Value)
	}
}

func TestMonster_DivineEminence(t *testing.T) {
	m := newTestMonster(t)
	m.SpecialAbilities.DivineEminenceNumDice = 3
	cfg := basicConfig()
	m.ActionManager = NewMonsterActionManager(m, m.RollManager, &cfg)

	// Setup spellcasting
	m.MonsterBase.IsSpellcaster = true
	m.SpellCastingManager = spellcasting_manager.NewSpellcastingManager(m, m.RollManager, core.CasterMonsterTrueCaster, 5, spells.SpellSlots{1: 4}, spells.SpellSlots{1: 4}, 3)

	tgt := targetStub{Entity: m}
	ctx := core.NewCombatContext(&core.SimulationOptions{EnableSpecialAbilities: true})
	m.AI.UpdateCombatContext(ctx)

	// Case 1: Not activated yet
	req := &core.AIRequest{
		ActionType:  core.ATMonsterAction,
		ActionIndex: 1,
		Target:      tgt,
		TargetID:    0,
		ActorID:     1,
		SimOptions:  ctx.Opt(),
	}

	out, err := m.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest error: %v", err)
	}
	if len(out.Effects) != 1 {
		t.Errorf("expected 1 effect (not activated), got %d", len(out.Effects))
	}

	// Case 2: Activation
	activateReq := &core.AIRequest{
		ActionType: core.ATMonsterDivineEminence,
		ActorID:    1,
		SimOptions: ctx.Opt(),
	}

	out, err = m.ExecuteAIRequest(activateReq)
	if err != nil {
		t.Fatalf("Activation error: %v", err)
	}
	if !m.EntityStateManager.GetIsDivineEminenceActive() {
		t.Errorf("expected Divine Eminence to be active")
	}
	if m.EntityStateManager.GetDivineEminenceDice() != 3 {
		t.Errorf("expected 3 dice (1st level slot used), got %d", m.EntityStateManager.GetDivineEminenceDice())
	}
	if !m.EntityStateManager.GetHasUsedBonusAction() {
		t.Errorf("expected bonus action to be used")
	}

	// Case 3: Hit with activated Divine Eminence
	out, err = m.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest error: %v", err)
	}

	// Should have base damage + Divine Eminence damage
	if len(out.Effects) != 2 {
		t.Errorf("expected 2 effects, got %d", len(out.Effects))
	}

	if out.Effects[1].Type != core.EffectDamage || out.Effects[1].DamageType != core.DamageRadiant {
		t.Errorf("unexpected Divine Eminence effect: %+v", out.Effects[1])
	}
}

func TestMonster_DivineEminence_Uplevel(t *testing.T) {
	m := newTestMonster(t)
	m.SpecialAbilities.DivineEminenceNumDice = 3
	cfg := basicConfig()
	m.ActionManager = NewMonsterActionManager(m, m.RollManager, &cfg)

	// Setup spellcasting with only 2nd level slots
	m.MonsterBase.IsSpellcaster = true
	m.SpellCastingManager = spellcasting_manager.NewSpellcastingManager(m, m.RollManager, core.CasterMonsterTrueCaster, 5, spells.SpellSlots{2: 1}, spells.SpellSlots{2: 1}, 3)

	ctx := core.NewCombatContext(&core.SimulationOptions{EnableSpecialAbilities: true})
	m.AI.UpdateCombatContext(ctx)

	activateReq := &core.AIRequest{
		ActionType: core.ATMonsterDivineEminence,
		ActorID:    1,
		SimOptions: ctx.Opt(),
	}

	_, err := m.ExecuteAIRequest(activateReq)
	if err != nil {
		t.Fatalf("Activation error: %v", err)
	}

	// 2nd level slot should give +1 dice -> 4 dice total
	if m.EntityStateManager.GetDivineEminenceDice() != 4 {
		t.Errorf("expected 4 dice (2nd level slot used), got %d", m.EntityStateManager.GetDivineEminenceDice())
	}
}

func TestMonster_DivineEminence_ImprovedCritical(t *testing.T) {
	m := newTestMonster(t)
	m.SpecialAbilities.DivineEminenceNumDice = 3
	cfg := basicConfig()
	m.ActionManager = NewMonsterActionManager(m, m.RollManager, &cfg)

	// Setup spellcasting
	m.MonsterBase.IsSpellcaster = true
	m.SpellCastingManager = spellcasting_manager.NewSpellcastingManager(m, m.RollManager, core.CasterMonsterTrueCaster, 5, spells.SpellSlots{1: 4}, spells.SpellSlots{1: 4}, 3)

	ctx := core.NewCombatContext(&core.SimulationOptions{
		EnableSpecialAbilities: true,
		UseImprovedCriticals:   true,
	})
	m.AI.UpdateCombatContext(ctx)

	// Activate
	m.EntityStateManager.SetDivineEminenceActive(true)
	m.EntityStateManager.SetDivineEminenceDice(3)

	// Normal critical: should double dice (6d6)
	ctx.Options.UseImprovedCriticals = false
	eff := m.resolveDivineEminence(true, ctx.Opt())
	// Min: 6, Max: 36.
	if eff.Value < 6 || eff.Value > 36 {
		t.Errorf("Normal crit: expected value between 6 and 36, got %d", eff.Value)
	}

	// Improved critical: should roll 3d6 and add 18 (3 * 6)
	// Min: 3 + 18 = 21, Max: 18 + 18 = 36.
	ctx.Options.UseImprovedCriticals = true
	eff = m.resolveDivineEminence(true, ctx.Opt())
	if eff.Value < 21 || eff.Value > 36 {
		t.Errorf("Improved crit: expected value between 21 and 36, got %d", eff.Value)
	}
}

type characterTargetStub struct {
	targetStub
}

func (s *characterTargetStub) GetEntityType() core.EntityType {
	//TODO implement me
	panic("implement me")
}

func (s *characterTargetStub) IsMonster() bool   { return false }
func (s *characterTargetStub) IsCharacter() bool { return true }
func (s *characterTargetStub) GetHPStatus() core.HPStatus {
	return core.NewHPStatusStub()
}
func (s *characterTargetStub) GetConditions() core.EntityConditions {
	return core.NewEntityConditions()
}
func (s *characterTargetStub) IsConcentrating() bool { return false }
func (s *characterTargetStub) IsUnconscious() bool   { return false }
func (s *characterTargetStub) IsDead() bool          { return false }
func (s *characterTargetStub) GetInstanceID() int    { return 0 }
func (s *characterTargetStub) GetName() string       { return "Stub" }

func TestMonster_MultipleActionsInTurn(t *testing.T) {
	m := newTestMonster(t)
	m.SpecialAbilities.DivineEminenceNumDice = 3
	cfg := basicConfig()
	m.ActionManager = NewMonsterActionManager(m, m.RollManager, &cfg)

	// Setup spellcasting
	m.MonsterBase.IsSpellcaster = true
	m.SpellCastingManager = spellcasting_manager.NewSpellcastingManager(m, m.RollManager, core.CasterMonsterTrueCaster, 5, spells.SpellSlots{1: 4}, spells.SpellSlots{1: 4}, 3)
	// Add a damage spell to satisfy AI preference
	dmgSpell := &spells.Spell{
		ID:        2,
		Name:      "Sacred Flame",
		Level:     0,
		SpellType: core.STDamage,
		Formulas: map[int]core.CastFormula{
			0: {NumberOfDice: 1, Die: core.D8},
		},
	}
	m.SpellCastingManager.AddKnownSpell(dmgSpell)

	m.AI = NewMonsterAI(m, nil)
	// targetStub is a monster because it embeds m.
	// We need a target that is NOT a monster.
	tgt := &characterTargetStub{}

	opts := &core.SimulationOptions{EnableSpecialAbilities: true}
	ctx := core.NewCombatContext(opts)
	// Add both to combat context info
	ctx.CombatantInfo[1] = &core.CombatantInfo{Combatant: core.NewCombatantWithInfo(m)}
	ctx.CombatantInfo[2] = &core.CombatantInfo{Combatant: core.NewCombatantWithInfo(tgt)}
	m.AI.UpdateCombatContext(ctx)

	// First call to GetAIRequest should return Divine Eminence (Bonus Action)
	req1, err := m.GetAIRequest(1, core.AIReqNormalAction)
	if err != nil {
		t.Fatalf("First GetAIRequest failed: %v", err)
	}
	if req1.ActionType != core.ATMonsterDivineEminence {
		t.Errorf("Expected ATMonsterDivineEminence, got %v", req1.ActionType)
	}

	// Execute it
	_, err = m.ExecuteAIRequest(req1)
	if err != nil {
		t.Fatalf("Execution 1 failed: %v", err)
	}

	if !m.EntityStateManager.GetHasUsedBonusAction() {
		t.Errorf("Expected bonus action to be used")
	}

	// We must NOT replenish actions if we want to test multiple actions in one turn logic
	// m.EntityStateManager.RefreshActions() is usually called by the combat engine at start of turn.
	// GetAIRequest uses action economy to decide what's available.

	// Second call to GetAIRequest should return a standard action (Action) or spell attack
	req2, err := m.GetAIRequest(1, core.AIReqNormalAction)
	if err != nil {
		t.Fatalf("Second GetAIRequest failed: %v", err)
	}
	if req2 == nil {
		// Log state to debug
		t.Errorf("State: HasUsedAction=%v, HasUsedBonusAction=%v", m.EntityStateManager.GetHasUsedAction(), m.EntityStateManager.GetHasUsedBonusAction())
		t.Fatalf("Expected second request, got nil")
	}
	if req2.ActionType != core.ATMonsterAction && req2.ActionType != core.ATMonsterMultiattack && req2.ActionType != core.ATSpell {
		t.Errorf("Expected standard action or spell, got %v", req2.ActionType)
	}

	// Execute it
	_, err = m.ExecuteAIRequest(req2)
	if err != nil {
		t.Fatalf("Execution 2 failed: %v", err)
	}

	if !m.EntityStateManager.GetHasUsedAction() {
		t.Errorf("Expected action to be used")
	}

	// Third call should return nil
	req3, err := m.GetAIRequest(1, core.AIReqNormalAction)
	if err != nil {
		t.Fatalf("Third GetAIRequest failed: %v", err)
	}
	if req3 != nil {
		t.Errorf("Expected nil for third request, got %v", req3.ActionType)
	}
}
