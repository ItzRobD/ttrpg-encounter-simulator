package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"math/rand/v2"
	"sort"
	"strings"
)

// ActionIntent represents what an actor wants to do.
type ActionIntent struct {
	ActivationType core.ActivationType // ActAction/ActBonus/ActReaction/ActLegendary
	ActorID        int
	TargetIDs      []int
	Action         core.Action
	ActionType     core.ActionType
}

// AIDirector is the centralized "brain" that makes decisions for actors.
type AIDirector struct {
	rng *rand.Rand
}

func NewAIDirector(rng *rand.Rand) *AIDirector {
	return &AIDirector{
		rng: rng,
	}
}

// SelectActionDecision picks a decision type based on actor preference and state.
func (aid *AIDirector) SelectActionDecision(a *actor.Actor, ed *EncounterDirector) core.Decision {
	basePreference := a.Behavior.ActionPreference
	if basePreference == "" {
		basePreference = core.APAttack
	}
	// Slayer will never heal
	if basePreference.IsSlayer() {
		return core.DecisionAttack
	}
	// Check if actor triggered berserk
	if a.StateManager.Conditions.Has(core.ConditionBerserk) {
		return core.DecisionAttack
	}

	currentPreference := basePreference
	// Review: preferRanged only used if we fall back to attack after having healing capacity but no needs.
	preferRanged := basePreference == core.APRanged || (a.Behavior.SecondaryActionPreference == core.APRanged && basePreference == core.APHeal)

	// Dynamic ActionPreference Override:
	if ed.Statistics != nil {
		hasHealingCapacity := aid.hasHealingAbilities(a)
		foundNeed := false

		// 1. Emergency Healing Check
		if hasHealingCapacity && len(ed.Statistics.NeedsEmergencyHealing) > 0 {
			for _, targetID := range ed.Statistics.NeedsEmergencyHealing {
				if targetActor, ok := ed.Actors[targetID]; ok && targetActor.Side == a.Side {
					currentPreference = core.APHeal
					foundNeed = true
					break
				}
			}
		}

		// 2. Regular Healing Check
		if hasHealingCapacity && !foundNeed && (currentPreference == core.APHeal || basePreference == core.APHeal || a.Behavior.SecondaryActionPreference == core.APHeal) {
			if len(ed.Statistics.NeedsHealing) > 0 {
				for _, targetID := range ed.Statistics.NeedsHealing {
					if targetActor, ok := ed.Actors[targetID]; ok && targetActor.Side == a.Side {
						currentPreference = core.APHeal
						foundNeed = true
						break
					}
				}
			}

			// If we can heal but no one needs it, fall back
			if !foundNeed && currentPreference == core.APHeal {
				if a.Behavior.SecondaryActionPreference != "" && a.Behavior.SecondaryActionPreference != core.APHeal {
					currentPreference = a.Behavior.SecondaryActionPreference
				} else if basePreference != core.APHeal {
					currentPreference = basePreference
				} else {
					// Default to attack
					if preferRanged {
						currentPreference = core.APRanged
					} else {
						currentPreference = core.APAttack
					}
				}
			}
		}
	}

	if currentPreference == core.APHeal {
		return core.DecisionHeal
	}

	return core.DecisionAttack
}

// SelectAction determines the intents for a given decision.
func (aid *AIDirector) SelectAction(a *actor.Actor, d core.Decision, ed *EncounterDirector) []ActionIntent {
	var intents []ActionIntent

	// Legendary action
	if d == core.DecisionLegendary {
		// Make sure we didn't end up here with the wrong actor
		if !a.IsLegendary() || a.StateManager.RemainingLegendaryActionCount() <= 0 {
			return nil
		}

		action, targetIDs := aid.chooseBestAction(a, core.ActLegendary, d, ed)
		if action != nil {
			intent := ActionIntent{
				ActivationType: core.ActLegendary,
				ActorID:        a.InstanceID,
				TargetIDs:      targetIDs,
				Action:         *action,
				ActionType:     action.ActionType,
			}
			intents = append(intents, intent)
		}
		return intents
	}

	// Main action
	if a.StateManager.ActionUsedCount < 1 {
		action, targetIDs := aid.chooseBestAction(a, core.ActAction, d, ed)
		if action != nil {
			intent := ActionIntent{
				ActivationType: core.ActAction,
				ActorID:        a.InstanceID,
				TargetIDs:      targetIDs,
				Action:         *action,
				ActionType:     action.ActionType,
			}
			intents = append(intents, intent)
		}
	}

	// Bonus action
	if a.StateManager.BonusActionUsedCount < 1 {
		action, targetIDs := aid.chooseBestAction(a, core.ActBonus, d, ed)
		if action != nil {
			intent := ActionIntent{
				ActivationType: core.ActBonus,
				ActorID:        a.InstanceID,
				TargetIDs:      targetIDs,
				Action:         *action,
				ActionType:     action.ActionType,
			}
			intents = append(intents, intent)
		}
	}

	return intents
}

// chooseBestAction selects an action and target for a decision.
func (aid *AIDirector) chooseBestAction(a *actor.Actor, costType core.ActivationType, d core.Decision, ed *EncounterDirector) (*core.Action, []int) {
	switch d {
	case core.DecisionHeal:
		if (len(ed.Statistics.NeedsHealing) == 0 && len(ed.Statistics.NeedsEmergencyHealing) == 0) || !aid.hasHealingAbilities(a) {
			// We shouldn't end up here, but if we do, choose an attack action instead
			return aid.chooseBestAction(a, costType, core.DecisionAttack, ed)
		}

		bestAction, targetIDs := aid.chooseBestHealingAction(a, costType, ed)
		if bestAction == nil {
			// If we can't find a valid healing target/action for this specific cost type,
			// fall back to attack instead of doing nothing.
			return aid.chooseBestAction(a, costType, core.DecisionAttack, ed)
		}
		return bestAction, targetIDs
	case core.DecisionAttack:
		bestAction, targetIDs := aid.chooseBestAttackAction(a, costType, ed)
		if bestAction == nil {
			return nil, nil
		}
		return bestAction, targetIDs
	case core.DecisionLegendary:
		bestAction, targetIDs := aid.chooseBestAttackAction(a, core.ActLegendary, ed)
		if bestAction == nil {
			return nil, nil
		}
		return bestAction, targetIDs
	default:
		return nil, nil
	}
}

// chooseBestAttackAction picks the most effective attack action.
func (aid *AIDirector) chooseBestAttackAction(a *actor.Actor, costType core.ActivationType, ed *EncounterDirector) (*core.Action, []int) {
	targets := ed.GetEnemyTargets(a)
	targetIDs := aid.SelectTarget(a, targets, core.TTDamage, ed)
	if len(targetIDs) == 0 {
		return nil, nil
	}
	// For attack selection heuristics, we use the primary target
	primaryTargetID := targetIDs[0]

	policy := core.ActionPolicyHighestDamage
	if ed.SimOptions != nil && ed.SimOptions.ActionSelectionPolicy != "" {
		policy = ed.SimOptions.ActionSelectionPolicy
	}

	// Combine available actions
	availableActions := make([]core.Action, 0)
	for _, act := range a.Actions {
		if act.Cost.ActivationType == costType && (act.AverageDamage > 0 || act.ActionType == core.ATMultiAttack || act.HasDC) {
			if act.Cost.ActivationType == core.ActLegendary &&
				a.StateManager.RemainingLegendaryActionCount() <= act.Cost.Value {
				continue
			}
			// Skip uncharged recharge actions
			if act.RechargeValue > 0 && a.StateManager.Resource[act.Name] == 0 {
				continue
			}
			availableActions = append(availableActions, act)
		}
	}

	// Filter spell actions for optimal upcasting/resource expenditure if using weighted AI
	useWeighted := ed.SimOptions != nil && ed.SimOptions.UseWeightedAI
	primaryTarget := ed.Actors[primaryTargetID]
	shouldExpend := aid.ShouldExpendHighResource(a, primaryTarget)

	filteredSpellActions := make([]core.Action, 0)
	for _, act := range a.SpellActions {
		if act.Cost.ActivationType == costType && (act.AverageDamage > 0 || act.HasDC) {
			// Skip if no spell slots or innate uses
			if act.CastLevel > 0 {
				if act.IsInnate {
					if a.StateManager.InnateCurrent[act.Name] <= 0 {
						continue
					}
				} else if a.StateManager.CurrentSlots[act.CastLevel] <= 0 {
					continue
				}
			}

			// If weighted AI is active, decide if this spell level is appropriate
			if useWeighted && act.CastLevel > 0 {
				// Get minimum level for this spell (some spells might be added to SpellActions multiple times for different levels)
				// We want to avoid using high level slots for weak targets unless AlwaysUpcast is on.
				alwaysUpcast := (a.IsCharacter() && ed.SimOptions.CharactersAlwaysUpcast) || (a.IsMonster() && ed.SimOptions.MonstersAlwaysUpcast)

				if !alwaysUpcast {
					// Logic: If we shouldn't expend high resources, only consider the base level of the spell
					// We'll find the lowest level for this spell name in SpellActions
					lowestLvl := 9
					for _, other := range a.SpellActions {
						baseName := other.Name
						if idx := strings.Index(baseName, " ("); idx != -1 {
							baseName = baseName[:idx]
						}
						currentBaseName := act.Name
						if idx := strings.Index(currentBaseName, " ("); idx != -1 {
							currentBaseName = currentBaseName[:idx]
						}

						if baseName == currentBaseName {
							if other.CastLevel > 0 && other.CastLevel < lowestLvl {
								lowestLvl = other.CastLevel
							}
						}
					}

					if !shouldExpend && act.CastLevel > lowestLvl {
						continue // Conserve high level slots
					}
				}
			}

			filteredSpellActions = append(filteredSpellActions, act)
		}
	}

	// Update availableActions with filtered spell actions
	availableActions = append(availableActions, filteredSpellActions...)

	if len(availableActions) == 0 {
		return nil, targetIDs
	}

	threshold := 3 // Default threshold
	if ed.SimOptions != nil && ed.SimOptions.AOETargetThreshold > 0 {
		threshold = ed.SimOptions.AOETargetThreshold
	}

	if policy == core.ActionPolicyPriority {
		// Priority 1: Recharge actions
		var rechargeActions []core.Action
		for _, act := range availableActions {
			if act.RechargeValue > 0 {
				rechargeActions = append(rechargeActions, act)
			}
		}
		if len(rechargeActions) > 0 {
			// Sort for determinism
			sort.Slice(rechargeActions, func(i, j int) bool {
				return rechargeActions[i].Name < rechargeActions[j].Name
			})
			// Pick random for tie-breaking
			idx := aid.rng.IntN(len(rechargeActions))
			actionCopy := rechargeActions[idx]

			if actionCopy.IsAOE {
				targetIDs := ed.Adjudicator.IdentifyAOETargets(a, primaryTargetID, false)
				return &actionCopy, targetIDs
			}

			return &actionCopy, []int{primaryTargetID}
		}

		// Priority 2: AOE (if threshold met)
		var aoeActions []core.Action
		for _, act := range availableActions {
			if act.IsAOE {
				aoeActions = append(aoeActions, act)
			}
		}
		if len(aoeActions) > 0 {
			// Check if any AOE hits enough targets
			for range aoeActions {
				targetIDs := ed.Adjudicator.IdentifyAOETargets(a, primaryTargetID, false)
				if len(targetIDs) >= threshold {
					// Found an AOE that meets threshold.
					// In Priority mode, we'll pick the first one found (after sort)
					sort.Slice(aoeActions, func(i, j int) bool {
						return aoeActions[i].Name < aoeActions[j].Name
					})
					// Actually, let's re-identify targets for the chosen one
					actionCopy := aoeActions[0]
					targetIDs = ed.Adjudicator.IdentifyAOETargets(a, primaryTargetID, false)
					return &actionCopy, targetIDs
				}
			}
		}

		// Priority 3: Multiattack
		var multiattackActions []core.Action
		for _, act := range availableActions {
			if act.ActionType == core.ATMultiAttack {
				multiattackActions = append(multiattackActions, act)
			}
		}
		if len(multiattackActions) > 0 {
			// Sort for determinism
			sort.Slice(multiattackActions, func(i, j int) bool {
				return multiattackActions[i].Name < multiattackActions[j].Name
			})
			// Pick random for tie-breaking
			idx := aid.rng.IntN(len(multiattackActions))
			actionCopy := multiattackActions[idx]
			return &actionCopy, []int{primaryTargetID}
		}

		// Priority 4: Normal actions - Pick random in sorted list for variance
		// Sort for determinism
		sort.Slice(availableActions, func(i, j int) bool {
			return availableActions[i].Name < availableActions[j].Name
		})
		idx := aid.rng.IntN(len(availableActions))
		actionCopy := availableActions[idx]

		if actionCopy.IsAOE {
			targetIDs := ed.Adjudicator.IdentifyAOETargets(a, primaryTargetID, false)
			return &actionCopy, targetIDs
		}

		return &actionCopy, []int{primaryTargetID}
	}

	if policy == core.ActionPolicyRandom {
		// Sort for determinism
		sort.Slice(availableActions, func(i, j int) bool {
			return availableActions[i].Name < availableActions[j].Name
		})
		idx := aid.rng.IntN(len(availableActions))
		actionCopy := availableActions[idx]

		if actionCopy.IsAOE {
			targetIDs := ed.Adjudicator.IdentifyAOETargets(a, primaryTargetID, false)
			return &actionCopy, targetIDs
		}

		return &actionCopy, []int{primaryTargetID}
	}

	// Default/Highest Damage policy
	// For AOE, we calculate total expected damage: AverageDamage * numTargets
	var bestActions []core.Action
	var bestActionTargetIDs [][]int
	maxDmg := -1.0

	for _, action := range availableActions {
		dmg := float64(action.AverageDamage)
		var currentTargetIDs []int
		if action.IsAOE {
			currentTargetIDs = ed.Adjudicator.IdentifyAOETargets(a, primaryTargetID, false)
			dmg *= float64(len(currentTargetIDs))
		} else {
			currentTargetIDs = []int{primaryTargetID}
		}

		if dmg > maxDmg {
			maxDmg = dmg
			bestActions = []core.Action{action}
			bestActionTargetIDs = [][]int{currentTargetIDs}
		} else if dmg == maxDmg {
			bestActions = append(bestActions, action)
			bestActionTargetIDs = append(bestActionTargetIDs, currentTargetIDs)
		}
	}

	if len(bestActions) == 0 {
		return nil, []int{primaryTargetID}
	}

	// Pick random for tie-breaking
	idx := aid.rng.IntN(len(bestActions))
	actionCopy := bestActions[idx]
	return &actionCopy, bestActionTargetIDs[idx]
}

// chooseBestHealingAction selects the best healing action.
func (aid *AIDirector) chooseBestHealingAction(a *actor.Actor, costType core.ActivationType, ed *EncounterDirector) (*core.Action, []int) {
	var bestAction *core.Action
	bestScore := -1.0
	var bestTargetIDs []int

	hpDiff := ed.Statistics.GetHPDiffValuePerSide(a.Side)
	if len(hpDiff) == 0 {
		return nil, nil
	}

	// Combine available heals
	availableHeals := make([]core.Action, 0)
	for _, act := range a.Actions {
		if act.Cost.ActivationType == costType && act.AverageHeal > 0 {
			// Skip uncharged recharge actions
			if act.RechargeValue > 0 && a.StateManager.Resource[act.Name] == 0 {
				continue
			}
			availableHeals = append(availableHeals, act)
		}
	}

	// Filter spell actions for optimal upcasting if using weighted AI
	useWeighted := ed.SimOptions != nil && ed.SimOptions.UseWeightedAI

	for _, act := range a.SpellActions {
		if act.Cost.ActivationType == costType && act.AverageHeal > 0 {
			// Skip if no spell slots or innate uses
			if act.CastLevel > 0 {
				if act.IsInnate {
					if a.StateManager.InnateCurrent[act.Name] <= 0 {
						continue
					}
				} else if a.StateManager.CurrentSlots[act.CastLevel] <= 0 {
					continue
				}
			}

			// If weighted AI is active, decide if this spell level is appropriate
			// For healing, "ShouldExpend" usually depends on how much HP is missing
			if useWeighted && act.CastLevel > 0 {
				alwaysUpcast := (a.IsCharacter() && ed.SimOptions.CharactersAlwaysUpcast) || (a.IsMonster() && ed.SimOptions.MonstersAlwaysUpcast)

				if !alwaysUpcast {
					// Logic: Don't upcast if the base level heal already covers most of the missing HP.
					// We'll compare this action's heal to the target's missing HP later in the loop,
					// but here we can at least filter out extremely wasteful upcasts.
					// To be simple, we'll keep all levels and let the scoring logic (which penalizes over-healing)
					// handle it.
				}
			}

			availableHeals = append(availableHeals, act)
		}
	}

	if len(availableHeals) == 0 {
		return nil, nil
	}

	// Check for the most effective heal action
	targetIDs := make([]int, 0, len(hpDiff))
	for id := range hpDiff {
		targetIDs = append(targetIDs, id)
	}
	sort.Ints(targetIDs)

	// Sort availableHeals for determinism
	sort.Slice(availableHeals, func(i, j int) bool {
		return availableHeals[i].Name < availableHeals[j].Name
	})

	for _, targetID := range targetIDs {
		missingHP := hpDiff[targetID]
		targetActor := ed.Actors[targetID]

		for _, action := range availableHeals {
			// Calculate actual restoration
			effectiveHeal := action.AverageHeal
			overHeal := 0
			if effectiveHeal > missingHP {
				overHeal = action.AverageHeal - missingHP
				effectiveHeal = missingHP
			}

			// Primary value is HP restored, secondary penalty for waste, multiplied by emergency need
			hpFactor := core.CalculateHPFactor(targetActor.StateManager.CurrentHP, targetActor.StateManager.MaxHP, core.HPVisible)

			// Weight effective heal highly, penalize waste moderately
			score := float64(effectiveHeal) - (float64(overHeal) * 0.5)

			// Apply multiplier for critical targets (Danger/Emergency)
			// Ensure healing is actually attractive by boosting it significantly for low HP targets
			score *= 1.0 + (hpFactor * 100.0)

			if score > bestScore {
				bestScore = score
				actionCopy := action
				bestAction = &actionCopy
				bestTargetIDs = []int{targetID}
			}
		}
	}
	if bestAction != nil {
		return bestAction, bestTargetIDs
	}
	return nil, nil
}

func (aid *AIDirector) hasHealingAbilities(a *actor.Actor) bool {
	for _, act := range a.Actions {
		if act.AverageHeal > 0 {
			if act.RechargeValue > 0 && a.StateManager.Resource[act.Name] == 0 {
				continue
			}
			return true
		}
	}
	// Check spell actions too
	for _, act := range a.SpellActions {
		if act.AverageHeal > 0 {
			if act.CastLevel == 0 {
				return true
			}
			if act.IsInnate {
				if a.StateManager.InnateCurrent[act.Name] > 0 {
					return true
				}
			} else if a.StateManager.CurrentSlots[act.CastLevel] > 0 {
				return true
			}
		}
	}
	return false
}

// SelectTarget picks the best target of the given type.
func (aid *AIDirector) SelectTarget(a *actor.Actor, targets map[int]*actor.Actor, targetType core.TargetType, ed *EncounterDirector) []int {
	if len(targets) == 0 {
		return nil
	}

	var targetID int
	if ed.SimOptions != nil && ed.SimOptions.UseWeightedAI {
		targetID = aid.selectTargetWeighted(a, targets, targetType, ed)
	} else {
		targetID = aid.selectTargetByPriority(a, targets, ed)
	}

	if targetID == -1 {
		return nil
	}
	return []int{targetID}
}

// selectTargetByPriority selects a target from a list of candidates based on the actor’s target priority behavior.
func (aid *AIDirector) selectTargetByPriority(a *actor.Actor, targets map[int]*actor.Actor, ed *EncounterDirector) int {
	if len(targets) == 0 {
		return -1
	}

	priority := a.Behavior.TargetPriority

	// Determine visibility
	visibility := core.HPStatusHidden
	if ed.SimOptions != nil {
		visibility = ed.SimOptions.HPVisibilityMode
	}

	// Sort IDs for deterministic tiebreaking
	ids := make([]int, 0, len(targets))
	for id := range targets {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	targetID := -1
	if priority != core.PriorityNone && priority != "" {
		targetID = aid.selectTargetBySpecificPriority(a, targets, ids, priority, visibility)
	}

	if targetID != -1 {
		return targetID
	}

	// Fallback to secondary priority if it exists
	secondary := a.Behavior.SecondaryTargetPriority
	if secondary != "" && secondary != core.PriorityNone && secondary != priority {
		targetID = aid.selectTargetBySpecificPriority(a, targets, ids, secondary, visibility)
		if targetID != -1 {
			return targetID
		}
	}

	// Final fallback: random
	return aid.selectTargetSimple(targets)
}

func (aid *AIDirector) selectTargetBySpecificPriority(a *actor.Actor, targets map[int]*actor.Actor, ids []int, priority core.TargetPriority, visibility core.HPVisibilityMode) int {
	switch priority {
	case core.PriorityLowestHP:
		return aid.selectLowestHPTarget(targets, ids, visibility)
	case core.PriorityMostDamaged:
		return aid.selectMostDamagedTarget(targets, ids, visibility)
	case core.PriorityLeastDamaged:
		return aid.selectLeastDamagedTarget(targets, ids, visibility)
	case core.PrioritySpellcaster:
		return aid.selectSpellcaster(targets, ids)
	case core.PriorityHealer:
		return aid.selectHealer(targets, ids)
	case core.PriorityHighestMaxHP:
		return aid.selectHighestMaxHP(targets, ids, visibility)
	case core.PriorityLowestMaxHP:
		return aid.selectLowestMaxHP(targets, ids, visibility)
	case core.PriorityHighestScaler:
		return aid.selectHighestScaler(a, targets, ids)
	case core.PriorityLowestScaler:
		return aid.selectLowestScaler(a, targets, ids)
	default:
		return -1
	}
}

// selectSpellcaster identifies and selects the first spellcaster from the given actor targets based on provided IDs.
func (aid *AIDirector) selectSpellcaster(targets map[int]*actor.Actor, ids []int) int {
	bestID := -1

	for _, id := range ids {
		t := targets[id]
		if t.IsSpellcaster() {
			bestID = id
			break
		}
	}

	if bestID != -1 {
		return bestID
	}

	return aid.selectTargetSimple(targets)
}

// selectHealer selects and returns the ID of a healer from the provided list of target IDs, or -1 if no healer is found.
func (aid *AIDirector) selectHealer(targets map[int]*actor.Actor, ids []int) int {
	bestID := -1

	for _, id := range ids {
		t := targets[id]
		if t.IsHealer() {
			bestID = id
			break
		}
	}

	if bestID != -1 {
		return bestID
	}

	return aid.selectTargetSimple(targets)
}

// selectHighestScaler identifies and returns the ID of the target with the highest scaling property (CR or Level) from the given IDs.
func (aid *AIDirector) selectHighestScaler(a *actor.Actor, targets map[int]*actor.Actor, ids []int) int {
	bestID := -1

	bestScaler := float64(-1)
	for _, id := range ids {
		v := targets[id].GetLevelOrCR()
		if v > bestScaler {
			bestScaler = v
			bestID = id
		}
	}
	return bestID
}

// selectLowestScaler selects the ID of the actor with the lowest challenge rating (CR) or level from the given list of IDs.
func (aid *AIDirector) selectLowestScaler(a *actor.Actor, targets map[int]*actor.Actor, ids []int) int {
	bestID := -1

	bestScaler := float64(core.SIM_BIG_NUMBER)
	for _, id := range ids {
		v := targets[id].GetLevelOrCR()
		if v < bestScaler {
			bestScaler = v
			bestID = id
		}
	}
	return bestID
}

// selectHighestMaxHP selects the actor with the highest MaxHP from the given ids based on the specified visibility mode.
func (aid *AIDirector) selectHighestMaxHP(targets map[int]*actor.Actor, ids []int, visibility core.HPVisibilityMode) int {
	bestID := -1

	if visibility == core.HPVisible {
		highestHP := -1
		for _, id := range ids {
			t := targets[id]
			if t.StateManager.MaxHP > highestHP {
				highestHP = t.StateManager.MaxHP
				bestID = id
			}
		}

	} else {
		// If we can't see HP, pick a random target
		return aid.selectTargetSimple(targets)
	}

	return bestID
}

// selectLowestMaxHP selects the target with the lowest maximum HP from a given list of target IDs if HP is visible.
// If HP visibility is off, a random target is selected instead. Returns the selected target's ID or -1 if no target.
func (aid *AIDirector) selectLowestMaxHP(targets map[int]*actor.Actor, ids []int, visibility core.HPVisibilityMode) int {
	bestID := -1

	if visibility == core.HPVisible {
		lowestHP := core.SIM_BIG_NUMBER
		for _, id := range ids {
			t := targets[id]
			if t.StateManager.MaxHP < lowestHP {
				lowestHP = t.StateManager.MaxHP
				bestID = id
			}
		}

	} else {
		// If we can't see HP, pick a random target
		return aid.selectTargetSimple(targets)
	}

	return bestID
}

// selectLowestHPTarget determines the target with the lowest HP or worst health state based on visibility mode and selection criteria.
func (aid *AIDirector) selectLowestHPTarget(targets map[int]*actor.Actor, ids []int, visibility core.HPVisibilityMode) int {
	bestID := -1

	if visibility == core.HPVisible {
		minHP := core.SIM_BIG_NUMBER
		for _, id := range ids {
			t := targets[id]
			if t.StateManager.CurrentHP < minHP {
				minHP = t.StateManager.CurrentHP
				bestID = id
			}
		}
	} else {
		// Use HealthState
		stateValue := func(s core.HealthState) int {
			switch s {
			case core.HealthStateCritical:
				return 1
			case core.HealthStateBloody:
				return 2
			case core.HealthStateWounded:
				return 3
			case core.HealthStateHealthy:
				return 4
			default:
				return 5
			}
		}

		minVal := core.SIM_BIG_NUMBER
		for _, id := range ids {
			t := targets[id]
			val := stateValue(t.StateManager.HealthState)
			if val < minVal {
				minVal = val
				bestID = id
			}
		}
	}
	return bestID
}

// selectMostDamagedTarget selects the target with the highest damage taken based on visible health or lowest HP if not visible.
func (aid *AIDirector) selectMostDamagedTarget(targets map[int]*actor.Actor, ids []int, visibility core.HPVisibilityMode) int {
	bestID := -1

	if visibility == core.HPVisible {
		maxDiff := -1
		for _, id := range ids {
			t := targets[id]
			diff := t.StateManager.MaxHP - t.StateManager.CurrentHP
			if diff > maxDiff {
				maxDiff = diff
				bestID = id
			}
		}
	} else {
		// If visibility is off, most damage == lowest hp
		return aid.selectLowestHPTarget(targets, ids, visibility)
	}
	return bestID
}

// selectLeastDamagedTarget selects the least damaged target from a set of targets based on HP or health state visibility mode.
func (aid *AIDirector) selectLeastDamagedTarget(targets map[int]*actor.Actor, ids []int, visibility core.HPVisibilityMode) int {
	bestID := -1

	if visibility == core.HPVisible {
		minDiff := core.SIM_BIG_NUMBER
		for _, id := range ids {
			t := targets[id]
			diff := t.StateManager.MaxHP - t.StateManager.CurrentHP
			if diff < minDiff {
				minDiff = diff
				bestID = id
			}
		}
	} else {
		// Use HealthState - pick the one with highest state value (healthiest)
		stateValue := func(s core.HealthState) int {
			switch s {
			case core.HealthStateCritical:
				return 1
			case core.HealthStateBloody:
				return 2
			case core.HealthStateWounded:
				return 3
			case core.HealthStateHealthy:
				return 4
			default:
				return 0
			}
		}

		maxVal := -1
		for _, id := range ids {
			t := targets[id]
			val := stateValue(t.StateManager.HealthState)
			if val > maxVal {
				maxVal = val
				bestID = id
			}
		}
	}
	return bestID
}

// selectTargetSimple selects a random target ID from the given map of targets.
func (aid *AIDirector) selectTargetSimple(targets map[int]*actor.Actor) int {
	if len(targets) == 0 {
		return -1
	}
	// Pick a random target from the map (deterministic order for RNG consumption)
	ids := make([]int, 0, len(targets))
	for id := range targets {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids[aid.rng.IntN(len(ids))]
}

// selectTargetDeterministic selects the first target ID from the sorted list of targets.
func (aid *AIDirector) selectTargetDeterministic(targets map[int]*actor.Actor) int {
	if len(targets) == 0 {
		return -1
	}
	ids := make([]int, 0, len(targets))
	for id := range targets {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids[0]
}

// GetUtilityWeights returns the utility weights for the given actor, providing defaults if none are set.
func (aid *AIDirector) GetUtilityWeights(a *actor.Actor) core.UtilityWeights {
	if a.Behavior.Weights != nil {
		return *a.Behavior.Weights
	}

	return core.UtilityWeights{
		TargetFactorWeights: core.TargetFactorWeights{
			HighThreat:         1.0,
			TargetPotency:      1.0,
			TargetHitability:   1.0,
			Vengeance:          0.5,
			LowHP:              1.5,
			CasterPriority:     1.2,
			ConcentrationBreak: 1.2,
			ElitePriority:      0.8,
			EmergencyHeal:      2.0,
		},
		ResourceExpenditureWeight: 1.0,
	}
}

// ShouldExpendHighResource determines if an actor should expend high-level resources (spell slots, smites)
// against a specific target based on its potency and the actor's weights.
func (aid *AIDirector) ShouldExpendHighResource(a *actor.Actor, target *actor.Actor) bool {
	if target == nil {
		return false
	}

	weights := aid.GetUtilityWeights(a)
	potency := core.CalculatePotencyFactor(target.AC, target.ProficiencyBonus, target.Metadata.AverageOffensiveValue, target.Metadata.HighestOffensiveValue)

	// Threshold logic: A target with potency > 0.8 is considered "Elite" or "Dangerous"
	// and worthy of high-level resources. This is modified by the ResourceExpenditureWeight.
	threshold := 0.8 / weights.ResourceExpenditureWeight

	// Also consider if the target is a boss (Legendary)
	if target.Metadata.IsLegendary {
		potency += 0.3
	}

	return potency >= threshold
}

// selectTargetWeighted selects the most suitable target from a set of candidates based on weighted scoring criteria.
func (aid *AIDirector) selectTargetWeighted(a *actor.Actor, targets map[int]*actor.Actor, targetType core.TargetType, ed *EncounterDirector) int {
	// Authorization check for weighted AI
	authorized := true
	if !authorized {
		return aid.selectTargetSimple(targets)
	}

	// Default weights if none are provided
	weights := aid.GetUtilityWeights(a)

	bestID := -1
	bestScore := -1e18

	// Sort IDs for deterministic behavior
	ids := make([]int, 0, len(targets))
	for id := range targets {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	avgEnemyDamage := ed.CalculateAvgEnemyDamage(a)
	maxDamageSeen := 0
	if ed.Statistics != nil {
		for _, stat := range ed.Statistics.statistics {
			if stat.LastDamageDealt > maxDamageSeen {
				maxDamageSeen = stat.LastDamageDealt
			}
		}
	}

	for _, id := range ids {
		target := targets[id]
		score := 0.0

		priority := a.Behavior.TargetPriority

		if targetType == core.TTDamage {
			// 1. Hitability
			// We use the actor's primary attack bonus for a general assessment
			attackBonus := a.ProficiencyBonus + a.Abilities.GetAbilityModifier(core.AbilityStrength)
			if a.Metadata.HighestOffensiveValue > float64(attackBonus) {
				// If Metadata has a precalculated bonus (e.g. from Equipment), use that
				// Actually Metadata.HighestOffensiveValue is damage, not attack bonus.
				// We don't have a precalculated attack bonus in Metadata yet.
			}
			hitability := core.CalculateHitabilityFactor(target.AC, attackBonus)
			score += hitability * weights.TargetFactorWeights.TargetHitability

			// 2. Potency (Target's offensive capability)
			potency := core.CalculatePotencyFactor(target.AC, target.ProficiencyBonus, target.Metadata.AverageOffensiveValue, target.Metadata.HighestOffensiveValue)
			score += potency * weights.TargetFactorWeights.TargetPotency

			// 3. Low HP
			hpFactor := core.CalculateHPFactor(target.StateManager.CurrentHP, target.StateManager.MaxHP, ed.SimOptions.HPVisibilityMode)
			score += hpFactor * weights.TargetFactorWeights.LowHP

			// 4. Elite Priority
			if target.Metadata.IsLegendary {
				score += weights.TargetFactorWeights.ElitePriority
			}

			// 5. Caster/Healer Priority
			if (priority == core.PrioritySpellcaster && target.Metadata.SpellcasterMetadata.IsSpellcaster) ||
				(priority == core.PriorityHealer && target.Metadata.SpellcasterMetadata.IsSpellcaster) { // Healer check is similar for now
				score += 2.0
			}

			// 6. Vengeance (Did they hit ME last?)
			if ed.Statistics != nil {
				myStats := ed.Statistics.GetCombatant(a.InstanceID)
				if myStats != nil && myStats.LastAttackerID == target.InstanceID {
					score += weights.TargetFactorWeights.Vengeance
				}
			}

			// 7. High Threat (Who is dealing the most damage overall?)
			if ed.Statistics != nil && maxDamageSeen > 0 {
				targetStats := ed.Statistics.GetCombatant(target.InstanceID)
				if targetStats != nil {
					threatFactor := float64(targetStats.LastDamageDealt) / float64(maxDamageSeen)
					score += threatFactor * weights.TargetFactorWeights.HighThreat
				}
			}

		} else if targetType == core.TTHealing {
			// Emergency Heal
			emergency := core.CalculateEmergencyHealFactor(target.StateManager.CurrentHP, avgEnemyDamage)
			score += emergency * weights.TargetFactorWeights.EmergencyHeal

			// Priority registries
			if ed.Statistics != nil {
				for _, needsHealID := range ed.Statistics.NeedsEmergencyHealing {
					if needsHealID == id {
						score += 20.0
						break
					}
				}
				for _, needsHealID := range ed.Statistics.NeedsHealing {
					if needsHealID == id {
						score += 10.0
						break
					}
				}
			}
		}

		// Add Noise if enabled
		if ed.SimOptions.EnableMonsterNoise && ed.SimOptions.MonsterNoiseWeight > 0 {
			noise := (aid.rng.Float64()*2.0 - 1.0) * ed.SimOptions.MonsterNoiseWeight
			score += noise
		}

		if score > bestScore {
			bestScore = score
			bestID = id
		}
	}

	return bestID
}
