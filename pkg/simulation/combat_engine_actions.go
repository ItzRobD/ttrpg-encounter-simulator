package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"fmt"
)

func (ce *CombatEngine) ProcessAIRequest(req *core.AIRequest) error {
	ce.attachOptionsToAIRequest(req)
	switch req.ActionType {
	case core.ATMelee, core.ATRanged:
		return ce.executeWeaponAttack(req)
	case core.ATSpell:
		return ce.executeSpellCast(req)
	case core.ATHeal, core.ATMonsterHeal:
		return ce.executeHeal(req)
		//case core.ATUnarmed:
		//	return ce.executeUnarmedAttack(req)
	case core.ATDragonbornBreathWeapon:
		outcome, err := req.Actor.ExecuteAIRequest(req)
		if err != nil {
			return err
		}
		return ce.processActionResults(req.Actor, outcome)
	case core.ATMonsterAction:
		return ce.executeMonsterAction(req)
	case core.ATMonsterMultiattack:
		return ce.executeMonsterMultiattack(req)
	case core.ATLegendaryAction:
		return ce.executeMonsterLegendaryAction(req)
	case core.ATLairAction:
		// Lair actions are executed by the Lair entity but follow the same
		// generic "actor executes request, engine processes effects" path.
		outcome, err := req.Actor.ExecuteAIRequest(req)
		if err != nil {
			return err
		}
		return ce.processActionResults(req.Actor, outcome)
	default:
		return fmt.Errorf("unknown action type: %v", req.ActionType)
	}

}

func (ce *CombatEngine) attachOptionsToAIRequest(aiReq *core.AIRequest) {
	aiReq.SimOptions = ce.SimOptions
}

func (ce *CombatEngine) executeWeaponAttack(aiReq *core.AIRequest) error {
	// If weapon slot is not specified, use primary slot
	if aiReq.WeaponSlot == "" {
		aiReq.WeaponSlot = core.WSPrimary
	}

	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	actionErr := ce.processActionResults(aiReq.Actor, outcome)
	if actionErr != nil {
		return actionErr
	}

	// get actor state manager
	actorESM, ok := aiReq.Actor.GetState().(*entity_state_manager.EntityStateManager)
	if !ok || actorESM == nil {
		return fmt.Errorf("actor state manager is nil or wrong type")
	}

	if !actorESM.GetHasUsedBonusAction() {
		// TODO: This is a new action id for the event context
		offhandReq, ohErr := aiReq.Actor.GetAIRequest(aiReq.ActorID, core.AIReqOffhandAttack)
		if ohErr != nil {
			return ohErr
		}

		if offhandReq != nil {
			ohOutcome, ohOutcomeErr := aiReq.Actor.ExecuteAIRequest(offhandReq)
			if ohOutcomeErr != nil {
				return ohOutcomeErr
			}
			if ohResError := ce.processActionResults(aiReq.Actor, ohOutcome); ohResError != nil {
				return ohResError
			}
			actorESM.ExpendBonusAction()
		}
	}

	return nil
}

func (ce *CombatEngine) executeSpellCast(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}
	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeHeal(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}
	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeMonsterAction(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeMonsterMultiattack(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeMonsterLegendaryAction(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	return ce.processActionResults(aiReq.Actor, outcome)
}
