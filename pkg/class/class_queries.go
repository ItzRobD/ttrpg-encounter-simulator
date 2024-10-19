package class

import (
	"context"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5/pgtype"
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

func getClassSpellSlotsByID(ctx context.Context, id int) (map[int]map[int]int, error) {
	var spellSlots map[int]map[int]int
	stmt := SELECT(
		ClassesSpellSlots.Level,
		ClassesSpellSlots.Level1,
		ClassesSpellSlots.Level2,
		ClassesSpellSlots.Level3,
		ClassesSpellSlots.Level4,
		ClassesSpellSlots.Level5,
		ClassesSpellSlots.Level6,
		ClassesSpellSlots.Level7,
		ClassesSpellSlots.Level8,
		ClassesSpellSlots.Level9,
	).FROM(
		ClassesSpellSlots,
	).WHERE(
		ClassesSpellSlots.ClassID.EQ(Int(int64(id))),
	)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return spellSlots, fmt.Errorf("failed to query spell slots by class id and level: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var level int
		var level1, level2, level3, level4, level5, level6, level7, level8, level9 int
		err = rows.Scan(&level, &level1, &level2, &level3, &level4, &level5, &level6, &level7, &level8, &level9)
		if err != nil {
			return spellSlots, fmt.Errorf("failed to scan spell slots by class id and level: %w", err)
		}
		if spellSlots == nil {
			spellSlots = make(map[int]map[int]int)
		}
		if spellSlots[level] == nil {
			spellSlots[level] = make(map[int]int)
			spellSlots[level][1] = level1
			spellSlots[level][2] = level2
			spellSlots[level][3] = level3
			spellSlots[level][4] = level4
			spellSlots[level][5] = level5
			spellSlots[level][6] = level6
			spellSlots[level][7] = level7
			spellSlots[level][8] = level8
			spellSlots[level][9] = level9
		}
	}
	return spellSlots, nil
}

func QueryClassData(ctx context.Context, params ClassQueryParams) (Class, error) {
	var classResult Class
	var err error

	if params.ID != 0 {
		classResult, err = getClassByID(ctx, params.ID)
		if err != nil {
			return classResult, err
		}
	} else if params.Name != "" {
		var id int
		id, err = getClassIDByName(ctx, params.Name)
		if err != nil {
			return classResult, err
		}
		classResult, err = getClassByID(ctx, id)
		if err != nil {
			return classResult, err
		}
	} else {
		return classResult, fmt.Errorf("no class id or name provided")
	}
	const barbarianClassID = 2
	if classResult.ID != barbarianClassID {
		classResult.Spellcasting.ClassHealingSpells, classResult.Spellcasting.ClassDamageSpells, err = spells.GetHealingAndDamageSpellsByClassID(ctx, classResult.ID)
		if err != nil {
			return classResult, err
		}

		classResult.Spellcasting.MaxSpellSlots, err = getClassSpellSlotsByID(ctx, classResult.ID)
	} else {
		classResult.SpellcastingMod = "none"
	}

	return classResult, err
}
