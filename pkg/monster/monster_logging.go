package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
)

func (m *Monster) LogEvent(eventType interface{}, data interface{}) {
	ctx := m.GetCurrentEventContext()
	listener := m.GetEventListener()
	if listener == nil {
		return
	}

	switch t := eventType.(type) {
	case events.EventType:
		switch t {
		case events.ETRollEvent, events.ETRollInitiative:
			if res, ok := data.(*roll_manager.RollResult); ok {
				events.LogDiceRollEvent(ctx, m, res, listener)
			} else if d, ok := data.(*events.DiceRollData); ok {
				events.LogDiceRollEventWithType(ctx, m, d.RollResult, d.DamageType, listener)
			}
		case events.ETAttackEvent:
			if res, ok := data.(*core.AttackResult); ok {
				events.LogMartialAttackEvent(ctx, m, res, listener)
			}
		case events.ETSpellAttackEvent:
			if res, ok := data.(*core.SpellResult); ok {
				events.LogSpellAttackEvent(ctx, m, *res, listener)
			} else if res, ok := data.(core.SpellResult); ok {
				events.LogSpellAttackEvent(ctx, m, res, listener)
			}
		case events.ETSpellDCEvent:
			if d, ok := data.(*events.SpellDCData); ok {
				events.LogSpellDCEvent(ctx, m, d.Target, d.Spell, d.DC, d.Save, d.IsHit, listener)
			}
		case events.ETHealEvent:
			if res, ok := data.(*core.SpellResult); ok {
				events.LogSpellHealEvent(ctx, m, *res, listener)
			}
		case events.ETDamageEvent:
			if d, ok := data.(*events.DamageData); ok {
				events.LogDamageEvent(ctx, m, d.Target, d.DamageType, d.Damage, d.Rolls, listener)
			}
		case events.ETDeathEvent:
			events.LogDeathEvent(ctx, m, listener)
		case events.ETUnconsciousEvent:
			events.LogUnconsciousEvent(ctx, m, listener)
		case events.ETHPRollEvent:
			if d, ok := data.(*events.HPRollData); ok {
				events.LogHPRollEvent(ctx, m, d.RollSum, d.Rolls, d.ToAdd, listener)
			}
		case events.ETActionChoiceEvent:
			if d, ok := data.(*events.ActionChoiceData); ok {
				events.LogMonsterActionChoiceEvent(ctx, m, d.Choice, d.AllScores, d.TopReasons, d.UtilityScore, listener)
			}
		case events.ETSpellChoiceEvent:
			if d, ok := data.(*events.SpellChoiceData); ok {
				events.LogSpellChoiceEvent(ctx, m, d.Choice, d.Status, d.Target, listener)
				if ctx != nil {
					ctx.AdvanceScope()
				}
			}
		case events.ETHPModifiedEvent:
			if d, ok := data.(*events.HPModifiedData); ok {
				events.LogHPModifiedEvent(ctx, m, d.Subject, d.Res, d.DamageType, d.SourceRollID, listener)
			}
		case events.ETSavingThrowEvent:
			if d, ok := data.(*events.SavingThrowData); ok {
				events.LogSavingThrowEvent(ctx, m, d.Result, d.Roll, d.Modifier, d.Success, listener)
			}
		case events.ETTargetChoiceEvent:
			if d, ok := data.(*events.TargetChoiceData); ok {
				events.LogTargetChoiceEvent(ctx, m, d.Target, d.Score, d.Factors, listener)
			}
		case events.ECombatEventMessage:
			if msg, ok := data.(string); ok {
				events.LogCombatEventMessage(ctx, m, msg, listener)
			}
		case events.ETDamageModifiedEvent:
			if d, ok := data.(*events.DamageModifiedData); ok {
				events.LogDamageModifiedEvent(ctx, m, d.Subject, d.Res, d.DamageType, d.SourceRollID, listener)
			}
		case events.ETDragonbornBreathWeaponEvent:
			if d, ok := data.(*events.DragonbornBreathWeaponData); ok {
				events.LogDragonbornBreathWeaponEvent(ctx, m, d.Target, d.DamageTotal, d.DamageType, d.DC, d.SaveAbility, d.SaveSuccess, d.SaveResult, listener)
				if ctx != nil {
					ctx.AdvanceScope()
				}
			}
		case events.ETSpecialAbilityEvent:
			if d, ok := data.(*events.SpecialAbilityData); ok {
				events.LogSpecialAbilityEvent(ctx, m, d.AbilityName, d.Description, d.TargetName, d.Value, listener)
			}
		case events.ETEquipmentEvent:
			if d, ok := data.(*events.EquipmentEvent); ok {
				events.LogEquipmentEvent(ctx, m, d, listener)
			}
		}
	}
}
