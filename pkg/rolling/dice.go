package rolling

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
	"math/rand/v2"
)

func RollDice(numDice int, numSides int) (int, []int, error) {
	if numDice < 1 {
		return 0, nil, fmt.Errorf("numDice must be greater than 0")
	}
	if !shared.ValidateDie(numSides) {
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
