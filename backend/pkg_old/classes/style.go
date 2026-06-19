package classes

type FightingStyle int

const (
	_ FightingStyle = iota
	StyleArchery
	StyleDefense
	StyleDueling
	StyleGWF
	StyleTWF
)

func (sid FightingStyle) String() string {
	switch sid {
	case StyleArchery:
		return "Archery"
	case StyleDefense:
		return "Defense"
	case StyleDueling:
		return "Dueling"
	case StyleGWF:
		return "Great Weapon Fighting"
	case StyleTWF:
		return "Two Weapon Fighting"
	default:
		return "unknown style"
	}
}
