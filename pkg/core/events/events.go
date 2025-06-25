package events

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"time"
)

type EventType string

const (
	ETAttackEvent       EventType = "attack"
	ETSpellAttackEvent  EventType = "spellattack"
	ETSpellDCEvent      EventType = "spelldc"
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
	ETTargetChoiceEvent EventType = "targetchoice"
)

// TODO: Create Event structs for specific purposes. One catch all is not a smart way to handle this
//type CombatEvent struct {
//	Round        int
//	EventType    EventType
//	Actor        string
//	Target       string
//	Attack       string
//	Success      bool
//	Value        int
//	DamageType   string
//	Rolls        []int
//	IsFatal      bool
//	Modifier     int
//	SavingThrow  int
//	CurrentHP    int
//	PreviousHP   int
//	ActionChoice shared.ActionType
//	SpellChoice  *spells.Spell
//	HasSlots     bool
//	SpellSlots   shared.SpellSlots
//}

type CombatEvent interface {
	GetRound() int
	GetActor() string
	GetTimestamp() time.Time
	SetRound(int)
	SetActor(string)
	SetTimestamp(time.Time)
}

type BaseEvent struct {
	Round     int
	Timestamp time.Time
	Actor     string
}

func (b *BaseEvent) GetRound() int {
	return b.Round
}
func (b *BaseEvent) GetTimestamp() time.Time {
	return b.Timestamp
}
func (b *BaseEvent) GetActor() string {
	return b.Actor
}
func (b *BaseEvent) SetRound(round int) {
	b.Round = round
}
func (b *BaseEvent) SetTimestamp(timestamp time.Time) {
	b.Timestamp = timestamp
}
func (b *BaseEvent) SetActor(actor string) {
	b.Actor = actor
}

type MeleeAttackEvent struct {
	BaseEvent
	Target         string
	AttackName     string
	AttackCount    int
	AttackRoll     int
	AttackModifier int
	AttackTotal    int
	Success        bool
	CriticalHit    bool
}

func (e *MeleeAttackEvent) GetEventType() EventType { return ETAttackEvent }

type ActionChoiceEvent struct {
	BaseEvent
	ActionChoice shared.ActionType
}

func (e *ActionChoiceEvent) GetEventType() EventType { return ETActionChoiceEvent }

//type SpellSlotsEvent struct {
//	BaseEvent
//	SpellSlots shared.SpellSlots
//}

//func (e *SpellSlotsEvent) GetEventType() EventType { return ETSpellSlotsEvent }

type SpellChoiceEvent struct {
	BaseEvent
	SpellChoice *spells.Spell
	CastLevel   int
	HasSlots    bool
}

func (e *SpellChoiceEvent) GetEventType() EventType { return ETSpellChoiceEvent }

type SpellAttackEvent struct {
	BaseEvent
	Target         string
	SpellChoice    *spells.Spell
	AttackRoll     int
	AttackModifier int
	AttackTotal    int
	Success        bool
}

func (e *SpellAttackEvent) GetEventType() EventType { return ETSpellAttackEvent }

type SpellDCEvent struct {
	BaseEvent
	Target      string
	SpellChoice *spells.Spell
	DC          int
	SavingThrow int
	Success     bool
}

func (e *SpellDCEvent) GetEventType() EventType { return ETSpellDCEvent }

type DamageEvent struct {
	BaseEvent
	Target     string
	DamageType string
	Amount     int
	Rolls      []int
}

func (e *DamageEvent) GetEventType() EventType { return ETDamageEvent }

type HealEvent struct {
	BaseEvent
	Target string
	Amount int
	Rolls  []int
}

func (e *HealEvent) GetEventType() EventType { return ETHealEvent }

type DeathEvent struct {
	BaseEvent
}

func (e *DeathEvent) GetEventType() EventType { return ETDeathEvent }

type UnconsciousEvent struct {
	BaseEvent
}

func (e *UnconsciousEvent) GetEventType() EventType { return ETUnconsciousEvent }

type HPModifiedEvent struct {
	BaseEvent
	Amount     int
	PreviousHP int
	CurrentHP  int
}

func (e *HPModifiedEvent) GetEventType() EventType { return ETHPModifiedEvent }

type DiceRollEvent struct {
	BaseEvent
	RollType shared.DiceRollType
	Value    int
	Rolls    []int
	Modifier int
}

func (e *DiceRollEvent) GetEventType() EventType { return ETRollEvent }

type HPRollEvent struct {
	BaseEvent
	Value    int
	Rolls    []int
	Modifier int
}

func (e *HPRollEvent) GetEventType() EventType { return ETHPRollEvent }

type SavingThrowEvent struct {
	BaseEvent
	Actor    string
	Result   int
	Roll     int
	Modifier int
	Success  bool
}

func (e *SavingThrowEvent) GetEventType() EventType { return ETSavingThrowEvent }

type TargetChoiceEvent struct {
	BaseEvent
	Target string
}

func (e *TargetChoiceEvent) GetEventType() EventType { return ETTargetChoiceEvent }

type CombatLogger interface {
	LogEvent(event CombatEvent)
}

type CombatListener interface {
	HandleEvent(event CombatEvent)
}
