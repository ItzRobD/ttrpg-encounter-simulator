package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	"math/rand/v2"
)

type MonsterAIMode int

const (
	MAISimple MonsterAIMode = iota
	MAITactical
)

type MonsterAI struct {
	parent             *Monster
	combatCtx          *core.CombatContext
	rng                *rand.Rand
	isLegendary        bool
	hasRechargeActions bool
	hasMultiattack     bool
	hasSpecialAbility  bool
	AIMode             MonsterAIMode
	PrioritizeTotalAvg bool
}

func NewMonsterAI(m *Monster) *MonsterAI {
	hasRecharge := false
	hasMultiattack := false
	hasSpecialAbility := false

	if m.ActionManager.SpecialAbilities != nil {
		hasSpecialAbility = true
	}
	if m.ActionManager.Multiattacks != nil {
		hasMultiattack = true
	}
	if m.ActionManager.RechargeActions != nil {
		hasRecharge = true
	}

	return &MonsterAI{
		parent:             m,
		combatCtx:          nil,
		rng:                m.GetRNG(),
		isLegendary:        m.IsLegendary,
		hasRechargeActions: hasRecharge,
		hasMultiattack:     hasMultiattack,
		hasSpecialAbility:  hasSpecialAbility,
	}
}

func (mai *MonsterAI) UpdateCombatContext(ctx *core.CombatContext) {
	mai.combatCtx = ctx
}

func (mai *MonsterAI) createMonsterLegendaryActionRequest() (*core.AIRequest, error) {
	if !mai.isLegendary {
		return nil, fmt.Errorf("monster is not legendary")
	}

	if mai.parent.EntityState.LegendaryActionPoints == 0 {
		return nil, fmt.Errorf("monster has no legendary action points")
	}
	actionChoiceID := -1
	var err error
	legendaryIndexes := mai.getAvailableLegendaryActions(mai.parent.EntityState.LegendaryActionPoints)
	if len(legendaryIndexes) > 0 {
		actionChoiceID, err = mai.chooseLegendaryAction(legendaryIndexes)
	}

	if actionChoiceID != -1 {
		return mai.buildAIRequest(actionChoiceID, nil, core.ATLegendaryAction)
	} else {
		return nil, err
	}
}

func (mai *MonsterAI) getAvailableLegendaryActions(legPointsRemaining int) []int {
	availableIdx := make([]int, 0, len(mai.parent.ActionManager.LegendaryActions))
	for idx, la := range mai.parent.ActionManager.LegendaryActions {
		if la.Cost <= legPointsRemaining {
			availableIdx = append(availableIdx, idx)
		}
	}
	return availableIdx
}

func (mai *MonsterAI) chooseLegendaryAction(indexes []int) (int, error) {
	if len(indexes) == 0 {
		return -1, fmt.Errorf("no legendary actions available")
	}
	bestIndex := -1
	bestAvg := 0

	for _, idx := range indexes {
		if bestIndex == -1 {
			bestIndex = idx
			bestAvg = mai.parent.ActionManager.LegendaryAttackData[idx].Average
			continue
		}

		idxAvg := mai.parent.ActionManager.LegendaryAttackData[idx].Average
		if idxAvg > bestAvg {
			bestIndex = idx
			bestAvg = idxAvg
		}
	}

	if bestIndex == -1 {
		return -1, fmt.Errorf("no legendary actions available")
	}

	return bestIndex, nil
}

func (mai *MonsterAI) createMonsterHealActionRequest() (*core.AIRequest, error) {
	var req core.AIRequest
	var choice *core.SpellChoice

	targetID, err := mai.selectTargetID(core.TTHealing)
	if err != nil {
		return nil, err
	}

	// TODO: Will need to account for custom monsters with healing actions, maybe?

	// choose spell
	if mai.parent.SpellCastingManager.HasHealingSpells() {
		targetValue := mai.combatCtx.CombatantInfo[targetID].Combatant.Entity.GetHPStatus().GetHPDifference()
		choice, err = mai.parent.ChooseSpellByHealingEfficiency(targetValue)
		if err != nil {
			return nil, err
		}

		req = core.AIRequest{
			Actor:       mai.parent,
			ActorType:   core.EntityMonster,
			TargetID:    targetID,
			ActionType:  core.ATSpell,
			SpellChoice: choice,
		}

	} else {
		// Handle healing actions
	}

	return &req, nil
}

func (mai *MonsterAI) createMonsterDamageActionRequest() (*core.AIRequest, error) {
	var actionChoiceID int
	var spellChoice *core.SpellChoice
	var err error

	// Use recharge actions first
	if mai.hasRechargeActions {
		rechargeIndexes := mai.getAvailableRechargeActionIndexes()
		if len(rechargeIndexes) > 0 {
			actionChoiceID, err = mai.chooseRechargeAction(rechargeIndexes)
			if err != nil {
				fmt.Println(err)
			} else {
				return mai.buildAIRequest(actionChoiceID, nil, core.ATMonsterAction)
			}
		}
	}

	// Consider spellcasting vs multi attack
	if mai.hasMultiattack {
		switch mai.AIMode {
		case MAISimple:
			if mai.parent.IsSpellcaster() && mai.rng.IntN(2) == 0 {
				spellChoice, err = mai.chooseSpell()
				if err != nil {
					fmt.Println(err)
				} else {
					return mai.buildAIRequest(-1, spellChoice, core.ATSpell)
				}
			} else {
				actionChoiceID, err = mai.chooseMultiattackOption()
				if err != nil {
					fmt.Println(err)
				} else {
					return mai.buildAIRequest(actionChoiceID, nil, core.ATMonsterMultiattack)
				}
			}
			// The tactical AI is going to be replaced by the more advanced weighted decision process
		}
	}

	// No multiattack -> Spellcasting first
	if mai.parent.IsSpellcaster() {
		spellChoice, err = mai.chooseSpell()
		if err != nil {
			fmt.Println(err)
		} else {
			return mai.buildAIRequest(-1, spellChoice, core.ATSpell)
		}
	}

	actionChoiceID, err = mai.chooseMonsterAction()
	if err != nil {
		return nil, err
	} else {
		return mai.buildAIRequest(actionChoiceID, nil, core.ATMonsterAction)
	}
}

func (mai *MonsterAI) getAvailableRechargeActionIndexes() []int {
	availableIdx := make([]int, 0, len(mai.parent.ActionManager.RechargeActions))
	for idx, status := range mai.parent.EntityState.GetRechargeActionStatus() {
		if status {
			availableIdx = append(availableIdx, idx)
		}
	}

	return availableIdx
}

func (mai *MonsterAI) chooseRechargeAction(indexes []int) (int, error) {
	if len(indexes) == 0 {
		return -1, fmt.Errorf("no recharge actions available")
	}
	bestIndex := -1
	bestAvg := 0

	for _, idx := range indexes {
		if bestIndex == -1 {
			bestIndex = idx
			bestAvg = mai.parent.ActionManager.ActionAttackData[idx].Average
			continue
		}

		idxAvg := mai.parent.ActionManager.ActionAttackData[idx].Average
		if idxAvg > bestAvg {
			bestIndex = idx
			bestAvg = idxAvg
		}
	}

	if bestIndex == -1 {
		return -1, fmt.Errorf("no recharge actions available")
	}

	return bestIndex, nil
}

func (mai *MonsterAI) chooseMultiattackOption() (int, error) {
	if len(mai.parent.ActionManager.Multiattacks) == 0 {
		return -1, fmt.Errorf("no multiattack options available")
	}

	switch mai.AIMode {
	case MAISimple:
		idx := mai.rng.IntN(len(mai.parent.ActionManager.Multiattacks))
		return idx, nil
	case MAITactical:
		bestIndex := -1
		bestAvg := 0

		for idx, option := range mai.parent.ActionManager.MultiAttackData {
			if bestIndex == -1 {
				bestIndex = idx
				if mai.PrioritizeTotalAvg {
					bestAvg = option.TotalAverage
				} else {
					bestAvg = option.AveragePerAttack
				}
			}
			if mai.PrioritizeTotalAvg {
				if option.TotalAverage > bestAvg {
					bestIndex = idx
					bestAvg = option.TotalAverage
				}
			} else {
				if option.AveragePerAttack > bestAvg {
					bestIndex = idx
					bestAvg = option.AveragePerAttack
				}
			}
		}

		if bestIndex == -1 {
			return -1, fmt.Errorf("no multiattack options available")
		}

		return bestIndex, nil
	}

	return -1, fmt.Errorf("invalid monster AI mode")
}

func (mai *MonsterAI) chooseMonsterAction() (int, error) {
	actionIDs := mai.getActionIDs()
	if len(actionIDs) == 0 {
		return -1, fmt.Errorf("no actions available")
	}

	switch mai.AIMode {
	case MAISimple:
		idx := mai.rng.IntN(len(mai.parent.ActionManager.Actions))
		return actionIDs[idx], nil
	case MAITactical:
		bestIndex := -1
		bestAvg := 0

		for idx, action := range mai.parent.ActionManager.ActionAttackData {
			if bestIndex == -1 {
				bestIndex = actionIDs[idx]
				bestAvg = action.Average
			}
			if action.Average > bestAvg {
				bestIndex = actionIDs[idx]
				bestAvg = action.Average
			}
		}

		if bestIndex == -1 {
			return -1, fmt.Errorf("no actions available")
		}

		return bestIndex, nil
	}

	return -1, fmt.Errorf("invalid monster AI mode")
}

func (mai *MonsterAI) chooseSpell() (*core.SpellChoice, error) {
	if mai.parent.SpellCastingManager == nil || !mai.parent.IsSpellcaster() {
		return nil, fmt.Errorf("monster is not a spellcaster")
	}
	spellChoice, err := mai.parent.SpellCastingManager.ChooseSpellByPriority(core.STDamage, mai.parent.EntityState.SpellcastingPriority)
	if err != nil {
		return nil, err
	}

	return spellChoice, nil
}

func (mai *MonsterAI) buildAIRequest(actionIndex int, spellChoice *core.SpellChoice, actionType core.ActionType) (*core.AIRequest, error) {
	if mai.combatCtx == nil {
		return nil, fmt.Errorf("combat context not set")
	}
	if actionType == core.ATSpell && spellChoice == nil {
		return nil, fmt.Errorf("spell choice not set")
	}
	if actionType != core.ATSpell && actionIndex == -1 {
		return nil, fmt.Errorf("action index not set")
	}

	targetID, err := mai.selectTargetID(core.TTDamage)
	if err != nil {
		return nil, err
	}
	target := mai.combatCtx.CombatantInfo[targetID].Combatant.GetEntity()

	req := core.AIRequest{
		Actor:       mai.parent,
		ActorType:   core.EntityMonster,
		TargetID:    targetID,
		Target:      target,
		ActionType:  actionType,
		ActionIndex: actionIndex,
		SpellChoice: spellChoice,
	}

	// TODO: Logging

	return &req, nil
}

func (mai *MonsterAI) selectTargetID(targetType core.TargetType) (int, error) {
	if mai.combatCtx == nil {
		return -1, fmt.Errorf("combat context not set")
	}

	var validTargets map[int]*core.Combatant
	switch targetType {
	case core.TTDamage:
		validTargets = mai.getEnemyTargets()
	case core.TTHealing:
		validTargets = mai.getAllyTargets()
	default:
		return -1, fmt.Errorf("invalid target type")
	}

	target, err := core.SelectTargetFromMap(validTargets, mai.parent.EntityState.TargetPrioritization, mai.rng)
	if err != nil {
		return -1, err
	}

	return target, nil
}

func (mai *MonsterAI) getEnemyTargets() map[int]*core.Combatant {
	if mai.combatCtx == nil {
		return nil
	}

	enemies := make(map[int]*core.Combatant)
	self := mai.parent

	for id, combatant := range mai.combatCtx.CombatantInfo {
		e := combatant.Combatant.GetEntity()
		if !e.IsUnconscious() && (self.IsMonster() != e.IsMonster()) {
			enemies[id] = combatant.Combatant
		}
	}

	return enemies
}

func (mai *MonsterAI) getAllyTargets() map[int]*core.Combatant {
	allies := make(map[int]*core.Combatant)
	self := mai.parent

	for id, combatant := range mai.combatCtx.CombatantInfo {
		e := combatant.Combatant.GetEntity()
		if !e.IsUnconscious() && (self.IsMonster() == e.IsMonster()) {
			allies[id] = combatant.Combatant
		}
	}

	return allies
}

func (mai *MonsterAI) getActionIDs() []int {
	if mai.parent.ActionManager == nil {
		return nil
	}

	actionIDs := make([]int, 0, len(mai.parent.ActionManager.Actions))
	for idx := range mai.parent.ActionManager.Actions {
		actionIDs = append(actionIDs, idx)
	}
	return actionIDs
}

func (mai *MonsterAI) chooseMonsterActionType() (core.ActionType, error) {
	if mai.combatCtx == nil {
		return core.ATNoAction, fmt.Errorf("combat context not set")
	}

	if mai.combatCtx.Options.AllowMonsterHeals {
		if (mai.parent.SpellCastingManager.HasHealingSpells() || mai.parent.ActionManager.HasHealingAbilities()) && len(mai.combatCtx.MonstersInNeedOfHealing) > 0 {
			return core.ATMonsterHeal, nil
		}
	}

	return core.ATMonsterDamage, nil
}
