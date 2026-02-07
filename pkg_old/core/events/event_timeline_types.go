package events

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"time"
)

type TimelineType string

const (
	TimelineChoiceType         TimelineType = "choice"
	TimelineAttackType         TimelineType = "attack"
	TimelineSaveType           TimelineType = "saving_throw"
	TimelineDamageType         TimelineType = "damage_roll"
	TimelineEffectType         TimelineType = "effect"
	TimelineHPModifiedType     TimelineType = "hp_modified"
	TimelineDamageModifiedType TimelineType = "damage_modified"
	TimelineHealType           TimelineType = "heal"
	TimelineDeathType          TimelineType = "death"
	TimelineUnconsciousType    TimelineType = "unconscious"
	TimelineConditionType      TimelineType = "condition"
	TimelineVictoryType        TimelineType = "victory"
	TimelineEquipmentType      TimelineType = "equipment"
	TimelineActionDetailType   TimelineType = "action_detail"
	TimelineTurnStartType      TimelineType = "turn_start"
	TimelineInitiativeType     TimelineType = "initiative"
	TimelineMultiattackType    TimelineType = "multiattack"
	TimelineActionType         TimelineType = "action"
	TimelineMessageType        TimelineType = "message"
)

type TimelineEntity struct {
	Name       string          `json:"name"`
	InstanceID int             `json:"instance_id"`
	Type       core.EntityType `json:"type"`
}

type TimelineEvent struct {
	Timestamp  time.Time    `json:"timestamp"`
	ID         string       `json:"id"`
	SequenceID string       `json:"sequence_id"`
	ParentID   string       `json:"parent_id"`
	Round      int          `json:"round"`
	Type       TimelineType `json:"type"`
	Data       interface{}  `json:"data"`
}

type TimelineChoice struct {
	Actor      TimelineEntity `json:"actor,omitempty"`
	Target     TimelineEntity `json:"target"`
	ChoiceType string         `json:"choice_type"`
	Choice     *string        `json:"choice,omitempty"`
	Scores     TimelineScores `json:"scores"`
}

type TimelineAttack struct {
	Actor        TimelineEntity `json:"actor,omitempty"`
	Target       TimelineEntity `json:"target"`
	AttackType   string         `json:"attack_type"`
	DiceRoll     interface{}    `json:"dice_roll"`
	ActionDetail *ActionDetail  `json:"action_detail,omitempty"`
}

type TimelineSavingThrow struct {
	Actor    TimelineEntity `json:"actor"`
	Target   TimelineEntity `json:"target"`
	DC       int            `json:"dc"`
	DiceRoll interface{}    `json:"dice_roll"`
}

type TimelineEquipment struct {
	Actor        TimelineEntity `json:"actor"`
	Name         string         `json:"name"`
	NumberOfDice int            `json:"number_of_dice"`
	Die          string         `json:"die"`
	DamageType   string         `json:"damage_type"`
	AttackBonus  int            `json:"attack_bonus"`
	DamageBonus  int            `json:"damage_bonus"`
	IsRanged     bool           `json:"is_ranged"`
	Properties   []string       `json:"properties"`
	Modifiers    []string       `json:"modifiers"`
}

type TimelineRoll struct {
	Actor      TimelineEntity `json:"actor"`
	Target     TimelineEntity `json:"target"`
	Roll       interface{}    `json:"roll"`
	DamageType string         `json:"damage_type"`
}

type TimelineEffect struct {
	Actor            TimelineEntity       `json:"actor"`
	Target           TimelineEntity       `json:"target"`
	Type             core.EffectType      `json:"type"`
	Value            int                  `json:"value"`
	DamageType       core.DamageType      `json:"damage_type"`
	ResistBreakers   []core.ResistBreaker `json:"resist_breakers"`
	Condition        *core.Condition      `json:"condition"`
	SourceRollID     string               `json:"source_roll_id"`
	Note             string               `json:"note"`
	OriginalHP       int                  `json:"original_hp"`
	FinalHP          int                  `json:"final_hp"`
	OriginalTempHP   int                  `json:"original_temp_hp"`
	FinalTempHP      int                  `json:"final_temp_hp"`
	OriginalValue    int                  `json:"original_value"`
	FinalValue       int                  `json:"final_value"`
	WasModified      bool                 `json:"was_modified"`
	ResistanceType   core.ResistanceType  `json:"resistance_type"`
	ResistanceBroken bool                 `json:"resistance_broken"`
}

type TimelineScores struct {
	UtilityScore float64            `json:"utility_score"`
	Factors      map[string]float64 `json:"factors"`
	TopFactors   []string           `json:"top_factors"`
}

type TimelineVictory struct {
	Winner WinningSide `json:"winner"`
	Rounds int         `json:"rounds"`
}

type TimelineMessage struct {
	Actor   TimelineEntity `json:"actor,omitempty"`
	Message string         `json:"message"`
}

type TimelineTurnStart struct {
	Actor TimelineEntity `json:"actor"`
}
