package monster_action_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
)

type MonsterActionManager struct {
	parent      *core.Entity
	rollManager *roll_manager.RollManager

	// Raw
	Actions          map[int]Action
	Multiattacks     map[int][]Multiattack
	LegendaryActions []LegendaryAction
	SpecialAbilities []SpecialAbility

	// Precomputed
	ActionAttackData    map[int]core.AttackData   // actionID -> AttackData
	MultiAttackData     map[int][]core.AttackData // option index -> complete action
	LegendaryAttackData map[int]core.AttackData   // Legendary action ID -> Attack Data
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

func (mam *MonsterActionManager) GetSpecialAbilities() []SpecialAbility {
	return mam.SpecialAbilities
}

// InitializeActions sets up actions, multiattacks, legendary actions, and special abilities from the provided configuration.
func (mam *MonsterActionManager) InitializeActions(config *MAMConfig) {
	mam.Actions = config.Actions
	mam.Multiattacks = config.Multiattacks
	mam.LegendaryActions = config.LegendaryActions
	mam.SpecialAbilities = config.SpecialAbilities

	mam.precomputeAttackData()
}

// NewMonsterActionManager initializes and returns a MonsterActionManager with the provided parent, roll manager, and configuration.
// If the provided config is nil, an empty configuration is used for initialization. Actions must be initialized later.
func NewMonsterActionManager(parent *core.Entity, rm *roll_manager.RollManager, config *MAMConfig) *MonsterActionManager {
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
	return core.AttackData{
		Name:              action.Name,
		NumberOfDice:      action.NumberOfDice,
		Die:               action.Die,
		AttackModifier:    action.AttackBonus,
		DamageModifier:    action.AmountToAdd,
		DamageType:        action.DamageType,
		IsVersatileAttack: false,
	}
}

// precomputeAttackData precomputes attack data for actions, multiattacks, and legendary actions, optimizing runtime performance.
func (mam *MonsterActionManager) precomputeAttackData() {
	mam.ActionAttackData = make(map[int]core.AttackData)
	mam.MultiAttackData = make(map[int][]core.AttackData)
	mam.LegendaryAttackData = make(map[int]core.AttackData)

	// Compute actions attack data
	for actionID, action := range mam.Actions {
		mam.ActionAttackData[actionID] = mam.createAttackDataFromAction(action)
	}

	// Compute mutltiattack actions attack data
	for optionIndex, multiattacks := range mam.Multiattacks {
		var attackDataSlice []core.AttackData

		for _, ma := range multiattacks {
			attackActionData := mam.ActionAttackData[ma.ActionID]
			for i := 0; i < ma.Count; i++ {
				attackDataSlice = append(attackDataSlice, attackActionData)
			}
		}

		mam.MultiAttackData[optionIndex] = attackDataSlice
	}

	// Compute legendary actions attack data
	for i, la := range mam.LegendaryActions {
		mam.LegendaryAttackData[i] = mam.createAttackDataFromAction(la.Action)
	}
}
