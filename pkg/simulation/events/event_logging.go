package events

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
)

func LogCharacterActionChoiceEvent(actor shared.Entity, choice shared.ActionType, listener func(event CombatEvent)) {
	if listener != nil {
		event := CombatEvent{
			EventType:    ETActionChoiceEvent,
			Actor:        actor.GetName(),
			ActionChoice: choice,
		}
		listener(event)
	}
}

func LogWeaponAttackEvent(actor shared.Entity, target shared.Entity, weapon *weapon.Weapon, attackRoll, attackModifier int, isHit bool, listener func(event CombatEvent)) {
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

func LogSpellChoiceEvent(actor shared.Entity, spell *spells.Spell, hasSlots bool, listener func(event CombatEvent)) {
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

func LogSpellSlotsEvent(actor shared.Entity, spellSlots shared.SpellSlots, listener func(event CombatEvent)) {
	if listener != nil {
		event := CombatEvent{
			EventType:  ETSpellSlotsEvent,
			Actor:      actor.GetName(),
			SpellSlots: spellSlots,
		}
		listener(event)
	}
}

func LogSpellAttackEvent(actor shared.Entity, target shared.Entity, spell *spells.Spell, attackRoll, attackModifier int, isHit bool, listener func(event CombatEvent)) {
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

func LogSpellDCEvent(actor shared.Entity, target shared.Entity, spell *spells.Spell, dc int, save int, isHit bool, listener func(event CombatEvent)) {
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

func LogDamageEvent(actor shared.Entity, target shared.Entity, damageType string, damage int, rolls []int, listener func(event CombatEvent)) {
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

func LogHealEvent(actor shared.Entity, target shared.Entity, amt int, rolls []int, listener func(event CombatEvent)) {
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

func LogHPModifiedEvent(actor shared.Entity, amt int, prevHP int, newHP int, listener func(event CombatEvent)) {
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

func LogHPRollEvent(actor shared.Entity, rollSum int, rolls []int, toAdd int, listener func(event CombatEvent)) {
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

func LogDiceRollEvent(actor shared.Entity, rollSum int, rolls []int, listener func(event CombatEvent)) {
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

func LogSavingThrowEvent(actor shared.Entity, result int, rolls []int, modifier int, success bool, listener func(event CombatEvent)) {
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
