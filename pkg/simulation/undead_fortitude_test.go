package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleUndeadFortitude(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Zombie with Undead Fortitude
	zombie := &actor.Actor{
		InstanceID: 1,
		Name:       "Zombie",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{
				Constitution: 16, // +3
			},
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 22,
			MaxHP:     22,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityUndeadFortitude,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfDamageTaken: true,
				},
				Data: core.FeatureData{
					DamageType: []core.DamageType{core.DamageRadiant}, // Radiant bypasses fortitude
				},
			},
		},
	}

	attacker := &actor.Actor{
		InstanceID: 2,
		Name:       "Cleric",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
	}

	ed.Actors[1] = zombie
	ed.Actors[2] = attacker
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Successful Save keeps zombie at 1 HP", func(t *testing.T) {
		zombie.StateManager.CurrentHP = 5
		damage := 10
		isCrit := false

		// DC will be 5 + 10 = 15. Zombie has +3 Con mod. Needs 12+ on d20.
		// With Seed 1,1, let's see.
		// Actually, I can just set the DC to 0 to ensure success if I want to test logic.

		ctx := &FeatureContext{
			Target: attacker,
			DamageContext: &DamageContext{
				DamageValue: &damage,
				DamageType:  core.DamageBludgeoning,
			},
			AttackContext: &AttackContext{
				IsCritical: &isCrit,
				Action: &core.Action{
					Name: "Mace",
				},
			},
		}

		// Force success by setting high Con or low DC
		zombie.Abilities.AbilityScores.Constitution = 100 // Ridiculous mod

		err := ed.HandleUndeadFortitude(zombie, zombie.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleUndeadFortitude failed: %v", err)
		}

		if *ctx.DamageContext.DamageValue != 4 { // currentHP (5) - 1 = 4
			t.Errorf("Expected damage to be adjusted to 4, got %d", *ctx.DamageContext.DamageValue)
		}
	})

	t.Run("Failed Save kills zombie", func(t *testing.T) {
		zombie.StateManager.CurrentHP = 5
		damage := 10
		isCrit := false

		ctx := &FeatureContext{
			Target: attacker,
			DamageContext: &DamageContext{
				DamageValue: &damage,
				DamageType:  core.DamageBludgeoning,
			},
			AttackContext: &AttackContext{
				IsCritical: &isCrit,
				Action: &core.Action{
					Name: "Mace",
				},
			},
		}

		// Force failure
		zombie.Abilities.AbilityScores.Constitution = 1 // -5 mod
		// DC 15. Impossible to succeed.

		err := ed.HandleUndeadFortitude(zombie, zombie.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleUndeadFortitude failed: %v", err)
		}

		if *ctx.DamageContext.DamageValue != 10 {
			t.Errorf("Expected damage to remain 10, got %d", *ctx.DamageContext.DamageValue)
		}
	})

	t.Run("Radiant Damage bypasses Fortitude", func(t *testing.T) {
		zombie.StateManager.CurrentHP = 5
		damage := 10
		isCrit := false

		ctx := &FeatureContext{
			Target: attacker,
			DamageContext: &DamageContext{
				DamageValue: &damage,
				DamageType:  core.DamageRadiant,
			},
			AttackContext: &AttackContext{
				IsCritical: &isCrit,
				Action: &core.Action{
					Name: "Sacred Flame",
				},
			},
		}

		zombie.Abilities.AbilityScores.Constitution = 100

		err := ed.HandleUndeadFortitude(zombie, zombie.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleUndeadFortitude failed: %v", err)
		}

		if *ctx.DamageContext.DamageValue != 10 {
			t.Errorf("Expected damage to remain 10 (radiant bypass), got %d", *ctx.DamageContext.DamageValue)
		}
	})

	t.Run("Critical Hit bypasses Fortitude", func(t *testing.T) {
		zombie.StateManager.CurrentHP = 5
		damage := 10
		isCrit := true

		ctx := &FeatureContext{
			Target: attacker,
			DamageContext: &DamageContext{
				DamageValue: &damage,
				DamageType:  core.DamageBludgeoning,
			},
			AttackContext: &AttackContext{
				IsCritical: &isCrit,
				Action: &core.Action{
					Name: "Mace",
				},
			},
		}

		zombie.Abilities.AbilityScores.Constitution = 100

		err := ed.HandleUndeadFortitude(zombie, zombie.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleUndeadFortitude failed: %v", err)
		}

		if *ctx.DamageContext.DamageValue != 10 {
			t.Errorf("Expected damage to remain 10 (critical bypass), got %d", *ctx.DamageContext.DamageValue)
		}
	})

	t.Run("Non-lethal damage does not trigger Fortitude", func(t *testing.T) {
		zombie.StateManager.CurrentHP = 20
		damage := 10
		isCrit := false

		ctx := &FeatureContext{
			Target: attacker,
			DamageContext: &DamageContext{
				DamageValue: &damage,
				DamageType:  core.DamageBludgeoning,
			},
			AttackContext: &AttackContext{
				IsCritical: &isCrit,
				Action: &core.Action{
					Name: "Mace",
				},
			},
		}

		zombie.Abilities.AbilityScores.Constitution = 100

		err := ed.HandleUndeadFortitude(zombie, zombie.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleUndeadFortitude failed: %v", err)
		}

		if *ctx.DamageContext.DamageValue != 10 {
			t.Errorf("Expected damage to remain 10 (not lethal), got %d", *ctx.DamageContext.DamageValue)
		}
	})
}
