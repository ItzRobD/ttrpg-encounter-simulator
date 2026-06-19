package lair

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"math/rand/v2"
	"testing"
)

func TestNewLairFromConfig_BasicAttackAndDC(t *testing.T) {
	cfg := &LairConfig{
		Enabled:    true,
		Name:       "Test Lair",
		Initiative: 20,
		Actions: []LairActionInput{
			{
				Name:         "Falling Rubble",
				Mode:         LAMAttack,
				TargetSide:   TargetCharacters,
				TargetPolicy: "lowest max hp",
				IsAOE:        false,
				Recharge:     0,
				AttackBonus:  5,
				NumberOfDice: 2,
				Die:          core.D8,
				AmountToAdd:  0,
				DamageType:   core.DamageBludgeoning,
			},
			{
				Name:         "Scalding Steam",
				Mode:         LAMDC,
				TargetSide:   TargetMonsters,
				TargetPolicy: "lowest max hp",
				IsAOE:        true,
				Recharge:     5,
				DCAbility:    core.AbilityDexterity,
				DCValue:      12,
				OnSuccess:    core.DCOnSuccessHalf,
				NumberOfDice: 2,
				Die:          core.D6,
				AmountToAdd:  0,
				DamageType:   core.DamageFire,
			},
		},
	}

	l, err := NewLairFromConfig(cfg, rand.New(rand.NewPCG(1, 2)))
	if err != nil {
		t.Fatalf("NewLairFromConfig error: %v", err)
	}

	lam := l.GetActionManager()
	if lam == nil {
		t.Fatalf("nil action manager")
	}
	if len(lam.Actions) != 2 {
		t.Fatalf("want 2 actions, got %d", len(lam.Actions))
	}

	// Validate attack action fields present
	a0 := lam.Actions[0]
	if a0.Mode != LAMAttack {
		t.Errorf("action 0 mode = %v want attack", a0.Mode)
	}
	if a0.AttackData.Name != "Falling Rubble" {
		t.Errorf("attack name mismatch: %s", a0.AttackData.Name)
	}

	// Validate DC action fields present
	a1 := lam.Actions[1]
	if a1.Mode != LAMDC {
		t.Errorf("action 1 mode = %v want dc", a1.Mode)
	}
	if a1.DCAbility != core.AbilityDexterity || a1.DCValue != 12 {
		t.Errorf("dc fields mismatch")
	}
}

func TestNewLairFromConfig_Disabled(t *testing.T) {
	_, err := NewLairFromConfig(&LairConfig{Enabled: false}, nil)
	if err == nil {
		t.Fatalf("expected error for disabled lair config")
	}
}
