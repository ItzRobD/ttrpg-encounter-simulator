package core

import (
	"fmt"
	"math/rand/v2"
)

//type AdvantageType int
//
//const (
//	RollNormal AdvantageType = iota
//	RollAdvantage
//	RollDisadvantage
//)
//
//type DiceRollType string
//
//const (
//	DiceRollAttack     DiceRollType = "attack"
//	DiceRollDamage     DiceRollType = "damage"
//	DiceRollInitiative DiceRollType = "initiative"
//)

//type DiceType int
//
//const (
//	D4   DiceType = 4
//	D6   DiceType = 6
//	D8   DiceType = 8
//	D10  DiceType = 10
//	D12  DiceType = 12
//	D20  DiceType = 20
//	D100 DiceType = 100
//)

// InitiativeRoll calculates an initiative score based on dexterity, bonus, and advantage, returning the score, base roll, and error.
// dexterity must be between 1 and 30. Returns an error if inputs are invalid or rolling fails.
func InitiativeRoll(dexterity int, bonus int, advantage AdvantageType) (int, int, error) {
	if dexterity < 1 || dexterity > 30 {
		return 0, 0, fmt.Errorf("initiative roll - dexterity must be between 1 and 30")
	}
	modifier, err := GetAbilityScoreModifier(dexterity)
	if err != nil {
		return 0, 0, err
	}

	roll, err := RollD20WithAdvantage(advantage)
	if err != nil {
		return 0, 0, err
	}

	return roll + modifier + bonus, roll, nil
}

// RollD20WithAdvantage rolls a 20-sided die with normal, advantage, or disadvantage based on the given AdvantageType.
// Returns the final roll result, all individual roll values, and an error if rolling fails or the advantage type is invalid.
func RollD20WithAdvantage(advantage AdvantageType) (int, []int, error) {
	switch advantage {
	case RollNormal:
		_, rolls, err := RollDice(1, 20)
		if err != nil {
			return 0, err
		}
		return rolls[0], rolls, nil
	case RollAdvantage:
		_, rolls, err := RollDice(2, 20)
		if err != nil {
			return 0, err
		}
		return max(rolls[0], rolls[1]), rolls, nil
	case RollDisadvantage:
		_, rolls, err := RollDice(2, 20)
		if err != nil {
			return 0, err
		}
		return min(rolls[0], rolls[1]), rolls, nil
	default:
		return 0, nil, fmt.Errorf("invalid advantage type")
	}
}

func RerollLowerD20(rolls []int) ([]int, []int, error) {

}

// AttackRoll calculates the attack roll value by adding a roll of a D20 (considering advantage/disadvantage) to a modifier.
// Returns the final attack roll value, the raw roll value, and an error if rolling fails.
func AttackRoll(modifier int, advantage AdvantageType) (int, []int, error) {
	roll, rolls, err := RollD20WithAdvantage(advantage)
	if err != nil {
		return 0, nil, err
	}
	return roll + modifier, rolls, nil
}

func AttackRollD20(advantage AdvantageType) (int, error) {
	return RollD20WithAdvantage(advantage)
}

// DiceRollWithModifier rolls a number of dice, adds specified modifier, and returns total, individual rolls, and errors if any.
// numberOfDice specifies how many dice should be rolled and must be greater than 0.
// numberOfSides defines the type of dice to roll, which must be a valid die type.
// amountToAdd is a flat modifier added to the total result.
// Returns the total roll result, a slice of individual rolls, and an error if input validation fails.
func DiceRollWithModifier(numberOfDice, numberOfSides int, amountToAdd int) (int, []int, error) {
	if numberOfDice < 1 {
		return 0, nil, fmt.Errorf("number of dice to roll must be greater than 0")
	}
	if !ValidateDie(numberOfSides) {
		return 0, nil, fmt.Errorf("invalid die type")
	}
	s, rolls, err := RollDice(numberOfDice, numberOfSides)
	if err != nil {
		return 0, nil, err
	}
	total := s + amountToAdd

	return total, rolls, nil
}

func CalculateDamageCriticalHit(numberOfDice, numberOfSides int, amountToAdd int, improvedCritical bool) (int, []int, error) {
	if numberOfDice < 1 {
		return 0, nil, fmt.Errorf("number of dice to roll must be greater than 0")
	}
	if !ValidateDie(numberOfSides) {
		return 0, nil, fmt.Errorf("invalid die type")
	}
	if improvedCritical {
		s, rolls, err := RollDice(numberOfDice, numberOfSides)
		if err != nil {
			return 0, nil, err
		}
		total := s + amountToAdd + (numberOfDice * numberOfSides)
		for range numberOfDice {
			rolls = append(rolls, numberOfSides)
		}
		return total, rolls, nil
	}

	s, rolls, err := RollDice(numberOfDice*2, numberOfSides)
	if err != nil {
		return 0, nil, err
	}
	total := s + amountToAdd
	return total, rolls, nil
}

// RollDice rolls a specified number of dice with a given number of sides and returns the total sum, individual rolls, and any error.
// numDice is the number of dice to roll and must be greater than 0.
// numSides is the number of sides per die and must be a valid die type (4, 6, 8, 10, 12, 20, or 100).
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
