package character

import (
	"dnd5e-encounter-simulator-backend/pkg/events"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/rolling"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
)

func (c *Character) MakeAttack(target *monster.Monster, slot string) (bool, error) {
	var w *weapon.Weapon
	switch slot {
	case "primary":
		w = &c.Eq.Primary
	case "secondary":
		w = &c.Eq.Secondary
	case "ranged":
		w = &c.Eq.Ranged
	default:
		return false, fmt.Errorf("invalid slot identifier provided: %s", slot)
	}

	aMod, err := w.GetAttackModifier(&c.AbilityScores, c.Level)
	if err != nil {
		return false, err
	}

	ar, err := rolling.AttackRoll(aMod)
	if err != nil {
		return false, err
	}

	if rolling.AttackHits(ar+aMod, target.AC) {
		if c.EventListener != nil {
			event := events.CombatEvent{
				EventType: events.AttackEvent,
				Actor:     c.Name,
				Target:    target.Name,
				Value:     ar + aMod,
				Rolls:     []int{ar},
				Hit:       true,
			}
			c.EventListener(event)
		}
		dMod, err2 := w.GetWeaponModifier(&c.AbilityScores)
		if err2 != nil {
			return false, err2
		}

		dmg, rolls, err2 := rolling.DamageRoll(w.NumberOfDice, w.Die, dMod)
		if err2 != nil {
			return false, err2
		}
		if c.EventListener != nil {
			event := events.CombatEvent{
				EventType: events.DamageEvent,
				Actor:     c.Name,
				Target:    target.Name,
				Value:     dmg,
				Rolls:     rolls,
			}
			c.EventListener(event)
		}
		return true, nil
	} else {
		// Handle a miss
		event := events.CombatEvent{
			EventType: events.AttackEvent,
			Actor:     c.Name,
			Target:    target.Name,
			Value:     ar + aMod,
			Rolls:     []int{ar},
			Hit:       false,
		}
		c.EventListener(event)
	}

	return false, nil
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
