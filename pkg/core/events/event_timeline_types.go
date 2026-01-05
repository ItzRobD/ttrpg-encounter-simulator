package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"time"
)

type TimelineType string

const (
	TimelineChoiceType TimelineType = "choice"
	TimelineAttackType TimelineType = "attack"
	TimelineSaveType   TimelineType = "savingthrow"
	TimelineDamageType TimelineType = "damageroll"
	TimelineEffectType TimelineType = "effect"
)

type TimelineEntity struct {
	Name       string
	InstanceID int
	Type       core.EntityType
}

type TimelineEvent struct {
	Timestamp  time.Time
	ID         string
	SequenceID string
	ParentID   string
	Type       TimelineType
	Data       interface{}
}

type TimelineChoice struct {
	Actor      TimelineEntity
	Target     TimelineEntity
	ChoiceType string
	Choice     *string // Spell/Weapon/if target choice then nil
	Scores     TimelineScores
}

type TimelineAttack struct {
	Actor    TimelineEntity
	Target   TimelineEntity
	DiceRoll core.RollResult
}

type TimelineSavingThrow struct {
	Actor    TimelineEntity
	DC       int
	DiceRoll core.RollResult
}

type TimelineDamage struct {
	Actor  TimelineEntity
	Target TimelineEntity
	Damage core.RollResult
}

type TimelineEffect struct {
	Actor  TimelineEntity
	Target TimelineEntity
	Type   core.EffectType
	Value  int
	// Details for Damage/Heal
	DamageType     core.DamageType
	ResistBreakers []core.ResistBreaker
	// Details for Conditions
	Condition *core.Condition
	// ctx about the specific roll or save that caused this effect
	SourceRollID string
	// Metadata for the UI (e.g., "Half damage (Save Success)", "Resistant")
	Note string
}

type TimelineScores struct {
	UtilityScore float64
	Factors      map[DecisionFactor]float64
	TopReasons   []DecisionFactor
}
