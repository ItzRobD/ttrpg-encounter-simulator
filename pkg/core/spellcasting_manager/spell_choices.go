package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"errors"
	"math/rand/v2"
)

func (s *SpellcastingManager) GetMostEfficientHealingSpell(targetValue int) (*spells.SpellChoice, error) {
	if !s.HasHealingSpells() {
		return nil, NewSpellcastingError("", "no healing spells found", ERROR_SPELL_NOT_FOUND)
	}

	pool, err := s.getSpellPoolOfType(spells.STHealing)
	if err != nil {
		return nil, err
	}

	var options []HealingOption

	for _, spellsAtLevel := range pool {
		for _, spell := range spellsAtLevel {
			castLevel, formula, avg := s.getBestCastOptionForSpell(spell)

			if castLevel == -1 {
				continue
			}

			var efficiency float64
			if castLevel == 0 { // Cantrip
				efficiency = 1.0
			} else {
				efficiency = float64(formula.AverageValue) / float64(avg)
			}

			options = append(options, HealingOption{
				Spell:        spell,
				Formula:      formula,
				CastLevel:    castLevel,
				Efficiency:   efficiency,
				TargetDelta:  targetValue - avg,
				AverageValue: avg,
			})
		}
	}

	if len(options) == 0 {
		return nil, NewSpellcastingError("", "no available healing options", ERROR_SPELL_NOT_FOUND)
	}

	bestOption := s.selectBestHealingOption(options, targetValue)
	return &spells.SpellChoice{
		Spell:   bestOption.Spell,
		Formula: bestOption.Formula,
	}, nil
}

func (s *SpellcastingManager) selectBestHealingOption(options []HealingOption, targetValue int) HealingOption {
	// Sort by multiple criteria:
	// 1. Prefer spells that can exactly meet or slightly exceed the target
	// 2. Among those, prefer higher efficiency
	// 3. Prefer lower spell slot usage (less waste)

	// First, separate options that can meet the target from those that can't
	var exactMatches []HealingOption
	var overheals []HealingOption
	var underheals []HealingOption

	for _, option := range options {
		if option.TargetDelta == 0 {
			exactMatches = append(exactMatches, option)
		} else if option.TargetDelta < 0 { // Negative delta means overheal
			overheals = append(overheals, option)
		} else { // Positive delta means underheal
			underheals = append(underheals, option)
		}
	}

	// Prefer exact matches first
	if len(exactMatches) > 0 {
		return s.findMostEfficient(exactMatches)
	}

	// Then prefer minimal overheals
	if len(overheals) > 0 {
		return s.findMinimalOverheal(overheals)
	}

	// Finally, take the best underheal option
	return s.findBestUnderheal(underheals)

}

func (s *SpellcastingManager) findMostEfficient(options []HealingOption) HealingOption {
	best := options[0]
	for _, option := range options[1:] {
		if option.Efficiency > best.Efficiency {
			best = option
		}
	}
	return best
}

func (s *SpellcastingManager) findMinimalOverheal(options []HealingOption) HealingOption {
	best := options[0]
	for _, option := range options[1:] {
		// Prefer less overheal (less negative delta)
		if option.TargetDelta > best.TargetDelta {
			best = option
		} else if option.TargetDelta == best.TargetDelta {
			// If same overheal, prefer higher efficiency
			if option.Efficiency > best.Efficiency {
				best = option
			}
		}
	}
	return best
}

func (s *SpellcastingManager) findBestUnderheal(options []HealingOption) HealingOption {
	best := options[0]
	for _, option := range options[1:] {
		// Prefer less underheal (smaller positive delta)
		if option.TargetDelta < best.TargetDelta {
			best = option
		} else if option.TargetDelta == best.TargetDelta {
			// If same underheal, prefer higher efficiency
			if option.Efficiency > best.Efficiency {
				best = option
			}
		}
	}
	return best
}

func (s *SpellcastingManager) ChooseSpellByPriority(t spells.SpellType, priority shared.SpellPriority) (*spells.SpellChoice, error) {
	switch priority {
	case shared.SPNoPreference: // Random known spell
		choice, err := s.getRandomSpellChoice(t, false)
		if err != nil {
			var spellErr *SpellcastingError
			if errors.As(err, &spellErr) && spellErr.Type == ERROR_SPELL_NOT_FOUND {
				choice, err = s.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
		return choice, nil
	case shared.SPHighestLevel:
		choice, err := s.getHighestLevelSpellChoice(t)
		if err != nil {
			var spellSlotErr *SpellSlotError
			if errors.As(err, &spellSlotErr) && spellSlotErr.Type == ERROR_NO_SLOTS_AVAILABLE {
				choice, err = s.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
		return choice, nil
	case shared.SPLowestLevel:
		choice, err := s.getLowestLevelSpellChoice(t)
		if err != nil {
			var spellSlotErr *SpellSlotError
			if errors.As(err, &spellSlotErr) && spellSlotErr.Type == ERROR_NO_SLOTS_AVAILABLE {
				choice, err = s.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
		return choice, nil
	case shared.SPCantrip: // Prioritizes highest value cantrip
		choice, err := s.getHighestAverageCantrip(t)
		if err != nil {
			return nil, err
		}
		return choice, nil
	case shared.SPRandomCantrip:
		choice, err := s.getRandomCantripChoice(t)
		if err != nil {
			return nil, err
		}
		return choice, nil
	case shared.SPRandomLeveledSpell:
		choice, err := s.getRandomSpellChoice(t, true)
		if err != nil {
			var spellErr *SpellcastingError
			if errors.As(err, &spellErr) && spellErr.Type == ERROR_SPELL_NOT_FOUND {
				choice, err = s.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
		return choice, nil
	case shared.SPAreaOfEffect:
		choice, err := s.getHighestAverageAvailableAOESpellChoice(t)
		if err != nil {
			var spellErr *SpellcastingError
			if errors.As(err, &spellErr) && spellErr.Type == ERROR_SPELL_NOT_FOUND {
				choice, err = s.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
		return choice, nil
	case shared.SPHighestDamage:
		choice, err := s.getHighestAverageAvailableSpellChoice(t)
		if err != nil {
			var spellErr *SpellcastingError
			if errors.As(err, &spellErr) && spellErr.Type == ERROR_SPELL_NOT_FOUND {
				choice, err = s.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
	default:
		return nil, NewSpellcastingError("", "invalid spell priority", ERROR_GENERIC_SPELL)
	}
	return nil, nil
}

func (s *SpellcastingManager) getHighestAverageAvailableAOESpellChoice(t spells.SpellType) (*spells.SpellChoice, error) {
	pool, err := s.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no aoe spells found", ERROR_SPELL_NOT_FOUND)
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
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_FOUND)
	}

	return &spells.SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
}

// getHighestAverageAvailableOfPool determines the spell from the given pool with the highest average output and returns it.
// This method evaluates both leveled spells and cantrips, considering the spell's formulas and caster level.
// Returns a SpellChoice containing the selected spell and its formula or an error if no valid spell is found.
func (s *SpellcastingManager) getHighestAverageAvailableOfPool(pool []*spells.Spell) (*spells.SpellChoice, error) {
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

	return &spells.SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
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

// getRandomCantripChoice retrieves a random cantrip of the specified spell type (healing or damage).
// Returns an error if no cantrips are found for the provided type or if the type is invalid.
func (s *SpellcastingManager) getRandomCantripChoice(t spells.SpellType) (*spells.SpellChoice, error) {
	pool, err := s.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	var castableChoices []*spells.SpellChoice
	if cantrips, exists := pool[0]; exists {
		for _, spell := range cantrips {
			castLevel, formula, _ := s.getBestCastOptionForSpell(spell)
			if castLevel != -1 {
				castableChoices = append(castableChoices, &spells.SpellChoice{Spell: spell, Formula: formula})
			}
		}
	}
	if len(castableChoices) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_FOUND)
	}

	rand.NewPCG(rand.Uint64(), rand.Uint64())
	i := rand.IntN(len(castableChoices))
	return castableChoices[i], nil
}

// getRandomSpellChoice selects a random leveled spell of the given type (healing or damage).
// Returns an error if no spells are found for the specified type or if the type is invalid.
func (s *SpellcastingManager) getRandomSpellChoice(t spells.SpellType, excludeCantrips bool) (*spells.SpellChoice, error) {
	pool, err := s.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_FOUND)
	}

	var castableChoices []*spells.SpellChoice
	for level, spellsAtLevel := range pool {
		if level == 0 && excludeCantrips {
			continue
		}
		for _, spell := range spellsAtLevel {
			castLevel, formula, _ := s.getBestCastOptionForSpell(spell)
			if castLevel != -1 {
				castableChoices = append(castableChoices, &spells.SpellChoice{Spell: spell, Formula: formula})
			}
		}
	}

	if len(castableChoices) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_FOUND)
	}

	rand.NewPCG(rand.Uint64(), rand.Uint64())
	i := rand.IntN(len(castableChoices))
	return castableChoices[i], nil
}

func (s *SpellcastingManager) getHighestAverageAvailableSpellChoice(t spells.SpellType) (*spells.SpellChoice, error) {
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

	return &spells.SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
}

// getBestCastOptionForSpell determines the optimal way to cast a spell based on level, available slots, and upcast options.
// Returns the cast level, formula, and average value of the spell.
// Returns -1 if unable to cast the spell || error
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
		} else {
			return -1, nil, 0
		}
	}

	return -1, nil, 0
}

func (s *SpellcastingManager) getHighestAverageCantrip(t spells.SpellType) (*spells.SpellChoice, error) {
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

	return &spells.SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
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
		if len(pool[level]) > 0 && s.HasSpellSlotsAtLevel(level) {
			return pool[level], nil
		}
	}

	return nil, NewSpellSlotErrorOutOfSlots(0)
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
		if len(pool[level]) > 0 && s.HasSpellSlotsAtLevel(level) {
			return pool[level], nil
		}
	}

	return nil, NewSpellSlotErrorOutOfSlots(0)
}

func (s *SpellcastingManager) isAbleToCast(choice spells.SpellChoice) bool {
	if choice.Spell.Level == 0 {
		return true
	}

	return s.HasSpellSlotsAtLevel(choice.Formula.CastLevel)
}

func (s *SpellcastingManager) getHighestLevelSpellChoice(t spells.SpellType) (*spells.SpellChoice, error) {
	pool, err := s.getHighestLevelSpells(t)
	if err != nil {
		return nil, err
	}
	choice, err := s.getHighestAverageAvailableOfPool(pool)
	if err != nil {
		return nil, err
	}
	return choice, nil
}

func (s *SpellcastingManager) getLowestLevelSpellChoice(t spells.SpellType) (*spells.SpellChoice, error) {
	pool, err := s.getLowestLeveledSpells(t)
	if err != nil {
		return nil, err
	}
	choice, err := s.getHighestAverageAvailableOfPool(pool)
	if err != nil {
		return nil, err
	}
	return choice, nil
}

func (s *SpellcastingManager) flattenSpellPool(pool map[int][]*spells.Spell) []*spells.Spell {
	var flattened []*spells.Spell
	for _, spellsAtLevel := range pool {
		flattened = append(flattened, spellsAtLevel...)
	}
	return flattened
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
