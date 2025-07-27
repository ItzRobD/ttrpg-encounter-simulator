package simulation

import "dnd5e-encounter-simulator-backend/pkg/core"

type SimulationConfig struct {
	Seed                      int
	UseHPAverageMonster       bool
	UseHPAverageCharacter     bool
	CanMonstersCrit           bool
	CanCharactersCrit         bool
	HasIncreasedCrits         bool
	UseImprovedCriticals      bool
	CharactersAlwaysUpcast    bool
	MonstersAlwaysUpcast      bool
	AllowCharacterHeals       bool
	AllowMonsterHeals         bool
	TargetPriority            core.TargetPriority
	HealPriority              core.TargetPriority
	ActionPreference          core.ActionPreference
	AOEHitsAllEnemies         bool
	CharacterHealThresholdPct int
	MonsterHealThresholdPct   int
}
