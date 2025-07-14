package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"errors"
	"math/rand/v2"
)

func (scm *SpellcastingManager) GetMostEfficientHealingSpell(targetValue int) (*core.SpellChoice, error) {
	if !scm.HasHealingSpells() {
		return nil, NewSpellcastingError("", "no healing spells found", ERROR_SPELL_NOT_FOUND)
	}

	pool, err := scm.getSpellPoolOfType(core.STHealing)
	if err != nil {
		return nil, err
	}

	var options []HealingOption

	for _, spellsAtLevel := range pool {
		for _, spell := range spellsAtLevel {
			castLevel, formula, avg := scm.getBestCastOptionForSpell(spell)

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

	bestOption := scm.selectBestHealingOption(options, targetValue)
	return &core.SpellChoice{
		Spell:   bestOption.Spell,
		Formula: bestOption.Formula,
	}, nil
}

func (scm *SpellcastingManager) selectBestHealingOption(options []HealingOption, targetValue int) HealingOption {
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
		return scm.findMostEfficient(exactMatches)
	}

	// Then prefer minimal overheals
	if len(overheals) > 0 {
		return scm.findMinimalOverheal(overheals)
	}

	// Finally, take the best underheal option
	return scm.findBestUnderheal(underheals)

}

func (scm *SpellcastingManager) findMostEfficient(options []HealingOption) HealingOption {
	best := options[0]
	for _, option := range options[1:] {
		if option.Efficiency > best.Efficiency {
			best = option
		}
	}
	return best
}

func (scm *SpellcastingManager) findMinimalOverheal(options []HealingOption) HealingOption {
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

func (scm *SpellcastingManager) findBestUnderheal(options []HealingOption) HealingOption {
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

func (scm *SpellcastingManager) ChooseSpellByPriority(t core.SpellType, priority shared.SpellPriority) (*core.SpellChoice, error) {
	switch priority {
	case shared.SPNoPreference: // Random known spell
		choice, err := scm.getRandomSpellChoice(t, false)
		if err != nil {
			var spellErr *SpellcastingError
			if errors.As(err, &spellErr) && spellErr.Type == ERROR_SPELL_NOT_FOUND {
				choice, err = scm.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
		return choice, nil
	case shared.SPHighestLevel:
		choice, err := scm.getHighestLevelSpellChoice(t)
		if err != nil {
			var spellSlotErr *SpellSlotError
			if errors.As(err, &spellSlotErr) && spellSlotErr.Type == ERROR_NO_SLOTS_AVAILABLE {
				choice, err = scm.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
		return choice, nil
	case shared.SPLowestLevel:
		choice, err := scm.getLowestLevelSpellChoice(t)
		if err != nil {
			var spellSlotErr *SpellSlotError
			if errors.As(err, &spellSlotErr) && spellSlotErr.Type == ERROR_NO_SLOTS_AVAILABLE {
				choice, err = scm.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
		return choice, nil
	case shared.SPCantrip: // Prioritizes highest value cantrip
		choice, err := scm.getHighestAverageCantrip(t)
		if err != nil {
			return nil, err
		}
		return choice, nil
	case shared.SPRandomCantrip:
		choice, err := scm.getRandomCantripChoice(t)
		if err != nil {
			return nil, err
		}
		return choice, nil
	case shared.SPRandomLeveledSpell:
		choice, err := scm.getRandomSpellChoice(t, true)
		if err != nil {
			var spellErr *SpellcastingError
			if errors.As(err, &spellErr) && spellErr.Type == ERROR_SPELL_NOT_FOUND {
				choice, err = scm.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
		return choice, nil
	case shared.SPAreaOfEffect:
		choice, err := scm.getHighestAverageAvailableAOESpellChoice(t)
		if err != nil {
			var spellErr *SpellcastingError
			if errors.As(err, &spellErr) && spellErr.Type == ERROR_SPELL_NOT_FOUND {
				choice, err = scm.getHighestAverageCantrip(t)
				if err != nil {
					return nil, err
				}
				return choice, nil
			}
			return nil, err
		}
		return choice, nil
	case shared.SPHighestDamage:
		choice, err := scm.getHighestAverageAvailableSpellChoice(t)
		if err != nil {
			var spellErr *SpellcastingError
			if errors.As(err, &spellErr) && spellErr.Type == ERROR_SPELL_NOT_FOUND {
				choice, err = scm.getHighestAverageCantrip(t)
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

func (scm *SpellcastingManager) getHighestAverageAvailableAOESpellChoice(t core.SpellType) (*core.SpellChoice, error) {
	pool, err := scm.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no aoe spells found", ERROR_SPELL_NOT_FOUND)
	}

	var highestSpell *spells.Spell
	var highestFormula *core.CastFormula
	var highestValue int

	for _, spellsAtLevel := range pool {
		for _, spell := range spellsAtLevel {
			if !spell.IsAOE {
				continue
			}

			castLevel, formula, value := scm.getBestCastOptionForSpell(spell)
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

	return &core.SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
}

// getHighestAverageAvailableOfPool determines the spell from the given pool with the highest average output and returns it.
// This method evaluates both leveled spells and cantrips, considering the spell's formulas and caster level.
// Returns a SpellChoice containing the selected spell and its formula or an error if no valid spell is found.
func (scm *SpellcastingManager) getHighestAverageAvailableOfPool(pool []*spells.Spell) (*core.SpellChoice, error) {
	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found", ERROR_SPELL_NOT_KNOWN)
	}

	var highestSpell *spells.Spell
	var highestFormula *core.CastFormula
	var highestValue int
	casterLevel := scm.GetCasterLevel()
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
			castLevel, formula, value := scm.getBestCastOptionForSpell(spell)
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

	return &core.SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
}

func (scm *SpellcastingManager) GetBestFormulaForSpell(spell spells.Spell, p shared.SpellPriority) (*core.CastFormula, error) {
	var formula *core.CastFormula
	var err error
	var value int

	if spell.Level != 0 && !scm.HasAnySpellSlots() {
		return nil, NewSpellSlotError(0, "actor has no available spell slots", ERROR_SLOT_NOT_AVAILABLE)
	}

	if p != shared.SPLowestLevel && p != shared.SPHighestLevel {
		return nil, NewSpellcastingError(spell.Name, "cannot get best formula for spell - invalid priority", ERROR_GENERIC_SPELL)
	}

	switch p {
	case shared.SPLowestLevel:
		if spell.Level == 0 {
			value, formula, err = spell.GetAverageDamageCantrip(scm.parent.GetCasterLevel(), scm.spellcastModifierValue)
			if err != nil {
				return nil, err
			}
		} else {
			var canCast bool
			castLevel := spell.Level
			for !canCast {
				if scm.HasSpellSlotsAtLevel(castLevel) {
					canCast = true
					value, formula, err = spell.GetAverageDamageAtLevel(castLevel, scm.spellcastModifierValue)
					if err != nil {
						return nil, err
					}
				} else if scm.canUpcast && castLevel < 9 {
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
			value, formula, err = spell.GetAverageDamageCantrip(scm.parent.GetCasterLevel(), scm.spellcastModifierValue)
			if err != nil {
				return nil, err
			}
		} else {
			var canCast bool
			castLevel := spell.Level
			for !canCast {
				if scm.HasSpellSlotsAtLevel(castLevel) {
					canCast = true
					value, formula, err = spell.GetAverageDamageAtLevel(castLevel, scm.spellcastModifierValue)
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
func (scm *SpellcastingManager) getRandomCantripChoice(t core.SpellType) (*core.SpellChoice, error) {
	pool, err := scm.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	var castableChoices []*core.SpellChoice
	if cantrips, exists := pool[0]; exists {
		for _, spell := range cantrips {
			castLevel, formula, _ := scm.getBestCastOptionForSpell(spell)
			if castLevel != -1 {
				castableChoices = append(castableChoices, &core.SpellChoice{Spell: spell, Formula: formula})
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
func (scm *SpellcastingManager) getRandomSpellChoice(t core.SpellType, excludeCantrips bool) (*core.SpellChoice, error) {
	pool, err := scm.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_FOUND)
	}

	var castableChoices []*core.SpellChoice
	for level, spellsAtLevel := range pool {
		if level == 0 && excludeCantrips {
			continue
		}
		for _, spell := range spellsAtLevel {
			castLevel, formula, _ := scm.getBestCastOptionForSpell(spell)
			if castLevel != -1 {
				castableChoices = append(castableChoices, &core.SpellChoice{Spell: spell, Formula: formula})
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

func (scm *SpellcastingManager) getHighestAverageAvailableSpellChoice(t core.SpellType) (*core.SpellChoice, error) {
	pool, err := scm.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	var highestSpell *spells.Spell
	var highestFormula *core.CastFormula
	var highestValue int

	for _, spellsAtLevel := range pool {
		for _, spell := range spellsAtLevel {
			castLevel, formula, value := scm.getBestCastOptionForSpell(spell)
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

	return &core.SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
}

// getBestCastOptionForSpell determines the optimal way to cast a spell based on level, available slots, and upcast options.
// Returns the cast level, formula, and average value of the spell.
// Returns -1 if unable to cast the spell || error
func (scm *SpellcastingManager) getBestCastOptionForSpell(spell *spells.Spell) (int, *core.CastFormula, int) {
	if spell.Level == 0 {
		formula, err := spell.GetFormulaForCantrip(scm.casterLevel)
		if err != nil {
			return -1, nil, 0
		}
		return scm.casterLevel, formula, formula.AverageValue
	}

	if !scm.HasAnySpellSlots() {
		return -1, nil, 0
	}

	if scm.canUpcast {
		for level := 9; level >= spell.Level; level-- {
			formula, err := spell.GetClosestFormulaToLevel(level)
			if err != nil {
				continue
			}

			if !scm.HasSpellSlotsAtLevel(formula.CastLevel) {
				continue
			}

			return formula.CastLevel, formula, formula.AverageValue
		}
	} else {
		if scm.HasSpellSlotsAtLevel(spell.Level) {
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

func (scm *SpellcastingManager) getHighestAverageCantrip(t core.SpellType) (*core.SpellChoice, error) {
	var pool []*spells.Spell
	switch t {
	case core.STHealing:
		pool = scm.GetHealingCantrips()
	case core.STDamage:
		pool = scm.GetDamageCantrips()
	default:
		return nil, NewSpellcastingError("", "invalid spell type", ERROR_GENERIC_SPELL)
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	var highestSpell *spells.Spell
	var highestFormula *core.CastFormula
	var highestValue int
	casterLevel := scm.GetCasterLevel()

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

	return &core.SpellChoice{Spell: highestSpell, Formula: highestFormula}, nil
}

func (scm *SpellcastingManager) getHighestLevelSpells(t core.SpellType) ([]*spells.Spell, error) {
	pool, err := scm.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	for level := 9; level > 0; level-- {
		if len(pool[level]) > 0 && scm.HasSpellSlotsAtLevel(level) {
			return pool[level], nil
		}
	}

	return nil, NewSpellSlotErrorOutOfSlots(0)
}

func (scm *SpellcastingManager) getLowestLeveledSpells(t core.SpellType) ([]*spells.Spell, error) {
	pool, err := scm.getSpellPoolOfType(t)
	if err != nil {
		return nil, err
	}

	if len(pool) == 0 {
		return nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	for level := 1; level <= 9; level++ {
		if len(pool[level]) > 0 && scm.HasSpellSlotsAtLevel(level) {
			return pool[level], nil
		}
	}

	return nil, NewSpellSlotErrorOutOfSlots(0)
}

func (scm *SpellcastingManager) isAbleToCast(choice core.SpellChoice) bool {
	if choice.Spell.GetLevel() == 0 {
		return true
	}

	return scm.HasSpellSlotsAtLevel(choice.Formula.CastLevel)
}

func (scm *SpellcastingManager) getHighestLevelSpellChoice(t core.SpellType) (*core.SpellChoice, error) {
	pool, err := scm.getHighestLevelSpells(t)
	if err != nil {
		return nil, err
	}
	choice, err := scm.getHighestAverageAvailableOfPool(pool)
	if err != nil {
		return nil, err
	}
	return choice, nil
}

func (scm *SpellcastingManager) getLowestLevelSpellChoice(t core.SpellType) (*core.SpellChoice, error) {
	pool, err := scm.getLowestLeveledSpells(t)
	if err != nil {
		return nil, err
	}
	choice, err := scm.getHighestAverageAvailableOfPool(pool)
	if err != nil {
		return nil, err
	}
	return choice, nil
}

func (scm *SpellcastingManager) flattenSpellPool(pool map[int][]*spells.Spell) []*spells.Spell {
	var flattened []*spells.Spell
	for _, spellsAtLevel := range pool {
		flattened = append(flattened, spellsAtLevel...)
	}
	return flattened
}

func (scm *SpellcastingManager) getSpellPoolOfType(t core.SpellType) (map[int][]*spells.Spell, error) {
	switch t {
	case core.STHealing:
		return scm.GetHealingSpells(), nil
	case core.STDamage:
		return scm.GetDamageSpells(), nil
	default:
		return nil, NewSpellcastingError("", "invalid spell type", ERROR_GENERIC_SPELL)
	}
}
