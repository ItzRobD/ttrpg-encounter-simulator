package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"time"
)

func LogCharacterActionChoiceEvent(actor core.Entity, choice core.ActionType, listener func(event interface{})) {
	event := &ActionChoiceEvent{
		ActionChoice: choice,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogMonsterActionChoiceEvent(actor core.Entity, choice core.ActionType, listener func(event interface{})) {
	event := &ActionChoiceEvent{
		ActionChoice: choice,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogMeleeAttackEvent(actor core.Entity, attackResult *core.AttackResult, listener func(event interface{})) {
	event := &MeleeAttackEvent{
		Target:         attackResult.GetTargetName(),
		AttackName:     attackResult.GetAttackName(),
		AttackCount:    attackResult.GetAttackCount(),
		AttackRoll:     attackResult.GetAttackRoll(),
		AttackModifier: attackResult.GetAttackTotal() - attackResult.GetAttackRoll(),
		AttackTotal:    attackResult.GetAttackTotal(),
		TargetValue:    attackResult.GetTargetValue(),
		Success:        attackResult.GetIsHit(),
		CriticalHit:    attackResult.GetIsCriticalHit(),
		DamageTotal:    attackResult.GetDamageResult().GetTotal(),
		DamageType:     attackResult.GetDamageType().String(),
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogDragonbornBreathWeaponEvent(actor core.Entity, target core.Entity, damageTotal int, damageType string, dc int, saveAbility string, saveSuccess bool, saveResult int, listener func(event interface{})) {
	event := &DragonbornBreathWeaponEvent{
		Target:             target.GetName(),
		DamageTotal:        damageTotal,
		DamageType:         damageType,
		DC:                 dc,
		SaveAbility:        saveAbility,
		SavingThrowSuccess: saveSuccess,
		SavingThrowResult:  saveResult,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

//func LogMonsterAttackEvent(actor core.Entity, attackResult *core.AttackResult, listener func(event interface{})) {
//	event := &
//}

func LogDamageEvent(actor core.Entity, target core.Entity, damageType string, damage int, rolls []int, listener func(event interface{})) {
	event := &DamageEvent{
		Target:     target.GetName(),
		Amount:     damage,
		DamageType: damageType,
		Rolls:      rolls,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSpellChoiceEvent(actor core.Entity, choice *core.SpellChoice, status *spells.SpellcastingManagerStatus, listener func(event interface{})) {
	event := &SpellChoiceEvent{
		SpellChoice:   choice,
		ManagerStatus: status,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSpellAttackEvent(actor core.Entity, res core.SpellResult, listener func(event interface{})) {
	var damageTotal int
	if res.GetValueResult() != nil {
		damageTotal = res.GetValueResult().GetTotal()
	} else {
		damageTotal = 0
	}

	event := &SpellAttackEvent{
		Target:             res.GetTargetName(),
		SpellName:          res.GetSpellName(),
		SpellLevel:         res.GetSpellLevel(),
		AttackTotal:        res.GetAttackTotal(),
		AttackModifier:     res.GetAttackTotal() - res.GetAttackRoll(),
		AttackRoll:         res.GetAttackRoll(),
		Success:            res.GetIsHit(),
		CriticalHit:        res.GetIsCriticalHit(),
		DamageTotal:        damageTotal,
		DamageType:         res.GetDamageType().String(),
		HasDC:              res.GetHasDC(),
		DCAbility:          res.GetSpellSaveAbility().String(),
		SaveEffect:         res.GetSpellSaveEffect().String(),
		DCValue:            res.GetTargetDCValue(),
		SavingThrowSuccess: res.GetSpellSaveSuccess(),
		SavingThrowResult:  res.GetSpellSaveTotal(),
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSpellDCEvent(actor core.Entity, target core.Entity, spell *spells.Spell, dc int, save int, isHit bool, listener func(event interface{})) {
	event := &SpellDCEvent{
		Target:      target.GetName(),
		SpellChoice: spell,
		DC:          dc,
		SavingThrow: save,
		Success:     isHit,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSpellHealEvent(actor core.Entity, res core.SpellResult, listener func(event interface{})) {
	event := &HealEvent{
		Target:     res.GetTargetName(),
		Name:       res.GetSpellName(),
		SpellLevel: res.GetSpellLevel(),
		IsSpell:    true,
		HealTotal:  res.GetValueResult().GetTotal(),
		HealRolls:  res.GetValueResult().GetFinalRolls(),
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogLayOnHandsHealEvent(actor core.Entity, subject core.Entity, value int, listener func(event interface{})) {
	event := &HealEvent{
		Target:     subject.GetName(),
		Name:       "Lay on Hands",
		SpellLevel: 0,
		IsSpell:    false,
		HealTotal:  value,
		HealRolls:  []int{value},
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogDamageModifiedEvent(actor core.Entity, subject core.Entity, res core.DamageModificationResult, listener func(event interface{})) {
	event := &DamageModifiedEvent{
		BaseEvent:        BaseEvent{},
		SubjectName:      subject.GetName(),
		OriginalValue:    res.OriginalValue,
		FinalValue:       res.FinalValue,
		WasModified:      res.WasModified,
		ResistanceType:   res.ResistanceType,
		ResistanceBroken: res.ResistanceBroken,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogHPModifiedEvent(actor core.Entity, subject core.Entity, res core.HPModificationResult, listener func(event interface{})) {
	event := &HPModifiedEvent{
		BaseEvent:         BaseEvent{},
		SubjectName:       subject.GetName(),
		ModificationValue: res.GetModificationValue(),
		OriginalHP:        res.GetOriginalHP(),
		OriginalTempHP:    res.GetOriginalTempHP(),
		NewHP:             res.GetNewHP(),
		NewTempHP:         res.GetNewTempHP(),
		DidHealHP:         res.GetDidHealHP(),
		DidHealTempHP:     res.GetDidHealTempHP(),
		DidTempDamage:     res.GetDidTempDamage(),
		DidHPDamage:       res.GetDidHPDamage(),
		IsUnconscious:     res.GetIsUnconscious(),
		IsMaxHealth:       res.GetIsMaxHealth(),
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogHPRollEvent(actor core.Entity, rollSum int, rolls []int, toAdd int, listener func(event interface{})) {
	event := &HPRollEvent{
		Value:    rollSum,
		Rolls:    rolls,
		Modifier: toAdd,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogDiceRollEvent(actor core.Entity, res core.RollResult, listener func(event interface{})) {
	event := &DiceRollEvent{
		RollType:       res.GetDiceRollType(),
		NumberOfDice:   res.GetNumberOfDice(),
		Die:            res.GetDiceType(),
		FinalRollValue: res.GetFinalRollValue(),
		FinalRolls:     res.GetFinalRolls(),
		Modifier:       res.GetModifier(),
		Total:          res.GetTotal(),
		Advantage:      res.GetAdvantage(),
		OriginalRolls:  res.GetOriginalRolls(),
		RerollEvents:   res.GetRerollEvents(),
		WasRerolled:    res.GetWasRerolled(),
		IsCritical:     res.GetIsCritical(),
		IsNaturalOne:   res.GetIsNaturalOne(),
		IsSuccess:      res.GetIsSuccess(),
		TargetValue:    res.GetTargetValue(),
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSavingThrowEvent(actor core.Entity, result int, roll int, modifier int, success bool, listener func(event interface{})) {
	event := &SavingThrowEvent{
		Actor:    actor.GetName(),
		Result:   result,
		Roll:     roll,
		Modifier: modifier,
		Success:  success,
	}

	if listener != nil {
		listener(event)
	}
}

func LogTargetChoiceEvent(actor core.Entity, target core.Entity, listener func(event interface{})) {
	event := &TargetChoiceEvent{
		Target: target.GetName(),
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSpecialAbilityEvent(actor core.Entity, abilityName string, description string, targetName string, value int, listener func(event interface{})) {
	event := &SpecialAbilityEvent{
		AbilityName: abilityName,
		Description: description,
		Target:      targetName,
		Value:       value,
	}
	event.SetActor(actor.GetName())
	event.SetTimestamp(time.Now())

	if listener != nil {
		listener(event)
	}
}

func LogCombatEventMessage(actor core.Entity, message string, listener func(event interface{})) {
	event := &CombatEventMessage{
		Message: message,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}
