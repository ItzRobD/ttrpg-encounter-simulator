package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"fmt"
)

func (m *Monster) ProcessTurn(actorID int, turnType core.TurnType) (*core.TurnResult, *core.AIRequest, error) {
	if turnType == core.TurnTypeLegendary {
		return m.processLegendaryTurn(actorID)
	}

	result := &core.TurnResult{
		TurnStatuses: make(map[core.TurnStatus]bool),
	}

	// Able to act
	if m.EntityState.CanTakeActions() {
		aiReq, err := m.GetAIRequest(actorID, core.AIReqNormalAction)
		if err != nil {
			return nil, nil, err
		}
		result.TurnStatuses[core.TurnActionReady] = true
		return result, aiReq, nil
	}

	// Unable to Act
	if m.EntityState.IsDead {
		result.TurnStatuses[core.TurnDead] = true
		return result, nil, nil
	}

	if m.EntityState.GetIsUnconscious() {
		ucResult, err := m.handleUnconsciousTurn(result)
		if ucResult.TurnStatuses[core.TurnRevived] {
			aiReq, err := m.GetAIRequest(actorID, core.AIReqNormalAction)
			if err != nil {
				return nil, nil, fmt.Errorf("error getting AI request: %s", err)
			}
			ucResult.TurnStatuses[core.TurnActionReady] = true
			ucResult.Conditions = nil
			return ucResult, aiReq, nil
		}
		return ucResult, nil, err
	}

	result.Conditions = m.EntityState.GetActiveIncapacitatingConditions()
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
		actionChoice, err = m.AI.chooseMonsterActionType()
		if err != nil {
			return nil, err
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
	default:
		return req, fmt.Errorf("invalid AI request type: %v", t)
	}

	events.LogMonsterActionChoiceEvent(m, req.ActionType, m.EventListener)

	req.ActorID = actorID

	return req, nil
}

func (m *Monster) ExecuteAIRequest(req *core.AIRequest) (*core.ActionOutcome, error) {
	switch req.ActionType {
	case core.ATMonsterAction, core.ATMonsterMultiattack, core.ATMonsterSpecial, core.ATLegendaryAction:
		attackReq, err := m.createAttackRequest(req.Target, req.ActionIndex, req.ActionType, req.Advantage, req.SimOptions)
		if err != nil {
			return nil, err
		}

		results, err := m.ActionManager.ProcessAttackRequest(attackReq)
		if err != nil {
			return nil, err
		}

		// Recharge action
		if m.ActionManager.Actions[req.ActionIndex].RechargeValue > 0 {
			m.EntityState.ExpendRechargeAction(req.ActionIndex)
		}

		// Legendary actions
		if req.ActionType == core.ATLegendaryAction {
			cost := m.ActionManager.LegendaryActions[req.ActionIndex].Cost
			m.EntityState.ExpendLegendaryActionPoints(cost)
		}

		var effects []core.Effect
		for _, res := range results {
			if res.GetIsHit() {
				effects = append(effects, core.Effect{
					Type:       core.EffectDamage,
					Value:      res.GetDamageResult().GetTotal(),
					DamageType: res.GetDamageType(),
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
		scReq, err := m.createSpellCastRequest(req.Target, *req.SpellChoice, req.Advantage, req.SimOptions)
		if err != nil {
			return nil, err
		}

		res, err := m.SpellCastingManager.CastSpell(scReq)
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
	default:
		return nil, fmt.Errorf("monster execute ai req - invalid action type: %s", req.ActionType)
	}
}

func (m *Monster) handleUnconsciousTurn(turnResult *core.TurnResult) (*core.TurnResult, error) {
	// Failsafes if character is already dead and this is called
	if m.EntityState.IsDead {
		turnResult.TurnStatuses[core.TurnDead] = true
		return turnResult, nil
	}

	if !m.EntityState.GetIsUnconscious() {
		return nil, fmt.Errorf("character is not unconscious")
	}

	// Character is not dead but is unconscious
	if m.EntityState.IsStable {
		turnResult.TurnStatuses[core.TurnUnconscious] = true
		return turnResult, nil
	}

	// Roll death saving throw
	res, err := m.RollManager.RollDeathSavingThrow()
	if err != nil {
		return nil, fmt.Errorf("failed to roll death saving throw: %v", err)
	}

	// Apply death saving throw turnResult
	err = m.EntityState.ApplyDeathSavingThrowResult(res)
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

	turnResult.Conditions = m.EntityState.GetActiveIncapacitatingConditions()
	return turnResult, nil
}

func (m *Monster) processLegendaryTurn(actorID int) (*core.TurnResult, *core.AIRequest, error) {
	if !m.IsLegendary || len(m.ActionManager.LegendaryActions) == 0 {
		return nil, nil, fmt.Errorf("monster is not legendary")
	}

	result := &core.TurnResult{
		TurnStatuses: make(map[core.TurnStatus]bool),
	}

	if !m.EntityState.HasLegendaryActionPointsRemaining() {
		result.TurnStatuses[core.TurnLegendaryUnavailable] = true
		return result, nil, nil
	}

	var canAffordLegAction bool
	pointsRemaining := m.EntityState.GetLegendaryActionPoints()
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
