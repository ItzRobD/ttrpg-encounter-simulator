package spells

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
)

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
	CasterType   core.CasterType `json:"caster_type"`
	CasterLevel  int             `json:"caster_level"`
	CurrentSlots SpellSlots      `json:"current_slots"`
	MaxSlots     SpellSlots      `json:"max_slots"`
}

type InnateSpell struct {
	Spell          Spell `json:"spell"`
	MaxCastsPerDay int   `json:"max_casts_per_day"`
}

type SpellDC struct {
	Ability   core.Ability     `json:"ability"`
	OnSuccess core.DCOnSuccess `json:"on_success"`
}

func (s SpellDC) GetAbility() core.Ability       { return s.Ability }
func (s SpellDC) GetOnSuccess() core.DCOnSuccess { return s.OnSuccess }
