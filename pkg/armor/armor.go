package armor

import "fmt"

type Armor struct {
	Name       string
	ArmorClass int
	DexBonus   bool
	MaxBonus   bool
	MinimumStr int
}

type ArmorQueryParams struct {
	Name string
	ID   int
}

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
