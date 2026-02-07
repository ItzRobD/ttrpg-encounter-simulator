package roll_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"math/rand/v2"
)

// RollManager is a centralized dice engine.
type RollManager struct {
	rng *rand.Rand
}

func NewRollManager(rng *rand.Rand) *RollManager {
	return &RollManager{
		rng: rng,
	}
}

// RollOptions defines the parameters for a roll.
type RollOptions struct {
	RollType          core.DiceRollType
	Advantage         core.AdvantageType
	Modifier          int
	CriticalThreshold int

	// Reroll/Utility flags
	RerollOnOne     bool // Halfling Lucky
	RerollThreshold int  // Great Weapon Fighting (usually 2)
	RerollMaxTimes  int  // Usually 1 for most features

	AdvantageCount    int // Track multiple sources of advantage
	DisadvantageCount int // Track multiple sources of disadvantage
}

func (o *RollOptions) CalculateAdvantage() core.AdvantageType {
	if o.AdvantageCount > 0 && o.DisadvantageCount > 0 {
		return core.RollNormal
	}
	if o.AdvantageCount > 0 {
		return core.RollAdvantage
	}
	if o.DisadvantageCount > 0 {
		return core.RollDisadvantage
	}
	return core.RollNormal
}

// RollResult contains the outcome of a roll, including traceability for the UI.
type RollResult struct {
	RollType       core.DiceRollType  `json:"roll_type"`
	Dice           core.DiceType      `json:"dice"`
	NumberOfDice   int                `json:"number_of_dice"`
	Modifier       int                `json:"modifier"`
	Total          int                `json:"total"`
	FinalRollValue int                `json:"final_roll_value"` // Sum of dice before modifier
	FinalRolls     []int              `json:"final_rolls"`
	OriginalRolls  []int              `json:"original_rolls"`
	Advantage      core.AdvantageType `json:"advantage"`

	IsCritical   bool `json:"is_critical"`
	IsNaturalOne bool `json:"is_natural_one"`

	RerollEvents []RerollEvent `json:"reroll_events"`
}

type RerollEvent struct {
	Reason       string        `json:"reason"`
	OriginalRoll int           `json:"original_roll"`
	NewRoll      int           `json:"new_roll"`
	Die          core.DiceType `json:"die"`
}

// RollD20 performs a standard d20 roll with advantage/disadvantage and Halfling Lucky support.
func (rm *RollManager) RollD20(opts RollOptions) *RollResult {
	res := &RollResult{
		RollType:     opts.RollType,
		Dice:         core.D20,
		NumberOfDice: 1,
		Modifier:     opts.Modifier,
		Advantage:    opts.Advantage,
	}

	var rolls []int
	switch opts.Advantage {
	case core.RollAdvantage, core.RollDisadvantage:
		rolls = []int{rm.RollDie(core.D20), rm.RollDie(core.D20)}
	default:
		rolls = []int{rm.RollDie(core.D20)}
	}
	res.OriginalRolls = rolls

	// Apply Halfling Lucky (Reroll on 1)
	finalRolls := make([]int, len(rolls))
	copy(finalRolls, rolls)

	if opts.RerollOnOne {
		for i, v := range finalRolls {
			if v == 1 {
				newRoll := rm.RollDie(core.D20)
				res.RerollEvents = append(res.RerollEvents, RerollEvent{
					Reason:       "Halfling Lucky",
					OriginalRoll: 1,
					NewRoll:      newRoll,
					Die:          core.D20,
				})
				finalRolls[i] = newRoll
				// Halfling Lucky only rerolls the first 1 found in the set (per PHB/Legacy logic)
				break
			}
		}
	}
	res.FinalRolls = finalRolls

	// Determine final value based on advantage
	switch opts.Advantage {
	case core.RollAdvantage:
		res.FinalRollValue = highest(finalRolls)
	case core.RollDisadvantage:
		res.FinalRollValue = lowest(finalRolls)
	default:
		res.FinalRollValue = finalRolls[0]
	}

	res.Total = res.FinalRollValue + opts.Modifier

	// Crit/Nat1 flags
	if res.FinalRollValue == 1 {
		res.IsNaturalOne = true
	}
	threshold := opts.CriticalThreshold
	if threshold <= 0 {
		threshold = 20
	}
	if res.FinalRollValue >= threshold {
		res.IsCritical = true
	}

	return res
}

// RollDice rolls multiple dice and applies reroll logic like GWF.
func (rm *RollManager) RollDice(numberOfDice int, die core.DiceType, opts RollOptions) *RollResult {
	res := &RollResult{
		RollType:     opts.RollType,
		Dice:         die,
		NumberOfDice: numberOfDice,
		Modifier:     opts.Modifier,
		Advantage:    core.RollNormal, // Advantage doesn't apply to non-d20 rolls
	}

	rawRolls := make([]int, numberOfDice)
	for i := 0; i < numberOfDice; i++ {
		rawRolls[i] = rm.RollDie(die)
	}
	res.OriginalRolls = rawRolls

	finalRolls := make([]int, numberOfDice)
	copy(finalRolls, rawRolls)

	// Apply Reroll Logic (GWF: reroll 1s and 2s)
	if opts.RollType == core.DiceRollDamage {
		if opts.RerollThreshold > 0 {
			for i, v := range finalRolls {
				if v <= opts.RerollThreshold {
					newRoll := rm.RollDie(die)
					res.RerollEvents = append(res.RerollEvents, RerollEvent{
						Reason:       "Great Weapon Fighting",
						OriginalRoll: v,
						NewRoll:      newRoll,
						Die:          die,
					})
					finalRolls[i] = newRoll
				}
			}
		}
	}

	res.FinalRolls = finalRolls
	sum := 0
	for _, v := range finalRolls {
		sum += v
	}
	res.FinalRollValue = sum
	res.Total = sum + opts.Modifier

	return res
}

// RollExtraMaxDice handles "Improved Critical" style rolls where you add max value instead of doubling dice.
func (rm *RollManager) RollExtraMaxDice(count int, die core.DiceType, opts RollOptions) *RollResult {
	res := rm.RollDice(count, die, opts)

	// Add the max possible value of the original dice
	maxVal := count * die.Int()
	res.Total += maxVal
	res.FinalRollValue += maxVal

	return res
}

// RollDie generates a random integer between 1 and the maximum value of the specified die, inclusive. Returns 0 for invalid dice.
func (rm *RollManager) RollDie(die core.DiceType) int {
	if die <= 0 {
		return 0
	}
	return rm.rng.IntN(die.Int()) + 1
}

func highest(s []int) int {
	if len(s) == 0 {
		return 0
	}
	m := s[0]
	for _, v := range s {
		if v > m {
			m = v
		}
	}
	return m
}

func lowest(s []int) int {
	if len(s) == 0 {
		return 0
	}
	m := s[0]
	for _, v := range s {
		if v < m {
			m = v
		}
	}
	return m
}
