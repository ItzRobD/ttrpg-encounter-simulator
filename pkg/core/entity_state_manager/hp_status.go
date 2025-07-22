package entity_state_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"math"
)

type HPStatus struct {
	CurrentHP int
	MaxHP     int
	HPPct     int
	TempHP    int
	HitDie    core.DiceType
}

func (esm *EntityStateManager) GetHPStatus() HPStatus {
	return HPStatus{
		CurrentHP: esm.CurrentHP,
		MaxHP:     esm.MaxHP,
		HPPct:     int(math.Floor(float64(esm.CurrentHP * 100 / esm.MaxHP))),
		TempHP:    esm.TempHP,
		HitDie:    esm.HitDie,
	}
}

func (hps HPStatus) GetHP() int               { return hps.CurrentHP }
func (hps HPStatus) GetMaxHP() int            { return hps.MaxHP }
func (hps HPStatus) GetTempHP() int           { return hps.TempHP }
func (hps HPStatus) GetHPPct() int            { return hps.HPPct }
func (hps HPStatus) GetHitDie() core.DiceType { return hps.HitDie }
