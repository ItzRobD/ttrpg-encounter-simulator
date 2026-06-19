package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/simulation/intermission_manager"
	"testing"
)

func TestRunAdventuringDay_Basic(t *testing.T) {
	ctx := context.Background()

	charCfg := actor.ActorConfig{
		ID:        "char-1",
		Name:      "Hero",
		Side:      core.SideCharacters,
		ActorType: core.ActorTypeCharacter,
		IsCustom:  true,
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{Strength: 16, Dexterity: 14, Constitution: 14, Intelligence: 10, Wisdom: 12, Charisma: 10},
		},
		HPConfig: core.HPConfig{
			HPMethod:  core.HPSetValue,
			Value:     100,
			HPAverage: 100,
		},
		Metadata: actor.Metadata{
			Level: 5,
		},
		Actions: []core.Action{
			{
				ID:          core.MakeID("hero-attack"),
				Name:        "Hero Attack",
				Cost:        core.ActionCost{ActivationType: core.ActAction, Value: 1},
				ActionType:  core.ATMelee,
				AttackBonus: 10,
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 1, Die: core.D8, Modifier: 5, DamageType: core.DamageSlashing},
				},
				AverageDamage: 9,
			},
		},
	}

	monsterCfg := actor.ActorConfig{
		ID:        "monster-1",
		Name:      "Goblin",
		Side:      core.SideMonsters,
		ActorType: core.ActorTypeMonster,
		IsCustom:  true,
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{Strength: 8, Dexterity: 14, Constitution: 10, Intelligence: 10, Wisdom: 8, Charisma: 8},
		},
		HPConfig: core.HPConfig{
			HPMethod:  core.HPSetValue,
			Value:     1, // Very weak so hero wins easily
			HPAverage: 1,
		},
		Metadata: actor.Metadata{
			CR: 0.25,
		},
		Actions: []core.Action{
			{
				ID:          core.MakeID("goblin-attack"),
				Name:        "Goblin Attack",
				Cost:        core.ActionCost{ActivationType: core.ActAction, Value: 1},
				ActionType:  core.ATMelee,
				AttackBonus: 0,
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 1, Die: core.D4, Modifier: 0, DamageType: core.DamageSlashing},
				},
				AverageDamage: 2,
			},
		},
	}

	req := AdventuringDayRequest{
		BaseOptions: core.SimulationOptions{
			Seed: core.Seed{Seed1: 1, Seed2: 2},
		},
		CharacterConfigs: []actor.ActorConfig{charCfg},
		Encounters: []EncounterConfig{
			{
				Name:           "Encounter 1",
				MonsterConfigs: []actor.ActorConfig{monsterCfg},
			},
			{
				Name:           "Encounter 2",
				MonsterConfigs: []actor.ActorConfig{monsterCfg},
			},
		},
		Intermission: intermission_manager.IntermissionOptions{
			MaxShortRests:          2,
			ShortRestHealThreshold: 0.7,
		},
		MaxRounds:   10,
		IncludeLogs: true,
	}

	res, err := RunAdventuringDay(ctx, req)
	if err != nil {
		t.Fatalf("RunAdventuringDay failed: %v", err)
	}

	for _, encRes := range res.EncounterResults {
		for _, event := range encRes.Logs {
			t.Logf("Encounter: %s - Event: %s - Actor: %+v - Data: %+v", encRes.EncounterName, event.Type, event.Actor, event.Data)
		}
	}

	if res.TotalEncounters != 2 {
		t.Errorf("Expected 2 encounters, got %d", res.TotalEncounters)
	}

	if res.SuccessRate != 100.0 {
		t.Errorf("Expected 100%% success rate, got %f", res.SuccessRate)
	}

	if len(res.EncounterResults) != 2 {
		t.Errorf("Expected 2 encounter results, got %d", len(res.EncounterResults))
	}

	if res.EncountersWon != 2 {
		t.Errorf("Expected 2 encounters won, got %d", res.EncountersWon)
	}
}
