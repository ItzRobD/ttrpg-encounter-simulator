package class

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
)

func getClassByName(ctx context.Context, name string) (Class, error) {
	var classResult Class
	query := `
		SELECT
		    id,
		    class_name,
		    hit_die,
		    spellmod		    
		FROM classes WHERE class_name ILIKE $1`

	row, err := database.QueryRow(ctx, query, name)
	if err != nil {
		return classResult, fmt.Errorf("failed to query class by name: %w", err)
	}
	var spellmod pgtype.Text
	err = row.Scan(&classResult.ID, &classResult.Name, &classResult.HitDie, &spellmod)
	if err != nil {
		return classResult, fmt.Errorf("failed to scan class by name: %w", err)
	}
	if spellmod.Valid {
		classResult.SpellcastingMod = spellmod.String
	}

	return classResult, nil
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

	row, err := database.QueryRow(ctx, query, id)
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
		classResult, err = getClassByName(ctx, params.Name)
	} else {
		return classResult, fmt.Errorf("no name or id provided for class query")
	}

	return classResult, err
}
