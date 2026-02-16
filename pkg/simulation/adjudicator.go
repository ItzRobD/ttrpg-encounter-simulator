package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"fmt"
	"slices"
	"sort"
)

// Adjudicator is the centralized "Rulebook" that resolves ActionIntents.
type Adjudicator struct {
	ed *EncounterDirector
}

func NewReferee(ed *EncounterDirector) *Adjudicator {
	return &Adjudicator{
		ed: ed,
	}
}

// ResolveAction resolves an action intent.
func (adj *Adjudicator) ResolveAction(a *actor.Actor, intent ActionIntent) error {
	id := adj.ed.LogEvent(events.EventActionStart, a, map[string]interface{}{
		"action_name":     intent.Action.Name,
		"activation_type": intent.ActivationType,
		"target_ids":      intent.TargetIDs,
	})
	adj.ed.EventContext.PushParent(id)
	defer adj.ed.EventContext.PopParent()

	// Track action usage
	switch intent.Action.Cost.ActivationType {
	case core.ActAction:
		a.StateManager.ActionUsedCount++
	case core.ActBonus:
		a.StateManager.BonusActionUsedCount++
	case core.ActLegendary:
		a.StateManager.LegendaryActionUsedCount += intent.Action.Cost.Value
	}

	// Consume recharge
	if intent.Action.RechargeValue > 0 {
		a.StateManager.Resource[intent.Action.Name] = 0
	}
	// Consume spell slots
	if intent.Action.ActionType == core.ATSpell && intent.Action.CastLevel > 0 {
		if intent.Action.IsInnate {
			a.StateManager.InnateCurrent[intent.Action.Name]--
		} else {
			a.StateManager.CurrentSlots[intent.Action.CastLevel]--
		}
	}

	switch intent.Action.ActionType {
	case core.ATMultiAttack:
		return adj.resolveMultiattack(a, intent)
	default:
		return adj.resolveSingleActionSequence(a, intent)
	}
}

func (adj *Adjudicator) resolveSingleActionSequence(a *actor.Actor, intent ActionIntent) error {
	// For AOE actions, we use the pre-identified target list
	if intent.Action.IsAOE && len(intent.TargetIDs) > 0 {
		return adj.resolveAOESequence(a, intent)
	}

	// For single-target actions, intent.TargetIDs should contain the target
	if len(intent.TargetIDs) == 0 {
		return nil
	}
	targetID := intent.TargetIDs[0]

	// For PCs, a single "Action" might involve multiple strikes if they have Extra Attack
	numAttacks := 1
	if a.IsCharacter() && intent.Action.Cost.ActivationType == core.ActAction {
		numAttacks = a.StateManager.AttackCount
		// Failsafe if number of attacks was set improperly
		if numAttacks == 0 {
			numAttacks = 1
		}
	}

	targetType := core.TTDamage
	if intent.Action.AverageHeal > 0 {
		targetType = core.TTHealing
	}

	for i := 0; i < numAttacks; i++ {
		// Parent 3: Resolution Scope
		id := adj.ed.LogEvent(events.EventResolution, a, map[string]interface{}{
			"index": i,
		})
		adj.ed.EventContext.PushParent(id)

		// Create a local copy of the action to avoid persistent state changes from hooks
		localAction := intent.Action

		// Check if target is still alive/valid
		target, ok := adj.ed.Actors[targetID]
		if !ok || (targetType == core.TTDamage && target.StateManager.CurrentHP <= 0) {
			// Try to retarget
			newTargetIDs := adj.ed.AIDirector.SelectTarget(a, adj.ed.GetEnemyTargets(a), targetType, adj.ed)
			if targetType == core.TTHealing {
				_, newTargetIDs = adj.ed.AIDirector.chooseBestHealingAction(a, intent.ActivationType, adj.ed)
			}
			if len(newTargetIDs) == 0 {
				adj.ed.EventContext.PopParent()
				return nil // No more targets
			}
			targetID = newTargetIDs[0]
			target = adj.ed.Actors[targetID]
		}

		var err error
		if targetType == core.TTHealing {
			err = adj.executeHealing(a, targetID, &localAction, adj.ed)
		} else {
			err = adj.executeIndividualStrike(a, target, &localAction)
		}

		adj.ed.EventContext.PopParent()
		if err != nil {
			return err
		}
	}

	return nil
}

func (adj *Adjudicator) resolveAOESequence(a *actor.Actor, intent ActionIntent) error {
	// AOE actions typically hit all targets simultaneously (e.g. Fireball)
	// We'll iterate through each target and resolve the strike
	for i, targetID := range intent.TargetIDs {
		target, ok := adj.ed.Actors[targetID]
		if !ok || target.StateManager.CurrentHP <= 0 {
			continue
		}

		// Resolution Scope per target
		id := adj.ed.LogEvent(events.EventResolution, a, map[string]interface{}{
			"index":  i,
			"is_aoe": true,
		})
		adj.ed.EventContext.PushParent(id)

		// Create a local copy of the action to avoid persistent state changes from hooks
		localAction := intent.Action

		var err error
		if localAction.AverageHeal > 0 {
			err = adj.executeHealing(a, targetID, &localAction, adj.ed)
		} else {
			err = adj.executeIndividualStrike(a, target, &localAction)
		}

		adj.ed.EventContext.PopParent()
		if err != nil {
			return err
		}
	}
	return nil
}

func (adj *Adjudicator) executeHealing(a *actor.Actor, initialTargetID int, action *core.Action, ed *EncounterDirector) error {
	totalHealing := 0
	healingPool := 0
	targetCount := 1
	healingTargetIDs := []int{initialTargetID}

	if action.Name == "Mass Cure Wounds" || action.Name == "Mass Healing Word" {
		targetCount = 6
	}
	if action.Name == "Aid" {
		targetCount = 3
	}
	if action.Name == "Mass Heal" {
		targetCount = len(ed.GetAllyTargets(a))
	}

	id := adj.ed.LogEvent(events.EventOutcome, a, map[string]interface{}{
		"type":           "healing",
		"action":         action.Name,
		"target_count":   targetCount,
		"initial_target": initialTargetID,
	})
	adj.ed.EventContext.PushParent(id)
	defer adj.ed.EventContext.PopParent()

	// Prioritize targets
	priorityLists := [][]int{
		ed.Statistics.NeedsEmergencyHealing,
		ed.Statistics.NeedsHealing,
	}

	if action.Name == "Mass Heal" {
		allAllies := ed.GetAllyTargets(a)
		allyIDs := make([]int, 0, len(allAllies))
		for id := range allAllies {
			allyIDs = append(allyIDs, id)
		}
		priorityLists = append(priorityLists, allyIDs)
	}

	for _, list := range priorityLists {
		sort.Ints(list)
		for _, id := range list {
			if len(healingTargetIDs) >= targetCount {
				break
			}
			if !slices.Contains(healingTargetIDs, id) {
				healingTargetIDs = append(healingTargetIDs, id)
			}
		}
	}

	if action.Name == "Mass Heal" {
		healingPool = 700
	} else {
		for _, block := range action.DiceBlock {
			opts := roll_manager.RollOptions{
				RollType: core.DiceRollHealing,
				Modifier: block.Modifier,
			}

			healRollRes := adj.ed.RollManager.RollDice(block.NumberOfDice, block.Die, opts)
			totalHealing += healRollRes.Total
		}
		if totalHealing == 0 && action.AverageHeal > 0 {
			totalHealing = action.AverageHeal
		}
	}

	for _, id := range healingTargetIDs {
		target, ok := ed.Actors[id]
		if !ok {
			continue
		}

		if action.Name == "Heal" || action.Name == "Mass Heal" {
			target.StateManager.Conditions.Remove(core.ConditionBlinded)
			target.StateManager.Conditions.Remove(core.ConditionDeafened)
		}
		if action.Name == "Aid" {
			target.StateManager.MaxHP += totalHealing
		}
		if action.Name == "Regenerate" {
			target.Features = append(target.Features, core.Feature{
				Name: core.SpecAbilityRegeneration,
				Hooks: map[core.HookType]bool{
					core.HookOnTurnStart: true,
				},
				Data: core.FeatureData{
					Value: 1,
				},
			})
		}

		actualHeal := totalHealing
		if action.Name == "Mass Heal" {
			missingHP := target.StateManager.MaxHP - target.StateManager.CurrentHP
			if missingHP > healingPool {
				actualHeal = healingPool
			} else {
				actualHeal = missingHP
			}
			healingPool -= actualHeal
		}

		hpRes := target.StateManager.ModifyHP(actualHeal, false, target.IsCharacter())
		adj.ed.LogEvent(events.EventHPModified, target, map[string]interface{}{
			"result":    hpRes,
			"healer_id": a.InstanceID,
		})

		// Update Statistics
		if adj.ed.Statistics != nil {
			adj.ed.Statistics.AddHeal(a.InstanceID, target.InstanceID, actualHeal)
			// Guard if legendary actions can be healing actions (custom monsters perhaps)
			if action.Cost.ActivationType == core.ActLegendary {
				adj.ed.Statistics.AddLegendaryActionUse(a.InstanceID)
			}

			// Check if we can clear NeedsHealing
			threshold := adj.ed.SimOptions.MonsterHealThresholdPct
			emergencyThreshold := adj.ed.SimOptions.MonsterEmergencyThresholdPct
			if target.IsCharacter() {
				threshold = adj.ed.SimOptions.CharacterHealThresholdPct
				emergencyThreshold = adj.ed.SimOptions.CharacterEmergencyThresholdPct
			}

			currentPct := (target.StateManager.CurrentHP * 100) / target.StateManager.MaxHP
			if currentPct > threshold || target.StateManager.CurrentHP == target.StateManager.MaxHP {
				adj.ed.Statistics.ClearNeedsHealing(target.InstanceID)
			}
			if currentPct > emergencyThreshold || target.StateManager.CurrentHP == target.StateManager.MaxHP {
				adj.ed.Statistics.ClearNeedsEmergencyHealing(target.InstanceID)
			}
		}
	}
	return nil
}

func (adj *Adjudicator) resolveMultiattack(a *actor.Actor, intent ActionIntent) error {
	if len(intent.TargetIDs) == 0 {
		return nil
	}
	targetID := intent.TargetIDs[0]

	for _, ma := range intent.Action.Multiattack {
		// Find the specific action by ID
		// In a real implementation, we'd have a map or lookup for actions
		var subAction *core.Action
		for i := range a.Actions {
			if a.Actions[i].ID == ma.ActionID {
				subAction = &a.Actions[i]
				break
			}
		}

		if subAction == nil {
			continue
		}

		for i := 0; i < ma.Count; i++ {
			// Parent 3: Resolution Scope
			pid := adj.ed.LogEvent(events.EventResolution, a, map[string]interface{}{
				"multiattack": true,
				"action":      subAction.Name,
				"index":       i,
			})
			adj.ed.EventContext.PushParent(pid)

			// Check if target is still valid
			target, ok := adj.ed.Actors[targetID]
			if !ok || target.StateManager.CurrentHP <= 0 {
				// Retargeting logic based on policy
				if adj.ed.SimOptions.MultiattackPolicy == core.MultiattackPolicyRetargetOnDown {
					newTargetIDs := adj.ed.AIDirector.SelectTarget(a, adj.ed.GetEnemyTargets(a), core.TTDamage, adj.ed)
					if len(newTargetIDs) == 0 {
						adj.ed.EventContext.PopParent()
						return nil
					}
					targetID = newTargetIDs[0]
					target = adj.ed.Actors[targetID]
				} else {
					adj.ed.EventContext.PopParent()
					return nil // Waste remaining swings
				}
			}

			err := adj.executeIndividualStrike(a, target, subAction)
			adj.ed.EventContext.PopParent()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (adj *Adjudicator) executeIndividualStrike(a *actor.Actor, target *actor.Actor, action *core.Action) error {
	// Dispatch "On Attack" Hooks
	ctx := &FeatureContext{
		Target: target,
		AttackContext: &AttackContext{
			Action:           action,
			WeaponModifiers:  action.WeaponModifiers,
			WeaponProperties: action.WeaponProperties,
		},
	}

	// isHit indicates if the attack landed or if the effect (for DC actions) triggered
	var isHit bool
	// saveSuccess indicates if the target succeeded on their saving throw (only if action.HasDC)
	var saveSuccess bool
	// isCritical indicates if the attack was a critical hit; if saving throw this is false
	var isCritical bool

	if action.HasDC {
		success, rollRes := adj.ResolveSavingThrow(a, target, action)
		saveSuccess = success
		isHit = true // A DC action "hits" regardless, but damage is mitigated in resolveDamage
		isCritical = false

		adj.ed.LogEvent(events.EventSavingThrow, target, map[string]interface{}{
			"action":       action.Name,
			"dc":           action.DCSaveDC,
			"ability":      action.DCAbility,
			"save_success": saveSuccess,
			"attacker_id":  a.InstanceID,
			"roll":         rollRes,
		})
	} else if action.IsAutoHit {
		isHit = true
		isCritical = false
	} else {
		opts := roll_manager.RollOptions{
			RollType:  core.DiceRollAttack,
			Modifier:  action.AttackBonus,
			Advantage: adj.computeAttackAdvantage(a, target),
		}

		ctx.AttackContext.AttackRoll = &opts
		ctx.AttackContext.WeaponModifiers = action.WeaponModifiers
		adj.ed.dispatchHooks(a, core.HookOnSelfAttack, ctx)

		// Dispatch to the Target (to check for Elusive, etc.)
		targetCtx := &FeatureContext{
			Target: a,
			AttackContext: &AttackContext{
				AttackRoll: &opts,
				Action:     action,
			},
		}
		adj.ed.dispatchHooks(target, core.HookOnSelfAttack, targetCtx)

		rollRes := adj.ed.RollManager.RollD20(opts)
		if rollRes == nil {
			return fmt.Errorf("failed to roll attack for %s; action: %s; id: %s	", a.Name, action.Name, action.ID)
		}

		// Compare vs AC
		isHit = rollRes.Total >= target.AC
		if rollRes.IsCritical {
			isHit = true
			isCritical = true
		}
		if rollRes.IsNaturalOne {
			isHit = false
		}
		adj.ed.LogEvent(events.EventAttackRoll, a, map[string]interface{}{
			"action":      action.Name,
			"roll":        rollRes,
			"is_hit":      isHit,
			"is_critical": isCritical,
			"target_id":   target.InstanceID,
			"target_ac":   target.AC,
		})
	}

	// Update Statistics
	// Adds a missed attack. Successful hits and damage are added within resolveDamage
	if adj.ed.Statistics != nil && !isHit {
		adj.ed.Statistics.AddAttack(a.InstanceID, target.InstanceID, false, isCritical, 0)
	}
	if action.Cost.ActivationType == core.ActLegendary {
		adj.ed.Statistics.AddLegendaryActionUse(a.InstanceID)
	}
	if action.ActionType == core.ATSpell {
		adj.ed.Statistics.AddSpellAttack(a.InstanceID, action.HasDC)
	}

	if !isHit {
		return nil
	}

	// Dispatch "On Hit" Hooks
	ctx.AttackContext.IsCritical = &isCritical
	adj.ed.dispatchHooks(a, core.HookOnSelfHit, ctx)

	// Dispatch to the hook owner, target, and set the ctx target to the attacker
	ctx.Target = a
	adj.ed.dispatchHooks(target, core.HookOnSelfHit, ctx)

	// Resolve Damage
	oid := adj.ed.LogEvent(events.EventOutcome, a, map[string]interface{}{
		"type":      "damage",
		"target_id": target.InstanceID,
	})
	adj.ed.EventContext.PushParent(oid)
	err := adj.resolveDamage(a, target, action, saveSuccess, isCritical)
	adj.ed.EventContext.PopParent()
	return err
}

// ResolveSavingThrow calculates whether the target succeeds in a saving throw against the given action's difficulty check.
func (adj *Adjudicator) ResolveSavingThrow(a *actor.Actor, target *actor.Actor, action *core.Action) (bool, *roll_manager.RollResult) {
	modifier := target.Abilities.GetAbilityModifier(action.DCAbility)
	if target.Abilities.GetIsProficientInAbility(action.DCAbility) {
		modifier += target.ProficiencyBonus
	}

	opts := roll_manager.RollOptions{
		RollType:  core.DiceRollSavingThrow,
		Advantage: adj.computeSavingThrowAdvantage(target, action),
		Modifier:  modifier,
	}

	ctx := &FeatureContext{
		Target: target,
		SaveContext: &SaveContext{
			Target:  target,
			Options: &opts,
		},
		AttackContext: &AttackContext{
			Action: action,
		},
	}

	// Dispatch to the Target BEFORE roll to allow for advantage/disadvantage modifications
	adj.ed.dispatchHooks(target, core.HookOnSelfSavingThrow, ctx)
	if action.ActionType == core.ATSpell {
		adj.ed.dispatchHooks(target, core.HookOnSelfSavingThrowAgainstMagic, ctx)
	}

	rollRes := adj.ed.RollManager.RollD20(opts)
	if rollRes == nil {
		fmt.Printf("Error: failed to roll saving throw for %s; action: %s; id: %s\n", target.Name, action.Name, action.ID)
		return false, nil
	}

	saveSuccess := rollRes.Total >= action.DCSaveDC
	ctx.SaveContext.SaveSuccess = saveSuccess

	// Dispatch again to allow for post-roll modifications (like Indomitable or Legendary Resistance)
	ctx.SaveContext.IsPostRoll = true
	adj.ed.dispatchHooks(target, core.HookOnSelfSavingThrow, ctx)
	saveSuccess = ctx.SaveContext.SaveSuccess

	return saveSuccess, rollRes
}

func (adj *Adjudicator) computeAttackAdvantage(a *actor.Actor, target *actor.Actor) core.AdvantageType {
	opts := adj.computeAttackAdvantageOptions(a, target)
	return opts.Advantage
}

func (adj *Adjudicator) computeAttackAdvantageOptions(a *actor.Actor, target *actor.Actor) roll_manager.RollOptions {
	opts := roll_manager.RollOptions{
		Advantage: core.RollNormal,
	}

	// 1. Attacker Conditions
	if a.StateManager.Conditions.Has(core.ConditionBlinded) {
		opts.DisadvantageCount++
	}
	if a.StateManager.Conditions.Has(core.ConditionReckless) {
		opts.AdvantageCount++
	}

	// 2. Target Conditions
	if target.StateManager.Conditions.Has(core.ConditionBlinded) {
		opts.AdvantageCount++
	}
	if target.StateManager.Conditions.Has(core.ConditionParalyzed) ||
		target.StateManager.Conditions.Has(core.ConditionStunned) ||
		target.StateManager.Conditions.Has(core.ConditionUnconscious) ||
		target.StateManager.Conditions.Has(core.ConditionRestrained) ||
		target.StateManager.Conditions.Has(core.ConditionProne) {
		opts.AdvantageCount++
	}
	if target.StateManager.Conditions.Has(core.ConditionReckless) {
		opts.AdvantageCount++
	}

	opts.Advantage = opts.CalculateAdvantage()
	return opts
}

func (adj *Adjudicator) computeSavingThrowAdvantage(target *actor.Actor, action *core.Action) core.AdvantageType {
	// Evaluates conditions and features to determine saving throw advantage.
	return core.RollNormal
}

func (adj *Adjudicator) resolveDamage(a *actor.Actor, target *actor.Actor, action *core.Action, saveSuccess bool, isCrit bool) error {
	totalDamage := 0

	for _, block := range action.DiceBlock {
		opts := roll_manager.RollOptions{
			RollType: core.DiceRollDamage,
			Modifier: block.Modifier,
		}

		var dmgRollRes *roll_manager.RollResult
		if isCrit {
			if adj.ed.SimOptions.UseImprovedCritical {
				dmgRollRes = adj.ed.RollManager.RollExtraMaxDice(block.NumberOfDice, block.Die, opts)
			} else {
				dmgRollRes = adj.ed.RollManager.RollDice(block.NumberOfDice*2, block.Die, opts)
			}
		} else {
			dmgRollRes = adj.ed.RollManager.RollDice(block.NumberOfDice, block.Die, opts)
		}

		if dmgRollRes == nil {
			return fmt.Errorf("failed to roll damage for %s", action.Name)
		}

		// Log the damage roll first so modifications are parented to it
		dmgRollID := adj.ed.LogEvent(events.EventDamageRoll, a, map[string]interface{}{
			"roll":        dmgRollRes,
			"damage_type": block.DamageType,
			"target_id":   target.InstanceID,
		})
		adj.ed.EventContext.PushParent(dmgRollID)

		// Modify damage based on resistances and saving throws
		dmgValue := dmgRollRes.Total

		if action.HasDC {
			// Create a local copy of the action to avoid persistent state changes from hooks
			localAction := *action

			fCtx := &FeatureContext{
				Target: target,
				SaveContext: &SaveContext{
					Target:      target,
					SaveSuccess: saveSuccess,
				},
				AttackContext: &AttackContext{
					Action: &localAction,
				},
				DamageContext: &DamageContext{
					DamageValue: &dmgValue,
					DamageType:  block.DamageType,
				},
			}
			fCtx.SaveContext.IsPostRoll = true

			adj.ed.dispatchHooks(target, core.HookOnSelfSavingThrow, fCtx)

			if fCtx.SaveContext.SaveSuccess {
				if fCtx.AttackContext.Action.DCOnSuccess == core.DCOnSuccessHalf {
					dmgValue /= 2
				} else if fCtx.AttackContext.Action.DCOnSuccess == core.DCOnSuccessNone {
					dmgValue = 0
				}
			}
		}

		// Ensure we dispatch HookOnSelfSavingThrow even if we don't have a DC?
		// No, the above block handles it when there IS a DC.

		adj.applyResistancesToDamage(target, &dmgValue, block.DamageType, target.GetResistances())

		// Dispatch "On Damage Taken" hooks for each block
		// This allows for things like Absorption to negate the damage and heal the actor
		dmgCtx := &FeatureContext{
			Target: a, // The attacker
			DamageContext: &DamageContext{
				DamageValue: &dmgValue,
				DamageType:  block.DamageType,
			},
			AttackContext: &AttackContext{
				Action: action,
			},
		}
		adj.ed.dispatchHooks(target, core.HookOnSelfDamageTaken, dmgCtx)

		// Check if absorption occurred
		if dmgCtx.DamageContext.IsAbsorption {
			// Heal the actor for the absorbed amount
			hpRes := target.StateManager.ModifyHP(dmgCtx.DamageContext.AbsorbedValue, false, target.IsCharacter())
			adj.ed.LogEvent(events.EventHPModified, target, map[string]interface{}{
				"result": hpRes,
				"note":   "Absorbed from " + string(block.DamageType),
			})
			// DamageValue was already set to 0 by the handler (if it followed the rule),
			// but we use 0 damage for the Statistics.
			dmgValue = 0
		}

		totalDamage += dmgValue

		// We could log a "final damage" event here if we wanted to show the post-reduction value
		// but for now we just close the scope.
		adj.ed.EventContext.PopParent()
	}

	// Apply Damage to target
	oldState := target.StateManager.HealthState
	hpRes := target.StateManager.ModifyHP(-totalDamage, false, target.IsCharacter())
	adj.ed.LogEvent(events.EventHPModified, target, map[string]interface{}{
		"result":      hpRes,
		"attacker_id": a.InstanceID,
	})

	// 5a. Damage while at 0 HP
	if target.StateManager.CurrentHP == 0 && target.IsCharacter() && oldState != core.HealthStateDead {
		target.StateManager.Conditions.Remove(core.ConditionStable)
		if isCrit {
			target.StateManager.DeathSaveFailures += 2
		} else {
			target.StateManager.DeathSaveFailures++
		}
		// Log death save failure due to damage
		if adj.ed.Statistics != nil {
			adj.ed.Statistics.DeathSave(target.InstanceID, false)
			if isCrit {
				adj.ed.Statistics.DeathSave(target.InstanceID, false)
			}
		}
		// Re-evaluate health state (might have died from failures)
		target.StateManager.HealthState = target.StateManager.GetHealthState(true)
	}

	// 5b. Falling to 0 HP adds Unconscious condition
	if target.StateManager.CurrentHP == 0 && target.StateManager.HealthState != core.HealthStateDead {
		if !target.StateManager.Conditions.Has(core.ConditionUnconscious) {
			adj.ed.LogEvent(events.EventUnconscious, target, nil)
		}
		target.StateManager.Conditions.Add(core.ConditionUnconscious)
		target.StateManager.Conditions.Add(core.ConditionProne)
	}

	// Dispatch death hook if the actor just died
	if oldState != core.HealthStateDead && target.StateManager.HealthState == core.HealthStateDead {
		adj.ed.LogEvent(events.EventDeath, target, nil)
		deathCtx := &FeatureContext{
			Target: a, // The killer
			AttackContext: &AttackContext{
				Action: action,
			},
		}
		adj.ed.dispatchHooks(target, core.HookOnSelfDeath, deathCtx)
	}

	// Dispatch damage taken hook for the entire action resolution (legacy support/general awareness)
	// We already dispatched per block for Absorption logic above.
	strikeCtx := &FeatureContext{
		Target: a, // The original attacker
		AttackContext: &AttackContext{
			Action: action,
		},
		DamageContext: &DamageContext{
			DamageValue: &totalDamage,
		},
	}
	adj.ed.dispatchHooks(target, core.HookOnSelfDamageTaken, strikeCtx)

	// Update Statistics
	if adj.ed.Statistics != nil {
		adj.ed.Statistics.AddAttack(a.InstanceID, target.InstanceID, true, isCrit, totalDamage)

		// Check Healing Threshold
		adj.UpdateHealTargetsAfterDamage(target)
	}

	return nil
}

func (adj *Adjudicator) UpdateHealTargetsAfterDamage(target *actor.Actor) {
	if adj.ed.SimOptions == nil {
		return
	}
	threshold := adj.ed.SimOptions.MonsterHealThresholdPct
	emergencyThreshold := adj.ed.SimOptions.MonsterEmergencyThresholdPct
	if target.IsCharacter() {
		threshold = adj.ed.SimOptions.CharacterHealThresholdPct
		emergencyThreshold = adj.ed.SimOptions.CharacterEmergencyThresholdPct
	}

	currentPct := (target.StateManager.CurrentHP * 100) / target.StateManager.MaxHP
	if threshold > 0 {
		if currentPct <= threshold && target.StateManager.CurrentHP > 0 {
			adj.ed.Statistics.MarkNeedsHealing(target.InstanceID)
		} else {
			adj.ed.Statistics.ClearNeedsHealing(target.InstanceID)
		}
	}
	if emergencyThreshold > 0 {
		if currentPct <= emergencyThreshold && target.StateManager.CurrentHP > 0 {
			adj.ed.Statistics.MarkNeedsEmergencyHealing(target.InstanceID)
		} else {
			adj.ed.Statistics.ClearNeedsEmergencyHealing(target.InstanceID)
		}
	}
}

// applyResistancesToDamage adjusts the damage value based on the specified damage type and associated resistances.
func (adj *Adjudicator) applyResistancesToDamage(target *actor.Actor, dmg *int, dt core.DamageType, resistances core.DamageResistances) {
	if resistances == nil {
		return
	}

	oldValue := *dmg
	resType := resistances.GetResistanceType(dt)

	switch resType {
	case core.ResistanceResistant:
		*dmg /= 2
	case core.ResistanceImmune:
		*dmg = 0
	case core.ResistanceVulnerable:
		*dmg *= 2
	}

	if resType != core.ResistanceNone {
		adj.ed.LogEvent(events.EventDamageModified, target, map[string]interface{}{
			"original_value":  oldValue,
			"final_value":     *dmg,
			"damage_type":     dt,
			"resistance_type": resType,
		})
	}
}

// IdentifyAOETargets determines which actors are affected by an AOE attack, based on input parameters and state.
func (adj *Adjudicator) IdentifyAOETargets(source *actor.Actor, initialTargetID int, includeAllies bool) []int {
	var targetIDs []int

	// If AOE hits all is disabled only select the initial target
	if !adj.ed.SimOptions.AOEHitsAllEnemies && initialTargetID > 0 {
		target, ok := adj.ed.Actors[initialTargetID]
		if ok && target.StateManager.GetHealthState(target.IsCharacter()) != core.HealthStateDead {
			targetIDs = append(targetIDs, initialTargetID)
		}
		return targetIDs
	}

	for id, other := range adj.ed.Actors {
		if id == source.InstanceID || other.StateManager.GetHealthState(other.IsCharacter()) == core.HealthStateDead || other.ActorType == core.ActorTypeLair {
			continue
		}

		shouldHit := false
		if other.Side != source.Side {
			if adj.ed.SimOptions.AOEHitsAllEnemies {
				shouldHit = true
			} else {
				// Fallback: if no initial target provided, select the first available enemy
				if len(targetIDs) == 0 {
					shouldHit = true
				}
			}
		} else if includeAllies {
			shouldHit = true
		}

		if shouldHit {
			targetIDs = append(targetIDs, id)
		}
	}
	sort.Ints(targetIDs) // Sort list to maintain determinism
	return targetIDs
}
