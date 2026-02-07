package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg_old/character"
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/events"
	"dnd5e-encounter-simulator-backend/pkg_old/lair"
	"dnd5e-encounter-simulator-backend/pkg_old/monster"
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"
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
	BaseOptions      core.SimulationOptions      `json:"base_options"`
	CharacterConfigs []character.CharacterConfig `json:"character_configs"`
	MonsterIDs       []int                       `json:"monster_ids"`
	MonsterConfigs   []monster.MonsterConfig     `json:"monster_configs"`
	LairConfig       *lair.LairConfig            `json:"lair_config"`
	NumberOfRuns     int                         `json:"number_of_runs"`
	MaxRounds        int                         `json:"max_rounds"`
	IncludeLogs      bool                        `json:"include_logs"`
}

// MultiSimulationResult aggregates the results of multiple simulation runs.
type MultiSimulationResult struct {
	TotalRuns          int                          `json:"total_runs"`
	CharacterVictories int                          `json:"character_victories"`
	MonsterVictories   int                          `json:"monster_victories"`
	OtherVictories     int                          `json:"other_victories"`
	AverageRounds      float64                      `json:"average_rounds"`
	WinRatePercentage  float64                      `json:"win_rate_percentage"`
	InitialState       map[int]core.CombatantInfo   `json:"initial_state,omitempty"`
	IndividualResults  []IndividualSimulationResult `json:"individual_results,omitempty"`
	Performance        *PerformanceMetrics          `json:"performance,omitempty"`
}

type PerformanceMetrics struct {
	ExecutionTimeMs    int64   `json:"execution_time_ms"`
	ExecutionTimeHuman string  `json:"execution_time_human"`
	MemoryAllocatedMb  float64 `json:"memory_allocated_mb"`
	PeakGoroutines     int     `json:"peak_goroutines"`
}

// IndividualSimulationResult holds data for a single simulation run within a multi-run.
type IndividualSimulationResult struct {
	RunID         int                        `json:"run_id"`
	VictoryStatus core.VictoryStatus         `json:"victory_status"`
	Rounds        int                        `json:"rounds"`
	Seed          core.Seed                  `json:"seed"`
	Logs          []events.TimelineEvent     `json:"logs,omitempty"`
	InitialState  map[int]core.CombatantInfo `json:"initial_state,omitempty"`
	FinalState    map[int]core.CombatantInfo `json:"final_state,omitempty"`
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
	startTime := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	if req.NumberOfRuns <= 0 {
		return nil, fmt.Errorf("number of runs must be greater than 0")
	}

	resultsChan := make(chan IndividualSimulationResult, req.NumberOfRuns)
	errChan := make(chan error, req.NumberOfRuns)

	// We use a master RNG to generate seeds for each run to ensure they are different but reproducible if the master seed is fixed.
	masterSeed := req.BaseOptions.Seed
	if masterSeed.Seed1 == 0 && masterSeed.Seed2 == 0 {
		masterSeed = core.Seed{Seed1: uint64(rand.Int64()), Seed2: uint64(rand.Int64())}
	}
	masterRNG := rand.New(rand.NewPCG(masterSeed.Seed1, masterSeed.Seed2))

	for i := 0; i < req.NumberOfRuns; i++ {
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
			}

			// Capture initial state from the first run
			var initialState map[int]core.CombatantInfo
			if runID == 0 {
				// We need to initialize combatants temporarily to capture the state
				// but we don't want to log these events yet if we haven't attached listeners.
				// However, RunSimulation will call InitializeCombatants again, which might
				// be redundant or cause double logging if we attach listeners now.
				// TO ENSURE INITIAL STATE IS CORRECT (e.g. HP is rolled), we call it here.
				sm.InitializeCombatants()
				initialState = make(map[int]core.CombatantInfo)
				for id, combatant := range sm.combatEngine.Combatants {
					if combatant.Info != nil {
						combatant.Info.UpdateState()
						initialState[id] = *combatant.Info
					}
				}
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
				InitialState:  initialState,
				FinalState:    make(map[int]core.CombatantInfo),
			}

			// Collect final state for all combatants
			for id, combatant := range sm.combatEngine.Combatants {
				if combatant.Info != nil {
					combatant.Info.UpdateState()
					res.FinalState[id] = *combatant.Info
				}
			}
			if req.IncludeLogs {
				res.Logs = sm.GetTimeline()
			}
			resultsChan <- res
		}(i, runSeed)
	}

	multiResult := &MultiSimulationResult{
		TotalRuns:         req.NumberOfRuns,
		IndividualResults: make([]IndividualSimulationResult, 0, req.NumberOfRuns),
	}

	var totalRounds int
	for i := 0; i < req.NumberOfRuns; i++ {
		select {
		case res := <-resultsChan:
			totalRounds += res.Rounds
			if res.RunID == 0 {
				multiResult.InitialState = res.InitialState
				// Clear from individual result to avoid duplication in JSON/DB
				res.InitialState = nil
			}
			multiResult.IndividualResults = append(multiResult.IndividualResults, res)
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
		multiResult.WinRatePercentage = float64(multiResult.CharacterVictories+multiResult.MonsterVictories) / float64(multiResult.TotalRuns) * 100
	}

	// Capture ending performance metrics
	executionTime := time.Since(startTime)
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)
	allocMB := float64(memEnd.TotalAlloc-memStart.TotalAlloc) / 1024 / 1024

	multiResult.Performance = &PerformanceMetrics{
		ExecutionTimeMs:    executionTime.Milliseconds(),
		ExecutionTimeHuman: executionTime.String(),
		MemoryAllocatedMb:  allocMB,
		PeakGoroutines:     runtime.NumGoroutine(),
	}

	return multiResult, nil
}

// RunSimulation executes a combat simulation for a given maximum number of rounds and returns an error if the simulation fails.
func (s *SimulationManager) RunSimulation(maxRounds int) error {
	// Attach event listeners after combatants (including lair) are fully set up
	// but BEFORE SetupCombat so initiative rolls and other round 0 events are captured.
	s.SetupEventListeners()

	// Initialize HP for all combatants if they haven't been already
	// (InitialState capture might have already called this, but it's idempotent for set values/average,
	// and for rolls it would re-roll which we might want to avoid, but SetupCombat will proceed with whatever HP is set)
	// Actually, InitializeHP is only called if HP is not already set in some implementations.
	// But let's be safe and ensure it's called once listeners are attached so it's logged.
	s.InitializeCombatants()

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

		// Log the victory event for the timeline
		var winningSide events.WinningSide
		switch victory {
		case core.VictoryStatusCharacters:
			winningSide = events.WinningSideCharacters
		case core.VictoryStatusMonsters:
			winningSide = events.WinningSideMonsters
		default:
			winningSide = events.WinningSideNone
		}

		events.LogVictoryEvent(s.combatEngine.EventContext, winningSide, s.combatEngine.CurrentRound, func(event interface{}) {
			if ce, ok := event.(events.CombatEvent); ok {
				s.LogEvent(ce)
			}
		})
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

		// Only initialize if HP hasn't been set yet (CurrentHP is 0)
		// This prevents re-rolling HP if InitializeCombatants is called multiple times.
		if combatant.GetEntity().GetHPStatus().GetHP() == 0 {
			combatant.GetEntity().InitializeHP()
		}
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
