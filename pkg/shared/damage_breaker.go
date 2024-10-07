package shared

type DamageBreaker struct {
	DamageType string
	Breaker    WeaponBreakerType
}

type WeaponBreakerType string

const (
	MAGIC      WeaponBreakerType = "magic"
	SILVERED   WeaponBreakerType = "silvered"
	ADAMANTINE WeaponBreakerType = "adamantine"
	COLDFORGED WeaponBreakerType = "cold forged steel"
)

func (wbt WeaponBreakerType) String() string {
	return string(wbt)
}

func (wbt WeaponBreakerType) IsValid() bool {
	switch wbt {
	case MAGIC, SILVERED, ADAMANTINE, COLDFORGED:
		return true
	}
	return false
}
