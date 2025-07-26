package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/monster_action_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
)

type MonsterConfig struct {
	Base               MonsterBase
	Actions            map[int]monster_action_manager.Action
	Multiattacks       map[int][]monster_action_manager.Multiattack
	LegendaryActions   []monster_action_manager.LegendaryAction
	SpecialAbilities   []monster_action_manager.SpecialAbility
	Resistances        core.DamageResistances
	DamageBreakers     []core.ResistBreaker
	spellcastingConfig MonsterSpellcastingConfig
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
