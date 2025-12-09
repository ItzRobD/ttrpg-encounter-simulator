package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"errors"
	"fmt"
)

func (c *Character) ProcessTurn(actorID int, turnType core.TurnType) (*core.TurnResult, *core.AIRequest, error) {
	if turnType == core.TurnTypeLegendary {
		return nil, nil, fmt.Errorf("invalid turn type for character: %s", turnType)
	}

	result := &core.TurnResult{
		TurnStatuses: make(map[core.TurnStatus]bool),
	}

	// Able to act
	if c.EntityState.CanTakeActions() {
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
	if c.EntityState.IsDead {
		result.TurnStatuses[core.TurnDead] = true
		return result, nil, nil
	}

	if c.EntityState.GetIsUnconscious() {
		ucResult, err := c.handleUnconsciousTurn(result)
		if ucResult.TurnStatuses[core.TurnRevived] {
			aiReq, err := c.GetAIRequest(actorID, core.AIReqNormalAction)
			if err != nil {
				return nil, nil, fmt.Errorf("error getting AI request: %s", err)
			}
			ucResult.TurnStatuses[core.TurnActionReady] = true
			// On revive: clear Unconscious, set Prone to true so ranged/melee modifiers apply correctly
			c.EntityState.SetUnconscious(false)
			c.EntityState.AddCondition(core.ConditionProne)
			ucResult.Conditions = []core.Condition{core.ConditionProne}
			events.LogCombatEventMessage(c, "Revived from 0 HP: now Prone", c.EventListener)
			return ucResult, aiReq, nil
		}
		return ucResult, nil, err
	}

	result.Conditions = c.EntityState.GetActiveIncapacitatingConditions()
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
		if err != nil {
			return nil, err
		}
	default:
		return req, fmt.Errorf("invalid AI request type: %v", t)
	}

	if req == nil {
		return nil, nil
	}
	events.LogCharacterActionChoiceEvent(c, req.ActionType, c.EventListener)
	req.ActorID = actorID
	return req, nil
}

func (c *Character) ExecuteAIRequest(req *core.AIRequest) (*core.ActionOutcome, error) {
	switch req.ActionType {
	case core.ATMelee, core.ATRanged:
		attackReq, err := c.CreateAttackRequest(req.Target, req.WeaponSlot, req.Advantage, req.UseVersatile, req.SimOptions)
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
		scReq, err := c.CreateSpellCastRequest(req.Target, *req.SpellChoice, req.Advantage, req.SimOptions)
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
	if c.EntityState.IsDead {
		turnResult.TurnStatuses[core.TurnDead] = true
		return turnResult, nil
	}

	if !c.EntityState.GetIsUnconscious() {
		return nil, fmt.Errorf("character is not unconscious")
	}

	// Character is not dead but is unconscious
	if c.EntityState.IsStable {
		turnResult.TurnStatuses[core.TurnUnconscious] = true
		return turnResult, nil
	}

	// Roll death saving throw
	res, err := c.RollManager.RollDeathSavingThrow()
	if err != nil {
		return nil, fmt.Errorf("failed to roll death saving throw: %v", err)
	}

	// Apply death saving throw turnResult
	err = c.EntityState.ApplyDeathSavingThrowResult(res)
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

	turnResult.Conditions = c.EntityState.GetActiveIncapacitatingConditions()
	return turnResult, nil
}
