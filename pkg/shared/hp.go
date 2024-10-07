package shared

type MonsterHP struct {
	HPAverage    int `json:"hpAverage"`
	NumberOfDice int `json:"numberOfDice"`
	Die          int `json:"die"`
	AmountToAdd  int `json:"amountToAdd"`
}

// This may be unnecessary -> placeholder
type PlayerHP struct {
	HP     int `json:"hp"`
	HitDie int `json:"hitDie"`
}
