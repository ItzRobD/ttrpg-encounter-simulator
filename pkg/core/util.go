package core

import (
	"fmt"
	"strings"
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

var characterPBTable = map[uint8]int{
	1: 2, 2: 2, 3: 2, 4: 2,
	5: 3, 6: 3, 7: 3, 8: 3,
	9: 4, 10: 4, 11: 4, 12: 4,
	13: 5, 14: 5, 15: 5, 16: 5,
	17: 6, 18: 6, 19: 6, 20: 6,
}

func GetCharacterProficiencyBonus(level uint8) (int, error) {
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

var dieAverageValues = map[int]float64{
	4:  2.5,
	6:  3.5,
	8:  4.5,
	10: 5.5,
	12: 6.5,
	20: 10.5,
}

func GetDieAverage(die DiceType) (float64, error) {
	return dieAverageValues[die.Int()], nil
}

// GetAverageRoll calculates the average roll of a specified number of dice with a modifier and returns the result or an error.
func GetAverageRoll(numDice int, die DiceType, amtToAdd int) (int, error) {
	dAvg, err := GetDieAverage(die)
	if err != nil {
		return 0, err
	}
	return int(dAvg*float64(numDice) + float64(amtToAdd)), nil
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

func GetNormalizedAbility(ability string) (Ability, error) {
	switch strings.ToLower(ability) {
	case "str", "strength":
		return AbilityStrength, nil
	case "dex", "dexterity":
		return AbilityDexterity, nil
	case "con", "constitution":
		return AbilityConstitution, nil
	case "int", "intelligence":
		return AbilityIntelligence, nil
	case "wis", "wisdom":
		return AbilityWisdom, nil
	case "cha", "charisma":
		return AbilityCharisma, nil
	default:
		return AbilityNone, fmt.Errorf("invalid ability")
	}
}

// DetermineAttackAdvantageFromConditions calculates attack roll advantage or disadvantage based on actor and target conditions.
// It evaluates each condition's effect on outgoing and incoming attack rolls and returns the final AdvantageType.
func DetermineAttackAdvantageFromConditions(actorConditions EntityConditions, targetConditions EntityConditions) AdvantageType {
	advCount := 0
	if actorConditions == nil && targetConditions == nil {
		return RollNormal
	}

	if actorConditions != nil {
		for c, _ := range actorConditions {
			if actorConditions[c] {
				e := GetConditionEffects(c)
				if e.OutgoingAttackRoll == RollAdvantage {
					advCount++
				}
				if e.OutgoingAttackRoll == RollDisadvantage {
					advCount--
				}
			}
		}
	}

	if targetConditions != nil {
		for c, _ := range targetConditions {
			if targetConditions[c] {
				e := GetConditionEffects(c)
				if e.IncomingAttackRoll == RollAdvantage {
					advCount++
				}
				if e.IncomingAttackRoll == RollDisadvantage {
					advCount--
				}
			}
		}
	}

	if advCount > 0 {
		return RollAdvantage
	} else if advCount < 0 {
		return RollDisadvantage
	}

	return RollNormal
}

// DetermineSaveAdvantageFromConditions evaluates the effective saving throw advantage or disadvantage based on current conditions.
func DetermineSaveAdvantageFromConditions(actorConditions EntityConditions, ability Ability) AdvantageType {
	advCount := 0
	if actorConditions == nil {
		return RollNormal
	}

	for c, _ := range actorConditions {
		if actorConditions[c] {
			e := GetConditionEffects(c)
			if e.SavingThrow[ability] == RollAdvantage {
				advCount++
			} else if e.SavingThrow[ability] == RollDisadvantage {
				advCount--
			}
		}
	}

	if advCount > 0 {
		return RollAdvantage
	} else if advCount < 0 {
		return RollDisadvantage
	}

	return RollNormal
}

func GetFinalAdvantageType(advs []AdvantageType) AdvantageType {
	advCount := 0
	for _, a := range advs {
		if a == RollAdvantage {
			advCount++
		} else if a == RollDisadvantage {
			advCount--
		}
	}

	if advCount > 0 {
		return RollAdvantage
	} else if advCount < 0 {
		return RollDisadvantage
	}

	return RollNormal
}
