package monster

import "dnd5e-encounter-simulator-backend/pkg_old/core"

type Action struct {
	ActionID      int                `json:"action_id"`
	Name          string             `json:"name"`
	RechargeValue int                `json:"recharge_value"`
	HasDC         bool               `json:"has_dc"`
	Index         int                `json:"index"`
	DamageBlocks  []core.DamageBlock `json:"damage_blocks"`
	AttackBonus   int                `json:"attack_bonus"`

	// Optional fields - use pointers
	DCAbility   *string `json:"dc_ability"`
	DCOnSuccess *string `json:"dc_on_success"`
	DC          *int    `json:"dc"`
}

type Multiattack struct {
	ActionID int `json:"action_id"`
	Count    int `json:"count"`
}

type MultiattackData struct {
	AttackDataBlocks []core.AttackData `json:"attack_data_blocks"`
	TotalAverage     int               `json:"total_average"`
	AveragePerAttack int               `json:"average_per_attack"`
}

type LegendaryAction struct {
	Cost   int    `json:"cost"`
	Action Action `json:"action"`
}

type MAMConfig struct {
	Actions          map[int]Action
	Multiattacks     map[int][]Multiattack
	LegendaryActions map[int]LegendaryAction
	SpecialAbilities SpecialAbilities
}
