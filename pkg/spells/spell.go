package spells

import (
	"fmt"
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
	Level           int // Minimum spell level
	SpellType       string
	IsAOE           bool
	HasDC           bool
	ApiURL          string
	LevelType       string // character || slot
	SpellDC
	Formulas []CastFormula
}

type SpellDC struct {
	Ability   string
	OnSuccess string
}

type CastFormula struct {
	CastLevel    int
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

type SpellResult struct {
	Success bool
	Amount  int
	Rolls   []int
}

func (s Spell) GetHighestAverageAmount() int {
	highestDie := s.Formulas[len(s.Formulas)-1].Die
	highestNumDice := s.Formulas[len(s.Formulas)-1].NumberOfDice
	if highestDie <= 0 {
		return 0
	}
	dieAvg := float64(highestDie+1) / 2.0
	spellAvg := dieAvg * float64(highestNumDice)
	return int(math.Floor(spellAvg))
}

func (s *Spell) GetForumlaAtLevel(castLevel int) (*CastFormula, error) {
	if castLevel < s.Level {
		castLevel = s.Level
	}

	var bestFormula *CastFormula
	for i := range s.Formulas {
		formula := &s.Formulas[i]
		if formula.CastLevel <= castLevel {
			bestFormula = formula
		} else {
			break
		}
	}

	if bestFormula == nil {
		return nil, fmt.Errorf("no formula available for spell %s at level %d", s.Name, castLevel)
	}

	return bestFormula, nil
}
