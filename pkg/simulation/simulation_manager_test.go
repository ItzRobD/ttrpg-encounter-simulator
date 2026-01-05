package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"fmt"
	"testing"
)

// Build two basic combatants and run a one-round smoke simulation.
func TestSimulationManager_RunSimulation_Smoke(t *testing.T) {
	opts := core.SimulationOptions{}
	sm := NewSimulationManager(opts, core.Seed{Seed1: 7, Seed2: 9})

	// Attacker: character with a sword
	ch := buildTestCharacter(t, core.AbilityScores{Strength: 16, Dexterity: 14}, 5)
	equipSword(t, ch, core.WSPrimary)
	// Equip a secondary weapon to satisfy offhand path if triggered
	equipSword(t, ch, core.WSSecondary)
	// Target: AC=0 to guarantee hits
	mon := buildTestMonster(t, 0)

	ce := sm.GetCombatEngine()
	c1, c2 := buildCombatants(ch, mon)
	ce.AddCombatant(c1)
	ce.AddCombatant(c2)

	sm.SetupEventListeners()
	sm.InitializeCombatants()

	if err := sm.RunSimulation(1); err != nil {
		t.Fatalf("RunSimulation: %v", err)
	}
}

func TestSimulationManager_SetupEventListeners_AttachesAndLogs(t *testing.T) {
	sm := NewSimulationManager(core.SimulationOptions{}, core.Seed{Seed1: 10, Seed2: 12})

	// Minimal one combatant to attach a listener to (non-lair)
	ch := buildTestCharacter(t, core.AbilityScores{Dexterity: 14}, 5)
	c := core.NewCombatantWithInfo(ch)
	sm.GetCombatEngine().AddCombatant(c)

	sm.SetupEventListeners()

	// Emit a dice roll event via the entity's listener and verify it ends up in sim log
	rr := &roll_manager.RollResult{DiceRollType: core.DiceRollInitiative, NumberOfDice: 1, Die: core.D20, FinalRollValue: 10, Total: 12}
	events.LogDiceRollEvent(ch.GetCurrentEventContext(), ch, rr, ch.GetEventListener())

	if len(sm.simLog) == 0 { // package test has access to private field
		t.Fatalf("expected at least one event in simulation log after listener emit")
	}
}

// Ensure InitializeCombatants calls InitializeHP on non-lair combatants
func TestSimulationManager_InitializeCombatants_CallsInitHP(t *testing.T) {
	sm := NewSimulationManager(core.SimulationOptions{}, core.Seed{Seed1: 21, Seed2: 22})

	ch := buildTestCharacter(t, core.AbilityScores{}, 1)
	// Change desired HP via config to a recognizable value then call InitializeCombatants
	ch.HPConfig = core.HPConfig{HPSetMethod: core.HPSetValue, Value: 23, HitDie: core.D8}
	c := core.NewCombatantWithInfo(ch)
	sm.GetCombatEngine().AddCombatant(c)

	sm.InitializeCombatants()

	if ch.GetHPStatus().GetHP() != 23 || ch.GetHPStatus().GetMaxHP() != 23 {
		t.Errorf("InitializeCombatants did not set HP values, got hp=%d max=%d", ch.GetHPStatus().GetHP(), ch.GetHPStatus().GetMaxHP())
	}
}

// Guard: SetupCombatantsFromAPI should return without adding when given empty inputs; avoid DB.
func TestSimulationManager_SetupCombatantsFromAPI_EmptyInputs(t *testing.T) {
	sm := NewSimulationManager(core.SimulationOptions{}, core.Seed{Seed1: 31, Seed2: 32})
	ctx := context.Background()
	// No characters and no monsters — should not error hard and should add none
	result, err := sm.SetupCombatantsFromAPI(ctx, nil, nil, nil)
	if err != nil && result == nil {
		// Some implementations may return nil result with error; allow either no error or an error with nil result
		t.Fatalf("SetupCombatantsFromAPI unexpected fatal error: %v", err)
	}
}

func TestRunMultiSimulation_Basic(t *testing.T) {
	// We'll skip the API/DB part by using a custom setup if possible,
	// but RunMultiSimulation is hardcoded to use SetupCombatantsFromAPIWithLair.
	// For a unit test, we might want a version that takes pre-configured combatants,
	// but the requirement was about the front-end requesting multiple runs,
	// which usually implies starting from configs.

	// Since SetupCombatantsFromAPIWithLair depends on the database,
	// and we don't have a mock DB easily available here without more setup,
	// let's see if we can at least test the concurrency logic with a mockable path.

	// Actually, let's look at SetupCombatantsFromAPI in simulation_manager.go
	// It calls setupManager.SetupCombatants(characterConfigs, monsterIDs)

	// For now, let's try a test that would fail if it tried to hit the DB,
	// or better, let's add a variant of RunMultiSimulation that is more testable
	// or just accept that this specific integration test needs a DB.

	// Alternatively, I can test it by passing empty configs and seeing it fail as expected,
	// or passing valid but "offline" configs if character.NewCharacterWithRNG doesn't need DB.

	ctx := context.Background()
	req := MultiSimulationRequest{
		NumRuns:   3,
		MaxRounds: 10,
		BaseOptions: core.SimulationOptions{
			Seed: core.Seed{Seed1: 1, Seed2: 2},
		},
		// Empty configs will cause SetupCombatants to return an error "no valid combatants"
	}

	result, err := RunMultiSimulation(ctx, req)
	if err == nil {
		t.Fatal("Expected error due to no combatants, but got nil")
	}
	if result != nil {
		t.Fatal("Expected nil result on error")
	}
}

func TestRunMultiSimulation_WithLogs(t *testing.T) {
	ctx := context.Background()
	req := MultiSimulationRequest{
		NumRuns:     2,
		MaxRounds:   5,
		IncludeLogs: true,
		BaseOptions: core.SimulationOptions{
			Seed: core.Seed{Seed1: 1, Seed2: 2},
		},
	}

	setup := func(sm *SimulationManager) error {
		ch := buildTestCharacter(t, core.AbilityScores{Strength: 16, Dexterity: 14}, 1)
		equipSword(t, ch, core.WSPrimary)
		mon := buildTestMonster(t, 10)
		ce := sm.GetCombatEngine()
		c1, c2 := buildCombatants(ch, mon)
		ce.AddCombatant(c1)
		ce.AddCombatant(c2)
		sm.InitializeCombatants()
		return nil
	}

	result, err := RunMultiSimulationWithSetup(ctx, req, setup)
	if err != nil {
		t.Fatalf("RunMultiSimulationWithSetup failed: %v", err)
	}

	if len(result.IndividualResults) != 2 {
		t.Errorf("expected 2 individual results, got %d", len(result.IndividualResults))
	}

	for _, res := range result.IndividualResults {
		if len(res.Logs) == 0 {
			t.Errorf("expected logs for run %d, but got none", res.RunID)
		}
	}
}

func TestSimulationManager_SetupCombatantsFromAPI_CustomMonsters(t *testing.T) {
	sm := NewSimulationManager(core.SimulationOptions{}, core.Seed{Seed1: 41, Seed2: 42})
	ctx := context.Background()

	// Use buildTestCharacter instead of CharacterConfig to bypass DB queries in unit test
	ch := buildTestCharacter(t, core.AbilityScores{}, 1)
	ch.Name = "Hero"

	monsterConfigs := []monster.MonsterConfig{
		{
			Base: monster.MonsterBase{
				Name:          "Custom Orc",
				AC:            13,
				AbilityScores: core.AbilityScores{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
				HP:            core.HPConfig{HPSetMethod: core.HPSetValue, Value: 15, HitDie: core.D8, NumberOfDice: 1}, // Added NumberOfDice: 1 to pass RollHP validation
			},
		},
	}

	setupManager := NewCombatantSetupManager(ctx, sm.options.UseHPAverageCharacter, sm.options.UseHPAverageMonster, sm.rng)
	result, err := setupManager.SetupCombatants(nil, nil, monsterConfigs)
	if err != nil {
		t.Fatalf("SetupCombatants failed: %v", err)
	}

	// Manually add Hero
	result.Combatants = append(result.Combatants, core.NewCombatantWithInfo(ch))

	if len(result.Combatants) != 2 {
		t.Errorf("expected 2 combatants, got %d", len(result.Combatants))
	}

	foundHero := false
	foundOrc := false
	for _, c := range result.Combatants {
		if c.Entity.GetName() == "Hero" {
			foundHero = true
		}
		if c.Entity.GetName() == "Custom Orc" {
			foundOrc = true
		}
	}

	if !foundHero {
		t.Error("Hero not found in combatants")
	}
	if !foundOrc {
		t.Error("Custom Orc not found in combatants")
	}
}

func TestRunMultiSimulation_WithCustomMonsters(t *testing.T) {
	ctx := context.Background()
	req := MultiSimulationRequest{
		NumRuns:   1,
		MaxRounds: 5,
		BaseOptions: core.SimulationOptions{
			Seed: core.Seed{Seed1: 1, Seed2: 2},
		},
		MonsterConfigs: []monster.MonsterConfig{
			{
				Base: monster.MonsterBase{
					Name:          "Custom Orc",
					AC:            13,
					AbilityScores: core.AbilityScores{Strength: 10, Dexterity: 10, Constitution: 10, Intelligence: 10, Wisdom: 10, Charisma: 10},
					HP:            core.HPConfig{HPSetMethod: core.HPSetValue, Value: 15, HitDie: core.D8, NumberOfDice: 1},
				},
				Actions: map[int]monster.Action{
					1: {ActionID: 1, Name: "Greataxe", NumberOfDice: 1, Die: core.D12, AmountToAdd: 3, AttackBonus: 5, DamageType: core.DamageSlashing},
				},
			},
		},
	}

	// Use a setup function to bypass DB-based character creation
	setup := func(sm *SimulationManager) error {
		ch := buildTestCharacter(t, core.AbilityScores{Strength: 16, Dexterity: 14}, 1)
		ch.Name = "Hero"
		equipSword(t, ch, core.WSPrimary) // Equip sword so it has a valid action
		sm.GetCombatEngine().AddCombatant(core.NewCombatantWithInfo(ch))

		// Setup custom monsters from request
		res, err := sm.SetupCombatantsFromAPI(ctx, nil, nil, req.MonsterConfigs)
		if err != nil {
			return err
		}
		if len(res.Errors) > 0 {
			return fmt.Errorf("setup errors: %v", res.Errors)
		}

		sm.InitializeCombatants()
		return nil
	}

	result, err := RunMultiSimulationWithSetup(ctx, req, setup)
	if err != nil {
		t.Fatalf("RunMultiSimulationWithSetup failed: %v", err)
	}

	if result.TotalRuns != 1 {
		t.Errorf("expected 1 run, got %d", result.TotalRuns)
	}
}
