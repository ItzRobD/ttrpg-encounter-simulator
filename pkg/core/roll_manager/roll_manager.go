package roll_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"fmt"
	"math/rand/v2"
)

type AdvantageType int

const (
	RollNormal AdvantageType = iota
	RollAdvantage
	RollDisadvantage
)

func (at AdvantageType) String() string {
	switch at {
	case RollNormal:
		return "Normal"
	case RollAdvantage:
		return "Advantage"
	case RollDisadvantage:
		return "Disadvantage"
	default:
		return "invalid"
	}
}

type DiceRollType string

const (
	DiceRollAttack       DiceRollType = "attack"
	DiceRollDamage       DiceRollType = "damage"
	DiceRollInitiative   DiceRollType = "initiative"
	DiceRollAbilityCheck DiceRollType = "ability check"
	DiceRollSavingThrow  DiceRollType = "saving throw"
)

type DiceType int

const (
	d4   DiceType = 4
	d6   DiceType = 6
	d8   DiceType = 8
	d10  DiceType = 10
	d12  DiceType = 12
	d20  DiceType = 20
	d100 DiceType = 100
)

func (dt DiceType) String() string {
	return fmt.Sprintf("%d", int(dt))
}

func (dt DiceType) Int() int {
	return int(dt)
}

func (dt DiceType) Max() int {
	return int(dt)
}

func (dt DiceType) Min() int {
	return 1
}

func (dt DiceType) Avg() float64 {
	return (float64(dt) / 2) + 0.5
}

func (dt DiceType) IsValid() bool {
	switch dt {
	case d4, d6, d8, d10, d12, d20, d100:
		return true
	}
	return false
}

type RollManager struct {
	parent             core.Entity
	options            RollOptions
	luckyUsesRemaining int
	rng                *rand.Rand
}

type RollOptions struct {
	// Base roll settings
	Advantage AdvantageType
	Modifier  int

	// Reroll abilities
	RerollAbilities RerollAbilities

	// Other modifiers
	CriticalThreshold int
	TreatOnesAsTwos   bool // Elemental Adept

	// Context for logging
	RollType    DiceRollType
	RollContext string // "Attack Roll", "Damage Roll", etc.
}

type RerollAbilities struct {
	HasHalflingLucky       bool
	HasLuckyFeat           bool
	HasElvenAccuracy       bool
	HasGreatWeaponFighting bool
}

type RollResult struct {
	FinalRollValue int
	FinalRolls     []int
	Modifier       int
	Total          int
	Advantage      AdvantageType

	// Reroll tracking
	OriginalRolls []int
	RerollEvents  []RerollEvent
	WasRerolled   bool

	// Special results
	IsCritical   bool
	IsNaturalOne bool
}

func (rr RollResult) GetFinalRollValue() int  { return rr.FinalRollValue }
func (rr RollResult) GetFinalRolls() []int    { return rr.FinalRolls }
func (rr RollResult) GetModifier() int        { return rr.Modifier }
func (rr RollResult) GetTotal() int           { return rr.Total }
func (rr RollResult) GetAdvantage() string    { return rr.Advantage.String() }
func (rr RollResult) GetOriginalRolls() []int { return rr.OriginalRolls }
func (rr RollResult) GetWasRerolled() bool    { return rr.WasRerolled }
func (rr RollResult) GetIsCritical() bool     { return rr.IsCritical }
func (rr RollResult) GetIsNaturalOne() bool   { return rr.IsNaturalOne }
func (rr RollResult) GetRerollEvents() []map[string]interface{} {
	var r []map[string]interface{}
	for _, event := range rr.RerollEvents {
		m := make(map[string]interface{})
		m["Reason"] = event.Reason
		m["Original Roll"] = event.OriginalRoll
		m["New Roll"] = event.NewRoll
		m["Die"] = event.Die.String()
		m["Reroll Type"] = event.RerollType.String()
		r = append(r, m)
	}
	return r
}

type RerollEvent struct {
	Reason       string
	OriginalRoll int
	NewRoll      int
	Die          DiceType
	RerollType   RerollType
}

type RerollType string

const (
	RerollHalflingLucky  RerollType = "Halfling Lucky"
	RerollLuckyFeat      RerollType = "Lucky Feat"
	RerollGWF            RerollType = "Great Weapon Fighting"
	RerollElementalAdept RerollType = "Elemental Adept"
)

func (rt RerollType) String() string {
	return string(rt)
}

func NewRollManager(parent core.Entity, options RollOptions) *RollManager {
	return &RollManager{
		parent:             parent,
		luckyUsesRemaining: 3, // Lucky feat gives 3 uses
		options:            options,
	}
}

func (rm *RollManager) SetOptions(options RollOptions) {
	rm.options = options
}

func (rm *RollManager) RollD20(options RollOptions) (*RollResult, error) {
	var res RollResult // Single return value
	res.Advantage = options.Advantage
	res.Modifier = options.Modifier

	// Handle d20 rolls with advantage/disadvantage

	switch rm.options.Advantage {
	case RollNormal:
		roll := rm.rollDie(d20)
		res.OriginalRolls = []int{roll}
		res.FinalRolls = []int{roll}
	case RollAdvantage:
		_, rolls := rm.rollDice(2, d20)
		res.OriginalRolls = rolls
		res.FinalRolls = rolls
	case RollDisadvantage:
		_, rolls := rm.rollDice(2, d20)
		res.OriginalRolls = rolls
		res.FinalRolls = rolls
	}

	// Apply Halfling Lucky
	if containsOne(res.OriginalRolls) {
		newRolls, rEvents := rm.applyHalflingLucky(res.OriginalRolls, d20)
		if len(rEvents) > 0 {
			res.FinalRolls = newRolls
			res.RerollEvents = append(res.RerollEvents, rEvents...)
		}
	}

	rm.calculateFinalValue(&res)
	if res.FinalRollValue == 1 {
		res.IsNaturalOne = true
	}
	if res.FinalRollValue >= options.CriticalThreshold {
		res.IsCritical = true
	}

	// Log all events
	events.LogDiceRollEvent(rm.parent, &res, rm.parent.GetEventListener())

	return &res, nil
}

func (rm *RollManager) RollInitiative(options RollOptions) (*RollResult, error) {
	// Roll a d20 and apply applicable modifiers
	mod, err := core.GetAbilityScoreModifier(rm.parent.GetAbilityScores().Dexterity)
	if err != nil {
		return nil, err
	}
	// TODO: Handle alert's +5 modifier -> access parent feats

	options.Modifier += mod
	options.RollType = DiceRollInitiative
	options.RollContext = "Initiative Roll"

	res, err := rm.RollD20(options)
	if err != nil {
		return nil, err
	}

	// Log initiative roll
	events.LogDiceRollEvent(rm.parent, res, rm.parent.GetEventListener())

	return res, nil
}

func (rm *RollManager) RollDamage(numDice, die int, options RollOptions) (*RollResult, error) {
	// Handle damage rolls
	// Apply Great Weapon Fighting, Elemental Adept
	// Log all events
}

// TODO: I was working on rolling saving throws
//
//	I also modified monster and character for getters of ability scores/saves
//	Verify that makes sense -> Character and monster is going to have to get
//	Reworked anyway based on the inclusion of this new roll manager.
func (rm *RollManager) RollSavingThrow(ability core.Ability, options RollOptions) (*RollResult, error) {
	// Specialized for saving throws
	// Can apply proficiency
	bonus := rm.parent.GetSavingThrowBonus(ability)

	options.Modifier += bonus
	options.RollType = DiceRollSavingThrow
	options.RollContext = "Saving Throw"

	res, err := rm.RollD20(options)
	if err != nil {
		return nil, err
	}

	// TODO: Modify the event handler to accept RollResult
	events.LogSavingThrowEvent()

	res.Total += 8

	return res, nil
}

func (rm *RollManager) RollAbilityCheck(ability core.Ability, options RollOptions) (*RollResult, error) {
	mod, err := rm.parent.GetAbilityScoreModifier(ability)
	if err != nil {
		return nil, err
	}
	options.Modifier += mod
	options.RollType = DiceRollAbilityCheck
	options.RollContext = "Ability Check"

	res, err := rm.RollD20(options)
	if err != nil {
		return nil, err
	}

	// Log ability check roll
	// TODO: Ability checks likely aren't going to be used
	// 		Do we need to actually need to keep this function?
	events.LogDiceRollEvent(rm.parent, res, rm.parent.GetEventListener())

	return res, nil
}

func (rm *RollManager) UseLuckyReroll() bool {
	if rm.luckyUsesRemaining > 0 {
		rm.luckyUsesRemaining--
		return true
	}
	return false
}

func (rm *RollManager) RestoreLuckyUses() {
	rm.luckyUsesRemaining = rm.maxLuckyUses
}

// applyHalflingLucky applies the Halfling Lucky trait to reroll dice rolls of 1, returning the updated rolls and reroll events.
func (rm *RollManager) applyHalflingLucky(rolls []int, die DiceType) ([]int, []RerollEvent) {
	if !rm.options.RerollAbilities.HasHalflingLucky {
		return nil, nil
	}

	var rerollEvents []RerollEvent
	newRolls := make([]int, len(rolls))
	copy(newRolls, rolls)

	for i, roll := range newRolls {
		if roll == 1 {
			originalRoll := roll
			newRoll := rm.rollDie(die)
			newRolls[i] = newRoll

			rerollEvents = append(rerollEvents, RerollEvent{
				Reason:       RerollHalflingLucky.String(),
				OriginalRoll: originalRoll,
				NewRoll:      newRoll,
				Die:          die,
				RerollType:   RerollHalflingLucky,
			})
		}
		break
	}

	return newRolls, rerollEvents
}

func (rm *RollManager) calculateFinalValue(res *RollResult) {
	switch res.Advantage {
	case RollNormal:
		res.FinalRollValue = res.FinalRolls[0]
		res.Total = res.FinalRollValue + res.Modifier
	case RollAdvantage:
		res.FinalRollValue = highest(res.FinalRolls)
		res.Total = res.FinalRollValue + res.Modifier
	case RollDisadvantage:
		res.FinalRollValue = lowest(res.FinalRolls)
		res.Total = res.FinalRollValue + res.Modifier
	}
}

func (rm *RollManager) rollDice(numDice int, die DiceType) (int, []int) {
	rolls := make([]int, numDice)
	for i := 0; i < numDice; i++ {
		rolls[i] = rm.rng.IntN(die.Int()) + 1
	}

	return sum(rolls), rolls
}

func (rm *RollManager) rollDie(die DiceType) int {
	return rm.rng.IntN(die.Int()) + 1
}

func sum(arr []int) int {
	s := 0
	for _, v := range arr {
		s += v
	}
	return s
}

func highest(arr []int) int {
	r := arr[0]
	for _, v := range arr {
		if v > r {
			r = v
		}
	}
	return r
}

func lowest(arr []int) int {
	r := arr[0]
	for _, v := range arr {
		if v < r {
			r = v
		}
	}
}

func containsOne(arr []int) bool {
	for _, v := range arr {
		if v == 1 {
			return true
		}
	}
	return false
}
