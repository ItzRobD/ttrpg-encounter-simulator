package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"fmt"
	"testing"
)

func TestAdjudicator_ResolveAction_ExtraAttack(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{})
	adj := ed.Adjudicator

	a := &actor.Actor{
		InstanceID: 1,
		ActorType:  core.ActorTypeCharacter,
		AC:         15,
		Abilities:  core.Abilities{AbilityScores: core.AbilityScores{Strength: 16}},
		StateManager: state_manager.StateManager{
			AttackCount: 2, // Extra Attack
		},
		Actions: []core.Action{
			{
				ID:          core.ID("1"),
				Name:        "Longsword",
				ActionType:  core.ATMelee,
				Cost:        core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AttackBonus: 5,
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 1, Die: core.D8, DamageType: core.DamageSlashing},
				},
			},
		},
	}

	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Target",
		AC:         10,
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
		},
	}

	ed.Actors[1] = a
	ed.Actors[2] = target

	intent := ActionIntent{
		ActorID:   1,
		TargetIDs: []int{2},
		Action:    a.Actions[0],
	}

	err := adj.ResolveAction(a, intent)
	if err != nil {
		t.Fatalf("ResolveAction failed: %v", err)
	}

	// With 2 attacks, target HP should likely be reduced (statistically very probable vs AC 10 with +5 bonus)
	// We'll just verify the call finishes without error and state is updated
	if a.StateManager.ActionUsedCount != 1 {
		t.Errorf("Expected ActionUsedCount 1, got %d", a.StateManager.ActionUsedCount)
	}
}

func TestAdjudicator_ResolveAction_Multiattack(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{})
	adj := ed.Adjudicator

	m := &actor.Actor{
		InstanceID: 1,
		ActorType:  core.ActorTypeMonster,
		Actions: []core.Action{
			{
				ID:         core.ID("1"),
				Name:       "Multiattack",
				ActionType: core.ATMultiAttack,
				Cost:       core.ActionCost{ActivationType: core.ActAction, Value: 1},
				Multiattack: []core.Multiattack{
					{ActionID: core.ID("2"), Count: 2}, // 2 Claws
				},
			},
			{
				ID:          core.ID("2"),
				Name:        "Claw",
				ActionType:  core.ATAction,
				AttackBonus: 4,
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 1, Die: core.D6, DamageType: core.DamageSlashing},
				},
			},
		},
	}

	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Target",
		AC:         10,
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
		},
	}

	ed.Actors[1] = m
	ed.Actors[2] = target

	intent := ActionIntent{
		ActorID:   1,
		TargetIDs: []int{2},
		Action:    m.Actions[0],
	}

	err := adj.ResolveAction(m, intent)
	if err != nil {
		t.Fatalf("ResolveAction failed: %v", err)
	}

	if m.StateManager.ActionUsedCount != 1 {
		t.Errorf("Expected ActionUsedCount 1, got %d", m.StateManager.ActionUsedCount)
	}
}

func TestAdjudicator_ExecuteHealing_Basic(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{})
	adj := ed.Adjudicator

	a := &actor.Actor{InstanceID: 1, Name: "Healer", Side: core.SideCharacters, ActorType: core.ActorTypeCharacter}
	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Target",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP: 5,
			MaxHP:     20,
		},
	}
	ed.Actors[1] = a
	ed.Actors[2] = target
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	action := core.Action{
		ID:          core.ID("heal-1"),
		Name:        "Cure Wounds",
		ActionType:  core.ATSpell,
		AverageHeal: 10,
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 1, Die: core.D8, Modifier: 2},
		},
	}

	err := adj.executeHealing(a, 2, &action, ed)
	if err != nil {
		t.Fatalf("executeHealing failed: %v", err)
	}

	if target.StateManager.CurrentHP <= 5 {
		t.Errorf("Target HP was not increased. HP: %d", target.StateManager.CurrentHP)
	}
}

func TestAdjudicator_ExecuteHealing_MassPriority(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{CharacterHealThresholdPct: 50})
	adj := ed.Adjudicator

	a := &actor.Actor{InstanceID: 1, Name: "Healer", Side: core.SideCharacters, ActorType: core.ActorTypeCharacter}
	ed.Actors[1] = a

	// 8 Allies, only 6 should be healed
	for i := 2; i <= 9; i++ {
		target := &actor.Actor{
			InstanceID: i,
			Name:       fmt.Sprintf("Ally %d", i),
			Side:       core.SideCharacters,
			ActorType:  core.ActorTypeCharacter,
			StateManager: state_manager.StateManager{
				CurrentHP: 5,
				MaxHP:     20,
			},
			HPConfig: core.HPConfig{HPAverage: 20},
		}
		ed.Actors[i] = target
	}
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	// Mark some as priority
	ed.Statistics.MarkNeedsEmergencyHealing(9)
	ed.Statistics.MarkNeedsEmergencyHealing(8)
	ed.Statistics.MarkNeedsHealing(7)
	ed.Statistics.MarkNeedsHealing(6)
	ed.Statistics.MarkNeedsHealing(5)
	ed.Statistics.MarkNeedsHealing(4)

	action := core.Action{
		ID:          core.ID("mass-heal-1"),
		Name:        "Mass Cure Wounds",
		ActionType:  core.ATSpell,
		AverageHeal: 10,
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 1, Die: core.D8, Modifier: 2},
		},
	}

	err := adj.executeHealing(a, 2, &action, ed)
	if err != nil {
		t.Fatalf("executeHealing failed: %v", err)
	}

	// Count how many were healed
	healedCount := 0
	for i := 2; i <= 9; i++ {
		if ed.Actors[i].StateManager.CurrentHP > 5 {
			healedCount++
		}
	}

	if healedCount != 6 {
		t.Errorf("Expected 6 allies healed, got %d", healedCount)
	}

	// Verify priority targets were healed
	if ed.Actors[9].StateManager.CurrentHP <= 5 {
		t.Error("Priority target 9 was not healed")
	}
	if ed.Actors[8].StateManager.CurrentHP <= 5 {
		t.Error("Priority target 8 was not healed")
	}
}

func TestAdjudicator_ExecuteHealing_MassHealPool(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{})
	adj := ed.Adjudicator

	a := &actor.Actor{InstanceID: 1, Name: "Healer", Side: core.SideCharacters, ActorType: core.ActorTypeCharacter}
	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Target",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP:  50,
			MaxHP:      1000, // Needs 950 HP
			Conditions: core.NewActorConditions(),
		},
	}
	target.StateManager.Conditions.Add(core.ConditionBlinded)

	ed.Actors[1] = a
	ed.Actors[2] = target
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	action := core.Action{
		ID:         core.ID("mass-heal"),
		Name:       "Mass Heal",
		ActionType: core.ATSpell,
	}

	err := adj.executeHealing(a, 2, &action, ed)
	if err != nil {
		t.Fatalf("executeHealing failed: %v", err)
	}

	// Pool is 700. 50 + 700 = 750
	if target.StateManager.CurrentHP != 750 {
		t.Errorf("Expected 750 HP, got %d", target.StateManager.CurrentHP)
	}

	if target.StateManager.Conditions.Has(core.ConditionBlinded) {
		t.Error("Target should no longer be blinded")
	}
}

func TestAdjudicator_ExecuteHealing_Aid(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{})
	adj := ed.Adjudicator

	a := &actor.Actor{InstanceID: 1, Name: "Healer", Side: core.SideCharacters, ActorType: core.ActorTypeCharacter}
	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Target",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP: 10,
			MaxHP:     10,
		},
	}
	ed.Actors[1] = a
	ed.Actors[2] = target
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	action := core.Action{
		ID:          core.ID("aid"),
		Name:        "Aid",
		ActionType:  core.ATSpell,
		AverageHeal: 5,
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 0, Modifier: 5},
		},
	}

	err := adj.executeHealing(a, 2, &action, ed)
	if err != nil {
		t.Fatalf("executeHealing failed: %v", err)
	}

	if target.StateManager.MaxHP != 15 {
		t.Errorf("Expected MaxHP 15, got %d", target.StateManager.MaxHP)
	}
	if target.StateManager.CurrentHP != 15 {
		t.Errorf("Expected CurrentHP 15, got %d", target.StateManager.CurrentHP)
	}
}

func TestAdjudicator_ResolveMultiattack_Retarget(t *testing.T) {
	simOptions := &core.SimulationOptions{
		MultiattackPolicy: core.MultiattackPolicyRetargetOnDown,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)
	adj := ed.Adjudicator

	m := &actor.Actor{
		InstanceID: 1,
		ActorType:  core.ActorTypeMonster,
		Actions: []core.Action{
			{
				ID:         core.ID("ma"),
				Name:       "Multiattack",
				ActionType: core.ATMultiAttack,
				Cost:       core.ActionCost{ActivationType: core.ActAction, Value: 1},
				Multiattack: []core.Multiattack{
					{ActionID: core.ID("claw"), Count: 2},
				},
			},
			{
				ID:          core.ID("claw"),
				Name:        "Claw",
				ActionType:  core.ATAction,
				AttackBonus: 100, // Guaranteed hit
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 1, Die: core.D6, Modifier: 100, DamageType: core.DamageSlashing}, // Guaranteed kill
				},
			},
		},
	}

	t1 := &actor.Actor{
		InstanceID: 2,
		Name:       "Target 1",
		AC:         10,
		Side:       core.SideCharacters,
		StateManager: state_manager.StateManager{
			CurrentHP: 10,
			MaxHP:     10,
		},
	}
	t2 := &actor.Actor{
		InstanceID: 3,
		Name:       "Target 2",
		AC:         10,
		Side:       core.SideCharacters,
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
		},
	}

	ed.Actors[1] = m
	ed.Actors[2] = t1
	ed.Actors[3] = t2
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	intent := ActionIntent{
		ActorID:   1,
		TargetIDs: []int{2},
		Action:    m.Actions[0],
	}

	err := adj.ResolveAction(m, intent)
	if err != nil {
		t.Fatalf("ResolveAction failed: %v", err)
	}

	if t1.StateManager.CurrentHP != 0 {
		t.Errorf("Target 1 should be dead (HP 0), got %d", t1.StateManager.CurrentHP)
	}
	if t2.StateManager.CurrentHP >= 100 {
		t.Errorf("Target 2 should have been hit after Target 1 went down, got HP %d", t2.StateManager.CurrentHP)
	}
}
