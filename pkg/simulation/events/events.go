package events

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
)

type EventType string

const (
	ETAttackEvent       EventType = "attack"
	ETSpellAttack       EventType = "spellattack"
	ETSpellDC           EventType = "spelldc"
	ETHealEvent         EventType = "heal"
	ETDamageEvent       EventType = "damage"
	ETDeathEvent        EventType = "death"
	ETUnconsciousEvent  EventType = "unconscious"
	ETRollEvent         EventType = "roll"
	ETHPRollEvent       EventType = "hproll"
	ETActionChoiceEvent EventType = "actionchoice"
	ETSpellChoiceEvent  EventType = "spellchoice"
)

type CombatEvent struct {
	Round        int
	EventType    EventType
	Actor        string
	Target       string
	Attack       string
	Hit          bool
	Value        int
	DamageType   string
	Rolls        []int
	IsFatal      bool
	Added        int
	SavingThrow  int
	ActionChoice shared.ActionType
	SpellChoice  *spells.Spell
}

type CombatLogger interface {
	LogEvent(event CombatEvent)
}

type CombatListener interface {
	HandleEvent(event CombatEvent)
}
