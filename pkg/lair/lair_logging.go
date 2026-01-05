package lair

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
)

func (l *Lair) LogEvent(eventType interface{}, data interface{}) {
	ctx := l.GetCurrentEventContext()
	listener := l.GetEventListener()
	if listener == nil {
		return
	}

	switch t := eventType.(type) {
	case events.EventType:
		switch t {
		case events.ETRollEvent:
			if res, ok := data.(*roll_manager.RollResult); ok {
				events.LogDiceRollEvent(ctx, l, *res, listener)
			}
		case events.ETAttackEvent:
			if res, ok := data.(*core.AttackResult); ok {
				events.LogMeleeAttackEvent(ctx, l, res, listener)
			}
		case events.ETSpellAttackEvent:
			if res, ok := data.(*core.SpellResult); ok {
				events.LogSpellAttackEvent(ctx, l, *res, listener)
			} else if res, ok := data.(core.SpellResult); ok {
				events.LogSpellAttackEvent(ctx, l, res, listener)
			}
		case events.ETSpellDCEvent:
			if d, ok := data.(*events.SpellDCData); ok {
				events.LogSpellDCEvent(ctx, l, d.Target, d.Spell, d.DC, d.Save, d.IsHit, listener)
			}
		case events.ETHealEvent:
			if res, ok := data.(*core.SpellResult); ok {
				events.LogSpellHealEvent(ctx, l, *res, listener)
			}
		case events.ETDamageEvent:
			if d, ok := data.(*events.DamageData); ok {
				events.LogDamageEvent(ctx, l, d.Target, d.DamageType, d.Damage, d.Rolls, listener)
			}
		case events.ETDeathEvent:
			events.LogDeathEvent(ctx, l, listener)
		case events.ETUnconsciousEvent:
			events.LogUnconsciousEvent(ctx, l, listener)
		case events.ETHPRollEvent:
			if d, ok := data.(*events.HPRollData); ok {
				events.LogHPRollEvent(ctx, l, d.RollSum, d.Rolls, d.ToAdd, listener)
			}
		case events.ETActionChoiceEvent:
			if d, ok := data.(*events.ActionChoiceData); ok {
				// Lair usually uses MonsterActionChoiceEvent structure if it logs choices
				events.LogMonsterActionChoiceEvent(ctx, l, d.Choice, d.AllScores, d.TopReasons, d.UtilityScore, listener)
			}
		case events.ETSpellChoiceEvent:
			if d, ok := data.(*events.SpellChoiceData); ok {
				events.LogSpellChoiceEvent(ctx, l, d.Choice, d.Status, listener)
			}
		case events.ETHPModifiedEvent:
			if d, ok := data.(*events.HPModifiedData); ok {
				events.LogHPModifiedEvent(ctx, l, d.Subject, d.Res, listener)
			}
		case events.ETSavingThrowEvent:
			if d, ok := data.(*events.SavingThrowData); ok {
				events.LogSavingThrowEvent(ctx, l, d.Result, d.Roll, d.Modifier, d.Success, listener)
			}
		case events.ETTargetChoiceEvent:
			if d, ok := data.(*events.TargetChoiceData); ok {
				events.LogTargetChoiceEvent(ctx, l, d.Target, d.Score, d.Factors, listener)
			}
		case events.ECombatEventMessage:
			if msg, ok := data.(string); ok {
				events.LogCombatEventMessage(ctx, l, msg, listener)
			}
		case events.ETDamageModifiedEvent:
			if d, ok := data.(*events.DamageModifiedData); ok {
				events.LogDamageModifiedEvent(ctx, l, d.Subject, d.Res, listener)
			}
		case events.ETDragonbornBreathWeaponEvent:
			if d, ok := data.(*events.DragonbornBreathWeaponData); ok {
				events.LogDragonbornBreathWeaponEvent(ctx, l, d.Target, d.DamageTotal, d.DamageType, d.DC, d.SaveAbility, d.SaveSuccess, d.SaveResult, listener)
			}
		case events.ETSpecialAbilityEvent:
			if d, ok := data.(*events.SpecialAbilityData); ok {
				events.LogSpecialAbilityEvent(ctx, l, d.AbilityName, d.Description, d.TargetName, d.Value, listener)
			}
		}
	}
}
