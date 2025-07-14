package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"fmt"
)

func (scm *SpellcastingManager) CastSpell(req *SpellCastRequest, options SpellOptions) (*SpellResult, error) {
	switch req.SpellCastData.SpellChoice.Spell.GetSpellType() {
	case core.STDamage:
		// Damage
	case core.STHealing:
		// Healing

	}
}

func (scm *SpellcastingManager) castDamageSpell(req *SpellCastRequest, options SpellOptions) (*SpellResult, error) {
	switch req.SpellCastData.SpellChoice.Spell.GetHasDC() {
	case true:
		// Has DC So no attack roll needed -> target makes saving throw
		ability := req.SpellCastData.SpellChoice.Spell.GetSpellDC().GetAbility()
		targetDC := scm.parent.GetSpellSaveDC(ability)

		saveRes, err := req.GetTarget().MakeSavingThrow(ability, targetDC)
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
				IsHit:            false,
				IsCriticalHit:    false,
				HasDC:            req.GetSpellCastData().GetSpellChoice().GetSpell().GetHasDC(),
				TargetDCValue:    targetDC,
				SpellSaveAbility: req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetAbility(),
				SpellSaveEffect:  req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetOnSuccess(),
				SpellSaveRolls:   saveRes.GetFinalRolls(),
				SpellSaveTotal:   saveRes.GetTotal(),
				SpellSaveSuccess: saveRes.GetIsSuccess(),
				DamageRoll:       nil,
				DamageType:       req.GetSpellCastData().GetSpellChoice().GetFormula().GetDamageType(),
			}

			events.LogSpellAttackEvent(scm.parent, &spellResult, scm.parent.GetEventListener())

			return &spellResult, nil
		}

		// Damage Roll
		rollOpts := roll_manager.RollOptions{
			Advantage:         core.RollNormal,
			Modifier:          0,
			CriticalThreshold: 0,                       // Not relevant to damage function
			TreatOnesAsTwos:   options.TreatOnesAsTwos, // TODO: Does this make sense to pull from options
			RollType:          core.DiceRollDamage,
			RollContext:       "Damage Roll",
			TargetValue:       0, // Not relevant
		}

		dmgRollResult, err := scm.rollManager.RollSpellDamage(req, false, rollOpts)
		if err != nil {
			return nil, err
		}

		spellResult := SpellResult{
			ActorName:        scm.parent.GetName(),
			TargetName:       req.Target.GetName(),
			SpellName:        req.GetSpellCastData().GetSpellChoice().GetSpell().GetName(),
			SpellLevel:       req.GetSpellCastData().GetSpellChoice().GetSpell().GetLevel(),
			SpellTotalValue:  dmgRollResult.Total,
			AttackRoll:       0,
			AttackTotal:      0,
			IsHit:            true,
			IsCriticalHit:    false,
			HasDC:            req.GetSpellCastData().GetSpellChoice().GetSpell().GetHasDC(),
			TargetDCValue:    targetDC,
			SpellSaveAbility: req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetAbility(),
			SpellSaveEffect:  req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetOnSuccess(),
			SpellSaveRolls:   saveRes.GetFinalRolls(),
			SpellSaveTotal:   saveRes.GetTotal(),
			SpellSaveSuccess: saveRes.GetIsSuccess(),
			DamageRoll:       dmgRollResult,
			DamageType:       req.GetSpellCastData().GetSpellChoice().GetFormula().GetDamageType(),
		}

		if saveRes.GetIsSuccess() && req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetOnSuccess() == core.DCOnSuccessHalf {
			// Target Takes half-damage
			dmgRollResult.Total = dmgRollResult.Total / 2
			spellResult.DamageRoll = dmgRollResult
			spellResult.SpellTotalValue = dmgRollResult.Total
		}

		events.LogSpellAttackEvent(scm.parent, &spellResult, scm.parent.GetEventListener())

		return &spellResult, nil
	case false:

		// Attack Roll
		attackMod := req.SpellCastData.AttackModifier + req.SpellOptions.BonusToAttackRoll
		cT := 20
		if options.ImprovedCritical {
			cT = 19
		}

		rollOpts := roll_manager.RollOptions{
			Advantage:         options.Advantage,
			Modifier:          attackMod,
			CriticalThreshold: cT,
			TreatOnesAsTwos:   options.TreatOnesAsTwos, // Not relevant to the attack roll
			RollType:          core.DiceRollAttack,
			RollContext:       "Attack Roll",
			TargetValue:       req.Target.GetAC(),
		}

		attackRollResult, err := scm.rollManager.RollD20(rollOpts)
		if err != nil {
			return nil, err
		}

		// Damage Roll
		rollOpts = roll_manager.RollOptions{
			Advantage:         core.RollNormal,
			Modifier:          0,
			CriticalThreshold: 0,                       // Not relevant to damage function
			TreatOnesAsTwos:   options.TreatOnesAsTwos, // TODO: Does this make sense to pull from options
			RollType:          core.DiceRollDamage,
			RollContext:       "Damage Roll",
			TargetValue:       0, // Not relevant
		}

		dmgRollResult, err := scm.rollManager.RollSpellDamage(req, attackRollResult, rollOpts)
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
			IsHit:            attackRollResult.IsSuccess,
			IsCriticalHit:    attackRollResult.IsCritical,
			HasDC:            req.GetSpellCastData().GetSpellChoice().GetSpell().GetHasDC(),
			SpellSaveAbility: req.GetSpellCastData().GetSpellChoice().GetSpell().GetSpellDC().GetAbility(),
			SpellSaveRolls:   nil,
			SpellSaveTotal:   0,
			SpellSaveSuccess: false,
			DamageRoll:       dmgRollResult,
			DamageType:       req.GetSpellCastData().GetSpellChoice().GetFormula().GetDamageType(),
		}

		events.LogSpellAttackEvent(scm.parent, &attackResult, scm.parent.GetEventListener())

		return &attackResult, nil
	}

	return &res, nil
}

func (scm *SpellcastingManager) castHealingSpell(target core.Entity, spell *core.SpellChoice) error {
	var result SpellResult

}

func (scm *SpellcastingManager) IsSaveSuccessful(roll int, dc int) bool {
	return roll >= dc
}
