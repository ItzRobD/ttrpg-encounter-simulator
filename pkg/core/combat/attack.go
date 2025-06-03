package combat

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/shared"
)

type AttackInfo struct {
	Name              string
	NumberOfDice      int
	Die               int
	AttackModifier    int
	DamageModifier    int
	DamageType        string
	IsVersatileAttack bool
}

func DoesAttackHit(attackTotal int, ac int) bool {
	if attackTotal > ac {
		return true
	}
	return false
}

func isCriticalHit(attackRoll int) bool {
	if attackRoll == 20 {
		return true
	}
	return false
}

func CalculateDamage(ai AttackInfo, isCritical bool, options core.Options) (int, []int, error) {
	if isCritical {
		return shared.CalculateDamageCriticalHit(ai.NumberOfDice, ai.Die, ai.DamageModifier, options.UseImprovedCriticals)
	}
	return shared.DiceRollWithModifier(ai.NumberOfDice, ai.Die, ai.DamageModifier)
}

// MakeMartialAttack rolls an attack, determines if it hits the target, and calculates damage if applicable.
// Returns a boolean indicating a hit, an integer for damage dealt, and an error if any issues occurred.
func MakeMartialAttack(attacker core.Entity, target core.Entity, attackInfo AttackInfo, advantage shared.AdvantageType, options core.Options) (bool, int, error) {
	// Make the attack roll
	attackTotal, attackRoll, err := shared.AttackRoll(attackInfo.AttackModifier, advantage)
	if err != nil {
		return false, 0, err
	}
	events.LogDiceRollEvent(attacker, attackTotal, []int{attackRoll}, shared.DiceRollAttack, attackInfo.AttackModifier, attacker.GetEventListener())

	// Check if the attack hits
	isCrit := isCriticalHit(attackRoll)
	if isCrit || DoesAttackHit(attackTotal, target.GetAC()) {
		events.LogMeleeAttackEvent(attacker, target, attackInfo.Name, attackRoll, attackInfo.AttackModifier, true, isCrit, attacker.GetEventListener())

		damage, rolls, err2 := CalculateDamage(attackInfo, isCrit, options)
		if err2 != nil {
			return false, 0, err2
		}
		events.LogDiceRollEvent(attacker, damage, rolls, shared.DiceRollDamage, attackInfo.DamageModifier, attacker.GetEventListener())
		return true, damage, nil
	}

	events.LogMeleeAttackEvent(attacker, target, attackInfo.Name, attackRoll, attackInfo.AttackModifier, false, isCrit, attacker.GetEventListener())
	return false, 0, nil
}
