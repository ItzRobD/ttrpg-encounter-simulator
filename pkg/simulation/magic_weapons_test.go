package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleMagicWeapons(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Monster with Magic Weapons (e.g. Balor, Marilith)
	monster := &actor.Actor{
		InstanceID: 1,
		Name:       "Magic Monster",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		Features: []core.Feature{
			{
				Name: core.SpecAbilityMagicWeapons,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfAttack: true,
				},
			},
		},
	}

	// Target with resistance to non-magic bludgeoning damage (e.g. Golem)
	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Resistant Target",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
		},
		Resistances: core.NewDamageResistances(),
	}
	// Resistance to non-magic bludgeoning
	target.Resistances.SetResistanceType(core.DamageBludgeoning, core.ResistanceResistant)

	ed.Actors[1] = monster
	ed.Actors[2] = target
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	// Melee Weapon Action
	weaponAction := &core.Action{
		Name:       "Slam",
		ActionType: core.ATMelee,
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 2, Die: core.D10, Modifier: 0, DamageType: core.DamageBludgeoning},
		},
		WeaponModifiers: &core.WeaponModifiers{
			IsMagic: false, // Start as non-magic
		},
	}

	t.Run("Magic Weapons Feature Makes Attacks Magic", func(t *testing.T) {
		// Reset target HP
		target.StateManager.CurrentHP = 100

		// executeIndividualStrike should trigger the hook which sets IsMagic = true
		// Then resolveDamage should apply full damage because it's now magic
		err := ed.Adjudicator.executeIndividualStrike(monster, target, weaponAction)
		if err != nil {
			t.Fatalf("executeIndividualStrike failed: %v", err)
		}

		if !weaponAction.WeaponModifiers.IsMagic {
			t.Error("Expected weapon action to be magic after hook dispatch")
		}

		// Full damage check. 2d10 average is 11.
		// If it was resisted, damage would be ~5.
		damageDealt := 100 - target.StateManager.CurrentHP
		if damageDealt < 10 {
			t.Errorf("Expected full damage (magic), but damage dealt was %d", damageDealt)
		}
		t.Logf("Damage dealt with Magic Weapons: %d", damageDealt)
	})

	t.Run("Non-Weapon Attacks Not Affected", func(t *testing.T) {
		// Reset target HP
		target.StateManager.CurrentHP = 100

		spellAction := &core.Action{
			Name:       "Firebolt",
			ActionType: core.ATSpell,
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 1, Die: core.D10, Modifier: 0, DamageType: core.DamageFire},
			},
			WeaponModifiers: &core.WeaponModifiers{
				IsMagic: false,
			},
		}

		err := ed.Adjudicator.executeIndividualStrike(monster, target, spellAction)
		if err != nil {
			t.Fatalf("executeIndividualStrike failed: %v", err)
		}

		if spellAction.WeaponModifiers.IsMagic {
			t.Error("Expected spell action to NOT be affected by Magic Weapons feature")
		}
	})
}
