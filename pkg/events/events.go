package events

type EventType string

const (
	AttackEvent      EventType = "attack"
	HealEvent        EventType = "heal"
	DamageEvent      EventType = "damage"
	DeathEvent       EventType = "death"
	UnconsciousEvent EventType = "unconscious"
	RollEvent        EventType = "roll"
	HPRollEvent      EventType = "hproll"
)

type CombatEvent struct {
	Round     int
	EventType EventType
	Actor     string
	Target    string
	Hit       bool
	Value     int
	Rolls     []int
	IsFatal   bool
	Added     int
}

type CombatLogger interface {
	LogEvent(event CombatEvent)
}

type CombatListener interface {
	HandleEvent(event CombatEvent)
}
