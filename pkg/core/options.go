package core

type SimulationOptions struct {
	Seed                          Seed
	UseHPAverageMonster           bool
	UseHPAverageCharacter         bool
	CanMonstersCrit               bool
	CanCharactersCrit             bool
	HasIncreasedCrits             bool
	UseImprovedCriticals          bool
	CharactersAlwaysUpcast        bool
	MonstersAlwaysUpcast          bool
	AllowCharacterHeals           bool
	AllowMonsterHeals             bool
	AOEHitsAllEnemies             bool
	CharacterHealThresholdPct     int
	MonsterHealThresholdPct       int
	LimitedLegendaryActions       bool
	AllowLairActions              bool
	AllowDragonbornBreathAttack   bool
	EnableClassFeatures           bool
	EnableRacialFeatures          bool
	BarbarianAlwaysRecklessAttack bool
	PaladinAlwaysSmite            bool
	PaladinUseHighestSmiteSlot    bool
	UseMassiveDamage              bool
	EnableSpecialAbilities        bool
	MonsterDeathEffectsHitAllies  bool
}
