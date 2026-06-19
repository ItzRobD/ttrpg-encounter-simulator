package intermission_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"dnd5e-encounter-simulator-backend/pkg_old/classes"
	"math/rand/v2"
	"testing"
)

func TestIntermissionManager_ShortRest(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	rm := roll_manager.NewRollManager(rng)
	im := NewIntermissionManager(rm)

	a := &actor.Actor{
		Name:       "Test Hero",
		InstanceID: 1,
		ActorType:  core.ActorTypeCharacter,
		HPConfig: core.HPConfig{
			HitDice: map[core.DiceType]int{core.D10: 5},
		},
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{Constitution: 14}, // +2
		},
		StateManager: state_manager.StateManager{
			MaxHP:          50,
			CurrentHP:      10,
			MaxHitDice:     map[core.DiceType]int{core.D10: 5},
			CurrentHitDice: map[core.DiceType]int{core.D10: 5},
		},
	}

	party := []*actor.Actor{a}
	opts := IntermissionOptions{
		MaxShortRests:          2,
		ShortRestHealThreshold: 0.5,
	}

	// Should take short rest because 10 < 50*0.5
	result := im.ProcessIntermission(party, opts)

	if im.ShortRestsTaken != 1 {
		t.Errorf("Expected 1 short rest taken, got %d", im.ShortRestsTaken)
	}

	if result.HealingReceived[a.InstanceID] == 0 {
		t.Error("Expected healing in return map")
	}

	if a.StateManager.CurrentHP <= 10 {
		t.Errorf("Expected HP to increase after short rest, still %d", a.StateManager.CurrentHP)
	}

	if a.StateManager.CurrentHitDice[core.D10] == 5 {
		t.Error("Expected hit dice to be spent")
	}
}

func TestIntermissionManager_WarlockRecovery(t *testing.T) {
	rm := roll_manager.NewRollManager(rand.New(rand.NewPCG(1, 2)))
	im := NewIntermissionManager(rm)

	warlock := &actor.Actor{
		Metadata: actor.Metadata{
			ClassID: classes.ClassID(core.Warlock),
		},
		StateManager: state_manager.StateManager{
			MaxSlots:     spells.SpellSlots{1: 2},
			CurrentSlots: spells.SpellSlots{1: 0},
			MaxHP:        20,
			CurrentHP:    20, // Doesn't need healing
		},
	}

	party := []*actor.Actor{warlock}
	opts := IntermissionOptions{
		MaxShortRests:          2,
		ShortRestHealThreshold: 1.0, // Force rest if any HP missing, but wait...
	}

	// If we want to force a short rest even if healthy (not supported by current logic)
	// Let's damage him slightly
	warlock.StateManager.CurrentHP = 19

	im.ProcessIntermission(party, opts)

	if warlock.StateManager.CurrentSlots[1] != 2 {
		t.Errorf("Expected warlock slots to recover, got %d", warlock.StateManager.CurrentSlots[1])
	}
}

func TestIntermissionManager_LayOnHands(t *testing.T) {
	rm := roll_manager.NewRollManager(rand.New(rand.NewPCG(1, 2)))
	im := NewIntermissionManager(rm)

	paladin := &actor.Actor{
		Features: []core.Feature{
			{Name: core.SpecAbilityLayOnHands},
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 50,
			MaxHP:     50,
			Resource: map[string]int{
				string(core.SpecAbilityLayOnHands): 10,
			},
		},
	}

	target := &actor.Actor{
		StateManager: state_manager.StateManager{
			CurrentHP: 10,
			MaxHP:     30,
		},
	}

	party := []*actor.Actor{paladin, target}
	im.performPostCombatHealing(party, 1.0)

	if target.StateManager.CurrentHP != 20 {
		t.Errorf("Expected target to be healed by 10, got %d", target.StateManager.CurrentHP)
	}

	if paladin.StateManager.Resource[string(core.SpecAbilityLayOnHands)] != 0 {
		t.Errorf("Expected Lay on Hands pool to be empty, got %d", paladin.StateManager.Resource[string(core.SpecAbilityLayOnHands)])
	}
}

func TestIntermissionManager_Aggregation(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	rm := roll_manager.NewRollManager(rng)
	im := NewIntermissionManager(rm)

	a := &actor.Actor{
		Name:       "Wounded Hero",
		InstanceID: 1,
		ActorType:  core.ActorTypeCharacter,
		HPConfig: core.HPConfig{
			HitDice: map[core.DiceType]int{core.D10: 1},
		},
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{Constitution: 10},
		},
		StateManager: state_manager.StateManager{
			MaxHP:          100,
			CurrentHP:      10,
			MaxHitDice:     map[core.DiceType]int{core.D10: 1},
			CurrentHitDice: map[core.DiceType]int{core.D10: 1},
		},
	}

	healer := &actor.Actor{
		Name:       "Paladin",
		InstanceID: 2,
		ActorType:  core.ActorTypeCharacter,
		Features: []core.Feature{
			{Name: core.SpecAbilityLayOnHands},
		},
		StateManager: state_manager.StateManager{
			MaxHP:     100,
			CurrentHP: 100,
			Resource: map[string]int{
				string(core.SpecAbilityLayOnHands): 5,
			},
		},
	}

	party := []*actor.Actor{a, healer}
	opts := IntermissionOptions{
		MaxShortRests:          1,
		ShortRestHealThreshold: 0.5,
		PostRestHealThreshold:  1.0,
	}

	result := im.ProcessIntermission(party, opts)

	// Should have some healing from HD and 5 from Lay on Hands
	if result.HealingReceived[a.InstanceID] == 0 {
		t.Errorf("Expected healing for actor, got 0")
	}
	if result.HealingReceived[a.InstanceID] < 6 {
		t.Errorf("Expected at least 6 healing (1 from HD + 5 from LOH), got %d", result.HealingReceived[a.InstanceID])
	}
}
