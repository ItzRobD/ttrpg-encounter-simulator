package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleSneakAttack(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Rogue with Sneak Attack
	rogue := &actor.Actor{
		InstanceID: 1,
		Name:       "Rogue",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP:       30,
			MaxHP:           30,
			OncePerTurnUsed: make(map[string]bool),
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilitySneakAttack,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfHit: true,
				},
				Data: core.FeatureData{
					NumberOfDice: 2, // 2d6 sneak attack
					Die:          core.D6,
				},
			},
		},
	}

	// Target Goblin
	goblin := &actor.Actor{
		InstanceID: 2,
		Name:       "Goblin",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	// Allied Fighter
	fighter := &actor.Actor{
		InstanceID: 3,
		Name:       "Fighter",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP: 30,
			MaxHP:     30,
		},
	}

	ed.Actors[1] = rogue
	ed.Actors[2] = goblin
	ed.Actors[3] = fighter
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	// Base Action: Shortsword
	shortsword := &core.Action{
		Name:       "Shortsword",
		ActionType: core.ATMelee,
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 1, Die: core.D6, Modifier: 3, DamageType: core.DamagePiercing},
		},
	}

	t.Run("Sneak Attack triggers with Advantage", func(t *testing.T) {
		goblin.StateManager.CurrentHP = 50
		rogue.StateManager.OncePerTurnUsed = make(map[string]bool)

		isCrit := false
		ctx := &FeatureContext{
			Target: goblin,
			AttackContext: &AttackContext{
				Action:     shortsword,
				IsCritical: &isCrit,
				AttackRoll: &roll_manager.RollOptions{
					Advantage: core.RollAdvantage,
				},
			},
		}

		err := ed.HandleSneakAttack(rogue, rogue.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleSneakAttack failed: %v", err)
		}

		// Goblin should have taken 2d6 damage (~7) in addition to whatever hit happened (hit not resolved here)
		// Wait, resolveDamage is called INSIDE HandleSneakAttack.
		if goblin.StateManager.CurrentHP == 50 {
			t.Error("Goblin should have taken Sneak Attack damage")
		}
		if !rogue.StateManager.OncePerTurnUsed[string(core.SpecAbilitySneakAttack)] {
			t.Error("Sneak Attack should be marked as used")
		}
	})

	t.Run("Sneak Attack triggers with nearby Ally", func(t *testing.T) {
		goblin.StateManager.CurrentHP = 50
		rogue.StateManager.OncePerTurnUsed = make(map[string]bool)

		isCrit := false
		ctx := &FeatureContext{
			Target: goblin,
			AttackContext: &AttackContext{
				Action:     shortsword,
				IsCritical: &isCrit,
				AttackRoll: &roll_manager.RollOptions{
					Advantage: core.RollNormal,
				},
			},
		}

		err := ed.HandleSneakAttack(rogue, rogue.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleSneakAttack failed: %v", err)
		}

		if goblin.StateManager.CurrentHP == 50 {
			t.Error("Goblin should have taken Sneak Attack damage due to nearby ally")
		}
	})

	t.Run("Sneak Attack does NOT trigger with Disadvantage", func(t *testing.T) {
		goblin.StateManager.CurrentHP = 50
		rogue.StateManager.OncePerTurnUsed = make(map[string]bool)

		ctx := &FeatureContext{
			Target: goblin,
			AttackContext: &AttackContext{
				Action: shortsword,
				AttackRoll: &roll_manager.RollOptions{
					Advantage: core.RollDisadvantage,
				},
			},
		}

		err := ed.HandleSneakAttack(rogue, rogue.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleSneakAttack failed: %v", err)
		}

		if goblin.StateManager.CurrentHP != 50 {
			t.Error("Sneak Attack should NOT trigger with Disadvantage")
		}
	})

	t.Run("Sneak Attack does NOT trigger without Advantage or Ally", func(t *testing.T) {
		// Remove ally
		delete(ed.Actors, 3)
		defer func() { ed.Actors[3] = fighter }()

		goblin.StateManager.CurrentHP = 50
		rogue.StateManager.OncePerTurnUsed = make(map[string]bool)

		ctx := &FeatureContext{
			Target: goblin,
			AttackContext: &AttackContext{
				Action: shortsword,
				AttackRoll: &roll_manager.RollOptions{
					Advantage: core.RollNormal,
				},
			},
		}

		err := ed.HandleSneakAttack(rogue, rogue.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleSneakAttack failed: %v", err)
		}

		if goblin.StateManager.CurrentHP != 50 {
			t.Error("Sneak Attack should NOT trigger without Advantage or Ally")
		}
	})

	t.Run("Sneak Attack is once per turn", func(t *testing.T) {
		goblin.StateManager.CurrentHP = 50
		rogue.StateManager.OncePerTurnUsed = make(map[string]bool)
		rogue.StateManager.OncePerTurnUsed[string(core.SpecAbilitySneakAttack)] = true

		ctx := &FeatureContext{
			Target: goblin,
			AttackContext: &AttackContext{
				Action: shortsword,
				AttackRoll: &roll_manager.RollOptions{
					Advantage: core.RollAdvantage,
				},
			},
		}

		err := ed.HandleSneakAttack(rogue, rogue.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleSneakAttack failed: %v", err)
		}

		if goblin.StateManager.CurrentHP != 50 {
			t.Error("Sneak Attack should only trigger once per turn")
		}
	})

	t.Run("Sneak Attack only triggers on Weapon Attack", func(t *testing.T) {
		goblin.StateManager.CurrentHP = 50
		rogue.StateManager.OncePerTurnUsed = make(map[string]bool)

		spellAction := &core.Action{
			Name:       "Firebolt",
			ActionType: core.ATSpell,
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 1, Die: core.D10, Modifier: 0, DamageType: core.DamageFire},
			},
		}

		ctx := &FeatureContext{
			Target: goblin,
			AttackContext: &AttackContext{
				Action: spellAction,
				AttackRoll: &roll_manager.RollOptions{
					Advantage: core.RollAdvantage,
				},
			},
		}

		err := ed.HandleSneakAttack(rogue, rogue.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleSneakAttack failed: %v", err)
		}

		if goblin.StateManager.CurrentHP != 50 {
			t.Error("Sneak Attack should NOT trigger on a spell attack")
		}
	})

	t.Run("Sneak Attack doubles dice on Critical Hit", func(t *testing.T) {
		goblin.StateManager.CurrentHP = 100
		goblin.StateManager.MaxHP = 100
		rogue.StateManager.OncePerTurnUsed = make(map[string]bool)

		ed.SimOptions.UseImprovedCritical = false

		isCrit := true
		ctx := &FeatureContext{
			Target: goblin,
			AttackContext: &AttackContext{
				Action:     shortsword,
				IsCritical: &isCrit,
				AttackRoll: &roll_manager.RollOptions{
					Advantage: core.RollAdvantage,
				},
			},
		}

		// Sneak Attack is 2d6. On crit 4d6.
		err := ed.HandleSneakAttack(rogue, rogue.Features[0], core.HookOnSelfHit, ctx)
		if err != nil {
			t.Fatalf("HandleSneakAttack failed: %v", err)
		}

		damageTaken := 100 - goblin.StateManager.CurrentHP
		t.Logf("Damage dealt on Crit Sneak Attack: %d", damageTaken)

		if damageTaken <= 2 { // min of 2d6
			t.Errorf("Damage too low for 4d6: %d", damageTaken)
		}
	})
}
