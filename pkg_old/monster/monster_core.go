package monster

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/events"
	"dnd5e-encounter-simulator-backend/pkg_old/core/roll_manager"
	"fmt"
)

func (m *Monster) ProcessTurn(actorID int, turnType core.TurnType) (*core.TurnResult, *core.AIRequest, error) {
	if turnType == core.TurnTypeLegendary {
		return m.processLegendaryTurn(actorID)
	}

	// Turn start logging moved to CombatEngine.turnStartEvents()

	result := &core.TurnResult{
		TurnStatuses: make(map[core.TurnStatus]bool),
	}

	// Able to act
	if m.EntityStateManager.CanTakeActions() {
		aiReq, err := m.GetAIRequest(actorID, core.AIReqNormalAction)
		if err != nil {
			return nil, nil, err
		}
		if aiReq != nil {
			result.TurnStatuses[core.TurnActionReady] = true
		}

		return result, aiReq, nil
	}

	// Unable to Act
	if m.EntityStateManager.GetIsDead() {
		result.TurnStatuses[core.TurnDead] = true
		return result, nil, nil
	}

	if m.EntityStateManager.GetIsUnconscious() {
		ucResult, err := m.handleUnconsciousTurn(result)
		if ucResult.TurnStatuses[core.TurnRevived] {
			aiReq, err := m.GetAIRequest(actorID, core.AIReqNormalAction)
			if err != nil {
				return nil, nil, fmt.Errorf("error getting AI request: %s", err)
			}
			if aiReq != nil {
				ucResult.TurnStatuses[core.TurnActionReady] = true
			}
			ucResult.Conditions = nil

			return ucResult, aiReq, nil
		}
		return ucResult, nil, err
	}

	result.Conditions = m.EntityStateManager.GetActiveIncapacitatingConditions()
	if len(result.Conditions) > 0 {
		result.TurnStatuses[core.TurnIncapacitated] = true
	}
	return result, nil, nil
}

func (m *Monster) GetAIRequest(actorID int, t core.AIRequestType) (*core.AIRequest, error) {
	var req *core.AIRequest
	var err error
	switch t {
	case core.AIReqNormalAction:
		var actionChoice core.ActionType
		actionChoice, err = m.AI.ChooseMonsterActionType()
		if err != nil {
			return nil, err
		}
		if actionChoice == core.ATNoAction {
			return nil, nil
		}
		if actionChoice == core.ATMonsterHeal {
			req, err = m.AI.createMonsterHealActionRequest()
			if err != nil {
				return nil, err
			}
		} else {
			req, err = m.AI.createMonsterDamageActionRequest()
			if err != nil {
				return nil, err
			}
		}
	case core.AIReqLegendaryAction:
		req, err = m.AI.createMonsterLegendaryActionRequest()
		if err != nil {
			return nil, err
		}
	case core.AIReqDeathEffect:
		req = &core.AIRequest{
			ActionType: core.ATMonsterDeathEffect,
			TargetID:   -1, // Hits all in radius
		}
	case core.AIReqRetaliatoryEffect:
		req = &core.AIRequest{
			ActionType: core.ATMonsterRetaliatoryEffect,
		}
	default:
		return req, fmt.Errorf("invalid AI request type: %v", t)
	}

	if req == nil {
		return nil, nil
	}
	if req.Actor == nil {
		req.Actor = m
	}
	// Logging for tactical action (only if weighted AI is not used, to avoid duplication)
	if m.AI.combatCtx == nil || !m.AI.combatCtx.Options.UseWeightedAI {
		m.LogEvent(events.ETActionChoiceEvent, &events.ActionChoiceData{
			Choice:       req.ActionType,
			AllScores:    nil,
			TopReasons:   nil,
			UtilityScore: 0,
		})
	}

	req.ActorID = actorID
	req.Actor = m
	return req, nil
}

func (m *Monster) ExecuteAIRequest(req *core.AIRequest) (*core.ActionOutcome, error) {
	if req.Actor == nil {
		req.Actor = m
	}

	switch req.ActionType {
	case core.ATMonsterDivineEminence:
		if m.SpellCastingManager != nil {
			// Find highest slot to expend
			slotToExpend := -1
			for i := 9; i >= 1; i-- {
				if m.SpellCastingManager.HasSpellSlotsAtLevel(i) {
					slotToExpend = i
					break
				}
			}

			if slotToExpend != -1 {
				m.SpellCastingManager.ExpendSpellSlot(slotToExpend)
				m.EntityStateManager.ExpendBonusAction()
				m.EntityStateManager.SetDivineEminenceActive(true)

				// Bonus damage: +1d6 for each level above 1st
				// We already have m.SpecialAbilities.DivineEminenceNumDice as the base (3d6 for level 1 slot)
				// The ability says: "If the priest expends a spell slot of 2nd level or higher, the extra damage increases by 1d6 for each level above 1st."
				// So if they expend a 2nd level slot, it should be base + 1.
				// However, our m.SpecialAbilities is fixed.
				// We need a way to track the CURRENT number of dice for the activation.
				// Let's add a field to ESM to track the active dice count.

				activeDice := m.SpecialAbilities.DivineEminenceNumDice + (slotToExpend - 1)
				m.EntityStateManager.SetDivineEminenceDice(activeDice)

				m.LogEvent(events.ETSpecialAbilityEvent, &events.SpecialAbilityData{
					AbilityName: "Divine Eminence Activation",
					Description: fmt.Sprintf("%s activates Divine Eminence (expended level %d slot, %d dice)!", m.Name, slotToExpend, activeDice),
					TargetName:  "",
					Value:       slotToExpend,
				})
			}
		}
		return &core.ActionOutcome{
			ActionType: req.ActionType,
			ActorID:    req.ActorID,
			Success:    true,
		}, nil

	case core.ATMonsterAction, core.ATMonsterMultiattack, core.ATMonsterSpecial:
		// ExpendAction() moved to CombatEngine.ProcessAIRequest path for better turn control
		// Reckless special ability
		if req.SimOptions != nil && req.SimOptions.EnableSpecialAbilities {
			m.EntityStateManager.SetIsRecklesslyAttacking(true)
			m.EntityStateManager.AddCondition(core.ConditionRecklessExposed)
		}

		attackReq, err := m.createAttackRequest(req.Target, req.ActionIndex, req.ActionType, req.SimOptions)
		if err != nil {
			return nil, err
		}

		results, err := m.ActionManager.ProcessAttackRequest(attackReq)
		if err != nil {
			return nil, err
		}

		// Recharge action
		if m.ActionManager.Actions[req.ActionIndex].RechargeValue > 0 {
			m.EntityStateManager.ExpendRechargeAction(req.ActionIndex)
		}

		var effects []core.Effect
		var attackResults []core.AttackResult
		for _, res := range results {
			attackResults = append(attackResults, res)
			if res.GetIsHit() {
				// Create an effect for each damage component
				dmgRes := res.GetDamageResult()
				comps := dmgRes.GetDamageComponents()
				if len(comps) > 0 {
					for _, comp := range comps {
						effects = append(effects, core.Effect{
							Type:         core.EffectDamage,
							Value:        comp.GetTotal(),
							BaseValue:    comp.GetTotal(),
							DamageType:   comp.GetDamageType(),
							SourceRollID: dmgRes.GetID(),
							AttackCtx: &core.AttackContext{
								IsRanged:   res.IsRanged,
								IsCritical: res.IsCriticalHit,
							},
						})
					}
				} else {
					// Fallback for single damage attacks
					effects = append(effects, core.Effect{
						Type:         core.EffectDamage,
						Value:        dmgRes.GetTotal(),
						BaseValue:    dmgRes.GetTotal(),
						DamageType:   res.GetDamageType(),
						SourceRollID: dmgRes.GetID(),
						AttackCtx: &core.AttackContext{
							IsRanged:   res.IsRanged,
							IsCritical: res.IsCriticalHit,
						},
					})
				}

				// Special abilities: Martial Advantage, Divine Eminence
				if req.SimOptions != nil && req.SimOptions.EnableSpecialAbilities {
					// Martial Advantage
					if m.SpecialAbilities.MartialAdvantageNumDice > 0 {
						maEffect := m.resolveMartialAdvantage(res.IsCriticalHit, req.SimOptions)
						if maEffect != nil {
							effects = append(effects, *maEffect)
						}
					}

					// Divine Eminence
					if m.SpecialAbilities.DivineEminenceNumDice > 0 && !res.IsRanged {
						deEffect := m.resolveDivineEminence(res.IsCriticalHit, req.SimOptions)
						if deEffect != nil {
							effects = append(effects, *deEffect)
						}
					}

					// Sneak Attack
					if m.SpecialAbilities.SneakAttackNumDice > 0 {
						saEffect := m.resolveSneakAttack(core.SneakAttackParams{
							IsCritical: res.IsCriticalHit,
							Advantage:  res.AdvantageUsed,
							DamageType: res.DamageType,
							IsRanged:   res.IsRanged,
							IsSpell:    false,
						}, req.SimOptions)
						if saEffect != nil {
							effects = append(effects, *saEffect)
						}
					}
				}
			}
		}

		return &core.ActionOutcome{
			ActionType:    req.ActionType,
			TargetID:      req.TargetID,
			ActorID:       req.ActorID,
			Effects:       effects,
			Success:       len(effects) > 0,
			AttackResults: attackResults,
		}, nil

	case core.ATLegendaryAction:
		attackReq, err := m.createAttackRequest(req.Target, req.ActionIndex, req.ActionType, req.SimOptions)
		if err != nil {
			return nil, err
		}

		results, err := m.ActionManager.ProcessAttackRequest(attackReq)
		if err != nil {
			return nil, err
		}

		// Legendary actions
		cost := m.ActionManager.LegendaryActions[req.ActionIndex].Cost
		m.EntityStateManager.ExpendLegendaryActionPoints(cost)

		var effects []core.Effect
		var attackResults []core.AttackResult
		for _, res := range results {
			attackResults = append(attackResults, res)
			if res.GetIsHit() {
				// Create an effect for each damage component
				dmgRes := res.GetDamageResult()
				comps := dmgRes.GetDamageComponents()
				if len(comps) > 0 {
					for _, comp := range comps {
						effects = append(effects, core.Effect{
							Type:         core.EffectDamage,
							Value:        comp.GetTotal(),
							BaseValue:    comp.GetTotal(),
							DamageType:   comp.GetDamageType(),
							SourceRollID: dmgRes.GetID(),
							AttackCtx: &core.AttackContext{
								IsRanged:   res.IsRanged,
								IsCritical: res.IsCriticalHit,
							},
						})
					}
				} else {
					// Fallback for single damage attacks
					effects = append(effects, core.Effect{
						Type:         core.EffectDamage,
						Value:        dmgRes.GetTotal(),
						BaseValue:    dmgRes.GetTotal(),
						DamageType:   res.GetDamageType(),
						SourceRollID: dmgRes.GetID(),
						AttackCtx: &core.AttackContext{
							IsRanged:   res.IsRanged,
							IsCritical: res.IsCriticalHit,
						},
					})
				}

				// Special abilities: Martial Advantage, Divine Eminence
				if req.SimOptions != nil && req.SimOptions.EnableSpecialAbilities {
					// Martial Advantage
					if m.SpecialAbilities.MartialAdvantageNumDice > 0 {
						maEffect := m.resolveMartialAdvantage(res.IsCriticalHit, req.SimOptions)
						if maEffect != nil {
							effects = append(effects, *maEffect)
						}
					}

					// Divine Eminence
					if m.SpecialAbilities.DivineEminenceNumDice > 0 && !res.IsRanged {
						deEffect := m.resolveDivineEminence(res.IsCriticalHit, req.SimOptions)
						if deEffect != nil {
							effects = append(effects, *deEffect)
						}
					}

					// Sneak Attack
					if m.SpecialAbilities.SneakAttackNumDice > 0 {
						saEffect := m.resolveSneakAttack(core.SneakAttackParams{
							IsCritical: res.IsCriticalHit,
							Advantage:  res.AdvantageUsed,
							DamageType: res.DamageType,
							IsRanged:   res.IsRanged,
							IsSpell:    false,
						}, req.SimOptions)
						if saEffect != nil {
							effects = append(effects, *saEffect)
						}
					}
				}
			}
		}

		return &core.ActionOutcome{
			ActionType:    req.ActionType,
			TargetID:      req.TargetID,
			ActorID:       req.ActorID,
			Effects:       effects,
			Success:       len(effects) > 0,
			AttackResults: attackResults,
		}, nil

	case core.ATSpell:
		// ExpendAction() moved to CombatEngine.ProcessAIRequest path
		scReq, err := m.createSpellCastRequest(req.Target, *req.SpellChoice, req.SimOptions)
		if err != nil {
			return nil, err
		}

		res, err := m.SpellCastingManager.CastSpell(scReq)
		if err != nil {
			return nil, err
		}

		var effects []core.Effect
		var attackResults []core.AttackResult
		if !req.SpellChoice.Spell.GetHasDC() && !req.SpellChoice.Spell.GetIsAutoHit() {
			attackResults = append(attackResults, core.AttackResult{
				IsHit:         res.GetIsHit(),
				IsCriticalHit: res.GetIsCriticalHit(),
				AttackRoll:    res.GetAttackRoll(),
				AttackTotal:   res.GetAttackTotal(),
				AttackName:    res.SpellName,
			})
		}
		if res.GetIsHit() || req.SpellChoice.Spell.GetIsAutoHit() {
			if req.SpellChoice.Spell.GetSpellType() == core.STDamage {
				effects = append(effects, core.Effect{
					Type:         core.EffectDamage,
					Value:        res.GetSpellTotalValue(),
					BaseValue:    res.ValueRoll.GetTotal(),
					DamageType:   res.GetDamageType(),
					SourceRollID: res.ValueRoll.GetID(),
					SaveCtx: &core.SaveContext{
						Ability:   res.SpellSaveAbility,
						TargetDC:  res.TargetDCValue,
						Success:   res.SpellSaveSuccess,
						OnSuccess: res.SpellSaveEffect,
					},
					AttackCtx: &core.AttackContext{
						IsRanged:   false,
						IsCritical: res.GetIsCriticalHit(),
					},
					SpellCtx: &core.SpellContext{
						SpellLevel: req.SpellChoice.Formula.CastLevel,
					},
				})
			} else if req.SpellChoice.Spell.GetSpellType() == core.STHealing {
				effects = append(effects, core.Effect{
					Type:         core.EffectHealing,
					Value:        res.GetSpellTotalValue(),
					SourceRollID: res.ValueRoll.GetID(),
					SpellCtx: &core.SpellContext{
						SpellLevel: req.SpellChoice.Formula.CastLevel,
					},
				})
			}
		}

		return &core.ActionOutcome{
			ActionType:      req.ActionType,
			TargetID:        req.TargetID,
			ActorID:         req.ActorID,
			Effects:         effects,
			Success:         len(effects) > 0,
			IsConcentration: res.IsConcentration,
			SpellName:       res.SpellName,
			IsAOE:           res.IsAOE,
			AttackResults:   attackResults,
		}, nil
	case core.ATMonsterHeal:
		// ExpendAction() moved to CombatEngine.ProcessAIRequest path
		hReq := req.HealRequest
		if hReq == nil {
			return nil, fmt.Errorf("missing heal request")
		}

		var healingValue int
		if hReq.Source == core.HealSourceSpell {
			scReq, err := m.createSpellCastRequest(hReq.Target, *hReq.SpellChoice, req.SimOptions)
			if err != nil {
				return nil, err
			}
			res, err := m.SpellCastingManager.CastSpell(scReq)
			if err != nil {
				return nil, err
			}
			healingValue = res.GetSpellTotalValue()
		} else {
			return nil, fmt.Errorf("unsupported healing source for monster: %v", hReq.Source)
		}

		return &core.ActionOutcome{
			ActionType: req.ActionType,
			TargetID:   req.TargetID,
			ActorID:    req.ActorID,
			Effects: []core.Effect{
				{
					Type:  core.EffectHealing,
					Value: healingValue,
				},
			},
			Success: true,
		}, nil
	case core.ATMonsterDeathEffect:
		var effects []core.Effect
		var isAOE bool
		var spellName string

		opts := roll_manager.NewRollOptions()

		if m.SpecialAbilities.DeathBurstNumDice > 0 {
			damage, _ := m.RollManager.RollDice(m.SpecialAbilities.DeathBurstNumDice, core.D8, opts)
			effects = append(effects, core.Effect{
				Type:       core.EffectDamage,
				Value:      damage.GetTotal(),
				BaseValue:  damage.GetTotal(),
				DamageType: m.SpecialAbilities.DeathBurstDamageType,
				SaveCtx: &core.SaveContext{
					Ability:   core.AbilityDexterity,
					TargetDC:  m.SpecialAbilities.DeathBurstDC,
					OnSuccess: core.DCOnSuccessHalf,
				},
			})
			isAOE = true
			spellName = "Death Burst"
		} else if m.SpecialAbilities.DeathThroesNumDice > 0 {
			damage, _ := m.RollManager.RollDice(m.SpecialAbilities.DeathThroesNumDice, core.D6, opts)
			effects = append(effects, core.Effect{
				Type:       core.EffectDamage,
				Value:      damage.GetTotal(),
				BaseValue:  damage.GetTotal(),
				DamageType: core.DamageFire,
				SaveCtx: &core.SaveContext{
					Ability:   core.AbilityDexterity,
					TargetDC:  m.SpecialAbilities.DeathThroesDC,
					OnSuccess: core.DCOnSuccessHalf,
				},
			})
			isAOE = true
			spellName = "Death Throes"
		}

		return &core.ActionOutcome{
			ActionType: req.ActionType,
			ActorID:    req.ActorID,
			TargetID:   -1,
			Effects:    effects,
			Success:    len(effects) > 0,
			IsAOE:      isAOE,
			SpellName:  spellName,
		}, nil
	case core.ATMonsterRetaliatoryEffect:
		var effects []core.Effect
		var spellName string

		opts := roll_manager.NewRollOptions()

		// Corrosive Form
		if m.SpecialAbilities.CorrosiveFormNumDice > 0 {
			damage, _ := m.RollManager.RollDice(m.SpecialAbilities.CorrosiveFormNumDice, core.D8, opts)
			effects = append(effects, core.Effect{
				Type:       core.EffectDamage,
				Value:      damage.GetTotal(),
				BaseValue:  damage.GetTotal(),
				DamageType: core.DamageAcid,
			})
			if spellName != "" {
				spellName += " & Corrosive Form"
			} else {
				spellName = "Corrosive Form"
			}
		}

		// Fire Form / Fire Aura (retaliatory part)
		if m.SpecialAbilities.FireForm || m.SpecialAbilities.FireAuraNumDice > 0 {
			// Balor's Fire Aura is 3d6, Fire Form is usually 1d10 or similar but often flat
			// In our csv Fire Aura is 3d6. Fire Form doesn't specify dice in csv yet but we have the flag.
			// Let's use FireAuraNumDice if available, fallback to 1d10 for generic Fire Form if we want,
			// but better to stick to what we have in SpecialAbilities struct.
			numDice := m.SpecialAbilities.FireAuraNumDice
			if numDice == 0 && m.SpecialAbilities.FireForm {
				numDice = 1 // default to 1 die if not specified?
			}

			if numDice > 0 {
				damage, _ := m.RollManager.RollDice(numDice, core.D6, opts)
				effects = append(effects, core.Effect{
					Type:       core.EffectDamage,
					Value:      damage.GetTotal(),
					BaseValue:  damage.GetTotal(),
					DamageType: core.DamageFire,
				})
				name := "Fire Form"
				if m.SpecialAbilities.FireAuraNumDice > 0 {
					name = "Fire Aura"
				}
				if spellName != "" {
					spellName += " & " + name
				} else {
					spellName = name
				}
			}
		}

		// Heated Body
		if m.SpecialAbilities.HeatedBodyNumDice > 0 {
			damage, _ := m.RollManager.RollDice(m.SpecialAbilities.HeatedBodyNumDice, core.D10, opts)
			effects = append(effects, core.Effect{
				Type:       core.EffectDamage,
				Value:      damage.GetTotal(),
				BaseValue:  damage.GetTotal(),
				DamageType: core.DamageFire,
			})
			if spellName != "" {
				spellName += " & Heated Body"
			} else {
				spellName = "Heated Body"
			}
		}

		return &core.ActionOutcome{
			ActionType: req.ActionType,
			ActorID:    req.ActorID,
			TargetID:   req.TargetID,
			Effects:    effects,
			Success:    len(effects) > 0,
			SpellName:  spellName,
		}, nil
	default:
		return nil, fmt.Errorf("monster execute ai req - invalid action type: %s", req.ActionType)
	}
}

func (m *Monster) handleUnconsciousTurn(turnResult *core.TurnResult) (*core.TurnResult, error) {
	// Failsafes if character is already dead and this is called
	if m.EntityStateManager.GetIsDead() {
		turnResult.TurnStatuses[core.TurnDead] = true
		return turnResult, nil
	}

	if !m.EntityStateManager.GetIsUnconscious() {
		return nil, fmt.Errorf("character is not unconscious")
	}

	// Character is not dead but is unconscious
	if m.EntityStateManager.GetIsStable() {
		turnResult.TurnStatuses[core.TurnUnconscious] = true
		return turnResult, nil
	}

	// Roll death saving throw
	res, err := m.RollManager.RollDeathSavingThrow()
	if err != nil {
		return nil, fmt.Errorf("failed to roll death saving throw: %v", err)
	}

	// Apply death saving throw turnResult
	err = m.EntityStateManager.ApplyDeathSavingThrowResult(res)
	if err != nil {
		return nil, fmt.Errorf("failed to apply death saving throw turnResult: %v", err)
	}

	// Determine turn status
	switch {
	case res.IsCritical:
		turnResult.TurnStatuses[core.TurnRevived] = true
	case res.IsNaturalOne:
		turnResult.TurnStatuses[core.TurnDeathSaveFailedDouble] = true
	case res.IsSuccess:
		turnResult.TurnStatuses[core.TurnDeathSaveSuccess] = true
	default:
		turnResult.TurnStatuses[core.TurnDeathSaveFailed] = true
	}

	turnResult.Conditions = m.EntityStateManager.GetActiveIncapacitatingConditions()
	return turnResult, nil
}

func (m *Monster) processLegendaryTurn(actorID int) (*core.TurnResult, *core.AIRequest, error) {
	if !m.IsLegendary || len(m.ActionManager.LegendaryActions) == 0 {
		return nil, nil, fmt.Errorf("monster is not legendary")
	}

	result := &core.TurnResult{
		TurnStatuses: make(map[core.TurnStatus]bool),
	}

	if !m.EntityStateManager.HasLegendaryActionPointsRemaining() {
		result.TurnStatuses[core.TurnLegendaryUnavailable] = true
		return result, nil, nil
	}

	var canAffordLegAction bool
	pointsRemaining := m.EntityStateManager.GetLegendaryActionPoints()
	for _, action := range m.ActionManager.LegendaryActions {
		if action.Cost <= pointsRemaining {
			canAffordLegAction = true
			break
		}
	}

	if !canAffordLegAction {
		result.TurnStatuses[core.TurnLegendaryUnavailable] = true
		return result, nil, nil
	}

	legAIReq, err := m.GetAIRequest(actorID, core.AIReqLegendaryAction)

	if err != nil {
		return nil, nil, err
	}
	result.TurnStatuses[core.TurnLegendaryReady] = true
	return result, legAIReq, nil
}
