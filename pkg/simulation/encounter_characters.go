package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core/combat"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
)

func (e *Encounter) AddPartyMember(c *character.Character) {
	e.Party = append(e.Party, c)
}

// handleCharacterTurn manages the actions of a character during their turn in an encounter.
func (e *Encounter) handleCharacterTurn(character *character.Character, advantage shared.AdvantageType) {
	actionType, err := e.ChooseCharacterActionType(character)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch actionType {
	case shared.ATHeal:
		e.performCharacterHealAction(character)
		return
	case shared.ATRanged:
		e.performCharacterRangedAttack(character, advantage)
	case shared.ATMelee:
		// TODO: Placeholder versatile bool
		e.performCharacterMeleeAttack(character, character.EntityModifiers.UseVersatileAttacks, advantage)
	case shared.ATSpell:
		// TODO: Placeholder cast level
		e.performCharacterSpellAttack(character, 5, advantage)
		return
	case shared.ATNoAction:
		fallthrough
	default:
		// No action to be taken
		return
	}
}

func (e *Encounter) performCharacterRangedAttack(character *character.Character, advantage shared.AdvantageType) {
	target, err := e.chooseDamageTargetByPriority(character)
	if err != nil {
		fmt.Println(err)
		return
	}
	events.LogTargetChoiceEvent(character, target, character.GetEventListener())

	aI, err := character.CreateWeaponAttackInfo(shared.WSRanged, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	didHit, dmg, err := combat.MakeMartialAttack(character, target, aI, advantage)
	if err != nil {
		fmt.Println(err)
	}

	if didHit {
		fmt.Printf("Hit! %d damage\n", dmg)
	}
}

func (e *Encounter) performCharacterMeleeAttack(character *character.Character, useVersatileAttack bool, advantage shared.AdvantageType) {
	target, err := e.chooseDamageTargetByPriority(character)
	if err != nil {
		fmt.Println(err)
		return
	}
	events.LogTargetChoiceEvent(character, target, character.GetEventListener())
	// TODO: Add secondary slot
	// Perform attack using main slot

	aI, err := character.CreateWeaponAttackInfo(shared.WSPrimary, useVersatileAttack)
	if err != nil {
		fmt.Println(err)
		return
	}

	didHit, dmg, err := combat.MakeMartialAttack(character, target, aI, advantage)
	if err != nil {
		fmt.Println(err)
		return
	}
	if didHit {
		fmt.Printf("Hit! %d damage\n", dmg)
	}

}

func (e *Encounter) performCharacterHealAction(c *character.Character) {
	target, err := e.chooseHealTargetByPriority(c)
	if err != nil {
		fmt.Println(err)
		return
	}
	events.LogTargetChoiceEvent(c, target, c.GetEventListener())

	healingSpell, err := e.chooseBestHealingSpell(c, target)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO: Handle has slots
	events.LogSpellChoiceEvent(c, healingSpell, false, c.GetEventListener())

	// TODO: Placeholder cast level
	_, err = c.CastSpell(target, healingSpell, 5)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (e *Encounter) performCharacterSpellAttack(c *character.Character, castLevel int, advantage shared.AdvantageType) {
	if !e.Options.AOEHitsAllEnemies {
		target, _ := e.chooseDamageTargetByPriority(c)
		damageSpell, err := e.chooseDamageSpell(c, shared.SPHighestLevel)
		if err != nil {
			fmt.Println(err)
		}
		if monsterTarget, ok := target.(*monster.Monster); ok {
			_, err2 := c.MakeSpellAttack(monsterTarget, damageSpell, castLevel, advantage)
			if err2 != nil {
				fmt.Println(err2)
			}
		}
	}
}
