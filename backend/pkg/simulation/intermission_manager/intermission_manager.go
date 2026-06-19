package intermission_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"maps"
	"math"
	"slices"
)

type IntermissionOptions struct {
	MaxShortRests          int     `json:"max_short_rests"`
	ShortRestHealThreshold float64 `json:"short_rest_heal_threshold"` // e.g., 0.7 (70%)
	PostRestHealThreshold  float64 `json:"post_rest_heal_threshold"`
}

type IntermissionManager struct {
	ShortRestsTaken int
	RollManager     *roll_manager.RollManager
}

func NewIntermissionManager(rm *roll_manager.RollManager) *IntermissionManager {
	return &IntermissionManager{
		RollManager: rm,
	}
}

type IntermissionResult struct {
	ShortRestsTaken   int
	HealingReceived   map[int]int
	HitDiceUsed       map[int]map[core.DiceType]int
	SpellsUsed        map[int]int
	SpellSlotsUsed    map[int]map[int]int
	HealerInstanceIDs []int // IDs of actors who performed healing
}

// ProcessIntermission manages the intermission phase for a party, handling short rests and post-combat healing operations.
// Returns: an IntermissionResult containing detailed statistics
func (im *IntermissionManager) ProcessIntermission(party []*actor.Actor, options IntermissionOptions) IntermissionResult {
	result := IntermissionResult{
		HealingReceived: make(map[int]int),
		HitDiceUsed:     make(map[int]map[core.DiceType]int),
		SpellsUsed:      make(map[int]int),
		SpellSlotsUsed:  make(map[int]map[int]int),
	}

	// Determine if the party should short rest, always before expending resources
	if im.ShortRestsTaken < options.MaxShortRests {
		if im.shouldTakeShortRest(party, options.ShortRestHealThreshold) {
			srHealing, hdUsedByActor := im.PerformShortRest(party)
			im.ShortRestsTaken++

			for id, heal := range srHealing {
				result.HealingReceived[id] += heal
			}
			for id, hd := range hdUsedByActor {
				result.HitDiceUsed[id] = hd
			}
		}
	}

	// Post combat healing using resources
	pcHealing, spellsUsed, slotsUsed := im.performPostCombatHealing(party, options.PostRestHealThreshold)

	for id, heal := range pcHealing {
		result.HealingReceived[id] += heal
	}
	for id, count := range spellsUsed {
		result.SpellsUsed[id] += count
	}
	for id, slots := range slotsUsed {
		result.SpellSlotsUsed[id] = slots
	}

	result.ShortRestsTaken = im.ShortRestsTaken
	return result
}

func (im *IntermissionManager) shouldTakeShortRest(party []*actor.Actor, threshold float64) bool {
	for _, member := range party {
		if member.StateManager.CurrentHP < int(float64(member.StateManager.MaxHP)*threshold) {
			return true
		}
	}
	return false
}

// PerformShortRest handles short rest recovery for the party, restoring hit points and available short rest resources.
// Returns a map of healing done by actor ID and a map of hit dice expended by type per actor.
func (im *IntermissionManager) PerformShortRest(party []*actor.Actor) (map[int]int, map[int]map[core.DiceType]int) {
	srHealing := make(map[int]int)
	hdUsedByActor := make(map[int]map[core.DiceType]int)

	for _, actor := range party {
		// Recover Hit Dice based recovery (spending hit dice)
		hdHealAmt, hdUsed := im.spendHitDice(actor)
		srHealing[actor.InstanceID] += hdHealAmt
		if len(hdUsed) > 0 {
			hdUsedByActor[actor.InstanceID] = hdUsed
		}

		// Recover Short Rest resources
		im.recoverShortRestResources(actor)
	}

	return srHealing, hdUsedByActor
}

func (im *IntermissionManager) spendHitDice(a *actor.Actor) (int, map[core.DiceType]int) {
	sm := &a.StateManager
	if len(sm.CurrentHitDice) == 0 {
		return 0, nil
	}

	// Sort highest dice to lowest
	diceTypes := slices.Collect(maps.Keys(sm.CurrentHitDice))
	slices.SortFunc(diceTypes, func(a, b core.DiceType) int {
		return int(b) - int(a) // Largest first
	})

	conMod := a.Abilities.GetAbilityModifier(core.AbilityConstitution)

	totalHealingDone := 0
	hitDiceUsed := make(map[core.DiceType]int)
	for _, die := range diceTypes {
		for sm.CurrentHitDice[die] > 0 && sm.CurrentHP < sm.MaxHP {
			res := im.RollManager.RollDice(1, die, roll_manager.RollOptions{})
			healAmount := res.Total + conMod
			if healAmount < 1 {
				healAmount = 1
			}
			hpModRes := sm.ModifyHP(healAmount, false, a.ActorType == core.ActorTypeCharacter)
			sm.CurrentHitDice[die]--
			totalHealingDone += hpModRes.ModificationValue
			hitDiceUsed[die]++
		}
	}

	return totalHealingDone, hitDiceUsed
}

func (im *IntermissionManager) recoverShortRestResources(a *actor.Actor) {
	sm := &a.StateManager

	// Warlock Spell Slots recovery
	if core.ClassID(a.Metadata.ClassID) == core.Warlock {
		for k, v := range sm.MaxSlots {
			sm.CurrentSlots[k] = v
		}
	}

	// Wizard Spell Slots Recovery
	// Rule: Once per day, recover a combined level that is <= (wizard level / 2) rounded up, slot level <= 5
	// We should prioritize the highest spell slots first
	if core.ClassID(a.Metadata.ClassID) == core.Wizard && a.StateManager.Resource["Arcane Recovery"] == 1 {
		maxSlotsToRecover := int(math.Ceil(float64(a.Metadata.Level) / 2))
		for slotLevel := 5; slotLevel >= 1; slotLevel-- {
			if sm.CurrentSlots[slotLevel] < sm.MaxSlots[slotLevel] && maxSlotsToRecover > 0 {
				sm.CurrentSlots[slotLevel]++
				maxSlotsToRecover--
				a.StateManager.Resource["Arcane Recovery"] = 0
			}
		}
	}

	// Recover features that reset on a short rest
	for _, f := range a.Features {
		if slices.Contains(sm.ShortRestRecoveredFeatures, string(f.Name)) {
			// Logic to recover the resource based on the feature name
			switch f.Name {
			case core.SpecAbilityBreathWeapon:
				sm.Resource[string(f.Name)] = 1
			case core.SpecAbilityRelentlessRage:
				sm.Resource[string(f.Name)] = 0
			case core.SpecAbilitySecondWind:
				sm.Resource[string(f.Name)] = 1
			case core.SpecAbilityIndomitable:
				// Indomitable might have multiple uses at high levels, but for now:
				sm.Resource[string(f.Name)] = 1
			case core.SpecAbilityActionSurge:
				sm.Resource[string(f.Name)] = f.Data.Value
			case core.SpecAbilityStrokeOfLuck:
				sm.Resource[string(f.Name)] = 1
			}
		}
	}
}

func (im *IntermissionManager) performPostCombatHealing(party []*actor.Actor, threshold float64) (map[int]int, map[int]int, map[int]map[int]int) {
	// Find party members who need healing
	needsHealing := make([]*actor.Actor, 0)
	for _, member := range party {
		if member.StateManager.CurrentHP < member.StateManager.MaxHP {
			needsHealing = append(needsHealing, member)
		}
	}

	if len(needsHealing) == 0 {
		return nil, nil, nil
	}

	totalHealing := make(map[int]int)
	totalSpellsUsed := make(map[int]int)
	totalSlotsUsed := make(map[int]map[int]int)

	for _, actor := range party {
		if actor.StateManager.CurrentHP <= 0 || !actor.IsHealer() {
			continue // Unconscious or not a healer can't heal
		}

		// Use healing features first (like Lay on Hands)
		featureHealing := im.useHealingFeatures(actor, needsHealing)

		// Use healing spells
		spellHealing, spellsUsed, slotsUsed := im.useHealingSpells(actor, needsHealing, threshold)

		for id, healAmount := range featureHealing {
			totalHealing[id] += healAmount
		}

		for id, healAmount := range spellHealing {
			totalHealing[id] += healAmount
		}

		if spellsUsed > 0 {
			totalSpellsUsed[actor.InstanceID] += spellsUsed
		}
		if len(slotsUsed) > 0 {
			if totalSlotsUsed[actor.InstanceID] == nil {
				totalSlotsUsed[actor.InstanceID] = make(map[int]int)
			}
			for lvl, count := range slotsUsed {
				totalSlotsUsed[actor.InstanceID][lvl] += count
			}
		}
	}

	return totalHealing, totalSpellsUsed, totalSlotsUsed
}

func (im *IntermissionManager) useHealingFeatures(healer *actor.Actor, party []*actor.Actor) map[int]int {
	totalHealing := make(map[int]int)
	sm := &healer.StateManager
	for _, f := range healer.Features {
		if f.Name == core.SpecAbilityLayOnHands {
			pool := sm.Resource[string(core.SpecAbilityLayOnHands)]
			if pool <= 0 {
				continue
			}

			for _, target := range party {
				needed := target.StateManager.MaxHP - target.StateManager.CurrentHP
				if needed <= 0 {
					continue
				}

				heal := pool
				if heal > needed {
					heal = needed
				}

				healRes := target.StateManager.ModifyHP(heal, false, target.ActorType == core.ActorTypeCharacter)
				pool -= heal
				sm.Resource[string(core.SpecAbilityLayOnHands)] = pool
				totalHealing[target.InstanceID] += healRes.ModificationValue

				if pool <= 0 {
					break
				}
			}
		}
	}

	return totalHealing
}

func (im *IntermissionManager) useHealingSpells(healer *actor.Actor, party []*actor.Actor, threshold float64) (map[int]int, int, map[int]int) {
	sm := &healer.StateManager
	if !healer.SpellManager.HasHealingSpells() {
		return nil, 0, nil
	}

	totalHealing := make(map[int]int)
	spellsUsed := 0
	slotsUsed := make(map[int]int)

	// Spend lowest slots first for out-of-combat healing
	levels := slices.Collect(maps.Keys(sm.CurrentSlots))
	slices.Sort(levels)

	for _, lvl := range levels {
		if lvl == 0 {
			continue // Cantrips handled separately or usually don't heal in 5e (except specific features)
		}

		for sm.CurrentSlots[lvl] > 0 {
			// Find most wounded target
			var target *actor.Actor
			minPct := 1.0

			for _, member := range party {
				pct := float64(member.StateManager.CurrentHP) / float64(member.StateManager.MaxHP)
				if pct < minPct && pct < threshold {
					minPct = pct
					target = member
				}
			}

			if target == nil {
				return totalHealing, spellsUsed, slotsUsed // No one needs healing below threshold
			}

			// Find a healing spell at this level
			spellsAtLvl := healer.SpellManager.GetHealingSpellsByLevel(lvl)
			if len(spellsAtLvl) == 0 {
				// Try to find a lower level spell to upcast
				for lower := lvl - 1; lower >= 1; lower-- {
					spellsAtLvl = healer.SpellManager.GetHealingSpellsByLevel(lower)
					if len(spellsAtLvl) > 0 {
						break
					}
				}
			}

			if len(spellsAtLvl) == 0 {
				break // No healing spells available to cast at this level or below
			}

			// Just use the first one for simplicity in simulation
			spell := spellsAtLvl[0]
			healAmount := spell.GetHighestAverageAmount() // Simple average for intermission
			// Adjust for cast level if upcasting
			if lvl > spell.Level {
				// Basic upcast logic: add one die avg per level
				// This is a simplification.
				healAmount += (lvl - spell.Level) * 5 // Rough estimate
			}

			healRes := target.StateManager.ModifyHP(healAmount, false, target.ActorType == core.ActorTypeCharacter)
			sm.CurrentSlots[lvl]--
			spellsUsed++
			slotsUsed[lvl]++
			totalHealing[target.InstanceID] += healRes.ModificationValue
		}
	}

	return totalHealing, spellsUsed, slotsUsed
}
