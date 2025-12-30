package weapon

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
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
	NumberOfDice int
	Die          core.DiceType
	DamageType   core.DamageType
	Properties   Properties
	modifiers    Modifiers
}

type Properties struct {
	IsVersatile  bool
	IsFinesse    bool
	IsRanged     bool
	IsHeavy      bool
	IsTwoHanded  bool
	IsLight      bool
	IsThrown     bool
	IsOnlyRanged bool
}

type Modifiers struct {
	IsMagic          bool
	IsSilvered       bool
	IsAdamantine     bool
	IsColdForgedIron bool
	AttackBonus      int
	DamageBonus      int
}

// WeaponQueryParams defines the parameters for querying weapon data, including weapon name and ID.
type WeaponQueryParams struct {
	Name string
	ID   int
}

// New creates a new weapon with specified attributes, validating inputs and returning an error for invalid configurations.
func New(name string, isVersatile bool, isFinesse bool, numberOfDice int, die core.DiceType,
	damageType core.DamageType, properties Properties) (Weapon, error) {
	if name == "" {
		name = "Unnamed weapon"
	}
	if numberOfDice < 1 {
		return Weapon{}, fmt.Errorf("number of rolling must be greater than 0")
	}

	return Weapon{
		Name:         name,
		NumberOfDice: numberOfDice,
		Die:          die,
		DamageType:   damageType,
		Properties:   properties,
	}, nil
}

func (w *Weapon) SetModifiers(mods Modifiers) {
	w.modifiers = mods
}

func (w *Weapon) GetModifiers() Modifiers {
	return w.modifiers
}

func (w *Weapon) SetDamageBonus(v int) {
	w.modifiers.DamageBonus = v
}

func (w *Weapon) SetAttackBonus(v int) {
	w.modifiers.AttackBonus = v
}

func (w *Weapon) GetAttackBonus() int {
	return w.modifiers.AttackBonus
}

func (w *Weapon) GetDamageBonus() int {
	return w.modifiers.DamageBonus
}

func (w *Weapon) GetResistBreakers() []core.ResistBreaker {
	breakers := make([]core.ResistBreaker, 0)
	if w.modifiers.IsMagic {
		breakers = append(breakers, core.ResistBreakerMagic)
	}
	if w.modifiers.IsSilvered {
		breakers = append(breakers, core.ResistBreakerSilvered)
	}
	if w.modifiers.IsAdamantine {
		breakers = append(breakers, core.ResistBreakerAdamantine)
	}
	if w.modifiers.IsColdForgedIron {
		breakers = append(breakers, core.ResistBreakerColdForgedIron)
	}
	return breakers
}

// GetAttackModifier calculates the attack modifier for a weapon based on ability scores, character level, and proficiency status.
// Returns the attack modifier or an error if any calculation fails.
func (w *Weapon) GetAttackModifier(as *core.AbilityScores, clvl uint8, isProficient bool) (int, error) {
	mod, _, err := w.GetWeaponModifier(as)
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
func (w *Weapon) GetWeaponModifier(as *core.AbilityScores) (int, core.Ability, error) {
	var mod int
	var err error
	ability := core.AbilityNone
	if w.Properties.IsRanged {
		mod, err = core.GetAbilityScoreModifier(as.Dexterity)
		if err != nil {
			return 0, ability, err
		}
		return mod, core.AbilityDexterity, nil
	} else if w.Properties.IsFinesse {
		mod, err = core.GetAbilityScoreModifier(as.Strength)
		if err != nil {
			return 0, ability, err
		}
		dexMod, dexErr := core.GetAbilityScoreModifier(as.Dexterity)
		if dexErr != nil {
			return 0, ability, dexErr
		}
		if dexMod > mod {
			mod = dexMod
			ability = core.AbilityDexterity
		} else {
			ability = core.AbilityStrength
		}
		return mod, ability, nil
	}

	mod, err = core.GetAbilityScoreModifier(as.Strength)
	if err != nil {
		return 0, ability, err
	}
	return mod, ability, nil
}
