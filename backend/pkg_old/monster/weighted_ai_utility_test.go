package monster

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/entity_state_manager"
	"fmt"
	"math/rand/v2"
	"testing"
)

type dummyCharacter struct {
	core.Entity
	name string
	id   int
	esm  *entity_state_manager.EntityStateManager
	hp   core.HPStatus
}

func (d *dummyCharacter) GetEntityType() core.EntityType {
	//TODO implement me
	panic("implement me")
}

func (d *dummyCharacter) GetName() string                      { return d.name }
func (d *dummyCharacter) IsMonster() bool                      { return false }
func (d *dummyCharacter) IsCharacter() bool                    { return true }
func (d *dummyCharacter) GetInstanceID() int                   { return d.id }
func (d *dummyCharacter) GetHPStatus() core.HPStatus           { return d.hp }
func (d *dummyCharacter) GetAC() int                           { return 10 }
func (d *dummyCharacter) IsConcentrating() bool                { return false }
func (d *dummyCharacter) GetIsLegendary() bool                 { return false }
func (d *dummyCharacter) GetState() interface{}                { return d.esm }
func (d *dummyCharacter) GetAttackBonus() int                  { return 0 }
func (d *dummyCharacter) GetConditions() core.EntityConditions { return core.NewEntityConditions() }
func (d *dummyCharacter) IsUnconscious() bool                  { return false }
func (d *dummyCharacter) IsDead() bool                         { return false }
func (d *dummyCharacter) GetCasterLevel() int                  { return 0 }
func (d *dummyCharacter) GetLevel() float64                    { return 1.0 }

type dummyHP struct {
	core.HPStatus
	hp, max int
}

func (h *dummyHP) GetHP() int    { return h.hp }
func (h *dummyHP) GetMaxHP() int { return h.max }

func TestMonsterWeightedAI_DynamicUtilityScaling(t *testing.T) {
	// 1. Setup actor (Goblin)
	weights := &core.UtilityWeights{
		ActionWeights: map[core.ActionType]float64{
			core.ATMonsterDamage: 1.0,
		},
	}
	weights.TargetFactorWeights.LowHP = 1.0 // Prioritize low HP

	actor := &Monster{}
	actor.Name = "Goblin"
	actor.InstanceID = 100
	actor.RNG = rand.New(rand.NewPCG(100, 100))
	actor.ActionManager = &MonsterActionManager{}
	actor.AI = NewMonsterAI(actor, weights)
	actor.EntityStateManager, _ = entity_state_manager.NewEntityStateManager(actor, entity_state_manager.EntityStateConfig{})

	// 2. Setup Targets (Dummy Characters to avoid import cycle)
	targetA := &dummyCharacter{name: "HealthyPlayer", id: 1, hp: &dummyHP{hp: 100, max: 100}}
	targetB := &dummyCharacter{name: "WoundedPlayer", id: 2, hp: &dummyHP{hp: 10, max: 100}}

	// 3. Setup Combat ctx
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

	fmt.Printf("[TEST] Scenario 2: Best Target id: %d, Score: %.2f, Action: %s\n", bestID, score2, action2)

	if bestID != 2 {
		t.Errorf("Expected WoundedPlayer (id 2) to be chosen, got %d", bestID)
	}
	if score2 <= score1 {
		t.Errorf("Expected wounded target score (%.2f) to be higher than healthy target score (%.2f)", score2, score1)
	}
}
