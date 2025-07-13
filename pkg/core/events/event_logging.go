package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
)

func LogCharacterActionChoiceEvent(actor core.Entity, choice shared.ActionType, listener func(event interface{})) {
	event := &ActionChoiceEvent{
		ActionChoice: choice,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogMeleeAttackEvent(actor core.Entity, target core.Entity, attackResult core.AttackResultData, listener func(event interface{})) {
	event := &MeleeAttackEvent{
		Target:         target.GetName(),
		AttackName:     attackResult.GetActorName(),
		AttackCount:    attackResult.GetAttackCount(),
		AttackRoll:     attackResult.GetAttackRoll(),
		AttackModifier: attackResult.GetAttackTotal() - attackResult.GetAttackRoll(),
		AttackTotal:    attackResult.GetAttackTotal(),
		Success:        attackResult.GetIsHit(),
		CriticalHit:    attackResult.GetIsCriticalHit(),
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSpellChoiceEvent(actor core.Entity, spellChoiceEvent SpellChoiceEvent, listener func(event interface{})) {
	event := spellChoiceEvent
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSpellAttackEvent(actor core.Entity, target core.Entity, res core.SpellResultData, listener func(event interface{})) {
	event := &SpellAttackEvent{
		Target:         target.GetName(),
		SpellName:      res.GetSpellName(),
		AttackTotal:    res.GetAttackTotal(),
		AttackModifier: res.GetAttackTotal() - res.GetAttackRoll(),
		AttackRoll:     res.GetAttackRoll(),
		Success:        res.GetHasDC(),
		CriticalHit:    res.GetIsCriticalHit(),
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

func LogHealEvent(actor core.Entity, target core.Entity, amt int, rolls []int, listener func(event interface{})) {
	event := &HealEvent{
		Target:      target.GetName(),
		Amount:      amt,
		Rolls:       rolls,
		IsMaxHealth: target.GetMaxHP() == target.GetCurrentHP(),
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogHPModifiedEvent(actor core.Entity, amt int, prevHP int, newHP int, listener func(event interface{})) {
	event := &HPModifiedEvent{
		Amount:     amt,
		PreviousHP: prevHP,
		CurrentHP:  newHP,
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

func LogDiceRollEvent(actor core.Entity, res core.DiceRollResultData, listener func(event interface{})) {
	event := &DiceRollEvent{
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
