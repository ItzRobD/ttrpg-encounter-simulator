package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"math/rand/v2"
)

func spellExists(spells []spells.Spell, spellID int) bool {
	for _, s := range spells {
		if s.ID == spellID {
			return true
		}
	}
	return false
}

func isValidSpell(c *Character, spellID int) bool {
	if spellID < 0 || spellID > 319 {
		return false
	}

	if spellExists(c.Class.Spellcasting.ClassHealingSpells, spellID) {
		return true
	}
	if spellExists(c.Class.Spellcasting.ClassDamageSpells, spellID) {
		return true
	}

	return false
}

func (c *Character) AddKnownSpell(ctx context.Context, spellID int) error {
	if !isValidSpell(c, spellID) {
		return fmt.Errorf("spell id invalid or not available to class: %d", spellID)
	}
	var params spells.SpellQueryParams
	params.ID = spellID
	// TODO: Perform tests on damage formulae
	s, err := spells.QuerySpellData(ctx, params)
	if err != nil {
		return fmt.Errorf("error querying spell data: %w", err)
	}

	if s.LevelType == "character" {
		formula, err2 := spells.GetSpellFormulaByLevel(ctx, spellID, c.Level)
		if err2 != nil {
			return fmt.Errorf("error getting spell formula by level: %w", err2)
		}
		s.CastFormula = *formula
	}

	c.KnownSpells = append(c.KnownSpells, s)
	return nil
}

func (c *Character) getKnownHealingSpells() ([]*spells.Spell, error) {
	if len(c.KnownSpells) <= 0 {
		return nil, fmt.Errorf("character has no known spells")
	}
	var healingSpells []*spells.Spell
	for _, s := range c.KnownSpells {
		if s.SpellType == "healing" {
			healingSpells = append(healingSpells, &s)
		}
	}
	if len(healingSpells) <= 0 {
		return nil, fmt.Errorf("character has no known healing spells")
	}
	return healingSpells, nil
}

func (c *Character) getKnownDamageSpells() ([]*spells.Spell, error) {
	if len(c.KnownSpells) <= 0 {
		return nil, fmt.Errorf("character has no known spells")
	}
	var damageSpells []*spells.Spell
	for _, s := range c.KnownSpells {
		if s.SpellType == "damage" {
			damageSpells = append(damageSpells, &s)
		}
	}
	if len(damageSpells) <= 0 {
		return nil, fmt.Errorf("character has no known damage spells")
	}
	return damageSpells, nil
}

func (c *Character) getSpellPoolOfType(spellType spells.SpellType) ([]*spells.Spell, error) {
	var spellPool []*spells.Spell
	switch spellType {
	case spells.STDamage:
		var err error
		spellPool, err = c.getKnownDamageSpells()
		if err != nil {
			return nil, fmt.Errorf("error getting known damage spells: %w", err)
		}
	case spells.STHealing:
		var err error
		spellPool, err = c.getKnownHealingSpells()
		if err != nil {
			return nil, fmt.Errorf("error getting known healing spells: %w", err)
		}
	}
	if len(spellPool) <= 0 {
		return nil, fmt.Errorf("character has no known spells of type %s", spellType)
	}
	return spellPool, nil
}

func (c *Character) getRandomCantrip(spellPool []*spells.Spell) (*spells.Spell, error) {
	if len(spellPool) <= 0 {
		return nil, fmt.Errorf("character has no known spells")
	}
	var cantrips []*spells.Spell
	for _, s := range spellPool {
		if s.Level == 0 {
			cantrips = append(cantrips, s)
		}
	}
	if len(cantrips) <= 0 {
		return nil, fmt.Errorf("character has no known cantrips")
	}
	rand.NewPCG(rand.Uint64(), rand.Uint64())
	i := rand.IntN(len(cantrips))
	return cantrips[i], nil
}

func (c *Character) getRandomSpell(spellPool []*spells.Spell) (*spells.Spell, error) {
	if len(spellPool) <= 0 {
		return nil, fmt.Errorf("character has no known spells")
	}
	rand.NewPCG(rand.Uint64(), rand.Uint64())
	i := rand.IntN(len(spellPool))
	return spellPool[i], nil
}

func (c *Character) hasSpellSlotAvailable(level int) bool {
	if c.SpellSlots[level] > 0 {
		return true
	}
	return false
}

func (c *Character) getHighestAverageSpell(spellPool []*spells.Spell) (*spells.Spell, error) {
	if len(spellPool) <= 0 {
		return nil, fmt.Errorf("character has no known spells")
	}
	var highestSpell *spells.Spell
	for _, s := range spellPool {
		sLvl := s.Level
		if !c.hasSpellSlotAvailable(sLvl) {
			// Character does not have a spell slot to cast this level of a spell
			// TODO: Add logging for spell selection
			continue
		}
		if highestSpell == nil || s.GetAverageAmount() > highestSpell.GetAverageAmount() {
			highestSpell = s
		}
	}
	if highestSpell == nil {
		return nil, fmt.Errorf("character has no known spells of highest level")
	}
	return highestSpell, nil
}

func (c *Character) getHighestCastableVersion(spell spells.Spell) (*spells.Spell, error) {
	if spell.LevelType == "character" {
		formula, err := spells.GetSpellFormulaByLevel(context.Background(), spell.ID, c.Level)
		if err != nil {
			return nil, fmt.Errorf("error getting spell formula by level: %w", err)
		}
		spell.CastFormula = *formula
		return &spell, nil
	} else if spell.LevelType == "slot" {
		formula, err := spells.GetSpellFormulaByLevel(context.Background(), spell.ID, c.getHighestAvailableSpellSlot())
		if err != nil {
			return nil, fmt.Errorf("error getting spell formula by level: %w", err)
		}
		spell.CastFormula = *formula
		return &spell, nil
	}
	return nil, fmt.Errorf("spell level type not supported: %s", spell.LevelType)
}

func (c *Character) getHighestAvailableSpellSlot() int {
	highestSlot := -1
	for level, count := range c.SpellSlots {
		if count > 0 && level > highestSlot {
			highestSlot = level
		}
	}
	return highestSlot
}
