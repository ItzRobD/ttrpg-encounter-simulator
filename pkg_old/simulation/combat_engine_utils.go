package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg_old/character"
	"dnd5e-encounter-simulator-backend/pkg_old/classes"
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg_old/core/events"
	"dnd5e-encounter-simulator-backend/pkg_old/monster"
	"fmt"
	"math"
)

func (ce *CombatEngine) computeDamageValueAfterResistances(target core.Entity, dt core.DamageType, b []core.ResistBreaker, value int) (core.DamageModificationResult, error) {
	if target == nil {
		return core.DamageModificationResult{}, fmt.Errorf("target is nil")
	}

	targetESM, ok := target.GetState().(*entity_state_manager.EntityStateManager)
	if !ok || targetESM == nil {
		return core.DamageModificationResult{}, fmt.Errorf("target state manager is nil or wrong type")
	}

	targetResistances := targetESM.GetResistances()
	isPetrified := targetESM.GetConditions().Has(core.ConditionPetrified)
	if isPetrified {
		targetResistances = core.GetConditionEffects(core.ConditionPetrified).TemporaryResistance
	}

	result := core.DamageModificationResult{
		OriginalValue:  value,
		FinalValue:     value,
		ResistanceType: core.ResistanceNone,
	}

	if targetResistances == nil {
		return result, fmt.Errorf("target resistances are nil")
	}

	// Safe lookup: default to ResistanceNone when key is missing
	resistance := targetResistances.GetResistanceType(dt)
	result.ResistanceType = resistance

	// Resistance can only be broken if the attacker actually provides at least one breaker
	// and those breakers satisfy the target's breaker requirements for this damage type.
	brokenRes := len(b) > 0 && targetResistances.DamageTypeContainsAllBreakers(dt, b)
	result.ResistanceBroken = brokenRes
	result.ResistanceType = resistance

	switch resistance {
	case core.ResistanceNone:
		break
	case core.ResistanceVulnerable:
		if !brokenRes {
			result.FinalValue *= 2
		}
	case core.ResistanceResistant:
		if !brokenRes {
			result.FinalValue /= 2
		}
	case core.ResistanceImmune:
		if !brokenRes {
			result.FinalValue = 0
		}
	default:
		// Provide richer diagnostics to help identify unexpected values
		return core.DamageModificationResult{}, fmt.Errorf(
			"unknown resistance type for %s: damageType=%s, rawType=%q",
			target.GetName(), dt.String(), resistance,
		)
	}

	result.WasModified = result.FinalValue != value

	return result, nil
}

func (ce *CombatEngine) applyLimitedMagicImmunity(target core.Entity, effect *core.Effect) {
	if effect.SpellCtx == nil {
		return
	}

	if m, ok := target.(*monster.Monster); ok {
		if ce.SimOptions != nil && ce.SimOptions.EnableSpecialAbilities {
			if m.SpecialAbilities.LimitedMagicImmunityLevel > 0 {
				if effect.SpellCtx.SpellLevel <= m.SpecialAbilities.LimitedMagicImmunityLevel {
					effect.Value = 0
				}
			}
		}
	}
}

func (ce *CombatEngine) applyLightningAbsorption(target core.Entity, effect *core.Effect) {
	if effect.Type != core.EffectDamage || effect.DamageType != core.DamageLightning {
		return
	}

	if m, ok := target.(*monster.Monster); ok {
		if ce.SimOptions != nil && ce.SimOptions.EnableSpecialAbilities {
			if m.SpecialAbilities.LightningAbsorption {
				// Convert damage to healing
				effect.Type = core.EffectHealing
				// Value remains the same (it was damage value, now it's healing value)
				m.LogEvent(events.ETSpecialAbilityEvent, &events.SpecialAbilityData{
					AbilityName: "Lightning Absorption",
					Description: fmt.Sprintf("%s absorbs lightning damage and is healed!", m.GetName()),
					TargetName:  "",
					Value:       effect.Value,
				})
			}
		}
	}
}

// applyEvasionToEffect applies the evasion feature effects for rogues and monks, modifying the effect value based on saving throws.
func (ce *CombatEngine) applyEvasionToEffect(target core.Entity, effect *core.Effect) {
	switch t := target.(type) {
	case *character.Character:
		// Require a valid saving throw context and that it is a Dexterity save
		if effect == nil || effect.SaveCtx == nil || effect.SaveCtx.Ability != core.AbilityDexterity {
			return
		}

		// Only apply when class features are enabled (if options provided)
		if ce.SimOptions != nil && !ce.SimOptions.EnableClassFeatures {
			return
		}

		// Only Rogues/Monks have access to Evasion in this model
		if !(t.Class.ID == classes.Rogue || t.Class.ID == classes.Monk) {
			return // Ignore non-rogue/monk targets
		}
		hasEvasion := false

		switch t.Class.ID {
		case classes.Rogue:
			hasEvasion = t.Class.ClassFeatures.RogueFeatures.HasEvasion
		case classes.Monk:
			hasEvasion = t.Class.ClassFeatures.MonkFeatures.HasEvasion
		default:
			return // Ignore non-rogue/monk targets
		}

		if !hasEvasion {
			return // Don't apply evasion if target doesn't have access
		}

		if effect.SaveCtx.Success && effect.SaveCtx.OnSuccess == core.DCOnSuccessHalf {
			effect.Value = 0 // Feature reduces damage to zero
		} else if !effect.SaveCtx.Success {
			effect.Value /= 2
		}
	case *monster.Monster:
		if effect == nil || effect.SaveCtx == nil || effect.SaveCtx.Ability != core.AbilityDexterity {
			return
		}

		if ce.SimOptions != nil && !ce.SimOptions.EnableSpecialAbilities {
			return
		}
		if !t.SpecialAbilities.Evasion {
			return
		}

		if effect.SaveCtx.Success && effect.SaveCtx.OnSuccess == core.DCOnSuccessHalf {
			effect.Value = 0 // Feature reduces damage to zero
		} else if !effect.SaveCtx.Success {
			effect.Value /= 2
		}
	default:
		return
	}
}

func (ce *CombatEngine) applyUncannyDodgeToEffect(target core.Entity, effect *core.Effect) {
	targetChar, ok := target.(*character.Character)
	if !ok {
		return // Ignore non-character targets
	}

	// Require a valid saving throw context and that it is a Dexterity save
	if effect == nil {
		return
	}

	// Only apply when class features are enabled (if options provided)
	if ce.SimOptions != nil && !ce.SimOptions.EnableClassFeatures {
		return
	}

	if targetChar.Class.ID != classes.Rogue {
		return
	}

	if targetChar.Class.ClassFeatures.RogueFeatures.HasUncannyDodge &&
		!targetChar.EntityStateManager.GetHasUsedReaction() {
		effect.Value /= 2
		targetChar.EntityStateManager.ExpendReaction()
		return
	}
}

func (ce *CombatEngine) applyDeflectMissiles(target core.Entity, effect *core.Effect) {
	targetChar, ok := target.(*character.Character)
	if !ok {
		return // Ignore non-character targets
	}

	// Require a valid saving throw context and that it is a Dexterity save
	if effect == nil {
		return
	}

	// Only apply when class features are enabled (if options provided)
	if ce.SimOptions != nil && !ce.SimOptions.EnableClassFeatures {
		return
	}

	if targetChar.Class.ID != classes.Monk {
		return
	}

	if targetChar.Class.ClassFeatures.MonkFeatures != nil &&
		targetChar.Class.ClassFeatures.MonkFeatures.HasDeflectMissiles &&
		effect.AttackCtx.IsRanged {
		dexMod, err := targetChar.GetAbilityScoreModifier(core.AbilityDexterity)
		if err != nil {
			return
		}
		roll := targetChar.RollManager.RollDie(core.D10)
		effect.Value = int(math.Max(0, float64(effect.Value)-float64(dexMod)-float64(roll)-float64(targetChar.Level)))
		targetChar.EntityStateManager.ExpendReaction()
		return
	}
}

func (ce *CombatEngine) recordDeathSaves(combatantID int, status *core.TurnResult) {
	if status == nil {
		return
	}
	combatant, exists := ce.Combatants[combatantID]
	if !exists {
		return
	}

	if status.TurnStatuses[core.TurnDeathSaveSuccess] {
		combatant.Info.Statistics.RecordDeathSave(true)
	} else if status.TurnStatuses[core.TurnDeathSaveFailed] {
		combatant.Info.Statistics.RecordDeathSave(false)
	} else if status.TurnStatuses[core.TurnDeathSaveFailedDouble] {
		// Natural 1 counts as two failures
		combatant.Info.Statistics.RecordDeathSave(false)
		combatant.Info.Statistics.RecordDeathSave(false)
	} else if status.TurnStatuses[core.TurnRevived] {
		// Natural 20 counts as a success and revives
		combatant.Info.Statistics.RecordDeathSave(true)
	}
}
