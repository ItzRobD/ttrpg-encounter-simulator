package roll_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"math/rand/v2"
	"testing"
)

func TestRollManager_RollD20_Advantage(t *testing.T) {
	// Fixed seed for determinism if needed, but here we just check logic
	rng := rand.New(rand.NewPCG(1, 1))
	rm := NewRollManager(rng)

	opts := RollOptions{
		Advantage: core.RollAdvantage,
	}

	for i := 0; i < 100; i++ {
		res := rm.RollD20(opts)
		if len(res.FinalRolls) != 2 {
			t.Errorf("Expected 2 rolls for advantage, got %d", len(res.FinalRolls))
		}
		expected := res.FinalRolls[0]
		if res.FinalRolls[1] > expected {
			expected = res.FinalRolls[1]
		}
		if res.FinalRollValue != expected {
			t.Errorf("Advantage failed: expected %d, got %d (rolls: %v)", expected, res.FinalRollValue, res.FinalRolls)
		}
	}
}

func TestRollManager_RollD20_HalflingLucky(t *testing.T) {
	// We want to force a 1.
	// Since we can't easily mock rand.Rand in Go 1.22+ math/rand/v2 without some effort,
	// we will just run it many times or use a seed that we know produces a 1.
	rng := rand.New(rand.NewPCG(42, 42))
	rm := NewRollManager(rng)

	opts := RollOptions{
		Advantage:   core.RollNormal,
		RerollOnOne: true,
	}

	foundReroll := false
	for i := 0; i < 1000; i++ {
		res := rm.RollD20(opts)
		if res.OriginalRolls[0] == 1 {
			foundReroll = true
			if len(res.RerollEvents) == 0 {
				t.Errorf("Expected reroll event when rolling a 1 with Halfling Lucky")
			}
			if res.FinalRolls[0] == 1 && res.RerollEvents[0].NewRoll != 1 {
				// This could happen if the second roll is also a 1, which is valid.
				// But we check if the NewRoll is recorded.
			}
		}
	}
	if !foundReroll {
		t.Log("Did not roll a 1 in 1000 trials, lucky test might be inconclusive")
	}
}

func TestRollManager_RollDamage_GWF(t *testing.T) {
	rng := rand.New(rand.NewPCG(123, 123))
	rm := NewRollManager(rng)

	opts := RollOptions{
		RollType:        core.DiceRollDamage,
		RerollThreshold: 2, // GWF
	}

	foundGWF := false
	for i := 0; i < 1000; i++ {
		res := rm.RollDice(1, core.D12, opts)
		if res.OriginalRolls[0] <= 2 {
			foundGWF = true
			if len(res.RerollEvents) == 0 {
				t.Errorf("Expected GWF reroll for roll %d", res.OriginalRolls[0])
			}
		}
	}
	if !foundGWF {
		t.Log("Did not roll a 1 or 2 in 1000 trials")
	}
}

func TestRollManager_RollExtraMaxDice(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	rm := NewRollManager(rng)

	opts := RollOptions{}
	count := 2
	die := core.D6

	res := rm.RollExtraMaxDice(count, die, opts)

	// FinalRollValue should be sum of 2d6 + (2 * 6)
	// Original RollDice result should be the same as if we called it directly with same RNG state
	// But since we just want to verify the math:
	// res.Total = sum(rolls) + modifier + (count * die)

	expectedBaseSum := 0
	for _, v := range res.FinalRolls {
		expectedBaseSum += v
	}
	// Note: RollExtraMaxDice currently adds maxVal to FinalRollValue AND Total.
	// FinalRollValue is usually the sum of dice.

	expectedTotal := expectedBaseSum + (count * die.Int())
	if res.Total != expectedTotal {
		t.Errorf("Expected Total %d, got %d", expectedTotal, res.Total)
	}
}
