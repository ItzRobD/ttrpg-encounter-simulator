package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleBrutalCritical(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{EnableSpecialAbilities: true})

	barb := &actor.Actor{
		InstanceID: 1,
		Name:       "Barbarian",
		ActorType:  core.ActorTypeCharacter,
		Features: []core.Feature{
			{
				Name: core.SpecAbilityBrutalCritical,
				Data: core.FeatureData{
					NumberOfDice: 2,
					Die:          core.D12,
				},
				Hooks: map[core.HookType]bool{core.HookOnSelfHit: true},
			},
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
		},
	}

	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Target",
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
		},
	}

	ed.Actors[1] = barb
	ed.Actors[2] = target
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Trigger on Critical Melee Hit", func(t *testing.T) {
		target.StateManager.CurrentHP = 100
		isCrit := true
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				IsCritical: &isCrit,
				Action: &core.Action{
					ActionType: core.ATMelee,
					DiceBlock: []core.DiceBlock{
						{NumberOfDice: 1, Die: core.D12, DamageType: core.DamageSlashing},
					},
				},
			},
		}

		err := ed.HandleBrutalCritical(barb, barb.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleBrutalCritical failed: %v", err)
		}

		// Initial HP 100. resolveDamage handles the damage application.
		// Since resolveDamage is called with the feature action, it should have dealt 2d12 damage.
		// 100 - 2d12.
		if target.StateManager.CurrentHP >= 100 {
			t.Errorf("Expected damage to be dealt, got %d", target.StateManager.CurrentHP)
		}

		damageDealt := 100 - target.StateManager.CurrentHP
		// 2d12 is 2 to 24 damage
		if damageDealt < 2 || damageDealt > 24 {
			t.Errorf("Damage %d outside expected range [2, 24]", damageDealt)
		}
	})

	t.Run("No Trigger on Normal Hit", func(t *testing.T) {
		target.StateManager.CurrentHP = 100
		isCrit := false
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				IsCritical: &isCrit,
				Action: &core.Action{
					ActionType: core.ATMelee,
					DiceBlock: []core.DiceBlock{
						{NumberOfDice: 1, Die: core.D12, DamageType: core.DamageSlashing},
					},
				},
			},
		}

		err := ed.HandleBrutalCritical(barb, barb.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleBrutalCritical failed: %v", err)
		}

		if target.StateManager.CurrentHP != 100 {
			t.Errorf("Expected no damage on normal hit, got %d", target.StateManager.CurrentHP)
		}
	})

	t.Run("No Trigger on Ranged Critical Hit", func(t *testing.T) {
		target.StateManager.CurrentHP = 100
		isCrit := true
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				IsCritical: &isCrit,
				Action: &core.Action{
					ActionType: core.ATRanged,
					DiceBlock: []core.DiceBlock{
						{NumberOfDice: 1, Die: core.D12, DamageType: core.DamageSlashing},
					},
				},
			},
		}

		err := ed.HandleBrutalCritical(barb, barb.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleBrutalCritical failed: %v", err)
		}

		if target.StateManager.CurrentHP != 100 {
			t.Errorf("Expected no damage on ranged critical hit, got %d", target.StateManager.CurrentHP)
		}
	})
}
