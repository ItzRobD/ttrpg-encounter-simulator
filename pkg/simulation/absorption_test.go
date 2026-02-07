package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleAbsorption(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Golem with Lightning Absorption
	golem := &actor.Actor{
		InstanceID: 1,
		Name:       "Flesh Golem",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     100,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityAbsorption,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfDamageTaken: true,
				},
				Data: core.FeatureData{
					DamageType: []core.DamageType{core.DamageLightning},
				},
			},
		},
	}

	// Wizard (Attacker)
	wizard := &actor.Actor{
		InstanceID: 2,
		Name:       "Wizard",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
	}

	ed.Actors[1] = golem
	ed.Actors[2] = wizard
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Absorb Lightning and Heal", func(t *testing.T) {
		// Shocking Grasp-like action
		lightningAction := core.Action{
			Name: "Lightning Bolt",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 8, Die: core.D6, DamageType: core.DamageLightning}, // ~28 damage
			},
		}

		golem.StateManager.CurrentHP = 50
		oldHP := golem.StateManager.CurrentHP

		err := ed.Adjudicator.resolveDamage(wizard, golem, &lightningAction, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if golem.StateManager.CurrentHP <= oldHP {
			t.Errorf("Expected HP to increase due to absorption, but it was %d (old: %d)", golem.StateManager.CurrentHP, oldHP)
		}

		if golem.StateManager.CurrentHP > oldHP+48 { // 8d6 max is 48
			t.Errorf("HP increased too much: %d", golem.StateManager.CurrentHP)
		}

		t.Logf("Glow HP after lightning: %d (Healed: %d)", golem.StateManager.CurrentHP, golem.StateManager.CurrentHP-oldHP)
	})

	t.Run("Take Normal Damage for Other Type", func(t *testing.T) {
		fireAction := core.Action{
			Name: "Fireball",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 8, Die: core.D6, DamageType: core.DamageFire},
			},
		}

		golem.StateManager.CurrentHP = 50
		oldHP := golem.StateManager.CurrentHP

		err := ed.Adjudicator.resolveDamage(wizard, golem, &fireAction, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if golem.StateManager.CurrentHP >= oldHP {
			t.Errorf("Expected HP to decrease due to fire damage, but it was %d (old: %d)", golem.StateManager.CurrentHP, oldHP)
		}

		t.Logf("Glow HP after fire: %d (Took: %d damage)", golem.StateManager.CurrentHP, oldHP-golem.StateManager.CurrentHP)
	})
}
