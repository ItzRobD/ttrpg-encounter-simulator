package entity_configuration

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/roll_manager"
)

type EntityConfiguration struct {
	CombatFeatures CombatFeatures       `json:"combat_features"`
	UtilityWeights *core.UtilityWeights `json:"utility_weights"`
}

type CombatFeatures struct {
	ReRollAbilities   roll_manager.RerollAbilities `json:"reroll_abilities"`
	CriticalThreshold int                          `json:"critical_threshold"`
	InitiativeBonus   int                          `json:"initiative_bonus"`
}
