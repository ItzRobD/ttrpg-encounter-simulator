package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"math/rand/v2"
	"testing"
)

func TestHandleIndomitable(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	rm := roll_manager.NewRollManager(rng)
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{EnableSpecialAbilities: true})
	ed.RollManager = rm

	fighter := &actor.Actor{
		InstanceID: 1,
		Name:       "Fighter",
		Features: []core.Feature{
			{
				Name: core.SpecAbilityIndomitable,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfSavingThrow: true,
				},
				Data: core.FeatureData{
					Value: 1,
				},
			},
		},
		StateManager: state_manager.StateManager{
			Resource: make(map[string]int),
		},
	}
	fighter.StateManager.Resource[string(core.SpecAbilityIndomitable)] = 1

	ed.Actors[1] = fighter

	t.Run("Indomitable Rerolls on Failure", func(t *testing.T) {
		// Reset state
		fighter.Features[0].Data.Value = 1
		fighter.StateManager.Resource[string(core.SpecAbilityIndomitable)] = 1

		// Mock failed save
		saveSuccess := false
		opts := &roll_manager.RollOptions{Advantage: core.RollNormal}
		action := &core.Action{
			HasDC:    true,
			DCSaveDC: 30, // Extremely high DC to likely fail again or check reroll logic
		}

		ctx := &FeatureContext{
			Target: fighter,
			SaveContext: &SaveContext{
				Target:      fighter,
				Options:     opts,
				SaveSuccess: saveSuccess,
				IsPostRoll:  true,
			},
			AttackContext: &AttackContext{
				Action: action,
			},
		}

		// We want to see if f.Data.Value decreases
		err := ed.HandleIndomitable(fighter, fighter.Features[0], core.HookOnSelfSavingThrow, ctx)
		if err != nil {
			t.Fatalf("HandleIndomitable failed: %v", err)
		}

		if fighter.Features[0].Data.Value != 0 {
			t.Errorf("Expected Indomitable uses to be 0, got %d", fighter.Features[0].Data.Value)
		}

		if fighter.StateManager.Resource[string(core.SpecAbilityIndomitable)] != 0 {
			t.Errorf("Expected Indomitable resource to be 0, got %d", fighter.StateManager.Resource[string(core.SpecAbilityIndomitable)])
		}
	})

	t.Run("Indomitable Does Not Trigger on Success", func(t *testing.T) {
		// Reset state
		fighter.Features[0].Data.Value = 1
		fighter.StateManager.Resource[string(core.SpecAbilityIndomitable)] = 1

		// Mock successful save
		saveSuccess := true
		opts := &roll_manager.RollOptions{Advantage: core.RollNormal}
		action := &core.Action{
			HasDC:    true,
			DCSaveDC: 10,
		}

		ctx := &FeatureContext{
			Target: fighter,
			SaveContext: &SaveContext{
				Target:      fighter,
				Options:     opts,
				SaveSuccess: saveSuccess,
				IsPostRoll:  true,
			},
			AttackContext: &AttackContext{
				Action: action,
			},
		}

		err := ed.HandleIndomitable(fighter, fighter.Features[0], core.HookOnSelfSavingThrow, ctx)
		if err != nil {
			t.Fatalf("HandleIndomitable failed: %v", err)
		}

		if fighter.Features[0].Data.Value != 1 {
			t.Errorf("Expected Indomitable uses to remain 1, got %d", fighter.Features[0].Data.Value)
		}
	})

	t.Run("Indomitable Does Not Trigger when out of uses", func(t *testing.T) {
		// Reset state
		fighter.Features[0].Data.Value = 0
		fighter.StateManager.Resource[string(core.SpecAbilityIndomitable)] = 0

		// Mock failed save
		saveSuccess := false
		opts := &roll_manager.RollOptions{Advantage: core.RollNormal}
		action := &core.Action{
			HasDC:    true,
			DCSaveDC: 20,
		}

		ctx := &FeatureContext{
			Target: fighter,
			SaveContext: &SaveContext{
				Target:      fighter,
				Options:     opts,
				SaveSuccess: saveSuccess,
				IsPostRoll:  true,
			},
			AttackContext: &AttackContext{
				Action: action,
			},
		}

		err := ed.HandleIndomitable(fighter, fighter.Features[0], core.HookOnSelfSavingThrow, ctx)
		if err != nil {
			t.Fatalf("HandleIndomitable failed: %v", err)
		}

		if ctx.SaveContext.SaveSuccess != false {
			t.Error("Expected SaveSuccess to remain false")
		}
	})
}
