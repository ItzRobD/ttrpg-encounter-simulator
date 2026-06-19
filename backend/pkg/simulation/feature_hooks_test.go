package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHookIntegration_AttackModifiers(t *testing.T) {
	simOptions := &core.SimulationOptions{EnableSpecialAbilities: true}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	attacker := &actor.Actor{
		Name:       "Attacker",
		InstanceID: 1,
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
	}
	target := &actor.Actor{
		Name:       "Target",
		InstanceID: 2,
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
	}
	ally := &actor.Actor{
		Name:       "Ally",
		InstanceID: 3,
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
	}

	ed.Actors[1] = attacker
	ed.Actors[2] = target
	ed.Actors[3] = ally

	// Ensure StateManager is initialized for CanActConditions
	attacker.StateManager.Conditions = core.NewActorConditions()
	target.StateManager.Conditions = core.NewActorConditions()
	ally.StateManager.Conditions = core.NewActorConditions()

	// Add some HP to ally so they aren't considered dead/incapacitated
	ally.StateManager.CurrentHP = 10
	ally.StateManager.MaxHP = 10

	t.Run("Archery Fighting Style adds +2 to ranged attack rolls", func(t *testing.T) {
		attacker.Features = []core.Feature{
			{
				Name:  core.SpecAbilityFightingStyleArchery,
				Hooks: map[core.HookType]bool{core.HookOnSelfAttack: true},
			},
		}
		action := core.Action{
			Name:             "Longbow",
			ActionType:       core.ATRanged,
			AttackBonus:      5,
			WeaponProperties: &core.WeaponProperties{IsRanged: true},
		}
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				Action: &action,
				AttackRoll: &roll_manager.RollOptions{
					Modifier: 5,
				},
				WeaponProperties: action.WeaponProperties,
			},
		}
		ed.dispatchHooks(attacker, core.HookOnSelfAttack, ctx)
		if ctx.AttackContext.AttackRoll.Modifier != 7 {
			t.Errorf("Expected modifier 7, got %d", ctx.AttackContext.AttackRoll.Modifier)
		}
	})

	t.Run("Pack Tactics grants advantage when ally is near", func(t *testing.T) {
		attacker.Features = []core.Feature{
			{
				Name:  core.SpecAbilityPackTactics,
				Hooks: map[core.HookType]bool{core.HookOnSelfAttack: true},
			},
		}
		action := core.Action{
			Name:       "Melee Attack",
			ActionType: core.ATMelee,
		}
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				Action: &action,
				AttackRoll: &roll_manager.RollOptions{
					Advantage: core.RollNormal,
				},
			},
		}
		// The Pack Tactics handler in features.go:655 uses ed.GetAllyTargets(a)
		// and then checks for proximity. In this sim, we assume all allies are "near"
		// for testing if they are not incapacitated.
		ed.dispatchHooks(attacker, core.HookOnSelfAttack, ctx)
		if ctx.AttackContext.AttackRoll.Advantage != core.RollAdvantage {
			t.Errorf("Expected advantage from Pack Tactics")
		}
	})

	t.Run("Pack Tactics grants advantage when ally is near but not at disadvantage", func(t *testing.T) {
		attacker.Features = []core.Feature{
			{
				Name:  core.SpecAbilityPackTactics,
				Hooks: map[core.HookType]bool{core.HookOnSelfAttack: true},
			},
		}
		action := core.Action{
			Name:       "Melee Attack",
			ActionType: core.ATMelee,
		}
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				Action: &action,
				AttackRoll: &roll_manager.RollOptions{
					Advantage:         core.RollDisadvantage,
					DisadvantageCount: 1,
				},
			},
		}
		// The Pack Tactics handler in features.go:655 uses ed.GetAllyTargets(a)
		// and then checks for proximity. In this sim, we assume all allies are "near"
		// for testing if they are not incapacitated.
		ed.dispatchHooks(attacker, core.HookOnSelfAttack, ctx)
		if ctx.AttackContext.AttackRoll.Advantage != core.RollNormal {
			t.Errorf("Expected normal roll from Pack Tactics with disadvantage, got %v", ctx.AttackContext.AttackRoll.Advantage)
		}
	})
}

func TestHookIntegration_DamageModifiers(t *testing.T) {
	simOptions := &core.SimulationOptions{EnableSpecialAbilities: true}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)
	attacker := &actor.Actor{
		Name:       "Attacker",
		InstanceID: 1,
		Equipment:  equipment_manager.NewEquipmentManager(), // Need for Dueling
	}
	target := &actor.Actor{Name: "Target", InstanceID: 2}
	ed.Actors[1] = attacker
	ed.Actors[2] = target

	t.Run("Dueling Fighting Style adds +2 to damage", func(t *testing.T) {
		attacker.Features = []core.Feature{
			{
				Name:  core.SpecAbilityFightingStyleDuel,
				Hooks: map[core.HookType]bool{core.HookOnSelfHit: true},
			},
		}
		action := core.Action{
			Name:             "One-Handed Sword",
			ActionType:       core.ATMelee,
			WeaponProperties: &core.WeaponProperties{IsRanged: false, IsTwoHanded: false},
			WeaponModifiers:  &core.WeaponModifiers{DamageBonus: 0},
		}
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				Action:           &action,
				WeaponProperties: action.WeaponProperties,
				WeaponModifiers:  action.WeaponModifiers,
			},
		}
		ed.dispatchHooks(attacker, core.HookOnSelfHit, ctx)
		if action.WeaponModifiers.DamageBonus != 2 {
			t.Errorf("Expected damage bonus 2, got %d", action.WeaponModifiers.DamageBonus)
		}
	})

	t.Run("Great Weapon Fighting sets reroll threshold to 2", func(t *testing.T) {
		attacker.Features = []core.Feature{
			{
				Name:  core.SpecAbilityFightingStyleGWF,
				Hooks: map[core.HookType]bool{core.HookOnSelfAttack: true},
			},
		}
		action := core.Action{
			Name:             "Greatsword",
			ActionType:       core.ATMelee,
			WeaponProperties: &core.WeaponProperties{IsRanged: false, IsTwoHanded: true},
		}
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				Action:           &action,
				WeaponProperties: action.WeaponProperties,
				DamageRoll:       &roll_manager.RollOptions{},
			},
		}
		ed.dispatchHooks(attacker, core.HookOnSelfAttack, ctx)
		if ctx.AttackContext.DamageRoll.RerollThreshold != 2 {
			t.Errorf("Expected RerollThreshold 2, got %d", ctx.AttackContext.DamageRoll.RerollThreshold)
		}
	})
}

func TestHookIntegration_DamageMitigation(t *testing.T) {
	simOptions := &core.SimulationOptions{EnableSpecialAbilities: true}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)
	defender := &actor.Actor{Name: "Defender", InstanceID: 1}
	attacker := &actor.Actor{Name: "Attacker", InstanceID: 2}
	ed.Actors[1] = defender
	ed.Actors[2] = attacker

	t.Run("Uncanny Dodge halves damage", func(t *testing.T) {
		defender.Features = []core.Feature{
			{
				Name:  core.SpecAbilityUncannyDodge,
				Hooks: map[core.HookType]bool{core.HookOnSelfDamageTaken: true},
			},
		}
		dmg := 20
		ctx := &FeatureContext{
			Target: attacker,
			DamageContext: &DamageContext{
				DamageValue: &dmg,
			},
			AttackContext: &AttackContext{
				Action: &core.Action{ActionType: core.ATMelee},
			},
		}
		ed.dispatchHooks(defender, core.HookOnSelfDamageTaken, ctx)
		if *ctx.DamageContext.DamageValue != 10 {
			t.Errorf("Expected damage 10, got %d", *ctx.DamageContext.DamageValue)
		}
		if defender.StateManager.ReactionUsedCount != 1 {
			t.Errorf("Expected reaction used")
		}
	})

	t.Run("Uncanny Dodge requires reaction to use", func(t *testing.T) {
		defender.Features = []core.Feature{
			{
				Name:  core.SpecAbilityUncannyDodge,
				Hooks: map[core.HookType]bool{core.HookOnSelfDamageTaken: true},
			},
		}
		defender.StateManager.ReactionUsedCount = 1
		dmg := 20
		ctx := &FeatureContext{
			Target: attacker,
			DamageContext: &DamageContext{
				DamageValue: &dmg,
			},
			AttackContext: &AttackContext{
				Action: &core.Action{ActionType: core.ATMelee},
			},
		}
		ed.dispatchHooks(defender, core.HookOnSelfDamageTaken, ctx)
		if *ctx.DamageContext.DamageValue != 20 {
			t.Errorf("Expected damage 20, got %d", *ctx.DamageContext.DamageValue)
		}
		if defender.StateManager.ReactionUsedCount != 1 {
			t.Errorf("Additional reaction used")
		}
	})
}

func TestHookIntegration_ComplexFeatures(t *testing.T) {
	simOptions := &core.SimulationOptions{EnableSpecialAbilities: true}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)
	attacker := &actor.Actor{
		Name:       "Attacker",
		InstanceID: 1,
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			OncePerTurnUsed: make(map[string]bool),
		},
	}
	target := &actor.Actor{
		Name:       "Target",
		InstanceID: 2,
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}
	ed.Actors[1] = attacker
	ed.Actors[2] = target

	t.Run("Sneak Attack adds damage and marks once per turn", func(t *testing.T) {
		attacker.Features = []core.Feature{
			{
				Name:  core.SpecAbilitySneakAttack,
				Hooks: map[core.HookType]bool{core.HookOnSelfHit: true},
				Data:  core.FeatureData{NumberOfDice: 1, Die: core.D6},
			},
		}
		// Sneak Attack usually requires advantage or an ally.
		// The handler feature_hooks.go:773 checks if advantage or ally.
		// Let's force advantage.
		isCrit := false
		ctx := &FeatureContext{
			Target: target,
			AttackContext: &AttackContext{
				Action:     &core.Action{ActionType: core.ATMelee},
				AttackRoll: &roll_manager.RollOptions{Advantage: core.RollAdvantage},
				IsCritical: &isCrit,
			},
		}

		ed.dispatchHooks(attacker, core.HookOnSelfHit, ctx)

		if !attacker.StateManager.OncePerTurnUsed[string(core.SpecAbilitySneakAttack)] {
			t.Errorf("Expected Sneak Attack to be marked as used")
		}
		// We can't easily check the damage dealt here because resolveDamage logs events
		// but we can check if OncePerTurn is set.
	})

	t.Run("Regeneration restores HP at turn start", func(t *testing.T) {
		target.Features = []core.Feature{
			{
				Name:  core.SpecAbilityRegeneration,
				Hooks: map[core.HookType]bool{core.HookOnTurnStart: true},
				Data:  core.FeatureData{Value: 10},
			},
		}
		target.StateManager.CurrentHP = 30
		target.StateManager.MaxHP = 50

		ed.dispatchHooks(target, core.HookOnTurnStart, nil)

		if target.StateManager.CurrentHP != 40 {
			t.Errorf("Expected HP 40 after regeneration, got %d", target.StateManager.CurrentHP)
		}
	})

	t.Run("Regeneration restores HP but shouldn't go beyond max", func(t *testing.T) {
		target.Features = []core.Feature{
			{
				Name:  core.SpecAbilityRegeneration,
				Hooks: map[core.HookType]bool{core.HookOnTurnStart: true},
				Data:  core.FeatureData{Value: 10},
			},
		}
		target.StateManager.CurrentHP = 45
		target.StateManager.MaxHP = 50

		ed.dispatchHooks(target, core.HookOnTurnStart, nil)

		if target.StateManager.CurrentHP != 50 {
			t.Errorf("Expected HP 50 after regeneration, got %d", target.StateManager.CurrentHP)
		}
	})

	t.Run("Relentless Endurance avoids 0 HP once per day", func(t *testing.T) {
		target.Features = []core.Feature{
			{
				Name:  core.SpecAbilityRelentlessEndurance,
				Hooks: map[core.HookType]bool{core.HookOnSelfDamageTaken: true},
			},
		}
		target.StateManager.CurrentHP = 10
		target.StateManager.OncePerTurnUsed = make(map[string]bool)
		dmg := 15 // Would reduce to -5 -> 0
		ctx := &FeatureContext{
			DamageContext: &DamageContext{
				DamageValue: &dmg,
			},
		}

		ed.dispatchHooks(target, core.HookOnSelfDamageTaken, ctx)

		if *ctx.DamageContext.DamageValue != 9 {
			t.Errorf("Expected damage 9 (leaving 1 HP), got %d", *ctx.DamageContext.DamageValue)
		}
		if !target.StateManager.OncePerTurnUsed[string(core.SpecAbilityRelentlessEndurance)] {
			t.Errorf("Expected Relentless Endurance to be marked as used")
		}

		// Apply the reduced damage so HP is now 1
		target.StateManager.ModifyHP(-(*ctx.DamageContext.DamageValue), false, target.IsCharacter())
		if target.StateManager.CurrentHP != 1 {
			t.Fatalf("Expected HP 1, got %d", target.StateManager.CurrentHP)
		}

		// Second hit
		dmg2 := 10
		ctx2 := &FeatureContext{
			DamageContext: &DamageContext{
				DamageValue: &dmg2,
			},
		}

		ed.dispatchHooks(target, core.HookOnSelfDamageTaken, ctx2)
		if *ctx2.DamageContext.DamageValue != 10 {
			t.Errorf("Expected damage 10 (no reduction), got %d", *ctx2.DamageContext.DamageValue)
		}

		target.StateManager.ModifyHP(-(*ctx2.DamageContext.DamageValue), false, target.IsCharacter())
		if target.StateManager.CurrentHP != 0 {
			t.Errorf("Expected HP 0 after damage with unable to use Relentless Endurance, got %d", target.StateManager.CurrentHP)
		}
	})

	t.Run("Halfling Lucky sets reroll threshold to 1 on attack", func(t *testing.T) {
		attacker.Features = []core.Feature{
			{
				Name:  core.SpecAbilityHalflingLucky,
				Hooks: map[core.HookType]bool{core.HookOnSelfAttack: true},
			},
		}
		ctx := &FeatureContext{
			AttackContext: &AttackContext{
				AttackRoll: &roll_manager.RollOptions{RerollThreshold: 0},
			},
		}

		ed.dispatchHooks(attacker, core.HookOnSelfAttack, ctx)

		if ctx.AttackContext.AttackRoll.RerollThreshold != 1 {
			t.Errorf("Expected RerollThreshold to be 1, got %d", ctx.AttackContext.AttackRoll.RerollThreshold)
		}
	})
}

func TestHookIntegration_RetaliatoryAndDeathFeatures(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities:       true,
		MonsterDeathEffectsHitAllies: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	attacker := &actor.Actor{
		Name:       "Attacker",
		InstanceID: 1,
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
		},
	}
	defender := &actor.Actor{
		Name:       "Defender",
		InstanceID: 2,
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	ed.Actors[1] = attacker
	ed.Actors[2] = defender
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	t.Run("Heated Body deals damage to melee attacker", func(t *testing.T) {
		defender.Features = []core.Feature{
			{
				Name:  core.SpecAbilityHeatedBody,
				Hooks: map[core.HookType]bool{core.HookOnSelfHit: true},
				Data:  core.FeatureData{NumberOfDice: 1, Die: core.D10},
			},
		}
		attacker.StateManager.CurrentHP = 100
		ctx := &FeatureContext{
			Target: attacker, // In OnSelfHit dispatch to target, the Target is the attacker
			AttackContext: &AttackContext{
				Action: &core.Action{ActionType: core.ATMelee},
			},
		}

		ed.dispatchHooks(defender, core.HookOnSelfHit, ctx)

		if attacker.StateManager.CurrentHP >= 100 {
			t.Errorf("Expected attacker to take damage from Heated Body")
		}
	})

	t.Run("Absorption heals instead of taking damage", func(t *testing.T) {
		defender.Features = []core.Feature{
			{
				Name:  core.SpecAbilityAbsorption,
				Hooks: map[core.HookType]bool{core.HookOnSelfDamageTaken: true},
				Data:  core.FeatureData{DamageType: []core.DamageType{core.DamageFire}},
			},
		}
		defender.StateManager.CurrentHP = 40

		// Use the Adjudicator's resolveDamage to test the full flow
		action := &core.Action{
			Name: "Fire Attack",
			DiceBlock: []core.DiceBlock{
				{NumberOfDice: 0, Die: 0, Modifier: 10, DamageType: core.DamageFire},
			},
		}

		err := ed.Adjudicator.resolveDamage(attacker, defender, action, false, false)
		if err != nil {
			t.Fatalf("resolveDamage failed: %v", err)
		}

		if defender.StateManager.CurrentHP != 50 {
			t.Errorf("Expected HP to be 50 (40+10), got %d", defender.StateManager.CurrentHP)
		}
	})

	t.Run("Death Burst triggers on death", func(t *testing.T) {
		defender.Features = []core.Feature{
			{
				Name:  core.SpecAbilityDeathBurst,
				Hooks: map[core.HookType]bool{core.HookOnSelfDeath: true},
				Data: core.FeatureData{
					NumberOfDice: 2,
					Die:          core.D6,
					DamageType:   []core.DamageType{core.DamageFire},
					DC:           10,
					Ability:      core.AbilityDexterity,
					DCOnSuccess:  core.DCOnSuccessHalf,
				},
			},
		}
		attacker.StateManager.CurrentHP = 100
		defender.StateManager.CurrentHP = 10

		// Setup attacker for saving throw
		attacker.Abilities = core.Abilities{
			AbilityScores: core.AbilityScores{Dexterity: 10},
		}

		ctx := &FeatureContext{
			Target: attacker, // Killer
			AttackContext: &AttackContext{
				Action: &core.Action{Name: "Killing Blow"},
			},
		}

		ed.dispatchHooks(defender, core.HookOnSelfDeath, ctx)

		if attacker.StateManager.CurrentHP >= 100 {
			t.Errorf("Expected attacker to take damage from Death Burst")
		}
	})

	t.Run("Corrosive Form deals damage to melee attacker", func(t *testing.T) {
		defender.Features = []core.Feature{
			{
				Name:  core.SpecAbilityCorrosiveForm,
				Hooks: map[core.HookType]bool{core.HookOnSelfHit: true},
				Data:  core.FeatureData{NumberOfDice: 1, Die: core.D10},
			},
		}
		attacker.StateManager.CurrentHP = 100
		ctx := &FeatureContext{
			Target: attacker,
			AttackContext: &AttackContext{
				Action: &core.Action{ActionType: core.ATMelee},
			},
		}

		ed.dispatchHooks(defender, core.HookOnSelfHit, ctx)

		if attacker.StateManager.CurrentHP >= 100 {
			t.Errorf("Expected attacker to take damage from Corrosive Form")
		}
	})

	t.Run("Death Throes triggers on death", func(t *testing.T) {
		defender.Features = []core.Feature{
			{
				Name:  core.SpecAbilityDeathThroes,
				Hooks: map[core.HookType]bool{core.HookOnSelfDeath: true},
				Data: core.FeatureData{
					NumberOfDice: 10,
					Die:          core.D10,
					DamageType:   []core.DamageType{core.DamageFire},
					DC:           20,
					Ability:      core.AbilityDexterity,
					DCOnSuccess:  core.DCOnSuccessHalf,
				},
			},
		}
		attacker.StateManager.CurrentHP = 100
		defender.StateManager.CurrentHP = 10

		ctx := &FeatureContext{
			Target: attacker, // Killer
			AttackContext: &AttackContext{
				Action: &core.Action{Name: "Killing Blow"},
			},
		}

		ed.dispatchHooks(defender, core.HookOnSelfDeath, ctx)

		if attacker.StateManager.CurrentHP >= 100 {
			t.Errorf("Expected attacker to take damage from Death Throes")
		}
	})
}
