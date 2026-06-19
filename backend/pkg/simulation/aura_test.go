package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestHandleAura(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
		AOEHitsAllEnemies:      true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Fire Elemental (Owner of Fire Aura)
	elemental := &actor.Actor{
		InstanceID: 1,
		Name:       "Fire Elemental",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityFireAura,
				Hooks: map[core.HookType]bool{
					core.HookOnTurnStart: true,
					core.HookOnSelfHit:   true,
				},
				Data: core.FeatureData{
					NumberOfDice: 1,
					Die:          core.D10,
					DamageType:   []core.DamageType{core.DamageFire},
				},
			},
		},
	}

	// Target Fighter (Next to Elemental)
	fighter := &actor.Actor{
		InstanceID: 2,
		Name:       "Fighter",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		AC:         10,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	ed.Actors[1] = elemental
	ed.Actors[2] = fighter
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	// 1. Test Start of Turn Damage
	err := ed.HandleAura(elemental, elemental.Features[0], core.HookOnTurnStart, nil)
	if err != nil {
		t.Fatalf("HandleAura failed on turn start: %v", err)
	}

	if fighter.StateManager.CurrentHP == 50 {
		t.Error("Fighter should have taken damage from Fire Aura on turn start")
	}
	t.Logf("Fighter HP after turn start aura: %d", fighter.StateManager.CurrentHP)

	// 2. Test On Hit Damage
	fighter.StateManager.CurrentHP = 50
	ctx := &FeatureContext{
		Target: fighter,
		AttackContext: &AttackContext{
			Action: &core.Action{
				ActionType: core.ATMelee,
			},
		},
	}

	err = ed.HandleAura(elemental, elemental.Features[0], core.HookOnSelfHit, ctx)
	if err != nil {
		t.Fatalf("HandleAura failed on self hit: %v", err)
	}

	if fighter.StateManager.CurrentHP == 50 {
		t.Error("Fighter should have taken damage from Fire Aura when hitting elemental with melee")
	}
	t.Logf("Fighter HP after hitting elemental: %d", fighter.StateManager.CurrentHP)
}

func TestHandleHeatedBody(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Remorhaz (Owner of Heated Body)
	remorhaz := &actor.Actor{
		InstanceID: 1,
		Name:       "Remorhaz",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 100,
			MaxHP:     100,
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityHeatedBody,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfHit: true,
				},
				Data: core.FeatureData{
					NumberOfDice: 3,
					Die:          core.D10,
					DamageType:   []core.DamageType{core.DamageFire},
				},
			},
		},
	}

	fighter := &actor.Actor{
		InstanceID: 2,
		Name:       "Fighter",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		AC:         10,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
		},
	}

	ed.Actors[1] = remorhaz
	ed.Actors[2] = fighter
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	ctx := &FeatureContext{
		Target: fighter,
		AttackContext: &AttackContext{
			Action: &core.Action{
				ActionType: core.ATMelee,
			},
		},
	}

	err := ed.HandleMeleeTouchDamage(remorhaz, remorhaz.Features[0], core.HookOnSelfHit, ctx)
	if err != nil {
		t.Fatalf("HandleMeleeTouchDamage failed: %v", err)
	}

	foundDmgRoll := false
	foundHPMod := false
	for _, event := range ed.CombatLog {
		if event.Type == "damage_roll" {
			foundDmgRoll = true
		}
		if event.Type == "hp_modified" {
			foundHPMod = true
		}
	}

	if !foundDmgRoll {
		t.Error("Expected damage_roll event in combat log")
	}
	if !foundHPMod {
		t.Error("Expected hp_modified event in combat log")
	}

	if fighter.StateManager.CurrentHP == 50 {
		t.Error("Fighter should have taken damage from Heated Body")
	}
	t.Logf("Fighter HP after hitting Remorhaz: %d", fighter.StateManager.CurrentHP)
}
