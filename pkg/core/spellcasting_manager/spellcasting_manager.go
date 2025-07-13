package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"math"
)

type HealingOption struct {
	Spell        *spells.Spell
	Formula      *spells.CastFormula
	CastLevel    int
	Efficiency   float64
	TargetDelta  int
	AverageValue int
}

type SpellcastingManager struct {
	parent                 core.Entity
	casterType             spells.CasterType
	casterLevel            int
	currentSlots           spells.SpellSlots
	maxSlots               spells.SpellSlots
	healingSpells          map[int][]*spells.Spell
	damageSpells           map[int][]*spells.Spell
	damageSpellCount       int
	healingSpellCount      int
	usableSpellIDs         []int // TODO: Not currently being used for anything
	canUpcast              bool
	spellcastModifierValue int
}

func NewSpellcastingManager(parent core.Entity, casterType spells.CasterType, casterLevel int, currentSlots spells.SpellSlots, maxSlots spells.SpellSlots, canUpcast bool, spellcastModValue int) *SpellcastingManager {
	return &SpellcastingManager{
		parent:                 parent,
		casterType:             casterType,
		casterLevel:            casterLevel,
		currentSlots:           currentSlots,
		maxSlots:               maxSlots,
		canUpcast:              canUpcast,
		spellcastModifierValue: spellcastModValue,
		healingSpells:          map[int][]*spells.Spell{},
		damageSpells:           map[int][]*spells.Spell{},
	}
}

func (s *SpellcastingManager) SetUsableSpellIDs(ids []int) {
	s.usableSpellIDs = ids
}

func (s *SpellcastingManager) AddKnownSpell(spell *spells.Spell) error {
	s.calculateFormulaAverages(spell)
	if spell.SpellType == string(spells.STHealing) {
		s.healingSpells[spell.Level] = append(s.healingSpells[spell.Level], spell)
		s.healingSpellCount++
		return nil
	} else if spell.SpellType == string(spells.STDamage) {
		s.damageSpells[spell.Level] = append(s.damageSpells[spell.Level], spell)
		s.healingSpellCount++
		return nil
	}

	return fmt.Errorf("Spells is of non healing or damage type")
}

func (s *SpellcastingManager) calculateFormulaAverages(spell *spells.Spell) {
	for level, formula := range spell.Formulas {
		dAvg, err := core.GetDieAverage(formula.Die)
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
	return s.healingSpellCount > 0
}

func (s *SpellcastingManager) GetHealingSpells() map[int][]*spells.Spell {
	return s.healingSpells
}

func (s *SpellcastingManager) GetHealingSpellCount() int {
	return s.healingSpellCount
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
	return s.damageSpellCount > 0
}

func (s *SpellcastingManager) HasAnyKnownSpells() bool {
	return s.HasHealingSpells() || s.HasDamageSpells()
}

func (s *SpellcastingManager) GetDamageSpells() map[int][]*spells.Spell {
	return s.damageSpells
}

func (s *SpellcastingManager) GetDamageSpellCount() int {
	return s.damageSpellCount
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

func (s *SpellcastingManager) GetCasterType() spells.CasterType {
	return s.casterType
}

func (s *SpellcastingManager) GetCasterLevel() int {
	return s.casterLevel
}

func (s *SpellcastingManager) GetStatus() *spells.SpellcastingManagerStatus {
	return &spells.SpellcastingManagerStatus{
		Parent:       s.parent,
		CasterType:   s.casterType,
		CasterLevel:  s.casterLevel,
		CurrentSlots: s.currentSlots,
		MaxSlots:     s.maxSlots,
	}
}
