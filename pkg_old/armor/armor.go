package armor

import "fmt"

// Armor represents a set of protective equipment with specific attributes like name, defense capability, and requirements.
// Name specifies the name of the armor.
// ArmorClass indicates base AC of the armor.
// DexBonus determines if the armor allows the addition of Dexterity bonuses to effective ArmorClass.
// MaxBonus specifies whether the dex bonus is limited to 2 maximum
// MinimumStr defines the minimum Strength required to equip the armor.
type Armor struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ArmorClass int    `json:"ac"`
	DexBonus   bool   `json:"dex_bonus"`
	MaxBonus   bool   `json:"max_bonus"`
	MinimumStr int    `json:"minimum_str"`
	Modifier   int    `json:"modifier"`
	IsCustom   bool   `json:"is_custom"`
}

// ArmorQueryParams is used to specify query parameters for fetching armor data.
// Name represents the name of the armor to query.
// ID specifies the unique identifier of the armor to query.
type ArmorQueryParams struct {
	Name string
	ID   int
}

// New creates a new Armor instance with the specified attributes and validates the input parameters.
// Returns an Armor object or an error if validation fails.
func New(name string, armorClass int, dexBonus bool, maxBonus bool, minimumStr int) (Armor, error) {
	if name == "" {
		name = "Unnamed Armor"
	}
	if armorClass < 0 {
		return Armor{}, fmt.Errorf("armor class must be greater than or equal to 0")
	}
	if minimumStr < 0 {
		return Armor{}, fmt.Errorf("minimum strength requirement must be greater than or equal to 0")
	}
	return Armor{
		Name:       name,
		ArmorClass: armorClass,
		DexBonus:   dexBonus,
		MaxBonus:   maxBonus,
		MinimumStr: minimumStr,
	}, nil
}
