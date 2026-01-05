package core

import (
	"fmt"
	"strings"
)

type EntityType string

const (
	EntityCharacter EntityType = "character"
	EntityMonster   EntityType = "monster"
	EntityLair      EntityType = "lair"
	EntityUnknown   EntityType = "unknown"
)

func (et EntityType) String() string {
	return string(et)
}

func NewEntityType(s string) (EntityType, error) {
	switch strings.ToLower(s) {
	case "character":
		return EntityCharacter, nil
	case "monster":
		return EntityMonster, nil
	default:
		return EntityCharacter, fmt.Errorf("invalid entity type")
	}
}

type DeathSaveEvaluation string

const (
	DeathSaveSuccess DeathSaveEvaluation = "success"
	DeathSaveFailure DeathSaveEvaluation = "failure"
	DeathSaveNone    DeathSaveEvaluation = "none"
)

type DeathSaves struct {
	SaveSuccess int
	SaveFailure int
}

func NewDeathSaves() DeathSaves {
	return DeathSaves{
		SaveSuccess: 0,
		SaveFailure: 0,
	}
}

func (ds *DeathSaves) Reset() {
	ds.SaveSuccess = 0
	ds.SaveFailure = 0
}

func (ds *DeathSaves) String() string {
	return fmt.Sprintf("Success: %d, Failure: %d", ds.SaveSuccess, ds.SaveFailure)
}

func (ds *DeathSaves) AddSuccess() {
	ds.SaveSuccess++
}

func (ds *DeathSaves) AddFailure(addDouble bool) {
	if addDouble {
		ds.SaveFailure += 2
	} else {
		ds.SaveFailure++
	}
}

func (ds *DeathSaves) Evaluate() DeathSaveEvaluation {
	if ds.SaveSuccess >= 3 {
		return DeathSaveSuccess
	}
	if ds.SaveFailure >= 3 {
		return DeathSaveFailure
	}

	return DeathSaveNone
}

type HPSetMethod uint8

const (
	HPSetValue HPSetMethod = iota
	HPSetAverage
	HPSetRoll
)

type HPConfig struct {
	HPSetMethod  HPSetMethod
	Value        int
	HPAverage    int
	NumberOfDice int
	HitDie       DiceType
	AmountToAdd  int
	Modifier     int
}

type Seed struct {
	Seed1 uint64
	Seed2 uint64
}

type TurnResult struct {
	TurnStatuses map[TurnStatus]bool
	Conditions   []Condition
}

type TurnStatus string

type TurnType string

const (
	TurnTypeNormal    TurnType = "normal"
	TurnTypeLegendary TurnType = "legendary"
)

const (
	TurnActionReady           TurnStatus = "ready"
	TurnIncapacitated         TurnStatus = "inactive"
	TurnDeathSaveFailed       TurnStatus = "death_save_failed"
	TurnDeathSaveFailedDouble TurnStatus = "death_save_failed_double"
	TurnDeathSaveSuccess      TurnStatus = "death_save_success"
	TurnDeathSaveStabilized   TurnStatus = "death_save_stabilized"
	TurnDead                  TurnStatus = "dead"
	TurnUnconscious           TurnStatus = "unconscious"
	TurnRevived               TurnStatus = "revived"
	TurnLegendaryReady        TurnStatus = "legendary_ready"
	TurnLegendaryUnavailable  TurnStatus = "legendary_unavailable"
)
