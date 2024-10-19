package weapon

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
)

type Weapon struct {
	Name         string
	IsVersatile  bool
	NumberOfDice int
	Die          int
	DamageType   string
	IsRanged     bool
}

type WeaponQueryParams struct {
	Name string
	ID   int
}

func New(name string, isVersatile bool, numberOfDice int, die int, damageType string, isRanged bool) (Weapon, error) {
	if name == "" {
		name = "Unnamed weapon"
	}
	if numberOfDice < 1 {
		return Weapon{}, fmt.Errorf("number of dice must be greater than 0")
	}
	if !shared.ValidateDie(die) {
		return Weapon{}, fmt.Errorf("invalid damage die: %d", die)
	}
	if !shared.ValidateDamageType(damageType) {
		return Weapon{}, fmt.Errorf("invalid damage type: %s", damageType)
	}
	return Weapon{
		Name:         name,
		IsVersatile:  isVersatile,
		NumberOfDice: numberOfDice,
		Die:          die,
		DamageType:   damageType,
		IsRanged:     isRanged,
	}, nil
}

func isRangedWeapon(id int) bool {
	rangedIDs := []int{2, 4, 5, 6, 10, 12, 14, 29, 33, 34, 35}
	for _, v := range rangedIDs {
		if v == id {
			return true
		}
	}
	return false
}
