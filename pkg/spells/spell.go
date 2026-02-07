package spells

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	"math"
)

type Spell struct {
	ID              core.ID                    `json:"id"`
	Name            string                     `json:"name"`
	Description     string                     `json:"description"`
	IsConcentration bool                       `json:"is_concentration"`
	CastingTime     core.CastingTime           `json:"casting_time"`
	IsRitual        bool                       `json:"is_ritual"`
	Level           int                        `json:"level"`
	SpellType       core.SpellType             `json:"spell_type"`
	IsTouch         bool                       `json:"is_touch"`
	IsAOE           bool                       `json:"is_aoe"`
	HasDC           bool                       `json:"has_dc"`
	IsAutoHit       bool                       `json:"is_auto_hit"`
	ApiURL          string                     `json:"api_url"`
	LevelType       string                     `json:"level_type"`
	SpellDC         SpellDC                    `json:"spell_dc"`
	Formulas        map[int][]core.CastFormula `json:"formulas"`
	IsCustom        bool                       `json:"is_custom"`
}

func (s *Spell) GetID() core.ID                          { return s.ID }
func (s *Spell) GetName() string                         { return s.Name }
func (s *Spell) GetDescription() string                  { return s.Description }
func (s *Spell) GetIsConcentration() bool                { return s.IsConcentration }
func (s *Spell) GetCastingTime() core.CastingTime        { return s.CastingTime }
func (s *Spell) GetIsRitual() bool                       { return s.IsRitual }
func (s *Spell) GetLevel() int                           { return s.Level }
func (s *Spell) GetSpellType() core.SpellType            { return s.SpellType }
func (s *Spell) GetIsTouch() bool                        { return s.IsTouch }
func (s *Spell) GetIsAOE() bool                          { return s.IsAOE }
func (s *Spell) GetHasDC() bool                          { return s.HasDC }
func (s *Spell) GetIsAutoHit() bool                      { return s.IsAutoHit }
func (s *Spell) GetApiURL() string                       { return s.ApiURL }
func (s *Spell) GetLevelType() string                    { return s.LevelType }
func (s *Spell) GetSpellDC() SpellDC                     { return s.SpellDC }
func (s *Spell) GetFormulas() map[int][]core.CastFormula { return s.Formulas }

type SpellQueryParams struct {
	Name []string
	ID   []int
}

// GetHighestAverageAmount returns the total average damage/healing for the highest available formula level.
// This is used for AI decision making to estimate the maximum potential of a spell.
func (s *Spell) GetHighestAverageAmount() int {
	if len(s.Formulas) == 0 {
		return 0
	}

	maxLvl := -1
	for lvl := range s.Formulas {
		if lvl > maxLvl {
			maxLvl = lvl
		}
	}

	if maxLvl == -1 {
		return 0
	}

	formulas := s.Formulas[maxLvl]
	totalAvg := 0
	for _, f := range formulas {
		totalAvg += f.AverageValue
	}
	return totalAvg
}

// GetClosestFormulaToLevel retrieves the most suitable cast formulas for the given spell at the specified cast level.
func (s *Spell) GetClosestFormulaToLevel(castLevel int) ([]core.CastFormula, error) {
	if s.Level == 0 {
		return nil, fmt.Errorf("spell is not a leveled spell")
	}
	if castLevel < s.Level {
		return nil, fmt.Errorf("cast level %d is below minimum spell level %d", castLevel, s.Level)
	}
	if len(s.Formulas) == 0 {
		return nil, fmt.Errorf("spell has no formulas")
	}

	var bestFormulas []core.CastFormula
	var bestLevel int
	found := false

	for level, formulas := range s.Formulas {
		if level <= castLevel && (!found || level > bestLevel) {
			bestFormulas = formulas
			bestLevel = level
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("no formula available for spell %s at level %d", s.Name, castLevel)
	}

	return bestFormulas, nil
}

// GetAverageDamageAtLevel calculates and returns the average damage for a spell cast at the given level with modifiers.
// It uses the most suitable formulas based on the cast level and considers options like spell modifiers and additional amounts.
func (s *Spell) GetAverageDamageAtLevel(castLevel int, spellModDmg int) (int, []core.CastFormula, error) {
	if s.Level == 0 {
		return 0, nil, fmt.Errorf("spell is not a leveled spell")
	}
	formulas, err := s.GetClosestFormulaToLevel(castLevel)
	if err != nil {
		return 0, nil, err
	}

	totalDmg := 0
	for _, formula := range formulas {
		dAvg := formula.Die.Avg()
		dmg := int(math.Floor(float64(formula.NumberOfDice)*dAvg)) + formula.AmountToAdd
		if formula.UseSpellmod {
			dmg += spellModDmg
		}
		totalDmg += dmg
	}

	return totalDmg, formulas, nil
}

func (s *Spell) GetAverageDamageCantrip(casterLevel int, spellModDmg int) (int, []core.CastFormula, error) {
	if s.Level != 0 {
		return 0, nil, fmt.Errorf("spell is not a cantrip")
	}

	formulas, err := s.GetFormulaForCantrip(casterLevel)
	if err != nil {
		return 0, nil, err
	}

	totalDmg := 0
	for _, formula := range formulas {
		dAvg := formula.Die.Avg()
		dmg := int(math.Floor(float64(formula.NumberOfDice)*dAvg)) + formula.AmountToAdd
		if formula.UseSpellmod {
			dmg += spellModDmg
		}
		totalDmg += dmg
	}

	return totalDmg, formulas, nil
}

func (s *Spell) GetFormulaForCantrip(casterLevel int) ([]core.CastFormula, error) {
	if s.Level != 0 {
		return nil, fmt.Errorf("spell is not a cantrip")
	}

	var bestFormulas []core.CastFormula
	var bestLevel int
	found := false

	for level, formulas := range s.Formulas {
		if level <= casterLevel && (!found || level > bestLevel) {
			bestLevel = level
			bestFormulas = formulas
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("no formula available for cantrip %s at level %d", s.Name, casterLevel)
	}

	return bestFormulas, nil
}
