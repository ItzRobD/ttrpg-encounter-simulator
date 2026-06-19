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
	"runtime/debug"
	"time"
)

// MultiSimulationRequest defines the parameters for running multiple adventuring day simulations.
type MultiSimulationRequest struct {
	AdventuringDayRequest
	NumberOfRuns int `json:"number_of_runs"`
}

// MultiSimulationResult aggregates the results of multiple adventuring day simulation runs.
type MultiSimulationResult struct {
	TotalRuns           int                          `json:"total_runs"`
	CharacterVictories  int                          `json:"character_victories"` // Entire day won
	MonsterVictories    int                          `json:"monster_victories"`   // Party wiped at some point
	OtherVictories      int                          `json:"other_victories"`     // Draw or other
	AverageRounds       float64                      `json:"average_rounds"`
	WinRatePercentage   float64                      `json:"win_rate_percentage"`
	ActorConfigs        map[int]actor.ActorConfig    `json:"actor_configs,omitempty"`
	IndividualResults   []IndividualSimulationResult `json:"individual_results,omitempty"`
	Performance         *PerformanceMetrics          `json:"performance,omitempty"`
	AggregateStatistics map[int]*CombatStatistics    `json:"aggregate_statistics,omitempty"`
}

type PerformanceMetrics struct {
	ExecutionTimeMs    int64   `json:"execution_time_ms"`
	ExecutionTimeHuman string  `json:"execution_time_human"`
	MemoryAllocatedMb  float64 `json:"memory_allocated_mb"`
	PeakGoroutines     int     `json:"peak_goroutines"`
}

// IndividualSimulationResult holds data for a single adventuring day simulation run.
type IndividualSimulationResult struct {
	RunID               int                         `json:"run_id"`
	VictoryStatus       core.VictoryStatus          `json:"victory_status"`
	TotalRounds         int                         `json:"total_rounds"`
	Seed                core.Seed                   `json:"seed"`
	LogsStripped        bool                        `json:"logs_stripped,omitempty"`
	EncounterResults    []IndividualEncounterResult `json:"encounter_results,omitempty"`
	ActorConfigs        map[int]actor.ActorConfig   `json:"actor_configs,omitempty"`
	AggregateStatistics map[int]*CombatStatistics   `json:"aggregate_statistics,omitempty"`
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
	TotalEncounters     int                         `json:"total_encounters"`
	EncountersWon       int                         `json:"encounters_won"`
	SuccessRate         float64                     `json:"success_rate"`
	AverageRounds       float64                     `json:"average_rounds"`
	EncounterResults    []IndividualEncounterResult `json:"encounter_results,omitempty"`
	FinalActorStates    map[int]actor.Actor         `json:"final_actor_states,omitempty"`
	ActorConfigs        map[int]actor.ActorConfig   `json:"actor_configs,omitempty"`
	Performance         *PerformanceMetrics         `json:"performance,omitempty"`
	AggregateStatistics map[int]*CombatStatistics   `json:"aggregate_statistics,omitempty"`
	ShortRestsTaken     int                         `json:"short_rests_taken"`
}

type IndividualEncounterResult struct {
	EncounterName string                    `json:"encounter_name"`
	VictoryStatus core.VictoryStatus        `json:"victory_status"`
	Rounds        int                       `json:"rounds"`
	Seed          core.Seed                 `json:"seed"`
	InitialState  map[int]ActorInitialState `json:"initial_state,omitempty"`
	Logs          []events.TimelineEvent    `json:"logs,omitempty"`
	Statistics    map[int]*CombatStatistics `json:"statistics,omitempty"`
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

			// Recover from a panic in a single run so it fails that run via
			// errChan instead of crashing the whole batch's goroutine.
			defer func() {
				if r := recover(); r != nil {
					errChan <- fmt.Errorf("run %d panicked: %v\n%s", runID, r, debug.Stack())
				}
			}()

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
				RunID:               runID,
				VictoryStatus:       dayVictory,
				TotalRounds:         totalRounds,
				Seed:                seed,
				EncounterResults:    dayRes.EncounterResults,
				ActorConfigs:        dayRes.ActorConfigs,
				AggregateStatistics: dayRes.AggregateStatistics,
			}

			resultsChan <- res
		}(i, daySeed)
	}

	multiResult := &MultiSimulationResult{
		TotalRuns:           req.NumberOfRuns,
		IndividualResults:   make([]IndividualSimulationResult, 0, req.NumberOfRuns),
		ActorConfigs:        make(map[int]actor.ActorConfig),
		AggregateStatistics: make(map[int]*CombatStatistics),
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

			// Aggregate statistics from all runs into the multi-result aggregate
			mergeCombatStatistics(multiResult.AggregateStatistics, res.AggregateStatistics)

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

		// Update averages for multiResult aggregate stats
		for _, stats := range multiResult.AggregateStatistics {
			stats.AverageDamagePerRound = float64(stats.TotalDamageDealt) / float64(totalRounds)
			stats.AverageHealingPerRound = float64(stats.TotalHealingDone) / float64(totalRounds)

			// Per run averages
			stats.AverageDamageDealtPerRun = float64(stats.TotalDamageDealt) / float64(multiResult.TotalRuns)
			stats.AverageHealingDonePerRun = float64(stats.TotalHealingDone) / float64(multiResult.TotalRuns)
			stats.AverageDamageTakenPerRun = float64(stats.TotalDamageTaken) / float64(multiResult.TotalRuns)
			stats.AverageHealingReceivedPerRun = float64(stats.TotalHealingReceived) / float64(multiResult.TotalRuns)
			stats.AverageAttacksMadePerRun = float64(stats.AttacksMade) / float64(multiResult.TotalRuns)
			stats.AverageAttacksHitPerRun = float64(stats.AttacksHit) / float64(multiResult.TotalRuns)
			stats.AverageSpellsUsedPerRun = float64(stats.SpellsUsed) / float64(multiResult.TotalRuns)
		}
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
		if a == nil {
			return nil, fmt.Errorf("character hydration failed for %s: received nil actor", cfg.Name)
		}
		// Assign InstanceID starting from 1 for characters
		// We ignore any pre-assigned InstanceID from the request to ensure a clean start
		a.InstanceID = nextInstanceID
		nextInstanceID++
		characters = append(characters, a)
		// Capture the fully hydrated config including the assigned InstanceID
		actorConfigs[a.InstanceID] = a.ToConfig()
	}

	type IntermissionStats struct {
		hitDiceUsed     map[core.DiceType]int
		healing         map[int]int
		shortRestsTaken int
	}

	intermissionStats := IntermissionStats{
		hitDiceUsed:     make(map[core.DiceType]int),
		healing:         make(map[int]int),
		shortRestsTaken: 0,
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
			if m == nil {
				return nil, fmt.Errorf("monster hydration failed for ID %d: received nil actor", mID)
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
			if m == nil {
				return nil, fmt.Errorf("monster hydration failed: received nil actor")
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
		var encLogs []events.TimelineEvent
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
			Statistics:    ed.Statistics.statistics,
		})

		// If party wiped, we stop the adventuring day, but we might want to continue
		// if the user specifically wants to see all encounters (e.g. for testing).
		// However, logically the day ends.
		if victory == core.VictoryStatusMonsters {
			break
		}

		encountersWon++

		// Intermission
		res := im.ProcessIntermission(characters, req.Intermission)
		if req.IncludeLogs && len(res.HealingReceived) > 0 {
			ed.LogEvent(events.EventIntermissionHealing, nil, map[string]interface{}{
				"healing": res.HealingReceived,
			})
			// Since we already captured logs, append the healing event specifically
			encounterResults[len(encounterResults)-1].Logs = append(encounterResults[len(encounterResults)-1].Logs, ed.CombatLog[len(ed.CombatLog)-1])
		}

		// Update day-level aggregate stats with intermission data
		// Attach intermission stats to the encounter results' statistics
		// so they get picked up by aggregateEncounterStatistics.
		lastEncStats := encounterResults[len(encounterResults)-1].Statistics
		for id, heal := range res.HealingReceived {
			if stats, ok := lastEncStats[id]; ok {
				stats.IntermissionHealingReceived += heal
			}
		}
		for id, hdMap := range res.HitDiceUsed {
			if stats, ok := lastEncStats[id]; ok {
				if stats.IntermissionHitDiceUsed == nil {
					stats.IntermissionHitDiceUsed = make(map[core.DiceType]int)
				}
				for die, count := range hdMap {
					stats.IntermissionHitDiceUsed[die] += count
				}
			}
		}
		for id, count := range res.SpellsUsed {
			if stats, ok := lastEncStats[id]; ok {
				stats.IntermissionSpellsUsed += count
			}
		}
		for id, slots := range res.SpellSlotsUsed {
			if stats, ok := lastEncStats[id]; ok {
				if stats.IntermissionSpellSlotsUsed == nil {
					stats.IntermissionSpellSlotsUsed = make(map[int]int)
				}
				for lvl, count := range slots {
					stats.IntermissionSpellSlotsUsed[lvl] += count
				}
			}
		}

		intermissionStats.healing = mergeIntMaps(intermissionStats.healing, res.HealingReceived)
		intermissionStats.shortRestsTaken = res.ShortRestsTaken
	}

	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	finalStates := make(map[int]actor.Actor)
	for _, char := range characters {
		finalStates[char.InstanceID] = *char
	}

	// Aggregate individual encounter result statistics
	aggStats := aggregateEncounterStatistics(encounterResults...)

	for _, stats := range aggStats {
		if totalRounds > 0 {
			stats.AverageDamagePerRound = float64(stats.TotalDamageDealt) / float64(totalRounds)
			stats.AverageHealingPerRound = float64(stats.TotalHealingDone) / float64(totalRounds)
		}
		// In a single run, per-run averages are the same as totals
		stats.AverageDamageDealtPerRun = float64(stats.TotalDamageDealt)
		stats.AverageHealingDonePerRun = float64(stats.TotalHealingDone)
		stats.AverageDamageTakenPerRun = float64(stats.TotalDamageTaken)
		stats.AverageHealingReceivedPerRun = float64(stats.TotalHealingReceived)
		stats.AverageAttacksMadePerRun = float64(stats.AttacksMade)
		stats.AverageAttacksHitPerRun = float64(stats.AttacksHit)
		stats.AverageSpellsUsedPerRun = float64(stats.SpellsUsed)
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
		AggregateStatistics: aggStats,
		ShortRestsTaken:     intermissionStats.shortRestsTaken,
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

// mergeIntMaps combines multiple maps by summing the values of identical keys.
func mergeIntMaps(maps ...map[int]int) map[int]int {
	result := make(map[int]int)
	for _, m := range maps {
		for k, v := range m {
			result[k] += v
		}
	}
	return result
}

// mergeDiceMaps combines multiple maps by summing the values of identical keys.
func mergeDiceMaps(maps ...map[core.DiceType]int) map[core.DiceType]int {
	result := make(map[core.DiceType]int)
	for _, m := range maps {
		for k, v := range m {
			result[k] += v
		}
	}
	return result
}

func aggregateEncounterStatistics(encounters ...IndividualEncounterResult) map[int]*CombatStatistics {
	aggStats := make(map[int]*CombatStatistics)
	for _, enc := range encounters {
		mergeCombatStatistics(aggStats, enc.Statistics)
	}
	return aggStats
}

func mergeCombatStatistics(target map[int]*CombatStatistics, source map[int]*CombatStatistics) {
	for instanceID, stats := range source {
		if _, exists := target[instanceID]; !exists {
			target[instanceID] = NewCombatStatistics()
		}

		as := target[instanceID]
		as.TotalDamageDealt += stats.TotalDamageDealt
		as.TotalHealingDone += stats.TotalHealingDone
		as.AttacksMade += stats.AttacksMade
		as.AttacksHit += stats.AttacksHit
		as.AttacksMissed += stats.AttacksMissed
		as.SpellsUsed += stats.SpellsUsed
		as.SpellAttackActions += stats.SpellAttackActions
		as.SpellSaveActions += stats.SpellSaveActions
		as.LegendaryActionsUsed += stats.LegendaryActionsUsed
		as.HealingActions += stats.HealingActions
		as.CriticalHits += stats.CriticalHits
		as.TimesDamaged += stats.TimesDamaged
		as.TimesHealed += stats.TimesHealed
		as.TotalDamageTaken += stats.TotalDamageTaken
		as.TotalHealingReceived += stats.TotalHealingReceived
		as.DeathSaveSuccesses += stats.DeathSaveSuccesses
		as.DeathSaveFailures += stats.DeathSaveFailures

		as.DamageByRound = mergeIntMaps(as.DamageByRound, stats.DamageByRound)
		as.HealingByRound = mergeIntMaps(as.HealingByRound, stats.HealingByRound)
		as.SpellSlotsUsed = mergeIntMaps(as.SpellSlotsUsed, stats.SpellSlotsUsed)

		// Aggregate intermission stats
		as.IntermissionHealingReceived += stats.IntermissionHealingReceived
		as.IntermissionSpellsUsed += stats.IntermissionSpellsUsed
		as.IntermissionHitDiceUsed = mergeDiceMaps(as.IntermissionHitDiceUsed, stats.IntermissionHitDiceUsed)
		as.IntermissionSpellSlotsUsed = mergeIntMaps(as.IntermissionSpellSlotsUsed, stats.IntermissionSpellSlotsUsed)
	}
}
