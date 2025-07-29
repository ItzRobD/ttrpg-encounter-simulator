package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"fmt"
)

type MonsterActionManager struct {
	parent      *Monster
	rollManager *roll_manager.RollManager

	// Raw
	Actions          map[int]Action        // Key: ActionID; Value: Action
	Multiattacks     map[int][]Multiattack // Key: Option Index; Value: Slice of ActionIDs and Count
	LegendaryActions []LegendaryAction
	SpecialAbilities []SpecialAbility
	RechargeActions  map[int]uint8

	// Precomputed
	ActionAttackData    map[int]core.AttackData // actionID -> AttackData
	MultiAttackData     map[int]MultiattackData // option index -> complete action
	LegendaryAttackData map[int]core.AttackData // Legendary action ID -> Attack Data
}

func (mam *MonsterActionManager) GetActions() map[int]Action {
	return mam.Actions
}

func (mam *MonsterActionManager) GetMulitattacks() map[int][]Multiattack {
	return mam.Multiattacks
}

func (mam *MonsterActionManager) GetLegendaryActions() []LegendaryAction {
	return mam.LegendaryActions
}

func (mam *MonsterActionManager) GetRechargeActionStatus() map[int]bool {
	return mam.parent.EntityState.RechargeActionStatus
}

func (mam *MonsterActionManager) ExpendRechargeAction(actionID int) {
	mam.parent.EntityState.ExpendRechargeAction(actionID)
}

func (mam *MonsterActionManager) RechargeAction(actionID int) {
	mam.parent.EntityState.RechargeRechargeAction(actionID)
}

func (mam *MonsterActionManager) GetSpecialAbilities() []SpecialAbility {
	return mam.SpecialAbilities
}

// InitializeActions sets up actions, multiattacks, legendary actions, and special abilities from the provided configuration.
func (mam *MonsterActionManager) InitializeActions(config *MAMConfig) {
	mam.Actions = config.Actions
	mam.Multiattacks = config.Multiattacks
	mam.LegendaryActions = config.LegendaryActions
	mam.SpecialAbilities = config.SpecialAbilities

	for _, action := range mam.Actions {
		if action.RechargeValue > 0 {
			mam.parent.EntityState.AddRechargeAction(action.ActionID)
			mam.RechargeActions[action.ActionID] = 1
		}
	}

	mam.precomputeAttackData()
}

// NewMonsterActionManager initializes and returns a MonsterActionManager with the provided parent, roll manager, and configuration.
// If the provided config is nil, an empty configuration is used for initialization. Actions must be initialized later.
func NewMonsterActionManager(parent *Monster, rm *roll_manager.RollManager, config *MAMConfig) *MonsterActionManager {
	if config == nil {
		return &MonsterActionManager{
			parent:      parent,
			rollManager: rm,
		}
	}
	mam := MonsterActionManager{
		parent:           parent,
		rollManager:      rm,
		Actions:          config.Actions,
		Multiattacks:     config.Multiattacks,
		LegendaryActions: config.LegendaryActions,
		SpecialAbilities: config.SpecialAbilities,
	}
	mam.precomputeAttackData()
	return &mam
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
		var attackDataSlice []core.AttackData

		for _, ma := range multiattacks {
			attackActionData := mam.ActionAttackData[ma.ActionID]
			for i := 0; i < ma.Count; i++ {
				attackDataSlice = append(attackDataSlice, attackActionData)
				totalAverage += attackActionData.Average
			}
		}

		maData.TotalAverage = totalAverage
		maData.AveragePerAttack = totalAverage / len(multiattacks)
		maData.AttackDataBlocks = attackDataSlice
		mam.MultiAttackData[optionIndex] = maData
	}

	// Compute legendary actions attack data
	for i, la := range mam.LegendaryActions {
		mam.LegendaryAttackData[i] = mam.createAttackDataFromAction(la.Action)
	}
}
