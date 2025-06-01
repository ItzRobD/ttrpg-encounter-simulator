package core

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
)

type Entity interface {
	ModifyHP(amount int)
	IsUnconscious() bool
	GetCurrentHP() int
	GetCurrentHPPct() int
	GetMaxHP() int
	GetName() string
	GetAC() int
	GetEventListener() func(event interface{})
	GetSavingThrowRollResult(ability string) (int, error)
}

type Combatant struct {
	InitiativeScore int
	Entity          Entity
}

type EntityModifiers struct {
	InitiativeAdvantage shared.AdvantageType
	InitiativeBonus     int
}

type CombatState struct {
	// HP Management
	CurrentHP int
	MaxHP     int
	TempHP    int

	// TODO: Evaluate if this is necessary
	// Action Economy
	HasUsedAction         bool
	HasUsedBonusAction    bool
	HasUsedReaction       bool
	LegendaryActionPoints int

	// Spell Resources
	SpellSlots    shared.SpellSlots
	Concentration *spells.Spell

	// TODO: Evaluate if conditions will be a feature
	// Conditions
	Conditions []string
}
