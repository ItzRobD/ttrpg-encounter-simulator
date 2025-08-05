package monster

import "dnd5e-encounter-simulator-backend/pkg/core"

type Action struct {
	ActionID      int             // Action
	Name          string          // Action
	RechargeValue int             // Action
	HasDC         bool            // Used to determine if embedded struct is of value
	Index         int             // Action
	NumberOfDice  int             // DC || Attack Bonus Blocks
	Die           core.DiceType   // DC || Attack Bonus Blocks
	AmountToAdd   int             // DC || Attack Bonus Blocks
	AttackBonus   int             // DC || Attack Bonus Blocks
	DamageType    core.DamageType // DC || Attack Bonus Blocks

	// Optional fields - use pointers
	DCAbility   *string // nil if no DC
	DCOnSuccess *string // nil if no DC
	DC          *int    // nil if no DC
}

type Multiattack struct {
	ActionID int
	Count    int
}

type MultiattackData struct {
	AttackDataBlocks []core.AttackData
	TotalAverage     int
	AveragePerAttack int
}

type LegendaryAction struct {
	Cost   int
	Action Action
}

type SpecialAbility struct {
	Name        string
	UsageCount  int
	Description string
}

type MAMConfig struct {
	Actions          map[int]Action
	Multiattacks     map[int][]Multiattack
	LegendaryActions map[int]LegendaryAction
	SpecialAbilities []SpecialAbility
}
