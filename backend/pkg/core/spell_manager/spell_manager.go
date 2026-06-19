package spell_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"math"
)

type SpellManager struct {
	HealingSpells       map[int][]*spells.Spell       `json:"healing_spells"`
	DamageSpells        map[int][]*spells.Spell       `json:"damage_spells"`
	HealingSpellsInnate map[int][]*spells.InnateSpell `json:"healing_spells_innate"`
	DamageSpellsInnate  map[int][]*spells.InnateSpell `json:"damage_spells_innate"`
	DamageSpellCount    int                           `json:"damage_spell_count"`
	HealingSpellCount   int                           `json:"healing_spell_count"`
	SpellcastingAbility core.Ability                  `json:"spellcasting_ability"`
}

func NewSpellManager() SpellManager {
	return SpellManager{
		HealingSpells:       map[int][]*spells.Spell{}, // Key is spell level
		DamageSpells:        map[int][]*spells.Spell{}, // Key is spell level
		HealingSpellsInnate: map[int][]*spells.InnateSpell{},
		DamageSpellsInnate:  map[int][]*spells.InnateSpell{},
	}
}

func (scm *SpellManager) GetSpellCount() int {
	return scm.DamageSpellCount + scm.HealingSpellCount
}
func (scm *SpellManager) GetHealingSpellCount() int {
	return scm.HealingSpellCount
}
func (scm *SpellManager) GetDamageSpellCount() int {
	return scm.DamageSpellCount
}
func (scm *SpellManager) HasHealingSpells() bool { return scm.HealingSpellCount > 0 }
func (scm *SpellManager) HasDamageSpells() bool  { return scm.DamageSpellCount > 0 }

func (scm *SpellManager) GetDamageSpellsByLevel(level int) []*spells.Spell {
	return scm.DamageSpells[level]
}
func (scm *SpellManager) GetHealingSpellsByLevel(level int) []*spells.Spell {
	return scm.HealingSpells[level]
}

func (scm *SpellManager) GetInnateHealingSpellsByLevel(level int) []*spells.InnateSpell {
	return scm.HealingSpellsInnate[level]
}
func (scm *SpellManager) GetInnateDamageSpellsByLevel(level int) []*spells.InnateSpell {
	return scm.DamageSpellsInnate[level]
}

func (scm *SpellManager) GetDamageCantrips() []*spells.Spell {
	return scm.DamageSpells[0]
}
func (scm *SpellManager) GetHealingCantrips() []*spells.Spell {
	return scm.HealingSpells[0]
}

func (scm *SpellManager) GetInnateDamageCantrips() []*spells.InnateSpell {
	return scm.DamageSpellsInnate[0]
}
func (scm *SpellManager) GetInnateHealingCantrips() []*spells.InnateSpell {
	return scm.HealingSpellsInnate[0]
}

func (scm *SpellManager) AddKnownSpell(s *spells.Spell) error {
	scm.calculateFormulaAverages(s)

	var targetMap map[int][]*spells.Spell
	if s.SpellType == core.STHealing {
		targetMap = scm.HealingSpells
	} else if s.SpellType == core.STDamage {
		targetMap = scm.DamageSpells
	} else {
		fmt.Printf("SpellID: %s, Name: %s - is of non healing or damage type. Skipping\n", s.ID, s.Name)
		return nil
	}

	// Deduplicate by ID
	existingList := targetMap[s.Level]
	for i, existing := range existingList {
		if existing.ID == s.ID {
			// Replace with new version (usually enriched from DB)
			existingList[i] = s
			targetMap[s.Level] = existingList
			return nil
		}
	}

	targetMap[s.Level] = append(existingList, s)
	if s.SpellType == core.STHealing {
		scm.HealingSpellCount++
	} else {
		scm.DamageSpellCount++
	}
	return nil
}

func (scm *SpellManager) AddKnownInnateSpell(s *spells.InnateSpell) error {
	scm.calculateFormulaAverages(&s.Spell)

	var targetMap map[int][]*spells.InnateSpell
	if s.Spell.SpellType == core.STHealing {
		targetMap = scm.HealingSpellsInnate
	} else if s.Spell.SpellType == core.STDamage {
		targetMap = scm.DamageSpellsInnate
	} else {
		fmt.Printf("SpellID: %s, Name: %s - is of non healing or damage type. Skipping\n", s.Spell.ID, s.Spell.Name)
		return nil
	}

	existingList := targetMap[s.Spell.Level]
	for i, existing := range existingList {
		if existing.Spell.ID == s.Spell.ID {
			existingList[i] = s
			return nil
		}
	}

	targetMap[s.Spell.Level] = append(existingList, s)
	if s.Spell.SpellType == core.STHealing {
		scm.HealingSpellCount++
	} else {
		scm.DamageSpellCount++
	}
	return nil
}

func (scm *SpellManager) calculateFormulaAverages(s *spells.Spell) {
	for level, formulas := range s.Formulas {
		for i, formula := range formulas {
			dAvg := formula.Die.Avg()
			baseAverage := int(math.Floor(float64(formula.NumberOfDice)*dAvg)) + formula.AmountToAdd

			s.Formulas[level][i].AverageValue = baseAverage
		}
	}
}
