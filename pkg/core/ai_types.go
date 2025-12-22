package core

import (
	"fmt"
)

type TargetType int

const (
	TTDamage TargetType = iota
	TTHealing
)

type AIRequestType int

const (
	AIReqNormalAction AIRequestType = iota
	AIReqLegendaryAction
	AIReqOffhandAttack
	AIReqDragonbornBreathWeapon
)

type AIRequest struct {
	Actor        Entity
	ActorType    EntityType
	ActorID      int
	Target       Entity
	TargetID     int
	ActionType   ActionType
	SpellChoice  *SpellChoice
	WeaponSlot   WeaponSlot
	UseVersatile bool
	ActionIndex  int // Monsters only
	Advantage    AdvantageType
	Request      AIRequestType
	SimOptions   *SimulationOptions
}

func (r *AIRequest) Validate() error {
	if r.Actor == nil {
		return fmt.Errorf("actor cannot be nil")
	}
	if r.ActionType == ATHeal && r.SpellChoice == nil {
		return fmt.Errorf("spell choice cannot be nil for healing")
	}
	if r.TargetID == -1 {
		return fmt.Errorf("target id was not selected")
	}
	return nil
}
