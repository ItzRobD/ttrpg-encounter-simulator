package roll_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"math/rand/v2"
)

type RollManager struct {
	parent             core.Entity
	RerollAbilities    RerollAbilities
	luckyUsesRemaining int
	rng                *rand.Rand
}

type RollOptions struct {
	// Base roll settings
	Advantage core.AdvantageType
	Modifier  int

	// Reroll abilities
	//RerollAbilities RerollAbilities

	// Other modifiers
	CriticalThreshold int
	TreatOnesAsTwos   bool // Elemental Adept

	// Context for logging
	RollType core.DiceRollType

	// Target for success
	TargetValue int
}

type RerollAbilities struct {
	HasHalflingLucky       bool
	HasLuckyFeat           bool
	HasElvenAccuracy       bool
	HasGreatWeaponFighting bool
	HasElementalAdept      bool
}

type RollResult struct {
	DiceRollType   core.DiceRollType
	NumberOfDice   int
	Die            core.DiceType
	FinalRollValue int
	FinalRolls     []int
	Modifier       int
	Total          int
	Advantage      core.AdvantageType

	// Reroll tracking
	OriginalRolls []int
	RerollEvents  []RerollEvent
	WasRerolled   bool

	// Special results
	IsCritical   bool
	IsNaturalOne bool
	IsSuccess    bool
	TargetValue  int
}

func (rr RollResult) GetDiceRollType() core.DiceRollType { return rr.DiceRollType }
func (rr RollResult) GetNumberOfDice() int               { return rr.NumberOfDice }
func (rr RollResult) GetDiceType() string                { return rr.Die.String() }
func (rr RollResult) GetFinalRollValue() int             { return rr.FinalRollValue }
func (rr RollResult) GetFinalRolls() []int               { return rr.FinalRolls }
func (rr RollResult) GetModifier() int                   { return rr.Modifier }
func (rr RollResult) GetTotal() int                      { return rr.Total }
func (rr RollResult) GetAdvantage() string               { return rr.Advantage.String() }
func (rr RollResult) GetOriginalRolls() []int            { return rr.OriginalRolls }
func (rr RollResult) GetWasRerolled() bool               { return rr.WasRerolled }
func (rr RollResult) GetIsCritical() bool                { return rr.IsCritical }
func (rr RollResult) GetIsNaturalOne() bool              { return rr.IsNaturalOne }
func (rr RollResult) GetIsSuccess() bool                 { return rr.IsSuccess }
func (rr RollResult) GetTargetValue() int                { return rr.TargetValue }
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
	Die          core.DiceType
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

func NewRollManager(parent core.Entity, abilities RerollAbilities) *RollManager {
	return &RollManager{
		parent:             parent,
		luckyUsesRemaining: 3, // Lucky feat gives 3 uses
		RerollAbilities:    abilities,
	}
}

func (rm *RollManager) RollD20(options RollOptions) (*RollResult, error) {
	var res RollResult // Single return value
	res.Advantage = options.Advantage
	res.Modifier = options.Modifier
	res.NumberOfDice = 1
	res.Die = core.D20

	// Handle d20 rolls with advantage/disadvantage
	switch options.Advantage {
	case core.RollNormal:
		roll := rm.rollDie(core.D20)
		res.OriginalRolls = []int{roll}
		res.FinalRolls = []int{roll}
	case core.RollAdvantage:
		_, rolls := rm.rollDice(2, core.D20)
		res.OriginalRolls = rolls
		res.FinalRolls = rolls
	case core.RollDisadvantage:
		_, rolls := rm.rollDice(2, core.D20)
		res.OriginalRolls = rolls
		res.FinalRolls = rolls
	}

	// Apply Halfling Lucky
	canUseLucky := options.RollType == core.DiceRollAttack ||
		options.RollType == core.DiceRollSavingThrow ||
		options.RollType == core.DiceRollAbilityCheck
	if containsOnes(res.OriginalRolls) && canUseLucky {
		newRolls, rEvents := rm.applyHalflingLucky(res.OriginalRolls, core.D20)
		if len(rEvents) > 0 {
			res.FinalRolls = newRolls
			res.RerollEvents = append(res.RerollEvents, rEvents...)
			res.WasRerolled = true
		}
	}

	rm.calculateSingleDieFinalValue(&res)
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
	// TODO: Handle alert's +5 modifier -> characters with the feat should add bonus to opts

	options.Modifier += mod
	options.RollType = core.DiceRollInitiative

	res, err := rm.RollD20(options)
	if err != nil {
		return nil, err
	}

	// Log initiative roll
	events.LogDiceRollEvent(rm.parent, res, rm.parent.GetEventListener())

	return res, nil
}

func (rm *RollManager) RollDamage(req core.AttackRequest, isCritical bool, opts RollOptions) (*RollResult, error) {
	// Handle damage rolls
	// Apply Great Weapon Fighting, Elemental Adept
	// Log all events
	var res RollResult

	// Calculate the appropriate damage modifier from the attack request
	dmgMod := req.GetAttackData().GetDamageModifier()
	if !req.GetAttackOptions().GetShouldApplyDamageMod() {
		dmgMod = 0
	}
	dmgMod += req.GetAttackOptions().GetBonusToAttackRoll()
	if req.GetAttackOptions().GetIsPowerAttack() {
		dmgMod += 10
	}

	// Calculate the appropriate amount of damage
	var dmgRollTotal int
	var dmgRolls []int

	attackData := req.GetAttackData()
	numDice := attackData.GetNumberOfDice()
	die := attackData.GetDie()

	// Determine if this attack can crit
	crit := isCritical &&
		((rm.parent.IsCharacter() && req.GetSimulationOptions().CanCharactersCrit) ||
			(rm.parent.IsMonster() && req.GetSimulationOptions().CanMonstersCrit))

	if crit {
		if req.GetSimulationOptions().UseImprovedCriticals {
			dmgRollTotal, dmgRolls = rm.rollExtraMaxDice(numDice, die)
		} else {
			dmgRollTotal, dmgRolls = rm.rollDoubleDice(numDice, die)
		}
	} else {
		dmgRollTotal, dmgRolls = rm.rollDice(numDice, die)
	}

	// Configure result
	res.DiceRollType = opts.RollType
	res.NumberOfDice = numDice
	res.Die = die
	res.FinalRollValue = dmgRollTotal
	res.FinalRolls = dmgRolls
	res.Modifier = dmgMod
	res.Total = dmgRollTotal
	res.Advantage = opts.Advantage
	res.OriginalRolls = dmgRolls

	// apply modifiers
	// Great Weapon Fighting - Reroll 1s and 2s
	if containsOnesOrTwos(dmgRolls) {
		newRolls, rEvents := rm.applyGreatWeaponFighting(dmgRolls, die)
		if len(rEvents) > 0 {
			res.FinalRolls = newRolls
			res.RerollEvents = append(res.RerollEvents, rEvents...)
			res.WasRerolled = true
			res.FinalRollValue = sum(res.FinalRolls)
		}
	}

	// log rolls
	events.LogDiceRollEvent(rm.parent, &res, rm.parent.GetEventListener())

	return &res, nil
}

func (rm *RollManager) RollSpellValue(req core.SpellCastRequest, isCritical bool, opts RollOptions) (*RollResult, error) {
	var res RollResult

	var valueMod int
	if !req.GetSpellCastData().GetSpellChoice().GetFormula().GetUseSpellModifier() {
		// Get the caster's spell ability modifier
		valueMod = req.GetSpellCastData().GetSpellcastingModifier()
	} else {
		valueMod = req.GetSpellCastData().GetSpellChoice().GetFormula().GetAmountToAdd()
	}

	// Calculate the appropriate amount of damage
	var valRollTotal int
	var valRolls []int

	formula := req.GetSpellCastData().GetSpellChoice().GetFormula()
	numDice := formula.GetNumberOfDice()
	die := formula.GetDie()

	// Determine if this spell can crit (spell attacks only, not saving throws)
	crit := !req.GetSpellCastData().GetSpellChoice().GetSpell().GetHasDC() &&
		isCritical &&
		((rm.parent.IsCharacter() && req.GetSimulationOptions().CanCharactersCrit) ||
			(rm.parent.IsMonster() && req.GetSimulationOptions().CanMonstersCrit))

	if crit {
		if req.GetSimulationOptions().UseImprovedCriticals {
			valRollTotal, valRolls = rm.rollExtraMaxDice(numDice, die)
		} else {
			valRollTotal, valRolls = rm.rollDoubleDice(numDice, die)
		}
	} else {
		valRollTotal, valRolls = rm.rollDice(numDice, die)
	}

	// Configure result
	res.DiceRollType = opts.RollType
	res.NumberOfDice = numDice
	res.Die = die
	res.FinalRollValue = valRollTotal
	res.FinalRolls = valRolls
	res.Modifier = valueMod
	res.Total = valRollTotal
	res.Advantage = opts.Advantage
	res.OriginalRolls = valRolls

	// Apply modifiers
	// Elemental Adept
	if containsOnes(res.OriginalRolls) {
		newRolls, rEvents := rm.applyElementalAdept(valRolls, die)
		if len(rEvents) > 0 {
			res.FinalRolls = newRolls
			res.RerollEvents = append(res.RerollEvents, rEvents...)
			res.WasRerolled = true
			res.FinalRollValue = sum(res.FinalRolls)
		}
	}

	// log rolls
	events.LogDiceRollEvent(rm.parent, &res, rm.parent.GetEventListener())

	return &res, nil
}

// RollSavingThrow rolls a d20 for a saving throw, applies bonuses and modifiers, and logs the result. Returns the roll result.
func (rm *RollManager) RollSavingThrow(ability core.Ability, options RollOptions) (*RollResult, error) {
	options.RollType = core.DiceRollSavingThrow

	res, err := rm.RollD20(options)
	if err != nil {
		return nil, err
	}

	rm.calculateSuccess(res, options)

	// Log the roll
	events.LogDiceRollEvent(rm.parent, res, rm.parent.GetEventListener())

	res.Total += 8

	return res, nil
}

func (rm *RollManager) RollAttack(options RollOptions) (*RollResult, error) {
	res, err := rm.RollD20(options)
	if err != nil {
		return nil, err
	}

	rm.calculateSuccess(res, options)

	return res, nil
}

func (rm *RollManager) RollAbilityCheck(ability core.Ability, options RollOptions) (*RollResult, error) {
	mod, err := rm.parent.GetAbilityScoreModifier(ability)
	if err != nil {
		return nil, err
	}
	options.Modifier += mod
	options.RollType = core.DiceRollAbilityCheck

	res, err := rm.RollD20(options)
	if err != nil {
		return nil, err
	}

	rm.calculateSuccess(res, options)
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
	rm.luckyUsesRemaining = 3
}

func (rm *RollManager) applyElementalAdept(rolls []int, die core.DiceType) ([]int, []RerollEvent) {
	if !rm.RerollAbilities.HasElementalAdept {
		return nil, nil
	}

	var rerollEvents []RerollEvent
	newRolls := make([]int, len(rolls))
	copy(newRolls, rolls)

	for i, roll := range newRolls {
		if roll == 1 {
			newRolls[i] = 2

			rerollEvents = append(rerollEvents, RerollEvent{
				Reason:       RerollElementalAdept.String(),
				OriginalRoll: 1,
				NewRoll:      2,
				Die:          die,
				RerollType:   RerollElementalAdept,
			})
		}
	}

	return newRolls, rerollEvents
}

func (rm *RollManager) applyGreatWeaponFighting(rolls []int, die core.DiceType) ([]int, []RerollEvent) {
	if !rm.RerollAbilities.HasGreatWeaponFighting {
		return nil, nil
	}

	var rerollEvents []RerollEvent
	newRolls := make([]int, len(rolls))
	copy(newRolls, rolls)

	for i, roll := range newRolls {
		if roll == 1 || roll == 2 {
			originalRoll := roll
			newRoll := rm.rollDie(die)
			newRolls[i] = newRoll

			rerollEvents = append(rerollEvents, RerollEvent{
				Reason:       RerollGWF.String(),
				OriginalRoll: originalRoll,
				NewRoll:      newRoll,
				Die:          die,
				RerollType:   RerollGWF,
			})
		}
	}

	return newRolls, rerollEvents
}

// applyHalflingLucky applies the Halfling Lucky trait to reroll dice rolls of 1, returning the updated rolls and reroll events.
func (rm *RollManager) applyHalflingLucky(rolls []int, die core.DiceType) ([]int, []RerollEvent) {
	if !rm.RerollAbilities.HasHalflingLucky {
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

func (rm *RollManager) calculateSingleDieFinalValue(res *RollResult) {
	switch res.Advantage {
	case core.RollNormal:
		res.FinalRollValue = res.FinalRolls[0]
		res.Total = res.FinalRollValue + res.Modifier
	case core.RollAdvantage:
		res.FinalRollValue = highest(res.FinalRolls)
		res.Total = res.FinalRollValue + res.Modifier
	case core.RollDisadvantage:
		res.FinalRollValue = lowest(res.FinalRolls)
		res.Total = res.FinalRollValue + res.Modifier
	}
}

func (rm *RollManager) calculateSuccess(res *RollResult, options RollOptions) {
	res.IsSuccess = res.FinalRollValue >= options.TargetValue
	res.TargetValue = options.TargetValue
}

func (rm *RollManager) rollDice(numberOfDice int, die core.DiceType) (int, []int) {
	rolls := make([]int, numberOfDice)
	for i := 0; i < numberOfDice; i++ {
		rolls[i] = rm.rng.IntN(die.Int()) + 1
	}

	return sum(rolls), rolls
}

func (rm *RollManager) rollDoubleDice(numberOfDice int, die core.DiceType) (int, []int) {
	rolls := make([]int, numberOfDice*2)
	for i := 0; i < numberOfDice*2; i++ {
		rolls[i] = rm.rng.IntN(die.Int()) + 1
	}

	return sum(rolls), rolls
}

func (rm *RollManager) rollExtraMaxDice(numberOfDice int, die core.DiceType) (int, []int) {
	rolls := make([]int, numberOfDice*2)
	for i := 0; i < numberOfDice*2; i++ {
		if i >= numberOfDice {
			rolls[i] = die.Int()
		} else {
			rolls[i] = rm.rng.IntN(die.Int()) + 1
		}
	}

	return sum(rolls), rolls
}

func (rm *RollManager) rollDie(die core.DiceType) int {
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
	return r
}

func containsOnes(arr []int) bool {
	for _, v := range arr {
		if v == 1 {
			return true
		}
	}
	return false
}

func containsOnesOrTwos(arr []int) bool {
	for _, v := range arr {
		if v == 1 || v == 2 {
			return true
		}
	}
	return false
}
