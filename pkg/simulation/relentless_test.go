package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleRelentless(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Actor with Relentless (Threshold 10)
	orc := &actor.Actor{
		InstanceID: 1,
		Name:       "Orc",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 5,
			MaxHP:     15,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityRelentless,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfDamageTaken: true,
				},
				Data: core.FeatureData{
					Value: 10,
				},
			},
		},
	}

	attacker := &actor.Actor{
		InstanceID: 2,
		Name:       "Fighter",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
	}

	ed.Actors[1] = orc
	ed.Actors[2] = attacker
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Relentless Triggers (Damage <= Threshold)", func(t *testing.T) {
		orc.StateManager.CurrentHP = 5

		// 8 damage would drop it to -3, but since 8 <= 10, it should stay at 1.
		action := &core.Action{
			Name: "Sword",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 0, Modifier: 8, DamageType: core.DamageSlashing},
			},
		}

		err := ed.Adjudicator.resolveDamage(attacker, orc, action, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if orc.StateManager.CurrentHP != 1 {
			t.Errorf("Expected Orc to have 1 HP, got %d", orc.StateManager.CurrentHP)
		}
	})

	t.Run("Relentless Does Not Trigger (Damage > Threshold)", func(t *testing.T) {
		orc.StateManager.CurrentHP = 5

		// 12 damage would drop it to -7. Since 12 > 10, it should die.
		action := &core.Action{
			Name: "Heavy Hit",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 0, Modifier: 12, DamageType: core.DamageSlashing},
			},
		}

		err := ed.Adjudicator.resolveDamage(attacker, orc, action, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if orc.StateManager.CurrentHP != 0 {
			t.Errorf("Expected Orc to have 0 HP, got %d", orc.StateManager.CurrentHP)
		}
		if orc.StateManager.HealthState != core.HealthStateDead {
			t.Errorf("Expected Orc to be Dead, got %s", orc.StateManager.HealthState)
		}
	})

	t.Run("Relentless Does Not Trigger (Damage would not reduce to 0)", func(t *testing.T) {
		orc.StateManager.CurrentHP = 15

		// 8 damage would drop it to 7. Relentless should not intervene.
		action := &core.Action{
			Name: "Sword",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 0, Modifier: 8, DamageType: core.DamageSlashing},
			},
		}

		err := ed.Adjudicator.resolveDamage(attacker, orc, action, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if orc.StateManager.CurrentHP != 7 {
			t.Errorf("Expected Orc to have 7 HP, got %d", orc.StateManager.CurrentHP)
		}
	})

	t.Run("Relentless Endurance Triggers (Once per day)", func(t *testing.T) {
		halfOrc := &actor.Actor{
			InstanceID: 3,
			Name:       "Half-Orc",
			Side:       core.SideCharacters,
			ActorType:  core.ActorTypeCharacter,
			StateManager: state_manager.StateManager{
				CurrentHP:       5,
				MaxHP:           15,
				OncePerTurnUsed: make(map[string]bool),
			},
			Features: []core.Feature{
				{
					Name: core.SpecAbilityRelentlessEndurance,
					Hooks: map[core.HookType]bool{
						core.HookOnSelfDamageTaken: true,
					},
				},
			},
		}
		ed.Actors[3] = halfOrc
		ed.Statistics.statistics[3] = NewCombatStatistics()

		// 100 damage would drop it to -95, but Relentless Endurance should stay at 1.
		action := &core.Action{
			Name: "Massive Hit",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 0, Modifier: 100, DamageType: core.DamageSlashing},
			},
		}

		err := ed.Adjudicator.resolveDamage(attacker, halfOrc, action, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if halfOrc.StateManager.CurrentHP != 1 {
			t.Errorf("Expected Half-Orc to have 1 HP, got %d", halfOrc.StateManager.CurrentHP)
		}

		// Second hit should kill it because it's once per day
		halfOrc.StateManager.CurrentHP = 5
		err = ed.Adjudicator.resolveDamage(attacker, halfOrc, action, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if halfOrc.StateManager.CurrentHP != 0 {
			t.Errorf("Expected Half-Orc to have 0 HP on second hit, got %d", halfOrc.StateManager.CurrentHP)
		}
	})
}
