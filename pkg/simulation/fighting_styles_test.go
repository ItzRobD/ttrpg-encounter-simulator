package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/equipment"
	"testing"
)

func TestHandleFightingStyle(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	t.Run("Archery", func(t *testing.T) {
		a := &actor.Actor{
			InstanceID: 1,
			Equipment:  equipment_manager.NewEquipmentManager(),
		}
		f := core.Feature{Name: core.SpecAbilityFightingStyleArchery}

		// Ranged weapon
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				Action:           &core.Action{ActionType: core.ATRanged},
				AttackRoll:       &roll_manager.RollOptions{Modifier: 0},
				WeaponProperties: &core.WeaponProperties{IsRanged: true},
			},
		}
		ed.HandleFightingStyle(a, f, core.HookOnSelfAttack, ctx)
		if ctx.AttackContext.AttackRoll.Modifier != 2 {
			t.Errorf("Expected +2 attack bonus for Archery, got %d", ctx.AttackContext.AttackRoll.Modifier)
		}

		// Melee weapon
		ctx = &FeatureContext{
			AttackContext: &AttackContext{
				Action:           &core.Action{ActionType: core.ATMelee},
				AttackRoll:       &roll_manager.RollOptions{Modifier: 0},
				WeaponProperties: &core.WeaponProperties{IsRanged: false},
			},
		}
		ed.HandleFightingStyle(a, f, core.HookOnSelfAttack, ctx)
		if ctx.AttackContext.AttackRoll.Modifier != 0 {
			t.Errorf("Expected no bonus for Archery with melee weapon, got %d", ctx.AttackContext.AttackRoll.Modifier)
		}
	})

	t.Run("Dueling", func(t *testing.T) {
		a := &actor.Actor{
			InstanceID: 1,
			Equipment:  equipment_manager.NewEquipmentManager(),
		}
		f := core.Feature{Name: core.SpecAbilityFightingStyleDuel}

		// One-handed melee weapon, no offhand
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				Action:           &core.Action{ActionType: core.ATMelee},
				WeaponModifiers:  &core.WeaponModifiers{DamageBonus: 0},
				WeaponProperties: &core.WeaponProperties{IsRanged: false, IsTwoHanded: false},
			},
		}
		ed.HandleFightingStyle(a, f, core.HookOnSelfHit, ctx)
		if ctx.AttackContext.WeaponModifiers.DamageBonus != 2 {
			t.Errorf("Expected +2 damage bonus for Dueling, got %d", ctx.AttackContext.WeaponModifiers.DamageBonus)
		}

		// Two-handed melee weapon
		ctx = &FeatureContext{
			AttackContext: &AttackContext{
				Action:           &core.Action{ActionType: core.ATMelee},
				WeaponModifiers:  &core.WeaponModifiers{DamageBonus: 0},
				WeaponProperties: &core.WeaponProperties{IsRanged: false, IsTwoHanded: true},
			},
		}
		ed.HandleFightingStyle(a, f, core.HookOnSelfHit, ctx)
		if ctx.AttackContext.WeaponModifiers.DamageBonus != 0 {
			t.Errorf("Expected no bonus for Dueling with two-handed weapon, got %d", ctx.AttackContext.WeaponModifiers.DamageBonus)
		}

		// One-handed melee weapon, but with offhand weapon
		a.Equipment.AddItem("secondary", equipment.Equipment{
			Type: equipment.EquipmentTypeWeapon,
			Weapon: &equipment.Weapon{
				DamageBlocks: []core.DiceBlock{{NumberOfDice: 1, Die: core.D6, DamageType: core.DamagePiercing}},
			},
		})
		ctx = &FeatureContext{
			AttackContext: &AttackContext{
				Action:           &core.Action{ActionType: core.ATMelee},
				WeaponModifiers:  &core.WeaponModifiers{DamageBonus: 0},
				WeaponProperties: &core.WeaponProperties{IsRanged: false, IsTwoHanded: false},
			},
		}
		ed.HandleFightingStyle(a, f, core.HookOnSelfHit, ctx)
		if ctx.AttackContext.WeaponModifiers.DamageBonus != 0 {
			t.Errorf("Expected no bonus for Dueling with offhand weapon, got %d", ctx.AttackContext.WeaponModifiers.DamageBonus)
		}
	})

	t.Run("Great Weapon Fighting", func(t *testing.T) {
		a := &actor.Actor{
			InstanceID: 1,
			Equipment:  equipment_manager.NewEquipmentManager(),
		}
		f := core.Feature{Name: core.SpecAbilityFightingStyleGWF}

		// Two-handed melee weapon
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				Action:           &core.Action{ActionType: core.ATMelee},
				DamageRoll:       &roll_manager.RollOptions{RerollThreshold: 0},
				WeaponProperties: &core.WeaponProperties{IsRanged: false, IsTwoHanded: true},
			},
		}
		ed.HandleFightingStyle(a, f, core.HookOnSelfAttack, ctx)
		if ctx.AttackContext.DamageRoll.RerollThreshold != 2 {
			t.Errorf("Expected reroll threshold 2 for GWF, got %d", ctx.AttackContext.DamageRoll.RerollThreshold)
		}

		// One-handed melee weapon
		ctx = &FeatureContext{
			AttackContext: &AttackContext{
				Action:           &core.Action{ActionType: core.ATMelee},
				DamageRoll:       &roll_manager.RollOptions{RerollThreshold: 0},
				WeaponProperties: &core.WeaponProperties{IsRanged: false, IsTwoHanded: false},
			},
		}
		ed.HandleFightingStyle(a, f, core.HookOnSelfAttack, ctx)
		if ctx.AttackContext.DamageRoll.RerollThreshold != 0 {
			t.Errorf("Expected no reroll threshold for GWF with one-handed weapon, got %d", ctx.AttackContext.DamageRoll.RerollThreshold)
		}
	})

	t.Run("Two-Weapon Fighting", func(t *testing.T) {
		a := &actor.Actor{
			InstanceID: 1,
			Abilities: core.Abilities{
				AbilityScores: core.AbilityScores{Strength: 16}, // +3 mod
			},
			Equipment: equipment_manager.NewEquipmentManager(),
		}
		f := core.Feature{Name: core.SpecAbilityFightingStyleTWF}

		// Offhand attack
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				Action:           &core.Action{ActionType: core.ATOffhand},
				WeaponModifiers:  &core.WeaponModifiers{DamageBonus: 0},
				WeaponProperties: &core.WeaponProperties{IsRanged: false, IsFinesse: false},
			},
		}
		ed.HandleFightingStyle(a, f, core.HookOnSelfHit, ctx)
		if ctx.AttackContext.WeaponModifiers.DamageBonus != 3 {
			t.Errorf("Expected +3 damage bonus for TWF, got %d", ctx.AttackContext.WeaponModifiers.DamageBonus)
		}

		// Main hand attack
		ctx = &FeatureContext{
			AttackContext: &AttackContext{
				Action:           &core.Action{ActionType: core.ATMelee},
				WeaponModifiers:  &core.WeaponModifiers{DamageBonus: 0},
				WeaponProperties: &core.WeaponProperties{IsRanged: false, IsFinesse: false},
			},
		}
		ed.HandleFightingStyle(a, f, core.HookOnSelfHit, ctx)
		if ctx.AttackContext.WeaponModifiers.DamageBonus != 0 {
			t.Errorf("Expected no bonus for TWF on main hand attack, got %d", ctx.AttackContext.WeaponModifiers.DamageBonus)
		}
	})
}
