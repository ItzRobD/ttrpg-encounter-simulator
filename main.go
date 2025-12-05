package main

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/simulation"
	"fmt"
)

// TODO: I completed the ai targeting for characters
// 		Still need to handle turn logic for characters, relatively simple
// 		Need to figure out what to do for monster ai with action selections etc
//		Need to also account for legendary actions, this would be a call every turn
//		Would be smart to have an is legendary monster present bool within the simulation

func main() {
	dbErr := database.InitDb(nil)

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
	//jack := setupJack()
	testSimulation([]character.CharacterConfig{frank}, []int{2})
}

func setupJack() character.CharacterConfig {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "CanUpcast", true)
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
	ctx := context.Background()
	ctx = context.WithValue(ctx, "CanUpcast", true)
	charConfig := character.CharacterConfig{
		Name:    "Frank",
		ClassID: classes.Fighter,
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
	config := core.SimulationOptions{
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
	//sim.InitializeCombatants()
	err := sim.RunSimulation(24)
	if err != nil {
		fmt.Println(err)
	}
	sim.GetCombatEngine().PrintCombatTracker()
	//sim.PrintSimulationLog()
}
