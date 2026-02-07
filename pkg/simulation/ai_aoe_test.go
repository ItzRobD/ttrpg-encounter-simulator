package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"math/rand/v2"
	"testing"
)

func TestAIDirector_AOE_Heuristics(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	aid := NewAIDirector(rng)
	simOptions := &core.SimulationOptions{
		HPVisibilityMode:   core.HPVisible,
		AOETargetThreshold: 2,
		AOEHitsAllEnemies:  true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	attacker := &actor.Actor{
		InstanceID: 1,
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			Resource: make(map[string]int),
		},
	}

	target1 := &actor.Actor{
		InstanceID: 2,
		Name:       "Target 1",
		Side:       core.SideCharacters,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	target2 := &actor.Actor{
		InstanceID: 3,
		Name:       "Target 2",
		Side:       core.SideCharacters,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	ed.Actors[1] = attacker
	ed.Actors[2] = target1
	ed.Actors[3] = target2

	// Setup actions: single target vs AOE
	singleAction := core.Action{
		ID:            core.MakeID("single"),
		Name:          "Single Strike",
		ActionType:    core.ATMelee,
		AverageDamage: 20,
		Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
	}

	aoeAction := core.Action{
		ID:            core.MakeID("aoe"),
		Name:          "Fireball",
		ActionType:    core.ATSpell,
		IsAOE:         true,
		AverageDamage: 15, // Lower than single target, but hits multiple
		Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
	}

	attacker.Actions = []core.Action{singleAction, aoeAction}

	t.Run("Highest Damage Policy chooses AOE when it hits enough targets", func(t *testing.T) {
		ed.SimOptions.ActionSelectionPolicy = core.ActionPolicyHighestDamage

		// Expected single damage: 20
		// Expected AOE damage: 15 * 2 = 30

		action, targetIDs := aid.chooseBestAttackAction(attacker, core.ActAction, ed)
		if action.Name != "Fireball" {
			t.Errorf("Expected Fireball (AOE), got %s", action.Name)
		}
		if len(targetIDs) != 2 {
			t.Errorf("Expected 2 targets for AOE, got %d", len(targetIDs))
		}
	})

	t.Run("Priority Policy chooses AOE when threshold met", func(t *testing.T) {
		ed.SimOptions.ActionSelectionPolicy = core.ActionPolicyPriority
		ed.SimOptions.AOETargetThreshold = 2

		action, targetIDs := aid.chooseBestAttackAction(attacker, core.ActAction, ed)
		if action.Name != "Fireball" {
			t.Errorf("Expected Fireball (AOE Priority), got %s", action.Name)
		}
		if len(targetIDs) != 2 {
			t.Errorf("Expected 2 targets for AOE, got %d", len(targetIDs))
		}
	})

	t.Run("Priority Policy ignores AOE when threshold NOT met", func(t *testing.T) {
		ed.SimOptions.ActionSelectionPolicy = core.ActionPolicyPriority
		ed.SimOptions.AOETargetThreshold = 3 // Only 2 targets available

		action, targetIDs := aid.chooseBestAttackAction(attacker, core.ActAction, ed)
		if action.Name == "Fireball" {
			t.Errorf("Did not expect Fireball when threshold not met")
		}
		if len(targetIDs) != 1 {
			t.Errorf("Expected 1 target for single target action, got %d", len(targetIDs))
		}
	})
}

func TestAdjudicator_AOE_Resolution(t *testing.T) {
	simOptions := &core.SimulationOptions{EnableSpecialAbilities: true}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	attacker := &actor.Actor{
		InstanceID: 1,
		Name:       "Attacker",
		Side:       core.SideMonsters,
	}
	target1 := &actor.Actor{
		InstanceID: 2,
		Name:       "Target 1",
		Side:       core.SideCharacters,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}
	target2 := &actor.Actor{
		InstanceID: 3,
		Name:       "Target 2",
		Side:       core.SideCharacters,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	ed.Actors[1] = attacker
	ed.Actors[2] = target1
	ed.Actors[3] = target2
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	aoeAction := core.Action{
		Name:          "Fireball",
		ActionType:    core.ATSpell,
		IsAOE:         true,
		AverageDamage: 10,
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 1, Die: core.D10, Modifier: 0, DamageType: core.DamageFire},
		},
		Cost: core.ActionCost{ActivationType: core.ActAction, Value: 1},
	}

	intent := ActionIntent{
		Action:         aoeAction,
		TargetIDs:      []int{2, 3},
		ActivationType: core.ActAction,
		ActorID:        1,
	}

	err := ed.Adjudicator.ResolveAction(attacker, intent)
	if err != nil {
		t.Fatalf("ResolveAction failed: %v", err)
	}

	if target1.StateManager.CurrentHP >= 50 {
		t.Errorf("Target 1 should have taken damage")
	}
	if target2.StateManager.CurrentHP >= 50 {
		t.Errorf("Target 2 should have taken damage")
	}
}
