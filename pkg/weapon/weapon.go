package weapon

type Weapon struct {
	Name         string
	IsVersatile  bool
	NumberOfDice int
	Die          int
	DamageType   string
}

type WeaponQueryParams struct {
	Name string
	ID   int
}

func New(name string, isVersatile bool, numberOfDice int, die int, damageType string) Weapon {
	return Weapon{
		Name:         name,
		IsVersatile:  isVersatile,
		NumberOfDice: numberOfDice,
		Die:          die,
		DamageType:   damageType,
	}
}
