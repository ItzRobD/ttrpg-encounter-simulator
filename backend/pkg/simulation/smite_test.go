package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"testing"
)

func TestHandleSmiteLikeFeature_DivineEminence(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	// Priest with Divine Eminence
	priest := &actor.Actor{
		InstanceID: 1,
		Name:       "Priest",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP: 27,
			MaxHP:     27,
			CurrentSlots: spells.SpellSlots{
				1: 4,
				2: 3,
			},
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityDivineEminence,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfHit: true,
				},
				Data: core.FeatureData{
					NumberOfDice: 3, // 3d6 base
					Die:          core.D6,
				},
			},
		},
	}

	target := &actor.Actor{
		InstanceID: 2,
		Name:       "Fighter",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP: 30,
			MaxHP:     30,
		},
	}

	ed.Actors[1] = priest
	ed.Actors[2] = target
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	ctx := &FeatureContext{
		Target: target,
		AttackContext: &AttackContext{
			Action: &core.Action{
				ActionType: core.ATMelee,
			},
		},
	}

	// Test 1st level slot (lowest by default)
	err := ed.HandleSmiteLikeFeature(priest, priest.Features[0], core.HookOnSelfHit, ctx)
	if err != nil {
		t.Fatalf("HandleSmiteLikeFeature failed: %v", err)
	}

	if priest.StateManager.CurrentSlots[1] != 3 {
		t.Errorf("Expected 3 slots of level 1, got %d", priest.StateManager.CurrentSlots[1])
	}

	if target.StateManager.CurrentHP == 30 {
		t.Error("Target should have taken damage")
	}

	// Reset target HP and priest bonus action
	target.StateManager.CurrentHP = 30
	priest.StateManager.BonusActionUsedCount = 0

	// Test 2nd level slot (highest)
	ed.SimOptions.PaladinUseHighestSmiteSlot = true
	err = ed.HandleSmiteLikeFeature(priest, priest.Features[0], core.HookOnSelfHit, ctx)
	if err != nil {
		t.Fatalf("HandleSmiteLikeFeature failed: %v", err)
	}

	if priest.StateManager.CurrentSlots[2] != 2 {
		t.Errorf("Expected 2 slots of level 2, got %d", priest.StateManager.CurrentSlots[2])
	}
}

func TestHandleSmiteLikeFeature_DivineSmite(t *testing.T) {
	simOptions := &core.SimulationOptions{
		EnableSpecialAbilities: true,
		PaladinAlwaysSmite:     true,
	}
	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, simOptions)

	paladin := &actor.Actor{
		InstanceID: 1,
		Name:       "Paladin",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
			CurrentSlots: spells.SpellSlots{
				1: 4,
			},
		},
		Features: []core.Feature{
			{
				Name: core.SpecAbilityDivineSmite,
				Hooks: map[core.HookType]bool{
					core.HookOnSelfHit: true,
				},
				Data: core.FeatureData{
					NumberOfDice:     2, // 2d8 base
					Die:              core.D8,
					BonusTargetTypes: []core.MonsterType{core.MTUndead, core.MTFiend},
				},
			},
		},
	}

	// Standard target
	goblin := &actor.Actor{
		InstanceID: 2,
		Name:       "Goblin",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		Metadata: actor.Metadata{
			MonsterType: core.MTHumanoid,
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 7,
			MaxHP:     7,
		},
	}

	// Undead target
	zombie := &actor.Actor{
		InstanceID: 3,
		Name:       "Zombie",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		Metadata: actor.Metadata{
			MonsterType: core.MTUndead,
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 22,
			MaxHP:     22,
		},
	}

	ed.Actors[1] = paladin
	ed.Actors[2] = goblin
	ed.Actors[3] = zombie
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	// Hit Goblin
	ctxGoblin := &FeatureContext{
		Target: goblin,
		AttackContext: &AttackContext{
			Action: &core.Action{
				ActionType: core.ATMelee,
			},
		},
	}

	ed.HandleSmiteLikeFeature(paladin, paladin.Features[0], core.HookOnSelfHit, ctxGoblin)
	// 2d8 damage should have been dealt
	if goblin.StateManager.CurrentHP == 7 {
		t.Error("Goblin should have taken damage")
	}

	// Hit Zombie
	ctxZombie := &FeatureContext{
		Target: zombie,
		AttackContext: &AttackContext{
			Action: &core.Action{
				ActionType: core.ATMelee,
			},
		},
	}

	ed.HandleSmiteLikeFeature(paladin, paladin.Features[0], core.HookOnSelfHit, ctxZombie)
	// 3d8 damage (2d8 + 1 bonus) should have been dealt
	if zombie.StateManager.CurrentHP == 22 {
		t.Error("Zombie should have taken damage")
	}
}
