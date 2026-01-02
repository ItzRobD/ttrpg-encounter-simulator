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
	AIReqDeathEffect
)

type AIRequest struct {
	Actor              Entity
	ActorType          EntityType
	ActorID            int
	Target             Entity
	TargetID           int
	ActionType         ActionType
	SpellChoice        *SpellChoice
	HealRequest        *HealRequest
	WeaponSlot         WeaponSlot
	UseVersatile       bool
	ActionIndex        int // Monsters only
	Advantage          AdvantageType
	Request            AIRequestType
	SimOptions         *SimulationOptions
	LayingOnHandsValue int
}

func (r *AIRequest) Validate() error {
	if r.Actor == nil {
		return fmt.Errorf("actor cannot be nil")
	}
	if r.ActionType == ATHeal && r.HealRequest == nil && r.SpellChoice == nil && r.LayingOnHandsValue == 0 {
		return fmt.Errorf("healing request must provide either a heal request, a spell, or an ability value")
	}
	if r.TargetID == -1 {
		return fmt.Errorf("target id was not selected")
	}
	return nil
}
