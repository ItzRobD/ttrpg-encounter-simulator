package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"testing"
)

func TestSpellSlotConsumption(t *testing.T) {
	seed := core.Seed{Seed1: 1, Seed2: 1}
	ed := NewEncounterDirector(seed, nil)

	// Create a wizard with 1 slot of level 1
	wizard := &actor.Actor{
		InstanceID: 1,
		Name:       "Wizard",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP:    10,
			MaxHP:        10,
			MaxSlots:     spells.SpellSlots{1: 1},
			CurrentSlots: spells.SpellSlots{1: 1},
		},
		SpellActions: []core.Action{
			{
				ID:            core.MakeID("spell:1"),
				Name:          "Magic Missile",
				ActionType:    core.ATSpell,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AverageDamage: 10,
				CastLevel:     1,
			},
		},
		Behavior: actor.BehaviorConfig{
			ActionPreference: core.APSpell,
		},
	}

	target := CreateMockGoblin(2, "Goblin", core.SideMonsters)

	ed.AddActor(wizard)
	ed.AddActor(target)
	ed.SetupEncounter()

	// 1. First cast - should succeed and consume the slot
	decision := ed.AIDirector.SelectActionDecision(wizard, ed)
	intents := ed.AIDirector.SelectAction(wizard, decision, ed)

	foundSpell := false
	for _, intent := range intents {
		if intent.Action.Name == "Magic Missile" {
			foundSpell = true
			err := ed.Adjudicator.ResolveAction(wizard, intent)
			if err != nil {
				t.Fatalf("Failed to resolve spell: %v", err)
			}
		}
	}

	if !foundSpell {
		t.Fatal("Wizard did not choose Magic Missile")
	}

	if wizard.StateManager.CurrentSlots[1] != 0 {
		t.Errorf("Expected 0 slots of level 1, got %d", wizard.StateManager.CurrentSlots[1])
	}

	// 2. Second turn - should NOT be able to cast Magic Missile again
	wizard.StateManager.ActionUsedCount = 0 // Reset for new turn simulation
	decision = ed.AIDirector.SelectActionDecision(wizard, ed)
	intents = ed.AIDirector.SelectAction(wizard, decision, ed)

	for _, intent := range intents {
		if intent.Action.Name == "Magic Missile" {
			t.Error("Wizard cast Magic Missile without available slots")
		}
	}
}

func TestInnateSpellConsumption(t *testing.T) {
	seed := core.Seed{Seed1: 1, Seed2: 1}
	ed := NewEncounterDirector(seed, nil)

	// Create a monster with 1 innate use of a spell
	monster := &actor.Actor{
		InstanceID: 1,
		Name:       "Aboleth",
		Side:       core.SideMonsters,
		ActorType:  core.ActorTypeMonster,
		StateManager: state_manager.StateManager{
			CurrentHP:     100,
			MaxHP:         100,
			InnateMax:     map[string]int{"Enslave": 1},
			InnateCurrent: map[string]int{"Enslave": 1},
		},
		SpellActions: []core.Action{
			{
				ID:            core.MakeID("spell:enslave"),
				Name:          "Enslave",
				ActionType:    core.ATSpell,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AverageDamage: 10,
				CastLevel:     1,
				IsInnate:      true,
			},
		},
		Behavior: actor.BehaviorConfig{
			ActionPreference: core.APSpell,
		},
	}

	target := CreateMockFighter(2, "Fighter", core.SideCharacters)

	ed.AddActor(monster)
	ed.AddActor(target)
	ed.SetupEncounter()

	// 1. First cast - should succeed and consume the innate use
	decision := ed.AIDirector.SelectActionDecision(monster, ed)
	intents := ed.AIDirector.SelectAction(monster, decision, ed)

	foundSpell := false
	for _, intent := range intents {
		if intent.Action.Name == "Enslave" {
			foundSpell = true
			err := ed.Adjudicator.ResolveAction(monster, intent)
			if err != nil {
				t.Fatalf("Failed to resolve spell: %v", err)
			}
		}
	}

	if !foundSpell {
		t.Fatal("Monster did not choose Enslave")
	}

	if monster.StateManager.InnateCurrent["Enslave"] != 0 {
		t.Errorf("Expected 0 innate uses of Enslave, got %d", monster.StateManager.InnateCurrent["Enslave"])
	}

	// 2. Second turn - should NOT be able to cast Enslave again
	monster.StateManager.ActionUsedCount = 0 // Reset for new turn simulation
	decision = ed.AIDirector.SelectActionDecision(monster, ed)
	intents = ed.AIDirector.SelectAction(monster, decision, ed)

	for _, intent := range intents {
		if intent.Action.Name == "Enslave" {
			t.Error("Monster cast Enslave without available innate uses")
		}
	}
}

func TestSpellFallbackToCantrip(t *testing.T) {
	seed := core.Seed{Seed1: 1, Seed2: 1}
	ed := NewEncounterDirector(seed, nil)

	// Create a wizard with 0 slots of level 1, but has a cantrip
	wizard := &actor.Actor{
		InstanceID: 1,
		Name:       "Wizard",
		Side:       core.SideCharacters,
		ActorType:  core.ActorTypeCharacter,
		StateManager: state_manager.StateManager{
			CurrentHP:    10,
			MaxHP:        10,
			MaxSlots:     spells.SpellSlots{1: 1},
			CurrentSlots: spells.SpellSlots{1: 0},
		},
		SpellActions: []core.Action{
			{
				ID:            core.MakeID("spell:1"),
				Name:          "Magic Missile",
				ActionType:    core.ATSpell,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AverageDamage: 10,
				CastLevel:     1,
			},
			{
				ID:            core.MakeID("spell:0"),
				Name:          "Fire Bolt",
				ActionType:    core.ATSpell,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AverageDamage: 5,
				CastLevel:     0,
			},
		},
		Behavior: actor.BehaviorConfig{
			ActionPreference: core.APSpell,
		},
	}

	target := CreateMockGoblin(2, "Goblin", core.SideMonsters)

	ed.AddActor(wizard)
	ed.AddActor(target)
	ed.SetupEncounter()

	// Wizard prefers spells. Magic Missile is out of slots. Should pick Fire Bolt.
	decision := ed.AIDirector.SelectActionDecision(wizard, ed)
	intents := ed.AIDirector.SelectAction(wizard, decision, ed)

	foundCantrip := false
	for _, intent := range intents {
		if intent.Action.Name == "Fire Bolt" {
			foundCantrip = true
		}
		if intent.Action.Name == "Magic Missile" {
			t.Error("Wizard chose Magic Missile when out of slots")
		}
	}

	if !foundCantrip {
		t.Error("Wizard did not fall back to Fire Bolt cantrip")
	}
}
