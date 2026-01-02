package core

import "fmt"

type DiceRollType string

const (
	DiceRollGeneral          DiceRollType = "general"
	DiceRollAttack           DiceRollType = "attack"
	DiceRollDamage           DiceRollType = "damage"
	DiceRollHealing          DiceRollType = "healing"
	DiceRollInitiative       DiceRollType = "initiative"
	DiceRollAbilityCheck     DiceRollType = "ability check"
	DiceRollSavingThrow      DiceRollType = "saving throw"
	DiceRollHP               DiceRollType = "hp"
	DiceRollHPAvgUsed        DiceRollType = "hp average"
	DiceRollHPAvgUsedMonster DiceRollType = "hp average"
	DiceRollHPValueUsed      DiceRollType = "hp value"
	DiceRollRecharge         DiceRollType = "recharge"
	DiceRollDeathSavingThrow DiceRollType = "death saving throw"
)

type DiceType int

const (
	D0   DiceType = 0
	D4   DiceType = 4
	D6   DiceType = 6
	D8   DiceType = 8
	D10  DiceType = 10
	D12  DiceType = 12
	D20  DiceType = 20
	D100 DiceType = 100
)

func (dt DiceType) String() string {
	return fmt.Sprintf("%d", int(dt))
}

func (dt DiceType) Int() int {
	return int(dt)
}

func (dt DiceType) Max() int {
	return int(dt)
}

func (dt DiceType) Min() int {
	return 1
}

func (dt DiceType) Avg() float64 {
	return (float64(dt) / 2) + 0.5
}

func (dt DiceType) IsValid() bool {
	switch dt {
	case D4, D6, D8, D10, D12, D20, D100:
		return true
	}
	return false
}

func MakeDiceType(v int) (DiceType, error) {
	switch v {
	case 4, 6, 8, 10, 12, 20, 100:
		return DiceType(v), nil
	}
	return DiceType(0), fmt.Errorf("invalid dice type")
}

type AdvantageType int

const (
	RollNormal       AdvantageType = 0
	RollAdvantage    AdvantageType = 1
	RollDisadvantage AdvantageType = -1
)

func (at AdvantageType) String() string {
	switch at {
	case RollNormal:
		return "Normal"
	case RollAdvantage:
		return "Advantage"
	case RollDisadvantage:
		return "Disadvantage"
	default:
		return "invalid"
	}
}
