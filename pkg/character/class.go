package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
)

// Class represents a character class with its attributes like ID, name, hit die, and spellcasting modifier.
type Class struct {
	ID              int
	Name            string
	HitDie          int
	SpellcastingMod core.Ability
}

// ClassQueryParams defines parameters for querying a class, including its name and ID.
type ClassQueryParams struct {
	Name string
	ID   int
}
