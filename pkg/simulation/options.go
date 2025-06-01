package simulation

import "dnd5e-encounter-simulator-backend/pkg/shared"

type Options struct {
	UseMonsterHPAverage     bool
	CanMonstersCrit         bool
	CanPlayersCrit          bool
	HasIncreasedCrits       bool
	AllowPlayerHeals        bool
	AllowMonsterHeals       bool
	TargetPriority          shared.Prioritization
	HealPriority            shared.Prioritization
	ActionPreference        shared.ActionPreference
	AOEHitsAllEnemies       bool
	PlayerHealThresholdPct  int
	MonsterHealThresholdPct int
}
