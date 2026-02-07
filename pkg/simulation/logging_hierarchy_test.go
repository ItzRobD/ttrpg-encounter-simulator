package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"testing"
)

func TestLoggingHierarchy(t *testing.T) {
	// Setup a simple encounter: 1 PC vs 1 Monster
	pc := &actor.Actor{
		Name:      "Warrior",
		Side:      core.SideCharacters,
		ActorType: core.ActorTypeCharacter,
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{Strength: 16},
		},
		HPConfig: core.HPConfig{HPAverage: 20},
		AC:       15,
		Actions: []core.Action{
			{
				ID:          "longsword",
				Name:        "Longsword",
				Cost:        core.ActionCost{ActivationType: core.ActAction, Value: 1},
				ActionType:  core.ATMelee,
				AttackBonus: 5,
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 1, Die: core.D8, Modifier: 3, DamageType: core.DamageSlashing},
				},
				AverageDamage: 7,
			},
		},
	}
	pc.StateManager.MaxHP = 20
	pc.StateManager.CurrentHP = 20
	pc.StateManager.AttackCount = 1

	monster := &actor.Actor{
		Name:      "Goblin",
		Side:      core.SideMonsters,
		ActorType: core.ActorTypeMonster,
		HPConfig:  core.HPConfig{HPAverage: 7},
		AC:        12,
	}
	monster.StateManager.MaxHP = 7
	monster.StateManager.CurrentHP = 7

	seed := core.Seed{Seed1: 1, Seed2: 1}
	options := &core.SimulationOptions{EnableSpecialAbilities: true}
	ed := NewEncounterDirector(seed, options)

	ed.AddActor(pc)
	ed.AddActor(monster)

	// Manually execute a turn for the PC
	err := ed.ExecuteTurn(pc.InstanceID)
	if err != nil {
		t.Fatalf("ExecuteTurn failed: %v", err)
	}

	timeline := ed.ExportTimeline()
	if len(timeline) == 0 {
		t.Fatal("Timeline is empty")
	}

	// Verify Hierarchy
	// Expected flow:
	// turn_start (Parent: "")
	//   decision_start (Parent: turn_start_id) -> wait, actually decision_start is Parent 1.
	//     action_start (Parent: decision_start_id)
	//       attack_roll (Parent: action_start_id - no wait, Parent 3 is created for Resolution)
	// Actually:
	// turn_start (Parent: "")
	// decision_start (Parent: "", ID: P1)
	//   action_start (Parent: P1, ID: P2)
	//     attack_roll (Parent: P2, ID: P3)
	//       outcome (Parent: P3, ID: P4)
	//         damage_roll (Parent: P4)
	//         hp_modified (Parent: P4)

	foundDecision := false
	foundAction := false
	foundAttackRoll := false
	foundOutcome := false
	foundHPModified := false

	var p1, p2, p3, p4 string

	for _, event := range timeline {
		t.Logf("Event: Type=%s, ID=%s, ParentID=%s", event.Type, event.ID, event.ParentID)
		switch event.Type {
		case events.EventDecisionStart:
			foundDecision = true
			p1 = event.ID
		case events.EventActionStart:
			if p1 != "" && event.ParentID == p1 {
				foundAction = true
				p2 = event.ID
			}
		case events.EventResolution:
			if p2 != "" && event.ParentID == p2 {
				p3 = event.ID
			}
		case events.EventAttackRoll:
			if p3 != "" && event.ParentID == p3 {
				foundAttackRoll = true
			}
		case events.EventOutcome:
			if p3 != "" && event.ParentID == p3 {
				foundOutcome = true
				p4 = event.ID
			}
		case events.EventHPModified:
			if p4 != "" && event.ParentID == p4 {
				foundHPModified = true
			}
		}
	}

	if !foundDecision {
		t.Error("DecisionStart not found")
	}
	if !foundAction {
		t.Error("ActionStart not found or wrong parent")
	}
	if !foundAttackRoll {
		t.Error("AttackRoll not found or wrong parent")
	}
	if !foundOutcome {
		t.Error("Outcome not found or wrong parent")
	}
	if !foundHPModified {
		t.Error("HPModified not found or wrong parent")
	}
}
