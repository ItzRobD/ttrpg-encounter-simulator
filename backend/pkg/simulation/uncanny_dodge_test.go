package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleUncannyDodge(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{EnableSpecialAbilities: true})

	rogue := &actor.Actor{
		InstanceID: 1,
		Name:       "Rogue",
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
		Features: []core.Feature{
			{
				Name:  core.SpecAbilityUncannyDodge,
				Hooks: map[core.HookType]bool{core.HookOnSelfDamageTaken: true},
			},
		},
	}

	attacker := &actor.Actor{
		InstanceID: 2,
		Name:       "Enemy",
	}

	ed.Actors[1] = rogue
	ed.Actors[2] = attacker
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Uncanny Dodge halves weapon damage", func(t *testing.T) {
		rogue.StateManager.CurrentHP = 50
		rogue.StateManager.ReactionUsedCount = 0

		damageVal := 20
		ctx := &FeatureContext{
			DamageContext: &DamageContext{
				DamageValue: &damageVal,
			},
			AttackContext: &AttackContext{
				Action: &core.Action{ActionType: core.ATMelee},
			},
		}

		err := ed.HandleUncannyDodge(rogue, rogue.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleUncannyDodge failed: %v", err)
		}

		if damageVal != 10 {
			t.Errorf("Expected 10 damage, got %d", damageVal)
		}
		if rogue.StateManager.ReactionUsedCount != 1 {
			t.Error("Expected ReactionUsedCount to be incremented")
		}
	})

	t.Run("Uncanny Dodge doesn't work without reaction", func(t *testing.T) {
		rogue.StateManager.CurrentHP = 50
		rogue.StateManager.ReactionUsedCount = 1

		damageVal := 20
		ctx := &FeatureContext{
			DamageContext: &DamageContext{
				DamageValue: &damageVal,
			},
			AttackContext: &AttackContext{
				Action: &core.Action{ActionType: core.ATMelee},
			},
		}

		err := ed.HandleUncannyDodge(rogue, rogue.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleUncannyDodge failed: %v", err)
		}

		if damageVal != 20 {
			t.Errorf("Expected 20 damage, got %d", damageVal)
		}
	})

	t.Run("Uncanny Dodge doesn't work on saving throw based actions", func(t *testing.T) {
		rogue.StateManager.CurrentHP = 50
		rogue.StateManager.ReactionUsedCount = 0

		damageVal := 20
		ctx := &FeatureContext{
			DamageContext: &DamageContext{
				DamageValue: &damageVal,
			},
			AttackContext: &AttackContext{
				Action: &core.Action{
					ActionType: core.ATSpell,
					HasDC:      true,
				},
			},
		}

		err := ed.HandleUncannyDodge(rogue, rogue.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleUncannyDodge failed: %v", err)
		}

		if damageVal != 20 {
			t.Errorf("Expected 20 damage, got %d", damageVal)
		}
	})
}
