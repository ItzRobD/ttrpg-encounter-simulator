package class

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"

	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	. "github.com/go-jet/jet/v2/postgres"
)

func getClassIDByName(ctx context.Context, name string) (int, error) {
	var id int
	stmt := SELECT(
		Classes.ID,
	).FROM(
		Classes,
	).WHERE(
		Classes.ClassName.EQ(String(name)))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return id, fmt.Errorf("failed to query class id by name: %w", err)
	}
	err = row.Scan(&id)
	if err != nil {
		return id, fmt.Errorf("failed to scan class id by name: %w", err)
	}

	return id, nil
}

func getClassByID(ctx context.Context, id int) (Class, error) {
	var classResult Class
	query := `
		SELECT
		    id,
		    class_name,
		    hit_die,
		    spellmod		    
		FROM classes WHERE id = $1`

	stmt := SELECT(
		Classes.ID,
		Classes.ClassName,
		Classes.HitDie,
		Classes.Spellmod,
	).FROM(
		Classes,
	).WHERE(Classes.ID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return classResult, fmt.Errorf("failed to query class by id: %w", err)
	}
	var spellmod pgtype.Text
	err = row.Scan(&classResult.ID, &classResult.Name, &classResult.HitDie, &spellmod)
	if err != nil {
		return classResult, fmt.Errorf("failed to scan class by id: %w", err)
	}
	if spellmod.Valid {
		classResult.SpellcastingMod = spellmod.String
	}

	return classResult, nil
}

func QueryClassData(ctx context.Context, params ClassQueryParams) (Class, error) {
	var classResult Class
	var err error

	if params.ID != 0 {
		classResult, err = getClassByID(ctx, params.ID)
	} else if params.Name != "" {
		var id int
		id, err = getClassIDByName(ctx, params.Name)
		if err != nil {
			return classResult, err
		}
		classResult, err = getClassByID(ctx, id)
	} else {
		return classResult, fmt.Errorf("no name or id provided for class query")
	}

	return classResult, err
}
