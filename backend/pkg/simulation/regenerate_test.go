package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestRegenerateSpell(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	caster := &actor.Actor{
		InstanceID: 1,
		Name:       "Cleric",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Fighter",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP:   0,
			MaxHP:       50,
			HealthState: core.HealthStateUnconscious,
			Conditions:  core.NewActorConditions(),
		},
	}
	target.StateManager.Conditions.Add(core.ConditionUnconscious)
	target.StateManager.Conditions.Add(core.ConditionProne)

	ed.Actors[1] = caster
	ed.Actors[2] = target
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	regenerateAction := core.Action{
		Name:        "Regenerate",
		ActionType:  core.ATSpell,
		AverageHeal: 0, // The initial healing is 4d8 + 15, but let's assume 0 for testing the regen part
	}

	t.Run("Regenerate spell adds feature", func(t *testing.T) {
		err := ed.Adjudicator.executeHealing(caster, target.InstanceID, &regenerateAction, ed)
		if err != nil {
			t.Fatalf("executeHealing failed: %v", err)
		}

		found := false
		for _, f := range target.Features {
			if f.Name == core.SpecAbilityRegeneration {
				found = true
				if f.Data.Value != 1 {
					t.Errorf("Expected regeneration value 1, got %d", f.Data.Value)
				}
				if !f.Hooks[core.HookOnTurnStart] {
					t.Errorf("Expected HookOnTurnStart to be true")
				}
			}
		}
		if !found {
			t.Errorf("Regeneration feature not added to target")
		}
	})

	t.Run("Regeneration restores consciousness at turn start", func(t *testing.T) {
		// Ensure target is at 0 HP and unconscious
		target.StateManager.CurrentHP = 0
		target.StateManager.HealthState = core.HealthStateUnconscious
		target.StateManager.Conditions.Add(core.ConditionUnconscious)
		target.StateManager.Conditions.Add(core.ConditionProne)

		// Process turn start
		ed.processTurnStart(target)

		if target.StateManager.CurrentHP != 1 {
			t.Errorf("Expected 1 HP after regeneration, got %d", target.StateManager.CurrentHP)
		}

		if target.StateManager.Conditions.Has(core.ConditionUnconscious) {
			t.Errorf("Target should no longer be unconscious")
		}

		if target.StateManager.Conditions.Has(core.ConditionProne) {
			t.Errorf("Target should no longer be prone")
		}

		if target.StateManager.HealthState == core.HealthStateUnconscious {
			t.Errorf("HealthState should not be unconscious")
		}
	})
}
