package simulation

import (
	"dnd5e-encounter-simulator-backend/internal/util"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"fmt"
	"sort"
)

type Encounter struct {
	sim           *Simulation
	Party         []*character.Character
	Monsters      []*monster.Monster
	CombatTracker []core.Combatant
	Options       core.SimulationOptions
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
		util.PrintStructFields(c, "")
	}
	for _, m := range e.Monsters {
		fmt.Printf("Name: %s\n", m.GetName())
		util.PrintStructFields(m, "")
	}
}

func (e *Encounter) AddCombatant(c core.Combatant) error {
	// TODO: Handle multiple combatants of the same type -> Bandit 1, Bandit 2, Bandit 3, etc.
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
		initiative, roll, err := core.InitiativeRoll(m.AbilityScores.Dexterity, m.EntityModifiers.InitiativeBonus, m.EntityModifiers.InitiativeAdvantage)
		if err != nil {
			return err
		}
		events.LogDiceRollEvent(m, initiative, []int{roll}, core.DiceRollInitiative, m.EntityModifiers.InitiativeBonus, m.EventListener)
		err = e.AddCombatant(core.Combatant{
			InitiativeScore: initiative,
			Entity:          m,
		})
		if err != nil {
			return err
		}
	}
	for _, p := range e.Party {
		initiative, roll, err := core.InitiativeRoll(p.AbilityScores.Dexterity, p.EntityModifiers.InitiativeBonus, p.EntityModifiers.InitiativeAdvantage)
		if err != nil {
			return err
		}
		events.LogDiceRollEvent(p, initiative, []int{roll}, core.DiceRollInitiative, p.EntityModifiers.InitiativeBonus, p.EventListener)
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
			e.handleCharacterTurn(creature, core.RollNormal)
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
