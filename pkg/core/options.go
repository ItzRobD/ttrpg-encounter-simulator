package core

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
	TargetPriority            TargetPrioritization
	HealPriority              TargetPrioritization
	ActionPreference          ActionPreference
	AOEHitsAllEnemies         bool
	CharacterHealThresholdPct int
	MonsterHealThresholdPct   int
}
