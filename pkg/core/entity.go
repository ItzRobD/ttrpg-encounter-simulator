package core

type Entity interface {
	ModifyHP(amount int)
	IsUnconscious() bool
	GetCurrentHP() int
	GetCurrentHPPct() int
	GetMaxHP() int
	GetName() string
	GetAC() int
	GetEventListener() func(event interface{})
	GetSavingThrowRollResult(ability string) (int, error)
}

type Combatant struct {
	InitiativeScore int
	Creature        Entity
}
