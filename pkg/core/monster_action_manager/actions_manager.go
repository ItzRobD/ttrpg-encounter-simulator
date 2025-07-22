package monster_action_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/monster"
)

type MonsterActionManager struct {
	parent *core.Entity

	ActionList       []monster.MonsterAction
	Mulitattacks     []monster.MonsterMultiattack
	LegendaryActions []monster.LegendaryAction
	SpecialAbilities []monster.SpecialAbility
}
