package core

import (
	"fmt"
	"strings"
)

type DamageType string

const (
	DamageAcid        DamageType = "acid"
	DamageCold        DamageType = "cold"
	DamageFire        DamageType = "fire"
	DamageForce       DamageType = "force"
	DamageLightning   DamageType = "lightning"
	DamageNecrotic    DamageType = "necrotic"
	DamagePoison      DamageType = "poison"
	DamagePsychic     DamageType = "psychic"
	DamageRadiant     DamageType = "radiant"
	DamageThunder     DamageType = "thunder"
	DamageSlashing    DamageType = "slashing"
	DamageBludgeoning DamageType = "bludgeoning"
	DamagePiercing    DamageType = "piercing"
	DamageNone        DamageType = "none"
)

func (dt DamageType) String() string {
	return string(dt)
}

func MakeDamageType(s string) (DamageType, error) {
	switch strings.ToLower(s) {
	case "acid":
		return DamageAcid, nil
	case "cold":
		return DamageCold, nil
	case "fire":
		return DamageFire, nil
	case "force":
		return DamageForce, nil
	case "lightning":
		return DamageLightning, nil
	case "necrotic":
		return DamageNecrotic, nil
	case "poison":
		return DamagePoison, nil
	case "psychic":
		return DamagePsychic, nil
	case "radiant":
		return DamageRadiant, nil
	case "thunder":
		return DamageThunder, nil
	case "slashing":
		return DamageSlashing, nil
	case "bludgeoning":
		return DamageBludgeoning, nil
	case "piercing":
		return DamagePiercing, nil
	default:
		return DamageAcid, fmt.Errorf("invalid damage type")
	}
}
