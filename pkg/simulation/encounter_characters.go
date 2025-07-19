package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/martial_attacks"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"fmt"
)

func (e *Encounter) AddPartyMember(c *character.Character) {
	e.Party = append(e.Party, c)
}

// handleCharacterTurn manages the actions of a character during their turn in an encounter.
func (e *Encounter) handleCharacterTurn(character *character.Character, advantage core.AdvantageType) {
	actionType, err := e.ChooseCharacterActionType(character)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch actionType {
	case core.ATHeal:
		e.performCharacterHealAction(character)
		return
	case core.ATRanged:
		e.performCharacterRangedAttack(character, advantage)
	case core.ATMelee:
		// TODO: Placeholder versatile bool
		e.performCharacterMeleeAttack(character, character.EntityModifiers.UseVersatileAttacks, advantage)
	case core.ATSpell:
		// TODO: Placeholder cast level
		e.performCharacterSpellAttack(character, 5, advantage)
		return
	case core.ATNoAction:
		fallthrough
	default:
		// No action to be taken
		return
	}
}

func (e *Encounter) performCharacterRangedAttack(c *character.Character, advantage core.AdvantageType) {
	target, err := e.chooseDamageTargetByPriority(c)
	if err != nil {
		fmt.Println(err)
		return
	}
	events.LogTargetChoiceEvent(c, target, c.GetEventListener())

	req, err := c.CreateAttackRequest(core.WSPrimary, false, advantage)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := martial_attacks.MakeMartialAttack(c, target, req, e.Options)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, atk := range res {
		if atk.IsHit {
			fmt.Printf("Hit! %d damage\n", atk.Damage)
		}
	}
}

func (e *Encounter) performCharacterMeleeAttack(c *character.Character, useVersatileAttack bool, advantage core.AdvantageType) {
	target, err := e.chooseDamageTargetByPriority(c)
	if err != nil {
		fmt.Println(err)
		return
	}
	events.LogTargetChoiceEvent(c, target, c.GetEventListener())
	// TODO: Add secondary slot
	// Perform attack using main slot

	req, err := c.CreateAttackRequest(core.WSPrimary, useVersatileAttack, advantage)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := martial_attacks.MakeMartialAttack(c, target, req, e.Options)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, atk := range res {
		if atk.IsHit {
			fmt.Printf("Hit! %d damage\n", atk.Damage)
		}
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

func (e *Encounter) performCharacterSpellAttack(c *character.Character, castLevel int, advantage core.AdvantageType) {
	if !e.Options.AOEHitsAllEnemies {
		target, _ := e.chooseDamageTargetByPriority(c)
		// TODO: how are we choosing spell preference
		damageSpell, err := e.chooseDamageSpell(c, core.SPHighestLevel)
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
