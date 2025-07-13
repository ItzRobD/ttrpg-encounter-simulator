package spells

import "dnd5e-encounter-simulator-backend/pkg/core"

type CasterType string

const (
	CasterCharacter         CasterType = "character"
	CasterMonsterInnate                = "innate_monster"
	CasterMonsterTrueCaster            = "spellcaster_monster"
)

type SpellSlots map[int]int

type SpellChoice struct {
	Spell   *Spell
	Formula *CastFormula
}

type SpellcastingManagerStatus struct {
	Parent       core.Entity
	CasterType   CasterType
	CasterLevel  int
	CurrentSlots SpellSlots
	MaxSlots     SpellSlots
}
