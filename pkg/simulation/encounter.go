package simulation

import (
	"dnd5e-encounter-simulator-backend/internal/helpers"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
	"sort"
)

type Encounter struct {
	sim           *Simulation
	Party         []*character.Character
	Monsters      []*monster.Monster
	CombatTracker []core.Combatant
	Options       Options
	CurrentRound  int
}

func (e *Encounter) PrintCombatTracker() {
	fmt.Println("Combat Tracker")
	for _, c := range e.CombatTracker {
		fmt.Printf("Initiative: %d, Name: %s\n", c.InitiativeScore, c.Creature.GetName())
	}
}

func (e *Encounter) PrintEncounterMembers() {
	fmt.Println("Encounter Members")
	for _, c := range e.Party {
		fmt.Printf("Name: %s\n", c)
		helpers.PrintStructFields(c, "")
	}
	for _, m := range e.Monsters {
		fmt.Printf("Name: %s\n", m.GetName())
		helpers.PrintStructFields(m, "")
	}
}

func (e *Encounter) AddCombatant(c core.Combatant) error {
	if c.InitiativeScore <= 0 {
		return fmt.Errorf("initiative score must be greater than zero")
	}
	if c.Creature == nil {
		return fmt.Errorf("creature cannot be nil")
	}
	e.CombatTracker = append(e.CombatTracker, c)
	return nil
}

func (e *Encounter) AddPartyMember(c *character.Character) {
	e.Party = append(e.Party, c)
}

func (e *Encounter) AddMonster(m *monster.Monster) error {
	m.EventListener = func(event interface{}) {
		if evt, ok := event.(events.CombatEvent); ok {
			e.sim.LogEvent(evt)
		}
	}
	if e.Options.UseMonsterHPAverage {
		hp, _, err := m.DetermineMonsterHP(true)
		if err != nil {
			return err
		}
		m.HP.HP = hp
	} else {
		hp, _, err := m.DetermineMonsterHP(false)
		if err != nil {
			return err
		}
		m.HP.HP = hp
	}
	e.Monsters = append(e.Monsters, m)
	return nil
}

func (e *Encounter) SetupCombatTracker() error {
	if e.CombatTracker == nil {
		e.CombatTracker = []core.Combatant{}
	}

	for _, m := range e.Monsters {
		initiative, err := shared.InitiativeRoll(m.AbilityScores.Dexterity)
		if err != nil {
			return err
		}
		err = e.AddCombatant(core.Combatant{
			InitiativeScore: initiative,
			Creature:        m,
		})
		if err != nil {
			return err
		}
	}
	for _, p := range e.Party {
		initiative, err := shared.InitiativeRoll(p.AbilityScores.Dexterity)
		if err != nil {
			return err
		}
		err = e.AddCombatant(core.Combatant{
			InitiativeScore: initiative,
			Creature:        p,
		})
		if err != nil {
			return err
		}
	}

	sort.Slice(e.CombatTracker, func(i, j int) bool {
		return e.CombatTracker[i].InitiativeScore > e.CombatTracker[j].InitiativeScore
	})

	return nil
}

// SimulateRound processes a single round of the encounter. It iterates through all entities in the combat tracker and executes their turns.
func (e *Encounter) SimulateRound() {
	for _, entity := range e.CombatTracker {
		switch creature := entity.Creature.(type) {
		case *character.Character:
			if creature.IsUnconscious() {
				continue // Skip if the character is unconscious
			}
			// TODO: I want to split turn logic, specifically spellcasting, into the shared package
			//       Since both monsters and characters will need to choose spells. Choosing actions
			//       Should be handled through the interface -> Add GetAction() or some similar method
			//       Move any shared logic with rolls etc to the shared package
			e.handleCharacterTurn(creature)
		case *monster.Monster:
			if creature.IsUnconscious() {
				continue // Skip if the monster is unconscious
			}
			//e.handleMonsterTurn(creature) // TODO: Implement monster logic
		default:
			fmt.Printf("Unknown creature type %T\n", creature)
		}
	}
}

// handleCharacterTurn manages the actions of a character during their turn in an encounter.
func (e *Encounter) handleCharacterTurn(character *character.Character) {
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
		e.performCharacterRangedAttack(character)
	case shared.ATMelee:
		e.performCharacterMeleeAttack(character)
	case shared.ATSpell:
		e.performCharacterSpellAttack(character)
		return
	case shared.ATNoAction:
		fallthrough
	default:
		// No action to be taken
		return
	}
}

func (e *Encounter) performCharacterRangedAttack(character *character.Character) {
	target, _ := e.chooseDamageTarget(character)
	if monsterTarget, ok := target.(*monster.Monster); ok {
		_, err := character.MakeWeaponAttack(monsterTarget, "ranged")
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("Target is not a monster\n")
	}
}

func (e *Encounter) performCharacterMeleeAttack(character *character.Character) {
	target, _ := e.chooseDamageTarget(character)
	if monsterTarget, ok := target.(*monster.Monster); ok {
		// TODO: Add secondary slot
		_, err := character.MakeWeaponAttack(monsterTarget, "primary")
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("Target is not a monster\n")
	}
}

func (e *Encounter) performCharacterHealAction(c *character.Character) {
	target, _ := e.chooseHealTarget(c)
	healingSpell, err := e.chooseBestHealingSpell(c, target)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = c.MakeSpellHeal(target, healingSpell)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (e *Encounter) performCharacterSpellAttack(c *character.Character) {
	if !e.Options.AOEHitsAllEnemies {
		target, _ := e.chooseDamageTarget(c)
		damageSpell, err := e.chooseDamageSpell(c, shared.SPHighestLevel)
		if err != nil {
			fmt.Println(err)
		}
		if monsterTarget, ok := target.(*monster.Monster); ok {
			_, err2 := c.MakeSpellAttack(monsterTarget, damageSpell)
			if err2 != nil {
				fmt.Println(err2)
			}
		}
	}
}

func (e *Encounter) handleMonsterTurn(monster *monster.Monster) {
	// TODO: Implement monster turn logic
}
