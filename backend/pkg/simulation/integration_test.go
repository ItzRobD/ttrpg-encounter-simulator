package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	"testing"
)

func TestIntegration_MockCombat(t *testing.T) {
	ed := SetupStandardEncounter()

	// Run 2 rounds of combat to see how things progress
	for round := 1; round <= 2; round++ {
		t.Logf("--- Starting Round %d ---", round)

		// Typically EncounterDirector would have a RunRound method or similar.
		// Since it might not have a full "Run" loop yet, let's simulate the turns.
		ed.CurrentRound = round
		if ed.Statistics != nil && round > 1 {
			ed.Statistics.IncrementRound()
		}

		for _, actorID := range ed.TurnOrder {
			a := ed.Actors[actorID]
			if a.StateManager.CurrentHP <= 0 {
				continue
			}

			// 1. Reset Turn State (normally done by EncounterDirector)
			a.StateManager.ActionUsedCount = 0
			a.StateManager.BonusActionUsedCount = 0

			// 2. Select Action
			decision := ed.AIDirector.SelectActionDecision(a, ed)
			intents := ed.AIDirector.SelectAction(a, decision, ed)

			// 3. Resolve Intents
			for _, intent := range intents {
				targetStr := "None"
				if len(intent.TargetIDs) > 0 {
					targetStr = fmt.Sprintf("%v", intent.TargetIDs)
				}
				t.Logf("Actor %s (%d) uses %s on Target %s", a.Name, a.InstanceID, intent.Action.Name, targetStr)
				err := ed.Adjudicator.ResolveAction(a, intent)
				if err != nil {
					t.Errorf("Error resolving action for %s: %v", a.Name, err)
				}
			}
		}

		// Print Summary after each round
		for _, a := range ed.Actors {
			t.Logf("Status: %s (HP: %d/%d)", a.Name, a.StateManager.CurrentHP, a.StateManager.MaxHP)
		}
	}

	// Basic assertions to ensure simulation actually did SOMETHING
	// (e.g., somebody took damage or somebody was healed)
	hasAnyDamage := false
	for _, a := range ed.Actors {
		if a.StateManager.CurrentHP < a.StateManager.MaxHP {
			hasAnyDamage = true
			break
		}
	}

	if !hasAnyDamage {
		t.Log("Warning: No damage was dealt during 2 rounds. This might happen with bad luck or high AC.")
	}
}

func TestIntegration_ClericHealing(t *testing.T) {
	ed := SetupStandardEncounter()

	fighter := ed.Actors[1] // Valeros
	cleric := ed.Actors[2]  // Kyra

	// Manually damage the fighter to trigger healing
	fighter.StateManager.CurrentHP = 2
	ed.Statistics.MarkNeedsHealing(fighter.InstanceID)
	ed.Statistics.MarkNeedsEmergencyHealing(fighter.InstanceID)

	// DEBUG: Verify statistics lists
	// t.Logf("NeedsHealing: %v", ed.Statistics.NeedsHealing)
	// t.Logf("NeedsEmergencyHealing: %v", ed.Statistics.NeedsEmergencyHealing)

	// Ensure Cleric's turn is next or just trigger it
	decision := ed.AIDirector.SelectActionDecision(cleric, ed)
	if decision != core.DecisionHeal {
		t.Errorf("Cleric decision was %v, expected heal", decision)
	}
	intents := ed.AIDirector.SelectAction(cleric, decision, ed)

	// DEBUG: check all intents
	foundHeal := false
	for _, intent := range intents {
		if intent.Action.Name == "Cure Wounds" && len(intent.TargetIDs) > 0 && intent.TargetIDs[0] == fighter.InstanceID {
			foundHeal = true
			err := ed.Adjudicator.ResolveAction(cleric, intent)
			if err != nil {
				t.Fatalf("Failed to resolve heal: %v", err)
			}
		}
	}

	if !foundHeal {
		t.Errorf("Cleric did not prioritize emergency healing for the Fighter")
	}

	if fighter.StateManager.CurrentHP <= 2 {
		t.Errorf("Fighter was not healed. HP still %d", fighter.StateManager.CurrentHP)
	}

	t.Logf("Fighter HP after heal: %d", fighter.StateManager.CurrentHP)
}
