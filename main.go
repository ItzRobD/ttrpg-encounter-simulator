package main

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/simulation"
	"fmt"
)

func main() {
	dbErr := database.InitDb()

	if dbErr != nil {
		fmt.Println(dbErr)
		return
	}
	defer database.CloseDb()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "CanUpcast", true)
	//params := spells.SpellQueryParams{Name: []string{"Fireball", "Acid Splash"}}
	//s, err := spells.QuerySpellData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//
	//}
	//fmt.Println(s)

	frank := setupFrank()
	testSimulation([]character.CharacterConfig{frank}, []int{2, 4, 52})
}

func setupFrank() character.CharacterConfig {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "CanUpcast", true)
	charConfig := character.CharacterConfig{
		Name:    "Frank",
		ClassID: classes.Wizard,
		Level:   5,
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

	return charConfig
}

func setupMonsters(ctx context.Context, ids []int) ([]core.Combatant, error) {
	var combatants []core.Combatant
	cfg, err := monster.QueryMonsterConfigData(ctx, monster.MonsterQueryParams{ID: ids})
	if err != nil {
		return nil, err
	}

	for _, c := range cfg {
		monster, err2 := monster.NewMonster(ctx, c)
		if err2 != nil {
			return nil, err2
		}
		combatants = append(combatants, core.NewCombatant(monster, 0))
	}
	return combatants, nil
}

func testSimulation(charCfgs []character.CharacterConfig, monsterIds []int) {
	config := simulation.SimulationConfig{
		Seed:                      0,
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
		TargetPriority:            0,
		HealPriority:              0,
		ActionPreference:          0,
		AOEHitsAllEnemies:         false,
		CharacterHealThresholdPct: 0,
		MonsterHealThresholdPct:   0,
	}

	sim := simulation.NewSimulationManager(config, core.Seed{})

	ctx := context.Background()
	ctx = context.WithValue(ctx, "CanUpcast", true)
	sim.SetupCombatantsFromAPI(ctx,
		charCfgs,
		monsterIds)
	sim.SetupEventListeners()
	sim.InitializeCombatants()
	sim.RunSimulation(50)
	sim.GetCombatEngine().PrintCombatTracker()
	//sim.PrintSimulationLog()
}
