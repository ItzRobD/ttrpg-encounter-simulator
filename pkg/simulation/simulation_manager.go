package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/lair"
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
	// Determine the master seed: prefer explicit seed param, else options.Seed, else fixed default
	master := seed
	if master.Seed1 == 0 && master.Seed2 == 0 {
		master = options.Seed
	}
	if master.Seed1 == 0 && master.Seed2 == 0 {
		// Fixed default for reproducibility if caller provides no seed
		master = core.Seed{Seed1: 0xC0FFEE, Seed2: 0xBEEF}
	}
	s.rng = rand.New(rand.NewPCG(master.Seed1, master.Seed2))

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
		// Attach listener to all combatants, including lair, so events can be logged uniformly
		combatant.Entity.SetEventListener(eventListener)
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
	// Attach event listeners after combatants (including lair) are fully set up
	s.SetupEventListeners()
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
	setupManager := NewCombatantSetupManager(ctx, s.options.UseHPAverageCharacter, s.options.UseHPAverageMonster, s.rng)

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

// SetupCombatantsFromAPIWithLair allows providing an optional lair configuration.
// When s.options.AllowLairActions is true and lairCfg.Enabled is true, a lair combatant
// is constructed and added with initiative 20.
func (s *SimulationManager) SetupCombatantsFromAPIWithLair(ctx context.Context, characterConfigs []character.CharacterConfig, monsterIDs []int, lairCfg *lair.LairConfig) (*SetupResult, error) {
	res, err := s.SetupCombatantsFromAPI(ctx, characterConfigs, monsterIDs)
	if err != nil {
		return res, err
	}

	if s.options.AllowLairActions && lairCfg != nil && lairCfg.Enabled {
		// Lair always acts on initiative 20 (auto-loses ties handled by engine sort)
		lr, err2 := lair.NewLairFromConfig(lairCfg, s.rng)
		if err2 != nil {
			return res, err2
		}
		cb := core.NewCombatantWithInfo(lr)
		cb.Initiative = 20
		cb.IsLair = true
		s.combatEngine.AddCombatant(cb)
	}
	return res, nil
}

// GetFinalResult returns the last recorded victory status for the simulation run.
func (s *SimulationManager) GetFinalResult() core.VictoryStatus { return s.finalResult }
