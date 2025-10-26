package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"errors"
	"fmt"
)

func (c *Character) GetAIRequest(actorID int, t core.AIRequestType) (*core.AIRequest, error) {
	var req *core.AIRequest
	var err error
	switch t {
	case core.AIReqChooseAction:
		req, err = c.AI.createCharacterActionRequest()
		if err != nil {
			return nil, err
		}
	default:
		return req, fmt.Errorf("invalid AI request type: %s", t)
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
		}, nil
	case core.ATHeal:
		return nil, errors.New("not implemented")
	}
	return nil, errors.New("invalid action type")
}
