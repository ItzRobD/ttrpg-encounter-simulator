package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"fmt"
	"math/rand/v2"
	"testing"
)

func TestWeightedAI_DynamicUtilityScaling(t *testing.T) {
	// 1. Setup Actor (Wizard)
	weights := &core.UtilityWeights{
		ActionWeights: map[core.ActionType]float64{
			core.ATDamage: 1.0,
		},
	}
	weights.TargetFactorWeights.LowHP = 1.0 // Prioritize low HP

	as := core.AbilityScores{Intelligence: 16}
	actor := &Character{
		Name:          "Wizard",
		Level:         5,
		AbilityScores: as,
		RNG:           rand.New(rand.NewPCG(42, 42)),
	}
	actor.AI = NewCharacterAI(actor, weights)
	actor.EntityStateManager, _ = entity_state_manager.NewEntityStateManager(actor, entity_state_manager.EntityStateConfig{})

	// 2. Setup Targets (Monsters)
	targetA := &monster.Monster{}
	targetA.Name = "HealthyTarget"
	targetA.InstanceID = 1
	targetA.EntityStateManager, _ = entity_state_manager.NewEntityStateManager(targetA, entity_state_manager.EntityStateConfig{})
	targetA.MonsterBase.HP = core.HPConfig{HPSetMethod: core.HPSetValue, Value: 100}
	targetA.InitializeHP()

	targetB := &monster.Monster{}
	targetB.Name = "WoundedTarget"
	targetB.InstanceID = 2
	targetB.EntityStateManager, _ = entity_state_manager.NewEntityStateManager(targetB, entity_state_manager.EntityStateConfig{})
	targetB.MonsterBase.HP = core.HPConfig{HPSetMethod: core.HPSetValue, Value: 100}
	targetB.InitializeHP()
	targetB.ModifyHP(-90, false, false, false, core.DamageFire, false) // 10/100 HP

	// 3. Setup Combat Context
	ctx := &core.CombatContext{
		CombatantInfo: make(map[int]*core.CombatantInfo),
		Options: &core.SimulationOptions{
			UseWeightedAI:    true,
			HPVisibilityMode: core.HPVisibilityWhite,
		},
	}
	actor.AI.UpdateCombatContext(ctx)

	// --- SCENARIO 1: Only Healthy Target is available ---
	ctx.CombatantInfo = map[int]*core.CombatantInfo{
		1: core.NewCombatantWithInfo(targetA).Info,
	}

	_, _, score1, _, _ := actor.AI.SelectTargetID(core.TTDamage)
	action1, _ := actor.AI.chooseActionWeighted()

	fmt.Printf("[TEST] Scenario 1: Healthy Target Score: %.2f, Action: %s\n", score1, action1)

	// --- SCENARIO 2: Both targets are available (Wounded should be chosen) ---
	ctx.CombatantInfo[2] = core.NewCombatantWithInfo(targetB).Info

	_, bestID, score2, _, _ := actor.AI.SelectTargetID(core.TTDamage)
	action2, _ := actor.AI.chooseActionWeighted()

	fmt.Printf("[TEST] Scenario 2: Best Target ID: %d, Score: %.2f, Action: %s\n", bestID, score2, action2)

	if bestID != 2 {
		t.Errorf("Expected WoundedTarget (ID 2) to be chosen, got %d", bestID)
	}
	if score2 <= score1 {
		t.Errorf("Expected wounded target score (%.2f) to be higher than healthy target score (%.2f)", score2, score1)
	}
}

func TestWeightedAI_DragonbornBreathNudge(t *testing.T) {
	// Verify that Dragonborn Breath is scaled by target score
	// ... (Similar setup)
}
