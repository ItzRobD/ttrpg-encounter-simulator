package monster

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"testing"
)

func TestMakeSavingThrow_LegendaryResistance(t *testing.T) {
	m := newSeededMonster(t)
	m.AbilityScores.Wisdom = 10

	// Set Legendary Resistance uses
	m.EntityStateManager.SetLegendaryResistanceUses(3)

	// We want a roll that fails.
	// newSeededMonster uses seed (3, 4) which I don't know the first result of.
	// But I can set a very high target value to ensure failure.
	targetValue := 100
	res, err := m.MakeSavingThrow(core.AbilityWisdom, targetValue, false, core.DamageNone, nil)
	if err != nil {
		t.Fatalf("MakeSavingThrow error: %v", err)
	}

	// It should have failed the roll initially, but Legendary Resistance makes it a success.
	if !res.GetIsSuccess() {
		t.Errorf("expected success due to Legendary Resistance, got failure")
	}

	if !res.GetWasRerolled() {
		t.Errorf("expected WasRerolled to be true (used as flag for LR)")
	}

	if m.EntityStateManager.GetLegendaryResistanceUses() != 2 {
		t.Errorf("expected 2 Legendary Resistance uses left, got %d", m.EntityStateManager.GetLegendaryResistanceUses())
	}
}

func TestMakeSavingThrow_LegendaryResistance_AutoFail(t *testing.T) {
	m := newSeededMonster(t)
	m.AbilityScores.Strength = 10

	// Set Legendary Resistance uses
	m.EntityStateManager.SetLegendaryResistanceUses(1)

	// Apply Paralyzed condition which causes auto-fail on STR/DEX saves
	m.GetConditions().Add(core.ConditionParalyzed)

	// Make a Strength saving throw
	targetValue := 15
	res, err := m.MakeSavingThrow(core.AbilityStrength, targetValue, false, core.DamageNone, nil)
	if err != nil {
		t.Fatalf("MakeSavingThrow error: %v", err)
	}

	// It should have failed due to condition, but Legendary Resistance makes it a success.
	if !res.GetIsSuccess() {
		t.Errorf("expected success due to Legendary Resistance on auto-fail, got failure")
	}

	if !res.GetWasRerolled() {
		t.Errorf("expected WasRerolled to be true (used as flag for LR)")
	}

	if m.EntityStateManager.GetLegendaryResistanceUses() != 0 {
		t.Errorf("expected 0 Legendary Resistance uses left, got %d", m.EntityStateManager.GetLegendaryResistanceUses())
	}
}

func TestMakeSavingThrow_LegendaryResistance_AutoFail_NoUses(t *testing.T) {
	m := newSeededMonster(t)
	m.AbilityScores.Strength = 10

	// Set Legendary Resistance uses to 0
	m.EntityStateManager.SetLegendaryResistanceUses(0)

	// Apply Paralyzed condition which causes auto-fail on STR/DEX saves
	m.GetConditions().Add(core.ConditionParalyzed)

	// Make a Strength saving throw
	targetValue := 15
	res, err := m.MakeSavingThrow(core.AbilityStrength, targetValue, false, core.DamageNone, nil)
	if err != nil {
		t.Fatalf("MakeSavingThrow error: %v", err)
	}

	// Should fail because no uses left
	if res.GetIsSuccess() {
		t.Errorf("expected failure as no Legendary Resistance uses left, got success")
	}
}

func TestMakeSavingThrow_NoLegendaryResistanceUses(t *testing.T) {
	m := newSeededMonster(t)
	m.AbilityScores.Wisdom = 10

	// Set Legendary Resistance uses to 0
	m.EntityStateManager.SetLegendaryResistanceUses(0)

	targetValue := 100

	res, err := m.MakeSavingThrow(core.AbilityWisdom, targetValue, false, core.DamageNone, nil)
	if err != nil {
		t.Fatalf("MakeSavingThrow error: %v", err)
	}

	// Should fail because no uses left
	if res.GetIsSuccess() {
		t.Errorf("expected failure as no Legendary Resistance uses left, got success")
	}
}

func TestMakeSavingThrow_MagicResistance(t *testing.T) {
	m := newSeededMonster(t)
	m.SpecialAbilities.MagicResistance = true
	m.AbilityScores.Wisdom = 10

	// Case 1: Is a spell, should have advantage
	res, err := m.MakeSavingThrow(core.AbilityWisdom, 10, true, core.DamageNone, nil)
	if err != nil {
		t.Fatalf("MakeSavingThrow error: %v", err)
	}
	if res.GetAdvantage() != core.RollAdvantage.String() {
		t.Errorf("expected advantage for spell save with Magic Resistance, got %s", res.GetAdvantage())
	}

	// Case 2: Not a spell, should be normal
	res, err = m.MakeSavingThrow(core.AbilityWisdom, 10, false, core.DamageNone, nil)
	if err != nil {
		t.Fatalf("MakeSavingThrow error: %v", err)
	}
	if res.GetAdvantage() != core.RollNormal.String() {
		t.Errorf("expected normal roll for non-spell save with Magic Resistance, got %s", res.GetAdvantage())
	}
}

func TestMakeSavingThrow_MagicResistance_EnabledFlag(t *testing.T) {
	m := newSeededMonster(t)
	m.SpecialAbilities.MagicResistance = true
	m.AbilityScores.Wisdom = 10

	// Case 1: Special abilities enabled
	simOptions := &core.SimulationOptions{EnableSpecialAbilities: true}
	res, err := m.MakeSavingThrow(core.AbilityWisdom, 10, true, core.DamageNone, simOptions)
	if err != nil {
		t.Fatalf("MakeSavingThrow error: %v", err)
	}
	if res.GetAdvantage() != core.RollAdvantage.String() {
		t.Errorf("expected advantage when special abilities are enabled")
	}

	// Case 2: Special abilities disabled
	simOptions.EnableSpecialAbilities = false
	res, err = m.MakeSavingThrow(core.AbilityWisdom, 10, true, core.DamageNone, simOptions)
	if err != nil {
		t.Fatalf("MakeSavingThrow error: %v", err)
	}
	if res.GetAdvantage() != core.RollNormal.String() {
		t.Errorf("expected normal roll when special abilities are disabled")
	}
}
