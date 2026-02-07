package entity_state_manager

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
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
		CurrentHP: esm.currentHP,
		MaxHP:     esm.maxHP,
		HPPct:     int(math.Floor(float64(esm.currentHP * 100 / esm.maxHP))),
		TempHP:    esm.tempHP,
		HitDie:    esm.hitDie,
	}
}

func (hp HPValues) GetHP() int               { return hp.CurrentHP }
func (hp HPValues) GetMaxHP() int            { return hp.MaxHP }
func (hp HPValues) GetTempHP() int           { return hp.TempHP }
func (hp HPValues) GetHPPct() int            { return hp.HPPct }
func (hp HPValues) GetHitDie() core.DiceType { return hp.HitDie }
func (hp HPValues) GetHPDifference() int     { return hp.MaxHP - hp.CurrentHP }
