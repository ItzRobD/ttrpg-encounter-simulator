package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg_old/classes"
	"dnd5e-encounter-simulator-backend/pkg_old/races"
	"testing"
)

func TestFeatureProcessor_UnarmoredDefense(t *testing.T) {
	// Barbarian: 10 + Dex + Con
	cfg := actor.ActorConfig{
		ActorType: core.ActorTypeCharacter,
		Metadata: actor.Metadata{
			ClassID: classes.ClassID(core.Barbarian), // 2
		},
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{
				Dexterity:    14, // +2
				Constitution: 16, // +3
			},
		},
	}

	a, err := actor.NewActorFromConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to create actor: %v", err)
	}

	// Should be 10 + 2 + 3 = 15
	a.ProcessFeatures()
	if a.AC != 15 {
		t.Errorf("Expected AC 15 for Barbarian, got %d", a.AC)
	}

	// Monk: 10 + Dex + Wis
	cfg.Metadata.ClassID = classes.ClassID(core.Monk) // 7
	cfg.Abilities.AbilityScores.Wisdom = 18           // +4
	a, _ = actor.NewActorFromConfig(&cfg)
	a.ProcessFeatures()
	// Should be 10 + 2 + 4 = 16
	if a.AC != 16 {
		t.Errorf("Expected AC 16 for Monk, got %d", a.AC)
	}
}

func TestFeatureProcessor_TieflingResistance(t *testing.T) {
	features := []core.Feature{
		{
			Name: core.SpecAbilityHellishResistance,
		},
	}
	cfg := actor.ActorConfig{
		ActorType: core.ActorTypeCharacter,
		Metadata: actor.Metadata{
			RaceID: races.RaceID(core.Tiefling),
		},
		Features: features,
	}

	a, _ := actor.NewActorFromConfig(&cfg)
	a.ProcessFeatures()

	if a.Resistances.GetResistanceType(core.DamageFire) != core.ResistanceResistant {
		t.Errorf("Expected fire resistance for Tiefling")
	}
}

func TestFeatureProcessor_DragonbornBreath(t *testing.T) {
	features := []core.Feature{
		{
			Name: core.SpecAbilityBreathWeapon,
			Data: core.FeatureData{
				NumberOfDice: 2,
				Die:          6,
				DamageType:   []core.DamageType{core.DamageFire},
			},
		},
		{
			Name: core.SpecAbilityDraconicResistance,
			Data: core.FeatureData{
				DamageType: []core.DamageType{core.DamageFire},
			},
		},
	}
	cfg := actor.ActorConfig{
		ActorType: core.ActorTypeCharacter,
		Metadata: actor.Metadata{
			RaceID: races.RaceID(core.Dragonborn),
		},
		Features: features,
	}

	a, _ := actor.NewActorFromConfig(&cfg)
	a.ProcessFeatures()

	found := false
	for _, act := range a.Actions {
		if act.Name == "Breath Weapon" {
			found = true
			if len(act.DiceBlock) == 0 {
				t.Errorf("Expected damage blocks for Breath Weapon")
			} else {
				db := act.DiceBlock[0]
				if db.NumberOfDice != 2 || db.Die != 6 || db.DamageType != core.DamageFire {
					t.Errorf("Unexpected damage block: %v", db)
				}
			}
			break
		}
	}

	if !found {
		t.Errorf("Expected Breath Weapon action for Dragonborn")
	}

	if a.Resistances.GetResistanceType(core.DamageFire) != core.ResistanceResistant {
		t.Errorf("Expected fire resistance for Dragonborn")
	}
}

func TestFeatureProcessor_HalflingLuckyHydration(t *testing.T) {
	// This test simulates what happens after HydrateRaceFeaturesSRD
	features := []core.Feature{
		{
			ID:          "42",
			Name:        core.SpecAbilityHalflingLucky,
			Description: "When you roll a 1 on an attack roll, ability check, or saving throw, you can rereoll the die.",
			Hooks: map[core.HookType]bool{
				core.HookOnSelfAttack:      true,
				core.HookOnSelfSavingThrow: true,
			},
			Data: core.FeatureData{
				RerollType:      "saving_throw, attack_roll",
				RerollThreshold: 1,
			},
		},
	}
	cfg := actor.ActorConfig{
		ActorType: core.ActorTypeCharacter,
		Metadata: actor.Metadata{
			RaceID: races.RaceID(core.Halfling),
		},
		Features: features,
	}

	a, err := actor.NewActorFromConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to create actor: %v", err)
	}

	// Verify the feature is present and has the correct hooks/data
	found := false
	for _, f := range a.Features {
		if f.Name == core.SpecAbilityHalflingLucky {
			found = true
			if !f.Hooks[core.HookOnSelfAttack] {
				t.Errorf("Expected HookOnSelfAttack to be true")
			}
			if !f.Hooks[core.HookOnSelfSavingThrow] {
				t.Errorf("Expected HookOnSelfSavingThrow to be true")
			}
			if f.Data.RerollThreshold != 1 {
				t.Errorf("Expected RerollThreshold to be 1, got %d", f.Data.RerollThreshold)
			}
			if f.Data.RerollType != "saving_throw, attack_roll" {
				t.Errorf("Expected RerollType to be 'saving_throw, attack_roll', got '%s'", f.Data.RerollType)
			}
		}
	}

	if !found {
		t.Errorf("Halfling Lucky feature not found")
	}

	// ProcessFeatures should run without error, even if it doesn't do anything specific yet for Lucky
	// (as Lucky is a hook-based feature handled during rolls)
	a.ProcessFeatures()
}

func TestFeatureProcessor_BarbarianRage(t *testing.T) {
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

	resistances := a.GetResistances()
	if resistances.GetResistanceType(core.DamageSlashing) != core.ResistanceResistant {
		t.Errorf("Expected slashing resistance while raging")
	}
	if resistances.GetResistanceType(core.DamageBludgeoning) != core.ResistanceResistant {
		t.Errorf("Expected bludgeoning resistance while raging")
	}
	if resistances.GetResistanceType(core.DamagePiercing) != core.ResistanceResistant {
		t.Errorf("Expected piercing resistance while raging")
	}

	// Test non-raging
	a, _ = actor.NewActorFromConfig(&cfg)
	a.StateManager.IsRaging = false
	resistances = a.GetResistances()
	if resistances.GetResistanceType(core.DamageSlashing) != core.ResistanceNone {
		t.Errorf("Expected no slashing resistance while not raging")
	}
}

func TestFeatureProcessor_Resources(t *testing.T) {
	features := []core.Feature{
		{
			Name: core.SpecAbilityIndomitable,
			Data: core.FeatureData{Value: 1},
		},
		{
			Name: core.SpecAbilityLayOnHands,
			Data: core.FeatureData{Value: 10},
		},
	}
	cfg := actor.ActorConfig{
		ActorType: core.ActorTypeCharacter,
		Features:  features,
	}

	a, _ := actor.NewActorFromConfig(&cfg)
	a.Metadata.Level = 2
	a.ProcessFeatures()

	if a.StateManager.Resource[string(core.SpecAbilityIndomitable)] != 1 {
		t.Errorf("Expected 1 use of Indomitable, got %d", a.StateManager.Resource[string(core.SpecAbilityIndomitable)])
	}
	if a.StateManager.Resource[string(core.SpecAbilityLayOnHands)] != 10 {
		t.Errorf("Expected 10 points of Lay on Hands, got %d", a.StateManager.Resource[string(core.SpecAbilityLayOnHands)])
	}
}
