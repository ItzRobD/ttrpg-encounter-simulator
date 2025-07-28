package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	"math/rand/v2"
)

type CharacterAI struct {
	parent    *Character
	combatCtx *core.CombatContext
	rng       *rand.Rand
	//simConfig *simulation.SimulationConfig
}

func NewCharacterAI(c *Character) *CharacterAI {
	return &CharacterAI{
		parent:    c,
		combatCtx: nil,
		rng:       c.GetRNG(),
	}
}
func (cai *CharacterAI) UpdateCombatContext(ctx *core.CombatContext) {
	cai.combatCtx = ctx
}

func (cai *CharacterAI) chooseCharacterAction() (*core.AIRequest, error) {
	if cai.combatCtx == nil {
		return nil, fmt.Errorf("combat context not set")
	}

	var req core.AIRequest
	var choice *core.SpellChoice
	if cai.combatCtx.AllowCharacterHeals && cai.parent.IsHealer() && len(cai.combatCtx.NeedHealingIDs) > 0 {
		// choose target
		targetID, err := cai.selectTargetID(core.TTHealing)
		if err != nil {
			return nil, err
		}
		// choose healing spell
		targetValue := cai.combatCtx.AllCombatants[targetID].GetEntity().GetHPStatus().GetHPDifference()
		choice, err = cai.parent.ChooseSpellByHealingEfficiency(targetValue)
		if err != nil {
			return nil, err
		}
		// return the action request
		req = core.AIRequest{
			TargetID:    targetID,
			Actor:       cai.parent,
			ActionType:  core.ATHeal,
			SpellChoice: choice,
		}

		// TODO: Logging

		return &req, nil
	}
	// Choose damage
	targetID, err := cai.selectTargetID(core.TTDamage)
	if err != nil {
		return nil, err
	}
	at, err := cai.chooseDamageActionType()
	if err != nil {
		return nil, err
	}
	if at == core.ATSpell {
		choice, err = cai.chooseDamageSpell()
		if err != nil {
			return nil, err
		}
	}
	req = core.AIRequest{
		TargetID:    targetID,
		Actor:       cai.parent,
		Target:      cai.combatCtx.AllCombatants[targetID].GetEntity(),
		ActionType:  at,
		SpellChoice: choice,
	}

	// TODO: Logging

	return &req, nil
}

func (cai *CharacterAI) chooseDamageSpell() (*core.SpellChoice, error) {
	return cai.parent.SpellCastingManager.ChooseSpellByPriority(core.STDamage, cai.parent.EntityState.SpellcastingPriority)
}

func (cai *CharacterAI) chooseDamageActionType() (core.ActionType, error) {
	actionPref := cai.parent.EntityState.ActionPreference
	actionType := core.ATNoAction
	switch actionPref {
	case core.APPreferMelee:
		if cai.parent.EquipmentManager.HasMeleeWeapon() {
			actionType = core.ATMelee
		} else {
			return cai.chooseFallbackAction(core.ATMelee), nil
		}
	case core.APPreferRanged:
		if cai.parent.EquipmentManager.HasRangedWeapon() {
			actionType = core.ATRanged
		} else {
			return cai.chooseFallbackAction(core.ATRanged), nil
		}
	case core.APNoPreference:
		fallthrough
	case core.APPreferSpells:
		if cai.parent.IsSpellcaster() && cai.parent.SpellCastingManager.GetDamageSpellCount() > 0 {
			actionType = core.ATSpell
		} else {
			if cai.parent.EquipmentManager.HasMeleeWeapon() {
				actionType = core.ATMelee
			} else {
				return cai.chooseFallbackAction(core.ATMelee), nil
			}
		}
	default:
		return core.ATNoAction, fmt.Errorf("unknown action preference %s", actionPref)
	}

	//TODO: Logging
	return actionType, nil
}

func (cai *CharacterAI) chooseFallbackAction(exclude core.ActionType) core.ActionType {
	// Standard fallback priority, excluding the preferred type
	if exclude != core.ATRanged && cai.parent.EquipmentManager.HasRangedWeapon() {
		return core.ATRanged
	}
	if exclude != core.ATMelee && cai.parent.EquipmentManager.HasMeleeWeapon() {
		return core.ATMelee
	}
	if exclude != core.ATSpell &&
		cai.parent.IsSpellcaster() &&
		cai.parent.SpellCastingManager.GetDamageSpellCount() > 0 {
		return core.ATSpell
	}
	return core.ATUnarmed
}

func (cai *CharacterAI) selectTargetID(targetType core.TargetType) (int, error) {
	var validTargets map[int]core.Combatant
	switch targetType {
	case core.TTDamage:
		validTargets = cai.getEnemyTargets()
	case core.TTHealing:
		validTargets = cai.getAllyTargets()
	default:
		return -1, fmt.Errorf("unknown target type")
	}

	target, err := core.SelectTargetFromMap(validTargets, cai.parent.EntityState.TargetPrioritization, cai.rng)
	if err != nil {
		return -1, err
	}
	return target, nil
}

func (cai *CharacterAI) getEnemyTargets() map[int]core.Combatant {
	enemies := make(map[int]core.Combatant)
	self := cai.parent

	for id, combatant := range cai.combatCtx.AllCombatants {
		e := combatant.GetEntity()
		if !e.IsUnconscious() && (self.IsCharacter() != e.IsCharacter()) {
			enemies[id] = combatant
		}
	}

	return enemies
}

func (cai *CharacterAI) getAllyTargets() map[int]core.Combatant {
	allies := make(map[int]core.Combatant)
	self := cai.parent

	for id, combatant := range cai.combatCtx.AllCombatants {
		e := combatant.GetEntity()
		if !e.IsUnconscious() && (self.IsCharacter() == e.IsCharacter()) {
			allies[id] = combatant
		}
	}

	return allies
}
