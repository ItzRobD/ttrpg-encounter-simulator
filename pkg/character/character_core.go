package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"errors"
	"fmt"
	"log"
)

func (c *Character) ProcessTurn(actorID int, turnType core.TurnType) (*core.TurnResult, *core.AIRequest, error) {
	if turnType == core.TurnTypeLegendary {
		return nil, nil, fmt.Errorf("invalid turn type for character: %s", turnType)
	}

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
	if c.EntityStateManager.IsDead {
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
			ucResult.TurnStatuses[core.TurnActionReady] = true
			// On revive: clear Unconscious, set Prone to true so ranged/melee modifiers apply correctly
			c.EntityStateManager.SetUnconscious(false)
			c.EntityStateManager.AddCondition(core.ConditionProne)
			ucResult.Conditions = []core.Condition{core.ConditionProne}
			events.LogCombatEventMessage(c, "Revived from 0 HP: now Prone", c.EventListener)
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
		actionChoice, err = c.AI.chooseCharacterActionType()
		if err != nil {
			return nil, err
		}

		switch actionChoice {
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
	events.LogCharacterActionChoiceEvent(c, req.ActionType, c.EventListener)
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
		for _, res := range results {
			if res.GetIsHit() {
				effects = append(effects, core.Effect{
					Type:           core.EffectDamage,
					Value:          res.GetDamageResult().GetTotal(),
					DamageType:     res.GetDamageType(),
					ResistBreakers: res.ResistBreakers,
				})
			}
		}

		return &core.ActionOutcome{
			ActionType: req.ActionType,
			TargetID:   req.TargetID,
			ActorID:    req.ActorID,
			Effects:    effects,
			Success:    len(effects) > 0,
		}, nil
	case core.ATOffhand:
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
		for _, res := range results {
			if res.GetIsHit() {
				effects = append(effects, core.Effect{
					Type:           core.EffectDamage,
					Value:          res.GetDamageResult().GetTotal(),
					DamageType:     res.GetDamageType(),
					ResistBreakers: res.ResistBreakers,
				})
			}
		}

		return &core.ActionOutcome{
			ActionType: req.ActionType,
			TargetID:   req.TargetID,
			ActorID:    req.ActorID,
			Effects:    effects,
			Success:    len(effects) > 0,
		}, nil
	case core.ATSpell:
		scReq, err := c.CreateSpellCastRequest(req.Target, *req.SpellChoice, adv, req.SimOptions)
		if err != nil {
			return nil, err
		}

		res, err := c.SpellCastingManager.CastSpell(scReq)
		if err != nil {
			return nil, err
		}

		var effects []core.Effect
		if res.GetIsHit() {
			if req.SpellChoice.Spell.GetSpellType() == core.STDamage {
				effects = append(effects, core.Effect{
					Type:       core.EffectDamage,
					Value:      res.GetSpellTotalValue(),
					DamageType: res.GetDamageType(),
				})
			} else if req.SpellChoice.Spell.GetSpellType() == core.STHealing {
				effects = append(effects, core.Effect{
					Type:  core.EffectHealing,
					Value: res.GetSpellTotalValue(),
				})
			}
		}

		return &core.ActionOutcome{
			ActionType: req.ActionType,
			TargetID:   req.TargetID,
			ActorID:    req.ActorID,
			Effects:    effects,
			Success:    len(effects) > 0,
		}, nil
	case core.ATHeal:
		return nil, errors.New("not implemented")
	}
	return nil, errors.New("invalid action type")
}

func (c *Character) handleUnconsciousTurn(turnResult *core.TurnResult) (*core.TurnResult, error) {
	// Failsafes if character is already dead and this is called
	if c.EntityStateManager.IsDead {
		turnResult.TurnStatuses[core.TurnDead] = true
		return turnResult, nil
	}

	if !c.EntityStateManager.GetIsUnconscious() {
		return nil, fmt.Errorf("character is not unconscious")
	}

	// Character is not dead but is unconscious
	if c.EntityStateManager.IsStable {
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
