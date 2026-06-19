package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleSavageAttacks(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	orc := &actor.Actor{
		InstanceID: 1,
		Name:       "Orc",
		Metadata: actor.Metadata{
			Level: 1,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilitySavageAttacks,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfHit: true,
				},
			},
		},
		StateManager: state_manager.StateManager{
			OncePerTurnUsed: make(map[string]bool),
		},
	}

	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Target",
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	ed.Actors[1] = orc
	ed.Actors[2] = target
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Savage Attacks on Critical Melee Hit", func(t *testing.T) {
		target.StateManager.CurrentHP = 50
		isCritical := true

		// Greataxe: 1d12
		action := &core.Action{
			Name:       "Greataxe",
			ActionType: core.ATMelee,
			DiceBlock: []core.DiceBlock{
				{
					NumberOfDice: 1,
					Die:          core.D12,
					DamageType:   core.DamageSlashing,
				},
			},
		}

		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				Action:     action,
				IsCritical: &isCritical,
			},
		}

		// Initial HP: 50
		// Greataxe crit: 2d12. Average 13.
		// Savage attack: +1d12. Average 6.5.
		// We want to verify that the savage attack handler triggers and calls resolveDamage.

		err := ed.HandleSavageAttacks(orc, orc.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleSavageAttacks failed: %v", err)
		}

		// Damage should be > 0
		damageDealt := 50 - target.StateManager.CurrentHP
		if damageDealt <= 0 {
			t.Errorf("Expected damage to be dealt, got %d", damageDealt)
		}

		// 1d12 should be between 1 and 12
		if damageDealt < 1 || damageDealt > 12 {
			t.Errorf("Savage attack damage %d outside expected range [1, 12]", damageDealt)
		}
	})

	t.Run("Savage Attacks NOT on Normal Hit", func(t *testing.T) {
		target.StateManager.CurrentHP = 50
		isCritical := false

		action := &core.Action{
			Name:       "Greataxe",
			ActionType: core.ATMelee,
			DiceBlock: []core.DiceBlock{
				{
					NumberOfDice: 1,
					Die:          core.D12,
					DamageType:   core.DamageSlashing,
				},
			},
		}

		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				Action:     action,
				IsCritical: &isCritical,
			},
		}

		err := ed.HandleSavageAttacks(orc, orc.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleSavageAttacks failed: %v", err)
		}

		if target.StateManager.CurrentHP != 50 {
			t.Errorf("Expected no damage on normal hit, got %d damage", 50-target.StateManager.CurrentHP)
		}
	})

	t.Run("Savage Attacks NOT on Ranged Critical Hit", func(t *testing.T) {
		target.StateManager.CurrentHP = 50
		isCritical := true

		action := &core.Action{
			Name:       "Longbow",
			ActionType: core.ATRanged,
			DiceBlock: []core.DiceBlock{
				{
					NumberOfDice: 1,
					Die:          core.D8,
					DamageType:   core.DamagePiercing,
				},
			},
		}

		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				Action:     action,
				IsCritical: &isCritical,
			},
		}

		err := ed.HandleSavageAttacks(orc, orc.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleSavageAttacks failed: %v", err)
		}

		if target.StateManager.CurrentHP != 50 {
			t.Errorf("Expected no damage on ranged hit, got %d damage", 50-target.StateManager.CurrentHP)
		}
	})
}
