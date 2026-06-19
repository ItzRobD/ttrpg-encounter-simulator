package core

import (
	"github.com/google/uuid"
)

func GetAverageRoll(numDice int, die DiceType, amtToAdd int) (int, error) {
	dAvg := die.Avg()
	return int(dAvg*float64(numDice) + float64(amtToAdd)), nil
}

// NewUUIDv7 generates and returns a new UUID version 7 as a string.
func NewUUIDv7() string {
	u, _ := uuid.NewV7()
	return u.String()
}
