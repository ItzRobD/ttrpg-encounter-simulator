package shared

type MonsterHP struct {
	HP           int   `json:"hp"`
	MaxHP        int   `json:"maxHP"`
	HPAverage    int   `json:"hpAverage"`
	NumberOfDice int   `json:"numberOfDice"`
	Die          int   `json:"die"`
	AmountToAdd  int   `json:"amountToAdd"`
	Rolls        []int `json:"hpRolls"`
}

// This may be unnecessary -> placeholder
type PlayerHP struct {
	HP     int   `json:"hp"`
	MaxHP  int   `json:"maxHP"`
	HitDie int   `json:"hitDie"`
	Rolls  []int `json:"hpRolls"`
}
