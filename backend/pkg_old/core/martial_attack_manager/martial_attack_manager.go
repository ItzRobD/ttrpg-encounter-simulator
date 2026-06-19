package martial_attack_manager

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/events"
	"dnd5e-encounter-simulator-backend/pkg_old/core/roll_manager"
	"fmt"
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

// ProcessAttackRequest processes an attack request by performing attack rolls and calculating damage for each attack.
// It uses the attack request data and options to execute attacks and returns a list of results for each attempt.
// Returns an error if the attack roll or damage roll fails at any point.
func (mam *MartialAttackManager) ProcessAttackRequest(req *core.AttackRequest) ([]core.AttackResult, error) {
	var results []core.AttackResult

	// Index corresponds to the entry in AttackData; characters repeat via numberOfAttacks, monsters expand AttackData per swing.
	for idx, ad := range req.AttackData {
		if req.GetAttackOptions().GetNumberOfAttacks() == 0 {
			return nil, fmt.Errorf("invalid number of attacks - 0")
		}
		for i := 1; i <= req.GetAttackOptions().GetNumberOfAttacks(); i++ {
			// 1. Attack Roll
			attackMod := ad.AttackModifier + req.AttackOptions.GetBonusToAttackRoll()

			cT := 20
			if req.AttackOptions.ImprovedCritical {
				cT = 19
			}

			rollOpts := roll_manager.NewRollOptions()
			rollOpts.Advantage = req.AttackOptions.Advantage
			rollOpts.Modifier = attackMod
			rollOpts.CriticalThreshold = cT
			rollOpts.RollType = core.DiceRollAttack
			rollOpts.TargetValue = req.Target.GetAC()

			attackRollResult, err := mam.rollManager.RollAttack(rollOpts)
			if err != nil {
				return nil, err
			}

			resistBreakers := make([]core.ResistBreaker, 0)
			resistBreakers = append(resistBreakers, ad.ResistBreakers...)

			// 2. Roll damage (without logging yet)
			rollOpts = roll_manager.NewRollOptions()
			rollOpts.RollType = core.DiceRollDamage
			dmgRollResult, err := mam.rollManager.RollDamage(req, idx, attackRollResult.IsCritical, rollOpts, false)
			if err != nil {
				return nil, err
			}

			attackResult := core.AttackResult{
				ActorName:      core.FormatEntityName(mam.parent),
				TargetName:     core.FormatEntityName(req.Target),
				Target:         req.Target,
				AttackName:     ad.Name,
				AttackCount:    i,
				TargetValue:    attackRollResult.TargetValue,
				IsHit:          attackRollResult.IsSuccess,
				IsCriticalHit:  attackRollResult.IsCritical,
				AttackTotal:    attackRollResult.Total,
				AttackRoll:     attackRollResult.FinalRollValue,
				DamageRoll:     dmgRollResult,
				DamageType:     core.DamageType(ad.GetDamageType()),
				ResistBreakers: resistBreakers,
				IsRanged:       ad.IsRangedWeapon,
				AdvantageUsed:  attackRollResult.Advantage,
			}

			// 3. Log Attack (This generates the Attack's currentID)
			mam.parent.LogEvent(events.ETAttackEvent, &attackResult)

			// 4. Advance Scope: The Attack becomes the parent for subsequent events (like damage)
			ctx := mam.parent.GetCurrentEventContext()
			if ctx != nil {
				actionID := ctx.GetParentID() // Store current action ID
				ctx.AdvanceScope()

				// 4b. Log Equipment Info (as child of Attack)
				props := make([]string, 0)
				if ad.IsVersatileAttack {
					props = append(props, "Versatile")
				}
				if ad.IsFinesseWeapon {
					props = append(props, "Finesse")
				}
				if ad.IsLightWeapon {
					props = append(props, "Light")
				}
				if ad.IsTwoHandedWeapon {
					props = append(props, "Two-Handed")
				}
				if ad.IsHeavyWeapon {
					props = append(props, "Heavy")
				}
				if ad.IsThrownWeapon {
					props = append(props, "Thrown")
				}

				mods := make([]string, 0)
				for _, rb := range ad.ResistBreakers {
					if rb != core.ResistBreakerNone {
						mods = append(mods, rb.String())
					}
				}

				// Log each damage block in equipment event
				for _, db := range ad.DamageBlocks {
					mam.parent.LogEvent(events.ETEquipmentEvent, &events.EquipmentEvent{
						Name:         ad.Name,
						NumberOfDice: db.NumberOfDice,
						Die:          db.Die.String(),
						DamageType:   db.DamageType.String(),
						AttackBonus:  ad.WeaponAttackBonus + req.AttackOptions.GetBonusToAttackRoll(),
						DamageBonus:  db.Modifier + ad.WeaponDamageBonus + req.AttackOptions.GetBonusToDamageRoll(),
						IsRanged:     ad.IsRangedWeapon,
						Properties:   props,
						Modifiers:    mods,
					})
				}

				// 5. Log Damage Roll manually (it will use the Attack as parent)
				for _, comp := range dmgRollResult.DamageComponents {
					mam.parent.LogEvent(events.ETRollEvent, &events.DiceRollData{
						RollResult: &roll_manager.RollResult{
							DiceRollType:   dmgRollResult.DiceRollType,
							FinalRollValue: comp.RollValue,
							FinalRolls:     comp.DiceRolls,
							Modifier:       comp.Modifier,
							Total:          comp.Total,
							IsCritical:     comp.IsCritical,
							RerollEvents:   comp.RerollEvents,
							NumberOfDice:   comp.NumberOfDice,
							Die:            comp.Die,
						},
						DamageType: comp.DamageType.String(),
					})
				}

				results = append(results, attackResult)

				// 6. Restore Action ID for the next attack in multi-attack sequence
				ctx.SetParentID(actionID)
			} else {
				// Fallback if no context (e.g. in some tests)
				// Log damage roll normally
				mam.parent.LogEvent(events.ETRollEvent, dmgRollResult)
				results = append(results, attackResult)
			}
		}
	}

	return results, nil
}
