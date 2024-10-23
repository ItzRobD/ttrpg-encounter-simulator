package character

import (
	"dnd5e-encounter-simulator-backend/pkg/events"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/rolling"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
	"math"
)

func (c *Character) MakeSpellAttack(t *monster.Monster, s *spells.Spell) (bool, error) {
	// returns true if damage will be dealt
	if s.SpellType != "damage" {
		return false, fmt.Errorf("spell is not a damage spell")
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
				damage, rolls, err := rolling.DamageRoll(s.NumberOfDice, s.Die, s.AmountToAdd)
				if err != nil {
					return false, err
				}
				c.logSpellDCEvent(t, s, charDC, saveVal, true)

				damage = int(math.Floor(float64(damage) / 2))
				c.logDamageEvent(t, s.DamageType, damage, rolls)
				return true, nil
			case "none":
				c.logSpellDCEvent(t, s, charDC, saveVal, false)
				return false, nil
			}
		} else {
			// fail
			damage, rolls, err := rolling.DamageRoll(s.NumberOfDice, s.Die, s.AmountToAdd)
			if err != nil {
				return false, err
			}
			c.logSpellDCEvent(t, s, charDC, saveVal, true)

			c.logDamageEvent(t, s.DamageType, damage, rolls)
			return true, nil
		}
		return false, fmt.Errorf("spell has DC")
	} else {
		aMod, err := c.GetSpellBonus()
		if err != nil {
			return false, err
		}
		ar, err := rolling.AttackRoll(aMod)
		if err != nil {
			return false, err
		}

		didHit := rolling.AttackHits(ar+aMod, t.AC)
		c.logSpellAttackEvent(t, s, ar, aMod, didHit)

		if didHit {
			damageModifier, err2 := c.GetSpellBonus()
			if err2 != nil {
				return false, err2
			}

			dmg, rolls, err2 := rolling.DamageRoll(s.NumberOfDice, s.Die, damageModifier)
			if err2 != nil {
				return false, err2
			}

			c.logDamageEvent(t, s.DamageType, dmg, rolls)

			return true, nil
		} else { // Miss
			c.logSpellAttackEvent(t, s, ar, aMod, didHit)

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

	ar, err := rolling.AttackRoll(aMod)
	if err != nil {
		return false, err
	}

	attackValue := ar + aMod

	didHit := rolling.AttackHits(attackValue, t.AC)
	c.logWeaponAttackEvent(t, w, ar, aMod, didHit)

	if didHit {
		damageModifier, err := w.GetWeaponModifier(&c.AbilityScores)
		if err != nil {
			return false, err
		}

		damage, rolls, err := rolling.DamageRoll(w.NumberOfDice, w.Die, damageModifier)
		if err != nil {
			return false, err
		}

		c.logDamageEvent(t, w.DamageType, damage, rolls)

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
	c.HP.HP += amount
	if c.HP.HP > c.HP.MaxHP {
		c.HP.HP = c.HP.MaxHP
	}
	if c.HP.HP < 0 {
		c.HP.HP = 0
	}
}

func (c *Character) logWeaponAttackEvent(target *monster.Monster, weapon *weapon.Weapon, attackRoll, attackModifier int, isHit bool) {
	if c.EventListener != nil {
		event := events.CombatEvent{
			EventType: events.AttackEvent,
			Actor:     c.Name,
			Target:    target.Name,
			Attack:    weapon.Name,
			Value:     attackRoll + attackModifier,
			Rolls:     []int{attackRoll},
			Hit:       isHit,
		}
		c.EventListener(event)
	}
}

func (c *Character) logSpellAttackEvent(target *monster.Monster, spell *spells.Spell, attackRoll, attackModifier int, isHit bool) {
	if c.EventListener != nil {
		event := events.CombatEvent{
			EventType: events.AttackEvent,
			Actor:     c.Name,
			Target:    target.Name,
			Attack:    spell.Name,
			Value:     attackRoll + attackModifier,
			Rolls:     []int{attackRoll},
			Hit:       isHit,
		}
		c.EventListener(event)
	}
}

func (c *Character) logSpellDCEvent(target *monster.Monster, spell *spells.Spell, dc int, save int, isHit bool) {
	if c.EventListener != nil {
		event := events.CombatEvent{
			EventType:   events.SpellDC,
			Actor:       c.Name,
			Target:      target.Name,
			Attack:      spell.Name,
			Value:       dc,
			SavingThrow: save,
			Hit:         isHit,
		}
		c.EventListener(event)
	}
}

func (c *Character) logDamageEvent(target *monster.Monster, damageType string, damage int, rolls []int) {
	if c.EventListener != nil {
		event := events.CombatEvent{
			EventType:  events.DamageEvent,
			Actor:      c.Name,
			Target:     target.Name,
			Value:      damage,
			DamageType: damageType,
			Rolls:      rolls,
		}
		c.EventListener(event)
	}
}
