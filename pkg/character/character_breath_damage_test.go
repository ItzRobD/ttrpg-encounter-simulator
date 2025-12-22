package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/testhelpers"
	"dnd5e-encounter-simulator-backend/pkg/races"
	"testing"
)

func TestExecuteAIRequest_DragonbornBreathWeapon_DamageEvent(t *testing.T) {
	// Setup
	color := races.DragonbornBlack
	charConfig := CharacterConfig{
		Name:            "DragonbornTester",
		RaceID:          races.Dragonborn,
		DragonbornColor: &color,
		ClassID:         1, // Fighter
		Level:           1,
		AsConfig: core.AbilityScoresConfig{
			AbilityScores: core.AbilityScores{Constitution: 14},
		},
	}

	char, err := NewCharacter(context.Background(), charConfig)
	if err != nil {
		t.Fatalf("Failed to create character: %v", err)
	}

	target := testhelpers.NewEmEntity(1, core.AbilityScores{}, nil)

	// Create request
	req := &core.AIRequest{
		ActionType: core.ATDragonbornBreathWeapon,
		Target:     target,
		TargetID:   1,
	}

	// Capture events
	var damageEventSeen bool
	char.SetEventListener(func(event interface{}) {
		if e, ok := event.(*events.DamageEvent); ok {
			damageEventSeen = true
			if e.Target != target.GetName() {
				t.Errorf("Expected target %s, got %s", target.GetName(), e.Target)
			}
			if len(e.Rolls) == 0 {
				t.Errorf("Expected rolls, got none")
			}
		}
	})

	// Execute
	_, err = char.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest failed: %v", err)
	}

	if !damageEventSeen {
		t.Errorf("DamageEvent was not emitted")
	}
}
