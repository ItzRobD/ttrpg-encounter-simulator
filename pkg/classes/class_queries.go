package classes

import (
	"context"
	"database/sql"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"errors"
	"fmt"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5/pgtype"
)

// getClassIDByName queries the database to find the class ID corresponding to the given class name.
func getClassIDByName(ctx context.Context, name string) (uint8, error) {
	var id uint8
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

// getClassByID retrieves a Class by its ID from the database using a context and an integer ID as inputs.
// It returns the Class struct and an error if the query or scan operation fails.
func getClassByID(ctx context.Context, id uint8) (Class, error) {
	var classResult Class
	stmt := SELECT(
		Classes.ID,
		Classes.ClassName,
		Classes.HitDie,
	).FROM(
		Classes,
	).WHERE(Classes.ID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return classResult, fmt.Errorf("failed to query class by id: %w", err)
	}
	err = row.Scan(&classResult.ID, &classResult.Name, &classResult.HitDie)
	if err != nil {
		return classResult, fmt.Errorf("failed to scan class by id: %w", err)
	}

	return classResult, nil
}

// getClassSpellSlotsByID retrieves spell slots of a class by its ID, returning a nested map of level and slots per level.
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

// QueryClassData retrieves a Class based on the provided query parameters (ID or Name) and fetches its spellcasting mod.
func QueryClassData(ctx context.Context, params ClassQueryParams) (Class, error) {
	var classResult Class
	var err error

	if params.ID != 0 {
		classResult, err = getClassByID(ctx, params.ID.Int())
		if err != nil {
			return classResult, err
		}
	} else if params.Name != "" {
		var id uint8
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

	classResult.SpellcastingMod, err = getSpellModByClassID(ctx, classResult.ID.Int())
	if err != nil {
		return Class{}, err
	}

	if classResult.ID == Rogue {
		classResult.SneakAttackDiceCount, err = GetNumberOfSneakAttackDiceFromLevel(ctx, params.Level)
		if err != nil {
			return Class{}, err
		}
	}

	classResult.AttackCount, err = GetNumberOfAttacksFromLevelAndClass(ctx, params.Level, classResult.ID.Int())
	if err != nil {
		return Class{}, err
	}

	return classResult, err
}

// GetNumberOfAttacksFromLevelAndClass determines the number of attacks a character has based on class and level inputs.
// ctx is the context for database queries.
// level is the character's level between 1 and 20.
// classID is the class identifier between 1 and 13.
// Returns the number of attacks or an error if inputs are invalid or the query fails.
func GetNumberOfAttacksFromLevelAndClass(ctx context.Context, level uint8, classID uint8) (int, error) {
	if level <= 0 || level > 20 {
		return 0, fmt.Errorf("invalid level provided: %d", level)
	}
	if classID <= 0 || classID > 13 {
		return 0, fmt.Errorf("invalid class id provided: %d", classID)
	}

	numberOfAttacks := 1

	stmt := SELECT(ClassesExtraAttack.NumberOfAttacks).
		FROM(ClassesExtraAttack).
		WHERE(ClassesExtraAttack.ClassID.EQ(Int(int64(classID))).AND(ClassesExtraAttack.ClassLevel.LT_EQ(Int(int64(level))))).
		ORDER_BY(ClassesExtraAttack.ClassLevel.DESC()).
		LIMIT(1)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return numberOfAttacks, fmt.Errorf("failed to query extra attacks by level and class: %w", err)
	}

	var numberOfAttacksNullable sql.NullInt64
	err = row.Scan(&numberOfAttacksNullable)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return numberOfAttacks, nil
		}
		return numberOfAttacks, fmt.Errorf("failed to scan extra attacks by level and class: %w", err)
	}

	if numberOfAttacksNullable.Valid {
		numberOfAttacks = int(numberOfAttacksNullable.Int64)
	} else {
		numberOfAttacks = 1
	}

	return numberOfAttacks, nil
}

// GetNumberOfSneakAttackDiceFromLevel retrieves the number of sneak attack dice for a given character level (1-20).
// ctx is the context for managing database queries.
// level specifies the character's level and must be between 1 and 20.
// Returns the number of dice or an error if the level is invalid or the query fails.
func GetNumberOfSneakAttackDiceFromLevel(ctx context.Context, level uint8) (int, error) {
	if level <= 0 || level > 20 {
		return 0, fmt.Errorf("invalid level provided: %d", level)
	}
	numberOfDice := 0

	stmt := SELECT(ClassesSneakAttack.NumberOfDice).
		FROM(ClassesSneakAttack).
		WHERE(ClassesSneakAttack.ClassLevel.LT_EQ(Int(int64(level)))).
		ORDER_BY(ClassesSneakAttack.ClassLevel.DESC()).
		LIMIT(1)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return numberOfDice, fmt.Errorf("failed to query sneak attack by level: %w", err)
	}
	err = row.Scan(&numberOfDice)
	if err != nil {
		return numberOfDice, fmt.Errorf("failed to scan sneak attack by level: %w", err)
	}

	return numberOfDice, nil
}

// getSpellModByClassID retrieves the spellcasting modifier for a character class based on its class ID.
// Returns a core.Ability and an error if the operation fails.
func getSpellModByClassID(ctx context.Context, classID uint8) (core.Ability, error) {
	if classID < 0 || classID > 13 {
		return "", fmt.Errorf("invalid class id provided: %d", classID)
	}

	stmt := SELECT(Classes.Spellmod).
		FROM(Classes).
		WHERE(Classes.ID.EQ(Int(int64(classID))))
	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("failed to query spellmod by class id: %w", err)
	}
	var spellmod pgtype.Text
	err = row.Scan(&spellmod)
	if err != nil {
		return "", fmt.Errorf("failed to scan spellmod by class id: %w", err)
	}
	if spellmod.Valid {
		a, err2 := core.GetNormalizedAbility(spellmod.String)
		if err2 != nil {
			return "", err2
		}
		return a, nil
	}
	return core.AbilityNone, nil
}

// GetHitDieByClassID retrieves the hit die value for a given class ID from the database.
// Returns the hit die as an integer or an error if the query fails or the class ID is invalid.
func GetHitDieByClassID(ctx context.Context, classID int) (int, error) {
	if classID < 0 || classID > 13 {
		return 0, fmt.Errorf("invalid class id provided: %d", classID)
	}

	stmt := SELECT(Classes.HitDie).
		FROM(Classes).
		WHERE(Classes.ID.EQ(Int(int64(classID))))
	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to query hit die by class id: %w", err)
	}
	var die int
	err = row.Scan(&die)
	if err != nil {
		return 0, fmt.Errorf("failed to scan hit die by class id: %w", err)
	}

	return die, nil
}

// GetSpellSlotsByLevelAndClassID retrieves the spell slots for a given character level and class ID.
// Returns a SpellSlots structure and an error if inputs are invalid or a database query fails.
func GetSpellSlotsByLevelAndClassID(ctx context.Context, level uint8, classID uint8) (spells.SpellSlots, error) {
	if level <= 0 || level > 20 {
		return nil, fmt.Errorf("invalid value provided: %d", level)
	}
	if classID < 0 || classID > 13 {
		return nil, fmt.Errorf("invalid class id provided: %d", classID)
	}
	if classID == 2 {
		return nil, fmt.Errorf("barbarian does not have spell slots")
	}

	stmt := SELECT(ClassesSpellSlots.Level1, ClassesSpellSlots.Level2, ClassesSpellSlots.Level3,
		ClassesSpellSlots.Level4, ClassesSpellSlots.Level5, ClassesSpellSlots.Level6,
		ClassesSpellSlots.Level7, ClassesSpellSlots.Level8, ClassesSpellSlots.Level9).
		FROM(ClassesSpellSlots).
		WHERE(ClassesSpellSlots.ClassID.EQ(Int(int64(classID))).AND(ClassesSpellSlots.Level.EQ(Int(int64(level)))))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query spell slots by value and class id: %w", err)
	}

	levels := make([]int, 9)
	err = row.Scan(&levels[0], &levels[1], &levels[2], &levels[3], &levels[4], &levels[5], &levels[6], &levels[7], &levels[8])
	if err != nil {
		return nil, fmt.Errorf("failed to scan spell slots by value and class id: %w", err)
	}

	slots := make(spells.SpellSlots)
	for i, value := range levels {
		slots[i+1] = value
	}

	return slots, nil
}
