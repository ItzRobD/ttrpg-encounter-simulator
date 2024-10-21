package shared

import (
	"fmt"
)

type ChallengeRatingPB struct {
	Ratings []float64
	Bonus   int
}

var monsterPBTable = []ChallengeRatingPB{
	{[]float64{0, 1.0 / 8, 1.0 / 4, 1.0 / 2, 1, 2, 3, 4}, 2},
	{[]float64{5, 6, 7, 8}, 3},
	{[]float64{9, 10, 11, 12}, 4},
	{[]float64{13, 14, 15, 16}, 5},
	{[]float64{17, 18, 19, 20}, 6},
	{[]float64{21, 22, 23, 24}, 7},
	{[]float64{25, 26, 27, 28}, 8},
	{[]float64{29, 30}, 9},
}

var characterPBTable = map[int]int{
	1: 2, 2: 2, 3: 2, 4: 2,
	5: 3, 6: 3, 7: 3, 8: 3,
	9: 4, 10: 4, 11: 4, 12: 4,
	13: 5, 14: 5, 15: 5, 16: 5,
	17: 6, 18: 6, 19: 6, 20: 6,
}

func GetCharacterProficiencyBonus(level int) (int, error) {
	if bonus, exists := characterPBTable[level]; exists {
		return bonus, nil
	}
	return 0, fmt.Errorf("level must be between 1 and 20")
}

func GetMonsterProficiencyBonus(challengeRating float64) (int, error) {
	if challengeRating < 0 || challengeRating > 30 {
		return 0, fmt.Errorf("challenge rating must be between 0 and 30")
	}

	for _, entry := range monsterPBTable {
		for _, cr := range entry.Ratings {
			if cr == challengeRating {
				return entry.Bonus, nil
			}
		}
	}
	return 0, fmt.Errorf("challenge rating not found")
}

type AbilityScoreModifier struct {
	Scores   []int
	Modifier int
}

var abilityScoreModifiers = []AbilityScoreModifier{
	{[]int{1}, -5},
	{[]int{2, 3}, -4},
	{[]int{4, 5}, -3},
	{[]int{6, 7}, -2},
	{[]int{8, 9}, -1},
	{[]int{10, 11}, 0},
	{[]int{12, 13}, 1},
	{[]int{14, 15}, 2},
	{[]int{16, 17}, 3},
	{[]int{18, 19}, 4},
	{[]int{20, 21}, 5},
	{[]int{22, 23}, 6},
	{[]int{24, 25}, 7},
	{[]int{26, 27}, 8},
	{[]int{28, 29}, 9},
	{[]int{30}, 10},
}

func GetAbilityScoreModifier(score int) (int, error) {
	if score < 1 || score > 30 {
		return 0, fmt.Errorf("score must be between 1 and 30")
	}
	for _, entry := range abilityScoreModifiers {
		for _, s := range entry.Scores {
			if s == score {
				return entry.Modifier, nil
			}
		}
	}

	return 0, fmt.Errorf("score modifier not found")
}

var validDieValues = []int{4, 6, 8, 10, 12, 20, 100}

func ValidateDie(die int) bool {
	for _, v := range validDieValues {
		if v == die {
			return true
		}
	}
	return false
}

var validDamageTypes = []string{"acid", "bludgeoning", "cold", "fire", "force", "lightning",
	"necrotic", "piercing", "poison", "psychic", "radiant", "slashing", "thunder"}

func ValidateDamageType(damageType string) bool {
	for _, v := range validDamageTypes {
		if v == damageType {
			return true
		}
	}
	return false
}
