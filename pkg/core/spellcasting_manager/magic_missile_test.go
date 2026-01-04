package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/testhelpers"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"testing"
)

func TestMagicMissile_AutoHit(t *testing.T) {
	actor := testhelpers.NewEmEntity(1, core.AbilityScores{Intelligence: 16}, nil)
	target := testhelpers.NewEmEntity(2, core.AbilityScores{}, nil)
	rm := roll_manager.NewRollManager(actor, roll_manager.RerollAbilities{})
	// Manually set RNG since EmEntity returns nil
	rm.SetRNG(1, 1)

	scm := NewSpellcastingManager(actor, rm, core.CasterCharacter, 1, spells.SpellSlots{1: 2}, spells.SpellSlots{1: 2}, 3)

	formula := core.CastFormula{NumberOfDice: 3, Die: core.D4, AmountToAdd: 3}
	mm := &spells.Spell{
		ID:        1,
		Name:      "Magic Missile",
		Level:     1,
		SpellType: core.STDamage,
		IsAutoHit: true,
		Formulas: map[int]core.CastFormula{
			1: formula, // 3 darts, each 1d4+1. Simplified to 3d4+3
		},
	}
	scm.AddKnownSpell(mm)

	choice := core.SpellChoice{
		Spell:   mm,
		Formula: &formula,
	}

	req := &SpellCastRequest{
		SpellCastData: SpellCastData{
			SpellChoice:          choice,
			AttackModifier:       5,
			SpellcastingModifier: 3,
		},
		SpellOptions:      SpellOptions{},
		SimulationOptions: &core.SimulationOptions{},
		Target:            target,
	}

	res, err := scm.CastSpell(req)
	if err != nil {
		t.Fatalf("CastSpell failed: %v", err)
	}

	if !res.GetIsHit() {
		t.Errorf("Expected Magic Missile to auto-hit")
	}

	if res.GetAttackRoll() != 0 {
		t.Errorf("Expected attack roll to be 0 for auto-hit spell, got %d", res.GetAttackRoll())
	}

	if res.GetSpellTotalValue() < 6 || res.GetSpellTotalValue() > 15 {
		t.Errorf("Expected damage between 6 and 15, got %d", res.GetSpellTotalValue())
	}
}
