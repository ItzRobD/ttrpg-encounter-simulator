package core

import (
	"fmt"
	"strings"
)

type ActorConditions map[Condition]bool

type Condition string

const (
	ConditionNone          Condition = "none"
	ConditionBlinded       Condition = "blinded"
	ConditionCharmed       Condition = "charmed"
	ConditionDeafened      Condition = "deafened"
	ConditionFrightened    Condition = "frightened"
	ConditionGrappled      Condition = "grappled"
	ConditionIncapacitated Condition = "incapacitated"
	ConditionInvisible     Condition = "invisible"
	ConditionParalyzed     Condition = "paralyzed"
	ConditionPetrified     Condition = "petrified"
	ConditionPoisoned      Condition = "poisoned"
	ConditionProne         Condition = "prone"
	ConditionRestrained    Condition = "restrained"
	ConditionStunned       Condition = "stunned"
	ConditionUnconscious   Condition = "unconscious"
	ConditionStable        Condition = "stable"
	// ConditionReckless Applied when a Barbarian (or similar feature) uses Reckless Attack; everyone has advantage to hit them
	ConditionReckless Condition = "reckless"
	// Feature Conditions
	ConditionBerserk Condition = "berserk"
)

func (c Condition) String() string {
	return string(c)
}

func NewActorConditions() ActorConditions {
	return map[Condition]bool{
		ConditionBlinded:       false,
		ConditionCharmed:       false,
		ConditionDeafened:      false,
		ConditionFrightened:    false,
		ConditionGrappled:      false,
		ConditionIncapacitated: false,
		ConditionInvisible:     false,
		ConditionParalyzed:     false,
		ConditionPetrified:     false,
		ConditionPoisoned:      false,
		ConditionProne:         false,
		ConditionRestrained:    false,
		ConditionStunned:       false,
		ConditionUnconscious:   false,
		ConditionStable:        false,
		ConditionReckless:      false,
		ConditionBerserk:       false,
	}
}

func (ec ActorConditions) Add(c Condition) {
	if ec == nil {
		return
	}
	ec[c] = true
}

func (ec ActorConditions) Remove(c Condition) {
	if ec == nil {
		return
	}
	delete(ec, c)
}

func (ec ActorConditions) Has(c Condition) bool {
	if ec == nil {
		return false
	}
	return ec[c]
}

func (ec ActorConditions) Clear() {
	for c := range ec {
		ec[c] = false
	}
}

func (ec ActorConditions) GetActive() []Condition {
	var active []Condition
	for c := range ec {
		if ec[c] {
			active = append(active, c)
		}
	}
	return active
}

func MakeCondition(s string) (Condition, error) {
	switch strings.ToLower(s) {
	case "blinded":
		return ConditionBlinded, nil
	case "charmed":
		return ConditionCharmed, nil
	case "deafened":
		return ConditionDeafened, nil
	case "frightened":
		return ConditionFrightened, nil
	case "grappled":
		return ConditionGrappled, nil
	case "incapacitated":
		return ConditionIncapacitated, nil
	case "invisible":
		return ConditionInvisible, nil
	case "paralyzed":
		return ConditionParalyzed, nil
	case "petrified":
		return ConditionPetrified, nil
	case "poisoned":
		return ConditionPoisoned, nil
	case "prone":
		return ConditionProne, nil
	case "restrained":
		return ConditionRestrained, nil
	case "stunned":
		return ConditionStunned, nil
	case "unconscious":
		return ConditionUnconscious, nil
	case "stable":
		return ConditionStable, nil
	case "reckless_exposed":
		return ConditionReckless, nil
	default:
		return ConditionNone, fmt.Errorf("invalid condition")
	}
}
