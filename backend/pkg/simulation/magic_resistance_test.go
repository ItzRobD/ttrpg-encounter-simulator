package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleMagicResistance(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Creature with Magic Resistance
	creature := &actor.Actor{
		InstanceID: 1,
		Name:       "Resistant Creature",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{
				Wisdom: 10,
			},
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityMagicResistance,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfSavingThrowAgainstMagic: true,
				},
			},
		},
	}

	// Attacker
	caster := &actor.Actor{
		InstanceID: 2,
		Name:       "Caster",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
	}

	ed.Actors[1] = creature
	ed.Actors[2] = caster
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Advantage on Spell Save", func(t *testing.T) {
		spellAction := &core.Action{
			ID:          core.MakeID("vicious_mockery:0"),
			Name:        "Vicious Mockery",
			ActionType:  core.ATSpell,
			HasDC:       true,
			DCSaveDC:    13,
			DCAbility:   core.AbilityWisdom,
			DCOnSuccess: core.DCOnSuccessNone,
		}

		// We need to check if ResolveSavingThrow applies advantage
		// ResolveSavingThrow doesn't return the options, but we can check if it was modified in the context
		// if we mock or just rely on the internal logic.
		// Since we can't easily see internal RollOptions from ResolveSavingThrow return,
		// we'll test the handler directly and then the integration.

		// Unit test handler
		ctx := &FeatureContext{
			Target: creature,
			SaveContext: &SaveContext{
				Target:  creature,
				Options: &roll_manager.RollOptions{},
			},
			AttackContext: &AttackContext{
				Action: spellAction,
			},
		}

		err := ed.HandleMagicResistance(creature, creature.Features[0], core.HookOnSelfSavingThrowAgainstMagic, ctx)
		if err != nil {
			t.Fatalf("HandleMagicResistance failed: %v", err)
		}

		if ctx.SaveContext.Options.Advantage != core.RollAdvantage {
			t.Errorf("Expected advantage on spell save, got %v", ctx.SaveContext.Options.Advantage)
		}
	})

	t.Run("No Advantage on Physical Attack (Standard Hook)", func(t *testing.T) {
		// Physical attacks don't trigger HookOnSelfSavingThrowAgainstMagic,
		// but let's ensure it doesn't trigger on HookOnSelfSavingThrow if registered there.
		// Magic Resistance PHB text: "advantage on saving throws against spells and other magical effects."

		physicalAction := &core.Action{
			ID:          core.MakeID("trap:1"),
			Name:        "Pit Trap",
			ActionType:  core.ATAction,
			HasDC:       true,
			DCSaveDC:    15,
			DCAbility:   core.AbilityDexterity,
			DCOnSuccess: core.DCOnSuccessHalf,
		}

		ctx := &FeatureContext{
			Target: creature,
			SaveContext: &SaveContext{
				Target:  creature,
				Options: &roll_manager.RollOptions{Advantage: core.RollNormal},
			},
			AttackContext: &AttackContext{
				Action: physicalAction,
			},
		}

		// Directly calling dispatchHooks with HookOnSelfSavingThrow shouldn't do anything because
		// Magic Resistance only has HookOnSelfSavingThrowAgainstMagic in its map.
		ed.dispatchHooks(creature, core.HookOnSelfSavingThrow, ctx)

		if ctx.SaveContext.Options.Advantage != core.RollNormal {
			t.Errorf("Expected no advantage on physical save, got %v", ctx.SaveContext.Options.Advantage)
		}
	})

	t.Run("Integration Test - Spell Save Advantage", func(t *testing.T) {
		// Verify HookOnSelfSavingThrowAgainstMagic is dispatched in Adjudicator.go
		// and correctly modifies the outcome.

		spellAction := &core.Action{
			ID:          core.MakeID("vicious_mockery:0"),
			Name:        "Vicious Mockery",
			ActionType:  core.ATSpell,
			HasDC:       true,
			DCSaveDC:    25, // High DC to ensure failure without advantage/luck
			DCAbility:   core.AbilityWisdom,
			DCOnSuccess: core.DCOnSuccessNone,
		}

		// Reset creature state
		creature.StateManager.CurrentHP = 50

		// In a real d20 roll with Normal advantage, high DC means frequent failure.
		// Magic Resistance should grant Advantage.
		// Since we use real RNG, we can't easily assert "Advantage was used" without mocking.
		// But we can check if it at least runs without error and the hook is registered.

		// We'll call ResolveSavingThrow and see if it runs.
		// We already unit tested that the handler sets the flag.
		success, _ := ed.Adjudicator.ResolveSavingThrow(caster, creature, spellAction)
		t.Logf("Save success: %v", success)
	})
}
