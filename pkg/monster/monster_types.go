package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
)

type MonsterConfig struct {
	Base               MonsterBase
	Actions            map[int]Action
	Multiattacks       map[int][]Multiattack
	LegendaryActions   []LegendaryAction
	SpecialAbilities   []SpecialAbility
	Resistances        core.DamageResistances
	DamageBreakers     []core.ResistBreaker
	spellcastingConfig MonsterSpellcastingConfig
	HPSetMethod        core.HPSetMethod
	Seed               core.Seed
}

type MonsterSpellcastingConfig struct {
	MonsterID      int
	CastingLevel   int
	Ability        core.Ability
	AttackModifier int
	SaveDC         int
	LeveledSpells  []spells.Spell
	InnateSpells   []spells.InnateSpell
	SpellSlots     spells.SpellSlots
}
