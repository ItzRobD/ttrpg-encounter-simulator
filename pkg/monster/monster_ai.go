package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"fmt"
	"math/rand/v2"
	"sort"
)

type MonsterAIMode int

const (
	MAISimple MonsterAIMode = iota
	MAITactical
)

type MonsterAI struct {
	parent             *Monster
	combatCtx          *core.CombatContext
	eventCtx           *core.EventContext
	Weights            *core.UtilityWeights
	rng                *rand.Rand
	isLegendary        bool
	hasRechargeActions bool
	hasMultiattack     bool
	hasSpecialAbility  bool
	AIMode             MonsterAIMode
	PrioritizeTotalAvg bool
}

func NewMonsterAI(m *Monster, weights *core.UtilityWeights) *MonsterAI {
	hasRecharge := false
	hasMultiattack := false
	hasSpecialAbility := false

	if m.ActionManager.SpecialAbilities != nil {
		hasSpecialAbility = true
	}
	if m.ActionManager.Multiattacks != nil {
		hasMultiattack = true
	}
	if m.ActionManager.RechargeActions != nil {
		hasRecharge = true
	}

	return &MonsterAI{
		parent:             m,
		combatCtx:          nil,
		eventCtx:           nil,
		Weights:            weights,
		rng:                m.GetRNG(),
		isLegendary:        m.IsLegendary,
		hasRechargeActions: hasRecharge,
		hasMultiattack:     hasMultiattack,
		hasSpecialAbility:  hasSpecialAbility,
	}
}

func (mai *MonsterAI) UpdateCombatContext(ctx *core.CombatContext) {
	mai.combatCtx = ctx
}

func (mai *MonsterAI) UpdateEventContext(ctx *core.EventContext) {
	mai.eventCtx = ctx
}

func (mai *MonsterAI) GetCombatContext() *core.CombatContext {
	return mai.combatCtx
}

func (mai *MonsterAI) createMonsterLegendaryActionRequest() (*core.AIRequest, error) {
	if !mai.isLegendary {
		return nil, fmt.Errorf("monster is not legendary")
	}

	if mai.parent.EntityStateManager.GetLegendaryActionPoints() == 0 {
		return nil, fmt.Errorf("monster has no legendary action points")
	}

	tStatus, targetID, _, _, err := mai.SelectTargetIDWithLogging(core.TTDamage, false)
	if err != nil {
		return nil, err
	}
	if tStatus == core.TargetNone {
		events.LogCombatEventMessage(mai.parent, "No valid targets", mai.parent.GetEventListener())
		return nil, nil
	}

	// Now that we've decided to use a legendary action, we can log the target choice
	if combatant, ok := mai.combatCtx.CombatantInfo[targetID]; ok {
		events.LogTargetChoiceEvent(mai.parent.GetCurrentEventContext(), mai.parent, combatant.Combatant.GetEntity(), 1.0, nil, mai.parent.GetEventListener())
	}

	actionChoiceID := -1
	legendaryIndexes := mai.getAvailableLegendaryActions(mai.parent.EntityStateManager.GetLegendaryActionPoints())
	if len(legendaryIndexes) > 0 {
		actionChoiceID, err = mai.chooseLegendaryAction(legendaryIndexes)
	}

	if actionChoiceID != -1 {
		return mai.buildAIRequest(actionChoiceID, targetID, nil, core.ATLegendaryAction)
	} else {
		return nil, err
	}
}

func (mai *MonsterAI) getAvailableLegendaryActions(legPointsRemaining int) []int {
	availableIdx := make([]int, 0, len(mai.parent.ActionManager.LegendaryActions))
	for idx, la := range mai.parent.ActionManager.LegendaryActions {
		if la.Cost <= legPointsRemaining {
			availableIdx = append(availableIdx, idx)
		}
	}
	sort.Ints(availableIdx)
	return availableIdx
}

func (mai *MonsterAI) chooseLegendaryAction(indexes []int) (int, error) {
	if len(indexes) == 0 {
		return -1, fmt.Errorf("no legendary actions available")
	}
	bestIndex := -1
	bestAvg := 0

	for _, idx := range indexes {
		if bestIndex == -1 {
			bestIndex = idx
			bestAvg = mai.parent.ActionManager.LegendaryAttackData[idx].Average
			continue
		}

		idxAvg := mai.parent.ActionManager.LegendaryAttackData[idx].Average
		if idxAvg > bestAvg {
			bestIndex = idx
			bestAvg = idxAvg
		}
	}

	if bestIndex == -1 {
		return -1, fmt.Errorf("no legendary actions available")
	}

	return bestIndex, nil
}

func (mai *MonsterAI) createMonsterHealActionRequest() (*core.AIRequest, error) {
	if mai.parent.EntityStateManager.GetHasUsedAction() {
		return nil, nil
	}
	if !mai.parent.IsHealer() {
		return nil, nil
	}

	tStatus, targetID, _, _, err := mai.SelectTargetID(core.TTHealing)
	if err != nil {
		return nil, err
	}
	if tStatus == core.TargetNone {
		events.LogCombatEventMessage(mai.parent, "No valid healing targets", mai.parent.GetEventListener())
		return nil, nil
	}

	target := mai.combatCtx.CombatantInfo[targetID].Combatant.Entity
	healReq, err := mai.parent.CreateHealRequest(target)
	if err != nil {
		return nil, err
	}

	if healReq.Source == core.HealSourceSpell {
		// Log spell choice event
		events.LogSpellChoiceEvent(mai.parent.GetCurrentEventContext(), mai.parent, healReq.SpellChoice, mai.parent.SpellCastingManager.GetStatus(), mai.parent.GetEventListener())
	}

	return &core.AIRequest{
		Actor:       mai.parent,
		ActorType:   core.EntityMonster,
		Target:      target,
		TargetID:    targetID,
		ActionType:  core.ATMonsterHeal,
		HealRequest: healReq,
	}, nil
}

func (mai *MonsterAI) createMonsterDamageActionRequest() (*core.AIRequest, error) {
	tStatus, targetID, _, _, err := mai.SelectTargetID(core.TTDamage)
	if err != nil {
		return nil, err
	}
	if tStatus == core.TargetNone {
		events.LogCombatEventMessage(mai.parent, "No valid targets", mai.parent.GetEventListener())
		return nil, nil
	}
	target := mai.combatCtx.CombatantInfo[targetID].Combatant.GetEntity()

	var actionChoiceID int
	var spellChoice *core.SpellChoice

	// Divine Eminence activation (Bonus Action)
	if mai.parent.SpecialAbilities.DivineEminenceNumDice > 0 &&
		!mai.parent.EntityStateManager.GetIsDivineEminenceActive() &&
		!mai.parent.EntityStateManager.GetHasUsedBonusAction() &&
		mai.parent.SpellCastingManager != nil &&
		mai.parent.SpellCastingManager.HasAnySpellSlots() {

		return &core.AIRequest{
			Actor:      mai.parent,
			ActorType:  core.EntityMonster,
			ActionType: core.ATMonsterDivineEminence,
		}, nil
	}

	// standard action check
	if mai.parent.EntityStateManager.GetHasUsedAction() {
		return nil, nil
	}

	// Use recharge actions first
	if mai.hasRechargeActions {
		rechargeIndexes := mai.getAvailableRechargeActionIndexes()
		if len(rechargeIndexes) > 0 {
			actionChoiceID, err = mai.chooseRechargeAction(rechargeIndexes)
			if err != nil {
				fmt.Println(err)
			} else {
				return mai.buildAIRequest(actionChoiceID, targetID, nil, core.ATMonsterAction)
			}
		}
	}

	// Consider spellcasting vs multi attack
	if mai.hasMultiattack {
		switch mai.AIMode {
		case MAISimple:
			// Deterministic choice: prefer spell if available
			if mai.parent.IsSpellcaster() {
				spellChoice, err = mai.chooseSpell(target)
				if err == nil {
					// Log spell choice event
					events.LogSpellChoiceEvent(mai.parent.GetCurrentEventContext(), mai.parent, spellChoice, mai.parent.SpellCastingManager.GetStatus(), mai.parent.GetEventListener())
					return mai.buildAIRequest(-1, targetID, spellChoice, core.ATSpell)
				}
				// If chooseSpell fails (no slots/spells), fallback to multiattack
			}
			actionChoiceID, err = mai.chooseMultiattackOption()
			if err != nil {
				fmt.Println(err)
			} else {
				return mai.buildAIRequest(actionChoiceID, targetID, nil, core.ATMonsterMultiattack)
			}
			// The tactical AI is going to be replaced by the more advanced weighted decision process
		}
	}

	// No multiattack -> Spellcasting first
	if mai.parent.IsSpellcaster() {
		spellChoice, err = mai.chooseSpell(target)
		if err != nil {
			fmt.Println(err)
		} else {
			// Log spell choice event
			events.LogSpellChoiceEvent(mai.parent.GetCurrentEventContext(), mai.parent, spellChoice, mai.parent.SpellCastingManager.GetStatus(), mai.parent.GetEventListener())
			return mai.buildAIRequest(-1, targetID, spellChoice, core.ATSpell)
		}
	}

	actionChoiceID, err = mai.chooseMonsterAction()
	if err != nil {
		return nil, err
	} else {
		return mai.buildAIRequest(actionChoiceID, targetID, nil, core.ATMonsterAction)
	}
}

func (mai *MonsterAI) getAvailableRechargeActionIndexes() []int {
	availableIdx := make([]int, 0, len(mai.parent.ActionManager.RechargeActions))
	for idx, status := range mai.parent.EntityStateManager.GetRechargeActionStatus() {
		if status {
			availableIdx = append(availableIdx, idx)
		}
	}
	sort.Ints(availableIdx)
	return availableIdx
}

func (mai *MonsterAI) chooseRechargeAction(indexes []int) (int, error) {
	if len(indexes) == 0 {
		return -1, fmt.Errorf("no recharge actions available")
	}
	bestIndex := -1
	bestAvg := 0

	for _, idx := range indexes {
		if bestIndex == -1 {
			bestIndex = idx
			bestAvg = mai.parent.ActionManager.ActionAttackData[idx].Average
			continue
		}

		idxAvg := mai.parent.ActionManager.ActionAttackData[idx].Average
		if idxAvg > bestAvg {
			bestIndex = idx
			bestAvg = idxAvg
		}
	}

	if bestIndex == -1 {
		return -1, fmt.Errorf("no recharge actions available")
	}

	return bestIndex, nil
}

func (mai *MonsterAI) chooseMultiattackOption() (int, error) {
	if len(mai.parent.ActionManager.Multiattacks) == 0 {
		return -1, fmt.Errorf("no multiattack options available")
	}

	indices := make([]int, 0, len(mai.parent.ActionManager.MultiAttackData))
	for idx := range mai.parent.ActionManager.MultiAttackData {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	switch mai.AIMode {
	case MAISimple:
		// Deterministic choice: pick the first available one
		return indices[0], nil
	case MAITactical:
		bestIndex := -1
		bestAvg := 0

		for _, idx := range indices {
			option := mai.parent.ActionManager.MultiAttackData[idx]
			if bestIndex == -1 {
				bestIndex = idx
				if mai.PrioritizeTotalAvg {
					bestAvg = option.TotalAverage
				} else {
					bestAvg = option.AveragePerAttack
				}
			}
			if mai.PrioritizeTotalAvg {
				if option.TotalAverage > bestAvg {
					bestIndex = idx
					bestAvg = option.TotalAverage
				}
			} else {
				if option.AveragePerAttack > bestAvg {
					bestIndex = idx
					bestAvg = option.AveragePerAttack
				}
			}
		}

		if bestIndex == -1 {
			return -1, fmt.Errorf("no multiattack options available")
		}

		return bestIndex, nil
	}

	return -1, fmt.Errorf("invalid monster AI mode")
}

func (mai *MonsterAI) chooseMonsterAction() (int, error) {
	actionIDs := mai.getActionIDs()
	if len(actionIDs) == 0 {
		return -1, fmt.Errorf("no actions available")
	}

	switch mai.AIMode {
	case MAISimple:
		// Deterministic choice: pick the first one
		return actionIDs[0], nil
	case MAITactical:
		bestIndex := -1
		bestAvg := 0

		// Sort ActionAttackData keys (indices) to ensure deterministic tie-breaking
		indices := make([]int, 0, len(mai.parent.ActionManager.ActionAttackData))
		for idx := range mai.parent.ActionManager.ActionAttackData {
			indices = append(indices, idx)
		}
		sort.Ints(indices)

		for _, idx := range indices {
			action := mai.parent.ActionManager.ActionAttackData[idx]
			if bestIndex == -1 {
				bestIndex = actionIDs[idx]
				bestAvg = action.Average
			}
			if action.Average > bestAvg {
				bestIndex = actionIDs[idx]
				bestAvg = action.Average
			}
		}

		if bestIndex == -1 {
			return -1, fmt.Errorf("no actions available")
		}

		return bestIndex, nil
	}

	return -1, fmt.Errorf("invalid monster AI mode")
}

func (mai *MonsterAI) chooseSpell(target core.Entity) (*core.SpellChoice, error) {
	if mai.parent.EntityStateManager.GetHasUsedAction() {
		return nil, fmt.Errorf("action already used")
	}
	if mai.parent.SpellCastingManager == nil || !mai.parent.IsSpellcaster() {
		return nil, fmt.Errorf("monster is not a spellcaster")
	}

	if mai.combatCtx != nil && mai.combatCtx.Options.UseWeightedAI && target != nil {
		upcast := mai.ShouldExpendResource(target, false)
		mai.parent.SpellCastingManager.SetForcedUpcast(upcast)
	}

	spellChoice, err := mai.parent.SpellCastingManager.ChooseSpellByPriority(core.STDamage, mai.parent.EntityStateManager.GetSpellcastingPriority())
	if err != nil {
		return nil, err
	}

	return spellChoice, nil
}

func (mai *MonsterAI) buildAIRequest(actionIndex int, targetID int, spellChoice *core.SpellChoice, actionType core.ActionType) (*core.AIRequest, error) {
	if mai.combatCtx == nil {
		return nil, fmt.Errorf("combat context not set")
	}
	if actionType == core.ATSpell && spellChoice == nil {
		return nil, fmt.Errorf("spell choice not set")
	}
	if actionType != core.ATSpell && actionIndex == -1 {
		return nil, fmt.Errorf("action index not set")
	}

	target := mai.combatCtx.CombatantInfo[targetID].Combatant.GetEntity()

	req := core.AIRequest{
		Actor:       mai.parent,
		ActorType:   core.EntityMonster,
		TargetID:    targetID,
		Target:      target,
		ActionType:  actionType,
		ActionIndex: actionIndex,
		SpellChoice: spellChoice,
	}

	return &req, nil
}

func (mai *MonsterAI) SelectTargetID(targetType core.TargetType) (core.TargetStatus, int, float64, map[events.DecisionFactor]float64, error) {
	return mai.SelectTargetIDWithLogging(targetType, true)
}

func (mai *MonsterAI) SelectTargetIDWithLogging(targetType core.TargetType, shouldLog bool) (core.TargetStatus, int, float64, map[events.DecisionFactor]float64, error) {
	if mai.combatCtx == nil {
		return core.TargetInvalidType, -1, 0, nil, fmt.Errorf("combat context not set")
	}

	var validTargets map[int]*core.Combatant
	switch targetType {
	case core.TTDamage:
		validTargets = mai.getEnemyTargets()
	case core.TTHealing:
		allies := mai.getAllyTargets()
		validTargets = make(map[int]*core.Combatant)
		needHealing := mai.combatCtx.MonstersInNeedOfHealing
		for _, id := range needHealing {
			if c, ok := allies[id]; ok {
				validTargets[id] = c
			}
		}
	default:
		return core.TargetInvalidType, -1, 0, nil, fmt.Errorf("invalid target type")
	}

	if mai.combatCtx.Options.UseWeightedAI {
		return mai.selectTargetWeighted(validTargets, targetType, shouldLog)
	}

	status, id, err := mai.selectTargetSimple(validTargets, targetType, shouldLog)
	return status, id, 1.0, nil, err
}

func (mai *MonsterAI) selectTargetSimple(validTargets map[int]*core.Combatant, targetType core.TargetType, shouldLog bool) (core.TargetStatus, int, error) {
	status, target, err := core.SelectTargetFromMap(validTargets, mai.parent.EntityStateManager.GetTargetPrioritization(), mai.rng)
	if err != nil || status != core.TargetOK {
		return status, -1, err
	}

	if shouldLog {
		if combatant, ok := validTargets[target]; ok && combatant != nil {
			events.LogTargetChoiceEvent(mai.parent.GetCurrentEventContext(), mai.parent, combatant.GetEntity(), 1.0, nil, mai.parent.GetEventListener())
		}
	}

	return core.TargetOK, target, nil
}

func (mai *MonsterAI) selectTargetWeighted(validTargets map[int]*core.Combatant, targetType core.TargetType, shouldLog bool) (core.TargetStatus, int, float64, map[events.DecisionFactor]float64, error) {
	if len(validTargets) == 0 {
		return core.TargetNone, -1, 0, nil, nil
	}

	avgEnemyDamage := mai.calculateAvgEnemyDamage()

	bestID := -1
	bestScore := -1e18 // Negative infinity initialization
	var bestFactors map[events.DecisionFactor]float64

	type factorContribution struct {
		id      int
		score   float64
		factors map[events.DecisionFactor]float64
	}
	contributions := make(map[int]factorContribution)

	ids := make([]int, 0, len(validTargets))
	for id := range validTargets {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		combatant := validTargets[id]
		score := 0.0
		factors := make(map[events.DecisionFactor]float64)

		if targetType == core.TTDamage {
			// 1. Hitability
			hitability := core.CalculateHitabilityFactor(combatant.Entity.GetAC(), mai.parent.GetAttackBonus())
			contribution := hitability * mai.Weights.TargetFactorWeights.TargetHitability
			score += contribution
			factors[events.FactorHighHitability] = contribution

			// 2. Potency
			potency := core.CalculatePotencyFactor(combatant.Entity.GetAC(), combatant.Entity.GetAttackBonus())
			contribution = potency * mai.Weights.TargetFactorWeights.TargetPotency
			score += contribution
			factors[events.FactorHighPotency] = contribution

			// 3. Vengeance
			if mai.parent.Info != nil && combatant.Entity.GetInstanceID() == mai.parent.Info.Statistics.LastAttackerID {
				score += mai.Weights.TargetFactorWeights.Vengeance
				factors[events.FactorVengeance] = mai.Weights.TargetFactorWeights.Vengeance
			}

			// 4. Low HP
			hpStatus := combatant.Entity.GetHPStatus()
			hpFactor := core.CalculateHPFactor(hpStatus.GetHP(), hpStatus.GetMaxHP(), mai.combatCtx.Options.HPVisibilityMode)
			contribution = hpFactor * mai.Weights.TargetFactorWeights.LowHP
			score += contribution
			factors[events.FactorBloodiedTarget] = contribution

			// 5. Concentration
			if combatant.Entity.IsConcentrating() {
				score += mai.Weights.TargetFactorWeights.ConcentrationBreak
				factors[events.FactorConcentration] = mai.Weights.TargetFactorWeights.ConcentrationBreak
			}

			// 6. High Threat
			if mai.combatCtx.MaxDamageSeen > 0 {
				threat := float64(combatant.Info.Statistics.LastDamageDealt) / float64(mai.combatCtx.MaxDamageSeen)
				contribution = threat * mai.Weights.TargetFactorWeights.HighThreat
				score += contribution
				factors[events.FactorHighThreat] = contribution
			}

			// 7. Elite Priority
			if combatant.Entity.GetIsLegendary() {
				score += mai.Weights.TargetFactorWeights.ElitePriority
				factors[events.FactorEliteThreat] = mai.Weights.TargetFactorWeights.ElitePriority
			}
		} else if targetType == core.TTHealing {
			// 1. Emergency Heal
			hpStatus := combatant.Entity.GetHPStatus()
			emergency := core.CalculateEmergencyHealFactor(hpStatus.GetHP(), avgEnemyDamage)
			contribution := emergency * mai.Weights.TargetFactorWeights.EmergencyHeal
			score += contribution
			factors[events.FactorEmergencyHeal] = contribution
		}

		// Apply Deterministic Noise for Monsters
		if mai.combatCtx.Options.EnableMonsterNoise {
			noise := mai.rng.Float64() * mai.combatCtx.Options.MonsterNoiseWeight
			score += noise
			factors[events.FactorDeterministicNoise] = noise
		}

		contributions[id] = factorContribution{id: id, score: score, factors: factors}

		if score > bestScore {
			bestScore = score
			bestID = id
			bestFactors = factors
		}
	}

	if bestID == -1 {
		return core.TargetNone, -1, 0, nil, nil
	}

	if mai.combatCtx.Options.DebugAI {
		fmt.Printf("[DEBUG TARGET] %s selects %s. Total Score: %.2f. Breakdown: ",
			mai.parent.GetName(), validTargets[bestID].Entity.GetName(), bestScore)
		// Sort factors for consistent output
		var factorKeys []events.DecisionFactor
		for k := range bestFactors {
			factorKeys = append(factorKeys, k)
		}
		sort.Slice(factorKeys, func(i, j int) bool {
			return factorKeys[i] < factorKeys[j]
		})
		for _, k := range factorKeys {
			fmt.Printf("[%s: %.2f] ", k, bestFactors[k])
		}
		fmt.Println()
	}

	if shouldLog {
		events.LogTargetChoiceEvent(mai.parent.GetCurrentEventContext(), mai.parent, validTargets[bestID].Entity, bestScore, bestFactors, mai.parent.GetEventListener())
	}
	return core.TargetOK, bestID, bestScore, bestFactors, nil
}

func (mai *MonsterAI) chooseActionWeighted() (core.ActionType, error) {
	healUtility := 0.0
	damageUtility := mai.Weights.ActionWeights[core.ATMonsterDamage]
	if damageUtility == 0 {
		damageUtility = 1.0
	}

	healFactors := make(map[events.DecisionFactor]float64)
	damageFactors := make(map[events.DecisionFactor]float64)

	// Evaluate Damage Utility based on best target
	tStatus, _, bestDamageScore, bestDamageFactors, _ := mai.SelectTargetIDWithLogging(core.TTDamage, false)
	if tStatus == core.TargetOK {
		damageUtility *= bestDamageScore
		if mai.combatCtx.Options.DebugAI {
			fmt.Printf("[DEBUG AI] %s: Damage Base Utility: %.2f, Best Target Score: %.2f, Final Damage Utility: %.2f\n",
				mai.parent.GetName(), mai.Weights.ActionWeights[core.ATMonsterDamage], bestDamageScore, damageUtility)
		}
		for f, v := range bestDamageFactors {
			damageFactors[f] = v * mai.Weights.ActionWeights[core.ATMonsterDamage]
		}
	} else {
		damageUtility = 0
	}
	damageFactors[events.FactorOptimalDamage] = damageUtility

	if (mai.parent.SpellCastingManager != nil && mai.parent.SpellCastingManager.HasHealingSpells()) ||
		(mai.parent.ActionManager != nil && mai.parent.ActionManager.HasHealingAbilities()) {
		// Evaluate ALL allies for healing utility
		allies := mai.getAllyTargets()
		if len(allies) > 0 {
			// Find best healing target without logging yet
			tStatus, _, bestHealScore, _, _ := mai.SelectTargetIDWithLogging(core.TTHealing, false)
			if tStatus == core.TargetOK {
				healUtility = mai.Weights.ActionWeights[core.ATMonsterHeal] * bestHealScore
				healFactors[events.FactorEmergencyHeal] = healUtility
				if mai.combatCtx.Options.DebugAI {
					fmt.Printf("[DEBUG AI] %s: Heal Base Utility: %.2f, Best Heal Score: %.2f, Final Heal Utility: %.2f\n",
						mai.parent.GetName(), mai.Weights.ActionWeights[core.ATMonsterHeal], bestHealScore, healUtility)
				}
			}
		}
	}

	var chosenAction core.ActionType
	var finalUtility float64
	var allScores []events.ActionUtilityScore

	allScores = append(allScores, events.ActionUtilityScore{
		ActionType: core.ATMonsterDamage,
		TotalScore: damageUtility,
		Factors:    damageFactors,
	})

	if healUtility > 0 {
		allScores = append(allScores, events.ActionUtilityScore{
			ActionType: core.ATMonsterHeal,
			TotalScore: healUtility,
			Factors:    healFactors,
		})
	}

	if healUtility > damageUtility {
		chosenAction = core.ATMonsterHeal
		finalUtility = healUtility
	} else {
		chosenAction = core.ATMonsterDamage
		finalUtility = damageUtility
	}

	// Determine top reasons
	topReasons := make([]events.DecisionFactor, 0)
	currentFactors := damageFactors
	if chosenAction == core.ATMonsterHeal {
		currentFactors = healFactors
	}

	type factorPair struct {
		f events.DecisionFactor
		v float64
	}
	pairs := make([]factorPair, 0, len(currentFactors))
	for f, v := range currentFactors {
		if v > 0 {
			pairs = append(pairs, factorPair{f, v})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].v > pairs[j].v
	})

	for i := 0; i < len(pairs) && i < 3; i++ {
		topReasons = append(topReasons, pairs[i].f)
	}

	if len(topReasons) == 0 {
		topReasons = append(topReasons, events.FactorOptimalDamage)
	}

	events.LogMonsterActionChoiceEvent(mai.parent.GetCurrentEventContext(), mai.parent, chosenAction, allScores, topReasons, finalUtility, mai.parent.GetEventListener())

	return chosenAction, nil
}

// ShouldExpendResource determines if a high-value resource should be used based on target potency.
func (mai *MonsterAI) ShouldExpendResource(target core.Entity, isCritical bool) bool {
	if mai.Weights == nil {
		return false
	}

	potency := core.CalculatePotencyFactor(target.GetAC(), target.GetAttackBonus())
	weight := mai.Weights.ResourceExpenditureWeight

	// DM "Noise" can also influence resource expenditure
	if mai.combatCtx != nil && mai.combatCtx.Options.EnableMonsterNoise {
		noise := mai.rng.Float64() * mai.combatCtx.Options.MonsterNoiseWeight
		weight += noise
	}

	if isCritical {
		weight *= 2.0
	}

	return core.ShouldExpendHighResource(potency, weight)
}

func (mai *MonsterAI) calculateAvgEnemyDamage() int {
	totalAvgDamage := 0.0
	enemyCount := 0

	ids := make([]int, 0, len(mai.combatCtx.CombatantInfo))
	for id := range mai.combatCtx.CombatantInfo {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		info := mai.combatCtx.CombatantInfo[id]
		if (mai.parent.IsMonster() != info.Combatant.Entity.IsMonster()) && !info.Combatant.Entity.IsDead() {
			totalAvgDamage += info.Statistics.AverageDamagePerRound
			enemyCount++
		}
	}
	if enemyCount > 0 {
		res := int(totalAvgDamage / float64(enemyCount))
		if res < 1 {
			return 1
		}
		return res
	}
	return 10
}

func (mai *MonsterAI) getEnemyTargets() map[int]*core.Combatant {
	if mai.combatCtx == nil {
		return nil
	}

	enemies := make(map[int]*core.Combatant)
	self := mai.parent

	for id, combatant := range mai.combatCtx.CombatantInfo {
		// Skip lair combatant entries; lair is not a valid target
		if combatant.Combatant.IsLair {
			continue
		}
		e := combatant.Combatant.GetEntity()
		if !e.IsUnconscious() && (self.IsMonster() != e.IsMonster()) {
			enemies[id] = combatant.Combatant
		}
	}

	return enemies
}

func (mai *MonsterAI) getAllyTargets() map[int]*core.Combatant {
	allies := make(map[int]*core.Combatant)
	self := mai.parent

	for id, combatant := range mai.combatCtx.CombatantInfo {
		// Skip lair combatant entries
		if combatant.Combatant.IsLair {
			continue
		}
		e := combatant.Combatant.GetEntity()
		if !e.IsDead() && (self.IsMonster() == e.IsMonster()) {
			allies[id] = combatant.Combatant
		}
	}

	return allies
}

func (mai *MonsterAI) getActionIDs() []int {
	if mai.parent.ActionManager == nil {
		return nil
	}

	actionIDs := make([]int, 0, len(mai.parent.ActionManager.Actions))
	for idx := range mai.parent.ActionManager.Actions {
		actionIDs = append(actionIDs, idx)
	}
	sort.Ints(actionIDs)
	return actionIDs
}

func (mai *MonsterAI) ChooseMonsterActionType() (core.ActionType, error) {
	if mai.combatCtx == nil {
		return core.ATNoAction, fmt.Errorf("combat context not set")
	}

	if mai.parent.EntityStateManager.GetHasUsedAction() {
		return core.ATNoAction, nil
	}

	if mai.combatCtx.Options.UseWeightedAI {
		return mai.chooseActionWeighted()
	}

	if mai.combatCtx.Options.AllowMonsterHeals {
		if (mai.parent.SpellCastingManager != nil && mai.parent.SpellCastingManager.HasHealingSpells()) ||
			(mai.parent.ActionManager != nil && mai.parent.ActionManager.HasHealingAbilities()) {
			if len(mai.combatCtx.MonstersInNeedOfHealing) > 0 {
				return core.ATMonsterHeal, nil
			}
		}
	}

	return core.ATMonsterDamage, nil
}
