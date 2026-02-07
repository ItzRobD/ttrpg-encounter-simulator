package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"testing"
)

func benchmarkSimulation(b *testing.B, numChars, numMonsters int) {
	options := &core.SimulationOptions{
		CanMonstersCrit:   true,
		CanCharactersCrit: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ed := NewEncounterDirector(core.Seed{Seed1: uint64(i), Seed2: uint64(i)}, options)

		// Add Characters
		for j := 0; j < numChars; j++ {
			pc := &actor.Actor{
				Name:      "Fighter",
				ActorType: core.ActorTypeCharacter,
				Side:      core.SideCharacters,
				AC:        16,
				Abilities: core.Abilities{AbilityScores: core.NewAbilityScores(16, 14, 14, 10, 12, 8)},
				Actions: []core.Action{
					{
						Name:        "Longsword",
						ActionType:  core.ATMelee,
						AttackBonus: 5,
						DiceBlock: []core.DiceBlock{
							{NumberOfDice: 1, Die: core.D8, DamageType: core.DamageSlashing, Modifier: 3},
						},
						Cost: core.ActionCost{ActivationType: core.ActAction, Value: 1},
					},
				},
				StateManager: state_manager.StateManager{
					CurrentHP: 20,
					MaxHP:     20,
				},
			}
			ed.AddActor(pc)
		}

		// Add Monsters
		for j := 0; j < numMonsters; j++ {
			monster := &actor.Actor{
				Name:      "Orc",
				ActorType: core.ActorTypeMonster,
				Side:      core.SideMonsters,
				AC:        13,
				Abilities: core.Abilities{AbilityScores: core.NewAbilityScores(16, 12, 16, 7, 11, 10)},
				Actions: []core.Action{
					{
						Name:        "Greataxe",
						ActionType:  core.ATMelee,
						AttackBonus: 5,
						DiceBlock: []core.DiceBlock{
							{NumberOfDice: 1, Die: core.D12, DamageType: core.DamageSlashing, Modifier: 3},
						},
						Cost: core.ActionCost{ActivationType: core.ActAction, Value: 1},
					},
				},
				StateManager: state_manager.StateManager{
					CurrentHP: 15,
					MaxHP:     15,
				},
			}
			ed.AddActor(monster)
		}

		ed.SetupEncounter()

		// Run for 10 rounds or until one side is dead
		for round := 0; round < 10; round++ {
			for _, id := range ed.TurnOrder {
				if ed.Actors[id].StateManager.CurrentHP > 0 {
					ed.ExecuteTurn(id)
				}
			}
			// Simple check for end of combat
			charsAlive := 0
			monstersAlive := 0
			for _, a := range ed.Actors {
				if a.StateManager.CurrentHP > 0 {
					if a.Side == core.SideCharacters {
						charsAlive++
					} else {
						monstersAlive++
					}
				}
			}
			if charsAlive == 0 || monstersAlive == 0 {
				break
			}
		}
	}
}

func BenchmarkSimulation_1v1(b *testing.B) { benchmarkSimulation(b, 1, 1) }
func BenchmarkSimulation_4v4(b *testing.B) { benchmarkSimulation(b, 4, 4) }
func BenchmarkSimulation_8v8(b *testing.B) { benchmarkSimulation(b, 8, 8) }
