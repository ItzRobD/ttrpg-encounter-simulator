package weapon

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"fmt"
)

type WeaponSummary struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	DamageBlocks []core.DamageBlock `json:"damage_blocks"`
	Properties   Properties         `json:"properties"`
	Modifiers    Modifiers          `json:"modifiers"`
	IsCustom     bool               `json:"is_custom"`
}

// Weapon represents a weapon with properties such as name, damage type, and special attributes.
// Name refers to the name of the weapon.
// IsVersatile indicates if the weapon can be used with one or two hands.
// IsFinesse determines if the weapon can use Dexterity for attack and damage rolls.
// DamageBlocks specifies the dice and damage types rolled for weapon damage.
// IsRanged indicates if the weapon is a ranged weapon.
type Weapon struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	DamageBlocks []core.DamageBlock `json:"damage_blocks"`
	Properties   Properties         `json:"properties"`
	Modifiers    Modifiers          `json:"modifiers"`
	IsCustom     bool               `json:"is_custom"`
}

type Properties struct {
	IsVersatile  bool `json:"is_versatile"`
	IsFinesse    bool `json:"is_finesse"`
	IsRanged     bool `json:"is_ranged"`
	IsHeavy      bool `json:"is_heavy"`
	IsTwoHanded  bool `json:"is_two_handed"`
	IsLight      bool `json:"is_light"`
	IsThrown     bool `json:"is_thrown"`
	IsOnlyRanged bool `json:"is_only_ranged"`
}

type Modifiers struct {
	IsMagic          bool `json:"is_magic"`
	IsSilvered       bool `json:"is_silvered"`
	IsAdamantine     bool `json:"is_adamantine"`
	IsColdForgedIron bool `json:"is_cold_forged_iron"`
	AttackBonus      int  `json:"attack_bonus"`
	DamageBonus      int  `json:"damage_bonus"`
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
		Name: name,
		DamageBlocks: []core.DamageBlock{
			{
				NumberOfDice: numberOfDice,
				Die:          die,
				DamageType:   damageType,
			},
		},
		Properties: properties,
	}, nil
}

func (w *Weapon) SetModifiers(mods Modifiers) {
	w.Modifiers = mods
}

func (w *Weapon) GetModifiers() Modifiers {
	return w.Modifiers
}

func (w *Weapon) SetDamageBonus(v int) {
	w.Modifiers.DamageBonus = v
}

func (w *Weapon) SetAttackBonus(v int) {
	w.Modifiers.AttackBonus = v
}

func (w *Weapon) GetAttackBonus() int {
	return w.Modifiers.AttackBonus
}

func (w *Weapon) GetDamageBonus() int {
	return w.Modifiers.DamageBonus
}

func (w *Weapon) GetResistBreakers() []core.ResistBreaker {
	breakers := make([]core.ResistBreaker, 0)
	if w.Modifiers.IsMagic {
		breakers = append(breakers, core.ResistBreakerMagic)
	}
	if w.Modifiers.IsSilvered {
		breakers = append(breakers, core.ResistBreakerSilvered)
	}
	if w.Modifiers.IsAdamantine {
		breakers = append(breakers, core.ResistBreakerAdamantine)
	}
	if w.Modifiers.IsColdForgedIron {
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
