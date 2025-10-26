package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
)

func (m *Monster) GetAIRequest(actorID int, t core.AIRequestType) (*core.AIRequest, error) {
	var req *core.AIRequest
	var err error
	switch t {
	case core.AIReqChooseAction:
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
		return req, fmt.Errorf("invalid AI request type: %s", t)
	}

	// TODO: Logging

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
		}, nil
	default:
		return nil, fmt.Errorf("monster execute ai req - invalid action type: %s", req.ActionType)
	}
}
