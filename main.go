package main

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/lair"
	"dnd5e-encounter-simulator-backend/pkg/races"
	"dnd5e-encounter-simulator-backend/pkg/simulation"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
)

// Legacy note: Main is used for local/manual runs. Core TODOs are tracked in todo-triage.md.

// TODO: Add spell ranges|touch

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
	//bob := setupBob()
	testSimulation([]character.CharacterConfig{frank}, []int{64, 64, 64})
}

func setupBob() character.CharacterConfig {
	color := races.DragonbornGold
	charConfig := character.CharacterConfig{
		Name:            "Bob",
		ClassID:         classes.Fighter,
		Level:           3,
		RaceID:          races.Dragonborn,
		DragonbornColor: &color,
		AsConfig: core.AbilityScoresConfig{
			AbilityScores: core.AbilityScores{
				Strength:     12,
				Dexterity:    10,
				Constitution: 8,
				Intelligence: 10,
				Wisdom:       10,
				Charisma:     10,
			},
			Proficiencies: core.AbilityScoresProficiencies{},
		},
		HPMethod: core.HPSetValue,
		HPValue:  150,
		Seed: core.Seed{
			Seed1: 0,
			Seed2: 0,
		},
		Resistances: core.NewDamageResistances(),
	}

	wMods := weapon.Modifiers{
		IsMagic:          true,
		IsSilvered:       false,
		IsAdamantine:     true,
		IsColdForgedIron: false,
		AttackBonus:      0,
		DamageBonus:      0,
	}
	charConfig.Equipment = character.EquipmentConfig{
		ArmorID:     1,
		PrimarySlot: []character.WeaponSlotConfig{{WeaponID: 22, IsProficient: true, Modifiers: &wMods}},
	}

	return charConfig
}

func setupFrank() character.CharacterConfig {
	charConfig := character.CharacterConfig{
		Name:    "Frank",
		ClassID: classes.Wizard,
		Level:   5,
		RaceID:  1,
		AsConfig: core.AbilityScoresConfig{
			AbilityScores: core.AbilityScores{
				Strength:     14,
				Dexterity:    14,
				Constitution: 14,
				Intelligence: 18,
				Wisdom:       18,
				Charisma:     12,
			},
			Proficiencies: core.AbilityScoresProficiencies{
				Strength:     false,
				Dexterity:    false,
				Constitution: false,
				Intelligence: false,
				Wisdom:       true,
				Charisma:     true,
			},
		},
		HPMethod: core.HPSetValue,
		HPValue:  200,
		Seed: core.Seed{
			Seed1: 0,
			Seed2: 0,
		},
		Resistances: core.NewDamageResistances(),
		KnownSpells: []string{"Fireball", "Lightning Bolt"},
	}

	charConfig.Equipment = character.EquipmentConfig{
		ArmorID:     5,
		PrimarySlot: []character.WeaponSlotConfig{{WeaponID: 1, IsProficient: true}},
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
	seed := core.Seed{Seed1: 42, Seed2: 42}
	config := core.SimulationOptions{
		Seed:                      seed,
		UseHPAverageCharacter:     false,
		UseHPAverageMonster:       false,
		CanMonstersCrit:           true,
		CanCharactersCrit:         true,
		HasIncreasedCrits:         false,
		UseImprovedCriticals:      false,
		CharactersAlwaysUpcast:    false,
		MonstersAlwaysUpcast:      false,
		AllowCharacterHeals:       true,
		AllowMonsterHeals:         true,
		AOEHitsAllEnemies:         true,
		CharacterHealThresholdPct: 50,
		MonsterHealThresholdPct:   50,
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
