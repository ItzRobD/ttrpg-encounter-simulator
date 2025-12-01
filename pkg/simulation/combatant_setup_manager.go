package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"fmt"
)

type CombatantSetupManager struct {
	ctx                    context.Context
	useHPAverageMonsters   bool
	useHPAverageCharacters bool
}

type SetupResult struct {
	Combatants []*core.Combatant
	Errors     []SetupError
}

type SetupError struct {
	Type    string `json:"type"` // "character" or "monster"
	ID      string `json:"id"`   // character name or monster ID
	Message string `json:"message"`
}

func NewCombatantSetupManager(ctx context.Context, useHPAverageCharacters bool, useHPAverageMonsters bool) *CombatantSetupManager {
	return &CombatantSetupManager{
		ctx:                    ctx,
		useHPAverageCharacters: useHPAverageCharacters,
		useHPAverageMonsters:   useHPAverageMonsters,
	}
}

func (csm *CombatantSetupManager) SetupCombatants(characterConfigs []character.CharacterConfig, monsterIDs []int) (*SetupResult, error) {
	result := &SetupResult{
		Combatants: make([]*core.Combatant, 0),
		Errors:     make([]SetupError, 0),
	}

	// Setup characters
	characters, charErrors := csm.createCharacters(characterConfigs)
	result.Combatants = append(result.Combatants, characters...)
	result.Errors = append(result.Errors, charErrors...)

	// Setup monsters
	monsters, monsterErrors := csm.createMonsters(monsterIDs)
	result.Combatants = append(result.Combatants, monsters...)
	result.Errors = append(result.Errors, monsterErrors...)

	// If we have no valid combatants, return error
	if len(result.Combatants) == 0 {
		return result, fmt.Errorf("no valid combatants could be created")
	}

	return result, nil
}

func (csm *CombatantSetupManager) createCharacters(configs []character.CharacterConfig) ([]*core.Combatant, []SetupError) {
	var combatants []*core.Combatant
	var errors []SetupError

	for _, config := range configs {
		char, err := character.NewCharacter(csm.ctx, config)
		if err != nil {
			errors = append(errors, SetupError{
				Type:    "character",
				ID:      config.Name,
				Message: fmt.Sprintf("Failed to create character %s: %v", config.Name, err),
			})
			continue
		}
		combatants = append(combatants, core.NewCombatantWithInfo(char))
	}

	return combatants, errors
}

func (csm *CombatantSetupManager) createMonsters(ids []int) ([]*core.Combatant, []SetupError) {
	var combatants []*core.Combatant
	var errors []SetupError

	if len(ids) == 0 {
		return combatants, errors
	}

	cfg, err := monster.QueryMonsterConfigData(csm.ctx, monster.MonsterQueryParams{ID: ids})
	if err != nil {
		// If we can't query any monsters, add errors for all requested IDs
		for _, id := range ids {
			errors = append(errors, SetupError{
				Type:    "monster",
				ID:      fmt.Sprintf("%d", id),
				Message: fmt.Sprintf("Failed to query monster data for ID %d: %v", id, err),
			})
		}
		return combatants, errors
	}

	// Track which monsters were successfully found
	foundIDs := make(map[int]bool)

	for _, monsterConfig := range cfg {
		if csm.useHPAverageMonsters {
			monsterConfig.HPSetMethod = core.HPSetAverage
		} else {
			monsterConfig.HPSetMethod = core.HPSetRoll
		}
		monster, err := monster.NewMonster(csm.ctx, monsterConfig)
		if err != nil {
			errors = append(errors, SetupError{
				Type:    "monster",
				ID:      fmt.Sprintf("%d", monster.ID),
				Message: fmt.Sprintf("Failed to create monster: %v", err),
			})
			continue
		}

		combatants = append(combatants, core.NewCombatantWithInfo(monster))

		foundIDs[monster.ID] = true
	}

	// Add errors for any requested IDs that weren't found
	for _, requestedID := range ids {
		if !foundIDs[requestedID] {
			errors = append(errors, SetupError{
				Type:    "monster",
				ID:      fmt.Sprintf("%d", requestedID),
				Message: fmt.Sprintf("Monster with ID %d not found", requestedID),
			})
		}
	}

	return combatants, errors
}
