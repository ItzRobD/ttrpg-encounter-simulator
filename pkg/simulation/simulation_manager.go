package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/lair"
	"dnd5e-encounter-simulator-backend/pkg/monster"
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

// MultiSimulationRequest defines the parameters for running multiple simulations.
type MultiSimulationRequest struct {
	BaseOptions      core.SimulationOptions
	CharacterConfigs []character.CharacterConfig
	MonsterIDs       []int
	MonsterConfigs   []monster.MonsterConfig
	LairConfig       *lair.LairConfig
	NumRuns          int
	MaxRounds        int
	IncludeLogs      bool // If true, full event logs will be included in the results.
}

// MultiSimulationResult aggregates the results of multiple simulation runs.
type MultiSimulationResult struct {
	TotalRuns          int                          `json:"total_runs"`
	CharacterVictories int                          `json:"character_victories"`
	MonsterVictories   int                          `json:"monster_victories"`
	OtherVictories     int                          `json:"other_victories"`
	AverageRounds      float64                      `json:"average_rounds"`
	IndividualResults  []IndividualSimulationResult `json:"individual_results,omitempty"`
}

// IndividualSimulationResult holds data for a single simulation run within a multi-run.
type IndividualSimulationResult struct {
	RunID         int                  `json:"run_id"`
	VictoryStatus core.VictoryStatus   `json:"victory_status"`
	Rounds        int                  `json:"rounds"`
	Seed          core.Seed            `json:"seed"`
	Logs          []events.CombatEvent `json:"logs,omitempty"`
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
		master = core.Seed{Seed1: 0xD20, Seed2: 0xD1CE}
	}
	s.rng = rand.New(rand.NewPCG(master.Seed1, master.Seed2))

	dispatcher := events.NewEventDispatcher()
	dispatcher.RegisterHandler(&events.UniversalEventHandler{})
	dispatcher.RegisterHandler(&events.TimelineHandler{})
	s.options = options
	s.dispatcher = dispatcher
	s.combatEngine = NewCombatEngine(&s.options)
	s.combatEngine.EventContext = core.NewEventContext()
	return &s
}

func (s *SimulationManager) SetupEventListeners() {
	eventListener := func(event interface{}) {
		if evt, ok := event.(events.CombatEvent); ok {
			s.LogEvent(evt)
		}
	}

	ids := s.combatEngine.getSortedCombatantIDs()
	for _, id := range ids {
		combatant := s.combatEngine.Combatants[id]
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

func (s *SimulationManager) GetTimeline() []events.TimelineEvent {
	for _, h := range s.dispatcher.GetHandlers() {
		if th, ok := h.(*events.TimelineHandler); ok {
			return th.Timeline
		}
	}
	return nil
}

// RunMultiSimulation executes multiple simulations in parallel and returns an aggregated result.
func RunMultiSimulation(ctx context.Context, req MultiSimulationRequest) (*MultiSimulationResult, error) {
	return RunMultiSimulationWithSetup(ctx, req, nil)
}

// RunMultiSimulationWithSetup allows providing an optional setup function for testing purposes.
// If setup is nil, it uses the default DB-based setup.
func RunMultiSimulationWithSetup(ctx context.Context, req MultiSimulationRequest, setup func(sm *SimulationManager) error) (*MultiSimulationResult, error) {
	if req.NumRuns <= 0 {
		return nil, fmt.Errorf("number of runs must be greater than 0")
	}

	resultsChan := make(chan IndividualSimulationResult, req.NumRuns)
	errChan := make(chan error, req.NumRuns)

	// We use a master RNG to generate seeds for each run to ensure they are different but reproducible if the master seed is fixed.
	masterSeed := req.BaseOptions.Seed
	if masterSeed.Seed1 == 0 && masterSeed.Seed2 == 0 {
		masterSeed = core.Seed{Seed1: uint64(rand.Int64()), Seed2: uint64(rand.Int64())}
	}
	masterRNG := rand.New(rand.NewPCG(masterSeed.Seed1, masterSeed.Seed2))

	for i := 0; i < req.NumRuns; i++ {
		runSeed := core.Seed{Seed1: masterRNG.Uint64(), Seed2: masterRNG.Uint64()}

		go func(runID int, seed core.Seed) {
			// Each goroutine gets its own SimulationManager
			opts := req.BaseOptions
			opts.Seed = seed
			sm := NewSimulationManager(opts, seed)

			if setup != nil {
				if err := setup(sm); err != nil {
					errChan <- fmt.Errorf("run %d setup failed: %w", runID, err)
					return
				}
			} else {
				// Setup combatants
				// Note: SetupCombatantsFromAPI uses the DB, so we need to ensure the DB connection is thread-safe.
				// In Go, sql.DB is thread-safe.
				_, err := sm.SetupCombatantsFromAPIWithLair(ctx, req.CharacterConfigs, req.MonsterIDs, req.MonsterConfigs, req.LairConfig)
				if err != nil {
					errChan <- fmt.Errorf("run %d setup failed: %w", runID, err)
					return
				}

				sm.InitializeCombatants()
			}

			err := sm.RunSimulation(req.MaxRounds)
			if err != nil {
				errChan <- fmt.Errorf("run %d execution failed: %w", runID, err)
				return
			}

			res := IndividualSimulationResult{
				RunID:         runID,
				VictoryStatus: sm.GetFinalResult(),
				Rounds:        sm.GetCombatEngine().CurrentRound,
				Seed:          seed,
			}
			if req.IncludeLogs {
				res.Logs = sm.simLog
			}
			resultsChan <- res
		}(i, runSeed)
	}

	multiResult := &MultiSimulationResult{
		TotalRuns:         req.NumRuns,
		IndividualResults: make([]IndividualSimulationResult, 0, req.NumRuns),
	}

	var totalRounds int
	for i := 0; i < req.NumRuns; i++ {
		select {
		case res := <-resultsChan:
			multiResult.IndividualResults = append(multiResult.IndividualResults, res)
			totalRounds += res.Rounds
			switch res.VictoryStatus {
			case core.VictoryStatusCharacters:
				multiResult.CharacterVictories++
			case core.VictoryStatusMonsters:
				multiResult.MonsterVictories++
			default:
				multiResult.OtherVictories++
			}
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if multiResult.TotalRuns > 0 {
		multiResult.AverageRounds = float64(totalRounds) / float64(multiResult.TotalRuns)
	}

	return multiResult, nil
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
	ids := s.combatEngine.getSortedCombatantIDs()
	for _, id := range ids {
		combatant := s.combatEngine.Combatants[id]
		// Skip lair combatants
		if combatant.IsLair {
			continue
		}

		combatant.GetEntity().InitializeHP()
	}
}

func (s *SimulationManager) SetupCombatantsFromAPI(ctx context.Context, characterConfigs []character.CharacterConfig, monsterIDs []int, monsterConfigs []monster.MonsterConfig) (*SetupResult, error) {
	setupManager := NewCombatantSetupManager(ctx, s.options.UseHPAverageCharacter, s.options.UseHPAverageMonster, s.rng)

	result, err := setupManager.SetupCombatants(characterConfigs, monsterIDs, monsterConfigs)
	if err != nil {
		return result, err
	}

	if result.Errors != nil && len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Printf("Encountered errors during combatant setup: %+v\n", e)
		}
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
func (s *SimulationManager) SetupCombatantsFromAPIWithLair(ctx context.Context, characterConfigs []character.CharacterConfig, monsterIDs []int, monsterConfigs []monster.MonsterConfig, lairCfg *lair.LairConfig) (*SetupResult, error) {
	res, err := s.SetupCombatantsFromAPI(ctx, characterConfigs, monsterIDs, monsterConfigs)
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
