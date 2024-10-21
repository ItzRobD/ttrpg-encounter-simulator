package rolling

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
)

func InitiativeRoll(dexterity int) (int, error) {
	if dexterity < 1 || dexterity > 30 {
		return 0, fmt.Errorf("dexterity must be between 1 and 30")
	}
	modifier, err := shared.GetAbilityScoreModifier(dexterity)
	if err != nil {
		return 0, err
	}
	_, rolls, err := RollDice(1, 20)
	if err != nil {
		return 0, err
	}
	i := rolls[0] + modifier
	if i < 1 {
		i = 1
	}
	return i, nil
}

func AttackRoll(modifier int) (int, error) {
	_, rolls, err := RollDice(1, 20)
	if err != nil {
		return 0, err
	}
	return rolls[0] + modifier, nil
}

func AttackHits(ar int, ac int) bool {
	if ar >= ac {
		return true
	}
	return false
}

func DamageRoll(numberOfDice, numberOfSides int, amountToAdd int) (int, []int, error) {
	if numberOfDice < 1 || numberOfDice > 100 {
		return 0, nil, fmt.Errorf("number of rolling must be between 1 and 100")
	}
	if !shared.ValidateDie(numberOfSides) {
		return 0, nil, fmt.Errorf("invalid die type")
	}
	s, rolls, err := RollDice(numberOfDice, numberOfSides)
	if err != nil {
		return 0, nil, err
	}
	dmg := s + amountToAdd

	return dmg, rolls, nil
}
