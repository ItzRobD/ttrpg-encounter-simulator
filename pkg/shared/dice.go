package shared

import (
	"fmt"
	"math/rand"
	"time"
)

func RollDice(numDice int, numSides int) ([]int, error) {
	if numDice < 1 {
		return nil, fmt.Errorf("numDice must be greater than 0")
	}
	if !ValidateDie(numSides) {
		return nil, fmt.Errorf("numSides must be 4, 6, 8, 10, 12, 20, or 100")
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))

	rolls := make([]int, numDice)
	for i := 0; i < numDice; i++ {
		rolls[i] = rand.Intn(numSides) + 1
	}

	return rolls, nil
}
