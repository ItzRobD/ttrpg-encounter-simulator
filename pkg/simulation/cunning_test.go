package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleCunning(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Gnome with Gnome Cunning
	gnome := &actor.Actor{
		InstanceID: 1,
		Name:       "Gnome",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{
				Intelligence: 16,
				Wisdom:       14,
				Charisma:     12,
				Strength:     10,
				Dexterity:    10,
				Constitution: 10,
			},
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 10,
			MaxHP:     10,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityCunning,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfSavingThrowAgainstMagic: true,
				},
			},
		},
	}

	enemy := &actor.Actor{
		InstanceID: 2,
		Name:       "Mage",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
	}

	ed.Actors[1] = gnome
	ed.Actors[2] = enemy

	tests := []struct {
		name     string
		ability  core.Ability
		isSpell  bool
		expected core.AdvantageType
	}{
		{
			name:     "Intelligence save against spell",
			ability:  core.AbilityIntelligence,
			isSpell:  true,
			expected: core.RollAdvantage,
		},
		{
			name:     "Wisdom save against spell",
			ability:  core.AbilityWisdom,
			isSpell:  true,
			expected: core.RollAdvantage,
		},
		{
			name:     "Charisma save against spell",
			ability:  core.AbilityCharisma,
			isSpell:  true,
			expected: core.RollAdvantage,
		},
		{
			name:     "Intelligence save against non-spell",
			ability:  core.AbilityIntelligence,
			isSpell:  false,
			expected: core.RollNormal,
		},
		{
			name:     "Dexterity save against spell",
			ability:  core.AbilityDexterity,
			isSpell:  true,
			expected: core.RollNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actionType := core.ATAction
			if tt.isSpell {
				actionType = core.ATSpell
			}

			action := &core.Action{
				Name:       "Test Action",
				ActionType: actionType,
				HasDC:      true,
				DCAbility:  tt.ability,
				DCSaveDC:   15,
			}

			// In this test, we can't easily check the final result of the roll
			// because it's random, but we can check if the Advantage flag was set
			// in the options by manually triggering the hook.

			opts := roll_manager.RollOptions{
				RollType:  core.DiceRollSavingThrow,
				Advantage: core.RollNormal,
			}

			ctx := &FeatureContext{
				Target: gnome,
				SaveContext: &SaveContext{
					Target:  gnome,
					Options: &opts,
				},
				AttackContext: &AttackContext{
					Action: action,
				},
			}

			hook := core.HookOnSelfSavingThrowAgainstMagic
			if tt.isSpell {
				err := ed.HandleCunning(gnome, gnome.Features[0], hook, ctx)
				if err != nil {
					t.Fatalf("HandleCunning failed: %v", err)
				}
			}

			if opts.Advantage != tt.expected {
				t.Errorf("Expected advantage %v, got %v", tt.expected, opts.Advantage)
			}
		})
	}
}
