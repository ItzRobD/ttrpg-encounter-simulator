package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"math"
)

type HealingOption struct {
	Spell        *spells.Spell
	Formula      *core.CastFormula
	CastLevel    int
	Efficiency   float64
	TargetDelta  int
	AverageValue int
}

type SpellcastingManager struct {
	parent                 core.Entity
	rollManager            *roll_manager.RollManager
	casterType             core.CasterType
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

func NewSpellcastingManager(parent core.Entity, rm *roll_manager.RollManager, casterType core.CasterType, casterLevel int, currentSlots spells.SpellSlots, maxSlots spells.SpellSlots, canUpcast bool, spellcastModValue int) *SpellcastingManager {
	return &SpellcastingManager{
		parent:                 parent,
		rollManager:            rm,
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

func (scm *SpellcastingManager) SetUsableSpellIDs(ids []int) {
	scm.usableSpellIDs = ids
}

func (scm *SpellcastingManager) AddKnownSpell(spell *spells.Spell) error {
	scm.calculateFormulaAverages(spell)
	if spell.SpellType == core.STHealing {
		scm.healingSpells[spell.Level] = append(scm.healingSpells[spell.Level], spell)
		scm.healingSpellCount++
		return nil
	} else if spell.SpellType == core.STDamage {
		scm.damageSpells[spell.Level] = append(scm.damageSpells[spell.Level], spell)
		scm.healingSpellCount++
		return nil
	}

	return fmt.Errorf("Spells is of non healing or damage type")
}

func (scm *SpellcastingManager) calculateFormulaAverages(spell *spells.Spell) {
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

func (scm *SpellcastingManager) HasHealingSpells() bool {
	return scm.healingSpellCount > 0
}

func (scm *SpellcastingManager) GetHealingSpells() map[int][]*spells.Spell {
	return scm.healingSpells
}

func (scm *SpellcastingManager) GetHealingSpellCount() int {
	return scm.healingSpellCount
}

func (scm *SpellcastingManager) GetHealingCantrips() []*spells.Spell {
	if !scm.HasHealingSpells() {
		return nil
	}
	return scm.healingSpells[0]
}

func (scm *SpellcastingManager) getHealingSpellsByLevel(level int) []*spells.Spell {
	if !scm.HasHealingSpells() {
		return nil
	}
	return scm.healingSpells[level]
}

func (scm *SpellcastingManager) GetHealingSpellsLeveled() []*spells.Spell {
	if !scm.HasHealingSpells() {
		return nil
	}
	var results []*spells.Spell

	for level := 1; level <= 9; level++ {
		spellsAtLevel := scm.getHealingSpellsByLevel(level)
		if spellsAtLevel != nil {
			results = append(results, spellsAtLevel...)
		}
	}

	return results
}

func (scm *SpellcastingManager) HasDamageSpells() bool {
	return scm.damageSpellCount > 0
}

func (scm *SpellcastingManager) HasAnyKnownSpells() bool {
	return scm.HasHealingSpells() || scm.HasDamageSpells()
}

func (scm *SpellcastingManager) GetDamageSpells() map[int][]*spells.Spell {
	return scm.damageSpells
}

func (scm *SpellcastingManager) GetDamageSpellCount() int {
	return scm.damageSpellCount
}

func (scm *SpellcastingManager) GetDamageCantrips() []*spells.Spell {
	if !scm.HasDamageSpells() {
		return nil
	}
	return scm.damageSpells[0]
}

func (scm *SpellcastingManager) getDamageSpellsByLevel(level int) []*spells.Spell {
	if !scm.HasDamageSpells() {
		return nil
	}

	return scm.damageSpells[level]
}

func (scm *SpellcastingManager) GetDamageSpellsLeveled() []*spells.Spell {
	if !scm.HasDamageSpells() {
		return nil
	}
	var results []*spells.Spell

	for level := 1; level <= 9; level++ {
		spellsAtLevel := scm.getDamageSpellsByLevel(level)
		if spellsAtLevel != nil {
			results = append(results, spellsAtLevel...)
		}
	}

	return results
}

func (scm *SpellcastingManager) GetCasterType() core.CasterType {
	return scm.casterType
}

func (scm *SpellcastingManager) GetCasterLevel() int {
	return scm.casterLevel
}

func (scm *SpellcastingManager) GetStatus() *spells.SpellcastingManagerStatus {
	return &spells.SpellcastingManagerStatus{
		Parent:       scm.parent,
		CasterType:   scm.casterType,
		CasterLevel:  scm.casterLevel,
		CurrentSlots: scm.currentSlots,
		MaxSlots:     scm.maxSlots,
	}
}
