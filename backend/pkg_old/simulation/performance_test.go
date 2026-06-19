package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg_old/character"
	"dnd5e-encounter-simulator-backend/pkg_old/classes"
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg_old/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg_old/core/martial_attack_manager"
	"dnd5e-encounter-simulator-backend/pkg_old/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg_old/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg_old/monster"
	"dnd5e-encounter-simulator-backend/pkg_old/races"
	"dnd5e-encounter-simulator-backend/pkg_old/weapon"
	"fmt"
	"math/rand/v2"
	"testing"
)

// getMockCharacterConfig returns a basic character config for benchmarking.
func getMockCharacterConfig(id string, name string) character.CharacterConfig {
	return character.CharacterConfig{
		ID:    id,
		Name:  name,
		Level: 5,
		AC:    15,
		AsConfig: core.AbilityScoresConfig{
			AbilityScores: core.NewAbilityScores(16, 14, 14, 10, 12, 8),
		},
		HPConfig: core.HPConfig{
			HPMethod:     core.HPSetAverage,
			Value:        40,
			HPAverage:    40,
			NumberOfDice: 5,
			HitDie:       core.D10,
		},
		UtilityWeights: &core.UtilityWeights{
			ActionWeights: map[core.ActionType]float64{
				core.ATMelee: 1.0,
			},
		},
	}
}

// getMockMonsterConfig returns a basic monster config for benchmarking.
func getMockMonsterConfig(id int, name string) monster.MonsterConfig {
	return monster.MonsterConfig{
		MonsterBase: monster.MonsterBase{
			ID:               id,
			Name:             name,
			AC:               13,
			CR:               2,
			ProficiencyBonus: 2,
			AbilityScores:    core.NewAbilityScores(15, 12, 14, 10, 10, 8),
			HP: core.HPConfig{
				HPMethod:     core.HPSetAverage,
				Value:        30,
				HPAverage:    30,
				NumberOfDice: 4,
				HitDie:       core.D8,
			},
		},
		Actions: map[int]monster.Action{
			1: {
				ActionID:    1,
				Name:        "Claw",
				Index:       0,
				AttackBonus: 4,
				DamageBlocks: []core.DamageBlock{
					{
						NumberOfDice: 1,
						Die:          core.D6,
						Modifier:     2,
						DamageType:   core.DamageSlashing,
					},
				},
			},
		},
		HPMethod: core.HPSetAverage,
		UtilityWeights: &core.UtilityWeights{
			ActionWeights: map[core.ActionType]float64{
				core.ATDamage: 1.0,
			},
		},
	}
}

// benchmarkSimulation runs a simulation with given parameters.
func benchmarkSimulation(b *testing.B, numChars, numMonsters, numRuns int, includeLogs bool) {
	charConfigs := make([]character.CharacterConfig, numChars)
	for i := 0; i < numChars; i++ {
		charConfigs[i] = getMockCharacterConfig(fmt.Sprintf("%d", i+1), fmt.Sprintf("Hero %d", i+1))
	}

	monsterConfigs := make([]monster.MonsterConfig, numMonsters)
	for i := 0; i < numMonsters; i++ {
		monsterConfigs[i] = getMockMonsterConfig(i+1, fmt.Sprintf("Monster %d", i+1))
	}

	req := MultiSimulationRequest{
		CharacterConfigs: charConfigs,
		MonsterConfigs:   monsterConfigs,
		NumberOfRuns:     numRuns,
		MaxRounds:        20,
		IncludeLogs:      includeLogs,
		BaseOptions: core.SimulationOptions{
			UseWeightedAI: true,
		},
	}

	ctx := context.Background()

	setup := func(sm *SimulationManager) error {
		// Mock Character Creation
		for _, cfg := range req.CharacterConfigs {
			char := &character.Character{
				Name:          cfg.Name,
				Level:         cfg.Level,
				AbilityScores: cfg.AsConfig.AbilityScores,
				Class: classes.Class{
					ID:          classes.Fighter,
					Name:        "Fighter",
					HitDie:      core.D10,
					AttackCount: 1,
				},
				Race: races.Race{
					ID:   races.Human,
					Name: "Human",
				},
				HPConfig:      cfg.HPConfig,
				Configuration: cfg.EntityConfiguration,
			}
			// Use internal initializers to set up managers properly
			char.RollManager = roll_manager.NewRollManager(char, roll_manager.RerollAbilities{})
			char.RollManager.SetRNG(sm.rng.Uint64(), sm.rng.Uint64())
			char.EntityStateManager, _ = entity_state_manager.NewEntityStateManager(char, entity_state_manager.EntityStateConfig{
				MaxHP:       cfg.HPConfig.Value,
				CurrentHP:   cfg.HPConfig.Value,
				AttackCount: 1,
			})
			char.EntityStateManager.SetHitDie(cfg.HPConfig.HitDie)
			char.SpellCastingManager = &spellcasting_manager.SpellcastingManager{}
			char.RNG = rand.New(rand.NewPCG(sm.rng.Uint64(), sm.rng.Uint64()))
			char.AI = character.NewCharacterAI(char, cfg.UtilityWeights)
			char.MartialAttackManager = martial_attack_manager.NewMartialAttackManager(char, char.RollManager)
			char.EquipmentManager, _ = equipment_manager.NewEquipmentManager(char)

			// Add a mock weapon
			w, _ := weapon.New("Longsword", false, false, 1, core.D8, core.DamageSlashing, weapon.Properties{})
			char.EquipmentManager.SetWeapon(core.WSPrimary, &w, true)

			sm.combatEngine.AddCombatant(core.NewCombatantWithInfo(char))
		}

		// Mock Monster Creation (using the csm but with config)
		csm := NewCombatantSetupManager(ctx, false, true, sm.rng)
		setupRes, err := csm.SetupCombatants(nil, nil, req.MonsterConfigs)
		if err != nil {
			return err
		}
		for _, c := range setupRes.Combatants {
			sm.combatEngine.AddCombatant(c)
		}

		sm.InitializeCombatants()
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := RunMultiSimulationWithSetup(ctx, req, setup)
		if err != nil {
			b.Fatalf("Simulation failed: %v", err)
		}
	}
}

func BenchmarkSimulation_1v1_10Runs_WithLogs(b *testing.B) { benchmarkSimulation(b, 1, 1, 10, true) }
func BenchmarkSimulation_1v1_10Runs_NoLogs(b *testing.B)   { benchmarkSimulation(b, 1, 1, 10, false) }

func BenchmarkSimulation_1v1_100Runs(b *testing.B)           { benchmarkSimulation(b, 1, 1, 100, true) }
func BenchmarkSimulation_4v4_10Runs(b *testing.B)            { benchmarkSimulation(b, 4, 4, 10, true) }
func BenchmarkSimulation_4v4_100Runs(b *testing.B)           { benchmarkSimulation(b, 4, 4, 100, true) }
func BenchmarkSimulation_Large_4v10_10Runs(b *testing.B)     { benchmarkSimulation(b, 4, 10, 10, true) }
func BenchmarkSimulation_Large_4v10_50Runs(b *testing.B)     { benchmarkSimulation(b, 4, 10, 50, true) }
func BenchmarkSimulation_VeryLarge_8v15_10Runs(b *testing.B) { benchmarkSimulation(b, 8, 15, 10, true) }
func BenchmarkSimulation_VeryLarge_8v15_50Runs(b *testing.B) { benchmarkSimulation(b, 8, 15, 50, true) }
func BenchmarkSimulation_TooBig_8v15_100Runs(b *testing.B)   { benchmarkSimulation(b, 8, 50, 100, true) }
