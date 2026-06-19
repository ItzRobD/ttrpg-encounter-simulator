package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleSecondWind(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	fighter := &actor.Actor{
		InstanceID: 1,
		Name:       "Fighter",
		Metadata: actor.Metadata{
			Level: 10,
		},
		HPConfig: core.HPConfig{
			HPAverage: 100,
		},
		StateManager: state_manager.StateManager{
			CurrentHP:   10,
			MaxHP:       100,
			HealthState: core.HealthStateHealthy,
			Resource:    make(map[string]int),
		},
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{
				Dexterity: 10,
			},
		},
	}
	fighter.StateManager.Resource[string(core.SpecAbilitySecondWind)] = 1
	ed.AddActor(fighter)
	ed.SetupEncounter()

	f := core.Feature{
		Name: core.SpecAbilitySecondWind,
		Data: core.FeatureData{
			NumberOfDice: 1,
			Die:          core.D10,
		},
	}

	t.Run("Successful Second Wind", func(t *testing.T) {
		fighter.StateManager.CurrentHP = 10
		fighter.StateManager.BonusActionUsedCount = 0

		err := ed.HandleSecondWind(fighter, f, core.HookOnTurnStart, nil)
		if err != nil {
			t.Fatalf("HandleSecondWind failed: %v", err)
		}

		// 1d10 + level 10 => min 11, max 20
		if fighter.StateManager.CurrentHP < 11 || fighter.StateManager.CurrentHP > 20+10 {
			t.Errorf("Expected HP between 21 and 30, got %d", fighter.StateManager.CurrentHP)
		}

		if fighter.StateManager.BonusActionUsedCount != 1 {
			t.Errorf("Expected BonusActionUsedCount to be 1, got %d", fighter.StateManager.BonusActionUsedCount)
		}
	})

	t.Run("Already used Bonus Action", func(t *testing.T) {
		fighter.StateManager.CurrentHP = 10
		fighter.StateManager.BonusActionUsedCount = 1

		err := ed.HandleSecondWind(fighter, f, core.HookOnTurnStart, nil)
		if err != nil {
			t.Fatalf("HandleSecondWind failed: %v", err)
		}

		if fighter.StateManager.CurrentHP != 10 {
			t.Errorf("Expected HP to remain 10, got %d", fighter.StateManager.CurrentHP)
		}
	})

	t.Run("Invalid Hook", func(t *testing.T) {
		fighter.StateManager.CurrentHP = 10
		fighter.StateManager.BonusActionUsedCount = 0

		err := ed.HandleSecondWind(fighter, f, core.HookOnSelfHit, nil)
		if err != nil {
			t.Fatalf("HandleSecondWind failed: %v", err)
		}

		if fighter.StateManager.CurrentHP != 10 {
			t.Errorf("Expected HP to remain 10, got %d", fighter.StateManager.CurrentHP)
		}
	})
}
