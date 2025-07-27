package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	"math/rand/v2"
)

type CharacterAI struct {
	entity    core.Entity
	combatCtx CombatContext
	rng       *rand.Rand
}

func NewCharacterAI(entity core.Entity, combatCtx CombatContext, rng *rand.Rand) *CharacterAI {
	return &CharacterAI{
		entity:    entity,
		combatCtx: combatCtx,
		rng:       rng,
	}
}

func (cai *CharacterAI) selectTargetID(targetType TargetType) (int, error) {
	var validTargets map[int]core.Combatant
	switch targetType {
	case Damage:
		validTargets = cai.getEnemyTargets()
	case Healing:
		validTargets = cai.getAllyTargets()
	default:
		return -1, fmt.Errorf("unknown target type")
	}

	target, err := SelectTargetFromMap(validTargets, cai.entity.GetTargetPriority(), cai.rng)
	if err != nil {
		return -1, err
	}
	return target, nil
}

func (cai *CharacterAI) getEnemyTargets() map[int]core.Combatant {
	enemies := make(map[int]core.Combatant)
	self := cai.entity

	for id, combatant := range cai.combatCtx.AllCombatants {
		e := combatant.GetEntity()
		if !e.IsUnconscious() && (self.IsCharacter() != e.IsCharacter()) {
			enemies[id] = combatant
		}
	}

	return enemies
}

func (cai *CharacterAI) getAllyTargets() map[int]core.Combatant {
	allies := make(map[int]core.Combatant)
	self := cai.entity

	for id, combatant := range cai.combatCtx.AllCombatants {
		e := combatant.GetEntity()
		if !e.IsUnconscious() && (self.IsCharacter() == e.IsCharacter()) {
			allies[id] = combatant
		}
	}

	return allies
}
