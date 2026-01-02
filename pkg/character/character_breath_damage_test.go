package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/testhelpers"
	"dnd5e-encounter-simulator-backend/pkg/races"
	"testing"
)

type targetWithSavingThrow struct{ testhelpers.EmEntity }

func (t targetWithSavingThrow) MakeSavingThrow(ability core.Ability, targetValue int, isSpell bool, damageType core.DamageType, simOptions *core.SimulationOptions) (core.RollResult, error) {
	return &roll_manager.RollResult{
		DiceRollType:   core.DiceRollSavingThrow,
		FinalRollValue: 10,
		Total:          10,
		IsSuccess:      false,
		TargetValue:    targetValue,
	}, nil
}

func TestExecuteAIRequest_DragonbornBreathWeapon_DamageEvent(t *testing.T) {
	// Setup
	color := races.DragonbornBlack
	as := core.AbilityScores{Constitution: 14}
	char := newTestCharacter(t, as, 1)

	char.Race = races.Race{
		ID:   races.Dragonborn,
		Name: "Dragonborn",
		DragonbornFeatures: &races.DragonbornFeatures{
			AncestryColor: color,
			DamageType:    core.DamageAcid,
			NumberOfDice:  2,
			Die:           core.D6,
		},
	}

	target := targetWithSavingThrow{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil)}

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
	_, err := char.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest failed: %v", err)
	}

	if !damageEventSeen {
		t.Errorf("DamageEvent was not emitted")
	}
}
