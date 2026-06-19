package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleAssassinate(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{EnableSpecialAbilities: true})

	assassin := &actor.Actor{
		InstanceID: 1,
		Features: []core.Feature{
			{
				Name:  core.SpecAbilityAssassinate,
				Hooks: map[core.HookType]bool{core.HookOnSelfAttack: true},
			},
		},
	}

	target := &actor.Actor{
		InstanceID: 2,
		StateManager: state_manager.StateManager{
			HasTakenTurnThisCombat: false,
		},
	}

	t.Run("Advantage against those who haven't acted", func(t *testing.T) {
		opts := roll_manager.RollOptions{Advantage: core.RollNormal}
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				AttackRoll: &opts,
			},
		}

		err := ed.HandleAssassinate(assassin, assassin.Features[0], core.HookOnSelfAttack, ctx)
		if err != nil {
			t.Fatalf("HandleAssassinate failed: %v", err)
		}

		if opts.Advantage != core.RollAdvantage {
			t.Errorf("Expected RollAdvantage, got %v", opts.Advantage)
		}
	})

	t.Run("No advantage against those who have acted", func(t *testing.T) {
		target.StateManager.HasTakenTurnThisCombat = true
		opts := roll_manager.RollOptions{Advantage: core.RollNormal}
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				AttackRoll: &opts,
			},
		}

		err := ed.HandleAssassinate(assassin, assassin.Features[0], core.HookOnSelfAttack, ctx)
		if err != nil {
			t.Fatalf("HandleAssassinate failed: %v", err)
		}

		if opts.Advantage != core.RollNormal {
			t.Errorf("Expected RollNormal, got %v", opts.Advantage)
		}
	})
}

func TestHandleBerserk(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{EnableSpecialAbilities: true})

	ironGolem := &actor.Actor{
		InstanceID: 1,
		StateManager: state_manager.StateManager{
			Conditions: core.NewActorConditions(),
		},
		Features: []core.Feature{
			{
				Name:  core.SpecAbilityBerserk,
				Hooks: map[core.HookType]bool{core.HookOnTurnStart: true},
				Data: core.FeatureData{
					Value: 6,
					Die:   core.D6,
				},
			},
		},
	}

	t.Run("Berserk triggers on specific roll", func(t *testing.T) {
		// Seed 1,1 roll for D6 is 1 (based on previous observations or we just try)
		// We'll iterate until we get a 6 or just mock it if we can.
		// Since we can't easily mock ed.RollManager.RollDie without interface, we rely on seed.

		found := false
		for i := 0; i < 100; i++ {
			err := ed.HandleBerserk(ironGolem, ironGolem.Features[0], core.HookOnTurnStart, nil)
			if err != nil {
				t.Fatalf("HandleBerserk failed: %v", err)
			}
			if ironGolem.StateManager.Conditions.Has(core.ConditionBerserk) {
				found = true
				break
			}
		}

		if !found {
			t.Error("Berserk condition never applied after 100 attempts")
		}
	})
}

func TestHandleBloodFrenzy(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{EnableSpecialAbilities: true})

	shark := &actor.Actor{
		InstanceID: 1,
		Features: []core.Feature{
			{
				Name:  core.SpecAbilityBloodFrenzy,
				Hooks: map[core.HookType]bool{core.HookOnTargetInjured: true},
			},
		},
	}

	target := &actor.Actor{
		InstanceID: 2,
		StateManager: state_manager.StateManager{
			CurrentHP:   5,
			MaxHP:       10,
			HealthState: core.HealthStateBloody,
		},
	}

	t.Run("Advantage against injured targets", func(t *testing.T) {
		opts := roll_manager.RollOptions{Advantage: core.RollNormal}
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				AttackRoll: &opts,
				Action:     &core.Action{ActionType: core.ATMelee},
			},
		}

		err := ed.HandleBloodFrenzy(shark, shark.Features[0], core.HookOnTargetInjured, ctx)
		if err != nil {
			t.Fatalf("HandleBloodFrenzy failed: %v", err)
		}

		if opts.Advantage != core.RollAdvantage {
			t.Errorf("Expected RollAdvantage, got %v", opts.Advantage)
		}
	})
}

func TestHandlePackTactics(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{EnableSpecialAbilities: true})

	wolf := &actor.Actor{
		InstanceID: 1,
		ActorType:  core.ActorTypeMonster,
		Features: []core.Feature{
			{
				Name:  core.SpecAbilityPackTactics,
				Hooks: map[core.HookType]bool{core.HookOnSelfAttack: true},
			},
		},
	}

	ally := &actor.Actor{
		InstanceID: 2,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 10,
			MaxHP:     10,
		},
	}

	ed.Actors[1] = wolf
	ed.Actors[2] = ally

	t.Run("Advantage with ally", func(t *testing.T) {
		opts := roll_manager.RollOptions{Advantage: core.RollNormal}
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				AttackRoll: &opts,
			},
		}

		err := ed.HandlePackTactics(wolf, wolf.Features[0], core.HookOnSelfAttack, ctx)
		if err != nil {
			t.Fatalf("HandlePackTactics failed: %v", err)
		}

		if opts.Advantage != core.RollAdvantage {
			t.Errorf("Expected RollAdvantage, got %v", opts.Advantage)
		}
	})
}

func TestHandleReckless(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{EnableSpecialAbilities: true})

	barbarian := &actor.Actor{
		InstanceID: 1,
		StateManager: state_manager.StateManager{
			Conditions: core.NewActorConditions(),
		},
	}

	err := ed.HandleReckless(barbarian, core.Feature{}, core.HookOnTurnStart, nil)
	if err != nil {
		t.Fatalf("HandleReckless failed: %v", err)
	}

	if !barbarian.StateManager.Conditions.Has(core.ConditionReckless) {
		t.Error("Expected Reckless condition to be applied")
	}
}

func TestHandleRegeneration(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{EnableSpecialAbilities: true})

	troll := &actor.Actor{
		InstanceID: 1,
		StateManager: state_manager.StateManager{
			CurrentHP: 10,
			MaxHP:     20,
		},
	}

	f := core.Feature{
		Name: core.SpecAbilityRegeneration,
		Data: core.FeatureData{Value: 5},
	}

	err := ed.HandleRegeneration(troll, f, core.HookOnTurnStart, nil)
	if err != nil {
		t.Fatalf("HandleRegeneration failed: %v", err)
	}

	if troll.StateManager.CurrentHP != 15 {
		t.Errorf("Expected 15 HP, got %d", troll.StateManager.CurrentHP)
	}
}
