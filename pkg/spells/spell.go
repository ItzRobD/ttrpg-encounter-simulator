package spells

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	"math"
)

type Spell struct {
	ID              int
	Name            string
	Description     string
	IsConcentration bool
	CastingTime     core.CastingTime
	IsRitual        bool
	Level           int // Minimum spell level
	SpellType       core.SpellType
	IsAOE           bool
	HasDC           bool
	ApiURL          string
	LevelType       string // character || slot
	SpellDC         SpellDC
	Formulas        map[int]core.CastFormula
}

func (s *Spell) GetID() int                            { return s.ID }
func (s *Spell) GetName() string                       { return s.Name }
func (s *Spell) GetDescription() string                { return s.Description }
func (s *Spell) GetIsConcentration() bool              { return s.IsConcentration }
func (s *Spell) GetCastingTime() core.CastingTime      { return s.CastingTime }
func (s *Spell) GetIsRitual() bool                     { return s.IsRitual }
func (s *Spell) GetLevel() int                         { return s.Level }
func (s *Spell) GetSpellType() core.SpellType          { return s.SpellType }
func (s *Spell) GetIsAOE() bool                        { return s.IsAOE }
func (s *Spell) GetHasDC() bool                        { return s.HasDC }
func (s *Spell) GetApiURL() string                     { return s.ApiURL }
func (s *Spell) GetLevelType() string                  { return s.LevelType }
func (s *Spell) GetSpellDC() core.SpellDC              { return s.SpellDC }
func (s *Spell) GetFormulas() map[int]core.CastFormula { return s.Formulas }

type SpellDC struct {
	Ability   core.Ability
	OnSuccess core.DCOnSuccess
}

func (s SpellDC) GetAbility() core.Ability       { return s.Ability }
func (s SpellDC) GetOnSuccess() core.DCOnSuccess { return s.OnSuccess }

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
func (s *Spell) GetClosestFormulaToLevel(castLevel int) (*core.CastFormula, error) {
	if s.Level == 0 {
		return nil, fmt.Errorf("spell is not a leveled spell")
	}
	if castLevel < s.Level {
		return nil, fmt.Errorf("cast level %d is below minimum spell level %d", castLevel, s.Level)
	}

	var bestFormula *core.CastFormula
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
func (s *Spell) GetAverageDamageAtLevel(castLevel int, spellModDmg int) (int, *core.CastFormula, error) {
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

func (s *Spell) GetAverageDamageCantrip(casterLevel int, spellModDmg int) (int, *core.CastFormula, error) {
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

func (s *Spell) GetFormulaForCantrip(casterLevel int) (*core.CastFormula, error) {
	if s.Level != 0 {
		return nil, fmt.Errorf("spell is not a cantrip")
	}

	var bestFormula *core.CastFormula
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
