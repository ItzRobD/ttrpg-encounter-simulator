package character

import (
	"context"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"
	. "github.com/go-jet/jet/v2/postgres"
)

func GetNumberOfAttacksFromLevelAndClass(ctx context.Context, level int, classID int) (int, error) {
	if level <= 0 || level > 20 {
		return 0, fmt.Errorf("invalid level provided: %d", level)
	}
	if classID < 0 || classID > 13 {
		return 0, fmt.Errorf("invalid class id provided: %d", classID)
	}

	numberOfAttacks := 1

	stmt := SELECT(ClassesExtraAttack.NumberOfAttacks).
		FROM(ClassesExtraAttack).
		WHERE(ClassesExtraAttack.ClassID.EQ(Int(int64(classID))).AND(ClassesExtraAttack.ClassLevel.GT_EQ(Int(int64(level))))).
		ORDER_BY(ClassesExtraAttack.ClassLevel.DESC()).
		LIMIT(1)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return numberOfAttacks, fmt.Errorf("failed to query extra attacks by level and class: %w", err)
	}
	err = row.Scan(&numberOfAttacks)
	if err != nil {
		return numberOfAttacks, fmt.Errorf("failed to scan extra attacks by level and class: %w", err)
	}

	if numberOfAttacks == 0 {
		numberOfAttacks = 1
	}
	return numberOfAttacks, nil
}

func GetNumberOfSneakAttackDiceFromLevel(ctx context.Context, level int) (int, error) {
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
