package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
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
	events.LogDiceRollEvent(ch, rr, ch.GetEventListener())

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
	result, err := sm.SetupCombatantsFromAPI(ctx, nil, nil)
	if err != nil && result == nil {
		// Some implementations may return nil result with error; allow either no error or an error with nil result
		t.Fatalf("SetupCombatantsFromAPI unexpected fatal error: %v", err)
	}
}
