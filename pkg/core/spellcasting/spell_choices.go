package spellcasting

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"math/rand/v2"
)

func (s *SpellcastingManager) ChooseSpellByPriority(t spells.SpellType, priority shared.SpellPriority) (*SpellChoice, error) {
	switch priority {
	case shared.SPNoPreference: // Random known spell
	case shared.SPCantrip: // Prioritizes highest value cantrip
		choice, err := s.GetHighestAverageCantrip(t)
		if err != nil {
			return nil, err
		}
		return choice, nil
	case shared.SPLowestLevel:
		pool, err := s.getLowestLeveledSpells(t)
		if err != nil {
			return nil, err
		}
		choice, err := s.getHighestAverageAvailableOfPool(pool)
		if err != nil {
			return nil, err
		}
		return choice, nil
	case shared.SPHighestLevel:
		pool, err := s.getHighestLevelSpells(t)
		if err != nil {
			return nil, err
		}
		choice, err := s.getHighestAverageAvailableOfPool(pool)
		if err != nil {
			return nil, err
		}
		return choice, nil
	case shared.SPAreaOfEffect:
		return s.getHighestAverageAvailableAOESpell(t)
	case shared.SPHighestDamage:
		return s.getHighestAverageAvailableSpell(t)
	default:
		return nil, NewSpellcastingError("", "invalid spell priority", ERROR_GENERIC_SPELL)
	}
	return nil, nil
}

func (s *SpellcastingManager) getSpellPoolOfType(t spells.SpellType) (map[int][]*spells.Spell, error) {
	switch t {
	case spells.STHealing:
		return s.GetHealingSpells(), nil
	case spells.STDamage:
		return s.GetDamageSpells(), nil
	default:
		return nil, NewSpellcastingError("", "invalid spell type", ERROR_GENERIC_SPELL)
	}
}

func (s *SpellcastingManager) getHighestAverageAvailableAOESpell(t spells.SpellType) (*SpellChoice, error) {
	pool, err := s.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	var highestSpell *spells.Spell
	var highestFormula *spells.CastFormula
	var highestValue int

	for _, spellsAtLevel := range pool {
		for _, spell := range spellsAtLevel {
			if !spell.IsAOE {
				continue
			}

			castLevel, formula, value := s.getBestCastOptionForSpell(spell)
			if castLevel == -1 {
				continue
			}

			if value > highestValue {
				highestValue = value
				highestSpell = spell
				highestFormula = formula
			}
		}
	}

	if highestSpell == nil {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	return &SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
}

// getHighestAverageAvailableOfPool determines the spell from the given pool with the highest average output and returns it.
// This method evaluates both leveled spells and cantrips, considering the spell's formulas and caster level.
// Returns a SpellChoice containing the selected spell and its formula or an error if no valid spell is found.
func (s *SpellcastingManager) getHighestAverageAvailableOfPool(pool []*spells.Spell) (*SpellChoice, error) {
	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found", ERROR_SPELL_NOT_KNOWN)
	}

	var highestSpell *spells.Spell
	var highestFormula *spells.CastFormula
	var highestValue int
	casterLevel := s.GetCasterLevel()
	for _, spell := range pool {
		if spell.Level == 0 {
			formula, err := spell.GetFormulaForCantrip(casterLevel)
			if err != nil {
				return nil, err
			}
			value := formula.AverageValue
			if value > highestValue {
				highestValue = value
				highestSpell = spell
				highestFormula = formula
			}
		} else {
			castLevel, formula, value := s.getBestCastOptionForSpell(spell)
			if castLevel == -1 {
				continue
			}

			if value > highestValue {
				highestValue = value
				highestSpell = spell
				highestFormula = formula
			}
		}
	}

	if highestSpell == nil {
		return nil, NewSpellcastingError("", "no spells found", ERROR_SPELL_NOT_KNOWN)
	}

	return &SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
}

func (s *SpellcastingManager) GetBestFormulaForSpell(spell spells.Spell, p shared.SpellPriority) (*spells.CastFormula, error) {
	var formula *spells.CastFormula
	var err error
	var value int

	if spell.Level != 0 && !s.HasAnySpellSlots() {
		return nil, NewSpellSlotError(0, "actor has no available spell slots", ERROR_SLOT_NOT_AVAILABLE)
	}

	if p != shared.SPLowestLevel && p != shared.SPHighestLevel {
		return nil, NewSpellcastingError(spell.Name, "cannot get best formula for spell - invalid priority", ERROR_GENERIC_SPELL)
	}

	switch p {
	case shared.SPLowestLevel:
		if spell.Level == 0 {
			value, formula, err = spell.GetAverageDamageCantrip(s.parent.GetCasterLevel(), s.spellcastModifierValue)
			if err != nil {
				return nil, err
			}
		} else {
			var canCast bool
			castLevel := spell.Level
			for !canCast {
				if s.HasSpellSlotsAtLevel(castLevel) {
					canCast = true
					value, formula, err = spell.GetAverageDamageAtLevel(castLevel, s.spellcastModifierValue)
					if err != nil {
						return nil, err
					}
				} else if s.canUpcast && castLevel < 9 {
					castLevel++
				} else {
					break
				}
			}
		}
		if formula == nil || value == 0 {
			return nil, NewSpellcastingError(spell.Name, "unable to cast spell", ERROR_GENERIC_SPELL)
		}
		return formula, nil
	case shared.SPHighestLevel:
		if spell.Level == 0 {
			value, formula, err = spell.GetAverageDamageCantrip(s.parent.GetCasterLevel(), s.spellcastModifierValue)
			if err != nil {
				return nil, err
			}
		} else {
			var canCast bool
			castLevel := spell.Level
			for !canCast {
				if s.HasSpellSlotsAtLevel(castLevel) {
					canCast = true
					value, formula, err = spell.GetAverageDamageAtLevel(castLevel, s.spellcastModifierValue)
					if err != nil {
						return nil, err
					}
				} else {
					castLevel--
					if castLevel < 1 {
						break
					}
				}
			}

			if formula == nil || value == 0 {
				return nil, NewSpellcastingError(spell.Name, "unable to cast spell", ERROR_GENERIC_SPELL)
			}
			return formula, nil
		}
	default:
		return nil, NewSpellcastingError(spell.Name, "cannot get best formula for spell - invalid priority", ERROR_GENERIC_SPELL)
	}
	return nil, nil
}

// GetRandomCantrip retrieves a random cantrip of the specified spell type (healing or damage).
// Returns an error if no cantrips are found for the provided type or if the type is invalid.
func (s *SpellcastingManager) GetRandomCantrip(t spells.SpellType) (*spells.Spell, error) {
	var selected *spells.Spell
	switch t {
	case spells.STHealing:
		pool := s.GetHealingCantrips()
		if len(pool) == 0 {
			return nil, NewSpellcastingError("", "no healing cantrips found", ERROR_SPELL_NOT_KNOWN)
		}
		rand.NewPCG(rand.Uint64(), rand.Uint64())
		i := rand.IntN(len(pool))
		selected = pool[i]
		return selected, nil
	case spells.STDamage:
		pool := s.GetDamageCantrips()
		if len(pool) == 0 {
			return nil, NewSpellcastingError("", "no damage cantrips found", ERROR_SPELL_NOT_KNOWN)
		}
		rand.NewPCG(rand.Uint64(), rand.Uint64())
		i := rand.IntN(len(pool))
		selected = pool[i]
		return selected, nil
	default:
		return nil, NewSpellcastingError("", "invalid spell type", ERROR_GENERIC_SPELL)
	}
}

// GetRandomLeveledSpell selects a random leveled spell of the given type (healing or damage).
// Returns an error if no spells are found for the specified type or if the type is invalid.
func (s *SpellcastingManager) GetRandomLeveledSpell(t spells.SpellType) (*spells.Spell, error) {
	var selected *spells.Spell
	switch t {
	case spells.STHealing:
		pool := s.GetHealingSpells()
		if len(pool) == 0 {
			return nil, NewSpellcastingError("", "no healing spells found", ERROR_SPELL_NOT_KNOWN)
		}

		var allSpells []*spells.Spell
		for _, spellsAtLevel := range pool {
			allSpells = append(allSpells, spellsAtLevel...)
		}
		if len(allSpells) == 0 {
			return nil, NewSpellcastingError("", "no healing spells found", ERROR_SPELL_NOT_KNOWN)
		}

		rand.NewPCG(rand.Uint64(), rand.Uint64())
		i := rand.IntN(len(allSpells))
		selected = allSpells[i]
		return selected, nil
	case spells.STDamage:
		pool := s.GetDamageSpells()
		if len(pool) == 0 {
			return nil, NewSpellcastingError("", "no damage spells found", ERROR_SPELL_NOT_KNOWN)
		}

		var allSpells []*spells.Spell
		for _, spellsAtLevel := range pool {
			allSpells = append(allSpells, spellsAtLevel...)
		}
		if len(allSpells) == 0 {
			return nil, NewSpellcastingError("", "no healing spells found", ERROR_SPELL_NOT_KNOWN)
		}

		rand.NewPCG(rand.Uint64(), rand.Uint64())
		i := rand.IntN(len(allSpells))
		selected = allSpells[i]
		return selected, nil
	default:
		return nil, NewSpellcastingError("", "invalid spell type", ERROR_GENERIC_SPELL)
	}
}

func (s *SpellcastingManager) getHighestAverageAvailableSpell(t spells.SpellType) (*SpellChoice, error) {
	pool, err := s.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	var highestSpell *spells.Spell
	var highestFormula *spells.CastFormula
	var highestValue int

	for _, spellsAtLevel := range pool {
		for _, spell := range spellsAtLevel {
			castLevel, formula, value := s.getBestCastOptionForSpell(spell)
			if castLevel == -1 {
				continue
			}

			if value > highestValue {
				highestValue = value
				highestSpell = spell
				highestFormula = formula
			}
		}
	}

	if highestSpell == nil {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	return &SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
}

// getBestCastOptionForSpell determines the optimal way to cast a spell based on level, available slots, and upcast options.
// Returns the cast level, formula, and average value of the spell.
func (s *SpellcastingManager) getBestCastOptionForSpell(spell *spells.Spell) (int, *spells.CastFormula, int) {
	if spell.Level == 0 {
		formula, err := spell.GetFormulaForCantrip(s.casterLevel)
		if err != nil {
			return -1, nil, 0
		}
		return s.casterLevel, formula, formula.AverageValue
	}

	if !s.HasAnySpellSlots() {
		return -1, nil, 0
	}

	if s.canUpcast {
		for level := 9; level >= spell.Level; level-- {
			formula, err := spell.GetClosestFormulaToLevel(level)
			if err != nil {
				continue
			}

			if !s.HasSpellSlotsAtLevel(formula.CastLevel) {
				continue
			}

			return formula.CastLevel, formula, formula.AverageValue
		}
	} else {
		if s.HasSpellSlotsAtLevel(spell.Level) {
			formula, err := spell.GetClosestFormulaToLevel(spell.Level)
			if err != nil {
				return -1, nil, 0
			}

			return formula.CastLevel, formula, formula.AverageValue
		}
	}

	return -1, nil, 0
}

func (s *SpellcastingManager) GetHighestAverageCantrip(t spells.SpellType) (*SpellChoice, error) {
	var pool []*spells.Spell
	switch t {
	case spells.STHealing:
		pool = s.GetHealingCantrips()
	case spells.STDamage:
		pool = s.GetDamageCantrips()
	default:
		return nil, NewSpellcastingError("", "invalid spell type", ERROR_GENERIC_SPELL)
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	var highestSpell *spells.Spell
	var highestFormula *spells.CastFormula
	var highestValue int
	casterLevel := s.GetCasterLevel()

	for _, spell := range pool {
		formula, err := spell.GetFormulaForCantrip(casterLevel)
		if err != nil {
			return nil, err
		}

		value := formula.AverageValue
		if value > highestValue {
			highestValue = value
			highestSpell = spell
			highestFormula = formula
		}
	}

	if highestSpell == nil {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	return &SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
}

func (s *SpellcastingManager) getHighestLevelSpells(t spells.SpellType) ([]*spells.Spell, error) {
	pool, err := s.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	for level := 9; level > 0; level-- {
		if len(pool[level]) > 0 {
			return pool[level], nil
		}
	}

	return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
}

func (s *SpellcastingManager) getLowestLeveledSpells(t spells.SpellType) ([]*spells.Spell, error) {
	pool, err := s.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	for level := 1; level <= 9; level++ {
		if len(pool[level]) > 0 {
			return pool[level], nil
		}
	}

	return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
}

func (s *SpellcastingManager) isAbleToCast(choice SpellChoice) bool {
	if choice.Spell.Level == 0 {
		return true
	}

	return s.HasSpellSlotsAtLevel(choice.Formula.CastLevel)
}
