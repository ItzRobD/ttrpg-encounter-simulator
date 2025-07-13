package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
)

// DEPRECATED
//func (c *Character) CastSpell(target core.Entity, spell *spells.Spell, castLevel int) (*spells.SpellResult, error) {
//	// Validate spell
//	// TODO: Check if spell slots exist
//	if spell.Level > 0 {
//		if err := c.ExpendSpellSlot(castLevel); err != nil {
//			return nil, err
//		}
//	}
//
//	switch spell.SpellType {
//	case string(spells.STHealing):
//		return c.castHealingSpell(target, spell, castLevel)
//	case string(spells.STDamage):
//		//return c.castDamageSpell(target, spell, ctx)
//		fallthrough
//	default:
//		return nil, fmt.Errorf("unknown spell type: %s", spell.SpellType)
//	}
//}

//DEPRECATED
//func (c *Character) castHealingSpell(target core.Entity, spell *spells.Spell, castLevel int) (*spells.SpellResult, error) {
//	formula, err := spell.GetClosestFormulaToLevel(castLevel)
//	if err != nil {
//		return nil, err
//	}
//
//	amount, rolls, err := shared.DiceRollWithModifier(formula.NumberOfDice, formula.Die, formula.AmountToAdd)
//	if err != nil {
//		return nil, err
//	}
//
//	initialHP := target.GetCurrentHP()
//	target.ModifyHP(amount)
//	actualHealing := target.GetCurrentHP() - initialHP
//
//	events.LogHealEvent(c, target, actualHealing, rolls, c.EventListener)
//
//	return &spells.SpellResult{
//		Success: true,
//		Amount:  actualHealing,
//		Rolls:   rolls,
//	}, nil
//}

// DEPRECATED
//func (c *Character) MakeSpellHeal(t core.Entity, s *spells.Spell) (bool, error) {
//	if s.SpellType != "healing" {
//		return false, fmt.Errorf("spell is not a heal spell")
//	}
//
//	// This is done first. This may cause an issue in the future
//	if s.Level != 0 {
//		err := c.ExpendSpellSlot(s.Level)
//		if err != nil {
//			return false, err
//		}
//	}
//
//	sum, rolls, err := shared.DiceRollWithModifier(s.NumberOfDice, s.Die, s.AmountToAdd)
//	if err != nil {
//		return false, err
//	}
//
//	t.ModifyHP(sum)
//
//	events.LogHealEvent(c, t, sum, rolls, c.EventListener)
//	return true, nil
//}

//DEPRECATED
//func (c *Character) MakeSpellAttack(t *monster.Monster, s *spells.Spell, castLevel int, advantage shared.AdvantageType) (bool, error) {
//	// returns true if damage will be dealt
//	if s.SpellType != "damage" {
//		return false, fmt.Errorf("spell is not a damage spell")
//	}
//
//	// This is done first. This may cause an issue in the future
//	if s.Level != 0 {
//		err := c.ExpendSpellSlot(s.Level)
//		if err != nil {
//			return false, err
//		}
//	}
//
//	formula, err := s.GetClosestFormulaToLevel(castLevel)
//	if err != nil {
//		return false, err
//	}
//
//	if s.HasDC {
//		sBonus, err := c.GetSpellBonus()
//		if err != nil {
//			return false, err
//		}
//		pb, err := shared.GetCharacterProficiencyBonus(c.Level)
//		if err != nil {
//			return false, err
//		}
//		charDC := 8 + sBonus + pb
//		saveVal, err := t.GetSavingThrowRollResult(s.SpellDC.Ability)
//		if err != nil {
//			return false, err
//		}
//		if saveVal >= charDC {
//			// save
//			switch s.SpellDC.OnSuccess {
//			case "half":
//				damage, rolls, err := shared.DiceRollWithModifier(formula.NumberOfDice, formula.Die, formula.AmountToAdd)
//				if err != nil {
//					return false, err
//				}
//				//c.logSpellDCEvent(t, s, charDC, saveVal, true)
//				events.LogSpellDCEvent(c, t, s, charDC, saveVal, true, c.EventListener)
//
//				damage = int(math.Floor(float64(damage) / 2))
//				//c.logDamageEvent(t, s.DamageType, damage, rolls)
//				events.LogDamageEvent(c, t, formula.DamageType, damage, rolls, c.EventListener)
//				return true, nil
//			case "none":
//				//c.logSpellDCEvent(t, s, charDC, saveVal, false)
//				events.LogSpellDCEvent(c, t, s, charDC, saveVal, false, c.EventListener)
//				return false, nil
//			}
//		} else {
//			// fail
//			damage, rolls, err := shared.DiceRollWithModifier(formula.NumberOfDice, formula.Die, formula.AmountToAdd)
//			if err != nil {
//				return false, err
//			}
//			//c.logSpellDCEvent(t, s, charDC, saveVal, true)
//			events.LogSpellDCEvent(c, t, s, charDC, saveVal, true, c.EventListener)
//
//			//c.logDamageEvent(t, s.DamageType, damage, rolls)
//			events.LogDamageEvent(c, t, formula.DamageType, damage, rolls, c.EventListener)
//
//			return true, nil
//		}
//		return false, fmt.Errorf("spell has DC")
//	} else {
//		aMod, err := c.GetSpellBonus()
//		if err != nil {
//			return false, err
//		}
//		attackTotal, attackRoll, err := shared.AttackRoll(aMod, advantage)
//		if err != nil {
//			return false, err
//		}
//
//		didHit := shared.DoesAttackHit(attackTotal, t.AC)
//		//c.logSpellAttackEvent(t, s, ar, aMod, didHit)
//		events.LogSpellAttackEvent(c, t, s, attackRoll, aMod, didHit, c.EventListener)
//
//		if didHit {
//			damageModifier, err2 := c.GetSpellBonus()
//			if err2 != nil {
//				return false, err2
//			}
//
//			dmg, rolls, err2 := shared.DiceRollWithModifier(formula.NumberOfDice, formula.Die, damageModifier)
//			if err2 != nil {
//				return false, err2
//			}
//
//			//c.logDamageEvent(t, s.DamageType, dmg, rolls)
//			events.LogDamageEvent(c, t, formula.DamageType, dmg, rolls, c.EventListener)
//
//			return true, nil
//		} else { // Miss
//			//c.logSpellAttackEvent(t, s, ar, aMod, didHit)
//			events.LogSpellAttackEvent(c, t, s, attackRoll, aMod, didHit, c.EventListener)
//			return false, nil
//		}
//	}
//}

// DEPRECATED
//func (c *Character) MakeWeaponAttack(t *monster.Monster, slot shared.WeaponSlot, advantage shared.AdvantageType) (bool, int, error) {
//	// returns true if damage will be dealt, damage amount, error
//	w, err := c.getWeaponFromSlot(slot)
//	if err != nil {
//		return false, 0, err
//	}
//	wProf, err := c.GetWeaponProficiencyFromSlot(slot)
//	if err != nil {
//		return false, 0, err
//	}
//	attackModifier, err := w.GetAttackModifier(&c.AbilityScores, c.Level, wProf)
//	if err != nil {
//		return false, 0, err
//	}
//
//	attackTotal, attackRoll, err := shared.AttackRoll(attackModifier, advantage)
//	if err != nil {
//		return false, 0, err
//	}
//
//	didHit := shared.DoesAttackHit(attackTotal, t.AC)
//
//	events.LogMeleeAttackEvent(c, t, w.Name, attackRoll, attackModifier, didHit, c.EventListener)
//
//	if didHit {
//		damageModifier, err := w.GetWeaponModifier(&c.AbilityScores)
//		if err != nil {
//			return false, 0, err
//		}
//
//		damage, rolls, err := shared.DiceRollWithModifier(w.NumberOfDice, w.Die, damageModifier)
//		if err != nil {
//			return false, 0, err
//		}
//
//		events.LogDamageEvent(c, t, w.DamageType, damage, rolls, c.EventListener)
//
//		return true, damage, nil
//	}
//
//	return false, 0, nil
//}

func (c *Character) getWeaponFromSlot(slot shared.WeaponSlot) (*weapon.Weapon, error) {
	switch slot {
	case shared.WSPrimary:
		return &c.Eq.Primary, nil
	case shared.WSSecondary:
		return &c.Eq.Secondary, nil
	case shared.WSRanged:
		return &c.Eq.Ranged, nil
	default:
		return nil, fmt.Errorf("invalid slot identifier provided: %s", slot)
	}
}

func (c *Character) ModifyHP(value int) {
	prevHP := c.HP.HP
	c.HP.HP += value
	if c.HP.HP > c.HP.MaxHP {
		c.HP.HP = c.HP.MaxHP
	}
	if c.HP.HP < 0 {
		c.HP.HP = 0
	}
	events.LogHPModifiedEvent(c, value, prevHP, c.HP.HP, c.EventListener)
}

//func (c *Character) ExpendSpellSlot(level int) error {
//	if level == 0 || level > 9 {
//		return fmt.Errorf("invalid spell slot level provided: %d", level)
//	}
//	c.SpellSlots[level] -= 1
//	return nil
//}

//func (c *Character) logWeaponAttackEvent(target *monster.Monster, weapon *weapon.Weapon, attackRoll, attackModifier int, isHit bool) {
//	if c.EventListener != nil {
//		event := events.CombatEvent{
//			EventType: events.ETAttackEvent,
//			ActorName:     c.Name,
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
//			ActorName:        c.Name,
//			ActionChoice: choice,
//		}
//		c.EventListener(event)
//	}
//}

//func (c *Character) logSpellChoiceEvent(spell *spells.Spell) {
//	if c.EventListener != nil {
//		event := events.CombatEvent{
//			EventType:   events.ETSpellChoiceEvent,
//			ActorName:       c.Name,
//			SpellChoice: spell,
//		}
//		c.EventListener(event)
//	}
//}

//func (c *Character) logSpellAttackEvent(target *monster.Monster, spell *spells.Spell, attackRoll, attackModifier int, isHit bool) {
//	if c.EventListener != nil {
//		event := events.CombatEvent{
//			EventType: events.ETAttackEvent,
//			ActorName:     c.Name,
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
//			ActorName:       c.Name,
//			Target:      target.Name,
//			Attack:      spell.Name,
//			Value:       dc,
//			GetSavingThrowRollResult: save,
//			Success:         isHit,
//		}
//		c.EventListener(event)
//	}
//}

//func (c *Character) logDamageEvent(target *monster.Monster, damageType string, damage int, rolls []int) {
//	if c.EventListener != nil {
//		event := events.CombatEvent{
//			EventType:  events.ETDamageEvent,
//			ActorName:      c.Name,
//			Target:     target.Name,
//			Value:      damage,
//			DamageType: damageType,
//			Rolls:      rolls,
//		}
//		c.EventListener(event)
//	}
//}
