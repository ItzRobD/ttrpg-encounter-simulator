package events

type EventType string

const (
	AttackEvent      EventType = "attack"
	SpellAttack      EventType = "spellattack"
	SpellDC          EventType = "spelldc"
	HealEvent        EventType = "heal"
	DamageEvent      EventType = "damage"
	DeathEvent       EventType = "death"
	UnconsciousEvent EventType = "unconscious"
	RollEvent        EventType = "roll"
	HPRollEvent      EventType = "hproll"
)

type CombatEvent struct {
	Round       int
	EventType   EventType
	Actor       string
	Target      string
	Attack      string
	Hit         bool
	Value       int
	DamageType  string
	Rolls       []int
	IsFatal     bool
	Added       int
	SavingThrow int
}

type CombatLogger interface {
	LogEvent(event CombatEvent)
}

type CombatListener interface {
	HandleEvent(event CombatEvent)
}
