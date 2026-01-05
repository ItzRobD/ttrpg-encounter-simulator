package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"fmt"
	"math/rand/v2"
)

type CombatantSetupManager struct {
	ctx                    context.Context
	useHPAverageMonsters   bool
	useHPAverageCharacters bool
	rng                    *rand.Rand
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

func NewCombatantSetupManager(ctx context.Context, useHPAverageCharacters bool, useHPAverageMonsters bool, rng *rand.Rand) *CombatantSetupManager {
	return &CombatantSetupManager{
		ctx:                    ctx,
		useHPAverageCharacters: useHPAverageCharacters,
		useHPAverageMonsters:   useHPAverageMonsters,
		rng:                    rng,
	}
}

func (csm *CombatantSetupManager) SetupCombatants(characterConfigs []character.CharacterConfig, monsterIDs []int, monsterConfigs []monster.MonsterConfig) (*SetupResult, error) {
	result := &SetupResult{
		Combatants: make([]*core.Combatant, 0),
		Errors:     make([]SetupError, 0),
	}

	// Setup characters
	characters, charErrors := csm.createCharacters(characterConfigs)
	result.Combatants = append(result.Combatants, characters...)
	result.Errors = append(result.Errors, charErrors...)

	// Setup monsters from IDs
	monsters, monsterErrors := csm.createMonsters(monsterIDs)
	result.Combatants = append(result.Combatants, monsters...)
	result.Errors = append(result.Errors, monsterErrors...)

	// Setup monsters from configs
	customMonsters, customMonsterErrors := csm.createMonstersFromConfigs(monsterConfigs)
	result.Combatants = append(result.Combatants, customMonsters...)
	result.Errors = append(result.Errors, customMonsterErrors...)

	// If we have no valid combatants, return error
	if len(result.Combatants) == 0 {
		return result, fmt.Errorf("no valid combatants could be created")
	}

	return result, nil
}

func (csm *CombatantSetupManager) createMonstersFromConfigs(configs []monster.MonsterConfig) ([]*core.Combatant, []SetupError) {
	var combatants []*core.Combatant
	var errors []SetupError

	for _, config := range configs {
		if csm.useHPAverageMonsters {
			config.HPSetMethod = core.HPSetAverage
		} else {
			config.HPSetMethod = core.HPSetRoll
		}
		m, err := monster.NewMonsterWithRNG(csm.ctx, config, csm.rng)
		if err != nil {
			errors = append(errors, SetupError{
				Type:    "monster",
				ID:      config.Base.Name,
				Message: fmt.Sprintf("Failed to create monster %s: %v", config.Base.Name, err),
			})
			continue
		}
		combatants = append(combatants, core.NewCombatantWithInfo(m))
	}

	return combatants, errors
}

func (csm *CombatantSetupManager) createCharacters(configs []character.CharacterConfig) ([]*core.Combatant, []SetupError) {
	var combatants []*core.Combatant
	var errors []SetupError

	for _, config := range configs {
		char, err := character.NewCharacterWithRNG(csm.ctx, config, csm.rng)
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
				Message: fmt.Sprintf("Failed to query monster data for id %d: %v", id, err),
			})
		}
		return combatants, errors
	}

	// Track which monsters were successfully found
	foundIDs := make(map[int]bool)

	// Iterate over the original ids slice to ensure deterministic creation order
	for _, id := range ids {
		monsterConfig, exists := cfg[id]
		if !exists {
			continue
		}
		if csm.useHPAverageMonsters {
			monsterConfig.HPSetMethod = core.HPSetAverage
		} else {
			monsterConfig.HPSetMethod = core.HPSetRoll
		}
		m, err := monster.NewMonsterWithRNG(csm.ctx, monsterConfig, csm.rng)
		if err != nil {
			errors = append(errors, SetupError{
				Type:    "monster",
				ID:      fmt.Sprintf("%d", monsterConfig.Base.ID),
				Message: fmt.Sprintf("Failed to create monster: %v", err),
			})
			continue
		}

		combatants = append(combatants, core.NewCombatantWithInfo(m))

		foundIDs[m.ID] = true
	}

	// Add errors for any requested IDs that weren't found
	for _, requestedID := range ids {
		if !foundIDs[requestedID] {
			errors = append(errors, SetupError{
				Type:    "monster",
				ID:      fmt.Sprintf("%d", requestedID),
				Message: fmt.Sprintf("Monster with id %d not found", requestedID),
			})
		}
	}

	return combatants, errors
}
