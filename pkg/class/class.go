package class

import "dnd5e-encounter-simulator-backend/pkg/spells"

type Class struct {
	ID              int
	Name            string
	HitDie          int
	SpellcastingMod string
	Spellcasting    CSpellcasting
}

type CSpellcasting struct {
	ClassHealingSpells []spells.Spell
	ClassDamageSpells  []spells.Spell
	MaxSpellSlots      map[int]map[int]int
}

type ClassQueryParams struct {
	Name string
	ID   int
}

func New(id int, name string, hitDie int, spellcastingMod string) Class {
	return Class{
		// TODO: Fix the ID problem
		ID:              id, // This is wrong because ID is the primary serial key
		Name:            name,
		HitDie:          hitDie,
		SpellcastingMod: spellcastingMod,
	}
}
