package spells

import "dnd5e-encounter-simulator-backend/pkg/core"

type SpellSlots map[int]int

type SpellcastingManagerStatus struct {
	Parent       core.Entity
	CasterType   core.CasterType
	CasterLevel  int
	CurrentSlots SpellSlots
	MaxSlots     SpellSlots
}

type InnateSpell struct {
	Spell          Spell
	MaxCastsPerDay int
	CastsRemaining int
}
