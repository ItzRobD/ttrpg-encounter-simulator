package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	"math/rand/v2"
)

type TargetType int

const (
	Damage TargetType = iota
	Healing
)

type AIManager struct {
}

func SelectTargetFromMap(validTargets map[int]core.Combatant, priority core.TargetPriority, rng *rand.Rand) (int, error) {
	if len(validTargets) == 0 {
		return -1, fmt.Errorf("no valid targets found")
	}
	targetID := -1
	switch priority {
	case core.NoPriority:
		targetIDs := make([]int, 0, len(validTargets))
		for id := range validTargets {
			targetIDs = append(targetIDs, id)
		}
		targetID = targetIDs[rng.IntN(len(targetIDs))]
	case core.PrioritizeHighestLevel:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetLevel() > validTargets[targetID].GetEntity().GetLevel() {
				targetID = id
			}
		}
	case core.PrioritizeLowestLevel:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetLevel() < validTargets[targetID].GetEntity().GetLevel() {
				targetID = id
			}
		}
	case core.PrioritizeMostDamaged:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetHPStatus().GetHPPct() < validTargets[targetID].GetEntity().GetHPStatus().GetHPPct() {
				targetID = id
			}
		}
	case core.PrioritizeLeastDamaged:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetHPStatus().GetHPPct() > validTargets[targetID].GetEntity().GetHPStatus().GetHPPct() {
				targetID = id
			}
		}
	case core.PrioritizeLowestHealth:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetHPStatus().GetHP() < validTargets[targetID].GetEntity().GetHPStatus().GetHP() {
				targetID = id
			}
		}
	case core.PrioritizeHighestMaxHP:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetHPStatus().GetMaxHP() > validTargets[targetID].GetEntity().GetHPStatus().GetMaxHP() {
				targetID = id
			}
		}
	case core.PrioritizeLowestMaxHP:
		for id, c := range validTargets {
			if targetID == -1 || c.GetEntity().GetHPStatus().GetMaxHP() < validTargets[targetID].GetEntity().GetHPStatus().GetMaxHP() {
				targetID = id
			}
		}
	case core.PrioritizeHealer:
		var targets map[int]core.Combatant = make(map[int]core.Combatant)
		for id, c := range validTargets {
			if c.GetEntity().IsSpellcaster() {
				if c.GetEntity().IsHealer() {
					targets[id] = c
				}
			}
		}
		if len(targets) > 0 {
			return SelectTargetFromMap(targets, core.PrioritizeMostDamaged, rng)
		}
	case core.PrioritizeSpellcaster:
		var targets map[int]core.Combatant = make(map[int]core.Combatant)
		for id, c := range validTargets {
			if c.GetEntity().IsSpellcaster() {
				targets[id] = c
			}
		}
		if len(targets) > 0 {
			return SelectTargetFromMap(targets, core.PrioritizeMostDamaged, rng)
		}
	default:
		return targetID, fmt.Errorf("unknown target prioritization strategy")
	}
	if targetID == -1 {
		return targetID, fmt.Errorf("no valid target found")
	}
	return targetID, nil
}
