package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/races"
	"fmt"
	"math/rand/v2"
)

type CharacterAI struct {
	parent    *Character
	combatCtx *core.CombatContext
	rng       *rand.Rand
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

func (cai *CharacterAI) chooseDamageSpell() (*core.SpellChoice, error) {
	return cai.parent.SpellCastingManager.ChooseSpellByPriority(core.STDamage, cai.parent.EntityStateManager.GetSpellcastingPriority())
}

func (cai *CharacterAI) chooseDamageActionType() (core.ActionType, error) {
	actionPref := cai.parent.EntityStateManager.GetActionPreference()
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
		// When no preference is set, we need a deterministic choice.
		// For now, let's prefer Spells if available, then Melee, then Ranged.
		if cai.parent.IsSpellcaster() && cai.parent.SpellCastingManager.GetDamageSpellCount() > 0 {
			actionType = core.ATSpell
		} else if cai.parent.EquipmentManager.HasMeleeWeapon() {
			actionType = core.ATMelee
		} else if cai.parent.EquipmentManager.HasRangedWeapon() {
			actionType = core.ATRanged
		} else {
			actionType = core.ATUnarmed
		}
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

	// Structured logging: chosen damage action type
	events.LogCharacterActionChoiceEvent(cai.parent, actionType, cai.parent.GetEventListener())
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

func (cai *CharacterAI) selectTargetID(targetType core.TargetType) (core.TargetStatus, int, error) {
	if cai.combatCtx == nil {
		return core.TargetInvalidType, -1, fmt.Errorf("combat context not set")
	}

	var validTargets map[int]*core.Combatant
	switch targetType {
	case core.TTDamage:
		validTargets = cai.getEnemyTargets()
	case core.TTHealing:
		allies := cai.getAllyTargets()
		validTargets = make(map[int]*core.Combatant)
		needHealing := cai.combatCtx.CharactersInNeedOfHealing
		for _, id := range needHealing {
			if c, ok := allies[id]; ok {
				validTargets[id] = c
			}
		}
	default:
		return core.TargetInvalidType, -1, fmt.Errorf("unknown target type")
	}

	status, target, err := core.SelectTargetFromMap(validTargets, cai.parent.EntityStateManager.GetTargetPrioritization(), cai.rng)
	if err != nil || status != core.TargetOK {
		return status, -1, err
	}
	// Structured logging: chosen target
	if combatant, ok := validTargets[target]; ok && combatant != nil {
		events.LogTargetChoiceEvent(cai.parent, combatant.GetEntity(), cai.parent.GetEventListener())
	}
	return core.TargetOK, target, nil
}

func (cai *CharacterAI) getEnemyTargets() map[int]*core.Combatant {
	enemies := make(map[int]*core.Combatant)
	self := cai.parent

	for id, combatant := range cai.combatCtx.CombatantInfo {
		// Skip lair combatant entries
		if combatant.Combatant.IsLair {
			continue
		}
		e := combatant.Combatant.GetEntity()
		if !e.IsUnconscious() && (self.IsCharacter() != e.IsCharacter()) {
			enemies[id] = combatant.Combatant
		}
	}

	return enemies
}

func (cai *CharacterAI) getAllyTargets() map[int]*core.Combatant {
	allies := make(map[int]*core.Combatant)
	self := cai.parent

	for id, combatant := range cai.combatCtx.CombatantInfo {
		e := combatant.Combatant.GetEntity()
		if !e.IsDead() && (self.IsCharacter() == e.IsCharacter()) {
			allies[id] = combatant.Combatant
		}
	}

	return allies
}

func (cai *CharacterAI) chooseCharacterActionType() (core.ActionType, error) {
	if cai.combatCtx == nil {
		return core.ATNoAction, fmt.Errorf("combat context not set")
	}

	if cai.parent.Race.ID == races.Dragonborn && !cai.parent.EntityStateManager.GetDBBreathWeaponUsed() {
		// Use breath weapon if there are targets (simple AI logic)
		tStatus, _, _ := cai.selectTargetID(core.TTDamage)
		if tStatus == core.TargetOK {
			return core.ATDragonbornBreathWeapon, nil
		}
	}

	if cai.combatCtx.Options.AllowCharacterHeals {
		if cai.parent.IsHealer() && len(cai.combatCtx.CharactersInNeedOfHealing) > 0 {
			return core.ATHeal, nil
		}
	}

	return core.ATDamage, nil
}

func (cai *CharacterAI) createCharacterHealActionRequest() (*core.AIRequest, error) {
	tStatus, targetID, err := cai.selectTargetID(core.TTHealing)
	if err != nil {
		return nil, err
	}
	if tStatus == core.TargetNone {
		events.LogCombatEventMessage(cai.parent, "No valid healing targets", cai.parent.GetEventListener())
		return nil, nil
	}

	target := cai.combatCtx.CombatantInfo[targetID].Combatant.Entity
	healReq, err := cai.parent.CreateHealRequest(target)
	if err != nil {
		return nil, err
	}

	return &core.AIRequest{
		Actor:       cai.parent,
		ActorType:   core.EntityCharacter,
		Target:      target,
		TargetID:    targetID,
		ActionType:  core.ATHeal,
		HealRequest: healReq,
	}, nil
}

func (cai *CharacterAI) createCharacterDamageActionRequest() (*core.AIRequest, error) {
	var req core.AIRequest
	var choice *core.SpellChoice
	var useVersatile bool
	var slot core.WeaponSlot

	tStatus, targetID, err := cai.selectTargetID(core.TTDamage)
	if err != nil {
		return nil, err
	}
	if tStatus == core.TargetNone {
		events.LogCombatEventMessage(cai.parent, "No valid targets", cai.parent.GetEventListener())
		return nil, nil
	}

	at, err := cai.chooseDamageActionType()
	if err != nil {
		return nil, err
	}

	switch at {
	case core.ATSpell:
		choice, err = cai.chooseDamageSpell()
		if err != nil {
			return nil, err
		}
	case core.ATMelee:
		slot = core.WSPrimary
		em := cai.parent.EquipmentManager
		hasShield := em.HasShieldEquipped

		preferVersatile := false
		switch cai.parent.EntityStateManager.GetVersatileWeaponPreference() {
		case core.VWPPreferVersatile:
			preferVersatile = true
		case core.VWPNoPreference:
			preferVersatile = cai.rng.IntN(2) == 1
		case core.VWPPreferNonVersatile:
			preferVersatile = false
		}

		if preferVersatile && !hasShield {
			primaryWeapon, wErr := em.GetWeaponFromSlot(core.WSPrimary)
			if wErr != nil {
				return nil, wErr
			}
			useVersatile = primaryWeapon.Properties.IsVersatile
		}
	case core.ATRanged:
		slot = core.WSPrimary
		if w, wErr := cai.parent.EquipmentManager.GetWeaponFromSlot(core.WSRanged); wErr == nil {
			if w.Properties.IsRanged {
				slot = core.WSRanged
			} else {
				// explicit fallback to primary; check capability if you want to enforce it
				slot = core.WSPrimary
			}
		} else {
			// Ranged slot missing: try primary (and optionally secondary) as fallbacks
			if pw, pwErr := cai.parent.EquipmentManager.GetWeaponFromSlot(core.WSPrimary); pwErr == nil {
				if pw.Properties.IsRanged {
					slot = core.WSPrimary
				} else {
					// Optional: try secondary for thrown/ranged; otherwise, surface a clear error
					if sw, swErr := cai.parent.EquipmentManager.GetWeaponFromSlot(core.WSSecondary); swErr == nil && sw.Properties.IsRanged {
						slot = core.WSSecondary
					} else {
						return nil, fmt.Errorf("no valid ranged weapon available in ranged/primary/secondary slots")
					}
				}
			} else {
				return nil, fmt.Errorf("no valid ranged weapon available (ranged slot missing; primary lookup failed: %v)", pwErr)
			}
		}
	}

	req = core.AIRequest{
		Actor:        cai.parent,
		ActorType:    core.EntityCharacter,
		TargetID:     targetID,
		Target:       cai.combatCtx.CombatantInfo[targetID].Combatant.GetEntity(),
		ActionType:   at,
		WeaponSlot:   slot,
		UseVersatile: useVersatile,
		SpellChoice:  choice,
	}

	return &req, nil
}

func (cai *CharacterAI) createCharacterOffhandActionRequest() (*core.AIRequest, error) {
	if cai.combatCtx == nil {
		return nil, fmt.Errorf("combat context not set")
	}

	em := cai.parent.EquipmentManager

	// Check conditions, quietly fail
	if cai.parent.EntityStateManager.GetHasUsedBonusAction() {
		return nil, nil // No bonus action available
	}
	if em.HasShieldEquipped {
		return nil, nil // Shield blocks offhand attacks
	}
	if _, err := em.GetWeaponFromSlot(core.WSSecondary); err != nil {
		return nil, nil // No offhand
	}

	tStatus, targetID, err := cai.selectTargetID(core.TTDamage)
	if err != nil || tStatus != core.TargetOK {
		return nil, nil
	}

	req := &core.AIRequest{
		Actor:      cai.parent,
		ActorType:  core.EntityCharacter,
		TargetID:   targetID,
		Target:     cai.combatCtx.CombatantInfo[targetID].Combatant.GetEntity(),
		ActionType: core.ATOffhand,
		WeaponSlot: core.WSSecondary,
	}

	return req, nil
}

func (cai *CharacterAI) createDragonbornBreathWeaponRequest() (*core.AIRequest, error) {
	tStatus, targetID, err := cai.selectTargetID(core.TTDamage)
	if err != nil {
		return nil, err
	}
	if tStatus == core.TargetNone {
		events.LogCombatEventMessage(cai.parent, "No valid targets for breath weapon", cai.parent.GetEventListener())
		return nil, nil
	}

	req := &core.AIRequest{
		Actor:      cai.parent,
		ActorType:  core.EntityCharacter,
		TargetID:   targetID,
		Target:     cai.combatCtx.CombatantInfo[targetID].Combatant.GetEntity(),
		ActionType: core.ATDragonbornBreathWeapon,
		Request:    core.AIReqDragonbornBreathWeapon,
	}

	return req, nil
}
