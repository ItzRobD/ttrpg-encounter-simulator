package simulation_bk

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
)

// ChooseCharacterActionType selects the action type for a character, preferring healing if available and needed.
func (e *Encounter) ChooseCharacterActionType(actor *character.Character) (core.ActionType, error) {
	// Choose between a damage action or a healing action
	if e.Options.AllowCharacterHeals && actor.SpellCastingManager.HasHealingSpells() {
		if e.doesAPlayerNeedHealing() {
			events.LogCharacterActionChoiceEvent(actor, core.ATHeal, actor.EventListener)
			return core.ATHeal, nil
		}
	}
	return e.chooseCharacterDamageAction(actor)
}

func (e *Encounter) chooseCharacterDamageAction(actor *character.Character) (core.ActionType, error) {
	switch e.Options.ActionPreference {
	case core.APNoPreference:
		if !actor.SpellCastingManager.HasDamageSpells() {
			return chooseFromNoSpellsPreference(actor)
		} else {
			return chooseFromHasSpellsPreference(actor)
		}
	case core.APPreferMelee:
		events.LogCharacterActionChoiceEvent(actor, core.ATMelee, actor.EventListener)
		return core.ATMelee, nil
	case core.APPreferRanged:
		events.LogCharacterActionChoiceEvent(actor, core.ATRanged, actor.EventListener)
		return core.ATRanged, nil
	case core.APPreferSpells:
		if !actor.SpellCastingManager.HasDamageSpells() {
			return chooseFromNoSpellsPreference(actor)
		}
		// TODO: There is a logic issue here where if spells are preferred but there are no spell slots available it tries anyway
		events.LogCharacterActionChoiceEvent(actor, core.ATSpell, actor.EventListener)
		return core.ATSpell, nil
	default:
		return core.ATNoAction, fmt.Errorf("unknown action preference %s", e.Options.ActionPreference)
	}
}

func chooseFromNoSpellsPreference(actor *character.Character) (core.ActionType, error) {
	if actor.ActionPreference == core.APNoPreference {
		// TODO: This is simply checking if the character class is a ranger which may not prefer ranged attacks
		// TODO: Add a prefer ranged preference?
		//if actor.PrefersRanged() {
		//	events.LogCharacterActionChoiceEvent(actor, shared.ATRanged, actor.EventListener)
		//	return shared.ATRanged, nil
		//}
		r, _, err := core.RollDice(1, 100)
		if err != nil {
			return core.ATNoAction, nil
		}
		if r <= 50 {
			events.LogCharacterActionChoiceEvent(actor, core.ATMelee, actor.EventListener)
			return core.ATMelee, nil
		}
		events.LogCharacterActionChoiceEvent(actor, core.ATRanged, actor.EventListener)
		return core.ATRanged, nil
	}
	actionType := core.GetActionFromPreference(actor.ActionPreference)
	events.LogCharacterActionChoiceEvent(actor, actionType, actor.EventListener)
	return actionType, nil

}

func chooseFromHasSpellsPreference(actor *character.Character) (core.ActionType, error) {
	if actor.ActionPreference == core.APNoPreference {
		if actor.PrefersRanged() {
			events.LogCharacterActionChoiceEvent(actor, core.ATRanged, actor.EventListener)
			return core.ATRanged, nil
		} else if actor.PrefersSpells() {
			events.LogCharacterActionChoiceEvent(actor, core.ATSpell, actor.EventListener)
			return core.ATSpell, nil
		}
		r, _, err := core.RollDice(1, 100)
		if err != nil {
			return core.ATNoAction, nil
		}
		if r > 66 {
			events.LogCharacterActionChoiceEvent(actor, core.ATSpell, actor.EventListener)
			return core.ATSpell, nil
		} else if r > 33 {
			events.LogCharacterActionChoiceEvent(actor, core.ATMelee, actor.EventListener)
			return core.ATMelee, nil
		}
		events.LogCharacterActionChoiceEvent(actor, core.ATRanged, actor.EventListener)
		return core.ATRanged, nil
	}
	actionType := core.GetActionFromPreference(actor.ActionPreference)
	events.LogCharacterActionChoiceEvent(actor, actionType, actor.EventListener)
	return actionType, nil
}

func (e *Encounter) chooseBestHealingSpell(actor core.Entity, target core.Entity) (*spellcasting_manager.SpellChoice, error) {
	switch a := actor.(type) {
	case *character.Character:
		hpDiff := target.GetMaxHP() - target.GetCurrentHP()
		s, err := a.SpellCastingManager.GetMostEfficientHealingSpell(hpDiff)
		event := events.SpellChoiceEvent{
			SpellChoice:   s,
			ManagerStatus: a.SpellCastingManager.GetStatus(),
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

func (e *Encounter) chooseDamageSpell(actor core.Entity, priority core.SpellPriority) (*spellcasting_manager.SpellChoice, error) {
	switch a := actor.(type) {
	case *character.Character:
		damageSpell, err := a.SpellCastingManager.ChooseSpellByPriority(spells.STDamage, priority)
		if err != nil {
			return nil, err
		}
		event := events.SpellChoiceEvent{
			SpellChoice:   damageSpell,
			ManagerStatus: a.SpellCastingManager.GetStatus(),
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
	if !e.Options.AllowCharacterHeals {
		return false
	}
	for _, c := range e.filterCharacters() {
		if c.GetCurrentHPPct() < e.Options.CharacterHealThresholdPct {
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
