package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"fmt"
	"math"
)

func (ce *CombatEngine) processActionResults(actor core.Entity, outcome *core.ActionOutcome) error {
	// Update attack statistics
	if len(outcome.AttackResults) > 0 {
		actorCombatant, exists := ce.Combatants[outcome.ActorID]
		if exists {
			for _, res := range outcome.AttackResults {
				actorCombatant.Info.Statistics.RecordAttack(res.IsHit, res.IsCriticalHit)
			}
		}
	}

	// Handle concentration start if the action is a concentration spell
	if outcome.IsConcentration {
		actorCombatant, exists := ce.Combatants[outcome.ActorID]
		if exists {
			// Determine targets - for now we just use the primary target of the action
			// In the future, this could be expanded for AOE concentration spells
			targets := []int{outcome.TargetID}
			// Duration is usually 10 rounds (1 minute) for most combat spells, but we should ideally get it from the spell
			// For now, default to 10 rounds as a placeholder if not specified.
			// Most concentration spells in this system are likely 1 minute.
			duration := 10

			// We need the current round
			currentRound := ce.CurrentRound

			actorCombatant.Info.StartConcentration(outcome.SpellName, targets, duration, currentRound)
		}
	}

	// Identify targets for the action
	targetIDs := []int{outcome.TargetID}
	if outcome.IsAOE {
		if ce.SimOptions.AOEHitsAllEnemies || outcome.ActionType == core.ATMonsterDeathEffect {
			targetIDs = []int{}
			ids := ce.getSortedCombatantIDs()
			for _, id := range ids {
				combatant := ce.Combatants[id]
				if combatant.Entity.IsDead() {
					continue
				}
				// For death effects, we hit everyone EXCEPT the dying actor (who is already dead/processed)
				if outcome.ActionType == core.ATMonsterDeathEffect {
					if id != outcome.ActorID {
						// Check if we should only hit enemies
						if !ce.SimOptions.MonsterDeathEffectsHitAllies {
							if actor.IsMonster() && combatant.Entity.IsMonster() {
								continue
							}
							if actor.IsCharacter() && combatant.Entity.IsCharacter() {
								continue
							}
						}
						targetIDs = append(targetIDs, id)
					}
					continue
				}

				// Standard AOE hits all enemies
				if actor.IsCharacter() && combatant.Entity.IsMonster() {
					targetIDs = append(targetIDs, id)
				} else if actor.IsMonster() && combatant.Entity.IsCharacter() {
					targetIDs = append(targetIDs, id)
				}
			}
		}
	}

	for _, targetID := range targetIDs {
		target, exists := ce.Combatants[targetID]
		if !exists {
			continue // Should not happen for primary target, but possible for AOE if someone died mid-process
		}

		if len(outcome.Effects) > 0 {
			var hpModResult core.HPModificationResult
			var err error
			for _, effect := range outcome.Effects {
				currentEffect := effect
				if targetID != outcome.TargetID && currentEffect.SaveCtx != nil {
					// Re-evaluate saving throw for secondary targets
					saveRes, err := target.GetEntity().MakeSavingThrow(
						currentEffect.SaveCtx.Ability,
						currentEffect.SaveCtx.TargetDC,
						true,
						currentEffect.DamageType,
						ce.SimOptions)
					if err != nil {
						return fmt.Errorf("failed to make saving throw for AOE target: %v", err)
					}

					// Update currentEffect based on new save result
					if saveRes.GetIsSuccess() {
						if currentEffect.SaveCtx.OnSuccess == core.DCOnSuccessHalf {
							currentEffect.Value = currentEffect.BaseValue / 2
						} else if currentEffect.SaveCtx.OnSuccess == core.DCOnSuccessNone {
							currentEffect.Value = 0
						}
						// If OnSuccessOther, we might need more logic, but current system uses None/Half mostly
					} else {
						currentEffect.Value = currentEffect.BaseValue
					}
				}

				// Pre-processing effects that might change the type or value before standard logic
				ce.applyLimitedMagicImmunity(target.GetEntity(), &currentEffect)
				ce.applyLightningAbsorption(target.GetEntity(), &currentEffect)

				switch currentEffect.Type {
				case core.EffectDamage:
					ce.applyDeflectMissiles(target.GetEntity(), &currentEffect)
					ce.applyUncannyDodgeToEffect(target.GetEntity(), &currentEffect)
					ce.applyEvasionToEffect(target.GetEntity(), &currentEffect)
					res, rErr := ce.computeDamageValueAfterResistances(
						target.GetEntity(),
						currentEffect.DamageType,
						currentEffect.ResistBreakers,
						-currentEffect.Value)
					if rErr != nil {
						return rErr
					}
					actor.LogEvent(events.ETDamageModifiedEvent, &events.DamageModifiedData{
						Subject:      target.GetEntity(),
						Res:          res,
						SourceRollID: currentEffect.SourceRollID,
					})
					hpModResult, err = target.GetEntity().ModifyHP(res.FinalValue, false, false, ce.SimOptions.UseMassiveDamage, currentEffect.DamageType, currentEffect.AttackCtx != nil && currentEffect.AttackCtx.IsCritical)
					if err != nil {
						return fmt.Errorf("failed to modify target entity HP: %v", err)
					}

					// Update statistics and global combat context
					damageValue := hpModResult.GetDamageTaken()
					if damageValue > 0 {
						// Global tracking
						if ce.CombatContext == nil {
							ce.initializeCombatContext()
						}
						if damageValue > ce.CombatContext.MaxDamageSeen {
							ce.CombatContext.MaxDamageSeen = damageValue
						}

						// actor statistics
						actorCombatant, actorExists := ce.Combatants[outcome.ActorID]
						if actorExists {
							actorCombatant.Info.Statistics.RecordDamageDealt(damageValue, ce.CurrentRound)
						}

						// Target statistics
						target.Info.Statistics.RecordDamageTaken(damageValue)
						target.Info.Statistics.LastAttackerID = outcome.ActorID
					}

					// Check for death effects
					if target.GetEntity().IsDead() {
						if m, ok := target.GetEntity().(*monster.Monster); ok && ce.SimOptions.EnableSpecialAbilities {
							if m.SpecialAbilities.DeathBurstNumDice > 0 || m.SpecialAbilities.DeathThroesNumDice > 0 {
								deathReq, _ := m.GetAIRequest(m.GetInstanceID(), core.AIReqDeathEffect)
								if deathReq != nil {
									deathOutcome, _ := m.ExecuteAIRequest(deathReq)
									if deathOutcome != nil {
										events.LogCombatEventMessage(m.GetCurrentEventContext(), m, fmt.Sprintf("%s triggers %s!", m.GetName(), deathOutcome.SpellName), m.GetEventListener())
										// Process death effect recursively
										ce.processActionResults(m, deathOutcome)
									}
								}
							}
						}
					} else {
						// Check for retaliatory effects (Corrosive Form, Fire Form, Heated Body)
						// Triggers if hit by a melee attack within 5ft.
						// Our system doesn't explicitly track range in distance units yet, but we have isRanged in AttackContext.
						// Assume all melee hits are within 5ft for these purposes.
						if m, ok := target.GetEntity().(*monster.Monster); ok && ce.SimOptions.EnableSpecialAbilities {
							if currentEffect.AttackCtx != nil && !currentEffect.AttackCtx.IsRanged {
								if m.SpecialAbilities.CorrosiveFormNumDice > 0 || m.SpecialAbilities.FireForm ||
									m.SpecialAbilities.FireAuraNumDice > 0 || m.SpecialAbilities.HeatedBodyNumDice > 0 {

									retalliationReq, _ := m.GetAIRequest(m.GetInstanceID(), core.AIReqRetaliatoryEffect)
									if retalliationReq != nil {
										retalliationReq.Target = actor
										retalliationReq.TargetID = outcome.ActorID
										retalOutcome, _ := m.ExecuteAIRequest(retalliationReq)
										if retalOutcome != nil {
											events.LogCombatEventMessage(m.GetCurrentEventContext(), m, fmt.Sprintf("%s triggers retaliatory %s against %s!", m.GetName(), retalOutcome.SpellName, actor.GetName()), m.GetEventListener())
											// Process retaliatory effect
											ce.processActionResults(m, retalOutcome)
										}
									}
								}
							}
						}
					}
				case core.EffectHealing:
					v := math.Abs(float64(currentEffect.Value))
					hpModResult, err = target.GetEntity().ModifyHP(int(v), false, false, ce.SimOptions.UseMassiveDamage, core.DamageNone, false)
					if err != nil {
						return fmt.Errorf("failed to modify target entity HP: %v", err)
					}

					// Update statistics
					healingValue := hpModResult.GetHealingReceived()
					if healingValue > 0 {
						// actor statistics
						actorCombatant, actorExists := ce.Combatants[outcome.ActorID]
						if actorExists {
							actorCombatant.Info.Statistics.RecordHealingDone(healingValue, ce.CurrentRound)
						}

						// Target statistics
						target.Info.Statistics.RecordHealingReceived(healingValue)
						target.Info.Statistics.TurnsSinceLastHeal = 0
					}
				case core.EffectTempHP:
					v := math.Abs(float64(currentEffect.Value))
					hpModResult, err = target.GetEntity().ModifyHP(int(v), true, false, ce.SimOptions.UseMassiveDamage, core.DamageNone, false)
					if err != nil {
						return fmt.Errorf("failed to modify target entity HP: %v", err)
					}
				case core.EffectCondition:
					return fmt.Errorf("effects of type %v are not supported", core.EffectCondition)
				}

				// Log after each effect's HP modification for clarity
				actor.LogEvent(events.ETHPModifiedEvent, &events.HPModifiedData{
					Subject:      target.GetEntity(),
					Res:          hpModResult,
					SourceRollID: currentEffect.SourceRollID,
				})

				// Handle concentration check if triggered
				if hpModResult.GetTriggeredConcentrationCheck() {
					damageTaken := hpModResult.GetDamageTaken()
					dc := max(10, damageTaken/2)
					saveResult, err := target.GetEntity().MakeSavingThrow(core.AbilityConstitution, dc, false, core.DamageNone, ce.SimOptions)
					if err != nil {
						return fmt.Errorf("failed to make concentration check: %v", err)
					}

					if !saveResult.GetIsSuccess() {
						target.Info.BreakConcentration()
						events.LogCombatEventMessage(target.GetEntity().GetCurrentEventContext(), target.GetEntity(), "Failed concentration check. Concentration broken.", target.GetEntity().GetEventListener())
					} else {
						events.LogCombatEventMessage(target.GetEntity().GetCurrentEventContext(), target.GetEntity(), "Succeeded concentration check. Concentration maintained.", target.GetEntity().GetEventListener())
					}
				}

				// Persist any state changes immediately
				ce.Combatants[targetID] = target

				// If target is down or dead, check victory
				if target.GetEntity().IsDead() || target.GetEntity().IsUnconscious() {
					if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
						return nil
					}
				}

				// Early victory check after each effect
				if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
					return nil
				}
			}
			// Final state persist
			ce.Combatants[targetID] = target
		}
	}

	return nil
}
