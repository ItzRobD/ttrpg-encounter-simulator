package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleImprovedDivineSmite(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	paladin := &actor.Actor{
		InstanceID: 1,
		Name:       "Paladin",
		Features: []core.Feature{
			{
				Name: core.SpecAbilityImprovedDivineSmite,
				Data: core.FeatureData{
					NumberOfDice: 1,
					Die:          core.D8,
				},
				Hooks: map[core.HookType]bool{
					core.HookOnSelfHit: true,
				},
			},
		},
	}

	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Goblin",
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	ed.Actors[1] = paladin
	ed.Actors[2] = target
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	isCrit := false
	ctx := &FeatureContext{
		Target: target,
		AttackContext: &AttackContext{
			Action: &core.Action{
				ActionType: core.ATMelee,
			},
			IsCritical: &isCrit,
		},
		DamageContext: &DamageContext{
			DamageType: core.DamageRadiant,
		},
	}

	t.Run("Improved Divine Smite adds damage on melee hit", func(t *testing.T) {
		target.StateManager.CurrentHP = 50
		err := ed.HandleImprovedDivineSmite(paladin, paladin.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleImprovedDivineSmite failed: %v", err)
		}

		if target.StateManager.CurrentHP >= 50 {
			t.Error("Expected target to take damage from Improved Divine Smite")
		}
	})

	t.Run("Improved Divine Smite does not trigger on non-melee", func(t *testing.T) {
		target.StateManager.CurrentHP = 50
		ctx.AttackContext.Action.ActionType = core.ATRanged
		err := ed.HandleImprovedDivineSmite(paladin, paladin.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleImprovedDivineSmite failed: %v", err)
		}

		if target.StateManager.CurrentHP != 50 {
			t.Error("Expected target to NOT take damage on non-melee attack")
		}
	})
}

func TestHandleHalflingLucky(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	halfling := &actor.Actor{
		InstanceID: 1,
		Features: []core.Feature{
			{
				Name: core.SpecAbilityHalflingLucky,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfAttack:      true,
					core.HookOnSelfSavingThrow: true,
				},
			},
		},
	}

	t.Run("Lucky sets RerollThreshold on AttackRoll", func(t *testing.T) {
		opts := &roll_manager.RollOptions{RerollThreshold: 0}
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				AttackRoll: opts,
			},
		}

		err := ed.HandleHalflingLucky(halfling, halfling.Features[0], core.HookOnSelfAttack, ctx)
		if err != nil {
			t.Fatalf("HandleHalflingLucky failed: %v", err)
		}

		if opts.RerollThreshold != 1 {
			t.Errorf("Expected RerollThreshold 1, got %d", opts.RerollThreshold)
		}
	})

	t.Run("Lucky sets RerollThreshold on SavingThrow", func(t *testing.T) {
		opts := &roll_manager.RollOptions{RerollThreshold: 0}
		ctx := &FeatureContext{
			SaveContext: &SaveContext{
				Options: opts,
			},
		}

		err := ed.HandleHalflingLucky(halfling, halfling.Features[0], core.HookOnSelfSavingThrow, ctx)
		if err != nil {
			t.Fatalf("HandleHalflingLucky failed: %v", err)
		}

		if opts.RerollThreshold != 1 {
			t.Errorf("Expected RerollThreshold 1, got %d", opts.RerollThreshold)
		}
	})
}

func TestHandleDwarvenResilience(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	dwarf := &actor.Actor{
		InstanceID: 1,
		Features: []core.Feature{
			{
				Name: core.SpecAbilityDwarvenResilience,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfSavingThrow: true,
				},
			},
		},
	}

	t.Run("Dwarven Resilience grants advantage against poison", func(t *testing.T) {
		opts := &roll_manager.RollOptions{Advantage: core.RollNormal}
		ctx := &FeatureContext{
			SaveContext: &SaveContext{
				Options: opts,
			},
			AttackContext: &AttackContext{
				Action: &core.Action{
					HasDC: true,
					DiceBlock: []core.DiceBlock{
						{DamageType: core.DamagePoison},
					},
				},
			},
		}

		err := ed.HandleSaveAdvantage(dwarf, dwarf.Features[0], core.HookOnSelfSavingThrow, ctx)
		if err != nil {
			t.Fatalf("HandleSaveAdvantage failed: %v", err)
		}

		if opts.Advantage != core.RollAdvantage {
			t.Error("Expected advantage against poison save")
		}
	})

	t.Run("Dwarven Resilience handles missing AttackContext gracefully", func(t *testing.T) {
		opts := &roll_manager.RollOptions{Advantage: core.RollNormal}
		ctx := &FeatureContext{
			SaveContext: &SaveContext{
				Options: opts,
			},
			// AttackContext is nil
		}

		err := ed.HandleSaveAdvantage(dwarf, dwarf.Features[0], core.HookOnSelfSavingThrow, ctx)
		if err != nil {
			t.Fatalf("HandleSaveAdvantage failed on nil AttackContext: %v", err)
		}

		if opts.Advantage != core.RollNormal {
			t.Error("Did not expect advantage when AttackContext is missing")
		}
	})
}
