package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"time"
)

type EventType string

const (
	EventInitiative       EventType = "initiative"
	EventTurnStart        EventType = "turn_start"
	EventTurnEnd          EventType = "turn_end"
	EventRecharge         EventType = "recharge"
	EventDecisionStart    EventType = "decision_start"
	EventActionStart      EventType = "action_start"
	EventResolution       EventType = "resolution"
	EventAttackRoll       EventType = "attack_roll"
	EventSavingThrow      EventType = "saving_throw"
	EventOutcome          EventType = "outcome"
	EventDamageRoll       EventType = "damage_roll"
	EventHealRoll         EventType = "heal_roll"
	EventDamageModified   EventType = "damage_modified"
	EventHPModified       EventType = "hp_modified"
	EventConditionAdded   EventType = "condition_added"
	EventConditionRemoved EventType = "condition_removed"
	EventFeatureTrigger   EventType = "feature_trigger"
	EventDeath            EventType = "death"
	EventDeathSave        EventType = "death_save"
	EventUnconscious      EventType = "unconscious"
	EventVictory          EventType = "victory"
	EventCombatStart      EventType = "combat_start"
	EventLegendaryAction  EventType = "legendary_action"
	EventMessage          EventType = "message"
)

type ActorInfo struct {
	Name       string         `json:"name"`
	InstanceID int            `json:"instance_id"`
	Type       core.ActorType `json:"type"`
	Side       core.Side      `json:"side"`
}

type ActorSnapshot struct {
	CurrentHP   int              `json:"current_hp"`
	TempHP      int              `json:"temp_hp"`
	Conditions  []core.Condition `json:"conditions"`
	HealthState core.HealthState `json:"health_state"`
}

type TimelineEvent struct {
	Timestamp   time.Time             `json:"timestamp"`
	ID          string                `json:"id"`
	ParentID    string                `json:"parent_id"`
	Round       int                   `json:"round"`
	Type        EventType             `json:"type"`
	Actor       *ActorInfo            `json:"actor,omitempty"`
	Data        interface{}           `json:"data,omitempty"`
	ActorStates map[int]ActorSnapshot `json:"actor_states,omitempty"`
}
