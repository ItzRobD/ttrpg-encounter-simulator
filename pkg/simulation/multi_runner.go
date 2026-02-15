package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/simulation/intermission_manager"
	"fmt"
	"math/rand/v2"
	"runtime"
	"time"
)

// MultiSimulationRequest defines the parameters for running multiple adventuring day simulations.
type MultiSimulationRequest struct {
	AdventuringDayRequest
	NumberOfRuns int `json:"number_of_runs"`
}

// MultiSimulationResult aggregates the results of multiple adventuring day simulation runs.
type MultiSimulationResult struct {
	TotalRuns          int                          `json:"total_runs"`
	CharacterVictories int                          `json:"character_victories"` // Entire day won
	MonsterVictories   int                          `json:"monster_victories"`   // Party wiped at some point
	OtherVictories     int                          `json:"other_victories"`     // Draw or other
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

// IndividualSimulationResult holds data for a single adventuring day simulation run.
type IndividualSimulationResult struct {
	RunID            int                         `json:"run_id"`
	VictoryStatus    core.VictoryStatus          `json:"victory_status"`
	TotalRounds      int                         `json:"total_rounds"`
	Seed             core.Seed                   `json:"seed"`
	LogsStripped     bool                        `json:"logs_stripped,omitempty"`
	EncounterResults []IndividualEncounterResult `json:"encounter_results,omitempty"`
	ActorConfigs     map[int]actor.ActorConfig   `json:"actor_configs,omitempty"`
}

type ActorInitialState struct {
	CurrentHP   int                  `json:"current_hp"`
	MaxHP       int                  `json:"max_hp"`
	TempHP      int                  `json:"temp_hp"`
	Conditions  core.ActorConditions `json:"conditions"`
	HealthState core.HealthState     `json:"health_state"`
}

type EncounterConfig struct {
	Name           string              `json:"name"`
	MonsterIDs     []int               `json:"monster_ids,omitempty"`
	MonsterConfigs []actor.ActorConfig `json:"monster_configs"`
}

type AdventuringDayRequest struct {
	BaseOptions      core.SimulationOptions                   `json:"base_options"`
	CharacterConfigs []actor.ActorConfig                      `json:"character_configs"`
	Encounters       []EncounterConfig                        `json:"encounters"`
	Intermission     intermission_manager.IntermissionOptions `json:"intermission"`
	MaxRounds        int                                      `json:"max_rounds"`
	IncludeLogs      bool                                     `json:"include_logs"`
}

type AdventuringDayResult struct {
	TotalEncounters  int                         `json:"total_encounters"`
	EncountersWon    int                         `json:"encounters_won"`
	SuccessRate      float64                     `json:"success_rate"`
	AverageRounds    float64                     `json:"average_rounds"`
	EncounterResults []IndividualEncounterResult `json:"encounter_results,omitempty"`
	FinalActorStates map[int]actor.Actor         `json:"final_actor_states,omitempty"`
	ActorConfigs     map[int]actor.ActorConfig   `json:"actor_configs,omitempty"`
	Performance      *PerformanceMetrics         `json:"performance,omitempty"`
}

type IndividualEncounterResult struct {
	EncounterName string                    `json:"encounter_name"`
	VictoryStatus core.VictoryStatus        `json:"victory_status"`
	Rounds        int                       `json:"rounds"`
	Seed          core.Seed                 `json:"seed"`
	InitialState  map[int]ActorInitialState `json:"initial_state,omitempty"`
	Logs          []events.TimelineEvent    `json:"logs,omitempty"`
}

// RunMultiSimulation executes multiple adventuring day simulations and returns an aggregated result.
func RunMultiSimulation(ctx context.Context, req MultiSimulationRequest) (*MultiSimulationResult, error) {
	startTime := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	if req.NumberOfRuns <= 0 {
		return nil, fmt.Errorf("number of runs must be greater than 0")
	}

	resultsChan := make(chan IndividualSimulationResult, req.NumberOfRuns)
	errChan := make(chan error, req.NumberOfRuns)

	// Limit concurrency to avoid exhausting resources
	maxConcurrency := runtime.GOMAXPROCS(0) * 2
	if maxConcurrency > 10 {
		maxConcurrency = 10
	}
	sem := make(chan struct{}, maxConcurrency)

	masterSeed := req.BaseOptions.Seed
	if masterSeed.Seed1 == 0 && masterSeed.Seed2 == 0 {
		masterSeed = core.Seed{Seed1: uint64(time.Now().UnixNano()), Seed2: 42}
	}
	masterRNG := rand.New(rand.NewPCG(masterSeed.Seed1, masterSeed.Seed2))

	for i := 0; i < req.NumberOfRuns; i++ {
		daySeed := core.Seed{Seed1: masterRNG.Uint64(), Seed2: masterRNG.Uint64()}

		go func(runID int, seed core.Seed) {
			sem <- struct{}{}
			defer func() { <-sem }()

			// Clone request and set seed
			dayReq := req.AdventuringDayRequest
			dayReq.BaseOptions.Seed = seed

			dayRes, err := RunAdventuringDay(ctx, dayReq)
			if err != nil {
				errChan <- fmt.Errorf("run %d failed: %w", runID, err)
				return
			}

			// Determine day-level victory status
			dayVictory := core.VictoryStatusCharacters
			totalRounds := 0
			for _, enc := range dayRes.EncounterResults {
				totalRounds += enc.Rounds
				if enc.VictoryStatus != core.VictoryStatusCharacters {
					dayVictory = enc.VictoryStatus
				}
			}

			res := IndividualSimulationResult{
				RunID:            runID,
				VictoryStatus:    dayVictory,
				TotalRounds:      totalRounds,
				Seed:             seed,
				EncounterResults: dayRes.EncounterResults,
				ActorConfigs:     dayRes.ActorConfigs,
			}

			resultsChan <- res
		}(i, daySeed)
	}

	multiResult := &MultiSimulationResult{
		TotalRuns:         req.NumberOfRuns,
		IndividualResults: make([]IndividualSimulationResult, 0, req.NumberOfRuns),
		ActorConfigs:      make(map[int]actor.ActorConfig),
	}

	maxLogs := req.BaseOptions.MaxLoggedRuns
	if maxLogs <= 0 {
		maxLogs = 10
	}

	var charVicLogs []IndividualSimulationResult
	var monsterVicLogs []IndividualSimulationResult
	var otherVicLogs []IndividualSimulationResult

	var totalRounds int
	for i := 0; i < req.NumberOfRuns; i++ {
		select {
		case res := <-resultsChan:
			totalRounds += res.TotalRounds

			// Merge ActorConfigs
			for id, cfg := range res.ActorConfigs {
				if _, exists := multiResult.ActorConfigs[id]; !exists {
					multiResult.ActorConfigs[id] = cfg
				}
			}
			// Clear from individual result to save space if not needed there
			// res.ActorConfigs = nil // Keep it for now as per user request shape

			switch res.VictoryStatus {
			case core.VictoryStatusCharacters:
				multiResult.CharacterVictories++
				if req.IncludeLogs && len(charVicLogs) < maxLogs {
					charVicLogs = append(charVicLogs, res)
				} else {
					res.EncounterResults = nil
					res.LogsStripped = req.IncludeLogs
					multiResult.IndividualResults = append(multiResult.IndividualResults, res)
				}
			case core.VictoryStatusMonsters:
				multiResult.MonsterVictories++
				if req.IncludeLogs && len(monsterVicLogs) < maxLogs {
					monsterVicLogs = append(monsterVicLogs, res)
				} else {
					res.EncounterResults = nil
					res.LogsStripped = req.IncludeLogs
					multiResult.IndividualResults = append(multiResult.IndividualResults, res)
				}
			default:
				multiResult.OtherVictories++
				if req.IncludeLogs && len(otherVicLogs) < maxLogs {
					otherVicLogs = append(otherVicLogs, res)
				} else {
					res.EncounterResults = nil
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

	if req.IncludeLogs {
		balanced := assembleBalancedLogs(charVicLogs, monsterVicLogs, otherVicLogs, maxLogs)
		multiResult.IndividualResults = append(multiResult.IndividualResults, balanced...)
	}

	if multiResult.TotalRuns > 0 {
		multiResult.AverageRounds = float64(totalRounds) / float64(multiResult.TotalRuns)
		multiResult.WinRatePercentage = (float64(multiResult.CharacterVictories) / float64(multiResult.TotalRuns)) * 100
	}

	// Capture character configs from first run or provided configs
	// In MultiSimulation, we now aggregate them from all runs.

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

func RunAdventuringDay(ctx context.Context, req AdventuringDayRequest) (*AdventuringDayResult, error) {
	startTime := time.Now()
	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	masterSeed := req.BaseOptions.Seed
	if masterSeed.Seed1 == 0 && masterSeed.Seed2 == 0 {
		masterSeed = core.Seed{Seed1: uint64(time.Now().UnixNano()), Seed2: 42}
	}
	dayRNG := rand.New(rand.NewPCG(masterSeed.Seed1, masterSeed.Seed2))
	im := intermission_manager.NewIntermissionManager(roll_manager.NewRollManager(dayRNG))

	actorConfigs := make(map[int]actor.ActorConfig)
	var characters []*actor.Actor
	encounterResults := make([]IndividualEncounterResult, 0)
	encountersWon := 0
	totalRounds := 0
	nextInstanceID := 1

	// Initial character hydration
	sm := NewSetupManager(ctx, roll_manager.NewRollManager(dayRNG))
	for _, cfg := range req.CharacterConfigs {
		a, err := sm.SetupActor(cfg)
		if err != nil {
			return nil, fmt.Errorf("character hydration failed for %s: %w", cfg.Name, err)
		}
		// Assign InstanceID starting from 1 for characters
		// We ignore any pre-assigned InstanceID from the request to ensure a clean start
		a.InstanceID = nextInstanceID
		nextInstanceID++
		characters = append(characters, a)
		// Capture the fully hydrated config including the assigned InstanceID
		actorConfigs[a.InstanceID] = a.ToConfig()
	}

	for _, encCfg := range req.Encounters {
		// Each encounter gets its own seed for the director, but derived from dayRNG
		encSeed := core.Seed{Seed1: dayRNG.Uint64(), Seed2: dayRNG.Uint64()}
		ed := NewEncounterDirector(encSeed, &req.BaseOptions)

		// Characters carry over state
		for _, char := range characters {
			char.StateManager.ResetStateForNewEncounter()
			ed.AddActor(char)
		}

		// Fresh monsters for each encounter
		for _, mID := range encCfg.MonsterIDs {
			mCfg := actor.ActorConfig{
				ID:        fmt.Sprintf("%d", mID),
				ActorType: core.ActorTypeMonster,
				Side:      core.SideMonsters,
			}
			m, err := sm.SetupActor(mCfg)
			if err != nil {
				return nil, fmt.Errorf("monster hydration failed for ID %d: %w", mID, err)
			}
			// Assign a unique InstanceID across the entire adventuring day
			m.InstanceID = nextInstanceID
			nextInstanceID++
			ed.AddActor(m)

			// Capture the fully hydrated config including the assigned InstanceID
			actorConfigs[m.InstanceID] = m.ToConfig()
		}
		for _, mCfg := range encCfg.MonsterConfigs {
			m, err := sm.SetupActor(mCfg)
			if err != nil {
				return nil, fmt.Errorf("monster hydration failed: %w", err)
			}
			// Assign a unique InstanceID across the entire adventuring day
			m.InstanceID = nextInstanceID
			nextInstanceID++
			ed.AddActor(m)

			// Capture the fully hydrated config including the assigned InstanceID
			actorConfigs[m.InstanceID] = m.ToConfig()
		}

		ed.SetupEncounter()

		// Capture initial state
		initialState := make(map[int]ActorInitialState)
		for _, a := range ed.Actors {
			initialState[a.InstanceID] = ActorInitialState{
				CurrentHP:   a.StateManager.CurrentHP,
				MaxHP:       a.StateManager.MaxHP,
				TempHP:      a.StateManager.TempHP,
				Conditions:  a.StateManager.Conditions,
				HealthState: a.StateManager.HealthState,
			}
		}

		var victory core.VictoryStatus
		for round := 0; round < req.MaxRounds; round++ {
			vic, err := ed.SimulateRound()
			if err != nil {
				return nil, fmt.Errorf("encounter failed: %w", err)
			}
			victory = vic
			if victory != core.VictoryStatusNone {
				break
			}
		}

		totalRounds += ed.CurrentRound
		encLogs := make([]events.TimelineEvent, 0)
		if req.IncludeLogs {
			encLogs = ed.ExportTimeline()
		}

		encounterResults = append(encounterResults, IndividualEncounterResult{
			EncounterName: encCfg.Name,
			VictoryStatus: victory,
			Rounds:        ed.CurrentRound,
			Seed:          encSeed,
			InitialState:  initialState,
			Logs:          encLogs,
		})

		// If party wiped, we stop the adventuring day, but we might want to continue
		// if the user specifically wants to see all encounters (e.g. for testing).
		// However, logically the day ends.
		if victory == core.VictoryStatusMonsters {
			break
		}

		encountersWon++

		// Intermission
		healing := im.ProcessIntermission(characters, req.Intermission)
		if req.IncludeLogs && len(healing) > 0 {
			ed.LogEvent(events.EventIntermissionHealing, nil, map[string]interface{}{
				"healing": healing,
			})
			// Since we already captured logs, append the healing event specifically
			if len(encounterResults) > 0 {
				encounterResults[len(encounterResults)-1].Logs = append(encounterResults[len(encounterResults)-1].Logs, ed.CombatLog[len(ed.CombatLog)-1])
			}
		}
	}

	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	finalStates := make(map[int]actor.Actor)
	for _, char := range characters {
		finalStates[char.InstanceID] = *char
	}

	return &AdventuringDayResult{
		TotalEncounters:  len(req.Encounters),
		EncountersWon:    encountersWon,
		SuccessRate:      float64(encountersWon) / float64(len(req.Encounters)) * 100,
		AverageRounds:    float64(totalRounds) / float64(len(req.Encounters)),
		EncounterResults: encounterResults,
		FinalActorStates: finalStates,
		ActorConfigs:     actorConfigs,
		Performance: &PerformanceMetrics{
			ExecutionTimeMs:    time.Since(startTime).Milliseconds(),
			ExecutionTimeHuman: time.Since(startTime).String(),
			MemoryAllocatedMb:  float64(memEnd.TotalAlloc-memStart.TotalAlloc) / 1024 / 1024,
			PeakGoroutines:     runtime.NumGoroutine(),
		},
	}, nil
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
	for _, res := range chars {
		res.EncounterResults = nil
		res.LogsStripped = true
		result = append(result, res)
	}
	for _, res := range monsters {
		res.EncounterResults = nil
		res.LogsStripped = true
		result = append(result, res)
	}
	for _, res := range others {
		res.EncounterResults = nil
		res.LogsStripped = true
		result = append(result, res)
	}

	return result
}
