package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleDeflectMissiles(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	monk := &actor.Actor{
		InstanceID: 1,
		Name:       "Monk",
		Metadata: actor.Metadata{
			Level: 5,
		},
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{Dexterity: 16}, // +3 modifier
		},
		StateManager: state_manager.StateManager{
			CurrentHP:       50,
			OncePerTurnUsed: make(map[string]bool),
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityDeflectMissiles,
				Data: core.FeatureData{
					NumberOfDice: 1,
					Die:          core.D10,
				},
				Hooks: map[core.HookType]bool{
					core.HookOnSelfDamageTaken: true,
				},
			},
		},
	}

	testCases := []struct {
		name       string
		actionType core.ActionType
		damage     int
		expectUse  bool
	}{
		{
			name:       "Ranged Attack Reduction",
			actionType: core.ATRanged,
			damage:     20,
			expectUse:  true,
		},
		{
			name:       "Melee Attack No Reduction",
			actionType: core.ATMelee,
			damage:     20,
			expectUse:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state
			monk.StateManager.ReactionUsedCount = 0
			damageVal := tc.damage

			ctx := &FeatureContext{
				DamageContext: &DamageContext{
					DamageValue: &damageVal,
				},
				AttackContext: &AttackContext{
					Action: &core.Action{
						ActionType: tc.actionType,
					},
				},
			}

			err := ed.HandleDeflectMissiles(monk, monk.Features[0], core.HookOnSelfDamageTaken, ctx)
			if err != nil {
				t.Fatalf("HandleDeflectMissiles failed: %v", err)
			}

			if tc.expectUse {
				if monk.StateManager.ReactionUsedCount == 0 {
					t.Error("Expected ReactionUsedCount to be incremented")
				}
				// 1d10 (min 1, max 10) + 3 (Dex) + 5 (Level) = 9 to 18 reduction
				expectedMinDamage := tc.damage - 18
				expectedMaxDamage := tc.damage - 9
				if damageVal < expectedMinDamage || damageVal > expectedMaxDamage {
					t.Errorf("Damage %d outside expected range [%d, %d]", damageVal, expectedMinDamage, expectedMaxDamage)
				}
			} else {
				if monk.StateManager.ReactionUsedCount > 0 {
					t.Error("Expected ReactionUsedCount to be 0")
				}
				if damageVal != tc.damage {
					t.Errorf("Expected damage to remain %d, got %d", tc.damage, damageVal)
				}
			}
		})
	}
}

func TestHandleDeflectMissiles_OncePerTurn(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{})

	monk := &actor.Actor{
		InstanceID: 1,
		Metadata:   actor.Metadata{Level: 5},
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{Dexterity: 16},
		},
		StateManager: state_manager.StateManager{
			ReactionUsedCount: 1,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityDeflectMissiles,
				Data: core.FeatureData{NumberOfDice: 1, Die: core.D10},
			},
		},
	}

	damageVal := 20
	ctx := &FeatureContext{
		DamageContext: &DamageContext{DamageValue: &damageVal},
		AttackContext: &AttackContext{
			Action: &core.Action{ActionType: core.ATRanged},
		},
	}

	ed.HandleDeflectMissiles(monk, monk.Features[0], core.HookOnSelfDamageTaken, ctx)

	if damageVal != 20 {
		t.Errorf("Expected no reduction when already used, got %d", damageVal)
	}
}
