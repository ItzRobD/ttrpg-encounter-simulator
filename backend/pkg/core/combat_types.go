package core

import "fmt"

// WeaponModifiers represents various traits and bonuses that can modify a weapon's properties and effectiveness.
type WeaponModifiers struct {
	IsMagic          bool `json:"is_magic"`
	IsSilvered       bool `json:"is_silvered"`
	IsAdamantine     bool `json:"is_adamantine"`
	IsColdForgedIron bool `json:"is_cold_forged_iron"`
	AttackBonus      int  `json:"attack_bonus"`
	DamageBonus      int  `json:"damage_bonus"`
}

// WeaponProperties represents a set of boolean attributes that define specific characteristics of a weapon.
type WeaponProperties struct {
	IsVersatile  bool `json:"is_versatile"`
	IsFinesse    bool `json:"is_finesse"`
	IsRanged     bool `json:"is_ranged"`
	IsHeavy      bool `json:"is_heavy"`
	IsTwoHanded  bool `json:"is_two_handed"`
	IsLight      bool `json:"is_light"`
	IsThrown     bool `json:"is_thrown"`
	IsOnlyRanged bool `json:"is_only_ranged"`
}

// DiceBlock represents a configuration for rolling dice in the context of actions, attacks, heals, or damage calculations.
type DiceBlock struct {
	NumberOfDice int        `json:"number_of_dice"` // NumberOfDice represents the number of dice used
	Die          DiceType   `json:"die"`            // DiceType represents the type of dice used
	DamageType   DamageType `json:"damage_type"`    // DamageType represents the type of damage dealt
	Modifier     int        `json:"modifier"`       // Flat bonus for this specific block
}

func (db DiceBlock) String() string {
	return fmt.Sprintf("%dd%d %s damage", db.NumberOfDice, db.Die, db.DamageType)
}

func MakeDiceBlock(numberOfDice int, die DiceType, damageType interface{}, modifier int) DiceBlock {
	var dt DamageType
	switch damageType.(type) {
	case string:
		dt, _ = MakeDamageType(damageType.(string))
	case DamageType:
		dt = damageType.(DamageType)
	default:
		dt = DamageNone
	}
	return DiceBlock{
		NumberOfDice: numberOfDice,
		Die:          die,
		DamageType:   dt,
		Modifier:     modifier,
	}
}

type HealthState string

const (
	HealthStateHealthy     HealthState = "healthy"     // 100%
	HealthStateWounded     HealthState = "wounded"     // > 75% < 100%
	HealthStateBloody      HealthState = "bloody"      // > 15% < 75%
	HealthStateCritical    HealthState = "critical"    // > 0% < 15%
	HealthStateDead        HealthState = "dead"        // 0
	HealthStateUnconscious HealthState = "unconscious" // 0 HP (for PCs)
)
