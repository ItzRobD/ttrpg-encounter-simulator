package lair

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"math/rand/v2"
	"testing"
)

func TestLairActionManager_Attack_HitAndMiss(t *testing.T) {
	l := NewLair("Test Lair", rand.New(rand.NewPCG(11, 12)))
	lam := l.GetActionManager()
	lam.AddLairAction(0, LairAction{
		Name:       "Rocks",
		Mode:       LAMAttack,
		TargetSide: TargetCharacters,
		AttackData: core.AttackData{Name: "Rocks", NumberOfDice: 1, Die: core.D6, AttackModifier: 20, DamageModifier: 0, DamageType: core.DamageBludgeoning},
	})

	// Force a hit: target AC = 0
	req := &core.AttackRequest{
		AttackData:        []core.AttackData{lam.Actions[0].AttackData},
		AttackOptions:     core.AttackOptions{Advantage: core.RollNormal, ShouldApplyDamageMod: true},
		SimulationOptions: &core.SimulationOptions{},
		Target:            newChar("C1", 0, true),
	}
	results, err := lam.ProcessAttackRequest(req)
	if err != nil {
		t.Fatalf("ProcessAttackRequest(hit): %v", err)
	}
	if len(results) != 1 || !results[0].GetIsHit() {
		t.Fatalf("expected 1 hit result, got %+v", results)
	}

	// Force a miss: target AC extremely high
	req2 := &core.AttackRequest{
		AttackData:        []core.AttackData{lam.Actions[0].AttackData},
		AttackOptions:     core.AttackOptions{Advantage: core.RollNormal, ShouldApplyDamageMod: true},
		SimulationOptions: &core.SimulationOptions{},
		Target:            newChar("C2", 100, true),
	}
	results2, err := lam.ProcessAttackRequest(req2)
	if err != nil {
		t.Fatalf("ProcessAttackRequest(miss): %v", err)
	}
	if len(results2) != 1 || results2[0].GetIsHit() {
		t.Fatalf("expected miss result, got %+v", results2)
	}
}

func TestLairActionManager_DC_AOE_HalfOnSuccess(t *testing.T) {
	l := NewLair("Test Lair", rand.New(rand.NewPCG(13, 14)))
	lam := l.GetActionManager()
	lam.AddLairAction(0, LairAction{
		Name:       "Steam",
		Mode:       LAMDC,
		TargetSide: TargetCharacters,
		IsAOE:      true,
		DCAbility:  core.AbilityDexterity,
		DCValue:    10,
		OnSuccess:  core.DCOnSuccessHalf,
		AttackData: core.AttackData{Name: "Steam", NumberOfDice: 2, Die: core.D6, DamageType: core.DamageFire},
	})

	// Build a fake combat context with two characters: one succeeds save, one fails
	cc := core.NewCombatContext(&core.SimulationOptions{})
	c1 := newChar("C1", 10, true)
	c2 := newChar("C2", 10, false)
	cc.CombatantInfo = map[int]*core.CombatantInfo{
		1: {Combatant: &core.Combatant{Entity: c1}},
		2: {Combatant: &core.Combatant{Entity: c2}},
	}
	_ = l.UpdateAICombatContext(cc)

	// Execute advanced DC action (AOE). Use c1 as primary target; manager will collect both.
	results, effects, err := lam.ExecuteAdvanced(0, c1)
	if err != nil {
		t.Fatalf("ExecuteAdvanced: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results for AOE, got %d", len(results))
	}
	if len(effects) == 0 {
		t.Fatalf("expected damage effects for at least one target")
	}
}

func TestLairActionManager_RechargeFlow(t *testing.T) {
	l := NewLair("Test Lair", rand.New(rand.NewPCG(15, 16)))
	lam := l.GetActionManager()
	lam.AddLairAction(0, LairAction{
		Name:       "Steam",
		Mode:       LAMDC,
		TargetSide: TargetCharacters,
		IsAOE:      false,
		DCAbility:  core.AbilityDexterity,
		DCValue:    1,
		OnSuccess:  core.DCOnSuccessNone,
		AttackData: core.AttackData{Name: "Steam", NumberOfDice: 1, Die: core.D6, DamageType: core.DamageFire},
		Recharge:   1, // easiest to recharge
	})

	// Available at start
	if !lam.IsActionAvailable(0) {
		t.Fatalf("action should be available initially")
	}

	// Use it once (mark expended)
	_, _, err := lam.ExecuteAdvanced(0, newChar("C1", 10, false))
	if err != nil {
		t.Fatalf("ExecuteAdvanced first use: %v", err)
	}
	if lam.IsActionAvailable(0) {
		t.Fatalf("action should be on cooldown after use")
	}

	// Roll recharge; with Recharge=1, it should always succeed
	lam.RollRechargeActions()
	if !lam.IsActionAvailable(0) {
		t.Fatalf("action should have recharged")
	}
}
