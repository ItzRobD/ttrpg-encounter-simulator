package character

import (
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/races"
	"fmt"
)

func (c *Character) RollInitiative() (int, error) {
	var err error
	opts := roll_manager.NewRollOptions()
	dexMod, err := c.getAbilityScoreModifier(core.AbilityDexterity)
	if err != nil {
		return 0, err
	}
	opts.Modifier = dexMod + c.Configuration.CombatFeatures.InitiativeBonus

	res, err := c.RollManager.RollInitiative(opts)
	if err != nil {
		return 0, err
	}

	c.EntityStateManager.SetInitiative(res.Total)

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

	damageMod, _, err := w.GetWeaponModifier(&c.AbilityScores)
	if err != nil {
		return core.AttackData{}, err
	}

	die := w.Die
	var v bool
	if useVersatile && w.Properties.IsVersatile {
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

func (c *Character) CreateHealRequest(target core.Entity) (*core.HealRequest, error) {
	// 1. Determine how much healing the target actually needs
	hpNeeded := target.GetHPStatus().GetHPDifference()
	if hpNeeded <= 0 {
		return nil, fmt.Errorf("target does not require healing")
	}

	// 2. Priority 1: Paladin Lay on Hands (Saves spell slots for Smite)
	if c.Class.ID == classes.Paladin {
		pool := c.EntityStateManager.GetPaladinLayingOnHandsPool()
		if pool > 0 {
			// Use just enough to top them off, or whatever is left in the pool
			amount := hpNeeded
			if amount > pool {
				amount = pool
			}

			return &core.HealRequest{
				Source:       core.HealSourceLayingOnHands,
				Target:       target,
				AbilityValue: amount,
			}, nil
		}
	}

	// 3. Priority 2: Healing Spells
	// This uses your existing logic to find the best available spell slot
	choice, err := c.ChooseSpellByHealingEfficiency(hpNeeded)
	if err == nil && choice != nil {
		return &core.HealRequest{
			Source:      core.HealSourceSpell,
			Target:      target,
			SpellChoice: choice,
		}, nil
	}

	// 4. No healing resources available
	return nil, fmt.Errorf("no healing resources available (no pool/no slots)")
}

// CreateAttackRequest generates an attack request with specific weapon data, modifiers, and attack count.
// Advantage is computed internally using core.DetermineAttackAdvantageForEntities, with a baseline derived from
// Reckless Attack (if enabled via SimulationOptions and currently active) and weapon/context.
func (c *Character) CreateAttackRequest(target core.Entity, slot core.WeaponSlot, useVersatile bool, simulationOptions *core.SimulationOptions) (*core.AttackRequest, error) {
	adSlice := make([]core.AttackData, 0)
	attackData, err := c.EquipmentManager.GetWeaponAttackData(slot, useVersatile)
	if err != nil {
		return nil, err
	}

	adSlice = append(adSlice, attackData)

	// Determine ranged vs melee for condition rules (prefer precomputed attack data)
	isRanged := attackData.IsRangedWeapon
	// For style interactions we still need to know usage context
	isVersatile := attackData.IsVersatileAttack
	isTwoHanded := attackData.IsTwoHandedWeapon

	// Derive baseline advantage from Reckless Attack if applicable
	base := core.RollNormal
	if simulationOptions != nil && simulationOptions.EnableClassFeatures {
		if c.EntityStateManager.GetIsRecklesslyAttacking() {
			if !isRanged && attackData.AbilityUsed == core.AbilityStrength {
				base = core.RollAdvantage
			}
		}
	}

	// Compute final advantage using unified core helper
	computedAdv := core.DetermineAttackAdvantageForEntities(c, target, isRanged, base)

	// Feature-driven damage mods (Rage, Brutal Critical)
	extraCritDice := 0
	bonusDmg := 0
	if simulationOptions != nil && simulationOptions.EnableClassFeatures {
		if c.Class.ID == classes.Barbarian {
			extraCritDice = c.Class.ClassFeatures.BarbarianFeatures.NumberOfBrutalCritDice
			if c.EntityStateManager.BarbarianIsRaging {
				bonusDmg += c.Class.ClassFeatures.BarbarianFeatures.RageDamage
			}
		}

		if c.Race.ID == races.HalfOrc && c.EntityStateManager.HalfOrcHasSavageAttacks {
			extraCritDice += 1
		}
	}

	attackOptions := core.AttackOptions{
		NumberOfAttacks:      c.EntityStateManager.GetNumberOfAttacks(),
		BonusToAttackRoll:    0,
		BonusToDamageRoll:    bonusDmg,
		ShouldApplyDamageMod: true,
		PowerAttack:          false,
		ImprovedCritical:     (simulationOptions != nil && simulationOptions.UseImprovedCriticals) || (c.Configuration.CombatFeatures.CriticalThreshold > 0 && c.Configuration.CombatFeatures.CriticalThreshold < 20),
		RerollOnesAndTwos:    false,
		Advantage:            computedAdv,
		ExtraCritDice:        extraCritDice,
	}

	c.applyFightingStyles(&attackData, &attackOptions, isRanged, isVersatile, isTwoHanded, slot)

	if simulationOptions != nil && simulationOptions.EnableClassFeatures {
		if c.Class.ID == classes.Paladin && c.Class.ClassFeatures.PaladinFeatures.HasImprovedDivineSmite {
			adSlice = append(adSlice, core.AttackData{
				Name:         "Improved Divine Smite",
				NumberOfDice: 1,
				Die:          8,
				DamageType:   core.DamageRadiant,
			})
		}
	}

	return &core.AttackRequest{
		AttackData:        adSlice,
		AttackOptions:     attackOptions,
		SimulationOptions: simulationOptions,
		Target:            target,
	}, nil
}

// CreateOffhandAttackRequest generates an attack request for the character's offhand weapon against a specified target.
// Advantage derived internally similar to CreateAttackRequest.
func (c *Character) CreateOffhandAttackRequest(target core.Entity, simulationOptions *core.SimulationOptions) (*core.AttackRequest, error) {
	ad, err := c.EquipmentManager.GetWeaponAttackData(core.WSSecondary, false)
	if err != nil {
		return nil, err
	}

	// Derive baseline advantage from Reckless Attack if applicable (offhand melee, STR-based)
	base := core.RollNormal
	if simulationOptions != nil && simulationOptions.EnableClassFeatures {
		if c.EntityStateManager.GetIsRecklesslyAttacking() {
			if !ad.IsRangedWeapon && ad.AbilityUsed == core.AbilityStrength {
				base = core.RollAdvantage
			}
		}
	}
	computedAdv := core.DetermineAttackAdvantageForEntities(c, target, ad.IsRangedWeapon, base)

	// Feature-driven damage tweaks (Rage applies to melee weapon damage rolls, including offhand)
	extraCritDice := 0
	bonusDmg := 0
	if simulationOptions != nil && simulationOptions.EnableClassFeatures {
		if c.Class.ID == classes.Barbarian {
			extraCritDice = c.Class.ClassFeatures.BarbarianFeatures.NumberOfBrutalCritDice
			if c.EntityStateManager.BarbarianIsRaging {
				bonusDmg += c.Class.ClassFeatures.BarbarianFeatures.RageDamage
			}
		}
	}

	opts := core.AttackOptions{
		NumberOfAttacks:      1,
		BonusToAttackRoll:    0,
		BonusToDamageRoll:    bonusDmg,
		ShouldApplyDamageMod: false, // Standard is false
		PowerAttack:          false,
		ImprovedCritical:     (simulationOptions != nil && simulationOptions.UseImprovedCriticals) || (c.Configuration.CombatFeatures.CriticalThreshold > 0 && c.Configuration.CombatFeatures.CriticalThreshold < 20),
		RerollOnesAndTwos:    false,
		Advantage:            computedAdv,
		ExtraCritDice:        extraCritDice,
	}

	c.applyFightingStyles(&ad, &opts, ad.IsRangedWeapon, ad.IsVersatileAttack, false, core.WSSecondary)

	adSlice := []core.AttackData{ad}
	if simulationOptions != nil && simulationOptions.EnableClassFeatures {
		if c.Class.ID == classes.Paladin && c.Class.ClassFeatures.PaladinFeatures.HasImprovedDivineSmite {
			adSlice = append(adSlice, core.AttackData{
				Name:         "Improved Divine Smite",
				NumberOfDice: 1,
				Die:          8,
				DamageType:   core.DamageRadiant,
			})
		}
	}

	return &core.AttackRequest{
		AttackData:        adSlice,
		AttackOptions:     opts,
		SimulationOptions: simulationOptions,
		Target:            target,
	}, nil
}

// applyFightingStyles modifies attack data and options based on the character's equipped fighting styles and conditions.
func (c *Character) applyFightingStyles(ad *core.AttackData, opts *core.AttackOptions, isRanged bool,
	isVersatile bool, isTwoHanded bool, slot core.WeaponSlot) {
	if c.Class.FightingStyles == nil {
		return
	}

	for _, style := range c.Class.FightingStyles {
		switch style {
		case classes.StyleArchery:
			if isRanged && ad.IsRangedWeapon {
				ad.AttackModifier += 2
			}
		case classes.StyleDueling:
			// This requires that the character is only holding one weapon
			// Check for shield - cannot apply if present
			// Check for versatile - must be used with one hand
			if c.EquipmentManager.GetHasShieldEquipped() || ad.IsVersatileAttack {
				break
			}
			ad.DamageModifier += 2
		case classes.StyleGWF:
			if isTwoHanded || isVersatile {
				opts.RerollOnesAndTwos = true
			}
		case classes.StyleTWF:
			if slot == core.WSSecondary {
				opts.ShouldApplyDamageMod = true
			}
		}
	}
}

// computeAttackAdvantage derives advantage/disadvantage from basic conditions.
// It respects the passed-in adv as a baseline and then applies condition-based modifiers:
// - Attacker blinded or poisoned: disadvantage
// - Target prone: melee attacks advantage, ranged attacks disadvantage
// - Target restrained/paralyzed/unconscious: advantage
// If both advantage and disadvantage are present, result is Normal.
// Deprecated: computeAttackAdvantage has been replaced by core.DetermineAttackAdvantageForEntities

// MakeSavingThrow calculates a saving throw roll using the specified ability and returns the result, rolls, and an error if any.
func (c *Character) MakeSavingThrow(ability core.Ability, targetValue int, isSpell bool, damageType core.DamageType) (core.RollResult, error) {
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
	baseAdv := c.EntityStateManager.GetSavingThrowAdvantage(ability) // This should get the default advantage of the character
	condAdv := core.DetermineSaveAdvantageFromConditions(c.GetConditions(), ability)
	raceAdv := core.RollNormal
	if damageType == core.DamagePoison && c.Race.ID == races.Dwarf {
		raceAdv = core.RollAdvantage
	}
	if !c.Race.SavingThrowAdv.AdvantageOnlyAgainstSpells || isSpell {
		raceAdv = c.Race.SavingThrowAdv.Abilities[ability]
	}
	opts.Advantage = core.GetFinalAdvantageType([]core.AdvantageType{baseAdv, condAdv, raceAdv})
	opts.Modifier = mod
	opts.RollType = core.DiceRollSavingThrow
	opts.TargetValue = targetValue

	res, err := c.RollManager.RollSavingThrow(opts)
	if err != nil {
		return nil, err
	}

	return res, nil
}
