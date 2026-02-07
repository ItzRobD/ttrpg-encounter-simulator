package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/events"
	"dnd5e-encounter-simulator-backend/pkg_old/core/roll_manager"
	"fmt"
)

func (scm *SpellcastingManager) CastSpell(req *SpellCastRequest) (*SpellResult, error) {
	var res *SpellResult
	var err error
	switch req.SpellCastData.SpellChoice.Spell.GetSpellType() {
	case core.STDamage:
		res, err = scm.castDamageSpell(req)
		if err != nil {
			return nil, err
		}
	case core.STHealing:
		res, err = scm.castHealingSpell(req)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid spell cast data")
	}

	// Apply concentration to entity state manager if it's a concentration spell
	if req.SpellCastData.SpellChoice.Spell.GetIsConcentration() {
		scm.parent.SetConcentrating(true, req.SpellCastData.SpellChoice.Spell.GetName())
	}

	res.IsAOE = req.SpellCastData.SpellChoice.Spell.GetIsAOE()

	return res, nil
}

func (scm *SpellcastingManager) castDamageSpell(req *SpellCastRequest) (*SpellResult, error) {
	if req.SpellCastData.SpellChoice.Spell.GetIsAutoHit() {
		// Auto-hit spell (like Magic Missile)
		rollOpts := roll_manager.NewRollOptions()
		rollOpts.TreatOnesAsTwos = req.SpellOptions.TreatOnesAsTwos
		rollOpts.RollType = core.DiceRollDamage

		dmgRollResult, err := scm.rollManager.RollSpellValue(req, false, rollOpts, false)
		if err != nil {
			return nil, err
		}

		spellResult := SpellResult{
			ActorName:       core.FormatEntityName(scm.parent),
			TargetName:      core.FormatEntityName(req.Target),
			Target:          req.Target,
			SpellName:       req.GetSpellCastData().GetSpellChoice().GetSpell().GetName(),
			SpellLevel:      req.GetSpellCastData().GetSpellChoice().GetSpell().GetLevel(),
			SpellTotalValue: dmgRollResult.Total,
			AttackRoll:      0,
			AttackTotal:     0,
			IsSuccess:       true,
			IsCriticalHit:   false,
			HasDC:           false,
			ValueRoll:       dmgRollResult,
			DamageType:      req.GetSpellCastData().GetSpellChoice().GetFormula().GetDamageType(),
			IsConcentration: req.GetSpellCastData().GetSpellChoice().GetSpell().GetIsConcentration(),
		}

		// 1. Log Spell Cast/Attack (generates currentID)
		scm.parent.LogEvent(events.ETSpellAttackEvent, &spellResult)

		// 2. Advance Scope: Damage/Effects should be children of the spell cast
		ctx := scm.parent.GetCurrentEventContext()
		if ctx != nil {
			actionID := ctx.GetParentID()
			ctx.AdvanceScope()

			// (No explicit damage log here as it's already in spellResult or would be handled by ModifyHP)

			// 3. Restore Action ID
			ctx.SetParentID(actionID)
		}

		return &spellResult, nil
	}

	switch req.SpellCastData.SpellChoice.Spell.GetHasDC() {
	case true:
		// Has DC So no attack roll needed -> target makes saving throw
		ability := req.SpellCastData.SpellChoice.Spell.GetSpellDC().GetAbility()
		targetDC, err := scm.parent.GetSpellSaveDC(&ability)
		if err != nil {
			return nil, err
		}

		dt := req.SpellCastData.GetSpellChoice().GetFormula().GetDamageType()
		saveRes, err := req.GetTarget().MakeSavingThrow(ability, targetDC, true, dt, req.GetSimulationOptions())
		if err != nil {
			return nil, err
		}

		if saveRes.GetIsSuccess() && req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetOnSuccess() == core.DCOnSuccessNone {
			// Target takes no damage

			spellResult := SpellResult{
				ActorName:        core.FormatEntityName(scm.parent),
				TargetName:       core.FormatEntityName(req.Target),
				Target:           req.Target,
				SpellName:        req.GetSpellCastData().GetSpellChoice().GetSpell().GetName(),
				SpellLevel:       req.GetSpellCastData().GetSpellChoice().GetSpell().GetLevel(),
				SpellTotalValue:  0,
				AttackRoll:       0,
				AttackTotal:      0,
				IsSuccess:        false,
				IsCriticalHit:    false,
				HasDC:            req.GetSpellCastData().GetSpellChoice().GetSpell().GetHasDC(),
				TargetDCValue:    targetDC,
				SpellSaveAbility: req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetAbility(),
				SpellSaveEffect:  req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetOnSuccess(),
				SpellSaveRolls:   saveRes.GetFinalRolls(),
				SpellSaveTotal:   saveRes.GetTotal(),
				SpellSaveSuccess: saveRes.GetIsSuccess(),
				ValueRoll:        nil,
				DamageType:       req.GetSpellCastData().GetSpellChoice().GetFormula().GetDamageType(),
				IsConcentration:  req.GetSpellCastData().GetSpellChoice().GetSpell().GetIsConcentration(),
			}

			// 1. Log Spell Cast (generates currentID)
			scm.parent.LogEvent(events.ETSpellAttackEvent, &spellResult)

			// 2. Advance Scope: Effects (if any) should be children
			ctx := scm.parent.GetCurrentEventContext()
			if ctx != nil {
				actionID := ctx.GetParentID()
				ctx.AdvanceScope()

				// ... subsequent logic could log more events ...

				// 3. Restore Action ID
				ctx.SetParentID(actionID)
			}

			return &spellResult, nil
		}

		// Damage handling
		// Special-case: Guardian of Faith — fixed 20 radiant damage, half on successful save, no roll.
		var dmgRollResult *roll_manager.RollResult
		if req.SpellCastData.SpellChoice.Spell.GetName() == "Guardian of Faith" {
			tmp := roll_manager.RollResult{
				DiceRollType:   core.DiceRollDamage,
				NumberOfDice:   0,
				Die:            core.D0,
				FinalRollValue: 0,
				FinalRolls:     []int{},
				Modifier:       0,
				Total:          20,
				Advantage:      core.RollNormal,
			}
			dmgRollResult = &tmp
		} else {
			// Standard damage roll
			rollOpts := roll_manager.NewRollOptions()
			rollOpts.TreatOnesAsTwos = req.SpellOptions.TreatOnesAsTwos
			rollOpts.RollType = core.DiceRollDamage

			dmg, err2 := scm.rollManager.RollSpellValue(req, false, rollOpts, false)
			if err2 != nil {
				return nil, err2
			}
			dmgRollResult = dmg
		}

		spellResult := SpellResult{
			ActorName:        core.FormatEntityName(scm.parent),
			TargetName:       core.FormatEntityName(req.Target),
			Target:           req.Target,
			SpellName:        req.GetSpellCastData().GetSpellChoice().GetSpell().GetName(),
			SpellLevel:       req.GetSpellCastData().GetSpellChoice().GetSpell().GetLevel(),
			SpellTotalValue:  dmgRollResult.Total,
			AttackRoll:       0,
			AttackTotal:      0,
			IsSuccess:        true,
			IsCriticalHit:    false,
			HasDC:            req.GetSpellCastData().GetSpellChoice().GetSpell().GetHasDC(),
			TargetDCValue:    targetDC,
			SpellSaveAbility: req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetAbility(),
			SpellSaveEffect:  req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetOnSuccess(),
			SpellSaveRolls:   saveRes.GetFinalRolls(),
			SpellSaveTotal:   saveRes.GetTotal(),
			SpellSaveSuccess: saveRes.GetIsSuccess(),
			ValueRoll:        dmgRollResult,
			DamageType:       req.GetSpellCastData().GetSpellChoice().GetFormula().GetDamageType(),
			IsConcentration:  req.GetSpellCastData().GetSpellChoice().GetSpell().GetIsConcentration(),
		}

		if saveRes.GetIsSuccess() && req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetOnSuccess() == core.DCOnSuccessHalf {
			// Target Takes half-damage
			dmgRollResult.Total = dmgRollResult.Total / 2
			spellResult.ValueRoll = dmgRollResult
			spellResult.SpellTotalValue = dmgRollResult.Total
		}

		// 1. Log Spell Cast (generates currentID)
		scm.parent.LogEvent(events.ETSpellAttackEvent, &spellResult)

		// 2. Advance Scope
		ctx := scm.parent.GetCurrentEventContext()
		if ctx != nil {
			actionID := ctx.GetParentID()
			ctx.AdvanceScope()

			// 2b. Log Damage Roll manually (it will use the Spell as parent)
			scm.parent.LogEvent(events.ETRollEvent, &events.DiceRollData{
				RollResult: dmgRollResult,
				DamageType: spellResult.DamageType.String(),
			})

			// 3. Restore Action ID
			ctx.SetParentID(actionID)
		}

		return &spellResult, nil
	case false:
		// Attack Roll
		attackMod := req.SpellCastData.AttackModifier + req.SpellOptions.BonusToAttackRoll
		cT := 20
		if req.SpellOptions.ImprovedCritical {
			cT = 19
		}

		rollOpts := roll_manager.NewRollOptions()
		rollOpts.Advantage = req.SpellOptions.Advantage
		rollOpts.Modifier = attackMod
		rollOpts.CriticalThreshold = cT
		rollOpts.TreatOnesAsTwos = req.SpellOptions.TreatOnesAsTwos
		rollOpts.RollType = core.DiceRollAttack
		rollOpts.TargetValue = req.Target.GetAC()

		attackRollResult, err := scm.rollManager.RollD20(rollOpts, false)
		if err != nil {
			return nil, err
		}

		// Damage Roll
		rollOpts = roll_manager.NewRollOptions()
		rollOpts.TreatOnesAsTwos = req.SpellOptions.TreatOnesAsTwos
		rollOpts.RollType = core.DiceRollDamage

		dmgRollResult, err := scm.rollManager.RollSpellValue(req, attackRollResult.IsCritical, rollOpts, false)
		if err != nil {
			return nil, err
		}

		attackResult := SpellResult{
			ActorName:        core.FormatEntityName(scm.parent),
			TargetName:       core.FormatEntityName(req.Target),
			Target:           req.Target,
			SpellName:        req.GetSpellCastData().GetSpellChoice().GetSpell().GetName(),
			SpellLevel:       req.GetSpellCastData().GetSpellChoice().GetSpell().GetLevel(),
			SpellTotalValue:  dmgRollResult.Total,
			AttackRoll:       attackRollResult.FinalRollValue,
			AttackTotal:      attackRollResult.Total,
			IsSuccess:        attackRollResult.IsSuccess,
			IsCriticalHit:    attackRollResult.IsCritical,
			HasDC:            req.GetSpellCastData().GetSpellChoice().GetSpell().GetHasDC(),
			SpellSaveAbility: req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetAbility(),
			SpellSaveRolls:   nil,
			SpellSaveTotal:   0,
			SpellSaveSuccess: false,
			ValueRoll:        dmgRollResult,
			DamageType:       req.GetSpellCastData().GetSpellChoice().GetFormula().GetDamageType(),
			IsConcentration:  req.GetSpellCastData().GetSpellChoice().GetSpell().GetIsConcentration(),
		}

		// 1. Log Spell Attack (generates currentID)
		scm.parent.LogEvent(events.ETSpellAttackEvent, &attackResult)

		// 2. Advance Scope
		ctx := scm.parent.GetCurrentEventContext()
		if ctx != nil {
			actionID := ctx.GetParentID()
			ctx.AdvanceScope()

			// 2b. Log Damage Roll manually (it will use the Spell as parent)
			scm.parent.LogEvent(events.ETRollEvent, &events.DiceRollData{
				RollResult: dmgRollResult,
				DamageType: attackResult.DamageType.String(),
			})

			// 3. Restore Action ID
			ctx.SetParentID(actionID)
		}

		return &attackResult, nil
	default:
		return nil, fmt.Errorf("Invalid spell cast data")
	}
}

func (scm *SpellcastingManager) castHealingSpell(req *SpellCastRequest) (*SpellResult, error) {
	res := SpellResult{
		ActorName:        core.FormatEntityName(scm.parent),
		TargetName:       core.FormatEntityName(req.Target),
		Target:           req.Target,
		SpellName:        req.GetSpellCastData().GetSpellChoice().GetSpell().GetName(),
		SpellLevel:       req.GetSpellCastData().GetSpellChoice().GetSpell().GetLevel(),
		SpellTotalValue:  0,
		AttackRoll:       0,
		AttackTotal:      0,
		IsSuccess:        true,
		IsCriticalHit:    false,
		HasDC:            false,
		TargetDCValue:    0,
		SpellSaveAbility: "",
		SpellSaveEffect:  "",
		SpellSaveRolls:   nil,
		SpellSaveTotal:   0,
		SpellSaveSuccess: false,
		ValueRoll:        nil,
		DamageType:       "",
		IsConcentration:  req.GetSpellCastData().GetSpellChoice().GetSpell().GetIsConcentration(),
	}

	opts := roll_manager.NewRollOptions()
	opts.RollType = core.DiceRollHealing

	healRollResult, err := scm.rollManager.RollSpellValue(req, false, opts, false)
	if err != nil {
		return nil, err
	}

	res.SpellTotalValue = healRollResult.Total
	res.ValueRoll = healRollResult

	scm.parent.LogEvent(events.ETHealEvent, &res)

	return &res, nil
}

func (scm *SpellcastingManager) IsSaveSuccessful(roll int, dc int) bool {
	return roll >= dc
}
