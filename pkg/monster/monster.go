package monster

import "dnd5e-encounter-simulator-backend/pkg/shared"

type MonsterBase struct {
	ID                  int
	Name                string
	Size                string
	Type                string
	AC                  int
	ProficiencyBonus    int
	CR                  float64
	ApiURL              string
	IsLegendary         bool
	IsSpellcaster       bool
	IsInnateSpellcaster bool
	AbilityScores       shared.AbilityScores
	HP                  shared.MonsterHP
	SaveProficiencies   shared.SaveProficiencies
}

type MonsterDamageModifier struct {
	DamageType   string
	ModifierType string
}

type Monster struct {
	MonsterBase
	DamageModifiers []MonsterDamageModifier
	ResistBreakers  []shared.DamageBreaker
}

type MonsterQueryParams struct {
	Name string
	ID   int
}

// TODO: Add New function?
