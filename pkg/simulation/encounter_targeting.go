package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
	"math/rand/v2"
)

func (e *Encounter) chooseDamageTargetByPriority(actor core.Entity) (core.Entity, error) {
	switch actor.(type) {
	case *character.Character:
		// if attacker has been confused or chargmed
		// get lists of both
		monsters := e.filterMonsters()
		if len(monsters) == 0 {
			return nil, fmt.Errorf("no targets available")
		}
		return e.selectMonsterByPriority(monsters)

	case *monster.Monster:
		// if attacker has been confused or chargmed
		// get lists of both
		characters := e.filterCharacters()
		if len(characters) == 0 {
			return nil, fmt.Errorf("no targets available")
		}
		return e.selectCharacterByPriority(characters)

	default:
		return nil, fmt.Errorf("Unknown creature type %T\n", actor)
	}
	return nil, fmt.Errorf("Unknown creature type %T\n", actor)
}

// chooseHealTargetByPriority determines the most suitable healing target based on the actor and prioritization strategy.
// It returns the selected target entity or an error if no valid target is found.
func (e *Encounter) chooseHealTargetByPriority(actor core.Entity) (core.Entity, error) {
	switch actor.(type) {
	case *character.Character:
		characters := e.filterCharacters()
		if len(characters) == 0 {
			return nil, fmt.Errorf("no targets available")
		}

		needsHealing := e.filterCharactersNeedingHealing(characters)
		if len(needsHealing) == 0 {
			return nil, fmt.Errorf("no characters need healing")
		}

		switch e.Options.HealPriority {
		case shared.PrioritizeLowestHealth:
			return e.findLowestHPCharacter(characters), nil
		case shared.PrioritizeMostDamaged:
			return e.findMostDamagedCharacter(characters), nil
		case shared.PrioritizeHealer:
			return e.findBestHealer(characters), nil
		case shared.PrioritizeSpellcasting:
			return e.findBestSpellcaster(characters), nil
		case shared.NoPriority:
			fallthrough
		default:
			return characters[rand.IntN(len(characters))], nil
		}
	// TODO: Add monsters
	case *monster.Monster:
		return nil, fmt.Errorf("no targets available")
	}
	return nil, fmt.Errorf("no valid targets found")
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
		if m, ok := entity.Entity.(*monster.Monster); ok {
			monsters = append(monsters, m)
		}
	}
	return monsters
}

func (e *Encounter) filterCharacters() []*character.Character {
	var characters []*character.Character
	for _, entity := range e.CombatTracker {
		if c, ok := entity.Entity.(*character.Character); ok {
			characters = append(characters, c)
		}
	}
	return characters
}

func (e *Encounter) filterCharactersNeedingHealing(characters []*character.Character) []*character.Character {
	var needsHealing []*character.Character
	for _, c := range characters {
		if c.GetCurrentHPPct() < e.Options.PlayerHealThresholdPct {
			needsHealing = append(needsHealing, c)
		}
	}
	return needsHealing
}

func (e *Encounter) findLowestHPCharacter(characters []*character.Character) *character.Character {
	lowest := characters[0]
	for _, c := range characters {
		if lowest == nil || c.GetCurrentHP() < lowest.GetCurrentHP() {
			lowest = c
		}
	}
	return lowest
}

func (e *Encounter) findMostDamagedCharacter(characters []*character.Character) *character.Character {
	mostDamaged := characters[0]
	maxDamage := mostDamaged.GetMaxHP() - mostDamaged.GetCurrentHP()
	for _, c := range characters[1:] {
		damage := c.GetMaxHP() - c.GetCurrentHP()
		if damage > maxDamage {
			mostDamaged = c
			maxDamage = damage
		}
	}
	return mostDamaged
}

func (e *Encounter) findBestHealer(characters []*character.Character) *character.Character {
	bestHealer := characters[0]
	for _, c := range characters[1:] {
		if len(c.Class.Spellcasting.ClassHealingSpells) > len(bestHealer.Class.Spellcasting.ClassHealingSpells) {
			bestHealer = c
		}
	}
	return bestHealer
}

func (e *Encounter) findBestSpellcaster(characters []*character.Character) *character.Character {
	bestCaster := characters[0]
	for _, c := range characters[1:] {
		if len(c.Class.Spellcasting.ClassDamageSpells) > len(bestCaster.Class.Spellcasting.ClassDamageSpells) {
			bestCaster = c
		}
	}
	return bestCaster
}
