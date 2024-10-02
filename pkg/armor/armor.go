package armor

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

func New(name string, armorClass int, dexBonus, maxBonus bool, minimumStr int) Armor {
	return Armor{
		Name:       name,
		ArmorClass: armorClass,
		DexBonus:   dexBonus,
		MaxBonus:   maxBonus,
		MinimumStr: minimumStr,
	}
}
