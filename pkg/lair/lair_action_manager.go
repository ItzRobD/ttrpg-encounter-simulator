package lair

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
)

// LairActionManager is a minimal action processor for lair actions.
// It mirrors the monster action flow without recharge/legendary hooks.
type LairActionManager struct {
	parent      *Lair
	rollManager *roll_manager.RollManager
	// Precomputed actions for the lair. Key: action index
	Actions map[int]core.AttackData
}

func NewLairActionManager(parent *Lair, rm *roll_manager.RollManager) *LairActionManager {
	return &LairActionManager{
		parent:      parent,
		rollManager: rm,
		Actions:     make(map[int]core.AttackData),
	}
}

func (lam *LairActionManager) AddAction(index int, ad core.AttackData) {
	lam.Actions[index] = ad
}

func (lam *LairActionManager) GetAttackDataFromIndex(index int) []core.AttackData {
	if ad, ok := lam.Actions[index]; ok {
		return []core.AttackData{ad}
	}
	return nil
}

func (lam *LairActionManager) ProcessAttackRequest(req *core.AttackRequest) ([]core.AttackResult, error) {
	var results []core.AttackResult

	for idx, ad := range req.AttackData {
		// Attack Roll
		attackMod := ad.AttackModifier + req.AttackOptions.GetBonusToAttackRoll()
		cT := 20
		if req.AttackOptions.GetIsImprovedCritical() {
			cT = 19
		}

		rollOpts := roll_manager.NewRollOptions()
		rollOpts.Advantage = req.AttackOptions.GetAdvantage()
		rollOpts.Modifier = attackMod
		rollOpts.CriticalThreshold = cT
		rollOpts.RollType = core.DiceRollAttack
		rollOpts.TargetValue = req.Target.GetAC()

		attackRollResult, err := lam.rollManager.RollAttack(rollOpts)
		if err != nil {
			return nil, err
		}

		// Damage roll
		rollOpts = roll_manager.NewRollOptions()
		rollOpts.Modifier = ad.DamageModifier + req.AttackOptions.GetBonusToDamageRoll()
		rollOpts.RollType = core.DiceRollDamage

		dmgRollResult, err := lam.rollManager.RollDamage(req, idx, attackRollResult.IsCritical, rollOpts)
		if err != nil {
			return nil, err
		}

		attackResult := core.AttackResult{
			ActorName:     lam.parent.GetName(),
			TargetName:    req.Target.GetName(),
			AttackName:    ad.Name,
			AttackCount:   idx + 1,
			TargetValue:   attackRollResult.TargetValue,
			IsHit:         attackRollResult.IsSuccess,
			IsCriticalHit: attackRollResult.IsCritical,
			AttackTotal:   attackRollResult.Total,
			AttackRoll:    attackRollResult.FinalRollValue,
			DamageRoll:    dmgRollResult,
			DamageType:    ad.DamageType,
		}

		events.LogMeleeAttackEvent(lam.parent, &attackResult, lam.parent.GetEventListener())
		if attackRollResult.IsSuccess {
			events.LogDiceRollEvent(lam.parent, dmgRollResult, lam.parent.GetEventListener())
		}
		results = append(results, attackResult)
	}

	return results, nil
}
