package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func TestMonsterMultiattackRetargeting(t *testing.T) {
	// Setup Simulation Options
	options := &core.SimulationOptions{
		MultiattackPolicy:      core.MultiattackPolicyRetargetOnDown,
		CanMonstersCrit:        false,
		CanCharactersCrit:      false,
		EnableSpecialAbilities: true,
	}

	ed := NewEncounterDirector(core.Seed{Seed1: 1, Seed2: 1}, options)

	// Create a Dragon-like monster with Multiattack
	monster := &actor.Actor{
		InstanceID: 1,
		Name:       "Young Red Dragon",
		ActorType:  core.ActorTypeMonster,
		Side:       core.SideMonsters,
		Actions: []core.Action{
			{
				ID:         "multiattack",
				Name:       "Multiattack",
				ActionType: core.ATMultiAttack,
				Cost:       core.ActionCost{ActivationType: core.ActAction, Value: 1},
				Multiattack: []core.Multiattack{
					{ActionID: "bite", Count: 1},
					{ActionID: "claw", Count: 2},
				},
			},
			{
				ID:          "bite",
				Name:        "Bite",
				ActionType:  core.ATMelee,
				AttackBonus: 10,
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 2, Die: core.D10, DamageType: core.DamagePiercing, Modifier: 6},
				},
			},
			{
				ID:          "claw",
				Name:        "Claw",
				ActionType:  core.ATMelee,
				AttackBonus: 10,
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 2, Die: core.D6, DamageType: core.DamageSlashing, Modifier: 6},
				},
			},
		},
		StateManager: state_manager.StateManager{
			CurrentHP: 178,
			MaxHP:     178,
		},
	}

	// Create two characters: Bob and Frank
	bob := &actor.Actor{
		InstanceID: 2,
		Name:       "Bob",
		ActorType:  core.ActorTypeCharacter,
		Side:       core.SideCharacters,
		AC:         10,
		StateManager: state_manager.StateManager{
			CurrentHP: 1, // Bob is almost dead
			MaxHP:     20,
		},
	}

	frank := &actor.Actor{
		InstanceID: 3,
		Name:       "Frank",
		ActorType:  core.ActorTypeCharacter,
		Side:       core.SideCharacters,
		AC:         10,
		StateManager: state_manager.StateManager{
			CurrentHP: 20,
			MaxHP:     20,
		},
	}

	ed.Actors[1] = monster
	ed.Actors[2] = bob
	ed.Actors[3] = frank
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	// Execute Multiattack
	intent := ActionIntent{
		ActivationType: core.ActAction,
		ActorID:        monster.InstanceID,
		TargetIDs:      []int{bob.InstanceID},
		Action:         monster.Actions[0], // Multiattack
	}

	err := ed.Adjudicator.ResolveAction(monster, intent)
	if err != nil {
		t.Fatalf("Multiattack failed: %v", err)
	}

	// Frank should have been attacked after Bob was downed
	if frank.StateManager.CurrentHP == frank.StateManager.MaxHP {
		t.Errorf("Frank was not attacked after Bob was downed. Current HP: %d", frank.StateManager.CurrentHP)
	}

	if bob.StateManager.CurrentHP > 0 {
		t.Errorf("Bob should be downed, but has %d HP", bob.StateManager.CurrentHP)
	}
}
