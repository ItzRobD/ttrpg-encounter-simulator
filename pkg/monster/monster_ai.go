package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"math/rand/v2"
)

type MonsterAI struct {
	parent    *Monster
	combatCtx *core.CombatContext
	rng       *rand.Rand
}

func NewMonsterAI(m *Monster) *MonsterAI {
	return &MonsterAI{
		parent: m,
		rng:    m.GetRNG(),
	}
}
