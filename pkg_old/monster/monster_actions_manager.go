package monster

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/events"
	"dnd5e-encounter-simulator-backend/pkg_old/core/roll_manager"
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
				switch core.DamageType(ad.GetDamageType()) {
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
			// Multiattack message just occurred, advance scope so subsequent attacks are children of it.
			if ctx := mam.parent.GetCurrentEventContext(); ctx != nil {
				ctx.AdvanceScope()
			}
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

		// 2. Roll damage
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
			AttackCount:    idx + 1,
			TargetValue:    attackRollResult.TargetValue,
			IsHit:          attackRollResult.IsSuccess,
			IsCriticalHit:  attackRollResult.IsCritical,
			AttackTotal:    attackRollResult.Total,
			AttackRoll:     attackRollResult.FinalRollValue,
			DamageRoll:     dmgRollResult,
			DamageType:     core.DamageType(ad.GetDamageType()),
			ResistBreakers: ad.ResistBreakers,
			IsRanged:       ad.IsRangedWeapon,
			AdvantageUsed:  attackRollResult.Advantage,
		}

		// 3. Log Attack with embedded action detail (instead of separate EquipmentEvent children)
		mods := make([]string, 0)
		for _, rb := range ad.ResistBreakers {
			if rb != core.ResistBreakerNone {
				mods = append(mods, rb.String())
			}
		}
		// Prefer the first damage block for headline values; detailed per-component rolls are still logged below
		var numDice int
		var dieStr string
		var dmgTypeStr string
		var dmgBonus int
		if len(ad.DamageBlocks) > 0 {
			numDice = ad.DamageBlocks[0].NumberOfDice
			dieStr = ad.DamageBlocks[0].Die.String()
			dmgTypeStr = ad.DamageBlocks[0].DamageType.String()
			dmgBonus = ad.DamageBlocks[0].Modifier + ad.WeaponDamageBonus + req.AttackOptions.GetBonusToDamageRoll()
		}
		actionDetail := &events.ActionDetail{
			Name:         ad.Name,
			NumberOfDice: numDice,
			Die:          dieStr,
			DamageType:   dmgTypeStr,
			AttackBonus:  ad.WeaponAttackBonus + req.AttackOptions.GetBonusToAttackRoll(),
			DamageBonus:  dmgBonus,
			IsRanged:     ad.IsRangedWeapon,
			Properties:   nil,
			Modifiers:    mods,
		}

		// Directly log the attack event with detail
		events.LogMartialAttackEvent(mam.parent.GetCurrentEventContext(), mam.parent, &attackResult, actionDetail, mam.parent.GetEventListener())

		// 4. Advance Scope: The Attack becomes the parent for subsequent events (like damage)
		ctx := mam.parent.GetCurrentEventContext()
		if ctx != nil {
			actionID := ctx.GetParentID() // Store current action ID
			ctx.AdvanceScope()

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

// ProcessSingleAttack rolls and logs a single attack defined by the first (and only) entry in req.AttackData.
// It returns the AttackResult for immediate effect application by the engine.
func (mam *MonsterActionManager) ProcessSingleAttack(req *core.AttackRequest) (core.AttackResult, error) {
	if req == nil || len(req.AttackData) == 0 {
		return core.AttackResult{}, fmt.Errorf("ProcessSingleAttack: missing AttackData")
	}
	ad := req.AttackData[0]

	// Apply special abilities if enabled (e.g., magic weapons breaker)
	if req.SimulationOptions != nil && req.SimulationOptions.EnableSpecialAbilities {
		if mam.parent.SpecialAbilities.MagicWeapons {
			switch core.DamageType(ad.GetDamageType()) {
			case core.DamageSlashing, core.DamagePiercing, core.DamageBludgeoning:
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

	// Attack roll
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
		return core.AttackResult{}, err
	}

	// Damage roll (components included)
	rollOpts = roll_manager.NewRollOptions()
	rollOpts.RollType = core.DiceRollDamage
	dmgRollResult, err := mam.rollManager.RollDamage(req, 0, attackRollResult.IsCritical, rollOpts, false)
	if err != nil {
		return core.AttackResult{}, err
	}

	attackResult := core.AttackResult{
		ActorName:      core.FormatEntityName(mam.parent),
		TargetName:     core.FormatEntityName(req.Target),
		Target:         req.Target,
		AttackName:     ad.Name,
		AttackCount:    1,
		TargetValue:    attackRollResult.TargetValue,
		IsHit:          attackRollResult.IsSuccess,
		IsCriticalHit:  attackRollResult.IsCritical,
		AttackTotal:    attackRollResult.Total,
		AttackRoll:     attackRollResult.FinalRollValue,
		DamageRoll:     dmgRollResult,
		DamageType:     core.DamageType(ad.GetDamageType()),
		ResistBreakers: ad.ResistBreakers,
		IsRanged:       ad.IsRangedWeapon,
		AdvantageUsed:  attackRollResult.Advantage,
	}

	// Log attack with embedded detail
	mods := make([]string, 0)
	for _, rb := range ad.ResistBreakers {
		if rb != core.ResistBreakerNone {
			mods = append(mods, rb.String())
		}
	}
	var numDice int
	var dieStr string
	var dmgTypeStr string
	var dmgBonus int
	if len(ad.DamageBlocks) > 0 {
		numDice = ad.DamageBlocks[0].NumberOfDice
		dieStr = ad.DamageBlocks[0].Die.String()
		dmgTypeStr = ad.DamageBlocks[0].DamageType.String()
		dmgBonus = ad.DamageBlocks[0].Modifier + ad.WeaponDamageBonus + req.AttackOptions.GetBonusToDamageRoll()
	}
	actionDetail := &events.ActionDetail{
		Name:         ad.Name,
		NumberOfDice: numDice,
		Die:          dieStr,
		DamageType:   dmgTypeStr,
		AttackBonus:  ad.WeaponAttackBonus + req.AttackOptions.GetBonusToAttackRoll(),
		DamageBonus:  dmgBonus,
		IsRanged:     ad.IsRangedWeapon,
		Properties:   nil,
		Modifiers:    mods,
	}
	events.LogMartialAttackEvent(mam.parent.GetCurrentEventContext(), mam.parent, &attackResult, actionDetail, mam.parent.GetEventListener())

	// Scope under attack and log per-component damage rolls
	if ctx := mam.parent.GetCurrentEventContext(); ctx != nil {
		parent := ctx.GetParentID()
		ctx.AdvanceScope()
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
		ctx.SetParentID(parent)
	}

	return attackResult, nil
}

// createAttackDataFromAction converts an Action into an AttackData object used for computations and attack management.
func (mam *MonsterActionManager) createAttackDataFromAction(action Action) core.AttackData {
	totalAvg := 0
	for _, db := range action.DamageBlocks {
		avg, err := core.GetAverageRoll(db.NumberOfDice, db.Die, db.Modifier)
		if err != nil {
			fmt.Println("Error computing average roll")
		} else {
			totalAvg += avg
		}
	}

	return core.AttackData{
		Name:              action.Name,
		DamageBlocks:      action.DamageBlocks,
		AttackModifier:    action.AttackBonus,
		DamageModifier:    0, // Modifiers are in DamageBlocks
		IsVersatileAttack: false,
		Average:           totalAvg,
		WeaponAttackBonus: action.AttackBonus,
		WeaponDamageBonus: 0,
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
