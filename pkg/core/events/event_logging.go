package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
)

func LogCharacterActionChoiceEvent(actor core.Entity, choice shared.ActionType, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType:    ETActionChoiceEvent,
			Actor:        actor.GetName(),
			ActionChoice: choice,
		}
		listener(event)
	}
}

func LogWeaponAttackEvent(actor core.Entity, target core.Entity, weapon *weapon.Weapon, attackRoll, attackModifier int, isHit bool, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType: ETAttackEvent,
			Actor:     actor.GetName(),
			Target:    target.GetName(),
			Attack:    weapon.Name,
			Value:     attackRoll + attackModifier,
			Rolls:     []int{attackRoll},
			Success:   isHit,
		}
		listener(event)
	}
}

func LogSpellChoiceEvent(actor core.Entity, spell *spells.Spell, hasSlots bool, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType:   ETSpellChoiceEvent,
			Actor:       actor.GetName(),
			SpellChoice: spell,
			HasSlots:    hasSlots,
		}
		listener(event)
	}
}

func LogSpellSlotsEvent(actor core.Entity, spellSlots shared.SpellSlots, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType:  ETSpellSlotsEvent,
			Actor:      actor.GetName(),
			SpellSlots: spellSlots,
		}
		listener(event)
	}
}

func LogSpellAttackEvent(actor core.Entity, target core.Entity, spell *spells.Spell, attackRoll, attackModifier int, isHit bool, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType: ETAttackEvent,
			Actor:     actor.GetName(),
			Target:    target.GetName(),
			Attack:    spell.Name,
			Value:     attackRoll + attackModifier,
			Rolls:     []int{attackRoll},
			Success:   isHit,
		}
		listener(event)
	}
}

func LogSpellDCEvent(actor core.Entity, target core.Entity, spell *spells.Spell, dc int, save int, isHit bool, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType:   ETSpellDC,
			Actor:       actor.GetName(),
			Target:      target.GetName(),
			Attack:      spell.Name,
			Value:       dc,
			SavingThrow: save,
			Success:     isHit,
		}
		listener(event)
	}
}

func LogDamageEvent(actor core.Entity, target core.Entity, damageType string, damage int, rolls []int, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType:  ETDamageEvent,
			Actor:      actor.GetName(),
			Target:     target.GetName(),
			Value:      damage,
			DamageType: damageType,
			Rolls:      rolls,
		}
		listener(event)
	}
}

func LogHealEvent(actor core.Entity, target core.Entity, amt int, rolls []int, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType: ETHealEvent,
			Actor:     actor.GetName(),
			Target:    target.GetName(),
			Value:     amt,
			Rolls:     rolls,
		}
		listener(event)
	}
}

func LogHPModifiedEvent(actor core.Entity, amt int, prevHP int, newHP int, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType:  ETHPModifiedEvent,
			Actor:      actor.GetName(),
			Value:      amt,
			PreviousHP: prevHP,
			CurrentHP:  newHP,
		}
		listener(event)
	}
}

func LogHPRollEvent(actor core.Entity, rollSum int, rolls []int, toAdd int, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType: ETHPRollEvent,
			Actor:     actor.GetName(),
			Value:     rollSum,
			Rolls:     rolls,
			Modifier:  toAdd,
		}
		listener(event)
	}
}

func LogDiceRollEvent(actor core.Entity, rollSum int, rolls []int, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType: ETRollEvent,
			Actor:     actor.GetName(),
			Value:     rollSum,
			Rolls:     rolls,
		}
		listener(event)
	}
}

func LogSavingThrowEvent(actor core.Entity, result int, rolls []int, modifier int, success bool, listener func(event interface{})) {
	if listener != nil {
		event := CombatEvent{
			EventType: ETSavingThrowEvent,
			Actor:     actor.GetName(),
			Value:     result,
			Rolls:     rolls,
			Modifier:  modifier,
			Success:   success,
		}
		listener(event)
	}
}
