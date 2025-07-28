package core

type SimulationOptions struct {
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
	AOEHitsAllEnemies         bool
	CharacterHealThresholdPct int
	MonsterHealThresholdPct   int
}
