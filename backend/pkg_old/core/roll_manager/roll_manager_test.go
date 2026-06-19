package roll_manager

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"math/rand/v2"
	"testing"
)

// Test basic dice rolling with fixed seed
func TestRollDice(t *testing.T) {
	rm := &RollManager{
		rng: rand.New(rand.NewPCG(12345, 0)), // Fixed seed
	}

	tests := []struct {
		name         string
		numDice      int
		die          core.DiceType
		wantMinTotal int
		wantMaxTotal int
	}{
		{"1d20", 1, core.D20, 1, 20},
		{"2d6", 2, core.D6, 2, 12},
		{"3d8", 3, core.D8, 3, 24},
		{"8d6 fireball", 8, core.D6, 8, 48},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, rolls := rm.rollDice(tt.numDice, tt.die)

			// Check correct number of dice
			if len(rolls) != tt.numDice {
				t.Errorf("rollDice() rolled %d dice, want %d", len(rolls), tt.numDice)
			}

			// Check each die is in valid range
			for i, roll := range rolls {
				if roll < 1 || roll > tt.die.Int() {
					t.Errorf("die[%d] = %d, want 1-%d", i, roll, tt.die.Int())
				}
			}

			// Check total is sum of rolls
			expectedSum := 0
			for _, r := range rolls {
				expectedSum += r
			}
			if total != expectedSum {
				t.Errorf("total = %d, want %d (sum of rolls)", total, expectedSum)
			}

			// Check total is in valid range
			if total < tt.wantMinTotal || total > tt.wantMaxTotal {
				t.Errorf("total = %d, want range [%d-%d]", total, tt.wantMinTotal, tt.wantMaxTotal)
			}
		})
	}
}

// Function deprecated
// Test critical hits double dice
//func TestRollDoubleDice(t *testing.T) {
//	rm := &RollManager{
//		rng: rand.New(rand.NewPCG(54321, 0)),
//	}
//
//	tests := []struct {
//		name           string
//		numDice        int
//		die            core.DiceType
//		wantDiceRolled int
//	}{
//		{"1d6 crit becomes 2d6", 1, core.D6, 2},
//		{"2d6 crit becomes 4d6", 2, core.D6, 4},
//		{"1d8 crit becomes 2d8", 1, core.D8, 2},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			total, rolls := rm.rollDoubleDice(tt.numDice, tt.die)
//
//			if len(rolls) != tt.wantDiceRolled {
//				t.Errorf("rollDoubleDice() rolled %d dice, want %d", len(rolls), tt.wantDiceRolled)
//			}
//
//			// Verify all dice in valid range
//			for i, roll := range rolls {
//				if roll < 1 || roll > tt.die.Int() {
//					t.Errorf("die[%d] = %d, want 1-%d", i, roll, tt.die.Int())
//				}
//			}
//
//			// Verify total matches sum
//			if total != sum(rolls) {
//				t.Errorf("total = %d, want %d", total, sum(rolls))
//			}
//		})
//	}
//}

// Test improved criticals (roll normal + max dice)
func TestRollExtraMaxDice(t *testing.T) {
	rm := &RollManager{
		rng: rand.New(rand.NewPCG(99999, 0)),
	}

	tests := []struct {
		name         string
		numDice      int
		die          core.DiceType
		wantMinTotal int // numDice * 1 + numDice * max
		wantMaxTotal int // numDice * max + numDice * max
	}{
		{"1d6 improved crit", 1, core.D6, 7, 12},  // 1-6 + 6 = 7-12
		{"2d6 improved crit", 2, core.D6, 14, 24}, // 2-12 + 12 = 14-24
		{"1d8 improved crit", 1, core.D8, 9, 16},  // 1-8 + 8 = 9-16
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, rolls := rm.RollExtraMaxDice(tt.numDice, tt.die)

			// Should have double the dice
			if len(rolls) != tt.numDice*2 {
				t.Errorf("RollExtraMaxDice() rolled %d dice, want %d", len(rolls), tt.numDice*2)
			}

			// Second half should all be max
			for i := tt.numDice; i < len(rolls); i++ {
				if rolls[i] != tt.die.Int() {
					t.Errorf("roll[%d] = %d, want %d (max)", i, rolls[i], tt.die.Int())
				}
			}

			// First half should be in valid range
			for i := 0; i < tt.numDice; i++ {
				if rolls[i] < 1 || rolls[i] > tt.die.Int() {
					t.Errorf("roll[%d] = %d, want 1-%d", i, rolls[i], tt.die.Int())
				}
			}

			// Total should be in expected range
			if total < tt.wantMinTotal || total > tt.wantMaxTotal {
				t.Errorf("total = %d, want range [%d-%d]", total, tt.wantMinTotal, tt.wantMaxTotal)
			}
		})
	}
}

// Test advantage/disadvantage selection
func TestCalculateSingleDieFinalValue(t *testing.T) {
	rm := &RollManager{}

	tests := []struct {
		name      string
		rolls     []int
		advantage core.AdvantageType
		modifier  int
		wantValue int
		wantTotal int
	}{
		{
			name:      "normal roll",
			rolls:     []int{15},
			advantage: core.RollNormal,
			modifier:  3,
			wantValue: 15,
			wantTotal: 18,
		},
		{
			name:      "advantage takes higher",
			rolls:     []int{10, 18},
			advantage: core.RollAdvantage,
			modifier:  2,
			wantValue: 18,
			wantTotal: 20,
		},
		{
			name:      "disadvantage takes lower",
			rolls:     []int{18, 5},
			advantage: core.RollDisadvantage,
			modifier:  2,
			wantValue: 5,
			wantTotal: 7,
		},
		{
			name:      "advantage with equal rolls",
			rolls:     []int{12, 12},
			advantage: core.RollAdvantage,
			modifier:  0,
			wantValue: 12,
			wantTotal: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &RollResult{
				FinalRolls: tt.rolls,
				Advantage:  tt.advantage,
				Modifier:   tt.modifier,
			}

			rm.calculateSingleDieFinalValue(res)

			if res.FinalRollValue != tt.wantValue {
				t.Errorf("FinalRollValue = %d, want %d", res.FinalRollValue, tt.wantValue)
			}

			if res.Total != tt.wantTotal {
				t.Errorf("Total = %d, want %d", res.Total, tt.wantTotal)
			}
		})
	}
}

// Test Halfling Lucky reroll
func TestApplyHalflingLucky(t *testing.T) {
	rm := &RollManager{
		rng: rand.New(rand.NewPCG(11111, 0)),
		RerollAbilities: RerollAbilities{
			HasHalflingLucky: true,
		},
	}

	tests := []struct {
		name             string
		rolls            []int
		wantRerolled     bool
		checkRollChanged bool
	}{
		{"reroll single 1", []int{1}, true, true},
		{"reroll first 1 only", []int{1, 1}, true, true},
		{"no reroll without 1", []int{10, 15}, false, false},
		{"reroll 1 in mixed rolls", []int{15, 1}, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newRolls, events := rm.applyHalflingLucky(tt.rolls, core.D20)

			if tt.wantRerolled {
				if len(events) == 0 {
					t.Error("Expected reroll event, got none")
				}
				if newRolls == nil {
					t.Fatal("Expected new rolls, got nil")
				}
				if tt.checkRollChanged {
					// At least one roll should have changed
					changed := false
					for i := range tt.rolls {
						if tt.rolls[i] == 1 && newRolls[i] != 1 {
							changed = true
							break
						}
					}
					if !changed {
						t.Error("Expected at least one 1 to be rerolled")
					}
				}
			} else {
				if len(events) > 0 {
					t.Errorf("Expected no reroll events, got %d", len(events))
				}
			}
		})
	}
}

// Test Great Weapon Fighting reroll
func TestApplyGreatWeaponFighting(t *testing.T) {
	rm := &RollManager{
		rng: rand.New(rand.NewPCG(22222, 0)),
		RerollAbilities: RerollAbilities{
			HasGreatWeaponFighting: true,
		},
	}

	tests := []struct {
		name        string
		rolls       []int
		wantRerolls int // How many dice should be rerolled
	}{
		{"reroll 1s and 2s", []int{1, 2, 3}, 2},
		{"reroll only 1", []int{1, 5, 6}, 1},
		{"reroll only 2", []int{2, 5, 6}, 1},
		{"no reroll", []int{3, 4, 5}, 0},
		{"reroll all 1s and 2s", []int{1, 1, 2, 2}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newRolls, events := rm.applyGreatWeaponFighting(tt.rolls, core.D6)

			if len(events) != tt.wantRerolls {
				t.Errorf("Got %d reroll events, want %d", len(events), tt.wantRerolls)
			}

			if tt.wantRerolls > 0 && newRolls == nil {
				t.Fatal("Expected new rolls, got nil")
			}
		})
	}
}

// Test Elemental Adept (1s become 2s)
func TestApplyElementalAdept(t *testing.T) {
	rm := &RollManager{
		RerollAbilities: RerollAbilities{
			HasElementalAdept: true,
		},
	}

	tests := []struct {
		name      string
		rolls     []int
		wantRolls []int
	}{
		{"convert single 1 to 2", []int{1}, []int{2}},
		{"convert multiple 1s to 2s", []int{1, 1, 3}, []int{2, 2, 3}},
		{"no conversion needed", []int{2, 3, 4}, []int{2, 3, 4}},
		{"convert only 1s", []int{1, 2, 1, 3, 1}, []int{2, 2, 2, 3, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newRolls, events := rm.applyElementalAdept(tt.rolls, core.D6)

			if newRolls == nil {
				t.Fatal("Expected new rolls, got nil")
			}

			for i, want := range tt.wantRolls {
				if newRolls[i] != want {
					t.Errorf("roll[%d] = %d, want %d", i, newRolls[i], want)
				}
			}

			// Count how many 1s were in original
			onesCount := 0
			for _, r := range tt.rolls {
				if r == 1 {
					onesCount++
				}
			}

			if len(events) != onesCount {
				t.Errorf("Got %d events, want %d (one per 1)", len(events), onesCount)
			}
		})
	}
}

// Test helper functions
func TestHighest(t *testing.T) {
	tests := []struct {
		name  string
		rolls []int
		want  int
	}{
		{"simple", []int{5, 10, 3}, 10},
		{"first is highest", []int{20, 5, 10}, 20},
		{"last is highest", []int{5, 10, 15}, 15},
		{"all equal", []int{10, 10, 10}, 10},
		{"single value", []int{7}, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := highest(tt.rolls)
			if got != tt.want {
				t.Errorf("highest(%v) = %d, want %d", tt.rolls, got, tt.want)
			}
		})
	}
}

func TestLowest(t *testing.T) {
	tests := []struct {
		name  string
		rolls []int
		want  int
	}{
		{"simple", []int{5, 10, 3}, 3},
		{"first is lowest", []int{2, 5, 10}, 2},
		{"last is lowest", []int{15, 10, 5}, 5},
		{"all equal", []int{10, 10, 10}, 10},
		{"single value", []int{7}, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lowest(tt.rolls)
			if got != tt.want {
				t.Errorf("lowest(%v) = %d, want %d", tt.rolls, got, tt.want)
			}
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name  string
		rolls []int
		want  int
	}{
		{"simple", []int{1, 2, 3}, 6},
		{"single value", []int{10}, 10},
		{"zeros", []int{0, 0, 0}, 0},
		{"mixed", []int{5, 10, 15, 20}, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sum(tt.rolls)
			if got != tt.want {
				t.Errorf("sum(%v) = %d, want %d", tt.rolls, got, tt.want)
			}
		})
	}
}

func TestContainsOnes(t *testing.T) {
	tests := []struct {
		name  string
		rolls []int
		want  bool
	}{
		{"has one", []int{1, 5, 10}, true},
		{"multiple ones", []int{1, 1, 1}, true},
		{"no ones", []int{2, 3, 4}, false},
		{"empty", []int{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsOnes(tt.rolls)
			if got != tt.want {
				t.Errorf("containsOnes(%v) = %v, want %v", tt.rolls, got, tt.want)
			}
		})
	}
}

func TestContainsOnesOrTwos(t *testing.T) {
	tests := []struct {
		name  string
		rolls []int
		want  bool
	}{
		{"has one", []int{1, 5, 10}, true},
		{"has two", []int{2, 5, 10}, true},
		{"has both", []int{1, 2, 5}, true},
		{"has neither", []int{3, 4, 5}, false},
		{"empty", []int{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsOnesOrTwos(tt.rolls)
			if got != tt.want {
				t.Errorf("containsOnesOrTwos(%v) = %v, want %v", tt.rolls, got, tt.want)
			}
		})
	}
}
