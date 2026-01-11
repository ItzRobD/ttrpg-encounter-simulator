package main

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/lair"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/races"
	"dnd5e-encounter-simulator-backend/pkg/simulation"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"encoding/json"
	"fmt"
	"os"
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

	//frank := setupFrank()
	bob := setupBob()
	testSimulation([]character.CharacterConfig{bob}, []int{2})
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
		HPValue:  10, // Low HP to trigger unconsciousness quickly
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
	weights := &core.UtilityWeights{
		ActionWeights: map[core.ActionType]float64{
			core.ATDamage: 1.0,
			core.ATHeal:   1.5,
		},
	}
	weights.TargetFactorWeights.HighThreat = 0.8
	weights.TargetFactorWeights.TargetPotency = 0.7
	weights.TargetFactorWeights.TargetHitability = 0.4
	weights.TargetFactorWeights.LowHP = 0.6
	weights.TargetFactorWeights.Vengeance = 0.5
	weights.TargetFactorWeights.ConcentrationBreak = 0.9
	weights.TargetFactorWeights.ElitePriority = 0.8
	weights.TargetFactorWeights.EmergencyHeal = 1.0
	weights.ResourceExpenditureWeight = 0.7

	charConfig := character.CharacterConfig{
		Name:           "Frank",
		ClassID:        classes.Wizard,
		Level:          5,
		RaceID:         1,
		UtilityWeights: weights,
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

//func setupMonsters(ctx context.ctx, ids []int) ([]core.Combatant, error) {
//	var combatants []core.Combatant
//	cfg, err := monster.QueryMonsterConfigData(ctx, monster.MonsterQueryParams{id: ids})
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
	// Standard Monster Weights (Default)
	standardMonsterWeights := &core.UtilityWeights{
		ActionWeights: map[core.ActionType]float64{
			core.ATMonsterDamage: 1.0,
			core.ATMonsterHeal:   1.2,
		},
	}
	standardMonsterWeights.TargetFactorWeights.HighThreat = 0.5
	standardMonsterWeights.TargetFactorWeights.TargetPotency = 0.6
	standardMonsterWeights.TargetFactorWeights.TargetHitability = 0.8
	standardMonsterWeights.TargetFactorWeights.LowHP = 0.7
	standardMonsterWeights.TargetFactorWeights.Vengeance = 0.4
	standardMonsterWeights.TargetFactorWeights.EmergencyHeal = 1.0

	seed := core.Seed{Seed1: 42, Seed2: 42}
	config := core.SimulationOptions{
		Seed:                          seed,
		UseHPAverageMonster:           false,
		UseHPAverageCharacter:         false,
		CanMonstersCrit:               true,
		CanCharactersCrit:             true,
		HasIncreasedCrits:             false,
		UseImprovedCriticals:          false,
		CharactersAlwaysUpcast:        false,
		MonstersAlwaysUpcast:          false,
		AllowCharacterHeals:           true,
		AllowMonsterHeals:             true,
		AOEHitsAllEnemies:             true,
		CharacterHealThresholdPct:     50,
		MonsterHealThresholdPct:       50,
		LimitedLegendaryActions:       false,
		AllowLairActions:              false,
		AllowDragonbornBreathAttack:   true,
		EnableClassFeatures:           false,
		EnableRacialFeatures:          false,
		BarbarianAlwaysRecklessAttack: false,
		PaladinAlwaysSmite:            false,
		PaladinUseHighestSmiteSlot:    false,
		UseMassiveDamage:              false,
		EnableSpecialAbilities:        false,
		MonsterDeathEffectsHitAllies:  false,
		AlwaysUseSneakAttack:          false,
		UseWeightedAI:                 true,
		DebugAI:                       false,
		HPVisibilityMode:              core.HPVisibilityWhite,
		EnableMonsterNoise:            true,
		MonsterNoiseWeight:            0.1,
	}

	sim := simulation.NewSimulationManager(config, seed)

	ctx := context.Background()

	// Update Character Configs with weights if not set
	for i := range charCfgs {
		if charCfgs[i].UtilityWeights == nil {
			// Fallback weights if Joe or Paul didn't set them
			charCfgs[i].UtilityWeights = &core.UtilityWeights{
				ActionWeights: map[core.ActionType]float64{
					core.ATDamage: 1.0,
					core.ATHeal:   1.0,
				},
			}
			charCfgs[i].UtilityWeights.TargetFactorWeights.HighThreat = 0.5
			charCfgs[i].UtilityWeights.TargetFactorWeights.TargetPotency = 0.5
			charCfgs[i].UtilityWeights.TargetFactorWeights.TargetHitability = 0.5
			charCfgs[i].UtilityWeights.TargetFactorWeights.LowHP = 0.5
			charCfgs[i].UtilityWeights.TargetFactorWeights.Vengeance = 0.5
			charCfgs[i].UtilityWeights.TargetFactorWeights.ConcentrationBreak = 0.5
			charCfgs[i].UtilityWeights.TargetFactorWeights.ElitePriority = 0.5
			charCfgs[i].UtilityWeights.TargetFactorWeights.EmergencyHeal = 1.0
			charCfgs[i].UtilityWeights.ResourceExpenditureWeight = 0.5
		}
	}

	// SetupCombatantsFromAPI uses the simulation manager to create entities.
	// Since we want monsters to have weights, we need to ensure the initialization
	// path supports passing them.
	// Currently, SetupCombatantsFromAPI queries DB. For manual testing in main.go
	// with specific weights for monsters, we might need a small override or
	// rely on the default monster weights being applied in the engine if we implement that.

	// For now, let's assume the user wants to see the weighted AI in action with
	// the standard monster weights. I'll modify SetupCombatants to apply these defaults.

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
		nil,
		lc)

	// Since monsters are loaded from DB, we apply standard weights here
	for _, c := range sim.GetCombatEngine().Combatants {
		if c.Entity.IsMonster() {
			if m, ok := c.Entity.(*monster.Monster); ok {
				m.AI.Weights = standardMonsterWeights
			}
		}
	}

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

	// Export Timeline to JSON
	timeline := sim.GetTimeline()
	jsonData, err := json.MarshalIndent(timeline, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling timeline: %v\n", err)
		return
	}
	err = os.WriteFile("timeline_output.json", jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing timeline file: %v\n", err)
		return
	}
	fmt.Println("Timeline exported to timeline_output.json")
}
