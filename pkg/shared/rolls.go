package shared

import (
	"fmt"
	"math/rand/v2"
)

func InitiativeRoll(dexterity int) (int, error) {
	if dexterity < 1 || dexterity > 30 {
		return 0, fmt.Errorf("initiative roll - dexterity must be between 1 and 30")
	}
	modifier, err := GetAbilityScoreModifier(dexterity)
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

func DiceRollWithModifier(numberOfDice, numberOfSides int, amountToAdd int) (int, []int, error) {
	if numberOfDice < 1 || numberOfDice > 100 {
		return 0, nil, fmt.Errorf("number of rolling must be between 1 and 100")
	}
	if !ValidateDie(numberOfSides) {
		return 0, nil, fmt.Errorf("invalid die type")
	}
	s, rolls, err := RollDice(numberOfDice, numberOfSides)
	if err != nil {
		return 0, nil, err
	}
	dmg := s + amountToAdd

	return dmg, rolls, nil
}

func RollDice(numDice int, numSides int) (int, []int, error) {
	if numDice < 1 {
		return 0, nil, fmt.Errorf("numDice must be greater than 0")
	}
	if !ValidateDie(numSides) {
		return 0, nil, fmt.Errorf("numSides must be 4, 6, 8, 10, 12, 20, or 100")
	}

	rand.NewPCG(rand.Uint64(), rand.Uint64())

	rolls := make([]int, numDice)
	for i := 0; i < numDice; i++ {
		rolls[i] = rand.IntN(numSides) + 1
	}

	s := sum(rolls)

	return s, rolls, nil
}

func sum(arr []int) int {
	s := 0
	for _, v := range arr {
		s += v
	}
	return s
}
