package spellcasting

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"math"
)

type CasterType string

const (
	CasterCharacter         CasterType = "character"
	CasterMonsterInnate                = "innate_monster"
	CasterMonsterTrueCaster            = "spellcaster_monster"
)

type SpellSlots map[int]int

type SpellChoice struct {
	Spell   *spells.Spell
	Formula *spells.CastFormula
}

type SpellcastingManager struct {
	parent                 core.Entity
	casterType             CasterType
	casterLevel            int
	currentSlots           SpellSlots
	maxSlots               SpellSlots
	healingSpells          map[int][]*spells.Spell
	damageSpells           map[int][]*spells.Spell
	canUpcast              bool
	spellcastModifierValue int
}

func NewSpellcastingManager(parent core.Entity, casterType CasterType, casterLevel int, currentSlots SpellSlots, maxSlots SpellSlots, canUpcast bool, spellcastMod int) *SpellcastingManager {
	return &SpellcastingManager{
		parent:                 parent,
		casterType:             casterType,
		casterLevel:            casterLevel,
		currentSlots:           currentSlots,
		maxSlots:               maxSlots,
		canUpcast:              canUpcast,
		spellcastModifierValue: spellcastMod,
		healingSpells:          map[int][]*spells.Spell{},
		damageSpells:           map[int][]*spells.Spell{},
	}
}

func (s *SpellcastingManager) AddKnownSpell(spell *spells.Spell) error {
	s.calculateFormulaAverages(spell)
	if spell.SpellType == string(spells.STHealing) {
		s.healingSpells[spell.Level] = append(s.healingSpells[spell.Level], spell)
		return nil
	} else if spell.SpellType == string(spells.STDamage) {
		s.damageSpells[spell.Level] = append(s.damageSpells[spell.Level], spell)
		return nil
	}

	return fmt.Errorf("Spells is of non healing or damage type")
}

func (s *SpellcastingManager) calculateFormulaAverages(spell *spells.Spell) {
	for level, formula := range spell.Formulas {
		dAvg, err := shared.GetDieAverage(formula.Die)
		if err != nil {
			fmt.Println("Error invalid die")
			continue
		}

		baseAverage := int(math.Floor(float64(formula.NumberOfDice)*dAvg)) + formula.AmountToAdd

		formulaCopy := formula
		formulaCopy.AverageValue = baseAverage
		spell.Formulas[level] = formulaCopy
	}
}

func (s *SpellcastingManager) HasHealingSpells() bool {
	return len(s.healingSpells) > 0
}

func (s *SpellcastingManager) GetHealingSpells() map[int][]*spells.Spell {
	return s.healingSpells
}

func (s *SpellcastingManager) GetHealingCantrips() []*spells.Spell {
	if !s.HasHealingSpells() {
		return nil
	}
	return s.healingSpells[0]
}

func (s *SpellcastingManager) getHealingSpellsByLevel(level int) []*spells.Spell {
	if !s.HasHealingSpells() {
		return nil
	}
	return s.healingSpells[level]
}

func (s *SpellcastingManager) GetHealingSpellsLeveled() []*spells.Spell {
	if !s.HasHealingSpells() {
		return nil
	}
	var results []*spells.Spell

	for level := 1; level <= 9; level++ {
		spellsAtLevel := s.getHealingSpellsByLevel(level)
		if spellsAtLevel != nil {
			results = append(results, spellsAtLevel...)
		}
	}

	return results
}

func (s *SpellcastingManager) HasDamageSpells() bool {
	return len(s.damageSpells) > 0
}

func (s *SpellcastingManager) GetDamageSpells() map[int][]*spells.Spell {
	return s.damageSpells
}

func (s *SpellcastingManager) GetDamageCantrips() []*spells.Spell {
	if !s.HasDamageSpells() {
		return nil
	}
	return s.damageSpells[0]
}

func (s *SpellcastingManager) getDamageSpellsByLevel(level int) []*spells.Spell {
	if !s.HasDamageSpells() {
		return nil
	}

	return s.damageSpells[level]
}

func (s *SpellcastingManager) GetDamageSpellsLeveled() []*spells.Spell {
	if !s.HasDamageSpells() {
		return nil
	}
	var results []*spells.Spell

	for level := 1; level <= 9; level++ {
		spellsAtLevel := s.getDamageSpellsByLevel(level)
		if spellsAtLevel != nil {
			results = append(results, spellsAtLevel...)
		}
	}

	return results
}

func (s *SpellcastingManager) GetCasterType() CasterType {
	return s.casterType
}

func (s *SpellcastingManager) GetCasterLevel() int {
	return s.casterLevel
}
