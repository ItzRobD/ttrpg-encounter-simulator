package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	"sort"
)

type CombatEngine struct {
	CurrentRound  int
	Combatants    map[int]core.Combatant
	CombatTracker []int
}

func NewCombatEngine() *CombatEngine {
	return &CombatEngine{
		CurrentRound:  0,
		Combatants:    make(map[int]core.Combatant),
		CombatTracker: nil,
	}
}

func (ce *CombatEngine) AddCombatant(c core.Combatant) {
	if ce.Combatants == nil {
		ce.Combatants = make(map[int]core.Combatant)
	}

	ce.Combatants[len(ce.Combatants)] = c
}

func (ce *CombatEngine) SetupCombat() error {
	ce.CurrentRound = 0

	if len(ce.Combatants) <= 0 {
		return fmt.Errorf("combatants list is empty")
	}

	for id, c := range ce.Combatants {
		entity := c.GetEntity()

		initiative, err := entity.RollInitiative()
		if err != nil {
			return err
		}

		updatedCombatant := core.NewCombatant(entity, initiative)
		ce.Combatants[id] = updatedCombatant
	}

	return ce.setupCombatTracker()
}

func (ce *CombatEngine) setupCombatTracker() error {
	ce.CombatTracker = make([]int, 0, len(ce.Combatants))
	for id := range ce.Combatants {
		ce.CombatTracker = append(ce.CombatTracker, id)
	}

	sort.Slice(ce.CombatTracker, func(i, j int) bool {
		idxI := ce.CombatTracker[i]
		idxJ := ce.CombatTracker[j]

		initI := ce.Combatants[idxI].GetInitiative()
		initJ := ce.Combatants[idxJ].GetInitiative()

		return initI > initJ
	})

	return nil
}

func (ce *CombatEngine) rollInitiativeForAllCombatants() error {
	for id, c := range ce.Combatants {
		init, err := c.GetEntity().RollInitiative()
		if err != nil {
			return err
		}

		updatedCombatant := c
		updatedCombatant.Initiative = init
		ce.Combatants[id] = updatedCombatant
	}
	return nil
}

//func (ce *CombatEngine) SimulateRound() {
//	for _, combatantID := range ce.CombatTracker {
//		combatant := ce.Combatants[combatantID]
//
//		if !combatant.CanAct() {
//			continue // Skip unconscious combatants
//		}
//
//		switch combatant.GetEntity().IsCharacter() {
//		case true:
//			ce.handleCharacterTurn(entity)
//		case false:
//			ce.handleMonsterTurn(entity)
//		default:
//			fmt.Printf("Unknown entity type: %T\n", entity)
//		}
//	}
//}

// Debug function
func (ce *CombatEngine) PrintCombatTracker() {
	order := 0
	for _, index := range ce.CombatTracker {
		order++
		fmt.Printf("Order Index: %d - Initiative: %d - Name: %s\n", order, ce.Combatants[index].GetInitiative(), ce.Combatants[index].GetEntity().GetName())
	}
}

// Debug function
func (ce *CombatEngine) PrintCombatants() {
	for _, c := range ce.Combatants {
		fmt.Printf("Name: %s\n", c.GetEntity().GetName())
	}
}
