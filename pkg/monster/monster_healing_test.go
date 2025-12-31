package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"testing"
)

func TestMonster_CreateHealRequest(t *testing.T) {
	m := newSeededMonster(t)

	// Setup spellcasting with a healing spell
	scm := spellcasting_manager.NewSpellcastingManager(m, m.RollManager, core.CasterMonsterTrueCaster, 5, spells.SpellSlots{1: 4}, spells.SpellSlots{1: 4}, 3)
	healSpell := &spells.Spell{
		ID:        1,
		Name:      "Cure Wounds",
		Level:     1,
		SpellType: core.STHealing,
		Formulas: map[int]core.CastFormula{
			1: {NumberOfDice: 1, Die: core.D8, UseSpellmod: true},
		},
	}
	scm.AddKnownSpell(healSpell)
	m.SpellCastingManager = scm
	m.MonsterBase.IsSpellcaster = true

	target := newSeededMonster(t)
	target.EntityStateManager.ModifyHP(-5, false, false, false) // Needs 5 HP

	// Test CreateHealRequest
	req, err := m.CreateHealRequest(target)
	if err != nil {
		t.Fatalf("CreateHealRequest failed: %v", err)
	}

	if req.Source != core.HealSourceSpell {
		t.Errorf("Expected HealSourceSpell, got %v", req.Source)
	}
	if req.SpellChoice.Spell.GetName() != "Cure Wounds" {
		t.Errorf("Expected Cure Wounds, got %s", req.SpellChoice.Spell.GetName())
	}
}

func TestMonster_ExecuteAIRequest_Heal(t *testing.T) {
	m := newSeededMonster(t)

	// Setup spellcasting with a healing spell
	scm := spellcasting_manager.NewSpellcastingManager(m, m.RollManager, core.CasterMonsterTrueCaster, 5, spells.SpellSlots{1: 4}, spells.SpellSlots{1: 4}, 3)
	healSpell := &spells.Spell{
		ID:        1,
		Name:      "Cure Wounds",
		Level:     1,
		SpellType: core.STHealing,
		Formulas: map[int]core.CastFormula{
			1: {NumberOfDice: 1, Die: core.D8, UseSpellmod: true},
		},
	}
	scm.AddKnownSpell(healSpell)
	m.SpellCastingManager = scm
	m.MonsterBase.IsSpellcaster = true

	target := newSeededMonster(t)
	target.EntityStateManager.ModifyHP(-5, false, false, false) // Needs 5 HP

	healReq, err := m.CreateHealRequest(target)
	if err != nil {
		t.Fatalf("CreateHealRequest failed: %v", err)
	}

	aiReq := &core.AIRequest{
		Actor:       m,
		Target:      target,
		TargetID:    1,
		ActionType:  core.ATMonsterHeal,
		HealRequest: healReq,
	}

	outcome, err := m.ExecuteAIRequest(aiReq)
	if err != nil {
		t.Fatalf("ExecuteAIRequest failed: %v", err)
	}

	if outcome.ActionType != core.ATMonsterHeal {
		t.Errorf("Expected ActionType ATMonsterHeal, got %v", outcome.ActionType)
	}
	if len(outcome.Effects) != 1 || outcome.Effects[0].Type != core.EffectHealing {
		t.Errorf("Expected 1 healing effect, got %v", outcome.Effects)
	}
	if outcome.Effects[0].Value <= 0 {
		t.Errorf("Expected positive healing value, got %d", outcome.Effects[0].Value)
	}
}
