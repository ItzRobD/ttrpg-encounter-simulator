package core

import (
	"fmt"
	"math/rand/v2"
)

// SelectTargetFromMap chooses a target based on the given priority.
// It returns a TargetStatus to distinguish "no targets" from actual errors.
func SelectTargetFromMap(validTargets map[int]*Combatant, priority TargetPriority, rng *rand.Rand) (TargetStatus, int, error) {
	if len(validTargets) == 0 {
		return TargetNone, -1, nil
	}
	targetID := -1
	switch priority {
	case NoPriority:
		targetIDs := make([]int, 0, len(validTargets))
		for id := range validTargets {
			targetIDs = append(targetIDs, id)
		}
		targetID = targetIDs[rng.IntN(len(targetIDs))]
	case PrioritizeHighestLevel:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetLevel() > validTargets[targetID].GetEntity().GetLevel() {
				targetID = id
			}
		}
	case PrioritizeLowestLevel:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetLevel() < validTargets[targetID].GetEntity().GetLevel() {
				targetID = id
			}
		}
	case PrioritizeMostDamaged:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetHPStatus().GetHPPct() < validTargets[targetID].GetEntity().GetHPStatus().GetHPPct() {
				targetID = id
			}
		}
	case PrioritizeLeastDamaged:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetHPStatus().GetHPPct() > validTargets[targetID].GetEntity().GetHPStatus().GetHPPct() {
				targetID = id
			}
		}
	case PrioritizeLowestHealth:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetHPStatus().GetHP() < validTargets[targetID].GetEntity().GetHPStatus().GetHP() {
				targetID = id
			}
		}
	case PrioritizeHighestMaxHP:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetHPStatus().GetMaxHP() > validTargets[targetID].GetEntity().GetHPStatus().GetMaxHP() {
				targetID = id
			}
		}
	case PrioritizeLowestMaxHP:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetHPStatus().GetMaxHP() < validTargets[targetID].GetEntity().GetHPStatus().GetMaxHP() {
				targetID = id
			}
		}
	case PrioritizeHealer:
		targets := make(map[int]*Combatant)
		for id, c := range validTargets {
			if c.GetEntity().IsSpellcaster() && c.GetEntity().IsHealer() {
				targets[id] = c
			}
		}
		if len(targets) > 0 {
			return SelectTargetFromMap(targets, PrioritizeMostDamaged, rng)
		}
	case PrioritizeSpellcaster:
		targets := make(map[int]*Combatant)
		for id, c := range validTargets {
			if c.GetEntity().IsSpellcaster() {
				targets[id] = c
			}
		}
		if len(targets) > 0 {
			return SelectTargetFromMap(targets, PrioritizeMostDamaged, rng)
		}
	case PrioritizeUnconscious:
		targets := make(map[int]*Combatant)
		for id, c := range validTargets {
			if c.GetEntity().IsUnconscious() {
				targets[id] = c
			}
		}
		if len(targets) > 0 {
			return SelectTargetFromMap(targets, PrioritizeHighestMaxHP, rng)
		}
	default:
		return TargetInvalidType, -1, fmt.Errorf("unknown target prioritization strategy")
	}
	if targetID == -1 {
		return TargetNone, -1, nil
	}
	return TargetOK, targetID, nil
}
