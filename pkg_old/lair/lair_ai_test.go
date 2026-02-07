package lair

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"math/rand/v2"
	"testing"
)

func TestLairAI_TargetsCharacters_WithAttackAction(t *testing.T) {
	// Lair with single action targeting characters
	l := NewLair("Test Lair", rand.New(rand.NewPCG(7, 8)))
	a := LairAction{
		Name:       "Rocks",
		Mode:       LAMAttack,
		TargetSide: TargetCharacters,
		AttackData: core.AttackData{
			Name: "Rocks",
			DamageBlocks: []core.DamageBlock{
				{NumberOfDice: 1, Die: core.D6, DamageType: core.DamageBludgeoning},
			},
			AttackModifier: 20, DamageModifier: 0,
		},
	}
	l.actionManager.AddLairAction(0, a)

	// Build combat context with two characters and one monster
	cc := core.NewCombatContext(&core.SimulationOptions{})
	cc.CombatantInfo = map[int]*core.CombatantInfo{
		1: {Combatant: &core.Combatant{Entity: newChar("C1", 10, true)}},
		2: {Combatant: &core.Combatant{Entity: newChar("C2", 10, true)}},
		3: {Combatant: &core.Combatant{Entity: newMon("M1", 10, false)}},
	}
	_ = l.UpdateAICombatContext(cc)

	req, err := l.ai.BuildLairActionRequest()
	if err != nil {
		t.Fatalf("BuildLairActionRequest: %v", err)
	}
	if req.ActionIndex != 0 {
		t.Fatalf("unexpected action index: %d", req.ActionIndex)
	}
	if !req.Target.IsCharacter() {
		t.Fatalf("expected target to be a character")
	}
}

func TestLairAI_TargetsMonsters_WithDCAction(t *testing.T) {
	l := NewLair("Test Lair", rand.New(rand.NewPCG(9, 10)))
	a := LairAction{
		Name:       "Steam",
		Mode:       LAMDC,
		TargetSide: TargetMonsters,
		IsAOE:      true,
		DCAbility:  core.AbilityDexterity,
		DCValue:    10,
		AttackData: core.AttackData{
			Name: "Steam",
			DamageBlocks: []core.DamageBlock{
				{NumberOfDice: 1, Die: core.D6, DamageType: core.DamageFire},
			},
		},
	}
	l.actionManager.AddLairAction(0, a)

	cc := core.NewCombatContext(&core.SimulationOptions{})
	cc.CombatantInfo = map[int]*core.CombatantInfo{
		1: {Combatant: &core.Combatant{Entity: newChar("C1", 10, true)}},
		2: {Combatant: &core.Combatant{Entity: newMon("M1", 10, false)}},
		3: {Combatant: &core.Combatant{Entity: newMon("M2", 10, true)}},
	}
	_ = l.UpdateAICombatContext(cc)

	req, err := l.ai.BuildLairActionRequest()
	if err != nil {
		t.Fatalf("BuildLairActionRequest: %v", err)
	}
	if !req.Target.IsMonster() {
		t.Fatalf("expected target to be a monster")
	}
}
