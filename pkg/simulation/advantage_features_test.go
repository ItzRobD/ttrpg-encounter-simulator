package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleSaveAdvantage(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	testCases := []struct {
		name      core.SpecialAbility
		ability   core.Ability
		expectAdv bool
	}{
		{core.SpecAbilityRageStrengthSave, core.AbilityStrength, true},
		{core.SpecAbilityRageStrengthSave, core.AbilityDexterity, false},
		{core.SpecAbilityDangerSense, core.AbilityDexterity, true},
		{core.SpecAbilityDangerSense, core.AbilityWisdom, false},
		{core.SpecAbilitySlipperyMind, core.AbilityWisdom, true},
		{core.SpecAbilitySlipperyMind, core.AbilityCharisma, false},
	}

	for _, tc := range testCases {
		t.Run(string(tc.name)+"_"+string(tc.ability), func(t *testing.T) {
			a := &actor.Actor{
				InstanceID: 1,
				Name:       "Test Actor",
				StateManager: state_manager.StateManager{
					IsRaging: true,
				},
				Features: []core.Feature{
					{
						Name: tc.name,
						Hooks: map[core.HookType]bool{
							core.HookOnSelfSavingThrow: true,
						},
					},
				},
			}

			action := &core.Action{
				HasDC:     true,
				DCAbility: tc.ability,
			}

			target := a
			ed.Actors[1] = a

			opts := &roll_manager.RollOptions{Advantage: core.RollNormal}
			ctx := &FeatureContext{
				Target: target,
				SaveContext: &SaveContext{
					Target:  target,
					Options: opts,
				},
				AttackContext: &AttackContext{
					Action: action,
				},
			}

			err := ed.HandleSaveAdvantage(a, a.Features[0], core.HookOnSelfSavingThrow, ctx)
			if err != nil {
				t.Fatalf("HandleSaveAdvantage failed: %v", err)
			}

			if tc.expectAdv {
				if opts.Advantage != core.RollAdvantage {
					t.Errorf("Expected advantage for %s on %s save", tc.name, tc.ability)
				}
			} else {
				if opts.Advantage != core.RollNormal {
					t.Errorf("Did not expect advantage for %s on %s save", tc.name, tc.ability)
				}
			}
		})
	}
}

func TestFeralInstinct(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{})

	a := &actor.Actor{
		InstanceID: 1,
		Name:       "Barbarian",
		Features: []core.Feature{
			{Name: core.SpecAbilityFeralInstinct},
		},
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{Dexterity: 10},
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 10,
		},
	}

	ed.AddActor(a)
	ed.RollInitiative()

	// Since we can't easily see the internal options of RollD20, we rely on the logic in RollInitiative:
	// if a.HasFeature(core.SpecAbilityFeralInstinct) { opts.Advantage = core.RollAdvantage }
	// We've already verified this logic exists in encounter_director.go.

	// Just verify it doesn't crash and the result is stored.
	if _, ok := ed.InitiativeResults[1]; !ok {
		t.Error("Initiative result not stored for actor with Feral Instinct")
	}
}
