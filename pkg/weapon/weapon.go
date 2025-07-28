package weapon

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
)

var (
	rangedWeaponIDs = map[int]bool{
		2: true, 4: true, 5: true, 6: true, 10: true,
		12: true, 14: true, 29: true, 33: true, 34: true, 35: true,
	}

	dedicatedRangedIDs = map[int]bool{
		11: true, 12: true, 13: true, 14: true,
		33: true, 34: true, 35: true,
	}
)

// Weapon represents a weapon with properties such as name, damage type, and special attributes.
// Name refers to the name of the weapon.
// IsVersatile indicates if the weapon can be used with one or two hands.
// IsFinesse determines if the weapon can use Dexterity for attack and damage rolls.
// NumberOfDice specifies the count of dice rolled for damage calculation.
// Die represents the type of die rolled for weapon damage (e.g., d6, d8).
// DamageType specifies the type of damage the weapon inflicts (e.g., slashing, piercing).
// IsRanged indicates if the weapon is a ranged weapon.
type Weapon struct {
	Name         string
	IsVersatile  bool
	IsFinesse    bool
	NumberOfDice int
	Die          core.DiceType
	DamageType   core.DamageType
	IsRanged     bool
	IsMelee      bool
}

// WeaponQueryParams defines the parameters for querying weapon data, including weapon name and ID.
type WeaponQueryParams struct {
	Name string
	ID   int
}

// New creates a new weapon with specified attributes, validating inputs and returning an error for invalid configurations.
func New(name string, isVersatile bool, isFinesse bool, numberOfDice int, die core.DiceType, damageType core.DamageType, isRanged bool, isMelee bool) (Weapon, error) {
	if name == "" {
		name = "Unnamed weapon"
	}
	if numberOfDice < 1 {
		return Weapon{}, fmt.Errorf("number of rolling must be greater than 0")
	}
	return Weapon{
		Name:         name,
		IsVersatile:  isVersatile,
		IsFinesse:    isFinesse,
		NumberOfDice: numberOfDice,
		Die:          die,
		DamageType:   damageType,
		IsRanged:     isRanged,
		IsMelee:      isMelee,
	}, nil
}

// isRangedWeapon determines if the given weapon ID corresponds to a ranged weapon and returns true if it does.
func isRangedWeapon(id int) bool {
	return rangedWeaponIDs[id]

}

// isMeleeWeapon checks if the given weapon ID corresponds to a melee weapon by excluding certain ranged weapon IDs.
func isMeleeWeapon(id int) bool {
	return !dedicatedRangedIDs[id]
}

// GetAttackModifier calculates the attack modifier for a weapon based on ability scores, character level, and proficiency status.
// Returns the attack modifier or an error if any calculation fails.
func (w *Weapon) GetAttackModifier(as *core.AbilityScores, clvl uint8, isProficient bool) (int, error) {
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

// GetWeaponModifier determines the modifier to use for a weapon based on its type and the ability scores provided.
// Returns the calculated modifier or an error if the ability score calculation fails.
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
