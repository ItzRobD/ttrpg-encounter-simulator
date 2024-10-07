package monster

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"
)

func getMonsterDataByName(ctx context.Context, name string) (MonsterBase, error) {
	var monsterResult MonsterBase
	query := `
		SELECT
		    m.id,
			m.name,
			m.size,
			m.type,
			m.armor_class,
			m.proficiency_bonus,
			m.cr,
			m.api_url,
			m.is_legendary,
			m.is_spellcaster,
			m.is_innate_caster,
			masb.strength,
			masb.dexterity,
			masb.constitution,
			masb.intelligence,
			masb.wisdom,
			masb.charisma,
			mhp.hp_average,
			mhp.number_of_dice,
		    mhp.die,
		    mhp.amount_to_add
		FROM monsters m
		JOIN monster_ability_score_block masb ON masb.monster_id = m.id
		JOIN monster_hit_point_formulas mhp ON mhp.monster_id = m.id
		WHERE name ILIKE $1`

	row, err := database.QueryRow(ctx, query, name)
	if err != nil {
		return monsterResult, fmt.Errorf("failed to query monster data by name: %w", err)
	}
	err = row.Scan(
		&monsterResult.ID,
		&monsterResult.Name,
		&monsterResult.Size,
		&monsterResult.Type,
		&monsterResult.AC,
		&monsterResult.ProficiencyBonus,
		&monsterResult.CR,
		&monsterResult.ApiURL,
		&monsterResult.IsLegendary,
		&monsterResult.IsSpellcaster,
		&monsterResult.IsInnateSpellcaster,
		&monsterResult.AbilityScores.Strength,
		&monsterResult.AbilityScores.Dexterity,
		&monsterResult.AbilityScores.Constitution,
		&monsterResult.AbilityScores.Intelligence,
		&monsterResult.AbilityScores.Wisdom,
		&monsterResult.AbilityScores.Charisma,
		&monsterResult.HP.HPAverage,
		&monsterResult.HP.NumberOfDice,
		&monsterResult.HP.Die,
		&monsterResult.HP.AmountToAdd,
	)
	if err != nil {
		return monsterResult, fmt.Errorf("failed to scan monster data by name: %w", err)
	}

	return monsterResult, nil
}

func getMonsterDataByID(ctx context.Context, id int) (MonsterBase, error) {
	var monsterResult MonsterBase
	query := `
		SELECT
		    m.id,
			m.name,
			m.size,
			m.type,
			m.armor_class,
			m.proficiency_bonus,
			m.cr,
			m.api_url,
			m.is_legendary,
			m.is_spellcaster,
			m.is_innate_caster,
			masb.strength,
			masb.dexterity,
			masb.constitution,
			masb.intelligence,
			masb.wisdom,
			masb.charisma,
			mhp.hp_average,
			mhp.number_of_dice,
		    mhp.die,
		    mhp.amount_to_add
		FROM monsters m
		JOIN monster_ability_score_block masb ON masb.monster_id = m.id
		JOIN monster_hit_point_formulas mhp ON mhp.monster_id = m.id
		WHERE id = $1`

	row, err := database.QueryRow(ctx, query, id)
	if err != nil {
		return monsterResult, fmt.Errorf("failed to query monster data by id: %w", err)
	}
	err = row.Scan(
		&monsterResult.ID,
		&monsterResult.Name,
		&monsterResult.Size,
		&monsterResult.Type,
		&monsterResult.AC,
		&monsterResult.ProficiencyBonus,
		&monsterResult.CR,
		&monsterResult.ApiURL,
		&monsterResult.IsLegendary,
		&monsterResult.IsSpellcaster,
		&monsterResult.IsInnateSpellcaster,
		&monsterResult.AbilityScores.Strength,
		&monsterResult.AbilityScores.Dexterity,
		&monsterResult.AbilityScores.Constitution,
		&monsterResult.AbilityScores.Intelligence,
		&monsterResult.AbilityScores.Wisdom,
		&monsterResult.AbilityScores.Charisma,
		&monsterResult.HP.HPAverage,
		&monsterResult.HP.NumberOfDice,
		&monsterResult.HP.Die,
		&monsterResult.HP.AmountToAdd,
	)
	if err != nil {
		return monsterResult, fmt.Errorf("failed to scan monster data by id: %w", err)
	}

	return monsterResult, nil
}

func QueryMonsterData(ctx context.Context, params MonsterQueryParams) (MonsterBase, error) {
	var monsterResult MonsterBase
	var err error

	if params.ID != 0 {
		monsterResult, err = getMonsterDataByID(ctx, params.ID)
	} else if params.Name != "" {
		monsterResult, err = getMonsterDataByName(ctx, params.Name)
	} else {
		err = fmt.Errorf("no name or id provided for monster data query")
	}

	return monsterResult, err
}
