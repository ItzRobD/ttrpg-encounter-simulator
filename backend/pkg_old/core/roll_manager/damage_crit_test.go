package roll_manager

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/testhelpers"
	"math/rand/v2"
	"testing"
)

func TestRollDamage_CriticalNumberDice(t *testing.T) {
	parent := testhelpers.NewEmEntity(1, core.AbilityScores{}, nil)
	rm := NewRollManager(parent, RerollAbilities{})
	rm.rng = rand.New(rand.NewPCG(1, 1))

	req := &core.AttackRequest{
		AttackData: []core.AttackData{
			{
				DamageBlocks: []core.DamageBlock{
					{NumberOfDice: 1, Die: core.D8},
				},
			},
		},
		AttackOptions: core.AttackOptions{
			ShouldApplyDamageMod: true,
		},
		SimulationOptions: &core.SimulationOptions{
			CanCharactersCrit: true,
		},
	}

	// Normal damage
	res, err := rm.RollDamage(req, 0, false, RollOptions{RollType: core.DiceRollDamage}, false)
	if err != nil {
		t.Fatalf("RollDamage error: %v", err)
	}
	if res.NumberOfDice != 1 {
		t.Errorf("Normal damage NumberOfDice = %d, want 1", res.NumberOfDice)
	}
	if len(res.FinalRolls) != 1 {
		t.Errorf("Normal damage FinalRolls length = %d, want 1", len(res.FinalRolls))
	}

	// Critical damage (Standard)
	resCrit, err := rm.RollDamage(req, 0, true, RollOptions{RollType: core.DiceRollDamage}, false)
	if err != nil {
		t.Fatalf("RollDamage error: %v", err)
	}
	if len(resCrit.FinalRolls) != 2 {
		t.Errorf("Crit damage FinalRolls length = %d, want 2", len(resCrit.FinalRolls))
	}
	if resCrit.NumberOfDice != 2 {
		t.Errorf("Crit damage NumberOfDice = %d, want 2", resCrit.NumberOfDice)
	}

	// Critical damage (Improved)
	req.SimulationOptions.UseImprovedCritical = true
	resImp, err := rm.RollDamage(req, 0, true, RollOptions{RollType: core.DiceRollDamage}, false)
	if err != nil {
		t.Fatalf("RollDamage error: %v", err)
	}
	if len(resImp.FinalRolls) != 2 {
		t.Errorf("Improved crit damage FinalRolls length = %d, want 2", len(resImp.FinalRolls))
	}
	if resImp.NumberOfDice != 2 {
		t.Errorf("Improved crit damage NumberOfDice = %d, want 2", resImp.NumberOfDice)
	}
}
