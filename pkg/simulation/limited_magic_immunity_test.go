package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleLimitedMagicImmunity(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Rakshasa-like actor with Limited Magic Immunity (6th level or lower)
	rakshasa := &actor.Actor{
		InstanceID: 1,
		Name:       "Rakshasa",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		AC:         16,
		Abilities:  core.Abilities{AbilityScores: core.AbilityScores{Dexterity: 17, Wisdom: 16, Charisma: 20}},
		StateManager: state_manager.StateManager{
			CurrentHP:   110,
			MaxHP:       110,
			HealthState: core.HealthStateHealthy,
			Conditions:  core.NewActorConditions(),
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityLimitedMagicImmunity,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfSavingThrow:             true,
					core.HookOnSelfSavingThrowAgainstMagic: true,
					core.HookOnSelfDamageTaken:             true,
				},
				Data: core.FeatureData{
					Value: 6, // 6th level or lower
				},
			},
		},
	}

	attacker := &actor.Actor{
		InstanceID: 2,
		Name:       "Wizard",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			HealthState: core.HealthStateHealthy,
			Conditions:  core.NewActorConditions(),
		},
	}

	ed.Actors[1] = rakshasa
	ed.Actors[2] = attacker
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Immune to 3rd level Fireball", func(t *testing.T) {
		fireball := &core.Action{
			ID:          core.MakeID("fireball:3"),
			Name:        "Fireball",
			ActionType:  core.ATSpell,
			HasDC:       true,
			DCSaveDC:    20, // High DC to force failure
			DCAbility:   core.AbilityDexterity,
			DCOnSuccess: core.DCOnSuccessHalf,
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 8, Die: core.D6, Modifier: 0, DamageType: core.DamageFire},
			},
		}

		rakshasa.StateManager.CurrentHP = 110
		err := ed.Adjudicator.executeIndividualStrike(attacker, rakshasa, fireball)
		if err != nil {
			t.Fatalf("Action failed: %v", err)
		}

		if rakshasa.StateManager.CurrentHP != 110 {
			t.Errorf("Expected Rakshasa to take 0 damage from 3rd level Fireball, but HP is %d", rakshasa.StateManager.CurrentHP)
		}
	})

	t.Run("Advantage on 7th level Delayed Blast Fireball", func(t *testing.T) {
		// We can't easily check advantage in integration test without mocking,
		// but we can check the handler logic.
		dbf := &core.Action{
			ID:          core.MakeID("dbf:7"),
			Name:        "Delayed Blast Fireball",
			ActionType:  core.ATSpell,
			HasDC:       true,
			DCSaveDC:    15,
			DCAbility:   core.AbilityDexterity,
			DCOnSuccess: core.DCOnSuccessHalf,
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 12, Die: core.D6, Modifier: 0, DamageType: core.DamageFire},
			},
		}

		opts := roll_manager.RollOptions{
			RollType: core.DiceRollSavingThrow,
		}
		ctx := &FeatureContext{
			Target: rakshasa,
			SaveContext: &SaveContext{
				Target:  rakshasa,
				Options: &opts,
			},
			AttackContext: &AttackContext{
				Action: dbf,
			},
		}

		err := ed.HandleLimitedMagicImmunity(rakshasa, rakshasa.Features[0], core.HookOnSelfSavingThrowAgainstMagic, ctx)
		if err != nil {
			t.Fatalf("Handler failed: %v", err)
		}

		if opts.Advantage != core.RollAdvantage {
			t.Errorf("Expected advantage on 7th level spell, got %v", opts.Advantage)
		}

		// Ensure it takes damage
		rakshasa.StateManager.CurrentHP = 110
		// Force save failure
		dbf.DCSaveDC = 40
		err = ed.Adjudicator.executeIndividualStrike(attacker, rakshasa, dbf)
		if err != nil {
			t.Fatalf("Action failed: %v", err)
		}

		if rakshasa.StateManager.CurrentHP == 110 {
			t.Error("Rakshasa should have taken damage from 7th level spell")
		}
	})

	t.Run("Non-spell action unaffected", func(t *testing.T) {
		sword := &core.Action{
			ID:          core.MakeID("sword"),
			Name:        "Longsword",
			ActionType:  core.ATMelee,
			AttackBonus: 10,
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 1, Die: core.D8, Modifier: 5, DamageType: core.DamageSlashing},
			},
		}

		rakshasa.StateManager.CurrentHP = 110
		rakshasa.StateManager.MaxHP = 110
		rakshasa.StateManager.HealthState = core.HealthStateHealthy

		// Give the attacker a high bonus to guarantee a hit against AC 16
		sword.AttackBonus = 20

		err := ed.Adjudicator.executeIndividualStrike(attacker, rakshasa, sword)
		if err != nil {
			t.Fatalf("Action failed: %v", err)
		}

		if rakshasa.StateManager.CurrentHP == 110 {
			t.Error("Rakshasa should have taken damage from non-spell sword attack")
		}
	})
}
