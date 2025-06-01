package combat

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/shared"
)

type AttackInfo struct {
	Name           string
	NumberOfDice   int
	Die            int
	AttackModifier int
	DamageModifier int
	DamageType     string
}

func DoesAttackHit(attackTotal int, ac int) bool {
	if attackTotal > ac {
		return true
	}
	return false
}

// MakeMartialAttack rolls an attack, determines if it hits the target, and calculates damage if applicable.
// Returns a boolean indicating a hit, an integer for damage dealt, and an error if any issues occurred.
func MakeMartialAttack(attacker core.Entity, target core.Entity, attackInfo AttackInfo, advantage shared.AdvantageType) (bool, int, error) {
	// Make the attack roll
	attackTotal, attackRoll, err := shared.AttackRoll(attackInfo.AttackModifier, advantage)
	if err != nil {
		return false, 0, err
	}
	events.LogDiceRollEvent(attacker, attackTotal, []int{attackRoll}, shared.DiceRollAttack, attackInfo.AttackModifier, attacker.GetEventListener())

	// Check if the attack hits
	didHit := DoesAttackHit(attackTotal, target.GetAC())

	events.LogMeleeAttackEvent(attacker, target, attackInfo.Name, attackRoll, attackInfo.AttackModifier, didHit, attacker.GetEventListener())

	if didHit {
		damage, rolls, err2 := shared.DiceRollWithModifier(attackInfo.NumberOfDice, attackInfo.Die, attackInfo.DamageModifier)
		if err2 != nil {
			return false, 0, err2
		}
		events.LogDiceRollEvent(attacker, damage, rolls, shared.DiceRollAttack, attackInfo.DamageModifier, attacker.GetEventListener())

		return true, damage, nil
	}

	return false, 0, nil
}
