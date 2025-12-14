package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
)

func (c *Character) RollInitiative() (int, error) {
	var err error
	opts := roll_manager.NewRollOptions()
	opts.Modifier, err = c.getAbilityScoreModifier(core.AbilityDexterity)
	if err != nil {
		return 0, err
	}

	res, err := c.RollManager.RollInitiative(opts)
	if err != nil {
		return 0, err
	}

	c.EntityState.SetInitiative(res.Total)

	return res.Total, nil
}

// CreateWeaponAttackData generates an AttackData object for a given weapon slot, considering proficiency and versatility.
// slot specifies the weapon slot to retrieve the weapon from.
// useVersatile indicates whether to use the weapon in versatile mode, if applicable.
// Returns the constructed AttackData and an error if any issue occurs in retrieving or calculating weapon properties.
func (c *Character) CreateWeaponAttackData(slot core.WeaponSlot, useVersatile bool) (core.AttackData, error) {
	w, err := c.EquipmentManager.GetWeaponFromSlot(slot)
	if err != nil {
		return core.AttackData{}, err
	}

	prof := c.EquipmentManager.GetIsProficientWithSlot(slot)

	attackMod, err := w.GetAttackModifier(&c.AbilityScores, c.Level, prof)
	if err != nil {
		return core.AttackData{}, err
	}

	damageMod, err := w.GetWeaponModifier(&c.AbilityScores)
	if err != nil {
		return core.AttackData{}, err
	}

	die := w.Die
	var v bool
	if useVersatile && w.IsVersatile {
		die = w.Die + 2
		v = true
	}

	return core.AttackData{
		Name:              w.Name,
		NumberOfDice:      w.NumberOfDice,
		Die:               die,
		AttackModifier:    attackMod,
		DamageModifier:    damageMod,
		DamageType:        w.DamageType,
		IsVersatileAttack: v,
	}, nil
}

// CreateAttackRequest generates an attack request with specific weapon data, modifiers, advantage type, and attack count.
func (c *Character) CreateAttackRequest(target core.Entity, slot core.WeaponSlot, adv core.AdvantageType, useVersatile bool, simulationOptions *core.SimulationOptions) (*core.AttackRequest, error) {
	attackData, err := c.EquipmentManager.GetWeaponAttackData(slot, useVersatile)
	if err != nil {
		return nil, err
	}

	// Determine ranged vs melee for condition rules
	isRanged := false
	if w, wErr := c.EquipmentManager.GetWeaponFromSlot(slot); wErr == nil {
		isRanged = w.IsRanged
	}

	// Build advantage from attacker/target conditions (minimal set). Respect incoming adv as baseline.
	computedAdv := c.computeAttackAdvantage(target, isRanged, adv)

	// Build attack options (minimal; feat-driven bonuses deferred to PF section)
	improvedCrit := false
	if simulationOptions != nil && simulationOptions.UseImprovedCriticals {
		improvedCrit = true
	}

	attackOptions := core.AttackOptions{
		NumberOfAttacks:      c.EntityState.GetNumberOfAttacks(),
		BonusToAttackRoll:    0,
		BonusToDamageRoll:    0,
		ShouldApplyDamageMod: true,
		PowerAttack:          false,
		ImprovedCritical:     improvedCrit,
		RerollOnesAndTwos:    false,
		Advantage:            computedAdv,
	}

	return &core.AttackRequest{
		AttackData:        []core.AttackData{attackData},
		AttackOptions:     attackOptions,
		SimulationOptions: simulationOptions,
		Target:            target,
	}, nil
}

// computeAttackAdvantage derives advantage/disadvantage from basic conditions.
// It respects the passed-in adv as a baseline and then applies condition-based modifiers:
// - Attacker blinded or poisoned: disadvantage
// - Target prone: melee attacks advantage, ranged attacks disadvantage
// - Target restrained/paralyzed/unconscious: advantage
// If both advantage and disadvantage are present, result is Normal.
func (c *Character) computeAttackAdvantage(target core.Entity, isRanged bool, base core.AdvantageType) core.AdvantageType {
	hasAdv := false
	hasDis := false

	// Baseline from caller
	switch base {
	case core.RollAdvantage:
		hasAdv = true
	case core.RollDisadvantage:
		hasDis = true
	}

	attackerConds := c.GetConditions()
	targetConds := target.GetConditions()

	// Attacker conditions
	if attackerConds.Has(core.ConditionBlinded) {
		hasDis = true
	}
	if attackerConds.Has(core.ConditionPoisoned) {
		hasDis = true
	}

	// Target conditions
	if targetConds.Has(core.ConditionProne) {
		if isRanged {
			hasDis = true
		} else {
			hasAdv = true
		}
	}
	if targetConds.Has(core.ConditionRestrained) || targetConds.Has(core.ConditionParalyzed) || targetConds.Has(core.ConditionUnconscious) {
		hasAdv = true
	}

	// Resolve
	if hasAdv && hasDis {
		return core.RollNormal
	} else if hasAdv {
		return core.RollAdvantage
	} else if hasDis {
		return core.RollDisadvantage
	}
	return core.RollNormal
}

// MakeSavingThrow calculates a saving throw roll using the specified ability and returns the result, rolls, and an error if any.
func (c *Character) MakeSavingThrow(ability core.Ability, targetValue int) (core.RollResult, error) {
	activeConditions := c.GetConditions().GetActive()
	isStrDexSave := ability == core.AbilityStrength || ability == core.AbilityDexterity

	autoFailResult := func() core.RollResult {
		result := roll_manager.RollResult{
			DiceRollType:   core.DiceRollSavingThrow,
			NumberOfDice:   0,
			Die:            0,
			FinalRollValue: 0,
			FinalRolls:     []int{0},
			Modifier:       0,
			Total:          0,
			Advantage:      core.RollNormal,
			IsSuccess:      false,
			TargetValue:    targetValue,
		}
		events.LogDiceRollEvent(c, &result, c.EventListener)
		return result
	}

	if isStrDexSave && len(activeConditions) > 0 {
		for _, cond := range activeConditions {
			effect := core.GetConditionEffects(cond)
			if effect.AutoFailStrDexSave {
				return autoFailResult(), nil
			}
		}
	}

	mod, err := c.getSavingThrowBonus(ability)
	if err != nil {
		return nil, err
	}

	opts := roll_manager.NewRollOptions()
	baseAdv := c.EntityState.GetSavingThrowAdvantage(ability) // This should get the default advantage of the character
	condAdv := core.DetermineSaveAdvantageFromConditions(c.GetConditions(), ability)
	opts.Advantage = core.GetFinalAdvantageType([]core.AdvantageType{baseAdv, condAdv})
	opts.Modifier = mod
	opts.RollType = core.DiceRollSavingThrow
	opts.TargetValue = targetValue

	res, err := c.RollManager.RollSavingThrow(ability, opts)
	if err != nil {
		return nil, err
	}

	return res, nil
}
