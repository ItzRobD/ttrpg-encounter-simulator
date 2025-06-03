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

func LogMeleeAttackEvent(actor core.Entity, target core.Entity, attack string, attackRoll, attackModifier int, isHit bool, isCritical bool, listener func(event interface{})) {
	event := &MeleeAttackEvent{
		Target:         target.GetName(),
		AttackName:     attack,
		AttackRoll:     attackRoll,
		AttackModifier: attackModifier,
		AttackTotal:    attackRoll + attackModifier,
		Success:        isHit,
		CriticalHit:    isCritical,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSpellChoiceEvent(actor core.Entity, spell *spells.Spell, hasSlots bool, listener func(event interface{})) {
	event := &SpellChoiceEvent{
		SpellChoice: spell,
		HasSlots:    hasSlots,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSpellSlotsEvent(actor core.Entity, spellSlots shared.SpellSlots, listener func(event interface{})) {
	event := &SpellSlotsEvent{
		SpellSlots: spellSlots,
	}
	event.SetActor(actor.GetName())

	if listener != nil {
		listener(event)
	}
}

func LogSpellAttackEvent(actor core.Entity, target core.Entity, spell *spells.Spell, attackRoll, attackModifier int, isHit bool, listener func(event interface{})) {
	event := &SpellAttackEvent{
		Target:         target.GetName(),
		SpellChoice:    spell,
		AttackTotal:    attackRoll + attackModifier,
		AttackModifier: attackModifier,
		AttackRoll:     attackRoll,
		Success:        isHit,
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
		Target: target.GetName(),
		Amount: amt,
		Rolls:  rolls,
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

func LogDiceRollEvent(actor core.Entity, rollSum int, rolls []int, rollType shared.DiceRollType, modifier int, listener func(event interface{})) {
	event := &DiceRollEvent{
		RollType: rollType,
		Value:    rollSum,
		Rolls:    rolls,
		Modifier: modifier,
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
