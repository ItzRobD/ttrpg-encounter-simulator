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
		fmt.Printf("Initiative: %d, Name: %s\n", c.InitiativeScore, c.Entity.GetName())
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
	if c.Entity == nil {
		return fmt.Errorf("creature cannot be nil")
	}
	e.CombatTracker = append(e.CombatTracker, c)
	return nil
}

func (e *Encounter) SetupCombatTracker() error {
	if e.CombatTracker == nil {
		e.CombatTracker = []core.Combatant{}
	}

	for _, m := range e.Monsters {
		initiative, roll, err := shared.InitiativeRoll(m.AbilityScores.Dexterity, m.EntityModifiers.InitiativeBonus, m.EntityModifiers.InitiativeAdvantage)
		if err != nil {
			return err
		}
		events.LogDiceRollEvent(m, initiative, []int{roll}, shared.DiceRollInitiative, m.EntityModifiers.InitiativeBonus, m.EventListener)
		err = e.AddCombatant(core.Combatant{
			InitiativeScore: initiative,
			Entity:          m,
		})
		if err != nil {
			return err
		}
	}
	for _, p := range e.Party {
		initiative, roll, err := shared.InitiativeRoll(p.AbilityScores.Dexterity, p.EntityModifiers.InitiativeBonus, p.EntityModifiers.InitiativeAdvantage)
		if err != nil {
			return err
		}
		events.LogDiceRollEvent(p, initiative, []int{roll}, shared.DiceRollInitiative, p.EntityModifiers.InitiativeBonus, p.EventListener)
		err = e.AddCombatant(core.Combatant{
			InitiativeScore: initiative,
			Entity:          p,
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
		switch creature := entity.Entity.(type) {
		case *character.Character:
			if creature.IsUnconscious() {
				continue // Skip if the character is unconscious
			}
			// TODO: I want to split turn logic, specifically spellcasting, into the shared package
			//       Since both monsters and characters will need to choose spells. Choosing actions
			//       Should be handled through the interface -> Add GetAction() or some similar method
			//       Move any shared logic with rolls etc to the shared package
			e.handleCharacterTurn(creature, shared.RollNormal)
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
