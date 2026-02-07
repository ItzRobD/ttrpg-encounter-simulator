package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestAdjudicator_SavingThrow(t *testing.T) {
	// Seed fixed to ensure predictable rolls if needed,
	// but here we can also mock the roll manager if it was an interface.
	// Since it's not, we'll rely on the fact that we can check if it works.
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{})
	adj := ed.Adjudicator

	attacker := &actor.Actor{InstanceID: 1, Name: "Attacker", Side: core.SideMonsters}
	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Target",
		Side:       core.SideCharacters,
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{Dexterity: 10}, // Mod 0
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 20,
			MaxHP:     20,
		},
	}
	ed.Actors[1] = attacker
	ed.Actors[2] = target

	// 1. Test Half on Success
	action := core.Action{
		ID:          core.MakeID(1),
		Name:        "Fireball-ish",
		HasDC:       true,
		DCAbility:   core.AbilityDexterity,
		DCSaveDC:    15,
		DCOnSuccess: core.DCOnSuccessHalf,
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 2, Die: core.D10, Modifier: 0, DamageType: core.DamageFire}, // Avg 11
		},
	}

	// We need to control the roll. EncounterDirector uses ed.RollManager.
	// Let's see if we can force a success and a failure by changing the DC.

	// Case A: Target Fails (DC is very high)
	action.DCSaveDC = 30
	target.StateManager.CurrentHP = 20
	err := adj.executeIndividualStrike(attacker, target, &action)
	if err != nil {
		t.Fatalf("ExecuteIndividualStrike failed: %v", err)
	}
	// Damage should be full. 2d10 avg is 11.
	if target.StateManager.CurrentHP >= 20 {
		t.Errorf("Target took no damage on failed save")
	}
	damageDealt := 20 - target.StateManager.CurrentHP
	t.Logf("Damage dealt on failed save: %d", damageDealt)

	// Case B: Target Succeeds (DC is very low)
	action.DCSaveDC = 1
	target.StateManager.CurrentHP = 20
	err = adj.executeIndividualStrike(attacker, target, &action)
	if err != nil {
		t.Fatalf("ExecuteIndividualStrike failed: %v", err)
	}
	// Damage should be half.
	if target.StateManager.CurrentHP >= 20 {
		t.Errorf("Target took no damage on successful save")
	}
	damageDealtSuccess := 20 - target.StateManager.CurrentHP
	t.Logf("Damage dealt on successful save: %d", damageDealtSuccess)

	// Note: since we use real random rolls (unless seeded), we can't check exact numbers easily
	// without many samples, but we can check the logic.
	// To be absolutely sure, we'd need to mock RollManager.
}

func TestGetAbilityModifier_Negative(t *testing.T) {
	testCases := []struct {
		score    int
		expected int
	}{
		{10, 0},
		{11, 0},
		{12, 1},
		{13, 1},
		{8, -1},
		{9, -1},
		{7, -2},
		{6, -2},
		{1, -5},
	}

	for _, tc := range testCases {
		a := core.Abilities{AbilityScores: core.AbilityScores{Strength: tc.score}}
		mod := a.GetAbilityModifier(core.AbilityStrength)
		if mod != tc.expected {
			t.Errorf("Score %d: expected mod %d, got %d", tc.score, tc.expected, mod)
		}
	}
}
