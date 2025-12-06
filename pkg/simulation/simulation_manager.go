package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"fmt"
	"math/rand/v2"
)

type SimulationManager struct {
	rng          *rand.Rand
	options      core.SimulationOptions
	dispatcher   *events.EventDispatcher
	combatEngine *CombatEngine
	simLog       []events.CombatEvent
	finalResult  core.VictoryStatus
}

func NewSimulationManager(options core.SimulationOptions, seed core.Seed) *SimulationManager {
	var s SimulationManager
	if seed.Seed1 == 0 {
		seed.Seed1 = rand.Uint64()
	}
	if seed.Seed2 == 0 {
		seed.Seed2 = rand.Uint64()
	}
	s.rng = rand.New(rand.NewPCG(seed.Seed1, seed.Seed2))

	dispatcher := events.NewEventDispatcher()
	dispatcher.RegisterHandler(&events.UniversalEventHandler{})
	s.options = options
	s.dispatcher = dispatcher
	s.combatEngine = NewCombatEngine(&s.options)
	return &s
}

func (s *SimulationManager) SetupEventListeners() {
	eventListener := func(event interface{}) {
		if evt, ok := event.(events.CombatEvent); ok {
			s.LogEvent(evt)
		}
	}

	for _, combatant := range s.combatEngine.Combatants {
		// Skip lair combatants
		if combatant.IsLair {
			continue
		}

		entity := combatant.Entity
		entity.SetEventListener(eventListener)
	}
}

func (s *SimulationManager) LogEvent(e events.CombatEvent) {
	if s.combatEngine.CurrentRound == 0 {
		e.SetRound(0)
	} else {
		e.SetRound(s.combatEngine.CurrentRound)
	}

	s.simLog = append(s.simLog, e)
	s.dispatcher.DispatchEvent(e)
}

func (s *SimulationManager) PrintSimulationLog() {
	fmt.Println("Simulation Log")
	for _, e := range s.simLog {
		fmt.Printf("%+v\n", e)
	}
}

func (s *SimulationManager) GetCombatEngine() *CombatEngine {
	return s.combatEngine
}

// RunSimulation executes a combat simulation for a given maximum number of rounds and returns an error if the simulation fails.
func (s *SimulationManager) RunSimulation(maxRounds int) error {
	err := s.combatEngine.SetupCombat()
	if err != nil {
		return err
	}
	victory, err := s.combatEngine.RunCombat(maxRounds)
	if err != nil {
		return err
	}

	if victory != core.VictoryStatusNone {
		// Record the final result and emit a simple console message.
		s.finalResult = victory
		// Note: Structured event emission would require an actor; for now, print.
		switch victory {
		case core.VictoryStatusCharacters:
			fmt.Println("Victory: Characters win")
		case core.VictoryStatusMonsters:
			fmt.Println("Victory: Monsters win")
		default:
			fmt.Println("Victory: Resolved with status", victory)
		}
	}

	return nil
}

func (s *SimulationManager) InitializeCombatants() {
	for _, combatant := range s.combatEngine.Combatants {
		// Skip lair combatants
		if combatant.IsLair {
			continue
		}

		combatant.GetEntity().InitializeHP()
	}
}

func (s *SimulationManager) SetupCombatantsFromAPI(ctx context.Context, characterConfigs []character.CharacterConfig, monsterIDs []int) (*SetupResult, error) {
	setupManager := NewCombatantSetupManager(ctx, s.options.UseHPAverageCharacter, s.options.UseHPAverageMonster)

	result, err := setupManager.SetupCombatants(characterConfigs, monsterIDs)
	if err != nil {
		return result, err
	}

	// Add all valid combatants to the engine
	for _, combatant := range result.Combatants {
		fmt.Println(combatant.Entity.GetName() + " added to combat engine")
		s.combatEngine.AddCombatant(combatant)
	}

	return result, nil
}

// GetFinalResult returns the last recorded victory status for the simulation run.
func (s *SimulationManager) GetFinalResult() core.VictoryStatus { return s.finalResult }
