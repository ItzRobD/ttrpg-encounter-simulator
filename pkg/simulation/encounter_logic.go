package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
)

// ChooseCharacterActionType selects the action type for a character, preferring healing if available and needed.
func (e *Encounter) ChooseCharacterActionType(actor *character.Character) (shared.ActionType, error) {
	// Choose between a damage action or a healing action
	if e.Options.AllowPlayerHeals && actor.SpellcastingManager.HasHealingSpells() {
		if e.doesAPlayerNeedHealing() {
			events.LogCharacterActionChoiceEvent(actor, shared.ATHeal, actor.EventListener)
			return shared.ATHeal, nil
		}
	}
	return e.chooseCharacterDamageAction(actor)
}

func (e *Encounter) chooseCharacterDamageAction(actor *character.Character) (shared.ActionType, error) {
	switch e.Options.ActionPreference {
	case shared.APNoPreference:
		if !actor.SpellcastingManager.HasDamageSpells() {
			return chooseFromNoSpellsPreference(actor)
		} else {
			return chooseFromHasSpellsPreference(actor)
		}
	case shared.APPreferMelee:
		events.LogCharacterActionChoiceEvent(actor, shared.ATMelee, actor.EventListener)
		return shared.ATMelee, nil
	case shared.APPreferRanged:
		events.LogCharacterActionChoiceEvent(actor, shared.ATRanged, actor.EventListener)
		return shared.ATRanged, nil
	case shared.APPreferSpells:
		if !actor.SpellcastingManager.HasDamageSpells() {
			return chooseFromNoSpellsPreference(actor)
		}
		// TODO: There is a logic issue here where if spells are preferred but there are no spell slots available it tries anyway
		events.LogCharacterActionChoiceEvent(actor, shared.ATSpell, actor.EventListener)
		return shared.ATSpell, nil
	default:
		return shared.ATNoAction, fmt.Errorf("unknown action preference %s", e.Options.ActionPreference)
	}
}

func chooseFromNoSpellsPreference(actor *character.Character) (shared.ActionType, error) {
	if actor.ActionPreference == shared.APNoPreference {
		// TODO: This is simply checking if the character class is a ranger which may not prefer ranged attacks
		// TODO: Add a prefer ranged preference?
		//if actor.PrefersRanged() {
		//	events.LogCharacterActionChoiceEvent(actor, shared.ATRanged, actor.EventListener)
		//	return shared.ATRanged, nil
		//}
		r, _, err := core.RollDice(1, 100)
		if err != nil {
			return shared.ATNoAction, nil
		}
		if r <= 50 {
			events.LogCharacterActionChoiceEvent(actor, shared.ATMelee, actor.EventListener)
			return shared.ATMelee, nil
		}
		events.LogCharacterActionChoiceEvent(actor, shared.ATRanged, actor.EventListener)
		return shared.ATRanged, nil
	}
	actionType := shared.GetActionFromPreference(actor.ActionPreference)
	events.LogCharacterActionChoiceEvent(actor, actionType, actor.EventListener)
	return actionType, nil

}

func chooseFromHasSpellsPreference(actor *character.Character) (shared.ActionType, error) {
	if actor.ActionPreference == shared.APNoPreference {
		if actor.PrefersRanged() {
			events.LogCharacterActionChoiceEvent(actor, shared.ATRanged, actor.EventListener)
			return shared.ATRanged, nil
		} else if actor.PrefersSpells() {
			events.LogCharacterActionChoiceEvent(actor, shared.ATSpell, actor.EventListener)
			return shared.ATSpell, nil
		}
		r, _, err := core.RollDice(1, 100)
		if err != nil {
			return shared.ATNoAction, nil
		}
		if r > 66 {
			events.LogCharacterActionChoiceEvent(actor, shared.ATSpell, actor.EventListener)
			return shared.ATSpell, nil
		} else if r > 33 {
			events.LogCharacterActionChoiceEvent(actor, shared.ATMelee, actor.EventListener)
			return shared.ATMelee, nil
		}
		events.LogCharacterActionChoiceEvent(actor, shared.ATRanged, actor.EventListener)
		return shared.ATRanged, nil
	}
	actionType := shared.GetActionFromPreference(actor.ActionPreference)
	events.LogCharacterActionChoiceEvent(actor, actionType, actor.EventListener)
	return actionType, nil
}

func (e *Encounter) chooseBestHealingSpell(actor core.Entity, target core.Entity) (*spellcasting_manager.SpellChoice, error) {
	switch a := actor.(type) {
	case *character.Character:
		hpDiff := target.GetMaxHP() - target.GetCurrentHP()
		s, err := a.SpellcastingManager.GetMostEfficientHealingSpell(hpDiff)
		event := events.SpellChoiceEvent{
			SpellChoice:   s,
			ManagerStatus: a.SpellcastingManager.GetStatus(),
		}
		events.LogSpellChoiceEvent(a, event, a.EventListener)
		if err != nil {
			return nil, err
		}
		return s, nil
	}

	// TODO: Add monsters
	return nil, fmt.Errorf("unknown actor type %T", actor)
}

func (e *Encounter) chooseDamageSpell(actor core.Entity, priority shared.SpellPriority) (*spellcasting_manager.SpellChoice, error) {
	switch a := actor.(type) {
	case *character.Character:
		damageSpell, err := a.SpellcastingManager.ChooseSpellByPriority(spells.STDamage, priority)
		if err != nil {
			return nil, err
		}
		event := events.SpellChoiceEvent{
			SpellChoice:   damageSpell,
			ManagerStatus: a.SpellcastingManager.GetStatus(),
		}
		events.LogSpellChoiceEvent(a, event, a.EventListener)
		return damageSpell, nil
	case *monster.Monster:
		// TODO: Add monster spell choice, if spellcaster/if innate
		break
	}
	return nil, fmt.Errorf("unknown actor type %T", actor)
}

func (e *Encounter) doesAPlayerNeedHealing() bool {
	if !e.Options.AllowPlayerHeals {
		return false
	}
	for _, c := range e.filterCharacters() {
		if c.GetCurrentHPPct() < e.Options.PlayerHealThresholdPct {
			return true
		}
	}
	return false
}

func (e *Encounter) getLowestHPCharacter() *character.Character {
	var lowHPCharacter *character.Character
	for _, c := range e.filterCharacters() {
		if lowHPCharacter == nil ||
			(c.HP.HP/c.HP.MaxHP) < (lowHPCharacter.HP.HP/lowHPCharacter.HP.MaxHP) {
			lowHPCharacter = c
		}
	}
	return lowHPCharacter
}
