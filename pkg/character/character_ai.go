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
	if cai.combatCtx == nil {
		return -1, fmt.Errorf("combat context not set")
	}

	var validTargets map[int]*core.Combatant
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

func (cai *CharacterAI) getEnemyTargets() map[int]*core.Combatant {
	enemies := make(map[int]*core.Combatant)
	self := cai.parent

	for id, combatant := range cai.combatCtx.CombatantInfo {
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
		if !e.IsUnconscious() && (self.IsCharacter() == e.IsCharacter()) {
			allies[id] = combatant.Combatant
		}
	}

	return allies
}

func (cai *CharacterAI) chooseCharacterActionType() (core.ActionType, error) {
	if cai.combatCtx == nil {
		return core.ATNoAction, fmt.Errorf("combat context not set")
	}

	if cai.combatCtx.Options.AllowCharacterHeals {
		if cai.parent.IsHealer() && len(cai.combatCtx.CharactersInNeedOfHealing) > 0 {
			return core.ATHeal, nil
		}
	}

	return core.ATDamage, nil
}

func (cai *CharacterAI) createCharacterHealActionRequest() (*core.AIRequest, error) {
	var req core.AIRequest
	var choice *core.SpellChoice

	targetID, err := cai.selectTargetID(core.TTHealing)
	if err != nil {
		return nil, err
	}

	// choose spell
	targetValue := cai.combatCtx.CombatantInfo[targetID].Combatant.Entity.GetHPStatus().GetHPDifference()
	choice, err = cai.parent.ChooseSpellByHealingEfficiency(targetValue)
	if err != nil {
		return nil, err
	}

	req = core.AIRequest{
		Actor:       cai.parent,
		ActorType:   core.EntityCharacter,
		TargetID:    targetID,
		ActionType:  core.ATHeal,
		SpellChoice: choice,
	}

	return &req, nil
}

func (cai *CharacterAI) createCharacterDamageActionRequest() (*core.AIRequest, error) {
	var req core.AIRequest
	var choice *core.SpellChoice
	var useVersatile bool
	var slot core.WeaponSlot

	targetID, err := cai.selectTargetID(core.TTDamage)
	if err != nil {
		return nil, err
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
		switch cai.parent.EntityState.GetVersatileWeaponPreference() {
		case core.VWPPreferNonVersatile:
			useVersatile = false
			break
		case core.VWPNoPreference:
			if cai.rng.IntN(2) == 0 {
				useVersatile = false
				break
			}
			fallthrough
		case core.VWPPreferVersatile:
			if !cai.parent.EquipmentManager.HasShieldEquipped {
				w, wErr := cai.parent.EquipmentManager.GetWeaponFromSlot(core.WSPrimary)
				if wErr != nil {
					return nil, wErr
				}
				useVersatile = w.IsVersatile
			}
		}
	case core.ATRanged:
		slot = core.WSPrimary
		if w, wErr := cai.parent.EquipmentManager.GetWeaponFromSlot(core.WSRanged); wErr == nil {
			if w.IsRanged {
				slot = core.WSRanged
			} else {
				// explicit fallback to primary; check capability if you want to enforce it
				slot = core.WSPrimary
			}
		} else {
			// Ranged slot missing: try primary (and optionally secondary) as fallbacks
			if pw, pwErr := cai.parent.EquipmentManager.GetWeaponFromSlot(core.WSPrimary); pwErr == nil {
				if pw.IsRanged {
					slot = core.WSPrimary
				} else {
					// Optional: try secondary for thrown/ranged; otherwise, surface a clear error
					if sw, swErr := cai.parent.EquipmentManager.GetWeaponFromSlot(core.WSSecondary); swErr == nil && sw.IsRanged {
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
