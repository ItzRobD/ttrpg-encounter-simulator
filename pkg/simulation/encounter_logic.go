package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/rolling"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
)

func (e *Encounter) ChooseCharacterActionType(actor *character.Character) (shared.ActionType, error) {
	if e.Options.AllowPlayerHeals {
		if e.doesAPlayerNeedHealing() {
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
		return shared.ATMelee, nil
	case shared.APPreferRanged:
		return shared.ATRanged, nil
	case shared.APPreferSpells:
		if len(actor.KnownSpells) == 0 {
			return chooseFromNoSpellsPreference(actor)
		}
		return shared.ATSpell, nil
	default:
		return shared.ATNoAction, fmt.Errorf("unknown action preference %s", e.Options.ActionPreference)
	}
}

func chooseFromNoSpellsPreference(actor *character.Character) (shared.ActionType, error) {
	if actor.ActionPreference == shared.APNoPreference {
		if actor.PrefersRanged() {
			return shared.ATRanged, nil
		}
		r, _, err := rolling.RollDice(1, 100)
		if err != nil {
			return shared.ATNoAction, nil
		}
		if r < 50 {
			return shared.ATMelee, nil
		}
		return shared.ATRanged, nil
	}
	return shared.GetActionFromPreference(actor.ActionPreference), nil
}

func chooseFromHasSpellsPreference(actor *character.Character) (shared.ActionType, error) {
	if actor.ActionPreference == shared.APNoPreference {
		if actor.PrefersRanged() {
			return shared.ATRanged, nil
		} else if actor.PrefersSpells() {
			return shared.ATSpell, nil
		}
		r, _, err := rolling.RollDice(1, 100)
		if err != nil {
			return shared.ATNoAction, nil
		}
		if r > 66 {
			return shared.ATSpell, nil
		} else if r > 33 {
			return shared.ATMelee, nil
		}
		return shared.ATRanged, nil
	}
	return shared.GetActionFromPreference(actor.ActionPreference), nil
}

func (e *Encounter) doesAPlayerNeedHealing() bool {
	if !e.Options.AllowPlayerHeals {
		return false
	}
	for _, c := range e.filterCharacters() {
		hpPct := c.HP.HP / c.HP.MaxHP
		if hpPct < e.Options.PlayerHealThresholdPct {
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
