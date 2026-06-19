package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestActionSelectionPriority(t *testing.T) {
	seed := core.Seed{Seed1: 1, Seed2: 1}

	// Policy: Priority
	options := &core.SimulationOptions{
		ActionSelectionPolicy: core.ActionPolicyPriority,
	}
	ed := NewEncounterDirector(seed, options)

	monster := &actor.Actor{
		InstanceID: 1,
		Name:       "Dragon",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
			Resource:  make(map[string]int),
		},
		Actions: []core.Action{
			{
				ID:            core.MakeID("bite"),
				Name:          "Bite",
				ActionType:    core.ATMelee,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AverageDamage: 10,
			},
			{
				ID:            core.MakeID("multi"),
				Name:          "Multiattack",
				ActionType:    core.ATMultiAttack,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AverageDamage: 25,
			},
			{
				ID:            core.MakeID("breath"),
				Name:          "Breath Weapon",
				ActionType:    core.ATAction,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AverageDamage: 50,
				RechargeValue: 5,
			},
		},
	}

	target := CreateMockFighter(2, "Hero", core.SideCharacters)
	ed.AddActor(monster)
	ed.AddActor(target)
	ed.SetupEncounter()

	// 1. Recharge is available -> Should pick Breath Weapon
	monster.StateManager.Resource["Breath Weapon"] = 1
	action, _ := ed.AIDirector.chooseBestAttackAction(monster, core.ActAction, ed)
	if action == nil || action.Name != "Breath Weapon" {
		t.Errorf("Expected Breath Weapon, got %v", action)
	}

	// 2. Recharge NOT available -> Should pick Multiattack
	monster.StateManager.Resource["Breath Weapon"] = 0
	action, _ = ed.AIDirector.chooseBestAttackAction(monster, core.ActAction, ed)
	if action == nil || action.Name != "Multiattack" {
		t.Errorf("Expected Multiattack, got %v", action)
	}

	// 3. Neither Recharge nor Multiattack available -> Should pick Bite
	monster.Actions = []core.Action{monster.Actions[0]} // Only Bite remains
	action, _ = ed.AIDirector.chooseBestAttackAction(monster, core.ActAction, ed)
	if action == nil || action.Name != "Bite" {
		t.Errorf("Expected Bite, got %v", action)
	}
}

func TestActionSelectionHighestDamage(t *testing.T) {
	seed := core.Seed{Seed1: 1, Seed2: 1}

	// Policy: Highest Damage (Default)
	options := &core.SimulationOptions{
		ActionSelectionPolicy: core.ActionPolicyHighestDamage,
	}
	ed := NewEncounterDirector(seed, options)

	monster := &actor.Actor{
		InstanceID: 1,
		Name:       "Monster",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
			Resource:  make(map[string]int),
		},
		Actions: []core.Action{
			{
				ID:            core.MakeID("low"),
				Name:          "Low Damage",
				ActionType:    core.ATMelee,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AverageDamage: 10,
			},
			{
				ID:            core.MakeID("high"),
				Name:          "High Damage",
				ActionType:    core.ATMelee,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AverageDamage: 50,
			},
		},
	}

	target := CreateMockFighter(2, "Hero", core.SideCharacters)
	ed.AddActor(monster)
	ed.AddActor(target)
	ed.SetupEncounter()

	action, _ := ed.AIDirector.chooseBestAttackAction(monster, core.ActAction, ed)
	if action == nil || action.Name != "High Damage" {
		t.Errorf("Expected High Damage, got %v", action)
	}
}
