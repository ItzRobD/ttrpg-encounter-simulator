package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleLayOnHands(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	paladin := &actor.Actor{
		InstanceID: 1,
		Name:       "Paladin",
		Side:       core.SideCharacters,
		Metadata: actor.Metadata{
			Level: 5,
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
			Resource:  make(map[string]int),
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityLayOnHands,
				Hooks: map[core.HookType]bool{
					core.HookOnTurnStart: true,
				},
			},
		},
	}
	paladin.ProcessFeatures() // Should set pool to 25

	ally := &actor.Actor{
		InstanceID: 2,
		Name:       "Ally",
		Side:       core.SideCharacters,
		StateManager: state_manager.StateManager{
			CurrentHP: 10,
			MaxHP:     50,
		},
	}

	ed.Actors[1] = paladin
	ed.Actors[2] = ally
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Lay on Hands Heals Ally", func(t *testing.T) {
		// Reset state
		paladin.StateManager.ActionUsedCount = 0
		ally.StateManager.CurrentHP = 10
		paladin.StateManager.Resource[string(core.SpecAbilityLayOnHands)] = 25

		err := ed.HandleLayOnHands(paladin, paladin.Features[0], core.HookOnTurnStart, nil)
		if err != nil {
			t.Fatalf("HandleLayOnHands failed: %v", err)
		}

		// Ally missing 40 HP, Pool is 25. Should heal 25.
		if ally.StateManager.CurrentHP != 35 {
			t.Errorf("Expected Ally to have 35 HP, got %d", ally.StateManager.CurrentHP)
		}

		if paladin.StateManager.Resource[string(core.SpecAbilityLayOnHands)] != 0 {
			t.Errorf("Expected Pool to be 0, got %d", paladin.StateManager.Resource[string(core.SpecAbilityLayOnHands)])
		}

		if paladin.StateManager.ActionUsedCount != 1 {
			t.Error("Expected ActionUsedCount to be 1")
		}
	})

	t.Run("Lay on Hands Partially Heals Ally", func(t *testing.T) {
		// Reset state
		paladin.StateManager.ActionUsedCount = 0
		ally.StateManager.CurrentHP = 45
		paladin.StateManager.Resource[string(core.SpecAbilityLayOnHands)] = 25

		err := ed.HandleLayOnHands(paladin, paladin.Features[0], core.HookOnTurnStart, nil)
		if err != nil {
			t.Fatalf("HandleLayOnHands failed: %v", err)
		}

		// Ally missing 5 HP, Pool is 25. Should heal 5.
		if ally.StateManager.CurrentHP != 50 {
			t.Errorf("Expected Ally to have 50 HP, got %d", ally.StateManager.CurrentHP)
		}

		if paladin.StateManager.Resource[string(core.SpecAbilityLayOnHands)] != 20 {
			t.Errorf("Expected Pool to be 20, got %d", paladin.StateManager.Resource[string(core.SpecAbilityLayOnHands)])
		}
	})

	t.Run("Lay on Hands No Pool", func(t *testing.T) {
		// Reset state
		paladin.StateManager.ActionUsedCount = 0
		ally.StateManager.CurrentHP = 10
		paladin.StateManager.Resource[string(core.SpecAbilityLayOnHands)] = 0

		err := ed.HandleLayOnHands(paladin, paladin.Features[0], core.HookOnTurnStart, nil)
		if err != nil {
			t.Fatalf("HandleLayOnHands failed: %v", err)
		}

		if ally.StateManager.CurrentHP != 10 {
			t.Errorf("Expected Ally to have 10 HP, got %d", ally.StateManager.CurrentHP)
		}

		if paladin.StateManager.ActionUsedCount != 0 {
			t.Error("Expected ActionUsedCount to be 0")
		}
	})

	t.Run("Lay on Hands Already Used Action", func(t *testing.T) {
		// Reset state
		paladin.StateManager.ActionUsedCount = 1
		ally.StateManager.CurrentHP = 10
		paladin.StateManager.Resource[string(core.SpecAbilityLayOnHands)] = 25

		err := ed.HandleLayOnHands(paladin, paladin.Features[0], core.HookOnTurnStart, nil)
		if err != nil {
			t.Fatalf("HandleLayOnHands failed: %v", err)
		}

		if ally.StateManager.CurrentHP != 10 {
			t.Errorf("Expected Ally to have 10 HP, got %d", ally.StateManager.CurrentHP)
		}
	})
}
