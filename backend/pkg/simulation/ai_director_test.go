package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestAIDirector_SelectTarget(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{UseWeightedAI: true})
	aid := ed.AIDirector

	a := &actor.Actor{
		InstanceID: 1,
		ActorType:  core.ActorTypeCharacter,
		AC:         15,
		Abilities:  core.Abilities{AbilityScores: core.AbilityScores{Strength: 16}},
	}

	targets := map[int]*actor.Actor{
		2: {
			InstanceID: 2,
			Name:       "Weakling",
			AC:         10,
			StateManager: state_manager.StateManager{
				CurrentHP: 5,
				MaxHP:     20,
			},
		},
		3: {
			InstanceID: 3,
			Name:       "Boss",
			AC:         20,
			Metadata:   actor.Metadata{IsLegendary: true},
			StateManager: state_manager.StateManager{
				CurrentHP: 100,
				MaxHP:     100,
			},
		},
	}

	// With high weight on low HP, Weakling should be selected
	// In the weighted AI, we need to ensure visibility is set to visible for HP factor to work
	ed.SimOptions.HPVisibilityMode = core.HPVisible
	targetIDs := aid.SelectTarget(a, targets, core.TTDamage, ed)
	if len(targetIDs) == 0 || targetIDs[0] != 2 {
		t.Errorf("Expected target 2 (Weakling), got %v", targetIDs)
	}
}

func TestAIDirector_EmergencyHealOverride(t *testing.T) {
	simOptions := &core.SimulationOptions{
		CharacterHealThresholdPct:      50,
		CharacterEmergencyThresholdPct: 20,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)
	aid := ed.AIDirector

	// Aggressive Paladin
	paladin := &actor.Actor{
		InstanceID: 1,
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		Behavior:   actor.BehaviorConfig{ActionPreference: core.APAttack},
		Actions: []core.Action{
			{ID: core.ID("1"), Name: "Attack", AverageDamage: 10, Cost: core.ActionCost{ActivationType: core.ActAction}},
			{ID: core.ID("2"), Name: "Heal", AverageHeal: 10, Cost: core.ActionCost{ActivationType: core.ActAction}},
		},
		StateManager: state_manager.StateManager{CurrentHP: 50, MaxHP: 50},
		HPConfig:     core.HPConfig{HPAverage: 50},
	}

	// Ally
	ally := &actor.Actor{
		InstanceID:   2,
		Side:         core.SideCharacters,
		ActorType:    core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{CurrentHP: 50, MaxHP: 100}, // 50% HP
		HPConfig:     core.HPConfig{HPAverage: 100},
	}

	// Enemy
	enemy := &actor.Actor{
		InstanceID:   3,
		Side:         core.SideMonsters,
		ActorType:    core.ActorTypeMonster,
		StateManager: state_manager.StateManager{CurrentHP: 50, MaxHP: 50},
		HPConfig:     core.HPConfig{HPAverage: 50},
	}

	ed.Actors[1] = paladin
	ed.Actors[2] = ally
	ed.Actors[3] = enemy
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	// Ensure enemies are identified correctly in SelectAction
	// Step 1: Identify valid targets (enemies)
	// We need to add the actors to the director properly if we were using ed.AddActor,
	// but here we manually assigned them to the map.
	// ed.GetEnemyTargets(paladin) uses ed.Actors.

	// --- Case 1: Ally at 50% (Regular Threshold) ---
	// Paladin is aggressive, should stay aggressive
	ed.Statistics.MarkNeedsHealing(2)
	decision := aid.SelectActionDecision(paladin, ed)
	intents := aid.SelectAction(paladin, decision, ed)
	if len(intents) == 0 {
		t.Fatalf("Expected intent, got none")
	}
	if intents[0].Action.Name != "Attack" {
		t.Errorf("Expected Attack at 50%% HP, got %s", intents[0].Action.Name)
	}

	// --- Case 2: Ally at 10% (Emergency Threshold) ---
	// Paladin is aggressive, but should switch to Protective due to emergency
	ally.StateManager.CurrentHP = 10
	ed.Statistics.MarkNeedsEmergencyHealing(2)
	decision = aid.SelectActionDecision(paladin, ed)
	intents = aid.SelectAction(paladin, decision, ed)
	if len(intents) == 0 {
		t.Fatalf("Expected intent, got none")
	}
	if intents[0].Action.Name != "Heal" {
		t.Errorf("Expected Heal at 10%% HP (Emergency), got %s", intents[0].Action.Name)
	}

	// --- Case 3: Ally at 10% (Emergency Threshold) but actor is APHeal ---
	// Protective actor should switch on regular threshold too
	ally.StateManager.CurrentHP = 40
	ed.Statistics.ClearNeedsEmergencyHealing(2)
	ed.Statistics.MarkNeedsHealing(2)
	paladin.Behavior.ActionPreference = core.APHeal
	decision = aid.SelectActionDecision(paladin, ed)
	intents = aid.SelectAction(paladin, decision, ed)
	if intents[0].Action.Name != "Heal" {
		t.Errorf("Expected Heal at 40%% HP (Regular) for Protective strategy, got %s", intents[0].Action.Name)
	}
}

func TestAIDirector_NoHealingCapacityOverride(t *testing.T) {
	simOptions := &core.SimulationOptions{
		CharacterHealThresholdPct:      50,
		CharacterEmergencyThresholdPct: 20,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)
	aid := ed.AIDirector

	// Fighter (No healing actions)
	fighter := &actor.Actor{
		InstanceID: 1,
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		Behavior:   actor.BehaviorConfig{ActionPreference: core.APAttack},
		Actions: []core.Action{
			{ID: core.ID("1"), Name: "Attack", AverageDamage: 10, Cost: core.ActionCost{ActivationType: core.ActAction}},
		},
		StateManager: state_manager.StateManager{CurrentHP: 50, MaxHP: 50},
		HPConfig:     core.HPConfig{HPAverage: 50},
	}

	// Ally in emergency
	ally := &actor.Actor{
		InstanceID:   2,
		Side:         core.SideCharacters,
		ActorType:    core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{CurrentHP: 5, MaxHP: 100}, // 5% HP
		HPConfig:     core.HPConfig{HPAverage: 100},
	}

	// Enemy
	enemy := &actor.Actor{
		InstanceID:   3,
		Side:         core.SideMonsters,
		ActorType:    core.ActorTypeMonster,
		StateManager: state_manager.StateManager{CurrentHP: 50, MaxHP: 50},
		HPConfig:     core.HPConfig{HPAverage: 50},
	}

	ed.Actors[1] = fighter
	ed.Actors[2] = ally
	ed.Actors[3] = enemy
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	// Ally is in emergency
	ed.Statistics.MarkNeedsEmergencyHealing(2)

	// Fighter should NOT switch to APHeal because they have no healing capacity
	decision := aid.SelectActionDecision(fighter, ed)
	intents := aid.SelectAction(fighter, decision, ed)
	if len(intents) == 0 {
		t.Fatalf("Expected intent, got none")
	}
	// Action should still be "Attack" and strategy should NOT have switched internally (though we only see the action)
	if intents[0].Action.Name != "Attack" {
		t.Errorf("Expected Attack even with emergency because Fighter has no healing capacity, got %s", intents[0].Action.Name)
	}
}
