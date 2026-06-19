package roll_manager

import (
	"dnd5e-encounter-simulator-backend/pkg_old/classes"
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg_old/core/events"
	"fmt"
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

	// ctx for logging
	RollType core.DiceRollType

	// Target for success
	TargetValue int
}

// NewRollOptions initializes and returns a default RollOptions instance with predefined values for roll settings.
// Advantage: Normal
// Modifier: 0
// Critical Threshold: 20
// TreatOnesAsTwos: false
// RollType: ""
// TargetValue: 0;
func NewRollOptions() RollOptions {
	return RollOptions{
		Advantage:         core.RollNormal,
		Modifier:          0,
		CriticalThreshold: 20,
		TreatOnesAsTwos:   false,
		RollType:          "",
		TargetValue:       0,
	}
}

type RerollAbilities struct {
	HasHalflingLucky       bool `json:"has_halfling_lucky"`
	HasElvenAccuracy       bool `json:"has_elven_accuracy"`
	HasGreatWeaponFighting bool `json:"has_great_weapon_fighting"`
	HasElementalAdept      bool `json:"has_elemental_adept"`
	HasIndomitable         bool `json:"has_indomitable"`
}

type RollResult struct {
	ID             string             `json:"id"`
	DiceRollType   core.DiceRollType  `json:"dice_roll_type"`
	FinalRollValue int                `json:"final_roll_value"`
	FinalRolls     []int              `json:"final_rolls"`
	Modifier       int                `json:"modifier"`
	Total          int                `json:"total"`
	Advantage      core.AdvantageType `json:"advantage"`

	// Compatibility fields
	NumberOfDice int           `json:"number_of_dice"`
	Die          core.DiceType `json:"die"`

	// Multi-damage support
	DamageComponents []DamageComponent `json:"damage_components"`

	// Reroll tracking
	OriginalRolls []int         `json:"original_rolls"`
	RerollEvents  []RerollEvent `json:"reroll_events"`
	WasRerolled   bool          `json:"was_rerolled"`

	// Special results
	IsCritical   bool   `json:"is_critical"`
	IsNaturalOne bool   `json:"is_natural_one"`
	IsSuccess    bool   `json:"is_success"`
	TargetValue  int    `json:"target_value"`
	Name         string `json:"name"` // Used for recharge only

	Target core.Entity `json:"-"`
}

type DamageComponent struct {
	DamageType   core.DamageType `json:"damage_type"`
	RollValue    int             `json:"roll_value"`
	Modifier     int             `json:"modifier"`
	Total        int             `json:"total"`
	DiceRolls    []int           `json:"dice_rolls"`
	NumberOfDice int             `json:"number_of_dice"`
	Die          core.DiceType   `json:"die"`
	IsCritical   bool            `json:"is_critical"`
	RerollEvents []RerollEvent   `json:"reroll_events"`
}

func (rr *RollResult) GetID() string                      { return rr.ID }
func (rr *RollResult) SetID(id string)                    { rr.ID = id }
func (rr *RollResult) GetTarget() core.Entity             { return rr.Target }
func (rr *RollResult) GetDiceRollType() core.DiceRollType { return rr.DiceRollType }
func (rr *RollResult) GetNumberOfDice() int {
	if rr.NumberOfDice > 0 {
		return rr.NumberOfDice
	}
	if len(rr.DamageComponents) > 0 {
		return rr.DamageComponents[0].NumberOfDice
	}
	return 0
}
func (rr *RollResult) GetDiceType() string {
	if rr.Die > 0 {
		return rr.Die.String()
	}
	if len(rr.DamageComponents) > 0 {
		return rr.DamageComponents[0].Die.String()
	}
	return ""
}
func (rr *RollResult) GetFinalRollValue() int  { return rr.FinalRollValue }
func (rr *RollResult) GetFinalRolls() []int    { return rr.FinalRolls }
func (rr *RollResult) GetModifier() int        { return rr.Modifier }
func (rr *RollResult) GetTotal() int           { return rr.Total }
func (rr *RollResult) GetAdvantage() string    { return rr.Advantage.String() }
func (rr *RollResult) GetOriginalRolls() []int { return rr.OriginalRolls }
func (rr *RollResult) GetWasRerolled() bool    { return rr.WasRerolled }
func (rr *RollResult) GetIsCritical() bool     { return rr.IsCritical }
func (rr *RollResult) GetIsNaturalOne() bool   { return rr.IsNaturalOne }
func (rr *RollResult) GetIsSuccess() bool      { return rr.IsSuccess }
func (rr *RollResult) GetTargetValue() int     { return rr.TargetValue }
func (rr *RollResult) GetDamageComponents() []core.DamageComponent {
	var comps []core.DamageComponent
	for i := range rr.DamageComponents {
		comps = append(comps, &rr.DamageComponents[i])
	}
	return comps
}

func (dc *DamageComponent) GetDamageType() core.DamageType { return dc.DamageType }
func (dc *DamageComponent) GetTotal() int                  { return dc.Total }

func (rr *RollResult) GetRerollEvents() []map[string]interface{} {
	var r []map[string]interface{}
	for _, event := range rr.RerollEvents {
		m := make(map[string]interface{})
		m["reason"] = event.Reason
		m["original_roll"] = event.OriginalRoll
		m["new_roll"] = event.NewRoll
		m["die"] = event.Die.String()
		m["reroll_type"] = event.RerollType.String()
		r = append(r, m)
	}
	return r
}

type RerollEvent struct {
	Reason       string        `json:"reason"`
	OriginalRoll int           `json:"original_roll"`
	NewRoll      int           `json:"new_roll"`
	Die          core.DiceType `json:"die"`
	RerollType   RerollType    `json:"reroll_type"`
}

type RerollType string

const (
	RerollHalflingLucky  RerollType = "Halfling Lucky"
	RerollLuckyFeat      RerollType = "Lucky Feat"
	RerollGWF            RerollType = "Great Weapon Fighting"
	RerollElementalAdept RerollType = "Elemental Adept"
	RerollIndomitable    RerollType = "Fighter Indomitable"
)

func (rt RerollType) String() string {
	return string(rt)
}

func NewRollManager(parent core.Entity, abilities RerollAbilities) *RollManager {
	rm := RollManager{
		parent:             parent,
		luckyUsesRemaining: 3, // Lucky feat gives 3 uses
		RerollAbilities:    abilities,
		rng:                parent.GetRNG(),
	}

	return &rm
}

func (rm *RollManager) SetRNG(seed1, seed2 uint64) {
	rm.rng = rand.New(rand.NewPCG(seed1, seed2))
}

// RollD20 performs a d20 roll based on the given options, handles advantage/disadvantage, and calculates the final result.
// RollDice rolls the specified number of dice of the given type and returns a RollResult.
func (rm *RollManager) RollDice(numberOfDice int, die core.DiceType, opts RollOptions) (*RollResult, error) {
	total, rolls := rm.rollDice(numberOfDice, die)
	res := &RollResult{
		DiceRollType:   opts.RollType,
		NumberOfDice:   numberOfDice,
		Die:            die,
		FinalRollValue: total,
		FinalRolls:     rolls,
		Modifier:       opts.Modifier,
		Total:          total + opts.Modifier,
	}

	return res, nil
}

func (rm *RollManager) RollD20(options RollOptions, shouldLogEvent bool) (*RollResult, error) {
	var res RollResult // Single return value
	res.Advantage = options.Advantage
	res.Modifier = options.Modifier
	res.NumberOfDice = 1
	res.Die = core.D20
	res.DiceRollType = options.RollType

	// Handle d20 rolls with advantage/disadvantage
	switch options.Advantage {
	case core.RollNormal:
		roll := rm.RollDie(core.D20)
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
		options.RollType == core.DiceRollAbilityCheck ||
		options.RollType == core.DiceRollDeathSavingThrow
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

	rm.calculateSuccess(&res, options)

	// Log all events
	if shouldLogEvent {
		rm.parent.LogEvent(events.ETRollEvent, &res)
	}

	return &res, nil
}

// RollInitiative rolls a d20 for initiative and applies applicable modifiers, returning the result or an error if any occur.
// Options required: Advantage, Modifier
func (rm *RollManager) RollInitiative(options RollOptions) (*RollResult, error) {
	// Roll a d20 and apply applicable modifiers
	mod, err := core.GetAbilityScoreModifier(rm.parent.GetAbilityScores().Dexterity)
	if err != nil {
		return nil, err
	}

	options.Modifier += mod
	options.RollType = core.DiceRollInitiative

	res, err := rm.RollD20(options, false)
	if err != nil {
		return nil, err
	}
	res.DiceRollType = options.RollType

	// Log initiative roll
	rm.parent.LogEvent(events.ETRollInitiative, res)

	return res, nil
}

func (rm *RollManager) RollRecharge(options RollOptions) *RollResult {
	roll := rm.RollDie(core.D6)
	res := RollResult{
		DiceRollType:   options.RollType,
		NumberOfDice:   1,
		Die:            core.D6,
		FinalRollValue: roll,
		FinalRolls:     []int{roll},
		Modifier:       0,
		Total:          roll,
		Advantage:      core.RollNormal,
		OriginalRolls:  []int{roll},
		TargetValue:    options.TargetValue,
		IsSuccess:      roll >= options.TargetValue,
	}

	return &res
}

// RollDamage rolls damage dice for an attack, applies modifiers, handles reroll rules, critical hits, and logs events.
func (rm *RollManager) RollDamage(req *core.AttackRequest, adIndex int, isCritical bool, opts RollOptions, shouldLogEvent bool) (*RollResult, error) {
	var res RollResult
	res.DiceRollType = opts.RollType
	res.Advantage = opts.Advantage

	attackData := req.GetAttackData()[adIndex]

	// Determine if this attack can crit
	canCrit := false
	if req.GetSimulationOptions() != nil {
		canCrit = (rm.parent.IsCharacter() && req.GetSimulationOptions().CanCharactersCrit) ||
			(rm.parent.IsMonster() && req.GetSimulationOptions().CanMonstersCrit)
	}
	crit := isCritical && canCrit

	extraCritDice := req.GetAttackOptions().ExtraCritDice

	var totalDamage int
	var totalRollValue int
	var allRolls []int

	for i, db := range attackData.DamageBlocks {
		var dmgRollTotal int
		var dmgRolls []int

		numDice := db.NumberOfDice
		die := db.Die

		if crit {
			if req.GetSimulationOptions() != nil && req.GetSimulationOptions().UseImprovedCritical {
				dmgRollTotal, dmgRolls = rm.RollExtraMaxDice(numDice, die)
			} else {
				dmgRollTotal, dmgRolls = rm.rollDice(numDice*2, die)
			}
			// Only add extra crit dice to the first damage block (usually the main weapon damage)
			if i == 0 && extraCritDice > 0 {
				extraRollTotal, extraRoll := rm.rollDice(extraCritDice, die)
				dmgRollTotal += extraRollTotal
				dmgRolls = append(dmgRolls, extraRoll...)
			}
		} else {
			dmgRollTotal, dmgRolls = rm.rollDice(numDice, die)
		}

		// Calculate modifier for this block
		blockMod := db.Modifier
		if i == 0 {
			// Apply global modifiers to the first block
			globalMod := attackData.GetDamageModifier()
			if !req.GetAttackOptions().GetShouldApplyDamageMod() {
				globalMod = 0
			} else {
				globalMod += req.GetAttackOptions().GetBonusToDamageRoll()
			}
			if req.GetAttackOptions().GetIsPowerAttack() {
				globalMod += 10
			}
			blockMod += globalMod
		}

		// Great Weapon Fighting - Reroll 1s and 2s
		var rEvents []RerollEvent
		if containsOnesOrTwos(dmgRolls) {
			newRolls, events := rm.applyGreatWeaponFighting(dmgRolls, die)
			if len(events) > 0 {
				dmgRolls = newRolls
				dmgRollTotal = sum(dmgRolls)
				rEvents = append(rEvents, events...)
				res.RerollEvents = append(res.RerollEvents, events...)
				res.WasRerolled = true
			}
		}

		comp := DamageComponent{
			DamageType:   db.DamageType,
			RollValue:    dmgRollTotal,
			Modifier:     blockMod,
			Total:        dmgRollTotal + blockMod,
			DiceRolls:    dmgRolls,
			NumberOfDice: len(dmgRolls),
			Die:          die,
			IsCritical:   crit,
			RerollEvents: rEvents,
		}

		res.DamageComponents = append(res.DamageComponents, comp)
		totalDamage += comp.Total
		totalRollValue += dmgRollTotal
		allRolls = append(allRolls, dmgRolls...)
	}

	res.NumberOfDice = len(allRolls)
	if len(attackData.DamageBlocks) > 0 {
		res.Die = attackData.DamageBlocks[0].Die
	}

	res.FinalRollValue = totalRollValue
	res.FinalRolls = allRolls
	res.Total = totalDamage
	res.OriginalRolls = allRolls // Simplified, might need more detail for exact tracking
	res.IsCritical = crit

	// Log damage roll
	if shouldLogEvent {
		rm.parent.LogEvent(events.ETRollEvent, &res)
	}

	return &res, nil
}

func (rm *RollManager) RollSpellValue(req core.SpellCastRequest, isCritical bool, opts RollOptions, shouldLogEvent bool) (*RollResult, error) {
	var res RollResult

	var valueMod int
	if req.GetSpellCastData().GetSpellChoice().GetFormula().GetUseSpellModifier() {
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
		if req.GetSimulationOptions().UseImprovedCritical {
			valRollTotal, valRolls = rm.RollExtraMaxDice(numDice, die)
		} else {
			valRollTotal, valRolls = rm.rollDice(numDice*2, die)
		}
	} else {
		valRollTotal, valRolls = rm.rollDice(numDice, die)
	}

	// Configure result
	res.DiceRollType = opts.RollType
	res.NumberOfDice = len(valRolls)
	res.Die = die
	res.FinalRollValue = valRollTotal
	res.FinalRolls = valRolls
	res.Modifier = valueMod
	// Set initial total (will be finalized after any rerolls/adjustments below)
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

	// Finalize total to include modifier after any rerolls/adjustments
	res.Total = res.FinalRollValue + res.Modifier

	// log rolls
	if shouldLogEvent {
		rm.parent.LogEvent(events.ETRollEvent, &res)
	}

	return &res, nil
}

// RollSavingThrow rolls a d20 for a saving throw, applies bonuses and modifiers, and logs the result. Returns the roll result.
func (rm *RollManager) RollSavingThrow(options RollOptions) (*RollResult, error) {
	options.RollType = core.DiceRollSavingThrow

	res, err := rm.RollD20(options, false)
	if err != nil {
		return nil, err
	}

	// Fighter indomitable
	if rm.parent.GetClassID() == uint8(classes.Fighter) && !res.IsSuccess {
		// TODO: If sim options sets no class features, we need to set uses to zero
		newRoll, rEvent, rErr := rm.applyFighterIndomitable(res.FinalRollValue)
		if rErr != nil {
			return nil, rErr
		}
		if rEvent.Reason != "" {
			res.RerollEvents = append(res.RerollEvents, rEvent)
			res.WasRerolled = true
			res.FinalRolls = []int{newRoll}
			res.FinalRollValue = newRoll
			res.IsNaturalOne = newRoll == 1
			res.Total = newRoll + res.Modifier
		}
	}

	rm.calculateSuccess(res, options)

	// Log the roll
	rm.parent.LogEvent(events.ETRollEvent, res)

	return res, nil
}

func (rm *RollManager) RollAttack(options RollOptions) (*RollResult, error) {
	res, err := rm.RollD20(options, false)
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

	res, err := rm.RollD20(options, false)
	if err != nil {
		return nil, err
	}

	rm.calculateSuccess(res, options)

	// Log ability check roll
	rm.parent.LogEvent(events.ETRollEvent, res)

	return res, nil
}

// RollHP rolls hit points based on the provided configuration and returns the result or an error if input is invalid.
// Uses the parent type to determine if rolling for a character or a monster.
// Logs dice roll events using the parent's event listener.
func (rm *RollManager) RollHP(config core.HPConfig) (*RollResult, error) {
	if config.NumberOfDice <= 0 || config.HitDie == 0 {
		return nil, fmt.Errorf("invalid HP configuration: NumberOfDice=%d, hitDie=%v",
			config.NumberOfDice, config.HitDie)
	}
	var res RollResult
	options := NewRollOptions()
	options.Advantage = core.RollNormal
	options.Modifier = config.AmountToAdd
	options.RollType = core.DiceRollHP

	if rm.parent.IsCharacter() {
		res = rm.rollCharacterHP(config, options)
	} else {
		res = rm.rollMonsterHP(config)
	}

	rm.parent.LogEvent(events.ETRollEvent, &res)

	return &res, nil
}

// UseLuckyReroll attempts to use a lucky reroll, decrementing the remaining count and returning true if available.
func (rm *RollManager) UseLuckyReroll() bool {
	if rm.luckyUsesRemaining > 0 {
		rm.luckyUsesRemaining--
		return true
	}
	return false
}

// RestoreLuckyUses resets the count of lucky uses available to the default value of 3.
func (rm *RollManager) RestoreLuckyUses() {
	rm.luckyUsesRemaining = 3
}

// RollDeathSavingThrow performs a death saving throw roll using a D20 based on specified roll options.
// Returns the result of the roll or an error if the roll could not be completed.
func (rm *RollManager) RollDeathSavingThrow() (*RollResult, error) {
	opts := NewRollOptions()
	opts.Advantage = core.RollNormal
	opts.RollType = core.DiceRollDeathSavingThrow
	opts.TargetValue = 10
	res, err := rm.RollD20(opts, false)
	if err != nil {
		return nil, fmt.Errorf("failed to roll death saving throw: %w", err)
	}

	// Log the death saving throw
	rm.parent.LogEvent(events.ETRollEvent, res)

	return res, nil
}

// rollCharacterHP calculates and returns the HP roll result for a character based on level, hit die, and modifiers.
func (rm *RollManager) rollCharacterHP(config core.HPConfig, options RollOptions) RollResult {
	if rm.parent.GetLevel() == 1 {
		// Level 1 characters get max HP
		return RollResult{
			DiceRollType:   core.DiceRollHP,
			NumberOfDice:   1,
			Die:            config.HitDie,
			FinalRollValue: config.HitDie.Int(),
			FinalRolls:     []int{config.HitDie.Int()},
			Modifier:       options.Modifier,
			Total:          config.HitDie.Int() + options.Modifier,
			OriginalRolls:  []int{config.HitDie.Int()},
		}
	}

	// For levels 2+, roll for additional HP
	cLevel := rm.parent.GetLevel()
	numDice := int(cLevel - 1)

	rollValue, rolls := rm.rollDice(numDice, config.HitDie)
	totalHP := config.HitDie.Int() + rollValue + (options.Modifier * int(cLevel))

	return RollResult{
		DiceRollType:   core.DiceRollHP,
		NumberOfDice:   numDice,
		Die:            config.HitDie,
		FinalRollValue: rollValue,
		FinalRolls:     rolls,
		Modifier:       options.Modifier,
		Total:          totalHP,
		OriginalRolls:  rolls,
	}
}

// rollMonsterHP calculates the total hit points for a monster based on the dice rolls and configuration parameters.
// It rolls the specified number of dice, adds any modifiers and a fixed amount to determine the final HP value.
// Returns a RollResult containing details of the roll, including individual rolls, dice type, modifiers, and total HP.
func (rm *RollManager) rollMonsterHP(config core.HPConfig) RollResult {
	rollValue, rolls := rm.rollDice(config.NumberOfDice, config.HitDie)
	totalHP := rollValue + config.AmountToAdd

	return RollResult{
		DiceRollType:   core.DiceRollHP,
		NumberOfDice:   config.NumberOfDice,
		Die:            config.HitDie,
		FinalRollValue: rollValue,
		FinalRolls:     rolls,
		Modifier:       config.AmountToAdd,
		Total:          totalHP,
		OriginalRolls:  rolls,
	}
}

// applyElementalAdept adjusts dice rolls to avoid results of 1 if the Elemental Adept feature is active.
// Rolls of 1 are replaced with 2, and corresponding reroll events are generated.
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

// applyGreatWeaponFighting applies the Great Weapon Fighting rule to reroll dice with values of 1 or 2 in the given rolls.
// It returns the modified rolls and a slice of RerollEvent objects detailing each reroll.
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
			newRoll := rm.RollDie(die)
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

// applyHalflingLucky adjusts the rolls by applying the Halfling Lucky feature, rerolling a single roll of 1 per dice set.
// It returns the updated rolls and a list of reroll events detailing the changes made, if any.
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
			newRoll := rm.RollDie(die)
			newRolls[i] = newRoll

			rerollEvents = append(rerollEvents, RerollEvent{
				Reason:       RerollHalflingLucky.String(),
				OriginalRoll: originalRoll,
				NewRoll:      newRoll,
				Die:          die,
				RerollType:   RerollHalflingLucky,
			})
			break
		}
	}

	return newRolls, rerollEvents
}

func (rm *RollManager) applyFighterIndomitable(originalRoll int) (int, RerollEvent, error) {
	if rm.parent.GetClassID() != uint8(classes.Fighter) {
		return 0, RerollEvent{}, nil
	}

	esm, ok := rm.parent.GetState().(*entity_state_manager.EntityStateManager)
	if !ok {
		return 0, RerollEvent{}, fmt.Errorf("failed to cast parent entity state to EntityStateManager")
	}
	if esm.GetFighterIndomitableUses() > 0 {
		newRoll := rm.RollDie(core.D20)

		rerollEvent := RerollEvent{
			Reason:       RerollIndomitable.String(),
			OriginalRoll: originalRoll,
			NewRoll:      newRoll,
			Die:          core.D20,
			RerollType:   RerollIndomitable,
		}

		esm.ExpendFighterIndomitableUses()
		return newRoll, rerollEvent, nil
	}
	return 0, RerollEvent{}, nil
}

// calculateSingleDieFinalValue determines the final roll value and total based on roll type (normal, advantage, disadvantage).
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

// calculateSuccess determines if a roll result meets or exceeds the desired target value and updates the result accordingly.
func (rm *RollManager) calculateSuccess(res *RollResult, options RollOptions) {
	res.IsSuccess = res.Total >= options.TargetValue
	res.TargetValue = options.TargetValue
}

// rollDice rolls a specified number of dice of a given type and returns the sum of rolls and individual roll results.
func (rm *RollManager) rollDice(numberOfDice int, die core.DiceType) (int, []int) {
	rolls := make([]int, numberOfDice)
	for i := 0; i < numberOfDice; i++ {
		rolls[i] = rm.rng.IntN(die.Int()) + 1
	}

	return sum(rolls), rolls
}

// RollExtraMaxDice rolls additional dice and returns the sum and the list of rolled values.
// numberOfDice specifies the number of dice to roll initially and then adds the maximum value for each dice.
// die defines the type of dice used for the rolls.
func (rm *RollManager) RollExtraMaxDice(numberOfDice int, die core.DiceType) (int, []int) {
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

// RollDie simulates rolling a die of the specified type and returns the result as an integer between 1 and the die's maximum value.
func (rm *RollManager) RollDie(die core.DiceType) int {
	return rm.rng.IntN(die.Int()) + 1
}

// sum calculates and returns the sum of all integers in the given slice.
func sum(arr []int) int {
	s := 0
	for _, v := range arr {
		s += v
	}
	return s
}

// highest returns the largest integer value from a given slice of integers.
func highest(arr []int) int {
	r := arr[0]
	for _, v := range arr {
		if v > r {
			r = v
		}
	}
	return r
}

// lowest finds and returns the smallest integer in the provided slice of integers.
func lowest(arr []int) int {
	r := arr[0]
	for _, v := range arr {
		if v < r {
			r = v
		}
	}
	return r
}

// containsOnes checks if the given slice of integers contains at least one occurrence of the integer value 1.
func containsOnes(arr []int) bool {
	for _, v := range arr {
		if v == 1 {
			return true
		}
	}
	return false
}

// containsOnesOrTwos checks if the given slice of integers contains at least one occurrence of the values 1 or 2.
func containsOnesOrTwos(arr []int) bool {
	for _, v := range arr {
		if v == 1 || v == 2 {
			return true
		}
	}
	return false
}
