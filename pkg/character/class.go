package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
)

type Class struct {
	ID              int
	Name            string
	HitDie          int
	SpellcastingMod core.Ability
}

type ClassQueryParams struct {
	Name string
	ID   int
}
