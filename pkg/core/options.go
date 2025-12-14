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
	// TODO: Model data-driven lair actions and wire execution; currently a stub hook used for logging/placeholder only.
	AllowLairActions bool
}
