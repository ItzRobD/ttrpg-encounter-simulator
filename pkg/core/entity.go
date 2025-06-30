package core

import (
	"dnd5e-encounter-simulator-backend/pkg/spells"
)

type Entity interface {
	ModifyHP(value int)
	IsUnconscious() bool
	GetCurrentHP() int
	GetCurrentHPPct() int
	GetMaxHP() int
	GetName() string
	GetAC() int
	GetEventListener() func(event interface{})
	GetSavingThrowRollResult(ability string) (int, error)
	GetLevel() interface{}
	GetCasterLevel() int
}

type Combatant struct {
	InitiativeScore int
	Entity          Entity
}

type EntityModifiers struct {
	InitiativeAdvantage AdvantageType
	InitiativeBonus     int
	UseVersatileAttacks bool
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
	// TODO: Spellcasting manager
	Concentration *spells.Spell

	// TODO: Evaluate if conditions will be a feature
	// Conditions
	Conditions []string
}
