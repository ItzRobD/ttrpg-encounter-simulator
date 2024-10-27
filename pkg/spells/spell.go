package spells

import (
	"math"
)

const (
	SpellDCStrength     = "str"
	SpellDCDexterity    = "dex"
	SpellDCConstitution = "con"
	SpellDCIntelligence = "int"
	SpellDCWisdom       = "wis"
	SpellDCCharisma     = "cha"
)

type SpellType string

const (
	STDamage  SpellType = "damage"
	STHealing SpellType = "healing"
)

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

func (s Spell) GetAverageAmount() int {
	if s.Die <= 0 {
		return 0
	}
	dieAvg := float64(s.Die+1) / 2.0
	spellAvg := dieAvg * float64(s.NumberOfDice)
	return int(math.Floor(spellAvg))
}
