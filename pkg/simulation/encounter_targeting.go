package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
	"math/rand/v2"
)

func (e *Encounter) chooseDamageTarget(actor shared.Entity) (shared.Entity, error) {
	switch actor.(type) {
	case *character.Character:
		monsters := e.filterMonsters()
		if len(monsters) == 0 {
			return nil, fmt.Errorf("no targets available")
		}
		return e.selectMonsterByPriority(monsters)

	case *monster.Monster:
		characters := e.filterCharacters()
		if len(characters) == 0 {
			return nil, fmt.Errorf("no targets available")
		}
		return e.selectCharacterByPriority(characters)

	default:
		fmt.Printf("Unknown creature type %T\n", actor)
	}
	panic("unhandled actor type")
}

func (e *Encounter) chooseHealTarget(actor shared.Entity) (shared.Entity, error) {
	switch actor.(type) {
	case *character.Character:
		characters := e.filterCharacters()
		if len(characters) == 0 {
			return nil, fmt.Errorf("no targets available")
		}

		var lowestHP shared.Entity
		for _, c := range characters {
			if lowestHP == nil || c.GetCurrentHP() < lowestHP.GetCurrentHP() {
				lowestHP = c
			}
			if c.GetCurrentHP() == lowestHP.GetCurrentHP() {
				if len(c.Class.Spellcasting.ClassHealingSpells) > 0 {
					lowestHP = c
				}
			}
		}
		return lowestHP, nil
	}
	// TODO: Add monsters
	return nil, nil
}

func (e *Encounter) selectMonsterByPriority(monsters []*monster.Monster) (*monster.Monster, error) {
	var target *monster.Monster
	switch e.Options.TargetPriority {
	case shared.NoPriority:
		//fmt.Println("No TargetPriority")
		target = monsters[rand.IntN(len(monsters))]
	case shared.PrioritizeHighestCR:
		for _, m := range monsters {
			if target == nil || m.CR > target.CR {
				target = m
			}
		}
	case shared.PrioritizeLowestCR:
		for _, m := range monsters {
			if target == nil || m.CR < target.CR {
				target = m
			}
		}
	case shared.PrioritizeMostDamaged:
		for _, m := range monsters {
			if target == nil || m.GetMaxHP()-m.GetCurrentHP() > target.GetMaxHP()-target.GetCurrentHP() {
				target = m
			}
		}
	case shared.PrioritizeLowestHealth:
		for _, m := range monsters {
			if target == nil || m.GetCurrentHP() < target.GetCurrentHP() {
				target = m
			}
		}
	case shared.PrioritizeHighestMaxHP:
		for _, m := range monsters {
			if target == nil || m.HP.MaxHP > target.HP.MaxHP {
				target = m
			}
		}
	case shared.PrioritizeLowestMaxHP:
		for _, m := range monsters {
			if target == nil || m.HP.MaxHP < target.HP.MaxHP {
				target = m
			}
		}
	case shared.PrioritizeHealer:
		for _, m := range monsters {
			if m.IsSpellcaster {
				for _, s := range m.Spellcasting.SC.Spells {
					if s.SpellType == "Heal" {
						return m, nil
					}
				}
			}
		}
	case shared.PrioritizeSpellcasting:
		for _, m := range monsters {
			if m.IsSpellcaster || m.IsInnateSpellcaster {
				return m, nil
			}
		}
	default:
		panic("unhandled default case")
	}
	return target, nil
}

func (e *Encounter) selectCharacterByPriority(characters []*character.Character) (*character.Character, error) {
	var target *character.Character
	switch e.Options.TargetPriority {
	case shared.NoPriority:
		fmt.Println("No TargetPriority")
		target = characters[rand.IntN(len(characters))]
	case shared.PrioritizeHighestCR,
		shared.PrioritizeLowestCR:
		fallthrough
	case shared.PrioritizeMostDamaged:
		for _, c := range characters {
			if target == nil || c.GetMaxHP()-c.GetCurrentHP() > target.GetMaxHP()-target.GetCurrentHP() {
				target = c
			}
		}
	case shared.PrioritizeLowestHealth:
		for _, c := range characters {
			if target == nil || c.GetCurrentHP() < target.GetCurrentHP() {
				target = c
			}
		}
	case shared.PrioritizeHighestMaxHP:
		for _, c := range characters {
			if target == nil || c.HP.MaxHP > target.HP.MaxHP {
				target = c
			}
		}
	case shared.PrioritizeLowestMaxHP:
		for _, c := range characters {
			if target == nil || c.HP.MaxHP < target.HP.MaxHP {
				target = c
			}
		}
	case shared.PrioritizeHealer:
		for _, c := range characters {
			if c.Class.SpellcastingMod != "None" {
				for _, s := range c.KnownSpells {
					if s.SpellType == "Heal" {
						return c, nil
					}
				}
			}
		}
	case shared.PrioritizeSpellcasting:
		for _, c := range characters {
			if len(c.KnownSpells) > 0 {
				return c, nil
			}
		}
	default:
		panic("unhandled default case")
	}
	return target, nil
}

func (e *Encounter) filterMonsters() []*monster.Monster {
	var monsters []*monster.Monster
	for _, entity := range e.CombatTracker {
		if m, ok := entity.Creature.(*monster.Monster); ok {
			monsters = append(monsters, m)
		}
	}
	return monsters
}

func (e *Encounter) filterCharacters() []*character.Character {
	var characters []*character.Character
	for _, entity := range e.CombatTracker {
		if c, ok := entity.Creature.(*character.Character); ok {
			characters = append(characters, c)
		}
	}
	return characters
}
