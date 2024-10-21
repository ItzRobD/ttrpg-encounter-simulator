package shared

type Entity interface {
	ModifyHP(amount int)
	IsUnconscious() bool
	GetName() string
	GetCurrentHP() int
	GetMaxHP() int
}

type Combatant struct {
	InitiativeScore int
	Creature        Entity
}
