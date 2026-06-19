package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg_old/core/events"
	"dnd5e-encounter-simulator-backend/pkg_old/monster"
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

	// EXPEND ACTION: Weapon attack uses the main action
	actorESM.ExpendAction()

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

	// get actor state manager
	actorESM, ok := aiReq.Actor.GetState().(*entity_state_manager.EntityStateManager)
	if ok && actorESM != nil {
		if aiReq.SpellChoice != nil {
			if aiReq.SpellChoice.GetSpell().GetCastingTime() == "action" {
				actorESM.ExpendAction()
			} else if aiReq.SpellChoice.GetSpell().GetCastingTime() == "bonus action" {
				actorESM.ExpendBonusAction()
			}
		}
	}

	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeHeal(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	// get actor state manager
	actorESM, ok := aiReq.Actor.GetState().(*entity_state_manager.EntityStateManager)
	if ok && actorESM != nil {
		if aiReq.HealRequest.SpellChoice != nil {
			if aiReq.HealRequest.SpellChoice.GetSpell().GetCastingTime() == "action" {
				actorESM.ExpendAction()
			} else if aiReq.HealRequest.SpellChoice.GetSpell().GetCastingTime() == "bonus action" {
				actorESM.ExpendBonusAction()
			}
		} else {
			// Lay on Hands etc
			actorESM.ExpendAction()
		}
	}

	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeMonsterAction(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	// get actor state manager
	actorESM, ok := aiReq.Actor.GetState().(*entity_state_manager.EntityStateManager)
	if ok && actorESM != nil {
		actorESM.ExpendAction()
	}

	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeMonsterMultiattack(aiReq *core.AIRequest) error {
	// New engine-driven stepwise execution for monster multiattacks
	// If actor is not a monster, fall back to legacy path
	m, ok := aiReq.Actor.(*monster.Monster)
	if !ok || m.ActionManager == nil {
		outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
		if err != nil {
			return err
		}
		return ce.processActionResults(aiReq.Actor, outcome)
	}

	// Log a single multiattack anchor and scope subsequent swings under it
	events.LogCombatEventMessage(ce.EventContext, m, fmt.Sprintf("%s (%d) multiattack against %s", m.GetName(), aiReq.Target.GetInstanceID(), aiReq.Target.GetName()), m.GetEventListener())
	if ce.EventContext != nil {
		ce.EventContext.AdvanceScope()
	}

	// EXPEND ACTION: Ensure we don't pick this action again in the next turn loop iteration.
	// We call it after starting the multiattack sequence.
	m.EntityStateManager.ExpendAction()

	// Gather attack blocks for this multiattack option
	blocks := m.ActionManager.GetAttackDataFromIndex(aiReq.ActionIndex, core.ATMonsterMultiattack)
	if len(blocks) == 0 {
		return nil
	}

	currentTarget := aiReq.Target
	excluded := map[int]bool{}

	for i, ad := range blocks {
		// Build single-swing request
		// Compute advantage based on conditions and monster special abilities
		advSlice := make([]core.AdvantageType, 0)
		computedAdv := core.DetermineAttackAdvantageForEntities(m, currentTarget, ad.IsRangedWeapon, core.RollNormal)
		advSlice = append(advSlice, computedAdv)

		// Best effort reflection of createAttackRequest logic inside the engine
		if m.SpecialAbilities.PackTactics {
			if ce.CombatContext != nil && (ce.CombatContext.ConsciousMonsterCount-1) > 0 {
				advSlice = append(advSlice, core.RollAdvantage)
			}
		}
		if m.SpecialAbilities.BloodFrenzy {
			if currentTarget.GetHPStatus().GetHPDifference() > 0 {
				advSlice = append(advSlice, core.RollAdvantage)
			}
		}
		if m.EntityStateManager.GetIsRecklesslyAttacking() {
			advSlice = append(advSlice, core.RollAdvantage)
		}
		// Assassinate: advantage against any creature that hasn't taken a turn
		if m.SpecialAbilities.Assassinate && !currentTarget.GetHasTakenTurnInCombat() {
			advSlice = append(advSlice, core.RollAdvantage)
		}

		adv := core.GetFinalAdvantageType(advSlice)

		atkOpts := core.AttackOptions{
			Advantage:            adv,
			ShouldApplyDamageMod: true,
			ImprovedCritical:     aiReq.SimOptions != nil && aiReq.SimOptions.UseImprovedCritical,
		}
		req := &core.AttackRequest{
			AttackData:        []core.AttackData{ad},
			AttackOptions:     atkOpts,
			SimulationOptions: aiReq.SimOptions,
			Target:            currentTarget,
		}

		// Roll/log one attack and per-component damage rolls
		res, err := m.ActionManager.ProcessSingleAttack(req)
		if err != nil {
			return err
		}

		// Convert to effects and apply immediately
		effects := ce.effectsFromAttackResult(res)
		if err := ce.applyEffects(m, currentTarget, effects); err != nil {
			return err
		}

		// Early victory check
		if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
			return nil
		}

		// Policy check after applying damage
		policy := ce.SimOptions.GetMultiattackPolicy()
		if currentTarget.GetHPStatus().GetHP() <= 0 {
			switch policy {
			case core.MultiattackPolicyRetargetOnDown:
				excluded[currentTarget.GetInstanceID()] = true
				newTarget, status := m.AI.ChooseTargetForFollowUp(excluded)
				if status != core.TargetOK || newTarget == nil {
					if ce.SimOptions.DebugAI {
						fmt.Printf("[DEBUG AI] %s: No more valid targets for follow-up\n", m.GetName())
					}
					return nil // no valid targets; end early
				}
				events.LogCombatEventMessage(ce.EventContext, m, fmt.Sprintf("Target downed after swing %d; retargeting to %s", i+1, newTarget.GetName()), m.GetEventListener())
				currentTarget = newTarget
				// Explicitly update advantage for the next block based on the new target
				// Note: i+1 will use the next iteration's ad.IsRangedWeapon
			case core.MultiattackPolicyWasteRemaining:
				if ce.SimOptions.DebugAI {
					fmt.Printf("[DEBUG AI] %s: Wasting remaining attacks (WasteRemaining policy)\n", m.GetName())
				}
				return nil // stop executing remaining swings
			case core.MultiattackPolicyNone:
				// Continue to swing the same target (even if downed); no-op
			}
		}
	}

	if ce.SimOptions.DebugAI {
		fmt.Printf("[DEBUG AI] %s: Finished multiattack\n", m.GetName())
	}
	return nil
}

// effectsFromAttackResult converts a single AttackResult into effect slices with correct SourceRollID.
func (ce *CombatEngine) effectsFromAttackResult(res core.AttackResult) []core.Effect {
	var effects []core.Effect
	if !res.GetIsHit() {
		return effects
	}
	dmgRes := res.GetDamageResult()
	comps := dmgRes.GetDamageComponents()
	if len(comps) > 0 {
		for _, comp := range comps {
			effects = append(effects, core.Effect{
				Type:           core.EffectDamage,
				Value:          comp.GetTotal(),
				BaseValue:      comp.GetTotal(),
				DamageType:     comp.GetDamageType(),
				ResistBreakers: res.ResistBreakers,
				SourceRollID:   dmgRes.GetID(),
				AttackCtx: &core.AttackContext{
					IsRanged:   res.IsRanged,
					IsCritical: res.IsCriticalHit,
				},
			})
		}
	} else {
		effects = append(effects, core.Effect{
			Type:           core.EffectDamage,
			Value:          dmgRes.GetTotal(),
			BaseValue:      dmgRes.GetTotal(),
			DamageType:     res.GetDamageType(),
			ResistBreakers: res.ResistBreakers,
			SourceRollID:   dmgRes.GetID(),
			AttackCtx: &core.AttackContext{
				IsRanged:   res.IsRanged,
				IsCritical: res.IsCriticalHit,
			},
		})
	}
	return effects
}

// applyEffects applies effects immediately using the same pipeline used by action outcomes.
func (ce *CombatEngine) applyEffects(actor core.Entity, target core.Entity, effects []core.Effect) error {
	if len(effects) == 0 {
		return nil
	}
	outcome := &core.ActionOutcome{
		ActorID:  actor.GetInstanceID(),
		TargetID: target.GetInstanceID(),
		Effects:  effects,
	}
	return ce.processActionResults(actor, outcome)
}

func (ce *CombatEngine) executeMonsterLegendaryAction(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	// Legendary actions don't use the main action, but they expend points.
	// This is already handled in Monster.processLegendaryTurn -> GetAIRequest -> ExecuteAIRequest (implicitly or explicitly)
	// Actually, let's verify where legendary points are spent.

	return ce.processActionResults(aiReq.Actor, outcome)
}
