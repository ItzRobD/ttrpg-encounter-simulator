package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
)

func (c *Character) RollInitiative() (int, error) {
	var err error
	opts := roll_manager.NewRollOptions()
	opts.Modifier, err = c.getAbilityScoreModifier(core.AbilityDexterity)
	if err != nil {
		return 0, err
	}
	// TODO: Handle chaaracter feats such as Alert for +5

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

	// TODO: This will have to be handled internally by other functions to get the values of each of these
	//		Will have to account for character feats
	attackOptions := core.AttackOptions{
		NumberOfAttacks:      c.EntityState.GetNumberOfAttacks(),
		BonusToAttackRoll:    0,
		BonusToDamageRoll:    0,
		ShouldApplyDamageMod: true,
		PowerAttack:          false,
		ImprovedCritical:     simulationOptions.UseImprovedCriticals,
		RerollOnesAndTwos:    false,
		Advantage:            adv,
	}

	return &core.AttackRequest{
		AttackData:        []core.AttackData{attackData},
		AttackOptions:     attackOptions,
		SimulationOptions: simulationOptions,
		Target:            target,
	}, nil
}

// MakeSavingThrow calculates a saving throw roll using the specified ability and returns the result, rolls, and an error if any.
func (c *Character) MakeSavingThrow(ability core.Ability, targetValue int) (core.RollResult, error) {
	mod, err := c.getSavingThrowBonus(ability)
	if err != nil {
		return nil, err
	}

	opts := roll_manager.NewRollOptions()
	// TODO: Determining advantage needs to be handled ie racial traits
	opts.Modifier = mod
	opts.RollType = core.DiceRollSavingThrow
	opts.TargetValue = targetValue

	res, err := c.RollManager.RollSavingThrow(ability, opts)
	if err != nil {
		return nil, err
	}

	return res, nil
}
