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
