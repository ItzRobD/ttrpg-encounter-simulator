package martial_attack_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
)

type MartialAttackManager struct {
	parent      core.Entity
	rollManager *roll_manager.RollManager
}

func NewMartialAttackManager(parent core.Entity, rm *roll_manager.RollManager) *MartialAttackManager {
	return &MartialAttackManager{
		parent:      parent,
		rollManager: rm,
	}
}

/*
Melee Attack flow:

Encounter handle turn
	Character:
		Chooses melee or spell attack
		Choose target
		Melee attack:
			Decides which weapon to use -> Generates attack request
				Create Attack request contains target
			Call martial manager
				Make martial attack
					Gets parent info for multi attacks
					Sets roll options
					Call roll manager
						Roll attack dice
							Log
					If hit, call roll manager
						roll damage dice
							Log
	Monster:
		Chooses melee or spell attack
		Choose target
		Melee Attack:
			Decide how to proceed -> determine multiattack or single attacks
				Create attack request containing target and actions
			Call martial manager
				Make martial attack
					Sets roll options
					Call roll manager
						Roll attack dice
							Log
					If hit, call roll manager
						roll damage dice
							Log

*/

// ProcessAttackRequest processes an attack request by performing attack rolls and calculating damage for each attack.
// It uses the attack request data and options to execute attacks and returns a list of results for each attempt.
// Returns an error if the attack roll or damage roll fails at any point.
func (mam *MartialAttackManager) ProcessAttackRequest(req *AttackRequest) ([]AttackResult, error) {
	var results []AttackResult

	for i := 0; i < req.GetAttackOptions().GetNumberOfAttacks(); i++ {
		// Attack Roll
		attackMod := req.AttackData.AttackModifier + req.AttackOptions.BonusToAttackRoll

		cT := 20
		if req.AttackOptions.ImprovedCritical {
			cT = 19
		}

		rollOpts := roll_manager.RollOptions{
			Advantage: req.AttackOptions.Advantage,
			Modifier:  attackMod,
			//RerollAbilities:   mam.rollManager.RerollAbilities,
			CriticalThreshold: cT,
			TreatOnesAsTwos:   false,
			RollType:          core.DiceRollAttack,
			TargetValue:       req.Target.GetAC(),
		}

		//attackRollResult, err := mam.rollManager.RollD20(rollOpts)
		attackRollResult, err := mam.rollManager.RollAttack(rollOpts)
		if err != nil {
			return nil, err
		}

		// roll damage
		rollOpts = roll_manager.RollOptions{
			Advantage:         core.RollNormal,
			Modifier:          0,     // Set within damage function
			CriticalThreshold: 0,     // Not relevant to damage function
			TreatOnesAsTwos:   false, // Not relevant
			RollType:          core.DiceRollDamage,
			TargetValue:       0, // Not relevant
		}
		dmgRollResult, err := mam.rollManager.RollDamage(req, attackRollResult.IsCritical, rollOpts)
		if err != nil {
			return nil, err
		}

		attackResult := AttackResult{
			ActorName:     mam.parent.GetName(),
			TargetName:    req.Target.GetName(),
			AttackName:    req.AttackData.Name,
			AttackCount:   i,
			TargetValue:   attackRollResult.TargetValue,
			IsHit:         attackRollResult.IsSuccess,
			IsCriticalHit: attackRollResult.IsCritical,
			AttackTotal:   attackRollResult.Total,
			AttackRoll:    attackRollResult.FinalRollValue,
			DamageRoll:    dmgRollResult,
			DamageType:    req.AttackData.DamageType,
		}

		events.LogMeleeAttackEvent(mam.parent, &attackResult, mam.parent.GetEventListener())
		events.LogDiceRollEvent(mam.parent, dmgRollResult, mam.parent.GetEventListener())
		results = append(results, attackResult)
	}

	return results, nil
}
