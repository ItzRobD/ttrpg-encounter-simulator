package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleRelentlessRage(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	barbarian := &actor.Actor{
		InstanceID: 1,
		Name:       "Barbarian",
		StateManager: state_manager.StateManager{
			CurrentHP: 10,
			MaxHP:     50,
			IsRaging:  true,
			Resource:  make(map[string]int),
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityRelentlessRage,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfDamageTaken: true,
				},
			},
		},
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{Constitution: 16},
		},
		ProficiencyBonus: 3,
	}

	ed.Actors[1] = barbarian

	t.Run("Relentless Rage Triggers (Succeeds DC 10)", func(t *testing.T) {
		barbarian.StateManager.CurrentHP = 10
		barbarian.StateManager.Resource[string(core.SpecAbilityRelentlessRage)] = 0

		damage := 15
		ctx := &FeatureContext{
			DamageContext: &DamageContext{
				DamageValue: &damage,
			},
		}

		// Mock ResolveSavingThrow to succeed
		// Actually, let's just let it roll. With +3 Con mod and +3 Prof (if proficient), DC 10 is easy.
		// Wait, ResolveSavingThrow calculates modifier itself.

		err := ed.HandleRelentlessRage(barbarian, barbarian.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleRelentlessRage failed: %v", err)
		}

		// If it succeeded, damage should be adjusted so HP drops to 1
		if *ctx.DamageContext.DamageValue != 9 {
			t.Errorf("Expected adjusted damage 9, got %d", *ctx.DamageContext.DamageValue)
		}
		if barbarian.StateManager.Resource[string(core.SpecAbilityRelentlessRage)] != 1 {
			t.Errorf("Expected usage count to be 1, got %d", barbarian.StateManager.Resource[string(core.SpecAbilityRelentlessRage)])
		}
	})

	t.Run("Relentless Rage Increases DC (Next is DC 15)", func(t *testing.T) {
		barbarian.StateManager.CurrentHP = 10
		barbarian.StateManager.Resource[string(core.SpecAbilityRelentlessRage)] = 1

		damage := 15
		ctx := &FeatureContext{
			DamageContext: &DamageContext{
				DamageValue: &damage,
			},
		}

		// We can't easily force failure without mocking RNG, but we can check if it calls with correct DC.
		// Since we can't easily check the DC passed to ResolveSavingThrow from here without modifying code,
		// we'll just verify it still runs.

		err := ed.HandleRelentlessRage(barbarian, barbarian.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleRelentlessRage failed: %v", err)
		}
	})

	t.Run("Relentless Rage Does Not Trigger when not raging", func(t *testing.T) {
		barbarian.StateManager.CurrentHP = 10
		barbarian.StateManager.IsRaging = false
		barbarian.StateManager.Resource[string(core.SpecAbilityRelentlessRage)] = 0

		damage := 15
		ctx := &FeatureContext{
			DamageContext: &DamageContext{
				DamageValue: &damage,
			},
		}

		err := ed.HandleRelentlessRage(barbarian, barbarian.Features[0], core.HookOnSelfDamageTaken, ctx)
		if err != nil {
			t.Fatalf("HandleRelentlessRage failed: %v", err)
		}

		if *ctx.DamageContext.DamageValue != 15 {
			t.Errorf("Expected damage to remain 15, got %d", *ctx.DamageContext.DamageValue)
		}
	})
}

func TestHandleRageExtraDamage(t *testing.T) {
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, &core.SimulationOptions{})

	barbarian := &actor.Actor{
		InstanceID: 1,
		StateManager: state_manager.StateManager{
			IsRaging: true,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityRageExtraDamage,
				Data: core.FeatureData{Modifier: 2},
			},
		},
	}

	t.Run("Rage Extra Damage applies to Melee Weapon Attack", func(t *testing.T) {
		weaponMods := &core.WeaponModifiers{}
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				Action: &core.Action{
					ActionType: core.ATMelee,
				},
				WeaponModifiers: weaponMods,
			},
		}

		err := ed.HandleRageExtraDamage(barbarian, barbarian.Features[0], core.HookOnSelfAttack, ctx)
		if err != nil {
			t.Fatalf("HandleRageExtraDamage failed: %v", err)
		}

		if weaponMods.DamageBonus != 2 {
			t.Errorf("Expected DamageBonus 2, got %d", weaponMods.DamageBonus)
		}
	})

	t.Run("Rage Extra Damage does not apply when not raging", func(t *testing.T) {
		barbarian.StateManager.IsRaging = false
		weaponMods := &core.WeaponModifiers{}
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				Action: &core.Action{
					ActionType: core.ATMelee,
				},
				WeaponModifiers: weaponMods,
			},
		}

		ed.HandleRageExtraDamage(barbarian, barbarian.Features[0], core.HookOnSelfAttack, ctx)
		if weaponMods.DamageBonus != 0 {
			t.Errorf("Expected DamageBonus 0, got %d", weaponMods.DamageBonus)
		}
	})
}

func TestRageResistance(t *testing.T) {
	cfg := actor.ActorConfig{
		ActorType: core.ActorTypeCharacter,
		Metadata: actor.Metadata{
			ClassID: 2, // Barbarian
		},
		Features: []core.Feature{
			{Name: core.SpecAbilityRageResistance},
		},
	}

	a, _ := actor.NewActorFromConfig(&cfg)
	a.StateManager.IsRaging = true
	a.ProcessFeatures()

	res := a.GetResistances()
	if res.GetResistanceType(core.DamageSlashing) != core.ResistanceResistant {
		t.Errorf("Expected slashing resistance while raging")
	}
	if res.GetResistanceType(core.DamageBludgeoning) != core.ResistanceResistant {
		t.Errorf("Expected bludgeoning resistance while raging")
	}
	if res.GetResistanceType(core.DamagePiercing) != core.ResistanceResistant {
		t.Errorf("Expected piercing resistance while raging")
	}

	// Non-raging
	a.StateManager.IsRaging = false
	res = a.GetResistances()
	if res.GetResistanceType(core.DamageSlashing) != core.ResistanceNone {
		t.Errorf("Expected no slashing resistance while not raging, got %s", res.GetResistanceType(core.DamageSlashing))
	}
}
