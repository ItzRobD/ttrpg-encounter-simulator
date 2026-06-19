package core

type HPMethodType int

const (
	HPSetValue HPMethodType = iota
	HPSetAverage
	HPSetRoll
)

type HPConfig struct {
	HPMethod     HPMethodType     `json:"hp_set_method"`
	Value        int              `json:"value"`
	HPAverage    int              `json:"hp_average"`
	NumberOfDice int              `json:"number_of_dice"`
	HitDice      map[DiceType]int `json:"hit_dice"` // Key: Die Type, Value: Count
	AmountToAdd  int              `json:"amount_to_add"`
}
