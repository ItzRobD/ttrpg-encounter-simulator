package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg_old/character"
	"dnd5e-encounter-simulator-backend/pkg_old/lair"
	"dnd5e-encounter-simulator-backend/pkg_old/monster"
)

type SimulationStatus string

const (
	StatusNew      SimulationStatus = "new"
	StatusPending  SimulationStatus = "pending"
	StatusComplete SimulationStatus = "complete"
	StatusRunning  SimulationStatus = "running"
	StatusFailed   SimulationStatus = "failed"
)

type SimulationEntityConfigs struct {
	MonsterConfigs   []monster.MonsterConfig     `json:"monster_configs"`
	CharacterConfigs []character.CharacterConfig `json:"character_configs"`
	MonsterIDs       []int                       `json:"monster_ids"`
	LairConfig       lair.LairConfig             `json:"lair_config"`
}

func NewSimulationEntityConfigs() SimulationEntityConfigs {
	return SimulationEntityConfigs{
		MonsterConfigs:   make([]monster.MonsterConfig, 0),
		CharacterConfigs: make([]character.CharacterConfig, 0),
		MonsterIDs:       make([]int, 0),
		LairConfig:       lair.LairConfig{},
	}
}
