package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleLegendaryResistance(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Dragon with Legendary Resistance
	dragon := &actor.Actor{
		InstanceID: 1,
		Name:       "Ancient Dragon",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{
				Wisdom: 10, // Mod 0
			},
		},
		StateManager: state_manager.StateManager{
			Resource: make(map[string]int),
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityLegendaryResistance,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfSavingThrow: true,
				},
				Data: core.FeatureData{
					Value: 3, // 3 uses
				},
			},
		},
	}
	dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] = 3

	// Caster (Wizard)
	wizard := &actor.Actor{
		InstanceID: 2,
		Name:       "Wizard",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
	}

	ed.Actors[1] = dragon
	ed.Actors[2] = wizard

	// Polymorph-ish Action (Wis Save)
	polymorph := core.Action{
		Name:        "Polymorph",
		ActionType:  core.ATSpell,
		HasDC:       true,
		DCSaveDC:    25, // Unbeatable for this dragon
		DCAbility:   core.AbilityWisdom,
		DCOnSuccess: core.DCOnSuccessNone,
	}

	t.Run("Legendary Resistance turns failure into success", func(t *testing.T) {
		// Reset state
		dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] = 3

		// Inlined ResolveSavingThrow logic basically
		saveSuccess, _ := ed.Adjudicator.ResolveSavingThrow(wizard, dragon, &polymorph)

		// It should be true because Legendary Resistance kicked in
		if !saveSuccess {
			t.Error("Expected saveSuccess to be true after Legendary Resistance")
		}

		if dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] != 2 {
			t.Errorf("Expected 2 uses left, got %d", dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)])
		}
	})

	t.Run("Legendary Resistance does not trigger on already successful save", func(t *testing.T) {
		// Easy DC
		easyAction := polymorph
		easyAction.DCSaveDC = 0

		dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] = 3

		saveSuccess, _ := ed.Adjudicator.ResolveSavingThrow(wizard, dragon, &easyAction)

		if !saveSuccess {
			t.Error("Expected saveSuccess to be true")
		}

		if dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] != 3 {
			t.Errorf("Expected 3 uses left (not consumed on success), got %d", dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)])
		}
	})

	t.Run("Legendary Resistance does not trigger when out of uses", func(t *testing.T) {
		dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] = 0

		// Hard DC
		saveSuccess, _ := ed.Adjudicator.ResolveSavingThrow(wizard, dragon, &polymorph)

		if saveSuccess {
			t.Error("Expected saveSuccess to be false when out of Legendary Resistance uses")
		}

		if dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] != 0 {
			t.Errorf("Expected 0 uses left, got %d", dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)])
		}
	})

	t.Run("Legendary Resistance handles multiple uses correctly", func(t *testing.T) {
		dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] = 2

		// Use 1
		ed.Adjudicator.ResolveSavingThrow(wizard, dragon, &polymorph)
		if dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] != 1 {
			t.Errorf("Expected 1 use left, got %d", dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)])
		}

		// Use 2
		ed.Adjudicator.ResolveSavingThrow(wizard, dragon, &polymorph)
		if dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] != 0 {
			t.Errorf("Expected 0 uses left, got %d", dragon.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)])
		}

		// Use 3 (Failed)
		saveSuccess, _ := ed.Adjudicator.ResolveSavingThrow(wizard, dragon, &polymorph)
		if saveSuccess {
			t.Error("Expected saveSuccess to be false")
		}
	})
}
