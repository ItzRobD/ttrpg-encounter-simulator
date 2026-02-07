package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleEvasion(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Rogue with Evasion
	rogue := &actor.Actor{
		InstanceID: 1,
		Name:       "Rogue",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		AC:         15,
		Abilities:  core.Abilities{AbilityScores: core.AbilityScores{Dexterity: 18}}, // +4
		StateManager: state_manager.StateManager{
			CurrentHP:   30,
			MaxHP:       30,
			HealthState: core.HealthStateHealthy,
			Conditions:  core.NewActorConditions(),
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityEvasion,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfSavingThrow: true,
				},
			},
		},
	}

	// Attacker (Dragon)
	dragon := &actor.Actor{
		InstanceID: 2,
		Name:       "Dragon",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			HealthState: core.HealthStateHealthy,
			Conditions:  core.NewActorConditions(),
		},
	}

	ed.Actors[1] = rogue
	ed.Actors[2] = dragon
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	// Fire Breath Action (Dex Save, Half on Success)
	fireBreath := core.Action{
		Name:       "Fire Breath",
		ActionType: core.ATAction,
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 10, Die: core.D6, Modifier: 0, DamageType: core.DamageFire}, // ~35 damage
		},
		HasDC:       true,
		DCSaveDC:    15,
		DCAbility:   core.AbilityDexterity,
		DCOnSuccess: core.DCOnSuccessHalf,
	}

	// --- Case 1: Success on Save ---
	// Rogue has +4 Dex, so needs 11+ on roll.
	// To guarantee success, let's manipulate the roll or just test multiple times if needed,
	// but better to use a predictable setup if possible.
	// Actually, our RollManager is using ed.rng.
	// With Seed1:1, Seed2:1, let's see what happens.

	// Better yet, just call resolveDamage with saveSuccess = true/false after ResolveSavingThrow
	// But Evasion modifies the ACTION itself via hooks in ResolveSavingThrow.

	// Let's mock a fixed dice roller if we had one, but we don't.
	// I'll just check the logic by calling the handler directly first to ensure it flips the bits.

	t.Run("Evasion Flips Half to None on Success", func(t *testing.T) {
		actionCopy := fireBreath
		val := 100
		ctx := &FeatureContext{
			Target: rogue,
			SaveContext: &SaveContext{
				Target:      rogue,
				SaveSuccess: true,
				IsPostRoll:  true,
			},
			AttackContext: &AttackContext{
				Action: &actionCopy,
			},
			DamageContext: &DamageContext{
				DamageValue: &val,
			},
		}

		err := ed.HandleEvasion(rogue, rogue.Features[0], core.HookOnSelfSavingThrow, ctx)
		if err != nil {
			t.Fatalf("HandleEvasion failed: %v", err)
		}

		if val != 0 {
			t.Errorf("Expected DamageValue to be 0, got %d", val)
		}
		if actionCopy.DCOnSuccess != core.DCOnSuccessOther {
			t.Errorf("Expected DCOnSuccess to be Other, got %v", actionCopy.DCOnSuccess)
		}
	})

	t.Run("Evasion Flips Half on Failure", func(t *testing.T) {
		actionCopy := fireBreath
		val := 100
		ctx := &FeatureContext{
			Target: rogue,
			SaveContext: &SaveContext{
				Target:      rogue,
				SaveSuccess: false,
				IsPostRoll:  true,
			},
			AttackContext: &AttackContext{
				Action: &actionCopy,
			},
			DamageContext: &DamageContext{
				DamageValue: &val,
			},
		}

		err := ed.HandleEvasion(rogue, rogue.Features[0], core.HookOnSelfSavingThrow, ctx)
		if err != nil {
			t.Fatalf("HandleEvasion failed: %v", err)
		}

		if val != 50 {
			t.Errorf("Expected DamageValue to be 50, got %d", val)
		}
		if actionCopy.DCOnSuccess != core.DCOnSuccessOther {
			t.Errorf("Expected DCOnSuccess to be Other, got %v", actionCopy.DCOnSuccess)
		}
	})

	t.Run("Full Integration Test", func(t *testing.T) {
		// This tests the flow in Adjudicator
		// We want to see if resolveDamage ends up with 0 damage on success

		// Re-initialize EncounterDirector to ensure clean state
		ed = NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

		// Add actors properly
		ed.AddActor(rogue)
		ed.AddActor(dragon)

		simOptions.EnableSpecialAbilities = true
		ed.SetupEncounter()

		// Manually override feature to be sure it's the one we want
		rogue.Features = []core.Feature{
			{
				Name: core.SpecAbilityEvasion,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfSavingThrow: true,
				},
			},
		}

		// Force a success by setting DC very low
		easyAction := fireBreath
		easyAction.DCSaveDC = -100 // Ensure success even with modifiers
		easyAction.ID = core.MakeID("easy_fire_breath")
		rogue.StateManager.CurrentHP = 100
		rogue.StateManager.MaxHP = 100
		rogue.StateManager.HealthState = core.HealthStateHealthy

		err := ed.Adjudicator.executeIndividualStrike(dragon, rogue, &easyAction)
		if err != nil {
			t.Fatalf("executeIndividualStrike failed: %v", err)
		}

		// Since Evasion modifies the action, we check if DCOnSuccess was flipped to None
		// NOTE: HandleEvasion is called twice in ResolveSavingThrow, so it should be flipped.
		// However, it's called on the target's copy of the action in Adjudicator if it's passed by pointer.

		if rogue.StateManager.CurrentHP != 100 {
			t.Errorf("Expected 0 damage on successful save with Evasion, Rogue HP is %d", rogue.StateManager.CurrentHP)
		}

		// Force a failure by setting DC very high
		hardAction := fireBreath
		hardAction.DCSaveDC = 200 // Ensure failure even with modifiers
		hardAction.ID = core.MakeID("hard_fire_breath")

		// Reset Rogue HP
		rogue.StateManager.CurrentHP = 100
		rogue.StateManager.MaxHP = 100
		rogue.StateManager.HealthState = core.HealthStateHealthy

		err = ed.Adjudicator.executeIndividualStrike(dragon, rogue, &hardAction)
		if err != nil {
			t.Fatalf("executeIndividualStrike failed: %v", err)
		}

		// On failure, should take some damage
		if rogue.StateManager.CurrentHP >= 100 {
			t.Errorf("Expected damage on failed save with Evasion, Rogue HP is %d", rogue.StateManager.CurrentHP)
		}
		t.Logf("Rogue HP after failed save: %d", rogue.StateManager.CurrentHP)
	})
}
