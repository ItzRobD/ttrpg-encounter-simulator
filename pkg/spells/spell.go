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
	Level           int    // Minimum spell level
	SpellType       string // TODO: Consider changing this to the spelltype enum
	IsAOE           bool
	HasDC           bool
	ApiURL          string
	LevelType       string // character || slot
	SpellDC
	Formulas map[int]CastFormula
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
	UseSpellmod  bool // UseSpellmod specifies whether the spell modifier should be added to the calculated damage.
	DamageType   string
	AverageValue int
}

type SpellQueryParams struct {
	Name  string
	ID    int
	Level int
}

func (s *Spell) GetHighestAverageAmount() int {
	highestDie := s.Formulas[len(s.Formulas)-1].Die
	highestNumDice := s.Formulas[len(s.Formulas)-1].NumberOfDice
	if highestDie <= 0 {
		return 0
	}
	dieAvg := float64(highestDie+1) / 2.0
	spellAvg := dieAvg * float64(highestNumDice)
	return int(math.Floor(spellAvg))
}

// GetClosestFormulaToLevel retrieves the most suitable cast formula for the given spell at the specified cast level.
func (s *Spell) GetClosestFormulaToLevel(castLevel int) (*CastFormula, error) {
	if s.Level == 0 {
		return nil, fmt.Errorf("spell is not a leveled spell")
	}
	if castLevel < s.Level {
		return nil, fmt.Errorf("cast level %d is below minimum spell level %d", castLevel, s.Level)
	}

	var bestFormula *CastFormula
	var bestLevel int
	found := false

	for level, formula := range s.Formulas {
		if level <= castLevel && (!found || level > bestLevel) {
			bestFormula = &formula
			bestLevel = level
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("no formula available for spell %s at level %d", s.Name, castLevel)
	}

	return bestFormula, nil
}

// GetAverageDamageAtLevel calculates and returns the average damage for a spell cast at the given level with modifiers.
// It uses the most suitable formula based on the cast level and considers options like spell modifiers and additional amounts.
func (s *Spell) GetAverageDamageAtLevel(castLevel int, spellModDmg int) (int, *CastFormula, error) {
	if s.Level == 0 {
		return 0, nil, fmt.Errorf("spell is not a leveled spell")
	}
	formula, err := s.GetClosestFormulaToLevel(castLevel)
	if err != nil {
		return 0, nil, err
	}

	dAvg := float64(formula.Die+1) / 2.0
	dmg := int(math.Floor(float64(formula.NumberOfDice)*dAvg)) + formula.AmountToAdd
	if formula.UseSpellmod {
		dmg += spellModDmg
	}

	return dmg, formula, nil
}

func (s *Spell) GetAverageDamageCantrip(casterLevel int, spellModDmg int) (int, *CastFormula, error) {
	if s.Level != 0 {
		return 0, nil, fmt.Errorf("spell is not a cantrip")
	}

	formula, err := s.GetFormulaForCantrip(casterLevel)
	if err != nil {
		return 0, nil, err
	}

	dAvg := float64(formula.Die+1) / 2.0
	if err != nil {
		return 0, nil, err
	}
	dmg := int(math.Floor(float64(formula.NumberOfDice)*dAvg)) + formula.AmountToAdd
	if formula.UseSpellmod {
		dmg += spellModDmg
	}

	return dmg, formula, nil
}

func (s *Spell) GetFormulaForCantrip(casterLevel int) (*CastFormula, error) {
	if s.Level != 0 {
		return nil, fmt.Errorf("spell is not a cantrip")
	}

	var bestFormula *CastFormula
	var bestLevel int
	found := false

	for level, formula := range s.Formulas {
		if level <= casterLevel && (!found || level > bestLevel) {
			bestLevel = level
			formulaCopy := formula
			bestFormula = &formulaCopy
			found = true
		}
	}

	return bestFormula, nil

}
