package simulation

import "dnd5e-encounter-simulator-backend/pkg/core"

type CombatContext struct {
	AllCombatants  map[int]core.Combatant
	CurrentRound   int
	ActingEntityID int
}
