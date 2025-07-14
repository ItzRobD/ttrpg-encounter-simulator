package core

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
)

type SimulationOptions struct {
	UseMonsterHPAverage       bool
	CanMonstersCrit           bool
	CanCharactersCrit         bool
	HasIncreasedCrits         bool
	UseImprovedCriticals      bool
	CharactersAlwaysUpcast    bool
	MonstersAlwaysUpcast      bool
	AllowCharacterHeals       bool
	AllowMonsterHeals         bool
	TargetPriority            shared.Prioritization
	HealPriority              shared.Prioritization
	ActionPreference          shared.ActionPreference
	AOEHitsAllEnemies         bool
	CharacterHealThresholdPct int
	MonsterHealThresholdPct   int
}
