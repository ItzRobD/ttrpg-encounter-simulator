package monster

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
)

func getMonsterBaseDataByName(ctx context.Context, name string) (MonsterBase, error) {
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
		    mhp.amount_to_add,
		    msp.strength,
		    msp.dexterity,
		    msp.constitution,
		    msp.intelligence,
		    msp.wisdom,
		    msp.charisma
		FROM monsters m
		JOIN monster_ability_score_block masb ON masb.monster_id = m.id
		JOIN monster_hit_point_formulas mhp ON mhp.monster_id = m.id
		JOIN monster_save_proficiencies msp ON msp.monster_id = m.id
		WHERE name ILIKE $1`

	row, err := database.QueryRow(ctx, query, name)
	if err != nil {
		return monsterResult, fmt.Errorf("failed to query monster base data by name: %w", err)
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
		&monsterResult.SaveProficiencies.Strength,
		&monsterResult.SaveProficiencies.Dexterity,
		&monsterResult.SaveProficiencies.Constitution,
		&monsterResult.SaveProficiencies.Intelligence,
		&monsterResult.SaveProficiencies.Wisdom,
		&monsterResult.SaveProficiencies.Charisma,
	)
	if err != nil {
		return monsterResult, fmt.Errorf("failed to scan monster base data by name: %w", err)
	}

	return monsterResult, nil
}

func getMonsterBaseDataByID(ctx context.Context, id int) (MonsterBase, error) {
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
		    mhp.amount_to_add,
		    msp.strength,
		    msp.dexterity,
		    msp.constitution,
		    msp.intelligence,
		    msp.wisdom,
		    msp.charisma
		FROM monsters m
		JOIN monster_ability_score_block masb ON masb.monster_id = m.id
		JOIN monster_hit_point_formulas mhp ON mhp.monster_id = m.id
		JOIN monster_save_proficiencies msp ON msp.monster_id = m.id
		WHERE id = $1`

	row, err := database.QueryRow(ctx, query, id)
	if err != nil {
		return monsterResult, fmt.Errorf("failed to query monster base data by id: %w", err)
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
		&monsterResult.SaveProficiencies.Strength,
		&monsterResult.SaveProficiencies.Dexterity,
		&monsterResult.SaveProficiencies.Constitution,
		&monsterResult.SaveProficiencies.Intelligence,
		&monsterResult.SaveProficiencies.Wisdom,
		&monsterResult.SaveProficiencies.Charisma,
	)
	if err != nil {
		return monsterResult, fmt.Errorf("failed to scan monster base data by id: %w", err)
	}

	return monsterResult, nil
}

func getMonsterDamageModifiersByID(ctx context.Context, id int) ([]MonsterDamageModifier, error) {
	var monsterDamageModifiers []MonsterDamageModifier
	query := `
		SELECT
			mdm.damage_type,
			mdm.modifier_type
		FROM monsters m
		JOIN monster_damage_modifiers mdm ON m.id = mdm.monster_id
		WHERE m.id = $1`

	rows, err := database.Query(ctx, query, id)
	if err != nil {
		return monsterDamageModifiers, fmt.Errorf("failed to query monster damage modifiers by id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var monsterDamageModifier MonsterDamageModifier
		err = rows.Scan(&monsterDamageModifier.DamageType, &monsterDamageModifier.ModifierType)
		if err != nil {
			return monsterDamageModifiers, fmt.Errorf("failed to scan monster damage modifiers by id: %w", err)
		}
		monsterDamageModifiers = append(monsterDamageModifiers, monsterDamageModifier)
	}

	if err := rows.Err(); err != nil {
		return monsterDamageModifiers, fmt.Errorf("failed to query monster damage modifiers by id: %w", err)
	}

	return monsterDamageModifiers, nil
}

func getMonsterResistBreakersByID(ctx context.Context, id int) ([]shared.DamageBreaker, error) {
	var monsterResistBreakers []shared.DamageBreaker
	query := `
		SELECT
		    mdm.damage_type,
			mrb.resist_breaker_type
		FROM monsters m
		JOIN monster_damage_modifiers mdm ON m.id = mdm.monster_id
		LEFT JOIN monster_damage_resist_breakers mdrb ON mdrb.modifier_id = mdm.modifier_id
		LEFT JOIN monster_resist_breakers mrb ON mrb.resist_breaker_id = mdrb.resist_breaker_id
		WHERE m.id = $1`

	rows, err := database.Query(ctx, query, id)
	if err != nil {
		return monsterResistBreakers, fmt.Errorf("failed to query monster resist breakers by id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var damageBreaker shared.DamageBreaker
		var damageBreakerName pgtype.Text
		err = rows.Scan(&damageBreaker.DamageType, &damageBreakerName)
		if err != nil {
			return monsterResistBreakers, fmt.Errorf("failed to scan monster resist breakers by id: %w", err)
		}
		if damageBreakerName.Valid {
			damageBreaker.Breaker = shared.WeaponBreakerType(damageBreakerName.String)
			monsterResistBreakers = append(monsterResistBreakers, damageBreaker)
		}
	}

	if err := rows.Err(); err != nil {
		return monsterResistBreakers, fmt.Errorf("failed to query monster resist breakers by id: %w", err)
	}

	return monsterResistBreakers, nil
}

func QueryMonsterData(ctx context.Context, params MonsterQueryParams) (Monster, error) {
	var monsterResult Monster
	var monsterBaseResult MonsterBase
	var err error

	if params.ID != 0 {
		monsterBaseResult, err = getMonsterBaseDataByID(ctx, params.ID)
	} else if params.Name != "" {
		monsterBaseResult, err = getMonsterBaseDataByName(ctx, params.Name)
	} else {
		err = fmt.Errorf("no name or id provided for monster data query")
		return monsterResult, err
	}

	if monsterBaseResult.ID != 0 {
		var monsterDamageModifiers []MonsterDamageModifier
		monsterDamageModifiers, err = getMonsterDamageModifiersByID(ctx, monsterBaseResult.ID)
		if err != nil {
			return monsterResult, err
		}
		monsterResult.DamageModifiers = monsterDamageModifiers

		var monsterResistBreakers []shared.DamageBreaker
		monsterResistBreakers, err = getMonsterResistBreakersByID(ctx, monsterBaseResult.ID)
		if err != nil {
			return monsterResult, err
		}
		monsterResult.ResistBreakers = monsterResistBreakers
	} else {
		err = fmt.Errorf("invalid monster id to query additional data")
		return monsterResult, err
	}

	monsterResult.MonsterBase = monsterBaseResult

	return monsterResult, err
}
