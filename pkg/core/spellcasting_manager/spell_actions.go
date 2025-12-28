package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
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

	return res, nil
}

func (scm *SpellcastingManager) castDamageSpell(req *SpellCastRequest) (*SpellResult, error) {
	switch req.SpellCastData.SpellChoice.Spell.GetHasDC() {
	case true:
		// Has DC So no attack roll needed -> target makes saving throw
		ability := req.SpellCastData.SpellChoice.Spell.GetSpellDC().GetAbility()
		targetDC, err := scm.parent.GetSpellSaveDC(&ability)
		if err != nil {
			return nil, err
		}

		dt := req.SpellCastData.GetSpellChoice().GetFormula().GetDamageType()
		saveRes, err := req.GetTarget().MakeSavingThrow(ability, targetDC, true, dt)
		if err != nil {
			return nil, err
		}

		if saveRes.GetIsSuccess() && req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetOnSuccess() == core.DCOnSuccessNone {
			// Target takes no damage

			spellResult := SpellResult{
				ActorName:        scm.parent.GetName(),
				TargetName:       req.Target.GetName(),
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

			events.LogSpellAttackEvent(scm.parent, &spellResult, scm.parent.GetEventListener())

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

			dmg, err2 := scm.rollManager.RollSpellValue(req, false, rollOpts)
			if err2 != nil {
				return nil, err2
			}
			dmgRollResult = dmg
		}

		spellResult := SpellResult{
			ActorName:        scm.parent.GetName(),
			TargetName:       req.Target.GetName(),
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

		events.LogSpellAttackEvent(scm.parent, &spellResult, scm.parent.GetEventListener())

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

		dmgRollResult, err := scm.rollManager.RollSpellValue(req, attackRollResult.IsCritical, rollOpts)
		if err != nil {
			return nil, err
		}

		attackResult := SpellResult{
			ActorName:        scm.parent.GetName(),
			TargetName:       req.Target.GetName(),
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

		events.LogSpellAttackEvent(scm.parent, &attackResult, scm.parent.GetEventListener())

		return &attackResult, nil
	default:
		return nil, fmt.Errorf("Invalid spell cast data")
	}
}

func (scm *SpellcastingManager) castHealingSpell(req *SpellCastRequest) (*SpellResult, error) {
	res := SpellResult{
		ActorName:        scm.parent.GetName(),
		TargetName:       req.Target.GetName(),
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

	healRollResult, err := scm.rollManager.RollSpellValue(req, false, opts)
	if err != nil {
		return nil, err
	}

	res.SpellTotalValue = healRollResult.Total
	res.ValueRoll = healRollResult

	events.LogSpellHealEvent(scm.parent, &res, scm.parent.GetEventListener())

	return &res, nil
}

func (scm *SpellcastingManager) IsSaveSuccessful(roll int, dc int) bool {
	return roll >= dc
}
