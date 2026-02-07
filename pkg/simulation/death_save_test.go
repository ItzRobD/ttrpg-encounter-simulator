package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestDeathSaveLifecycle(t *testing.T) {
	simOptions := &core.SimulationOptions{}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	pc := &actor.Actor{
		InstanceID: 1,
		Name:       "Hero",
		ActorType:  core.ActorTypeCharacter,
		Side:       core.SideCharacters,
		StateManager: state_manager.StateManager{
			CurrentHP:   20,
			MaxHP:       20,
			HealthState: core.HealthStateHealthy,
			Conditions:  core.NewActorConditions(),
		},
	}

	monster := &actor.Actor{
		InstanceID: 2,
		Name:       "Monster",
		ActorType:  core.ActorTypeMonster,
		Side:       core.SideMonsters,
		StateManager: state_manager.StateManager{
			CurrentHP:   20,
			MaxHP:       20,
			HealthState: core.HealthStateHealthy,
			Conditions:  core.NewActorConditions(),
		},
	}

	ed.Actors[1] = pc
	ed.Actors[2] = monster
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("PC falls unconscious at 0 HP", func(t *testing.T) {
		action := core.Action{
			Name: "Bite",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 5, Die: core.D10, DamageType: core.DamagePiercing},
			},
		}
		err := ed.Adjudicator.resolveDamage(monster, pc, &action, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if pc.StateManager.CurrentHP != 0 {
			t.Errorf("Expected 0 HP, got %d", pc.StateManager.CurrentHP)
		}
		if pc.StateManager.HealthState != core.HealthStateUnconscious {
			t.Errorf("Expected HealthStateUnconscious, got %s", pc.StateManager.HealthState)
		}
		if !pc.StateManager.Conditions.Has(core.ConditionUnconscious) {
			t.Error("PC should have Unconscious condition")
		}
	})

	t.Run("Monster dies immediately at 0 HP", func(t *testing.T) {
		action := core.Action{
			Name: "Slash",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 5, Die: core.D10, DamageType: core.DamageSlashing},
			},
		}
		err := ed.Adjudicator.resolveDamage(pc, monster, &action, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if monster.StateManager.HealthState != core.HealthStateDead {
			t.Errorf("Expected HealthStateDead for monster, got %s", monster.StateManager.HealthState)
		}
	})

	t.Run("PC takes damage at 0 HP", func(t *testing.T) {
		// Reset PC to unconscious at 0 HP
		pc.StateManager.CurrentHP = 0
		pc.StateManager.HealthState = core.HealthStateUnconscious
		pc.StateManager.DeathSaveFailures = 0

		action := core.Action{
			Name: "Bite",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 1, Die: core.D4, DamageType: core.DamagePiercing},
			},
		}
		// Normal hit
		err := ed.Adjudicator.resolveDamage(monster, pc, &action, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if pc.StateManager.DeathSaveFailures != 1 {
			t.Errorf("Expected 1 death save failure, got %d", pc.StateManager.DeathSaveFailures)
		}

		// Critical hit
		err = ed.Adjudicator.resolveDamage(monster, pc, &action, false, true)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if pc.StateManager.DeathSaveFailures != 3 {
			t.Errorf("Expected 3 death save failures after crit, got %d", pc.StateManager.DeathSaveFailures)
		}

		if pc.StateManager.HealthState != core.HealthStateDead {
			t.Errorf("Expected HealthStateDead after 3 failures, got %s", pc.StateManager.HealthState)
		}
	})

	t.Run("Death Save at Turn Start", func(t *testing.T) {
		// Reset PC to unconscious
		pc.StateManager.CurrentHP = 0
		pc.StateManager.DeathSaveFailures = 0
		pc.StateManager.DeathSaveSuccesses = 0
		pc.StateManager.HealthState = core.HealthStateUnconscious
		pc.StateManager.Conditions.Add(core.ConditionUnconscious)

		// Mock RollManager to roll a 10 (Success)
		// Since we can't easily mock, we'll just run it and see it update
		// To be deterministic, we use the seed. With Seed(1,1), let's see what happens.

		ed.processDeathSaves(pc)

		if pc.StateManager.DeathSaveSuccesses == 0 && pc.StateManager.DeathSaveFailures == 0 {
			t.Error("Expected death save to be processed")
		}
	})

	t.Run("Natural 20 restores 1 HP", func(t *testing.T) {
		pc.StateManager.CurrentHP = 0
		pc.StateManager.HealthState = core.HealthStateUnconscious
		pc.StateManager.Conditions.Add(core.ConditionUnconscious)

		// Find a seed or loop until natural 20
		found := false
		for i := 0; i < 500; i++ {
			ed.processDeathSaves(pc)
			if pc.StateManager.CurrentHP == 1 {
				found = true
				break
			}
			// Reset for next try if not dead
			if pc.StateManager.HealthState == core.HealthStateDead || pc.StateManager.Conditions.Has(core.ConditionStable) {
				pc.StateManager.CurrentHP = 0
				pc.StateManager.DeathSaveFailures = 0
				pc.StateManager.DeathSaveSuccesses = 0
				pc.StateManager.HealthState = core.HealthStateUnconscious
				pc.StateManager.Conditions.Remove(core.ConditionStable)
			}
		}

		if !found {
			t.Error("Failed to roll a natural 20 in 500 tries")
		}
		if pc.StateManager.Conditions.Has(core.ConditionUnconscious) {
			t.Error("PC should no longer be unconscious after natural 20")
		}
	})

	t.Run("Stabilization and Damage", func(t *testing.T) {
		pc.StateManager.CurrentHP = 0
		pc.StateManager.DeathSaveFailures = 0
		pc.StateManager.DeathSaveSuccesses = 0
		pc.StateManager.HealthState = core.HealthStateUnconscious
		pc.StateManager.Conditions.Clear()
		pc.StateManager.Conditions.Add(core.ConditionUnconscious)

		// 1. Force stabilization
		for i := 0; i < 100; i++ {
			ed.processDeathSaves(pc)
			if pc.StateManager.Conditions.Has(core.ConditionStable) {
				break
			}
			if pc.StateManager.HealthState == core.HealthStateDead || pc.StateManager.CurrentHP > 0 {
				// Reset
				pc.StateManager.CurrentHP = 0
				pc.StateManager.DeathSaveFailures = 0
				pc.StateManager.DeathSaveSuccesses = 0
				pc.StateManager.HealthState = core.HealthStateUnconscious
				pc.StateManager.Conditions.Clear()
				pc.StateManager.Conditions.Add(core.ConditionUnconscious)
			}
		}

		if !pc.StateManager.Conditions.Has(core.ConditionStable) {
			t.Fatal("Failed to stabilize PC in 100 tries")
		}

		// 2. Verify no more death saves are made while stable
		pc.StateManager.DeathSaveSuccesses = 0
		pc.StateManager.DeathSaveFailures = 0
		ed.processDeathSaves(pc)
		if pc.StateManager.DeathSaveSuccesses != 0 || pc.StateManager.DeathSaveFailures != 0 {
			t.Error("Death saves should not be processed while stable")
		}

		// 3. Taking damage removes stability
		action := core.Action{
			Name: "Bite",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 1, Die: core.D4, DamageType: core.DamagePiercing},
			},
		}
		err := ed.Adjudicator.resolveDamage(monster, pc, &action, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if pc.StateManager.Conditions.Has(core.ConditionStable) {
			t.Error("Stable condition should be removed after taking damage")
		}
		if pc.StateManager.DeathSaveFailures != 1 {
			t.Errorf("Expected 1 death save failure after damage while stable, got %d", pc.StateManager.DeathSaveFailures)
		}

		// 4. Healing removes stability
		pc.StateManager.Conditions.Add(core.ConditionStable)
		pc.StateManager.ModifyHP(5, false, true)
		if pc.StateManager.Conditions.Has(core.ConditionStable) {
			t.Error("Stable condition should be removed after healing")
		}
	})
}
