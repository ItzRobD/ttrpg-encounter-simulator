package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleElusive(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{EnableSpecialAbilities: true})

	rogue := &actor.Actor{
		InstanceID: 1,
		Name:       "Rogue",
		StateManager: state_manager.StateManager{
			Conditions: core.NewActorConditions(),
		},
		Features: []core.Feature{
			{
				Name:  core.SpecAbilityElusive,
				Hooks: map[core.HookType]bool{core.HookOnSelfAttack: true},
			},
		},
	}

	attacker := &actor.Actor{
		InstanceID: 2,
		Name:       "Enemy",
	}

	ed.Actors[1] = rogue
	ed.Actors[2] = attacker

	t.Run("Elusive negates advantage", func(t *testing.T) {
		opts := &roll_manager.RollOptions{Advantage: core.RollAdvantage}
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				AttackRoll: opts,
			},
		}

		err := ed.HandleElusive(rogue, rogue.Features[0], core.HookOnSelfAttack, ctx)
		if err != nil {
			t.Fatalf("HandleElusive failed: %v", err)
		}

		if opts.Advantage != core.RollNormal {
			t.Errorf("Expected RollNormal, got %v", opts.Advantage)
		}
	})

	t.Run("Elusive doesn't work when incapacitated", func(t *testing.T) {
		rogue.StateManager.Conditions.Add(core.ConditionIncapacitated)
		defer rogue.StateManager.Conditions.Remove(core.ConditionIncapacitated)

		opts := &roll_manager.RollOptions{Advantage: core.RollAdvantage}
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				AttackRoll: opts,
			},
		}

		err := ed.HandleElusive(rogue, rogue.Features[0], core.HookOnSelfAttack, ctx)
		if err != nil {
			t.Fatalf("HandleElusive failed: %v", err)
		}

		if opts.Advantage != core.RollAdvantage {
			t.Errorf("Expected RollAdvantage, got %v", opts.Advantage)
		}
	})
}
