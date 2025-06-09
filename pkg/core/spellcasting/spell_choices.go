package spellcasting

import (
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"math/rand/v2"
)

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
		rand.NewPCG(rand.Uint64(), rand.Uint64())
		i := rand.IntN(len(pool))
		selected = pool[i]
		return selected, nil
	case spells.STDamage:
		pool := s.GetDamageSpells()
		if len(pool) == 0 {
			return nil, NewSpellcastingError("", "no damage spells found", ERROR_SPELL_NOT_KNOWN)
		}
		rand.NewPCG(rand.Uint64(), rand.Uint64())
		i := rand.IntN(len(pool))
		selected = pool[i]
		return selected, nil
	default:
		return nil, NewSpellcastingError("", "invalid spell type", ERROR_GENERIC_SPELL)
	}
}

func (s *SpellcastingManager) GetRandomAOESpell(t spells.SpellType) (*spells.Spell, error) {
	var selected *spells.Spell
	switch t {
	case spells.STHealing:
		var pool []*spells.Spell
		for _, spell := range s.GetHealingSpells() {
			if spell.IsAOE {
				pool = append(pool, spell)
			}
		}
		if len(pool) == 0 {
			return nil, NewSpellcastingError("", "no AOE healing spells found", ERROR_SPELL_NOT_KNOWN)
		}
		rand.NewPCG(rand.Uint64(), rand.Uint64())
		i := rand.IntN(len(pool))
		selected = pool[i]
		return selected, nil
	case spells.STDamage:
		var pool []*spells.Spell
		for _, spell := range s.GetDamageSpells() {
			if spell.IsAOE {
				pool = append(pool, spell)
			}
		}
		if len(pool) == 0 {
			return nil, NewSpellcastingError("", "no AOE damage spells found", ERROR_SPELL_NOT_KNOWN)
		}
		rand.NewPCG(rand.Uint64(), rand.Uint64())
		i := rand.IntN(len(pool))
		selected = pool[i]
		return selected, nil
	default:
		return nil, NewSpellcastingError("", "invalid spell type", ERROR_GENERIC_SPELL)
	}
}

func (s *SpellcastingManager) GetHighestAverageSpell(t spells.SpellType) (*spells.Spell, *spells.CastFormula, error) {
	var pool []*spells.Spell

	switch t {
	case spells.STHealing:
		pool = s.GetHealingSpells()
	case spells.STDamage:
		pool = s.GetDamageSpells()
	default:
		return nil, nil, NewSpellcastingError("", "invalid spell type", ERROR_GENERIC_SPELL)
	}

	if len(pool) == 0 {
		return nil, nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	var highestSpell *spells.Spell
	var highestFormula *spells.CastFormula
	var highestValue int
	var highestAvailableSlot int
	var hasSpellSlots bool

	if s.canUpcast {
		slot, err := s.GetHighestAvailableSpellSlot()
		if err != nil {
			hasSpellSlots = false
		}
		highestAvailableSlot = slot
		hasSpellSlots = true
	}

	for _, spell := range pool {
		if s.canUpcast && hasSpellSlots {
			v, f, err := spell.GetAverageDamageAtLevel(highestAvailableSlot, s.spellcastModifierValue)
			if err != nil {
				return nil, nil, err
			}

			if !s.HasSpellSlotsAtLevel(f.CastLevel) {
				continue
			}

			if v > highestValue {
				highestValue = v
				highestSpell = spell
				highestFormula = f
			}
		} else {
			v, f, err := spell.GetAverageDamageAtLevel(spell.Level, s.spellcastModifierValue)
			if err != nil {
				return nil, nil, err
			}

			if !s.HasSpellSlotsAtLevel(f.CastLevel) {
				continue
			}

			if v > highestValue {
				highestValue = v
				highestSpell = spell
				highestFormula = f
			}
		}
	}

	if highestSpell == nil {
		return nil, nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	return highestSpell, highestFormula, nil
}

func (s *SpellcastingManager) GetHighestAverageCantrip(t spells.SpellType) (*spells.Spell, *spells.CastFormula, error) {
	var pool []*spells.Spell
	switch t {
	case spells.STHealing:
		pool = s.GetHealingCantrips()
	case spells.STDamage:
		pool = s.GetDamageCantrips()
	default:
		return nil, nil, NewSpellcastingError("", "invalid spell type", ERROR_GENERIC_SPELL)
	}

	if len(pool) == 0 {
		return nil, nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	var highestSpell *spells.Spell
	var highestFormula *spells.CastFormula
	var highestValue int
	casterLevel := s.GetCasterLevel()

	for _, spell := range pool {
		v, f, err := spell.GetAverageDamageCantrip(casterLevel, s.spellcastModifierValue)
		if err != nil {
			return nil, nil, err
		}

		if v > highestValue {
			highestValue = v
			highestSpell = spell
			highestFormula = f
		}
	}

	if highestSpell == nil {
		return nil, nil, NewSpellcastingError("", "no spells found of type", ERROR_SPELL_NOT_KNOWN)
	}

	return highestSpell, highestFormula, nil
}
