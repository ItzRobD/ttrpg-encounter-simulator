package spells

import "dnd5e-encounter-simulator-backend/pkg/core"

type SpellSlots map[int]int

//type SpellChoice struct {
//	Spell   *Spell
//	Formula *core.CastFormula
//}
//
//func (sc SpellChoice) GetSpell() *Spell {
//	return sc.Spell
//}
//
//func (sc SpellChoice) GetFormula() *core.CastFormula {
//	return sc.Formula
//}

type SpellcastingManagerStatus struct {
	Parent       core.Entity
	CasterType   core.CasterType
	CasterLevel  int
	CurrentSlots SpellSlots
	MaxSlots     SpellSlots
}
