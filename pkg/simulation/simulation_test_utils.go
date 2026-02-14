package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
)

// CreateMockFighter creates a standard level 1 Fighter-like actor.
func CreateMockFighter(id int, name string, side core.Side) *actor.Actor {
	return &actor.Actor{
		InstanceID: id,
		Name:       name,
		Side:       side,
		ActorType:  core.ActorTypeCharacter,
		AC:         16, // Chain Mail
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{
				Strength:     16,
				Dexterity:    12,
				Constitution: 14,
				Intelligence: 8,
				Wisdom:       10,
				Charisma:     10,
			},
		},
		HPConfig: core.HPConfig{HPAverage: 12},
		StateManager: state_manager.StateManager{
			CurrentHP:   12,
			MaxHP:       12,
			AttackCount: 1,
		},
		Actions: []core.Action{
			{
				ID:            core.MakeID(1),
				Name:          "Greatsword",
				ActionType:    core.ATMelee,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AttackBonus:   5,  // Str + Prof
				AverageDamage: 10, // 2d6 + 3
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 2, Die: core.D6, Modifier: 3, DamageType: core.DamageSlashing},
				},
			},
		},
		Behavior: actor.BehaviorConfig{
			ActionPreference: core.APAttack,
		},
	}
}

// CreateMockCleric creates a standard level 1 Cleric-like actor with healing.
func CreateMockCleric(id int, name string, side core.Side) *actor.Actor {
	return &actor.Actor{
		InstanceID: id,
		Name:       name,
		Side:       side,
		ActorType:  core.ActorTypeCharacter,
		AC:         18, // Chain Mail + Shield
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{
				Strength:     14,
				Dexterity:    10,
				Constitution: 14,
				Intelligence: 10,
				Wisdom:       16,
				Charisma:     12,
			},
		},
		HPConfig: core.HPConfig{HPAverage: 10},
		StateManager: state_manager.StateManager{
			CurrentHP:   10,
			MaxHP:       10,
			AttackCount: 1,
		},
		Actions: []core.Action{
			{
				ID:            core.MakeID(1),
				Name:          "Mace",
				ActionType:    core.ATMelee,
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AttackBonus:   4, // Str + Prof
				AverageDamage: 5, // 1d6 + 2
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 1, Die: core.D6, Modifier: 2, DamageType: core.DamageBludgeoning},
				},
			},
			{
				ID:          core.MakeID(2),
				Name:        "Cure Wounds",
				ActionType:  core.ATAction,
				Cost:        core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AverageHeal: 7, // 1d8 + 3
			},
		},
		Behavior: actor.BehaviorConfig{
			ActionPreference: core.APHeal,
		},
	}
}

// CreateMockGoblin creates a standard CR 1/4 Goblin.
func CreateMockGoblin(id int, name string, side core.Side) *actor.Actor {
	return &actor.Actor{
		InstanceID: id,
		Name:       name,
		Side:       side,
		ActorType:  core.ActorTypeMonster,
		AC:         15,
		Abilities: core.Abilities{
			AbilityScores: core.AbilityScores{
				Strength:     8,
				Dexterity:    14,
				Constitution: 10,
				Intelligence: 10,
				Wisdom:       8,
				Charisma:     8,
			},
		},
		HPConfig: core.HPConfig{HPAverage: 7},
		StateManager: state_manager.StateManager{
			CurrentHP: 7,
			MaxHP:     7,
		},
		Actions: []core.Action{
			{
				ID:            core.MakeID(1),
				Name:          "Scimitar",
				ActionType:    core.ATRanged, // Goblins often use ranged too, but let's stick to scimitar
				Cost:          core.ActionCost{ActivationType: core.ActAction, Value: 1},
				AttackBonus:   4,
				AverageDamage: 5,
				DiceBlock: []core.DiceBlock{
					{NumberOfDice: 1, Die: core.D6, Modifier: 2, DamageType: core.DamageSlashing},
				},
			},
		},
		Behavior: actor.BehaviorConfig{
			ActionPreference: core.APAttack,
		},
	}
}

// SetupStandardEncounter initializes a director with 2 players (Fighter, Cleric) and 3 monsters (Goblins).
// It can be used as a baseline for integration tests.
func SetupStandardEncounter() *EncounterDirector {
	return SetupStandardEncounterWithSeed(core.Seed{Seed1: 12345, Seed2: 67890})
}

func SetupStandardEncounterWithSeed(seed core.Seed) *EncounterDirector {
	simOptions := &core.SimulationOptions{
		CharacterHealThresholdPct:      50,
		CharacterEmergencyThresholdPct: 25,
		MonsterHealThresholdPct:        30,
		UseWeightedAI:                  true,
	}
	ed := NewEncounterDirector(seed, simOptions)

	fighter := CreateMockFighter(1, "Valeros", core.SideCharacters)
	cleric := CreateMockCleric(2, "Kyra", core.SideCharacters)
	// Initialize slots for Cleric
	cleric.StateManager.MaxSlots = spells.SpellSlots{1: 2}
	cleric.StateManager.CurrentSlots = spells.SpellSlots{1: 2}

	ed.AddActor(fighter)
	ed.AddActor(cleric)
	ed.AddActor(CreateMockGoblin(3, "Goblin 1", core.SideMonsters))
	ed.AddActor(CreateMockGoblin(4, "Goblin 2", core.SideMonsters))
	ed.AddActor(CreateMockGoblin(5, "Goblin 3", core.SideMonsters))

	ed.SetupEncounter()
	return ed
}
