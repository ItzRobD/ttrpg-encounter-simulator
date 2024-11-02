package character

import (
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/simulation/events"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
	"math"
)

func (c *Character) MakeSpellHeal(t shared.Entity, s *spells.Spell) (bool, error) {
	if s.SpellType != "healing" {
		return false, fmt.Errorf("spell is not a heal spell")
	}

	// This is done first. This may cause an issue in the future
	if s.Level != 0 {
		err := c.ExpendSpellSlot(s.Level)
		if err != nil {
			return false, err
		}
	}

	sum, rolls, err := shared.DiceRollWithModifier(s.NumberOfDice, s.Die, s.AmountToAdd)
	if err != nil {
		return false, err
	}

	t.ModifyHP(sum)

	events.LogHealEvent(c, t, sum, rolls, c.EventListener)
	return true, nil
}

func (c *Character) MakeSpellAttack(t *monster.Monster, s *spells.Spell) (bool, error) {
	// returns true if damage will be dealt
	if s.SpellType != "damage" {
		return false, fmt.Errorf("spell is not a damage spell")
	}

	// This is done first. This may cause an issue in the future
	if s.Level != 0 {
		err := c.ExpendSpellSlot(s.Level)
		if err != nil {
			return false, err
		}
	}

	if s.HasDC {
		sBonus, err := c.GetSpellBonus()
		if err != nil {
			return false, err
		}
		pb, err := shared.GetCharacterProficiencyBonus(c.Level)
		if err != nil {
			return false, err
		}
		charDC := 8 + sBonus + pb
		saveVal, err := t.SavingThrow(s.SpellDC.Ability)
		if err != nil {
			return false, err
		}
		if saveVal >= charDC {
			// save
			switch s.SpellDC.OnSuccess {
			case "half":
				damage, rolls, err := shared.DiceRollWithModifier(s.NumberOfDice, s.Die, s.AmountToAdd)
				if err != nil {
					return false, err
				}
				//c.logSpellDCEvent(t, s, charDC, saveVal, true)
				events.LogSpellDCEvent(c, t, s, charDC, saveVal, true, c.EventListener)

				damage = int(math.Floor(float64(damage) / 2))
				//c.logDamageEvent(t, s.DamageType, damage, rolls)
				events.LogDamageEvent(c, t, s.DamageType, damage, rolls, c.EventListener)
				return true, nil
			case "none":
				//c.logSpellDCEvent(t, s, charDC, saveVal, false)
				events.LogSpellDCEvent(c, t, s, charDC, saveVal, false, c.EventListener)
				return false, nil
			}
		} else {
			// fail
			damage, rolls, err := shared.DiceRollWithModifier(s.NumberOfDice, s.Die, s.AmountToAdd)
			if err != nil {
				return false, err
			}
			//c.logSpellDCEvent(t, s, charDC, saveVal, true)
			events.LogSpellDCEvent(c, t, s, charDC, saveVal, true, c.EventListener)

			//c.logDamageEvent(t, s.DamageType, damage, rolls)
			events.LogDamageEvent(c, t, s.DamageType, damage, rolls, c.EventListener)

			return true, nil
		}
		return false, fmt.Errorf("spell has DC")
	} else {
		aMod, err := c.GetSpellBonus()
		if err != nil {
			return false, err
		}
		ar, err := shared.AttackRoll(aMod)
		if err != nil {
			return false, err
		}

		didHit := shared.AttackHits(ar+aMod, t.AC)
		//c.logSpellAttackEvent(t, s, ar, aMod, didHit)
		events.LogSpellAttackEvent(c, t, s, ar, aMod, didHit, c.EventListener)

		if didHit {
			damageModifier, err2 := c.GetSpellBonus()
			if err2 != nil {
				return false, err2
			}

			dmg, rolls, err2 := shared.DiceRollWithModifier(s.NumberOfDice, s.Die, damageModifier)
			if err2 != nil {
				return false, err2
			}

			//c.logDamageEvent(t, s.DamageType, dmg, rolls)
			events.LogDamageEvent(c, t, s.DamageType, dmg, rolls, c.EventListener)

			return true, nil
		} else { // Miss
			//c.logSpellAttackEvent(t, s, ar, aMod, didHit)
			events.LogSpellAttackEvent(c, t, s, ar, aMod, didHit, c.EventListener)
			return false, nil
		}
	}
}

func (c *Character) MakeWeaponAttack(t *monster.Monster, slot string) (bool, error) {
	// returns true if damage will be dealt
	w, err := c.getWeaponFromSlot(slot)
	if err != nil {
		return false, err
	}

	aMod, err := w.GetAttackModifier(&c.AbilityScores, c.Level)
	if err != nil {
		return false, err
	}

	ar, err := shared.AttackRoll(aMod)
	if err != nil {
		return false, err
	}

	attackValue := ar + aMod

	didHit := shared.AttackHits(attackValue, t.AC)
	//c.logWeaponAttackEvent(t, w, ar, aMod, didHit)
	events.LogWeaponAttackEvent(c, t, w, ar, aMod, didHit, c.EventListener)

	if didHit {
		damageModifier, err := w.GetWeaponModifier(&c.AbilityScores)
		if err != nil {
			return false, err
		}

		damage, rolls, err := shared.DiceRollWithModifier(w.NumberOfDice, w.Die, damageModifier)
		if err != nil {
			return false, err
		}

		//c.logDamageEvent(t, w.DamageType, damage, rolls)
		events.LogDamageEvent(c, t, w.DamageType, damage, rolls, c.EventListener)

		return true, nil
	} else {
		return false, nil
	}
}

func (c *Character) getWeaponFromSlot(slot string) (*weapon.Weapon, error) {
	switch slot {
	case "primary":
		return &c.Eq.Primary, nil
	case "secondary":
		return &c.Eq.Secondary, nil
	case "ranged":
		return &c.Eq.Ranged, nil
	default:
		return nil, fmt.Errorf("invalid slot identifier provided: %s", slot)
	}
}

func (c *Character) ModifyHP(amount int) {
	prevHP := c.HP.HP
	c.HP.HP += amount
	if c.HP.HP > c.HP.MaxHP {
		c.HP.HP = c.HP.MaxHP
	}
	if c.HP.HP < 0 {
		c.HP.HP = 0
	}
	events.LogHPModifiedEvent(c, amount, prevHP, c.HP.HP, c.EventListener)
}

func (c *Character) ExpendSpellSlot(level int) error {
	if level == 0 || level > 9 {
		return fmt.Errorf("invalid spell slot level provided: %d", level)
	}
	c.SpellSlots[level] -= 1
	return nil
}

//func (c *Character) logWeaponAttackEvent(target *monster.Monster, weapon *weapon.Weapon, attackRoll, attackModifier int, isHit bool) {
//	if c.EventListener != nil {
//		event := events.CombatEvent{
//			EventType: events.ETAttackEvent,
//			Actor:     c.Name,
//			Target:    target.Name,
//			Attack:    weapon.Name,
//			Value:     attackRoll + attackModifier,
//			Rolls:     []int{attackRoll},
//			Success:       isHit,
//		}
//		c.EventListener(event)
//	}
//}

//func (c *Character) logActionChoiceEvent(choice shared.ActionType) {
//	if c.EventListener != nil {
//		event := events.CombatEvent{
//			EventType:    events.ETActionChoiceEvent,
//			Actor:        c.Name,
//			ActionChoice: choice,
//		}
//		c.EventListener(event)
//	}
//}

//func (c *Character) logSpellChoiceEvent(spell *spells.Spell) {
//	if c.EventListener != nil {
//		event := events.CombatEvent{
//			EventType:   events.ETSpellChoiceEvent,
//			Actor:       c.Name,
//			SpellChoice: spell,
//		}
//		c.EventListener(event)
//	}
//}

//func (c *Character) logSpellAttackEvent(target *monster.Monster, spell *spells.Spell, attackRoll, attackModifier int, isHit bool) {
//	if c.EventListener != nil {
//		event := events.CombatEvent{
//			EventType: events.ETAttackEvent,
//			Actor:     c.Name,
//			Target:    target.Name,
//			Attack:    spell.Name,
//			Value:     attackRoll + attackModifier,
//			Rolls:     []int{attackRoll},
//			Success:       isHit,
//		}
//		c.EventListener(event)
//	}
//}

//func (c *Character) logSpellDCEvent(target *monster.Monster, spell *spells.Spell, dc int, save int, isHit bool) {
//	if c.EventListener != nil {
//		event := events.CombatEvent{
//			EventType:   events.ETSpellDC,
//			Actor:       c.Name,
//			Target:      target.Name,
//			Attack:      spell.Name,
//			Value:       dc,
//			SavingThrow: save,
//			Success:         isHit,
//		}
//		c.EventListener(event)
//	}
//}

//func (c *Character) logDamageEvent(target *monster.Monster, damageType string, damage int, rolls []int) {
//	if c.EventListener != nil {
//		event := events.CombatEvent{
//			EventType:  events.ETDamageEvent,
//			Actor:      c.Name,
//			Target:     target.Name,
//			Value:      damage,
//			DamageType: damageType,
//			Rolls:      rolls,
//		}
//		c.EventListener(event)
//	}
//}
