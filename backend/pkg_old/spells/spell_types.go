package spells

import "dnd5e-encounter-simulator-backend/pkg_old/core"

type SpellSlots map[int]int

type SpellSummary struct {
	ID              int            `json:"id"`
	Name            string         `json:"name"`
	IsConcentration bool           `json:"is_concentration"`
	Level           int            `json:"level"`
	SpellType       core.SpellType `json:"spell_type"`
	IsAOE           bool           `json:"is_aoe"`
	HasDC           bool           `json:"has_dc"`
	IsCustom        bool           `json:"is_custom"`
}

type SpellcastingManagerStatus struct {
	Parent       core.Entity     `json:"-"`
	CasterType   core.CasterType `json:"caster_type"`
	CasterLevel  int             `json:"caster_level"`
	CurrentSlots SpellSlots      `json:"current_slots"`
	MaxSlots     SpellSlots      `json:"max_slots"`
}

type InnateSpell struct {
	Spell          Spell `json:"spell"`
	MaxCastsPerDay int   `json:"max_casts_per_day"`
	CastsRemaining int   `json:"casts_remaining"`
}
