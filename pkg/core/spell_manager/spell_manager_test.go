package spell_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"testing"
)

func TestSpellManager_AddKnownSpell(t *testing.T) {
	scm := NewSpellManager()

	s := &spells.Spell{
		ID:        core.MakeID("1"),
		Name:      "Cure Wounds",
		Level:     1,
		SpellType: core.STHealing,
		Formulas: map[int][]core.CastFormula{
			1: {{NumberOfDice: 1, Die: core.D8, AmountToAdd: 3}},
		},
	}

	err := scm.AddKnownSpell(s)
	if err != nil {
		t.Fatalf("AddKnownSpell failed: %v", err)
	}

	if scm.HealingSpellCount != 1 {
		t.Errorf("Expected HealingSpellCount 1, got %d", scm.HealingSpellCount)
	}

	if len(scm.HealingSpells[1]) != 1 {
		t.Errorf("Expected 1 healing spell at level 1, got %d", len(scm.HealingSpells[1]))
	}

	// Check if average was calculated
	if s.Formulas[1][0].AverageValue == 0 {
		t.Error("AverageValue was not calculated")
	}
	// 1d8 (4.5) + 3 = 7.5 -> 7
	if s.Formulas[1][0].AverageValue != 7 {
		t.Errorf("Expected AverageValue 7, got %d", s.Formulas[1][0].AverageValue)
	}
}

func TestSpellManager_HasSpells(t *testing.T) {
	scm := NewSpellManager()

	if scm.HasHealingSpells() {
		t.Error("HasHealingSpells should be false initially")
	}

	s := &spells.Spell{
		ID:        core.MakeID("2"),
		Name:      "Fireball",
		Level:     3,
		SpellType: core.STDamage,
	}

	scm.AddKnownSpell(s)

	if !scm.HasDamageSpells() {
		t.Error("HasDamageSpells should be true")
	}

	if scm.GetSpellCount() != 1 {
		t.Errorf("Expected GetSpellCount 1, got %d", scm.GetSpellCount())
	}
}
