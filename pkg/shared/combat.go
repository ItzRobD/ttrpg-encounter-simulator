package shared

import "dnd5e-encounter-simulator-backend/pkg/simulation/events"

type Entity interface {
	ModifyHP(amount int)
	IsUnconscious() bool
	GetCurrentHP() int
	GetCurrentHPPct() int
	GetMaxHP() int
	GetName() string
	GetAC() int
	GetEventListener() func(event events.CombatEvent)
}

type Combatant struct {
	InitiativeScore int
	Creature        Entity
}
