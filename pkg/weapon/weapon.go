package weapon

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
)

type Weapon struct {
	Name         string
	IsVersatile  bool
	IsFinesse    bool
	NumberOfDice int
	Die          int
	DamageType   string
	IsRanged     bool
}

type WeaponQueryParams struct {
	Name string
	ID   int
}

func New(name string, isVersatile bool, isFinesse bool, numberOfDice int, die int, damageType string, isRanged bool) (Weapon, error) {
	if name == "" {
		name = "Unnamed weapon"
	}
	if numberOfDice < 1 {
		return Weapon{}, fmt.Errorf("number of rolling must be greater than 0")
	}
	if !core.ValidateDie(die) {
		return Weapon{}, fmt.Errorf("invalid damage die: %d", die)
	}
	if !core.ValidateDamageType(damageType) {
		return Weapon{}, fmt.Errorf("invalid damage type: %s", damageType)
	}
	return Weapon{
		Name:         name,
		IsVersatile:  isVersatile,
		IsFinesse:    isFinesse,
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

func (w *Weapon) GetAttackModifier(as *core.AbilityScores, clvl int, isProficient bool) (int, error) {
	mod, err := w.GetWeaponModifier(as)
	if err != nil {
		return 0, err
	}

	if !isProficient {
		return mod, nil
	}

	pb, err := core.GetCharacterProficiencyBonus(clvl)
	if err != nil {
		return 0, err
	}

	return mod + pb, nil
}

func (w *Weapon) GetWeaponModifier(as *core.AbilityScores) (int, error) {
	var mod int
	var err error
	if w.IsRanged || w.IsFinesse {
		mod, err = core.GetAbilityScoreModifier(as.Dexterity)
		if err != nil {
			return 0, err
		}
		return mod, nil
	} else {
		mod, err = core.GetAbilityScoreModifier(as.Strength)
		if err != nil {
			return 0, err
		}
		return mod, nil
	}
}
