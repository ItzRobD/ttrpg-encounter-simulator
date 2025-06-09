package spellcasting

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
)

type CasterType string

const (
	CasterCharacter         CasterType = "character"
	CasterMonsterInnate                = "innate_monster"
	CasterMonsterTrueCaster            = "spellcaster_monster"
)

type SpellSlots map[int]int

type SpellcastingManager struct {
	parent                 core.Entity
	casterType             CasterType
	casterLevel            int
	currentSlots           SpellSlots
	maxSlots               SpellSlots
	healingSpells          []*spells.Spell
	damageSpells           []*spells.Spell
	canUpcast              bool
	spellcastModifierValue int
}

func NewSpellcastingManager(parent *core.Entity, casterType CasterType, casterLevel int, currentSlots SpellSlots, maxSlots SpellSlots, healingSpells []*spells.Spell, damageSpells []*spells.Spell, canUpcast bool, spellcastMod int) *SpellcastingManager {
	return &SpellcastingManager{
		parent:                 parent,
		casterType:             casterType,
		casterLevel:            casterLevel,
		currentSlots:           currentSlots,
		maxSlots:               maxSlots,
		healingSpells:          healingSpells,
		damageSpells:           damageSpells,
		canUpcast:              canUpcast,
		spellcastModifierValue: spellcastMod,
	}
}

func (s *SpellcastingManager) AddKnownSpell(spell *spells.Spell) error {
	if spell.SpellType == "healing" {
		s.healingSpells = append(s.healingSpells, spell)
		return nil
	} else if spell.SpellType == "damage" {
		s.damageSpells = append(s.damageSpells, spell)
		return nil
	}

	return fmt.Errorf("Spells is of non healing or damage type")
}

func (s *SpellcastingManager) HasHealingSpells() bool {
	return len(s.healingSpells) > 0
}

func (s *SpellcastingManager) GetHealingSpells() []*spells.Spell {
	return s.healingSpells
}

func (s *SpellcastingManager) GetHealingCantrips() []*spells.Spell {
	if !s.HasHealingSpells() {
		return nil
	}
	var spells []*spells.Spell
	for _, spell := range s.healingSpells {
		if spell.Level == 0 {
			spells = append(spells, spell)
		}
	}
	return spells
}

func (s *SpellcastingManager) GetHealingSpellsByLevel(level int) []*spells.Spell {
	if !s.HasHealingSpells() {
		return nil
	}
	var spells []*spells.Spell
	for _, spell := range s.healingSpells {
		if spell.Level == level {
			spells = append(spells, spell)
		}
	}
	return spells
}

func (s *SpellcastingManager) GetHealingSpellsLeveled() []*spells.Spell {
	if !s.HasHealingSpells() {
		return nil
	}
	var spells []*spells.Spell
	for _, spell := range s.healingSpells {
		if spell.Level > 0 {
			spells = append(spells, spell)
		}
	}
	return spells
}

func (s *SpellcastingManager) HasDamageSpells() bool {
	return len(s.damageSpells) > 0
}

func (s *SpellcastingManager) GetDamageSpells() []*spells.Spell {
	return s.damageSpells
}

func (s *SpellcastingManager) GetDamageCantrips() []*spells.Spell {
	if !s.HasDamageSpells() {
		return nil
	}
	var spells []*spells.Spell
	for _, spell := range s.damageSpells {
		if spell.Level == 0 {
			spells = append(spells, spell)
		}
	}
	return spells
}

func (s *SpellcastingManager) GetDamageSpellsByLevel(level int) []*spells.Spell {
	if !s.HasDamageSpells() {
		return nil
	}
	var spells []*spells.Spell
	for _, spell := range s.damageSpells {
		if spell.Level == level {
			spells = append(spells, spell)
		}
	}
	return spells
}

func (s *SpellcastingManager) GetDamageSpellsLeveled() []*spells.Spell {
	if !s.HasDamageSpells() {
		return nil
	}
	var spells []*spells.Spell
	for _, spell := range s.damageSpells {
		if spell.Level > 0 {
			spells = append(spells, spell)
		}
	}
	return spells
}

func (s *SpellcastingManager) GetCasterType() CasterType {
	return s.casterType
}

func (s *SpellcastingManager) GetCasterLevel() int {
	return s.casterLevel
}
