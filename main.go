package main

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/lair"
	"dnd5e-encounter-simulator-backend/pkg/simulation"
	"fmt"
)

// Legacy note: Main is used for local/manual runs. Core TODOs are tracked in todo-triage.md.

func main() {
	dbErr := database.InitDb(nil)

	if dbErr != nil {
		fmt.Println(dbErr)
		return
	}
	defer database.CloseDb()

	// Sample playground for manual queries; keep commented to avoid unused variables
	// ctx := context.Background()
	//params := spells.SpellQueryParams{Name: []string{"Fireball", "Acid Splash"}}
	//s, err := spells.QuerySpellData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//
	//}
	//fmt.Println(s)

	frank := setupFrank()
	//jack := setupJack()
	testSimulation([]character.CharacterConfig{frank}, []int{153})
}

func setupJack() character.CharacterConfig {
	charConfig := character.CharacterConfig{
		Name:    "Jack",
		ClassID: classes.Wizard,
		Level:   4,
		AsConfig: core.AbilityScoresConfig{
			AbilityScores: core.AbilityScores{
				Strength:     18,
				Dexterity:    14,
				Constitution: 16,
				Intelligence: 10,
				Wisdom:       12,
				Charisma:     10,
			},
			Proficiencies: core.AbilityScoresProficiencies{
				Strength:     false,
				Dexterity:    true,
				Constitution: false,
				Intelligence: true,
				Wisdom:       false,
				Charisma:     true,
			},
		},
		HPMethod: core.HPSetRoll,
		HPValue:  0,
		Seed: core.Seed{
			Seed1: 0,
			Seed2: 0,
		},
	}

	charConfig.Equipment = character.EquipmentConfig{
		ArmorID:       5,
		PrimarySlot:   map[int]bool{22: true},
		SecondarySlot: nil,
		RangedSlot:    nil,
	}

	return charConfig
}

func setupFrank() character.CharacterConfig {
	charConfig := character.CharacterConfig{
		Name:    "Frank",
		ClassID: classes.Fighter,
		Level:   4,
		RaceID:  1,
		AsConfig: core.AbilityScoresConfig{
			AbilityScores: core.AbilityScores{
				Strength:     18,
				Dexterity:    14,
				Constitution: 16,
				Intelligence: 10,
				Wisdom:       12,
				Charisma:     10,
			},
			Proficiencies: core.AbilityScoresProficiencies{
				Strength:     false,
				Dexterity:    true,
				Constitution: false,
				Intelligence: true,
				Wisdom:       false,
				Charisma:     true,
			},
		},
		HPMethod: core.HPSetRoll,
		HPValue:  0,
		Seed: core.Seed{
			Seed1: 0,
			Seed2: 0,
		},
		Resistances: core.NewDamageResistances(),
		// Mocked Fighting Styles as if sent by a front-end request
		FightingStyles: []classes.FightingStyle{
			// Enable/disable as needed for local testing
			//classes.StyleDueling,
			// classes.StyleGWF,
			//classes.StyleArchery,
			// Common test: allow offhand to add ability mod to damage
			//classes.StyleTWF,
		},
	}

	charConfig.Equipment = character.EquipmentConfig{
		ArmorID:       5,
		PrimarySlot:   map[int]bool{22: true},
		SecondarySlot: map[int]bool{2: false},
		RangedSlot:    nil,
	}

	charConfig.Resistances.SetResistance(core.DamageBludgeoning, core.ResistanceResistant, nil)

	return charConfig
}

//func setupMonsters(ctx context.Context, ids []int) ([]core.Combatant, error) {
//	var combatants []core.Combatant
//	cfg, err := monster.QueryMonsterConfigData(ctx, monster.MonsterQueryParams{ID: ids})
//	if err != nil {
//		return nil, err
//	}
//
//	for _, c := range cfg {
//		monster, err2 := monster.NewMonster(ctx, c)
//		if err2 != nil {
//			return nil, err2
//		}
//		combatants = append(combatants, core.NewCombatant(monster, 0))
//	}
//	return combatants, nil
//}

func testSimulation(charCfgs []character.CharacterConfig, monsterIds []int) {
	seed := core.Seed{Seed1: 11, Seed2: 22}
	config := core.SimulationOptions{
		Seed:                      seed,
		UseHPAverageCharacter:     false,
		UseHPAverageMonster:       false,
		CanMonstersCrit:           false,
		CanCharactersCrit:         false,
		HasIncreasedCrits:         false,
		UseImprovedCriticals:      false,
		CharactersAlwaysUpcast:    false,
		MonstersAlwaysUpcast:      false,
		AllowCharacterHeals:       false,
		AllowMonsterHeals:         false,
		AOEHitsAllEnemies:         false,
		CharacterHealThresholdPct: 0,
		MonsterHealThresholdPct:   0,
		AllowLairActions:          false,
	}

	sim := simulation.NewSimulationManager(config, seed)

	ctx := context.Background()
	// Example lair configuration for local runs (simulates what the API will pass later)
	lc := &lair.LairConfig{
		Enabled:    true,
		Name:       "Goblin Warrens",
		Initiative: 20, // lair always acts at 20; ties auto-loss handled in engine
		Actions: []lair.LairActionInput{
			{
				Name:         "Falling Rubble",
				Mode:         lair.LAMAttack,
				TargetSide:   lair.TargetCharacters,
				TargetPolicy: "lowest max hp",
				IsAOE:        false,
				Recharge:     0,
				AttackBonus:  5,
				NumberOfDice: 2,
				Die:          core.D8,
				AmountToAdd:  0,
				DamageType:   core.DamageBludgeoning,
			},
			{
				Name:         "Scalding Steam",
				Mode:         lair.LAMDC,
				TargetSide:   lair.TargetCharacters,
				TargetPolicy: "lowest max hp",
				IsAOE:        true,
				Recharge:     5, // recharges on 5-6
				DCAbility:    core.AbilityDexterity,
				DCValue:      12,
				OnSuccess:    core.DCOnSuccessHalf,
				NumberOfDice: 2,
				Die:          core.D6,
				AmountToAdd:  0,
				DamageType:   core.DamageFire,
			},
		},
	}
	sim.SetupCombatantsFromAPIWithLair(ctx,
		charCfgs,
		monsterIds,
		lc)
	// Event listeners will also be attached inside RunSimulation after
	// SetupCombat so the lair (inserted at init 20) gets a listener too.
	sim.SetupEventListeners()
	//sim.InitializeCombatants()
	err := sim.RunSimulation(24)
	if err != nil {
		fmt.Println(err)
	}
	sim.GetCombatEngine().PrintCombatTracker()
	//sim.PrintSimulationLog()
}
