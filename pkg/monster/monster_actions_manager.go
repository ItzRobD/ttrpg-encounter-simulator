package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"fmt"
)

type MonsterActionManager struct {
	parent      *Monster
	rollManager *roll_manager.RollManager

	// Raw
	Actions          map[int]Action        // Key: ActionID; Value: Action
	Multiattacks     map[int][]Multiattack // Key: Option Index; Value: Slice of ActionIDs and Count
	LegendaryActions map[int]LegendaryAction
	RechargeActions  map[int]uint8

	// Precomputed
	ActionAttackData    map[int]core.AttackData // actionID -> AttackData
	MultiAttackData     map[int]MultiattackData // option index -> complete action
	LegendaryAttackData map[int]core.AttackData // Legendary action id -> Attack Data
	SpecialAbilities    *SpecialAbilities
}

func (mam *MonsterActionManager) GetActions() map[int]Action {
	return mam.Actions
}

func (mam *MonsterActionManager) GetActionAtIndex(index int) Action { return mam.Actions[index] }

func (mam *MonsterActionManager) GetMulitattacks() map[int][]Multiattack {
	return mam.Multiattacks
}

func (mam *MonsterActionManager) GetLegendaryActions() map[int]LegendaryAction {
	return mam.LegendaryActions
}

func (mam *MonsterActionManager) GetRechargeActionStatus() map[int]bool {
	return mam.parent.EntityStateManager.GetRechargeActionStatus()
}

func (mam *MonsterActionManager) ExpendRechargeAction(actionID int) {
	mam.parent.EntityStateManager.ExpendRechargeAction(actionID)
}

func (mam *MonsterActionManager) RechargeAction(actionID int) {
	mam.parent.EntityStateManager.RechargeRechargeAction(actionID)
}

func (mam *MonsterActionManager) RollRechargeActions() {
	idxs := mam.parent.EntityStateManager.GetExpendedRechargeActionsIndex()
	if len(idxs) == 0 {
		return
	}

	for _, idx := range idxs {
		action := mam.Actions[idx]
		rollOpts := roll_manager.NewRollOptions()
		rollOpts.RollType = core.DiceRollRecharge
		rollOpts.TargetValue = action.RechargeValue

		res := mam.rollManager.RollRecharge(rollOpts)
		res.Name = action.Name

		if res.IsSuccess {
			mam.parent.EntityStateManager.RechargeRechargeAction(idx)
		}

		mam.parent.LogEvent(events.ETRollEvent, res)
	}
}

func (mam *MonsterActionManager) GetAttackDataFromIndex(index int, actionType core.ActionType) []core.AttackData {
	switch actionType {
	case core.ATMonsterAction:
		if data, ok := mam.ActionAttackData[index]; ok {
			return []core.AttackData{data}
		}
	case core.ATMonsterMultiattack:
		if data, ok := mam.MultiAttackData[index]; ok {
			return data.AttackDataBlocks
		}
	case core.ATLegendaryAction:
		if data, ok := mam.LegendaryAttackData[index]; ok {
			return []core.AttackData{data}
		}
	case core.ATMonsterSpecial:
	}
	return nil
}

// InitializeActions sets up actions, multiattacks, legendary actions, and special abilities from the provided configuration.
func (mam *MonsterActionManager) InitializeActions(config *MAMConfig) {
	mam.Actions = config.Actions
	mam.Multiattacks = config.Multiattacks
	mam.LegendaryActions = config.LegendaryActions
	mam.SpecialAbilities = &config.SpecialAbilities

	if mam.RechargeActions == nil {
		mam.RechargeActions = make(map[int]uint8)
	}
	for _, action := range mam.Actions {
		if action.RechargeValue > 0 {
			mam.parent.EntityStateManager.AddRechargeAction(action.ActionID)
			mam.RechargeActions[action.ActionID] = 1
		}
	}

	mam.precomputeAttackData()
}

// NewMonsterActionManager initializes and returns a MonsterActionManager with the provided parent, roll manager, and configuration.
// If the provided config is nil, an empty configuration is used for initialization. Actions must be initialized later.
func NewMonsterActionManager(parent *Monster, rm *roll_manager.RollManager, config *MAMConfig) *MonsterActionManager {
	mam := &MonsterActionManager{
		parent:          parent,
		rollManager:     rm,
		RechargeActions: make(map[int]uint8), // ensure non-nil
	}
	if config != nil {
		mam.Actions = config.Actions
		mam.Multiattacks = config.Multiattacks
		mam.LegendaryActions = config.LegendaryActions
		mam.precomputeAttackData()

		for _, a := range mam.Actions {
			if a.RechargeValue > 0 {
				mam.RechargeActions[a.ActionID] = 1
			}
		}
	}
	return mam
}

func (mam *MonsterActionManager) ProcessAttackRequest(req *core.AttackRequest) ([]core.AttackResult, error) {
	var results []core.AttackResult

	for idx, ad := range req.AttackData {
		// Apply special abilities if enabled
		if req.SimulationOptions != nil && req.SimulationOptions.EnableSpecialAbilities {
			if mam.parent.SpecialAbilities.MagicWeapons {
				switch ad.DamageType {
				case core.DamageSlashing, core.DamagePiercing, core.DamageBludgeoning:
					// Check if it already has magic breaker to avoid duplicates
					hasMagic := false
					for _, b := range ad.ResistBreakers {
						if b == core.ResistBreakerMagic {
							hasMagic = true
							break
						}
					}
					if !hasMagic {
						ad.ResistBreakers = append(ad.ResistBreakers, core.ResistBreakerMagic)
					}
				}
			}
		}

		// Structured logging: action about to be executed
		if len(req.AttackData) == 1 {
			mam.parent.LogEvent(events.ECombatEventMessage, fmt.Sprintf("%s (%d) action '%s' against %s", mam.parent.GetName(), req.Target.GetInstanceID(), ad.Name, req.Target.GetName()))
		} else if idx == 0 {
			mam.parent.LogEvent(events.ECombatEventMessage, fmt.Sprintf("%s (%d) multiattack against %s", mam.parent.GetName(), req.Target.GetInstanceID(), req.Target.GetName()))
		}
		// Attack Roll
		attackMod := ad.AttackModifier
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

		// 2. Roll damage (without logging yet)
		rollOpts = roll_manager.NewRollOptions()
		rollOpts.Modifier = ad.DamageModifier
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
			AttackCount:    idx + 1,
			TargetValue:    attackRollResult.TargetValue,
			IsHit:          attackRollResult.IsSuccess,
			IsCriticalHit:  attackRollResult.IsCritical,
			AttackTotal:    attackRollResult.Total,
			AttackRoll:     attackRollResult.FinalRollValue,
			DamageRoll:     dmgRollResult,
			DamageType:     ad.DamageType,
			ResistBreakers: ad.ResistBreakers,
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

			// 5. Log Damage Roll manually (it will use the Attack as parent)
			mam.parent.LogEvent(events.ETRollEvent, dmgRollResult)

			results = append(results, attackResult)

			// 6. Restore Action ID for the next attack
			ctx.SetParentID(actionID)
		} else {
			// Fallback if no context
			// Log damage roll normally
			mam.parent.LogEvent(events.ETRollEvent, dmgRollResult)
			results = append(results, attackResult)
		}
	}

	return results, nil
}

// createAttackDataFromAction converts an Action into an AttackData object used for computations and attack management.
func (mam *MonsterActionManager) createAttackDataFromAction(action Action) core.AttackData {
	avg, err := core.GetAverageRoll(action.NumberOfDice, action.Die, action.AmountToAdd)
	if err != nil {
		fmt.Println("Error computing average roll")
		avg = -1
	}
	return core.AttackData{
		Name:              action.Name,
		NumberOfDice:      action.NumberOfDice,
		Die:               action.Die,
		AttackModifier:    action.AttackBonus,
		DamageModifier:    action.AmountToAdd,
		DamageType:        action.DamageType,
		IsVersatileAttack: false,
		Average:           avg,
	}
}

// precomputeAttackData precomputes attack data for actions, multiattacks, and legendary actions, optimizing runtime performance.
func (mam *MonsterActionManager) precomputeAttackData() {
	mam.ActionAttackData = make(map[int]core.AttackData)
	mam.MultiAttackData = make(map[int]MultiattackData)
	mam.LegendaryAttackData = make(map[int]core.AttackData)

	// Compute actions attack data
	for actionID, action := range mam.Actions {
		mam.ActionAttackData[actionID] = mam.createAttackDataFromAction(action)
	}

	// Compute mutltiattack actions attack data
	for optionIndex, multiattacks := range mam.Multiattacks {
		var maData MultiattackData
		var totalAverage int
		var totalNumberOfAttacks int
		var attackDataSlice []core.AttackData

		for _, ma := range multiattacks {
			attackActionData := mam.ActionAttackData[ma.ActionID]
			for i := 0; i < ma.Count; i++ {
				attackDataSlice = append(attackDataSlice, attackActionData)
				totalAverage += attackActionData.Average
				totalNumberOfAttacks += 1
			}
		}

		maData.TotalAverage = totalAverage
		if totalNumberOfAttacks > 0 {
			maData.AveragePerAttack = totalAverage / totalNumberOfAttacks
		} else {
			maData.AveragePerAttack = 0
		}
		maData.AttackDataBlocks = attackDataSlice
		mam.MultiAttackData[optionIndex] = maData
	}

	// Compute legendary actions attack data
	for i, la := range mam.LegendaryActions {
		mam.LegendaryAttackData[i] = mam.createAttackDataFromAction(la.Action)
	}
}

func (mam *MonsterActionManager) HasHealingAbilities() bool {
	// TODO: Implement this feature when the monster actions database is updated
	return false
}
