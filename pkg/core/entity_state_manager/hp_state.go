package entity_state_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"math"
)

type HPValues struct {
	CurrentHP int
	MaxHP     int
	HPPct     int
	TempHP    int
	HitDie    core.DiceType
}

func (esm *EntityStateManager) GetHPStatus() HPValues {
	return HPValues{
		CurrentHP: esm.CurrentHP,
		MaxHP:     esm.MaxHP,
		HPPct:     int(math.Floor(float64(esm.CurrentHP * 100 / esm.MaxHP))),
		TempHP:    esm.TempHP,
		HitDie:    esm.HitDie,
	}
}

func (hp HPValues) GetHP() int               { return hp.CurrentHP }
func (hp HPValues) GetMaxHP() int            { return hp.MaxHP }
func (hp HPValues) GetTempHP() int           { return hp.TempHP }
func (hp HPValues) GetHPPct() int            { return hp.HPPct }
func (hp HPValues) GetHitDie() core.DiceType { return hp.HitDie }
