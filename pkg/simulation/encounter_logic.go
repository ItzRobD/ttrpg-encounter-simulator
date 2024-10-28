package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/simulation/events"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
)

func (e *Encounter) ChooseCharacterActionType(actor *character.Character) (shared.ActionType, error) {
	// Choose between a damage action or a healing action
	if e.Options.AllowPlayerHeals && actor.HasHealingSpells() {
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
		if len(actor.KnownSpells) == 0 {
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
		if len(actor.KnownSpells) == 0 {
			return chooseFromNoSpellsPreference(actor)
		}
		events.LogCharacterActionChoiceEvent(actor, shared.ATSpell, actor.EventListener)
		return shared.ATSpell, nil
	default:
		return shared.ATNoAction, fmt.Errorf("unknown action preference %s", e.Options.ActionPreference)
	}
}

func chooseFromNoSpellsPreference(actor *character.Character) (shared.ActionType, error) {
	if actor.ActionPreference == shared.APNoPreference {
		if actor.PrefersRanged() {
			events.LogCharacterActionChoiceEvent(actor, shared.ATRanged, actor.EventListener)
			return shared.ATRanged, nil
		}
		r, _, err := shared.RollDice(1, 100)
		if err != nil {
			return shared.ATNoAction, nil
		}
		if r < 50 {
			events.LogCharacterActionChoiceEvent(actor, shared.ATMelee, actor.EventListener)
			return shared.ATMelee, nil
		}
		events.LogCharacterActionChoiceEvent(actor, shared.ATRanged, actor.EventListener)
		return shared.ATRanged, nil
	}
	return shared.GetActionFromPreference(actor.ActionPreference), nil
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
		r, _, err := shared.RollDice(1, 100)
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
	return shared.GetActionFromPreference(actor.ActionPreference), nil
}

func (e *Encounter) chooseBestHealingSpell(actor shared.Entity, target shared.Entity) (*spells.Spell, error) {
	switch a := actor.(type) {
	case *character.Character:
		hpDiff := target.GetMaxHP() - target.GetCurrentHP()
		s, err := a.GetMostEfficientHealingSpell(hpDiff)
		events.LogSpellChoiceEvent(a, s, a.EventListener)
		if err != nil {
			return nil, err
		}
		return s, nil
	}

	// TODO: Add monsters
	return nil, fmt.Errorf("unknown actor type %T", actor)
}

func (e *Encounter) chooseDamageSpell(actor shared.Entity, priority shared.SpellPriority) (*spells.Spell, error) {
	switch a := actor.(type) {
	case *character.Character:
		damageSpell, err := a.ChooseDamageSpell(priority)
		if err != nil {
			return nil, err
		}
		events.LogSpellChoiceEvent(a, damageSpell, a.EventListener)
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
