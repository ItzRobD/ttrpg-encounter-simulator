package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"time"
)

type TimelineType string

const (
	TimelineChoiceType         TimelineType = "choice"
	TimelineAttackType         TimelineType = "attack"
	TimelineSaveType           TimelineType = "savingthrow"
	TimelineDamageType         TimelineType = "damageroll"
	TimelineEffectType         TimelineType = "effect"
	TimelineHPModifiedType     TimelineType = "hpmodified"
	TimelineDamageModifiedType TimelineType = "damagemodified"
	TimelineHealType           TimelineType = "heal"
	TimelineDeathType          TimelineType = "death"
	TimelineUnconsciousType    TimelineType = "unconscious"
	TimelineConditionType      TimelineType = "condition"
	TimelineVictoryType        TimelineType = "victory"
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
	Round      int
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
	Actor      TimelineEntity
	Target     TimelineEntity
	AttackType string // "melee" or "spell"
	DiceRoll   interface{}
}

type TimelineSavingThrow struct {
	Actor    TimelineEntity
	Target   TimelineEntity
	DC       int
	DiceRoll interface{}
}

type TimelineRoll struct {
	Actor  TimelineEntity
	Target TimelineEntity
	Roll   interface{}
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

	// Added fields for better UI representation
	OriginalHP       int
	FinalHP          int
	OriginalTempHP   int
	FinalTempHP      int
	OriginalValue    int
	FinalValue       int
	WasModified      bool
	ResistanceType   core.ResistanceType
	ResistanceBroken bool
}

type TimelineScores struct {
	UtilityScore float64
	Factors      map[DecisionFactor]float64
	TopReasons   []DecisionFactor
}

type TimelineVictory struct {
	Winner WinningSide
	Rounds int
}
