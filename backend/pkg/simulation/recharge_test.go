package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestRechargeLifecycle(t *testing.T) {
	simOptions := &core.SimulationOptions{}
	// Fixed seed for predictability if needed, but we'll control rolls by iterating or checking state
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	rechargeAction := core.Action{
		ID:            "breath_weapon",
		Name:          "Fire Breath",
		ActionType:    core.ATRecharge,
		RechargeValue: 5, // Recharges on 5-6
		Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
		AverageDamage: 20,
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 6, Die: core.D6, DamageType: core.DamageFire},
		},
	}

	monster := &actor.Actor{
		InstanceID: 1,
		Name:       "Young Red Dragon",
		ActorType:  core.ActorTypeMonster,
		Side:       core.SideMonsters,
		Actions:    []core.Action{rechargeAction},
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
			Resource:  make(map[string]int),
		},
	}
	// Initial state: Charged
	monster.StateManager.Resource[rechargeAction.Name] = 1

	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Target",
		ActorType:  core.ActorTypeCharacter,
		Side:       core.SideCharacters,
		AC:         10,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	ed.Actors[1] = monster
	ed.Actors[2] = target
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Action consumption", func(t *testing.T) {
		intent := ActionIntent{
			ActivationType: core.ActAction,
			ActorID:        monster.InstanceID,
			TargetIDs:      []int{target.InstanceID},
			Action:         rechargeAction,
		}

		err := ed.Adjudicator.ResolveAction(monster, intent)
		if err != nil {
			t.Fatalf("ResolveAction failed: %v", err)
		}

		if monster.StateManager.Resource[rechargeAction.Name] != 0 {
			t.Errorf("Expected recharge action to be spent (Resource = 0), got %d", monster.StateManager.Resource[rechargeAction.Name])
		}
	})

	t.Run("AIDirector filters spent action", func(t *testing.T) {
		// Ensure monster is not incapacitated
		monster.StateManager.ActionUsedCount = 0
		monster.StateManager.BonusActionUsedCount = 0

		intents := ed.AIDirector.SelectAction(monster, core.DecisionAttack, ed)

		for _, intent := range intents {
			if intent.Action.Name == rechargeAction.Name {
				t.Errorf("AIDirector selected spent recharge action: %s", rechargeAction.Name)
			}
		}
	})

	t.Run("Recharge roll at turn start", func(t *testing.T) {
		// Force uncharged state
		monster.StateManager.Resource[rechargeAction.Name] = 0

		// Run processTurnStart multiple times until it recharges (or we can mock/check logic)
		// Since we use a real RollManager with seed, it's deterministic.
		// Alternatively, we can just verify that it DOES roll when Resource is 0.

		recharged := false
		for i := 0; i < 20; i++ {
			ed.processTurnStart(monster)
			if monster.StateManager.Resource[rechargeAction.Name] == 1 {
				recharged = true
				break
			}
		}

		if !recharged {
			t.Errorf("Action failed to recharge after 20 turns (Recharge 5-6)")
		}
	})

	t.Run("Recharge roll skipped when already charged", func(t *testing.T) {
		monster.StateManager.Resource[rechargeAction.Name] = 1

		// If we had a way to count rolls, we'd check it here.
		// But we can at least ensure it stays 1.
		ed.processTurnStart(monster)
		if monster.StateManager.Resource[rechargeAction.Name] != 1 {
			t.Errorf("Recharge action state changed unexpectedly")
		}
	})

	t.Run("Dragonborn Breath Weapon Resource Lifecycle", func(t *testing.T) {
		breathWeapon := core.Action{
			Name:          "Breath Weapon",
			ActionType:    core.ATAction,
			RechargeValue: 7, // Marks as limited resource
			Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
			AverageDamage: 10,
		}

		pc := &actor.Actor{
			InstanceID: 3,
			Name:       "Dragonborn PC",
			ActorType:  core.ActorTypeCharacter,
			Side:       core.SideCharacters,
			Actions:    []core.Action{breathWeapon},
			StateManager: state_manager.StateManager{
				CurrentHP: 50,
				MaxHP:     50,
				Resource:  map[string]int{"Breath Weapon": 1},
			},
		}

		ed.Actors[3] = pc
		ed.Statistics = NewEncounterStatistics(ed.Actors)

		// 1. Verify it's used and marked as spent
		intent := ActionIntent{
			ActivationType: core.ActAction,
			ActorID:        pc.InstanceID,
			TargetIDs:      []int{monster.InstanceID},
			Action:         breathWeapon,
		}

		err := ed.Adjudicator.ResolveAction(pc, intent)
		if err != nil {
			t.Fatalf("ResolveAction failed: %v", err)
		}

		if pc.StateManager.Resource["Breath Weapon"] != 0 {
			t.Errorf("Expected Breath Weapon to be spent, got %d", pc.StateManager.Resource["Breath Weapon"])
		}

		// 2. Verify AI skips it
		pc.StateManager.ActionUsedCount = 0
		intents := ed.AIDirector.SelectAction(pc, core.DecisionAttack, ed)
		for _, intent := range intents {
			if intent.Action.Name == "Breath Weapon" {
				t.Errorf("AI selected spent Breath Weapon")
			}
		}

		// 3. Verify it does NOT recharge at turn start (RechargeValue 7)
		ed.processTurnStart(pc)
		if pc.StateManager.Resource["Breath Weapon"] != 0 {
			t.Errorf("Breath Weapon recharged unexpectedly at turn start")
		}
	})
}
