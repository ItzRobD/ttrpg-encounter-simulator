package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/races"
	"fmt"
	"math/rand/v2"
	"sort"
)

type CharacterAI struct {
	parent    *Character
	combatCtx *core.CombatContext
	eventCtx  *core.EventContext
	Weights   *core.UtilityWeights
	rng       *rand.Rand
}

func NewCharacterAI(c *Character, weights *core.UtilityWeights) *CharacterAI {
	return &CharacterAI{
		parent:    c,
		combatCtx: nil,
		Weights:   weights,
		rng:       c.GetRNG(),
	}
}

func (cai *CharacterAI) UpdateCombatContext(ctx *core.CombatContext) {
	cai.combatCtx = ctx
}

func (cai *CharacterAI) UpdateEventContext(ctx *core.EventContext) {
	cai.eventCtx = ctx
}

func (cai *CharacterAI) chooseDamageSpell(target core.Entity) (*core.SpellChoice, error) {
	if cai.combatCtx != nil && cai.combatCtx.Options.UseWeightedAI {
		upcast := cai.ShouldExpendResource(target, false)
		cai.parent.SpellCastingManager.SetForcedUpcast(upcast)
	}
	return cai.parent.SpellCastingManager.ChooseSpellByPriority(core.STDamage, cai.parent.EntityStateManager.GetSpellcastingPriority())
}

func (cai *CharacterAI) chooseDamageActionType() (core.ActionType, error) {
	actionPref := cai.parent.EntityStateManager.GetActionPreference()
	actionType := core.ATNoAction
	switch actionPref {
	case core.APPreferMelee:
		if cai.parent.EquipmentManager.HasMeleeWeapon() {
			actionType = core.ATMelee
		} else {
			return cai.chooseFallbackAction(core.ATMelee), nil
		}
	case core.APPreferRanged:
		if cai.parent.EquipmentManager.HasRangedWeapon() {
			actionType = core.ATRanged
		} else {
			return cai.chooseFallbackAction(core.ATRanged), nil
		}
	case core.APNoPreference:
		// When no preference is set, we need a deterministic choice.
		// For now, let's prefer Spells if available, then Melee, then Ranged.
		if cai.parent.IsSpellcaster() && cai.parent.SpellCastingManager.GetDamageSpellCount() > 0 {
			actionType = core.ATSpell
		} else if cai.parent.EquipmentManager.HasMeleeWeapon() {
			actionType = core.ATMelee
		} else if cai.parent.EquipmentManager.HasRangedWeapon() {
			actionType = core.ATRanged
		} else {
			actionType = core.ATUnarmed
		}
	case core.APPreferSpells:
		if cai.parent.IsSpellcaster() && cai.parent.SpellCastingManager.GetDamageSpellCount() > 0 {
			actionType = core.ATSpell
		} else {
			if cai.parent.EquipmentManager.HasMeleeWeapon() {
				actionType = core.ATMelee
			} else {
				return cai.chooseFallbackAction(core.ATMelee), nil
			}
		}
	default:
		return core.ATNoAction, fmt.Errorf("unknown action preference %s", actionPref)
	}

	// Structured logging: chosen damage action type (only if weighted AI is not used, to avoid duplication)
	if !cai.combatCtx.Options.UseWeightedAI {
		cai.parent.LogEvent(events.ETActionChoiceEvent, &events.ActionChoiceData{
			Choice:       actionType,
			AllScores:    nil,
			TopReasons:   nil,
			UtilityScore: 0,
		})
	}
	return actionType, nil
}

func (cai *CharacterAI) chooseFallbackAction(exclude core.ActionType) core.ActionType {
	// Standard fallback priority, excluding the preferred type
	if exclude != core.ATRanged && cai.parent.EquipmentManager.HasRangedWeapon() {
		return core.ATRanged
	}
	if exclude != core.ATMelee && cai.parent.EquipmentManager.HasMeleeWeapon() {
		return core.ATMelee
	}
	if exclude != core.ATSpell &&
		cai.parent.IsSpellcaster() &&
		cai.parent.SpellCastingManager.GetDamageSpellCount() > 0 {
		return core.ATSpell
	}
	return core.ATUnarmed
}

func (cai *CharacterAI) hasValidTargets(targetType core.TargetType) bool {
	var validTargets map[int]*core.Combatant
	switch targetType {
	case core.TTDamage:
		validTargets = cai.getEnemyTargets()
	case core.TTHealing:
		allies := cai.getAllyTargets()
		validTargets = make(map[int]*core.Combatant)
		needHealing := cai.combatCtx.CharactersInNeedOfHealing
		for _, id := range needHealing {
			if c, ok := allies[id]; ok {
				validTargets[id] = c
			}
		}
	default:
		return false
	}
	return len(validTargets) > 0
}

func (cai *CharacterAI) SelectTargetID(targetType core.TargetType) (core.TargetStatus, int, float64, map[events.DecisionFactor]float64, error) {
	return cai.SelectTargetIDWithLogging(targetType, true)
}

func (cai *CharacterAI) SelectTargetIDWithLogging(targetType core.TargetType, shouldLog bool) (core.TargetStatus, int, float64, map[events.DecisionFactor]float64, error) {
	if cai.combatCtx == nil {
		return core.TargetInvalidType, -1, 0, nil, fmt.Errorf("combat context not set")
	}

	var validTargets map[int]*core.Combatant
	switch targetType {
	case core.TTDamage:
		validTargets = cai.getEnemyTargets()
	case core.TTHealing:
		allies := cai.getAllyTargets()
		validTargets = make(map[int]*core.Combatant)
		needHealing := cai.combatCtx.CharactersInNeedOfHealing
		for _, id := range needHealing {
			if c, ok := allies[id]; ok {
				validTargets[id] = c
			}
		}
	default:
		return core.TargetInvalidType, -1, 0, nil, fmt.Errorf("unknown target type")
	}

	if cai.combatCtx.Options.UseWeightedAI {
		return cai.selectTargetWeighted(validTargets, targetType, shouldLog)
	}

	status, id, err := cai.selectTargetSimple(validTargets, targetType, shouldLog)
	return status, id, 1.0, nil, err
}

func (cai *CharacterAI) selectTargetSimple(validTargets map[int]*core.Combatant, targetType core.TargetType, shouldLog bool) (core.TargetStatus, int, error) {
	status, target, err := core.SelectTargetFromMap(validTargets, cai.parent.EntityStateManager.GetTargetPrioritization(), cai.rng)
	if err != nil || status != core.TargetOK {
		return status, -1, err
	}
	// Structured logging: chosen target
	if shouldLog {
		if combatant, ok := validTargets[target]; ok && combatant != nil {
			cai.parent.LogEvent(events.ETTargetChoiceEvent, &events.TargetChoiceData{
				Target:  combatant.GetEntity(),
				Score:   1.0,
				Factors: nil,
			})
		}
	}
	return core.TargetOK, target, nil
}

func (cai *CharacterAI) selectTargetWeighted(validTargets map[int]*core.Combatant, targetType core.TargetType, shouldLog bool) (core.TargetStatus, int, float64, map[events.DecisionFactor]float64, error) {
	if len(validTargets) == 0 {
		return core.TargetNone, -1, 0, nil, nil
	}

	avgEnemyDamage := cai.calculateAvgEnemyDamage()

	bestID := -1
	bestScore := -1e18 // Standard practice for negative infinity initialization
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
			hitability := core.CalculateHitabilityFactor(combatant.Entity.GetAC(), cai.parent.GetAttackBonus())
			contribution := hitability * cai.Weights.TargetFactorWeights.TargetHitability
			score += contribution
			factors[events.FactorHighHitability] = contribution

			// 2. Potency
			potency := core.CalculatePotencyFactor(combatant.Entity.GetAC(), combatant.Entity.GetAttackBonus())
			contribution = potency * cai.Weights.TargetFactorWeights.TargetPotency
			score += contribution
			factors[events.FactorHighPotency] = contribution

			// 3. Vengeance
			if cai.parent.Info != nil && combatant.Entity.GetInstanceID() == cai.parent.Info.Statistics.LastAttackerID {
				score += cai.Weights.TargetFactorWeights.Vengeance
				factors[events.FactorVengeance] = cai.Weights.TargetFactorWeights.Vengeance
			}

			// 4. Low HP
			hpStatus := combatant.Entity.GetHPStatus()
			hpFactor := core.CalculateHPFactor(hpStatus.GetHP(), hpStatus.GetMaxHP(), cai.combatCtx.Options.HPVisibilityMode)
			contribution = hpFactor * cai.Weights.TargetFactorWeights.LowHP
			score += contribution
			factors[events.FactorBloodiedTarget] = contribution

			// 5. Concentration
			if combatant.Entity.IsConcentrating() {
				score += cai.Weights.TargetFactorWeights.ConcentrationBreak
				factors[events.FactorConcentration] = cai.Weights.TargetFactorWeights.ConcentrationBreak
			}

			// 6. High Threat
			if cai.combatCtx.MaxDamageSeen > 0 {
				threat := float64(combatant.Info.Statistics.LastDamageDealt) / float64(cai.combatCtx.MaxDamageSeen)
				contribution = threat * cai.Weights.TargetFactorWeights.HighThreat
				score += contribution
				factors[events.FactorHighThreat] = contribution
			}

			// 7. Elite Priority
			if combatant.Entity.GetIsLegendary() {
				score += cai.Weights.TargetFactorWeights.ElitePriority
				factors[events.FactorEliteThreat] = cai.Weights.TargetFactorWeights.ElitePriority
			}
		} else if targetType == core.TTHealing {
			// 1. Emergency Heal
			hpStatus := combatant.Entity.GetHPStatus()
			emergency := core.CalculateEmergencyHealFactor(hpStatus.GetHP(), avgEnemyDamage)
			contribution := emergency * cai.Weights.TargetFactorWeights.EmergencyHeal
			score += contribution
			factors[events.FactorEmergencyHeal] = contribution
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

	if cai.combatCtx.Options.DebugAI {
		fmt.Printf("[DEBUG TARGET] %s selects %s. Total Score: %.2f. Breakdown: ",
			cai.parent.GetName(), validTargets[bestID].Entity.GetName(), bestScore)
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
		cai.parent.LogEvent(events.ETTargetChoiceEvent, &events.TargetChoiceData{
			Target:  validTargets[bestID].Entity,
			Score:   bestScore,
			Factors: bestFactors,
		})
	}
	return core.TargetOK, bestID, bestScore, bestFactors, nil
}

func (cai *CharacterAI) chooseActionWeighted() (core.ActionType, error) {
	healUtility := 0.0
	damageUtility := cai.Weights.ActionWeights[core.ATDamage]
	if damageUtility == 0 {
		damageUtility = 1.0
	}
	breathUtility := 0.0

	healFactors := make(map[events.DecisionFactor]float64)
	damageFactors := make(map[events.DecisionFactor]float64)
	breathFactors := make(map[events.DecisionFactor]float64)

	// Evaluate Damage Utility based on best target
	tStatus, _, bestDamageScore, bestDamageFactors, _ := cai.SelectTargetIDWithLogging(core.TTDamage, false)
	if tStatus == core.TargetOK {
		damageUtility *= bestDamageScore
		if cai.combatCtx.Options.DebugAI {
			fmt.Printf("[DEBUG AI] %s: Damage Base Utility: %.2f, Best Target Score: %.2f, Final Damage Utility: %.2f\n",
				cai.parent.GetName(), cai.Weights.ActionWeights[core.ATDamage], bestDamageScore, damageUtility)
		}
		for f, v := range bestDamageFactors {
			damageFactors[f] = v * cai.Weights.ActionWeights[core.ATDamage]
		}
	} else {
		damageUtility = 0 // No targets, no damage utility
	}
	damageFactors[events.FactorOptimalDamage] = damageUtility

	if cai.parent.IsHealer() {
		// Evaluate ALL allies for healing utility
		allies := cai.getAllyTargets()
		if len(allies) > 0 {
			// Find ally with highest healing utility
			bestHealScore := 0.0
			avgEnemyDamage := cai.calculateAvgEnemyDamage()

			ids := make([]int, 0, len(allies))
			for id := range allies {
				ids = append(ids, id)
			}
			sort.Ints(ids)

			for _, id := range ids {
				ally := allies[id]
				hpStatus := ally.Entity.GetHPStatus()
				emergency := core.CalculateEmergencyHealFactor(hpStatus.GetHP(), avgEnemyDamage)
				score := emergency * cai.Weights.TargetFactorWeights.EmergencyHeal
				if score > bestHealScore {
					bestHealScore = score
				}
			}

			healUtility = cai.Weights.ActionWeights[core.ATHeal] * bestHealScore
			healFactors[events.FactorEmergencyHeal] = healUtility
			if cai.combatCtx.Options.DebugAI {
				fmt.Printf("[DEBUG AI] %s: Heal Base Utility: %.2f, Best Heal Score: %.2f, Final Heal Utility: %.2f\n",
					cai.parent.GetName(), cai.Weights.ActionWeights[core.ATHeal], bestHealScore, healUtility)
			}
		}
	}

	// Dragonborn Breath Weapon
	if cai.parent.Race.ID == races.Dragonborn &&
		!cai.parent.EntityStateManager.GetDBBreathWeaponUsed() &&
		cai.combatCtx.Options.AllowDragonbornBreathAttack {

		// For now, we use the same best damage score as a proxy for breath weapon quality
		// In a more advanced sim, we'd check if many targets are in range/AOE.
		if tStatus == core.TargetOK {
			breathUtility = cai.Weights.ActionWeights[core.ATDragonbornBreathWeapon]
			if breathUtility == 0 {
				breathUtility = 1.0 // Default to same as standard damage if not set
			}
			breathUtility *= bestDamageScore
			breathFactors[events.FactorOptimalDamage] = breathUtility
			if cai.combatCtx.Options.DebugAI {
				fmt.Printf("[DEBUG AI] %s: Breath Base Utility: %.2f, Best Target Score: %.2f, Final Breath Utility: %.2f\n",
					cai.parent.GetName(), cai.Weights.ActionWeights[core.ATDragonbornBreathWeapon], bestDamageScore, breathUtility)
			}
		}
	}

	var chosenAction core.ActionType
	var finalUtility float64
	var allScores []events.ActionUtilityScore

	allScores = append(allScores, events.ActionUtilityScore{
		ActionType: core.ATDamage,
		TotalScore: damageUtility,
		Factors:    damageFactors,
	})

	if cai.parent.IsHealer() {
		allScores = append(allScores, events.ActionUtilityScore{
			ActionType: core.ATHeal,
			TotalScore: healUtility,
			Factors:    healFactors,
		})
	}

	if breathUtility > 0 {
		allScores = append(allScores, events.ActionUtilityScore{
			ActionType: core.ATDragonbornBreathWeapon,
			TotalScore: breathUtility,
			Factors:    breathFactors,
		})
	}

	// Find the highest utility action
	chosenAction = core.ATDamage
	finalUtility = damageUtility

	if healUtility > finalUtility {
		chosenAction = core.ATHeal
		finalUtility = healUtility
	}

	if breathUtility > finalUtility {
		chosenAction = core.ATDragonbornBreathWeapon
		finalUtility = breathUtility
	}

	// Determine top reasons
	topReasons := make([]events.DecisionFactor, 0)
	var currentFactors map[events.DecisionFactor]float64
	switch chosenAction {
	case core.ATHeal:
		currentFactors = healFactors
	case core.ATDragonbornBreathWeapon:
		currentFactors = breathFactors
	default:
		currentFactors = damageFactors
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

	cai.parent.LogEvent(events.ETActionChoiceEvent, &events.ActionChoiceData{
		Choice:       chosenAction,
		AllScores:    allScores,
		TopReasons:   topReasons,
		UtilityScore: finalUtility,
	})

	return chosenAction, nil
}

// ShouldExpendResource determines if a high-value resource (like Smite or high-level spell slot)
// should be used based on target potency and simulation settings.
func (cai *CharacterAI) ShouldExpendResource(target core.Entity, isCritical bool) bool {
	if cai.Weights == nil {
		return false
	}

	potency := core.CalculatePotencyFactor(target.GetAC(), target.GetAttackBonus())
	weight := cai.Weights.ResourceExpenditureWeight

	// Critical hits significantly increase the desire to spend resources (the "Big Hit")
	if isCritical {
		weight *= 2.0
	}

	return core.ShouldExpendHighResource(potency, weight)
}

func (cai *CharacterAI) calculateAvgEnemyDamage() int {
	totalAvgDamage := 0.0
	enemyCount := 0

	ids := make([]int, 0, len(cai.combatCtx.CombatantInfo))
	for id := range cai.combatCtx.CombatantInfo {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		info := cai.combatCtx.CombatantInfo[id]
		if (cai.parent.IsCharacter() != info.Combatant.Entity.IsCharacter()) && !info.Combatant.Entity.IsDead() {
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
	return 10 // Safe default if no active enemies
}

func (cai *CharacterAI) getEnemyTargets() map[int]*core.Combatant {
	enemies := make(map[int]*core.Combatant)
	self := cai.parent

	for id, combatant := range cai.combatCtx.CombatantInfo {
		// Skip lair combatant entries
		if combatant.Combatant.IsLair {
			continue
		}
		e := combatant.Combatant.GetEntity()
		if !e.IsUnconscious() && (self.IsCharacter() != e.IsCharacter()) {
			enemies[id] = combatant.Combatant
		}
	}

	return enemies
}

func (cai *CharacterAI) getAllyTargets() map[int]*core.Combatant {
	allies := make(map[int]*core.Combatant)
	self := cai.parent

	for id, combatant := range cai.combatCtx.CombatantInfo {
		e := combatant.Combatant.GetEntity()
		if !e.IsDead() && (self.IsCharacter() == e.IsCharacter()) {
			allies[id] = combatant.Combatant
		}
	}

	return allies
}

func (cai *CharacterAI) ChooseCharacterActionType() (core.ActionType, error) {
	if cai.combatCtx == nil {
		return core.ATNoAction, fmt.Errorf("combat context not set")
	}

	if cai.parent.EntityStateManager.GetHasUsedAction() {
		return core.ATNoAction, nil
	}

	if cai.combatCtx.Options.UseWeightedAI {
		return cai.chooseActionWeighted()
	}

	if cai.parent.Race.ID == races.Dragonborn && !cai.parent.EntityStateManager.GetDBBreathWeaponUsed() {
		// Use breath weapon if there are targets (simple AI logic)
		if cai.hasValidTargets(core.TTDamage) {
			return core.ATDragonbornBreathWeapon, nil
		}
	}

	if cai.combatCtx.Options.AllowCharacterHeals {
		if cai.parent.IsHealer() && len(cai.combatCtx.CharactersInNeedOfHealing) > 0 {
			return core.ATHeal, nil
		}
	}

	return core.ATDamage, nil
}

func (cai *CharacterAI) createCharacterHealActionRequest() (*core.AIRequest, error) {
	if cai.parent.EntityStateManager.GetHasUsedAction() {
		return nil, nil
	}

	tStatus, targetID, _, _, err := cai.SelectTargetID(core.TTHealing)
	if err != nil {
		return nil, err
	}
	if tStatus == core.TargetNone {
		cai.parent.LogEvent(events.ECombatEventMessage, "No valid healing targets")
		return nil, nil
	}

	target := cai.combatCtx.CombatantInfo[targetID].Combatant.Entity
	healReq, err := cai.parent.CreateHealRequest(target)
	if err != nil {
		return nil, err
	}

	if healReq.Source == core.HealSourceSpell {
		// Log spell choice event
		cai.parent.LogEvent(events.ETSpellChoiceEvent, &events.SpellChoiceData{
			Choice: healReq.SpellChoice,
			Status: cai.parent.SpellCastingManager.GetStatus(),
			Target: target,
		})
	}

	// Logging for tactical action (only if weighted AI is not used, to avoid duplication)
	if !cai.combatCtx.Options.UseWeightedAI {
		cai.parent.LogEvent(events.ETActionChoiceEvent, &events.ActionChoiceData{
			Choice:       core.ATHeal,
			AllScores:    nil,
			TopReasons:   nil,
			UtilityScore: 0,
		})
	}

	return &core.AIRequest{
		Actor:       cai.parent,
		ActorType:   core.EntityCharacter,
		Target:      target,
		TargetID:    targetID,
		ActionType:  core.ATHeal,
		HealRequest: healReq,
	}, nil
}

func (cai *CharacterAI) createCharacterDamageActionRequest() (*core.AIRequest, error) {
	if cai.parent.EntityStateManager.GetHasUsedAction() {
		return nil, nil
	}

	var req core.AIRequest
	var choice *core.SpellChoice
	var useVersatile bool
	var slot core.WeaponSlot

	tStatus, targetID, _, _, err := cai.SelectTargetID(core.TTDamage)
	if err != nil {
		return nil, err
	}
	if tStatus == core.TargetNone {
		cai.parent.LogEvent(events.ECombatEventMessage, "No valid targets")
		return nil, nil
	}

	at, err := cai.chooseDamageActionType()
	if err != nil {
		return nil, err
	}

	switch at {
	case core.ATSpell:
		target := cai.combatCtx.CombatantInfo[targetID].Combatant.GetEntity()
		choice, err = cai.chooseDamageSpell(target)
		if err != nil {
			return nil, err
		}

		// Log spell choice event
		cai.parent.LogEvent(events.ETSpellChoiceEvent, &events.SpellChoiceData{
			Choice: choice,
			Status: cai.parent.SpellCastingManager.GetStatus(),
			Target: target,
		})
	case core.ATMelee:
		slot = core.WSPrimary
		em := cai.parent.EquipmentManager
		hasShield := em.HasShieldEquipped

		preferVersatile := false
		switch cai.parent.EntityStateManager.GetVersatileWeaponPreference() {
		case core.VWPPreferVersatile:
			preferVersatile = true
		case core.VWPNoPreference:
			preferVersatile = cai.rng.IntN(2) == 1
		case core.VWPPreferNonVersatile:
			preferVersatile = false
		}

		if preferVersatile && !hasShield {
			primaryWeapon, wErr := em.GetWeaponFromSlot(core.WSPrimary)
			if wErr != nil {
				return nil, wErr
			}
			useVersatile = primaryWeapon.Properties.IsVersatile
		}
	case core.ATRanged:
		slot = core.WSPrimary
		if w, wErr := cai.parent.EquipmentManager.GetWeaponFromSlot(core.WSRanged); wErr == nil {
			if w.Properties.IsRanged {
				slot = core.WSRanged
			} else {
				// explicit fallback to primary; check capability if you want to enforce it
				slot = core.WSPrimary
			}
		} else {
			// Ranged slot missing: try primary (and optionally secondary) as fallbacks
			if pw, pwErr := cai.parent.EquipmentManager.GetWeaponFromSlot(core.WSPrimary); pwErr == nil {
				if pw.Properties.IsRanged {
					slot = core.WSPrimary
				} else {
					// Optional: try secondary for thrown/ranged; otherwise, surface a clear error
					if sw, swErr := cai.parent.EquipmentManager.GetWeaponFromSlot(core.WSSecondary); swErr == nil && sw.Properties.IsRanged {
						slot = core.WSSecondary
					} else {
						return nil, fmt.Errorf("no valid ranged weapon available in ranged/primary/secondary slots")
					}
				}
			} else {
				return nil, fmt.Errorf("no valid ranged weapon available (ranged slot missing; primary lookup failed: %v)", pwErr)
			}
		}
	}

	req = core.AIRequest{
		Actor:        cai.parent,
		ActorType:    core.EntityCharacter,
		TargetID:     targetID,
		Target:       cai.combatCtx.CombatantInfo[targetID].Combatant.GetEntity(),
		ActionType:   at,
		WeaponSlot:   slot,
		UseVersatile: useVersatile,
		SpellChoice:  choice,
	}

	return &req, nil
}

func (cai *CharacterAI) createCharacterOffhandActionRequest() (*core.AIRequest, error) {
	if cai.combatCtx == nil {
		return nil, fmt.Errorf("combat context not set")
	}

	em := cai.parent.EquipmentManager

	// Check conditions, quietly fail
	if cai.parent.EntityStateManager.GetHasUsedBonusAction() {
		return nil, nil // No bonus action available
	}
	if em.HasShieldEquipped {
		return nil, nil // Shield blocks offhand attacks
	}
	if _, err := em.GetWeaponFromSlot(core.WSSecondary); err != nil {
		return nil, nil // No offhand
	}

	tStatus, targetID, _, _, err := cai.SelectTargetID(core.TTDamage)
	if err != nil || tStatus != core.TargetOK {
		return nil, nil
	}

	req := &core.AIRequest{
		Actor:      cai.parent,
		ActorType:  core.EntityCharacter,
		TargetID:   targetID,
		Target:     cai.combatCtx.CombatantInfo[targetID].Combatant.GetEntity(),
		ActionType: core.ATOffhand,
		WeaponSlot: core.WSSecondary,
	}

	return req, nil
}

func (cai *CharacterAI) createDragonbornBreathWeaponRequest() (*core.AIRequest, error) {
	if cai.parent.EntityStateManager.GetHasUsedAction() {
		return nil, nil
	}

	tStatus, targetID, _, _, err := cai.SelectTargetID(core.TTDamage)
	if err != nil {
		return nil, err
	}
	if tStatus == core.TargetNone {
		cai.parent.LogEvent(events.ECombatEventMessage, "No valid targets for breath weapon")
		return nil, nil
	}

	// Logging for tactical action (only if weighted AI is not used, to avoid duplication)
	if !cai.combatCtx.Options.UseWeightedAI {
		cai.parent.LogEvent(events.ETActionChoiceEvent, &events.ActionChoiceData{
			Choice:       core.ATDragonbornBreathWeapon,
			AllScores:    nil,
			TopReasons:   nil,
			UtilityScore: 0,
		})
	}

	req := &core.AIRequest{
		Actor:      cai.parent,
		ActorType:  core.EntityCharacter,
		TargetID:   targetID,
		Target:     cai.combatCtx.CombatantInfo[targetID].Combatant.GetEntity(),
		ActionType: core.ATDragonbornBreathWeapon,
		Request:    core.AIReqDragonbornBreathWeapon,
	}

	return req, nil
}
