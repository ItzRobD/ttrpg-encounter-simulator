package core

import (
	"errors"
	"strings"
)

type Feature struct {
	ID          string            `json:"id"`
	Name        SpecialAbility    `json:"name"`
	Description string            `json:"description"`
	Hooks       map[HookType]bool `json:"hooks"`
	Data        FeatureData       `json:"data"`
}

type HookType string

const (
	HookNone HookType = "none"

	// Turn Lifecycle
	HookOnTurnStart   HookType = "on_turn_start"
	HookOnTurnEnd     HookType = "on_turn_end"
	HookOnCombatStart HookType = "on_combat_start"
	HookOnFirstTurn   HookType = "on_first_turn" // Specific to Assassinate/Berserk logic

	// Self-Centric (Defensive/Passive)
	HookOnSelfDamageTaken             HookType = "on_self_take_damage"
	HookOnSelfDeath                   HookType = "on_self_death"
	HookOnSelfSavingThrow             HookType = "on_self_saving_throw"
	HookOnSelfSavingThrowAgainstMagic HookType = "on_self_saving_throw_magic"
	HookOnSelfCriticalHit             HookType = "on_self_critical_hit"
	HookOnSelfDamageRoll              HookType = "on_self_damage_roll"
	HookOnSelfInitiativeRoll          HookType = "on_self_initiative_roll"
	HookOnSelfTakeRangedDamage        HookType = "on_self_take_ranged_damage"
	HookOnSelfTakeDamage              HookType = "on_self_take_damage"

	// Outgoing-Centric (Offensive)
	HookOnSelfAttack HookType = "on_self_attack"     // Modifies the attack roll (e.g. Pack Tactics)
	HookOnSelfHit    HookType = "on_self_attack_hit" // Triggers after a hit is confirmed (e.g. Sneak Attack)

	// Target-Centric
	HookOnTargetInjured HookType = "on_attack_target_injured" // Triggers based on target HP (e.g. Blood Frenzy)
	HookOnSelfTargeted  HookType = "on_self_targeted"
)

// MakeHookType converts a database string into a strongly-typed HookType
func MakeHookType(s string) HookType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "on_first_turn":
		return HookOnFirstTurn
	case "on_attack_target_injured":
		return HookOnTargetInjured
	case "on_self_take_damage":
		return HookOnSelfDamageTaken
	case "on_self_death":
		return HookOnSelfDeath
	case "on_self_attack_hit":
		return HookOnSelfHit
	case "on_self_saving_throw":
		return HookOnSelfSavingThrow
	case "on_self_saving_throw_against_magic":
		return HookOnSelfSavingThrowAgainstMagic
	case "on_self_attack":
		return HookOnSelfAttack
	case "on_turn_start":
		return HookOnTurnStart
	case "on_combat_start":
		return HookOnCombatStart
	case "on_self_damage_roll":
		return HookOnSelfDamageRoll
	case "on_self_critical_hit":
		return HookOnSelfCriticalHit
	case "on_turn_end":
		return HookOnTurnEnd
	case "on_self_initiative_roll":
		return HookOnSelfInitiativeRoll
	case "on_self_take_ranged_damage":
		return HookOnSelfTakeRangedDamage
	case "on_self_targeted":
		return HookOnSelfTargeted
	default:
		return HookNone
	}
}

func (ht HookType) String() string {
	return string(ht)
}

type ScalerType string

const (
	ScalerDamage    ScalerType = "damage"
	ScalerSpellSlot ScalerType = "spell_slot"
	ScalerLevel     ScalerType = "level"
)

type FeatureData struct {
	Value            int           `json:"value,omitempty"`
	NumberOfDice     int           `json:"number_of_dice,omitempty"`
	Die              DiceType      `json:"die,omitempty"`
	Modifier         int           `json:"modifier,omitempty"`
	DamageType       []DamageType  `json:"damage_type,omitempty"`
	Scaler           int           `json:"scaler,omitempty"`
	ScalerType       ScalerType    `json:"scaler_type,omitempty"`
	DC               int           `json:"dc,omitempty"`
	Ability          Ability       `json:"ability,omitempty"`
	DCOnSuccess      DCOnSuccess   `json:"dc_on_success,omitempty"`
	BonusTargetTypes []MonsterType `json:"bonus_target_types,omitempty"`
	RerollType       string        `json:"reroll_type,omitempty"`
	RerollThreshold  int           `json:"reroll_threshold,omitempty"`
}

func NewFeature(id string, name SpecialAbility, desc string) Feature {
	return Feature{
		ID:          id,
		Name:        name,
		Description: desc,
		Hooks:       make(map[HookType]bool),
	}
}

func NewFeatureFromSpecialAbility(id string, name SpecialAbility, desc string, data FeatureData) Feature {
	f := NewFeature(id, name, desc)
	f.Data = data
	return f
}

func (f *Feature) SetData(data FeatureData) {
	f.Data = data
}

func (f *Feature) SetHooksFromStrings(hooks []string) {
	for _, hook := range hooks {
		f.Hooks[MakeHookType(hook)] = true
	}
}

func (f *Feature) ValidateData() error {
	if f.Data.Value < 0 {
		return errors.New("feature data value must be non-negative")
	}
	if f.Data.NumberOfDice < 0 {
		return errors.New("feature data number of dice must be non-negative")
	}
	if f.Data.DC < 0 {
		return errors.New("feature data DC must be non-negative")
	}
	return nil
}
