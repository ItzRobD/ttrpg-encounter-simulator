package shared

type Entity interface {
	ModifyHP(amount int)
	IsUnconscious() bool
	GetCurrentHP() int
	GetCurrentHPPct() int
	GetMaxHP() int
	GetName() string
}

type Combatant struct {
	InitiativeScore int
	Creature        Entity
}
