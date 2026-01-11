package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/races"
	"errors"
	"fmt"
	"log"
)

func (c *Character) ProcessTurn(actorID int, turnType core.TurnType) (*core.TurnResult, *core.AIRequest, error) {
	if turnType == core.TurnTypeLegendary {
		return nil, nil, fmt.Errorf("invalid turn type for character: %s", turnType)
	}

	c.LogEvent(events.ETTurnStartEvent, nil)

	result := &core.TurnResult{
		TurnStatuses: make(map[core.TurnStatus]bool),
	}

	// Start-of-turn cleanup for Reckless Attack lifecycle
	// Clear last turn's exposure and reset attacking flag; it will be re-enabled by AI/ExecuteAIRequest if chosen again
	if c.EntityStateManager.HasCondition(core.ConditionRecklessExposed) {
		c.EntityStateManager.RemoveCondition(core.ConditionRecklessExposed)
	}
	if c.EntityStateManager.GetIsRecklesslyAttacking() {
		c.EntityStateManager.SetIsRecklesslyAttacking(false)
	}

	// Able to act
	if c.EntityStateManager.CanTakeActions() {
		aiReq, err := c.GetAIRequest(actorID, core.AIReqNormalAction)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting AI request: %s", err)
		}
		if aiReq != nil {
			result.TurnStatuses[core.TurnActionReady] = true
		}

		return result, aiReq, nil
	}

	// Unable to Act
	if c.EntityStateManager.GetIsDead() {
		result.TurnStatuses[core.TurnDead] = true
		return result, nil, nil
	}

	if c.EntityStateManager.GetIsUnconscious() {
		ucResult, err := c.handleUnconsciousTurn(result)
		if err != nil {
			return ucResult, nil, err
		}
		if ucResult.TurnStatuses[core.TurnRevived] {
			aiReq, err := c.GetAIRequest(actorID, core.AIReqNormalAction)
			if err != nil {
				return nil, nil, fmt.Errorf("error getting AI request: %s", err)
			}
			if aiReq != nil {
				ucResult.TurnStatuses[core.TurnActionReady] = true
			}
			// On revive: clear Unconscious, set Prone to true so ranged/melee modifiers apply correctly
			c.EntityStateManager.SetUnconscious(false)
			c.EntityStateManager.AddCondition(core.ConditionProne)
			ucResult.Conditions = []core.Condition{core.ConditionProne}
			events.LogCombatEventMessage(c.GetCurrentEventContext(), c, "Revived from 0 HP: now Prone", c.EventListener)
			return ucResult, aiReq, nil
		}
		return ucResult, nil, err
	}

	result.Conditions = c.EntityStateManager.GetActiveIncapacitatingConditions()
	if len(result.Conditions) > 0 {
		result.TurnStatuses[core.TurnIncapacitated] = true
	}
	return result, nil, nil
}

func (c *Character) GetAIRequest(actorID int, t core.AIRequestType) (*core.AIRequest, error) {
	var req *core.AIRequest
	var err error
	switch t {
	case core.AIReqNormalAction:
		var actionChoice core.ActionType
		actionChoice, err = c.AI.ChooseCharacterActionType()
		if err != nil {
			return nil, err
		}

		if actionChoice == core.ATNoAction {
			return nil, nil
		}

		switch actionChoice {
		case core.ATDragonbornBreathWeapon:
			req, err = c.AI.createDragonbornBreathWeaponRequest()
			if err != nil {
				return nil, err
			}
		case core.ATDamage:
			req, err = c.AI.createCharacterDamageActionRequest()
			if err != nil {
				return nil, err
			}
		case core.ATHeal:
			req, err = c.AI.createCharacterHealActionRequest()
			if err != nil {
				return nil, err
			}
		}
	case core.AIReqOffhandAttack:
		return c.AI.createCharacterOffhandActionRequest()
	default:
		return req, fmt.Errorf("invalid AI request type: %v", t)
	}

	if req == nil {
		return nil, nil
	}
	req.ActorID = actorID
	req.Actor = c
	return req, nil
}

func (c *Character) ExecuteAIRequest(req *core.AIRequest) (*core.ActionOutcome, error) {
	if req.Actor == nil {
		req.Actor = c
		log.Printf("warning: monster execute ai req - actor is nil")
	}
	// Note: Advantage for weapon attacks is computed inside CreateAttackRequest using unified helper.
	// For spells, we still compute condition-based advantage below when needed.
	adv := core.DetermineAttackAdvantageFromConditions(req.Actor.GetConditions(), req.Target.GetConditions())

	switch req.ActionType {
	case core.ATMelee, core.ATRanged:
		c.EntityStateManager.ExpendAction()
		// If class features are enabled, decide whether to use Reckless Attack this turn (simple rule or override)
		if req.SimOptions != nil && req.SimOptions.EnableClassFeatures {
			// Basic policy: if config forces recklessness, enable; otherwise leave for future AI heuristics
			if req.SimOptions.BarbarianAlwaysRecklessAttack {
				c.EntityStateManager.SetIsRecklesslyAttacking(true)
				// Make the character exposed to incoming attacks until next turn
				c.EntityStateManager.AddCondition(core.ConditionRecklessExposed)
			}
		}

		attackReq, err := c.CreateAttackRequest(req.Target, req.WeaponSlot, req.UseVersatile, req.SimOptions)
		if err != nil {
			return nil, err
		}

		results, err := c.MartialAttackManager.ProcessAttackRequest(attackReq)
		if err != nil {
			return nil, err
		}

		var effects []core.Effect
		var attackResults []core.AttackResult
		for _, res := range results {
			attackResults = append(attackResults, res)
			if res.GetIsHit() {
				effects = append(effects, core.Effect{
					Type:           core.EffectDamage,
					Value:          res.GetDamageResult().GetTotal(),
					BaseValue:      res.GetDamageResult().GetTotal(),
					DamageType:     res.GetDamageType(),
					ResistBreakers: res.ResistBreakers,
					SourceRollID:   res.GetDamageResult().GetID(),
					AttackCtx: &core.AttackContext{
						IsRanged:   res.IsRanged,
						IsCritical: res.IsCriticalHit,
					},
				})

				// Divine Smite check
				if req.ActionType == core.ATMelee && req.SimOptions != nil && req.SimOptions.EnableClassFeatures {
					if c.Class.ClassFeatures.PaladinFeatures != nil && c.Class.ClassFeatures.PaladinFeatures.HasDivineSmite {
						useSmite := false
						if req.SimOptions.UseWeightedAI {
							useSmite = c.AI.ShouldExpendResource(req.Target, res.IsCriticalHit)
						} else {
							useSmite = req.SimOptions.PaladinAlwaysSmite
						}

						if useSmite {
							smiteEffect := c.resolveDivineSmite(req.Target, res.IsCriticalHit, req.SimOptions)
							if smiteEffect != nil {
								effects = append(effects, *smiteEffect)
							}
						}
					}
				}

				// Sneak Attack check
				if req.SimOptions != nil && req.SimOptions.EnableClassFeatures {
					if c.Class.ClassFeatures.RogueFeatures != nil && c.Class.ClassFeatures.RogueFeatures.NumOfSneakAttackDice > 0 {
						saEffect := c.resolveSneakAttack(core.SneakAttackParams{
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
	case core.ATOffhand:
		c.EntityStateManager.ExpendBonusAction()
		// Offhand attacks should not apply ability modifier to damage unless Two-Weapon Fighting style is present.
		attackReq, err := c.CreateOffhandAttackRequest(req.Target, req.SimOptions)
		if err != nil {
			return nil, err
		}

		results, err := c.MartialAttackManager.ProcessAttackRequest(attackReq)
		if err != nil {
			return nil, err
		}

		var effects []core.Effect
		var attackResults []core.AttackResult
		for _, res := range results {
			attackResults = append(attackResults, res)
			if res.GetIsHit() {
				effects = append(effects, core.Effect{
					Type:           core.EffectDamage,
					Value:          res.GetDamageResult().GetTotal(),
					BaseValue:      res.GetDamageResult().GetTotal(),
					DamageType:     res.GetDamageType(),
					ResistBreakers: res.ResistBreakers,
					SourceRollID:   res.GetDamageResult().GetID(),
					AttackCtx: &core.AttackContext{
						IsRanged:   false,
						IsCritical: res.IsCriticalHit,
					},
				})

				// Divine Smite check
				if req.SimOptions != nil && req.SimOptions.EnableClassFeatures {
					if c.Class.ClassFeatures.PaladinFeatures != nil && c.Class.ClassFeatures.PaladinFeatures.HasDivineSmite {
						useSmite := false
						if req.SimOptions.UseWeightedAI {
							useSmite = c.AI.ShouldExpendResource(req.Target, res.IsCriticalHit)
						} else {
							useSmite = req.SimOptions.PaladinAlwaysSmite
						}

						if useSmite {
							smiteEffect := c.resolveDivineSmite(req.Target, res.IsCriticalHit, req.SimOptions)
							if smiteEffect != nil {
								effects = append(effects, *smiteEffect)
							}
						}
					}
				}

				// Sneak Attack check
				if req.SimOptions != nil && req.SimOptions.EnableClassFeatures {
					if c.Class.ClassFeatures.RogueFeatures != nil && c.Class.ClassFeatures.RogueFeatures.NumOfSneakAttackDice > 0 {
						saEffect := c.resolveSneakAttack(core.SneakAttackParams{
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
		c.EntityStateManager.ExpendAction()
		scReq, err := c.CreateSpellCastRequest(req.Target, *req.SpellChoice, adv, req.SimOptions)
		if err != nil {
			return nil, err
		}

		res, err := c.SpellCastingManager.CastSpell(scReq)
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
				// Special handling for Magic Missile multi-dart logic if needed
				// For now, if it's Magic Missile, we might want to split it into multiple effects
				// but since they hit the same target (usually in this sim), one combined effect is fine.
				// However, if we want to be accurate to "multi-target logic", we should consider it.

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
	case core.ATDragonbornBreathWeapon:
		c.EntityStateManager.ExpendAction()
		if c.Race.DragonbornFeatures == nil {
			return nil, fmt.Errorf("character is not a dragonborn or missing breath weapon features")
		}

		// Calculate damage
		rollOpts := roll_manager.NewRollOptions()
		rollOpts.RollType = core.DiceRollDamage
		damage, err := c.RollManager.RollDice(c.Race.DragonbornFeatures.NumberOfDice, c.Race.DragonbornFeatures.Die, rollOpts)
		if err != nil {
			return nil, err
		}

		// Save DC: 8 + Con mod + Proficiency
		conMod, err := c.getAbilityScoreModifier(core.AbilityConstitution)
		if err != nil {
			return nil, err
		}
		pb, err := core.GetCharacterProficiencyBonus(c.Level)
		if err != nil {
			return nil, err
		}
		dc := 8 + conMod + pb

		// Target makes a saving throw based on color
		saveAbility := core.AbilityDexterity
		switch c.Race.DragonbornFeatures.AncestryColor {
		case races.DragonbornGreen, races.DragonbornSilver, races.DragonbornWhite:
			saveAbility = core.AbilityConstitution
		}

		saveRes, err := req.Target.MakeSavingThrow(saveAbility, dc, false, core.DamageNone, req.SimOptions)
		if err != nil {
			return nil, err
		}

		finalDamage := damage.GetTotal()
		if saveRes.GetIsSuccess() {
			finalDamage /= 2
		}

		// Log breath weapon attack event
		// TODO: Is this duplicated with the weighted ai
		c.LogEvent(events.ETDragonbornBreathWeaponEvent, &events.DragonbornBreathWeaponData{
			Target:      req.Target,
			DamageTotal: damage.GetTotal(),
			DamageType:  c.Race.DragonbornFeatures.DamageType.String(),
			DC:          dc,
			SaveAbility: saveAbility.String(),
			SaveSuccess: saveRes.GetIsSuccess(),
			SaveResult:  saveRes.GetTotal(),
		})
		c.LogEvent(events.ETDamageEvent, &events.DamageData{
			Target:     req.Target,
			DamageType: c.Race.DragonbornFeatures.DamageType.String(),
			Damage:     damage.GetTotal(),
			Rolls:      damage.GetFinalRolls(),
		})

		c.EntityStateManager.SetDBBreathWeaponUsed(true)

		return &core.ActionOutcome{
			ActionType: req.ActionType,
			TargetID:   req.TargetID,
			ActorID:    req.ActorID,
			IsAOE:      true,
			Effects: []core.Effect{
				{
					Type:       core.EffectDamage,
					Value:      finalDamage,
					BaseValue:  damage.GetTotal(),
					DamageType: c.Race.DragonbornFeatures.DamageType,
					SaveCtx: &core.SaveContext{
						Ability:   saveAbility,
						TargetDC:  dc,
						Success:   saveRes.GetIsSuccess(),
						OnSuccess: core.DCOnSuccessHalf,
					},
				},
			},
			Success: true,
		}, nil
	case core.ATHeal:
		c.EntityStateManager.ExpendAction()
		if req.SimOptions != nil && !req.SimOptions.AllowCharacterHeals {
			return nil, errors.New("character healing is disabled")
		}

		hReq := req.HealRequest
		if hReq == nil {
			return nil, errors.New("missing heal request")
		}

		var healingValue int
		if hReq.Source == core.HealSourceLayingOnHands {
			// Ability Logic: Lay on Hands
			pool := c.EntityStateManager.GetPaladinLayingOnHandsPool()
			if hReq.AbilityValue > pool {
				return nil, fmt.Errorf("insufficient Lay on Hands pool")
			}
			c.EntityStateManager.ModifyPaladinLayingOnHandsPool(-hReq.AbilityValue)
			healingValue = hReq.AbilityValue

			c.LogEvent(events.ETHealEvent, &events.LayOnHandsHealData{
				Subject: hReq.Target,
				Value:   healingValue,
			})
		} else if hReq.Source == core.HealSourceSpell {
			// Existing Spell Logic
			scReq, err := c.CreateSpellCastRequest(hReq.Target, *hReq.SpellChoice, hReq.Advantage, req.SimOptions)
			if err != nil {
				return nil, err
			}
			res, err := c.SpellCastingManager.CastSpell(scReq)
			if err != nil {
				return nil, err
			}
			healingValue = res.GetSpellTotalValue()
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
	}
	return nil, errors.New("invalid action type")
}

func (c *Character) handleUnconsciousTurn(turnResult *core.TurnResult) (*core.TurnResult, error) {
	// Failsafes if character is already dead and this is called
	if c.EntityStateManager.GetIsDead() {
		turnResult.TurnStatuses[core.TurnDead] = true
		return turnResult, nil
	}

	if !c.EntityStateManager.GetIsUnconscious() {
		return nil, fmt.Errorf("character is not unconscious")
	}

	// Character is not dead but is unconscious
	if c.EntityStateManager.GetIsStable() {
		turnResult.TurnStatuses[core.TurnUnconscious] = true
		return turnResult, nil
	}

	// Roll death saving throw
	res, err := c.RollManager.RollDeathSavingThrow()
	if err != nil {
		return nil, fmt.Errorf("failed to roll death saving throw: %v", err)
	}

	// Apply death saving throw turnResult
	err = c.EntityStateManager.ApplyDeathSavingThrowResult(res)
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

	turnResult.Conditions = c.EntityStateManager.GetActiveIncapacitatingConditions()
	return turnResult, nil
}

func (c *Character) resolveDivineSmite(target core.Entity, isCrit bool, simOptions *core.SimulationOptions) *core.Effect {
	// 1. Identify slot level to use
	slotLevel := -1
	if simOptions.PaladinUseHighestSmiteSlot {
		for l := 5; l >= 1; l-- {
			if c.SpellCastingManager.HasSpellSlotsAtLevel(l) {
				slotLevel = l
				break
			}
		}
	} else {
		for l := 1; l <= 5; l++ {
			if c.SpellCastingManager.HasSpellSlotsAtLevel(l) {
				slotLevel = l
				break
			}
		}
	}

	if slotLevel == -1 {
		return nil
	}

	// 2. Calculate damage dice
	// Divine Smite: 2d8 for 1st level, +1d8 per slot level above 1st, max 5d8.
	// 1st: 2, 2nd: 3, 3rd: 4, 4th: 5, 5th: 5
	numDice := 1 + slotLevel
	if numDice > 5 {
		numDice = 5
	}

	// Undead/Fiend bonus
	targetType := target.GetType()
	if targetType == monster.MTUndead || targetType == monster.MTFiend {
		numDice++
	}

	// 3. Roll damage
	rollOpts := roll_manager.NewRollOptions()
	rollOpts.RollType = core.DiceRollDamage

	var err error
	var res *roll_manager.RollResult

	if isCrit && simOptions.CanCharactersCrit {
		if simOptions.UseImprovedCriticals {
			dmgRollTotal, dmgRolls := c.RollManager.RollExtraMaxDice(numDice, core.D8)
			res = &roll_manager.RollResult{
				ID:             core.NewUUIDv7(),
				DiceRollType:   core.DiceRollDamage,
				Die:            core.D8,
				Name:           "Divine Smite",
				OriginalRolls:  dmgRolls,
				FinalRolls:     dmgRolls,
				Total:          dmgRollTotal,
				FinalRollValue: dmgRollTotal,
				NumberOfDice:   len(dmgRolls),
				IsCritical:     true,
			}
		} else {
			res, err = c.RollManager.RollDice(numDice*2, core.D8, rollOpts)
			if res != nil {
				res.Name = "Divine Smite"
				res.NumberOfDice = len(res.FinalRolls)
				res.IsCritical = true
			}
		}
	} else {
		res, err = c.RollManager.RollDice(numDice, core.D8, rollOpts)
		if res != nil {
			res.Name = "Divine Smite"
		}
	}

	if err != nil {
		log.Printf("error rolling divine smite damage: %v", err)
		return nil
	}

	// 4. Expend slot
	err = c.SpellCastingManager.ExpendSpellSlot(slotLevel)
	if err != nil {
		log.Printf("error expending spell slot for divine smite: %v", err)
		return nil
	}

	// 5. Log & Return
	c.LogEvent(events.ECombatEventMessage, fmt.Sprintf("Divine Smite! (Level %d slot)", slotLevel))
	c.LogEvent(events.ETRollEvent, res)

	return &core.Effect{
		Type:         core.EffectDamage,
		Value:        res.GetTotal(),
		BaseValue:    res.GetTotal(),
		DamageType:   core.DamageRadiant,
		SourceRollID: res.GetID(),
	}
}
