package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"
)

// MultiSimulationRequest defines the parameters for running multiple simulations.
type MultiSimulationRequest struct {
	BaseOptions  core.SimulationOptions `json:"base_options"`
	ActorConfigs []actor.ActorConfig    `json:"actor_configs"`
	MonsterIDs   []int                  `json:"monster_ids"`
	NumberOfRuns int                    `json:"number_of_runs"`
	MaxRounds    int                    `json:"max_rounds"`
	IncludeLogs  bool                   `json:"include_logs"`
}

// MultiSimulationResult aggregates the results of multiple simulation runs.
type MultiSimulationResult struct {
	TotalRuns          int                          `json:"total_runs"`
	CharacterVictories int                          `json:"character_victories"`
	MonsterVictories   int                          `json:"monster_victories"`
	OtherVictories     int                          `json:"other_victories"`
	AverageRounds      float64                      `json:"average_rounds"`
	WinRatePercentage  float64                      `json:"win_rate_percentage"`
	ActorConfigs       map[int]actor.ActorConfig    `json:"actor_configs,omitempty"`
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
	RunID              int                       `json:"run_id"`
	VictoryStatus      core.VictoryStatus        `json:"victory_status"`
	Rounds             int                       `json:"rounds"`
	Seed               core.Seed                 `json:"seed"`
	LogsStripped       bool                      `json:"logs_stripped,omitempty"`
	ActorInitialStates map[int]ActorInitialState `json:"actor_initial_states,omitempty"`
	Logs               []events.TimelineEvent    `json:"logs,omitempty"`
}

type ActorInitialState struct {
	CurrentHP   int                  `json:"current_hp"`
	MaxHP       int                  `json:"max_hp"`
	TempHP      int                  `json:"temp_hp"`
	Conditions  core.ActorConditions `json:"conditions"`
	HealthState core.HealthState     `json:"health_state"`
}

// RunMultiSimulation executes multiple simulations and returns an aggregated result.
func RunMultiSimulation(ctx context.Context, req MultiSimulationRequest) (*MultiSimulationResult, error) {
	startTime := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	if req.NumberOfRuns <= 0 {
		return nil, fmt.Errorf("number of runs must be greater than 0")
	}

	resultsChan := make(chan IndividualSimulationResult, req.NumberOfRuns)
	errChan := make(chan error, req.NumberOfRuns)

	// Limit concurrency to avoid exhausting resources (like DB connections)
	maxConcurrency := runtime.GOMAXPROCS(0) * 2
	if maxConcurrency > 10 {
		maxConcurrency = 10 // Safe default to avoid DB connection pool exhaustion
	}
	sem := make(chan struct{}, maxConcurrency)

	masterSeed := req.BaseOptions.Seed
	if masterSeed.Seed1 == 0 && masterSeed.Seed2 == 0 {
		masterSeed = core.Seed{Seed1: uint64(time.Now().UnixNano()), Seed2: 42}
	}
	masterRNG := rand.New(rand.NewPCG(masterSeed.Seed1, masterSeed.Seed2))

	for i := 0; i < req.NumberOfRuns; i++ {
		runSeed := core.Seed{Seed1: masterRNG.Uint64(), Seed2: masterRNG.Uint64()}

		go func(runID int, seed core.Seed) {
			sem <- struct{}{}
			defer func() { <-sem }()

			ed := NewEncounterDirector(seed, &req.BaseOptions)
			sm := NewSetupManager(ctx, ed.RollManager)

			for _, cfg := range req.ActorConfigs {
				a, err := sm.SetupActor(cfg)
				if err != nil {
					errChan <- fmt.Errorf("run %d hydration failed for %s: %w", runID, cfg.Name, err)
					return
				}
				ed.AddActor(a)
			}

			ed.SetupEncounter()

			// Capture initial states
			initialStates := make(map[int]ActorInitialState)
			for id, a := range ed.Actors {
				// Clone conditions map to prevent simulation runs from modifying initial state results
				clonedConditions := make(core.ActorConditions)
				for k, v := range a.StateManager.Conditions {
					clonedConditions[k] = v
				}

				initialStates[id] = ActorInitialState{
					CurrentHP:   a.StateManager.CurrentHP,
					MaxHP:       a.StateManager.MaxHP,
					TempHP:      a.StateManager.TempHP,
					Conditions:  clonedConditions,
					HealthState: a.StateManager.HealthState,
				}
			}

			var victory core.VictoryStatus
			var err error
			rounds := 0
			for round := 0; round < req.MaxRounds; round++ {
				victory, err = ed.SimulateRound()
				if err != nil {
					errChan <- fmt.Errorf("run %d execution failed at round %d: %w", runID, round, err)
					return
				}
				rounds = ed.CurrentRound
				if victory != core.VictoryStatusNone {
					break
				}
			}

			res := IndividualSimulationResult{
				RunID:              runID,
				VictoryStatus:      victory,
				Rounds:             rounds,
				Seed:               seed,
				ActorInitialStates: initialStates,
			}

			if req.IncludeLogs {
				res.Logs = ed.ExportTimeline()
			}

			resultsChan <- res
		}(i, runSeed)
	}

	multiResult := &MultiSimulationResult{
		TotalRuns:         req.NumberOfRuns,
		IndividualResults: make([]IndividualSimulationResult, 0, req.NumberOfRuns),
	}

	maxLogs := req.BaseOptions.MaxLoggedRuns
	if maxLogs <= 0 {
		maxLogs = 10
	}

	// Buckets for balanced logging
	var charVicLogs []IndividualSimulationResult
	var monsterVicLogs []IndividualSimulationResult
	var otherVicLogs []IndividualSimulationResult

	var totalRounds int
	for i := 0; i < req.NumberOfRuns; i++ {
		select {
		case res := <-resultsChan:
			totalRounds += res.Rounds

			switch res.VictoryStatus {
			case core.VictoryStatusCharacters:
				multiResult.CharacterVictories++
				if req.IncludeLogs && len(charVicLogs) < maxLogs {
					charVicLogs = append(charVicLogs, res)
				} else {
					res.Logs = nil
					res.LogsStripped = req.IncludeLogs
					multiResult.IndividualResults = append(multiResult.IndividualResults, res)
				}
			case core.VictoryStatusMonsters:
				multiResult.MonsterVictories++
				if req.IncludeLogs && len(monsterVicLogs) < maxLogs {
					monsterVicLogs = append(monsterVicLogs, res)
				} else {
					res.Logs = nil
					res.LogsStripped = req.IncludeLogs
					multiResult.IndividualResults = append(multiResult.IndividualResults, res)
				}
			default:
				multiResult.OtherVictories++
				if req.IncludeLogs && len(otherVicLogs) < maxLogs {
					otherVicLogs = append(otherVicLogs, res)
				} else {
					res.Logs = nil
					res.LogsStripped = req.IncludeLogs
					multiResult.IndividualResults = append(multiResult.IndividualResults, res)
				}
			}
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Assemble balanced logs if requested
	if req.IncludeLogs {
		balanced := assembleBalancedLogs(charVicLogs, monsterVicLogs, otherVicLogs, maxLogs)
		multiResult.IndividualResults = append(multiResult.IndividualResults, balanced...)
	}

	if multiResult.TotalRuns > 0 {
		multiResult.AverageRounds = float64(totalRounds) / float64(multiResult.TotalRuns)
		multiResult.WinRatePercentage = (float64(multiResult.CharacterVictories) / float64(multiResult.TotalRuns)) * 100
	}

	// Capture actor configs for the multi-result (assuming they are consistent across runs)
	// We'll run a quick setup to get the deterministic instance IDs
	if len(req.ActorConfigs) > 0 {
		tempED := NewEncounterDirector(core.Seed{}, &req.BaseOptions)
		tempSM := NewSetupManager(ctx, tempED.RollManager)
		multiResult.ActorConfigs = make(map[int]actor.ActorConfig)
		for _, cfg := range req.ActorConfigs {
			a, err := tempSM.SetupActor(cfg)
			if err == nil {
				tempED.AddActor(a)
				multiResult.ActorConfigs[a.InstanceID] = cfg
			}
		}
	}

	// Final performance metrics
	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)
	multiResult.Performance = &PerformanceMetrics{
		ExecutionTimeMs:    time.Since(startTime).Milliseconds(),
		ExecutionTimeHuman: time.Since(startTime).String(),
		MemoryAllocatedMb:  float64(memEnd.TotalAlloc-memStart.TotalAlloc) / 1024 / 1024,
		PeakGoroutines:     runtime.NumGoroutine(),
	}

	return multiResult, nil
}

func assembleBalancedLogs(chars, monsters, others []IndividualSimulationResult, limit int) []IndividualSimulationResult {
	result := make([]IndividualSimulationResult, 0, limit)

	// Try to get an even split between characters and monsters first
	perSide := limit / 2
	if len(others) > 0 {
		// If there are 'other' victories (draws, etc.), maybe try 1/3 each
		perSide = limit / 3
	}

	// 1. Take from Characters
	takeChars := min(len(chars), perSide)
	result = append(result, chars[:takeChars]...)
	chars = chars[takeChars:]

	// 2. Take from Monsters
	takeMonsters := min(len(monsters), perSide)
	result = append(result, monsters[:takeMonsters]...)
	monsters = monsters[takeMonsters:]

	// 3. Take from Others
	takeOthers := min(len(others), perSide)
	result = append(result, others[:takeOthers]...)
	others = others[takeOthers:]

	// 4. Fill remaining slots from any bucket that still has data
	remaining := limit - len(result)
	if remaining > 0 && len(chars) > 0 {
		take := min(len(chars), remaining)
		result = append(result, chars[:take]...)
		chars = chars[take:]
		remaining -= take
	}
	if remaining > 0 && len(monsters) > 0 {
		take := min(len(monsters), remaining)
		result = append(result, monsters[:take]...)
		monsters = monsters[take:]
		remaining -= take
	}
	if remaining > 0 && len(others) > 0 {
		take := min(len(others), remaining)
		result = append(result, others[:take]...)
		others = others[take:]
		remaining -= take
	}

	// Mark any leftovers in our temporary buckets as stripped
	// Actually, we don't need to do anything with them because they were never added to multiResult.IndividualResults
	// BUT wait, in my loop above, if I don't add them to IndividualResults yet, I MUST add them now or they will be lost.
	// The ones that DIDN'T make it into the balanced selection should still be in the IndividualResults slice but without logs.

	for _, res := range chars {
		res.Logs = nil
		res.LogsStripped = true
		result = append(result, res)
	}
	for _, res := range monsters {
		res.Logs = nil
		res.LogsStripped = true
		result = append(result, res)
	}
	for _, res := range others {
		res.Logs = nil
		res.LogsStripped = true
		result = append(result, res)
	}

	return result
}
