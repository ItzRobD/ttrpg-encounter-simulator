package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleDeathBurst(t *testing.T) {
	simOptions := &core.SimulationOptions{
		MonsterDeathEffectsHitAllies: true,
		EnableSpecialAbilities:       true,
		AOEHitsAllEnemies:            true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Mephit (The one who dies)
	mephit := &actor.Actor{
		InstanceID: 1,
		Name:       "Mephit",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 1,
			MaxHP:     10,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityDeathBurst,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfDeath: true,
				},
				Data: core.FeatureData{
					NumberOfDice: 2,
					Die:          core.D6,
					DC:           10,
					Ability:      core.AbilityDexterity,
					DCOnSuccess:  core.DCOnSuccessHalf,
					DamageType:   []core.DamageType{core.DamageFire},
				},
			},
		},
	}

	// Allied Goblin (Should be hit because MonsterDeathEffectsHitAllies is true)
	goblin := &actor.Actor{
		InstanceID: 2,
		Name:       "Goblin",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		AC:         10,
		Abilities:  core.Abilities{AbilityScores: core.AbilityScores{Dexterity: 10}},
		StateManager: state_manager.StateManager{
			CurrentHP: 20,
			MaxHP:     20,
		},
	}

	// Enemy Fighter (Should always be hit)
	fighter := &actor.Actor{
		InstanceID: 3,
		Name:       "Fighter",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		AC:         10,
		Abilities:  core.Abilities{AbilityScores: core.AbilityScores{Dexterity: 10}},
		StateManager: state_manager.StateManager{
			CurrentHP: 20,
			MaxHP:     20,
		},
	}

	ed.Actors[1] = mephit
	ed.Actors[2] = goblin
	ed.Actors[3] = fighter

	// Simulate Mephit death
	// We need an action and an attacker to trigger the death via Adjudicator.resolveDamage
	attacker := fighter
	action := core.Action{
		Name: "Greatsword",
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 2, Die: core.D6, Modifier: 10}, // High damage to guarantee death
		},
	}

	err := ed.Adjudicator.resolveDamage(attacker, mephit, &action, false, false)
	if err != nil {
		t.Fatalf("resolveDamage failed: %v", err)
	}

	if mephit.StateManager.GetHealthState(mephit.IsCharacter()) != core.HealthStateDead {
		t.Errorf("Mephit should be dead, got %s", mephit.StateManager.GetHealthState(mephit.IsCharacter()))
	}

	// Check if Goblin was hit
	if goblin.StateManager.CurrentHP == 20 {
		t.Error("Goblin should have taken damage from Death Burst")
	}

	// Check if Fighter was hit
	if fighter.StateManager.CurrentHP == 20 {
		t.Error("Fighter should have taken damage from Death Burst")
	}

	t.Logf("Goblin HP: %d, Fighter HP: %d", goblin.StateManager.CurrentHP, fighter.StateManager.CurrentHP)
}

func TestHandleDeathBurst_SingleTarget(t *testing.T) {
	simOptions := &core.SimulationOptions{
		MonsterDeathEffectsHitAllies: false,
		EnableSpecialAbilities:       true,
		AOEHitsAllEnemies:            false, // Should only hit ONE enemy
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Mephit (The one who dies)
	mephit := &actor.Actor{
		InstanceID: 1,
		Name:       "Mephit",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 1,
			MaxHP:     10,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityDeathBurst,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfDeath: true,
				},
				Data: core.FeatureData{
					NumberOfDice: 2,
					Die:          core.D6,
					DC:           10,
					Ability:      core.AbilityDexterity,
					DCOnSuccess:  core.DCOnSuccessHalf,
					DamageType:   []core.DamageType{core.DamageFire},
				},
			},
		},
	}

	// Enemy Fighter 1 (Should be hit)
	fighter1 := &actor.Actor{
		InstanceID: 2,
		Name:       "Fighter 1",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		AC:         10,
		Abilities:  core.Abilities{AbilityScores: core.AbilityScores{Dexterity: 10}},
		StateManager: state_manager.StateManager{
			CurrentHP: 20,
			MaxHP:     20,
		},
	}

	// Enemy Fighter 2 (Should NOT be hit because AOEHitsAllEnemies is false)
	fighter2 := &actor.Actor{
		InstanceID: 3,
		Name:       "Fighter 2",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		AC:         10,
		Abilities:  core.Abilities{AbilityScores: core.AbilityScores{Dexterity: 10}},
		StateManager: state_manager.StateManager{
			CurrentHP: 20,
			MaxHP:     20,
		},
	}

	ed.Actors[1] = mephit
	ed.Actors[2] = fighter1
	ed.Actors[3] = fighter2

	attacker := fighter1
	action := core.Action{
		Name: "Greatsword",
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 2, Die: core.D6, Modifier: 10},
		},
	}

	err := ed.Adjudicator.resolveDamage(attacker, mephit, &action, false, false)
	if err != nil {
		t.Fatalf("resolveDamage failed: %v", err)
	}

	// Count how many enemies took damage
	damagedCount := 0
	if fighter1.StateManager.CurrentHP < 20 {
		damagedCount++
	}
	if fighter2.StateManager.CurrentHP < 20 {
		damagedCount++
	}

	if damagedCount != 1 {
		t.Errorf("Expected exactly 1 enemy to be damaged, but %d were", damagedCount)
	}

	// Verify it was Fighter 1 (the killer) who took damage
	if fighter1.StateManager.CurrentHP == 20 {
		t.Error("Fighter 1 (the killer) should have taken damage from Death Burst")
	}
	if fighter2.StateManager.CurrentHP < 20 {
		t.Error("Fighter 2 should NOT have taken damage from Death Burst")
	}

	t.Logf("Fighter 1 HP: %d, Fighter 2 HP: %d", fighter1.StateManager.CurrentHP, fighter2.StateManager.CurrentHP)
}

func TestHandleDeathBurst_NoFriendlyFire(t *testing.T) {
	simOptions := &core.SimulationOptions{
		MonsterDeathEffectsHitAllies: false,
		EnableSpecialAbilities:       true,
		AOEHitsAllEnemies:            true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Mephit (The one who dies)
	mephit := &actor.Actor{
		InstanceID: 1,
		Name:       "Mephit",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 1,
			MaxHP:     10,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityDeathBurst,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfDeath: true,
				},
				Data: core.FeatureData{
					NumberOfDice: 2,
					Die:          core.D6,
					DC:           10,
					Ability:      core.AbilityDexterity,
					DCOnSuccess:  core.DCOnSuccessHalf,
					DamageType:   []core.DamageType{core.DamageFire},
				},
			},
		},
	}

	// Allied Goblin (Should NOT be hit)
	goblin := &actor.Actor{
		InstanceID: 2,
		Name:       "Goblin",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		AC:         10,
		Abilities:  core.Abilities{AbilityScores: core.AbilityScores{Dexterity: 10}},
		StateManager: state_manager.StateManager{
			CurrentHP: 20,
			MaxHP:     20,
		},
	}

	// Enemy Fighter (Should always be hit)
	fighter := &actor.Actor{
		InstanceID: 3,
		Name:       "Fighter",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		AC:         10,
		Abilities:  core.Abilities{AbilityScores: core.AbilityScores{Dexterity: 10}},
		StateManager: state_manager.StateManager{
			CurrentHP: 20,
			MaxHP:     20,
		},
	}

	ed.Actors[1] = mephit
	ed.Actors[2] = goblin
	ed.Actors[3] = fighter

	attacker := fighter
	action := core.Action{
		Name: "Greatsword",
		DiceBlock: []core.DiceBlock{
			{NumberOfDice: 2, Die: core.D6, Modifier: 10},
		},
	}

	err := ed.Adjudicator.resolveDamage(attacker, mephit, &action, false, false)
	if err != nil {
		t.Fatalf("resolveDamage failed: %v", err)
	}

	// Check if Goblin was NOT hit
	if goblin.StateManager.CurrentHP != 20 {
		t.Errorf("Goblin should NOT have taken damage from Death Burst, but HP is %d", goblin.StateManager.CurrentHP)
	}

	// Check if Fighter was hit
	if fighter.StateManager.CurrentHP == 20 {
		t.Error("Fighter should have taken damage from Death Burst")
	}

	t.Logf("Goblin HP: %d, Fighter HP: %d", goblin.StateManager.CurrentHP, fighter.StateManager.CurrentHP)
}
