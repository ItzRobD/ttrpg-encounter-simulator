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
	ETSpellSlotsEvent   EventType = "spellslots"
	ETHPModifiedEvent   EventType = "hpmodified"
	ETSavingThrowEvent  EventType = "savingthrow"
)

// TODO: Create Event structs for specific purposes. One catch all is not a smart way to handle this
type CombatEvent struct {
	Round        int
	EventType    EventType
	Actor        string
	Target       string
	Attack       string
	Success      bool
	Value        int
	DamageType   string
	Rolls        []int
	IsFatal      bool
	Modifier     int
	SavingThrow  int
	CurrentHP    int
	PreviousHP   int
	ActionChoice shared.ActionType
	SpellChoice  *spells.Spell
	HasSlots     bool
	SpellSlots   shared.SpellSlots
}

type CombatLogger interface {
	LogEvent(event CombatEvent)
}

type CombatListener interface {
	HandleEvent(event CombatEvent)
}
