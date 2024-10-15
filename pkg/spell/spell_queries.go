package spell

import (
	"context"
	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/enum"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"
)

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
		return level, fmt.Errorf("failed to query minimum spell level by id: %d - %w", id, err)
	}
	err = pgx.Row.Scan(row,
		&level,
	)

	return level, nil
}

func getMinimumSpellLevelByName(ctx context.Context, name string) (int, error) {
	var level int
	stmt := SELECT(
		MIN(Spells.Level),
	).
		FROM(
			Spells,
		).
		WHERE(
			Spells.Name.EQ(String(name)),
		)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return level, fmt.Errorf("failed to query minimum spell level by name: %w - %w", name, err)
	}
	err = pgx.Row.Scan(row,
		&level,
	)

	return level, nil
}

func getSpellByName(ctx context.Context, name string, level int) (Spell, error) {
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
			THEN(SpellDamage.UseSpellmod).
			ELSE(SpellHeal.UseSpellmod),
		CASE().
			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
			THEN(SpellDamage.DamageType).
			ELSE(enum.Dmg.None),
	).
		FROM(
			Spells.
				LEFT_JOIN(SpellDc, Spells.ID.EQ(SpellDc.SpellID)).
				LEFT_JOIN(SpellFormulas, Spells.ID.EQ(SpellFormulas.SpellID)).
				LEFT_JOIN(SpellDamage, SpellFormulas.FormulaID.EQ(SpellDamage.SpellFormulaID)).
				LEFT_JOIN(SpellHeal, SpellFormulas.FormulaID.EQ(SpellHeal.SpellFormulaID)),
		).
		WHERE(
			Spells.Name.LIKE(String(name)).
				AND(SpellFormulas.FormulaLevel.EQ(Int(int64(level)))),
		)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return spell, fmt.Errorf("failed to query spell: %w", err)
	}
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
		&spell.Ability,
		&spell.OnSuccess,
		&spell.CastLevel,
		&spell.LevelType,
		&spell.NumberOfDice,
		&spell.Die,
		&spell.AmountToAdd,
		&spell.UseSpellmod,
		&spell.DamageType,
	)
	if err != nil {
		return spell, fmt.Errorf("failed to collect spell: %w", err)
	}

	return spell, nil
}

func getSpellByID(ctx context.Context, id int, level int) (Spell, error) {
	var spell Spell
	stmt := SELECT(
		Spells.AllColumns,
		SpellDc.Ability,
		SpellDc.OnSuccess,
		SpellFormulas.FormulaLevel,
		SpellFormulas.LevelType,
		SpellDamage.NumberOfDice,
		SpellDamage.Die,
		SpellDamage.AmountToAdd,
		SpellDamage.UseSpellmod,
		SpellDamage.DamageType,
	).
		FROM(
			Spells.
				LEFT_JOIN(SpellDc, Spells.ID.EQ(SpellDc.SpellID)).
				LEFT_JOIN(SpellFormulas, Spells.ID.EQ(SpellFormulas.SpellID)).
				LEFT_JOIN(SpellDamage, SpellFormulas.FormulaID.EQ(SpellDamage.SpellFormulaID)),
		).
		WHERE(
			Spells.ID.EQ(Int(int64(id))).
				AND(SpellFormulas.FormulaLevel.EQ(Int(int64(level)))),
		)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return spell, fmt.Errorf("failed to query spell: %w", err)
	}
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
		&spell.Ability,
		&spell.OnSuccess,
		&spell.CastLevel,
		&spell.LevelType,
		&spell.NumberOfDice,
		&spell.Die,
		&spell.AmountToAdd,
		&spell.DamageType,
		&spell.UseSpellmod,
	)
	if err != nil {
		return spell, fmt.Errorf("failed to collect spell: %w", err)
	}

	return spell, nil
}

func QuerySpellData(ctx context.Context, params SpellQueryParams) (Spell, error) {
	var spell Spell
	if params.Level < 0 || params.Level > 9 {
		return spell, fmt.Errorf("level must be within range 0-9")
	}

	if params.ID != 0 {
		minLevel, err := getMinimumSpellLevelByID(ctx, params.ID)
		if err != nil {
			return spell, fmt.Errorf("failed to query minimum spell level: %w", err)
		}
		if params.Level < minLevel {
			spell, err = getSpellByID(ctx, params.ID, minLevel)
		} else {
			spell, err = getSpellByID(ctx, params.ID, params.Level)
		}
	} else {
		minLevel, err := getMinimumSpellLevelByName(ctx, params.Name)
		fmt.Println(minLevel, err)
		if err != nil {
			return spell, fmt.Errorf("failed to query minimum spell level: %w", err)
		}
		if params.Level < minLevel {
			spell, err = getSpellByName(ctx, params.Name, minLevel)
		} else {
			spell, err = getSpellByName(ctx, params.Name, params.Level)
		}
	}

	return spell, nil
}

func GetSpellsUsableByClassID(ctx context.Context, classID int) ([]Spell, error) {
	var csIDs []int
	stmt := SELECT(
		Spells.ID,
	).
		FROM(
			Spells.INNER_JOIN(SpellUsers, Spells.ID.EQ(SpellUsers.SpellID)),
		).
		WHERE(
			SpellUsers.ClassID.EQ(Int(int64(classID))))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query spells usable by class id (%d): %w", classID, err)
	}
	defer rows.Close()
	for rows.Next() {
		var spellID int
		err = rows.Scan(&spellID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan spell id: %w", err)
		}
		csIDs = append(csIDs, spellID)
	}

	var spells []Spell
	for _, spellID := range csIDs {
		params := SpellQueryParams{ID: spellID, Level: 0}
		spell, err := QuerySpellData(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("failed to get spell by id (%d): %w", spellID, err)
		}
		spells = append(spells, spell)
	}

	return spells, nil
}
