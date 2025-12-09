package core

type SimulationOptions struct {
	Seed                      Seed
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
	LimitedLegendaryActions   bool
	// AllowLairActions gates executing a lair action at initiative 20 each round.
	// Currently used only as a stub hook (logs/placeholder) until data-driven lair actions are modeled.
	AllowLairActions bool
}
