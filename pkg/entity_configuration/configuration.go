package entity_configuration

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
)

type EntityConfiguration struct {
	CombatFeatures CombatFeatures
	UtilityWeights *core.UtilityWeights
}

type CombatFeatures struct {
	ReRollAbilities   roll_manager.RerollAbilities
	CriticalThreshold int
	InitiativeBonus   int
}
