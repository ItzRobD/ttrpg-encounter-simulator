package spells

type Spell struct {
	ID              int
	Name            string
	Description     string
	IsConcentration bool
	CastingTime     string
	IsRitual        bool
	Level           int
	SpellType       string
	IsAOE           bool
	HasDC           bool
	ApiURL          string
	SpellDC
	CastFormula
}

type SpellDC struct {
	Ability   string
	OnSuccess string
}

type CastFormula struct {
	CastLevel    int
	LevelType    string
	NumberOfDice int
	Die          int
	AmountToAdd  int
	UseSpellmod  bool
	DamageType   string
}

type SpellQueryParams struct {
	Name  string
	ID    int
	Level int
}
