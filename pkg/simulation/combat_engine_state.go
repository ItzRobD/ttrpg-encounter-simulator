package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
)

func (ce *CombatEngine) isLegendaryCreature(id int) bool {
	if ce.CombatContext.LegendaryCreatures == nil {
		return false
	}
	_, exists := ce.CombatContext.LegendaryCreatures[id]
	return exists
}

func (ce *CombatEngine) initializeEventContext() {
	if ce.EventContext == nil {
		ce.EventContext = core.NewEventContext()
	}
}

func (ce *CombatEngine) initializeCombatContext() {
	if ce.CombatContext == nil {
		ce.CombatContext = core.NewCombatContext(ce.SimOptions)
	}
	ce.CombatContext.CurrentRound = ce.CurrentRound
	ce.CombatContext.TurnOrder = ce.TurnOrder

	ce.CombatContext.CombatantInfo = make(map[int]*core.CombatantInfo)

	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		combatant := ce.Combatants[id]
		// Skip lair combatants (they have no entity)
		if combatant.IsLair {
			continue
		}

		ce.CombatContext.CombatantInfo[id] = combatant.Info

		// Track legendary creatures
		if combatant.Entity.IsMonster() && combatant.Entity.GetIsLegendary() {
			ce.CombatContext.LegendaryCreatures[id] = true
		}
	}
}

func (ce *CombatEngine) updateCombatContext(actorID int) {
	ce.CombatContext.CurrentRound = ce.CurrentRound
	ce.CombatContext.ActingEntityID = actorID
	ce.CombatContext.ConsciousCharacterCount = 0
	ce.CombatContext.ConsciousMonsterCount = 0

	ce.CombatContext.CharactersInNeedOfHealing, ce.CombatContext.MonstersInNeedOfHealing = ce.calculateEntitiesNeedingHealing()
	ce.CombatContext.DeadCombatants = ce.getDeadCombatantIDs()

	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		c := ce.Combatants[id]
		if !c.GetEntity().IsUnconscious() {
			if c.GetEntity().IsCharacter() {
				ce.CombatContext.ConsciousCharacterCount++
			} else {
				ce.CombatContext.ConsciousMonsterCount++
			}
		}
	}

	for _, id := range ce.getSortedCombatantIDs() {
		if info, exists := ce.CombatContext.CombatantInfo[id]; exists {
			info.UpdateState()
		}
	}
}

func (ce *CombatEngine) calculateEntitiesNeedingHealing() ([]int, []int) {
	var charNeedHealing, monNeedHealing []int

	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		combatant := ce.Combatants[id]
		if combatant.IsLair {
			continue
		}

		entity := combatant.GetEntity()

		// Calculate HP percentage
		var threshold int
		if entity.IsCharacter() {
			threshold = ce.CombatContext.Options.CharacterHealThresholdPct
		} else {
			threshold = ce.CombatContext.Options.MonsterHealThresholdPct
		}

		// Entity needs healing if below threshold and not dead
		if entity.GetHPStatus().GetHPPct() <= threshold && !entity.IsDead() {
			if entity.IsCharacter() {
				charNeedHealing = append(charNeedHealing, id)
			} else {
				monNeedHealing = append(monNeedHealing, id)
			}
		}
	}

	return charNeedHealing, monNeedHealing
}

func (ce *CombatEngine) getDeadCombatantIDs() []int {
	var deadIDs []int
	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		combatant := ce.Combatants[id]
		if combatant.IsLair {
			continue
		}
		if combatant.GetEntity().IsDead() {
			deadIDs = append(deadIDs, id)
		}
	}
	return deadIDs
}

func (ce *CombatEngine) checkVictoryCondition() core.VictoryStatus {
	var aliveCharacters, aliveMonsters bool

	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		// Skip lair combatants
		if ce.Combatants[id].IsLair {
			continue
		}

		entity := ce.Combatants[id].GetEntity()
		// Treat unconscious as not alive for victory purposes
		if entity.IsDead() || entity.IsUnconscious() {
			continue
		}
		if entity.IsCharacter() {
			aliveCharacters = true
		} else if entity.IsMonster() {
			aliveMonsters = true
		}

		if aliveCharacters && aliveMonsters {
			return core.VictoryStatusNone
		}
	}

	if !aliveCharacters {
		return core.VictoryStatusMonsters
	}
	if !aliveMonsters {
		return core.VictoryStatusCharacters
	}
	return core.VictoryStatusNone
}
