package core

import (
	"fmt"
	"strings"
)

type DeathSaves map[SaveType]int

type SaveType string

const (
	SaveSuccess SaveType = "success"
	SaveFailure SaveType = "failure"
)

func (st SaveType) String() string {
	return string(st)
}

func NewSaveType(s string) (SaveType, error) {
	switch strings.ToLower(s) {
	case "success":
		return SaveSuccess, nil
	case "failure":
		return SaveFailure, nil
	default:
		return SaveSuccess, fmt.Errorf("invalid save type")
	}
}

func (ds DeathSaves) GetSave(saveType SaveType) int {
	return ds[saveType]
}

func (ds DeathSaves) SetSave(saveType SaveType, value int) {
	ds[saveType] = value
}

func (ds DeathSaves) AddSave(saveType SaveType, value int) {
	ds[saveType] += value
}

func (ds DeathSaves) SubtractSave(saveType SaveType, value int) {
	ds[saveType] -= value
}

func (ds DeathSaves) ResetDeathSaves() {
	ds = ds.NewDeathSaves()
}

func (ds DeathSaves) NewDeathSaves() DeathSaves {
	return DeathSaves{
		SaveSuccess: 0,
		SaveFailure: 0,
	}
}

type HPSetMethod uint8

const (
	HPSetValue HPSetMethod = iota
	HPSetAverage
	HPSetRoll
)
