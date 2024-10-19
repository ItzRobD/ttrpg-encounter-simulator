package spells

import (
	"context"
	"database/sql"
	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/enum"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"
)

func getSpellCastLevelsByID(ctx context.Context, id int) ([]int, error) {
	var levels []int
	stmt := SELECT(
		SpellFormulas.FormulaLevel,
	).FROM(
		SpellFormulas,
	).WHERE(SpellFormulas.SpellID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return levels, fmt.Errorf("failed to query spells cast levels by id: %d - %w", id, err)
	}
	defer rows.Close()
	for rows.Next() {
		var level int
		err = pgx.Row.Scan(rows,
			&level,
		)
		if err != nil {
			return levels, fmt.Errorf("failed to collect spells cast levels by id: %d - %w", id, err)
		}
		levels = append(levels, level)
	}
	return levels, nil
}

func getSpellTypeByID(ctx context.Context, id int) (string, error) {
	var levelType string
	stmt := SELECT(
		SpellFormulas.LevelType,
	).FROM(
		SpellFormulas,
	).WHERE(SpellFormulas.SpellID.EQ(Int(int64(id))))
	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return levelType, fmt.Errorf("failed to query spells level type by id: %d - %w", id, err)
	}
	err = pgx.Row.Scan(row,
		&levelType,
	)
	if err != nil {
		return levelType, fmt.Errorf("failed to collect spells level type by id: %d - %w", id, err)
	}
	return levelType, nil
}

func getMinimumSpellLevelByID(ctx context.Context, id int) (int, error) {
	var level int
	stmt := SELECT(
		MIN(Spells.Level),
	).
		FROM(
			Spells,
		).
		WHERE(
			Spells.ID.EQ(Int(int64(id))),
		)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return level, fmt.Errorf("failed to query minimum spells level by id: %d - %w", id, err)
	}
	err = pgx.Row.Scan(row,
		&level,
	)
	if err != nil {
		return level, fmt.Errorf("failed to collect minimum spells level by id: %d - %w", id, err)
	}

	return level, nil
}

func getSpellByID(ctx context.Context, id int, level int) (Spell, error) {
	var spell Spell
	stmt := SELECT(
		Spells.AllColumns,
		CASE().
			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
			THEN(SpellDc.Ability).
			ELSE(enum.Abilityscore.None),
		CASE().
			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
			THEN(SpellDc.OnSuccess).
			ELSE(enum.Dcsuccess.None),
		SpellFormulas.FormulaLevel,
		SpellFormulas.LevelType,
		CASE().
			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
			THEN(SpellDamage.NumberOfDice).
			ELSE(SpellHeal.NumberOfDice),
		CASE().
			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
			THEN(SpellDamage.Die).
			ELSE(SpellHeal.Die),
		CASE().
			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
			THEN(SpellDamage.AmountToAdd).
			ELSE(SpellHeal.AmountToAdd),
		CASE().
			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
			THEN(SpellDamage.DamageType).
			ELSE(enum.Dmg.None),
		CASE().
			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
			THEN(SpellDamage.UseSpellmod).
			ELSE(SpellHeal.UseSpellmod),
	).
		FROM(
			Spells.
				LEFT_JOIN(SpellDc, Spells.ID.EQ(SpellDc.SpellID)).
				LEFT_JOIN(SpellFormulas, Spells.ID.EQ(SpellFormulas.SpellID)).
				LEFT_JOIN(SpellDamage, SpellFormulas.FormulaID.EQ(SpellDamage.SpellFormulaID)).
				LEFT_JOIN(SpellHeal, SpellFormulas.FormulaID.EQ(SpellHeal.SpellFormulaID)),
		).
		WHERE(
			Spells.ID.EQ(Int(int64(id))).
				AND(SpellFormulas.FormulaLevel.EQ(Int(int64(level)))),
		)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return spell, fmt.Errorf("failed to query spells: %w", err)
	}
	var ability sql.NullString
	var onSuccess sql.NullString
	err = pgx.Row.Scan(row,
		&spell.ID,
		&spell.Name,
		&spell.Description,
		&spell.IsConcentration,
		&spell.CastingTime,
		&spell.IsRitual,
		&spell.Level,
		&spell.SpellType,
		&spell.IsAOE,
		&spell.HasDC,
		&spell.ApiURL,
		&ability,
		&onSuccess,
		&spell.CastLevel,
		&spell.LevelType,
		&spell.NumberOfDice,
		&spell.Die,
		&spell.AmountToAdd,
		&spell.DamageType,
		&spell.UseSpellmod,
	)
	if err != nil {
		return spell, fmt.Errorf("failed to collect spells: %w", err)
	}

	if ability.Valid {
		spell.Ability = ability.String
	} else {
		spell.Ability = ""
	}
	if onSuccess.Valid {
		spell.OnSuccess = onSuccess.String
	} else {
		spell.OnSuccess = ""
	}

	return spell, nil
}

func getSpellIDByName(ctx context.Context, name string) (int, error) {
	var id int
	stmt := SELECT(Spells.ID).
		FROM(Spells).
		WHERE(Spells.Name.EQ(String(name)))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return id, fmt.Errorf("failed to query spells id by name: %w", err)
	}
	err = pgx.Row.Scan(row, &id)
	if err != nil {
		return id, fmt.Errorf("failed to collect spells id by name: %w", err)
	}
	return id, nil
}

func getSpellQueryLevel(ctx context.Context, id int, paramLevel int) (int, error) {
	var queryLevel int
	var minLevel int
	var sType string
	var err error
	minLevel, err = getMinimumSpellLevelByID(ctx, id)
	if err != nil {
		return 0, err
	}
	sType, err = getSpellTypeByID(ctx, id)
	if err != nil {
		return 0, err
	}

	switch {
	case sType == "slot":
		switch {
		case paramLevel < minLevel:
			queryLevel = minLevel
		default:
			queryLevel = paramLevel
		}
	case sType == "character":
		castLevels := make([]int, 0, 9)
		castLevels, err = getSpellCastLevelsByID(ctx, paramLevel)
		if err != nil {
			return 0, err
		}
		queryLevel = 1
		for _, v := range castLevels {
			if v == paramLevel {
				queryLevel = paramLevel
				break
			}
		}
	}
	return queryLevel, nil
}

func QuerySpellData(ctx context.Context, params SpellQueryParams) (Spell, error) {
	var spell Spell
	var err error
	if params.Level < 0 || params.Level > 9 {
		return spell, fmt.Errorf("level must be within range 0-9")
	}

	var queryID int
	var queryLevel int

	if params.ID != 0 {
		queryID = params.ID
	} else {
		queryID, err = getSpellIDByName(ctx, params.Name)
		if err != nil {
			return spell, err
		}
	}
	queryLevel, err = getSpellQueryLevel(ctx, queryID, params.Level)
	if err != nil {
		return spell, err
	}
	spell, err = getSpellByID(ctx, queryID, queryLevel)
	if err != nil {
		return spell, err
	}

	return spell, nil
}

func GetUsableSpellIDsByClassID(ctx context.Context, classID int) ([]int, error) {
	var ids []int
	stmt := SELECT(
		Spells.ID,
	).FROM(
		Spells.
			INNER_JOIN(SpellUsers, Spells.ID.EQ(SpellUsers.SpellID)),
	).WHERE(
		SpellUsers.ClassID.EQ(Int(int64(classID))).
			AND(
				Spells.SpellType.EQ(enum.Stype.Damage).
					OR(Spells.SpellType.EQ(enum.Stype.Healing)),
			),
	)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query spells usable by class id: %d - %w", classID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to collect spells usable by class id: %d - %w", classID, err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func GetHealingAndDamageSpellsByClassID(ctx context.Context, classID int) ([]Spell, []Spell, error) {
	var healingSpells []Spell
	var damageSpells []Spell
	var err error
	var ids []int
	ids, err = GetUsableSpellIDsByClassID(ctx, classID)
	if err != nil {
		return healingSpells, damageSpells, err
	}
	for _, id := range ids {
		spell, err2 := QuerySpellData(ctx, SpellQueryParams{ID: id, Level: 0})
		if err2 != nil {
			return healingSpells, damageSpells, err2
		}
		if spell.SpellType == "healing" {
			healingSpells = append(healingSpells, spell)
		}
		if spell.SpellType == "damage" {
			damageSpells = append(damageSpells, spell)
		}
	}
	return healingSpells, damageSpells, nil
}
