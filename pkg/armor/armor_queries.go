package armor

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"

	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	. "github.com/go-jet/jet/v2/postgres"
)

// getArmorIDByName retrieves the id of an armor item from the database using its name.
// ctx is the context for managing request-scoped values, deadlines, and cancellation.
// name represents the name of the armor for which the id needs to be fetched.
// Returns the armor id as an integer and an error if the operation fails.
func getArmorIDByName(ctx context.Context, name string) (int, error) {
	var id int
	stmt := SELECT(
		EquipmentArmor.ID,
	).FROM(
		EquipmentArmor,
	).
		WHERE(EquipmentArmor.Name.EQ(String(name)))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return id, fmt.Errorf("error getting armor id by name: %w", err)
	}
	err = row.Scan(&id)
	if err != nil {
		return id, fmt.Errorf("error scanning armorResult by name: %w", err)
	}

	return id, nil
}

// getArmorByID retrieves an armor item from the database using the provided id.
// ctx is the context for managing request-scoped values, deadlines, and cancellation.
// id specifies the unique identifier of the armor to fetch.
// Returns an Armor object and an error if the operation fails.
func getArmorByID(ctx context.Context, id int) (Armor, error) {
	var armorResult Armor
	stmt := SELECT(
		EquipmentArmor.Name,
		EquipmentArmor.ArmorClass,
		EquipmentArmor.DexBonus,
		EquipmentArmor.MaxBonus,
		EquipmentArmor.MinimumStr,
	).FROM(
		EquipmentArmor,
	).
		WHERE(EquipmentArmor.ID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return armorResult, fmt.Errorf("error getting armorResult by id: %w", err)
	}
	err = row.Scan(&armorResult.Name, &armorResult.ArmorClass, &armorResult.DexBonus, &armorResult.MaxBonus, &armorResult.MinimumStr)
	if err != nil {
		return armorResult, fmt.Errorf("error scanning armorResult by id: %w", err)
	}

	return armorResult, nil
}

// QueryArmorData retrieves armor data based on the provided query parameters.
// ctx is the context for managing request-scoped values, deadlines, and cancellation.
// params specifies the query parameters, including armor id or name.
// Returns an Armor object if successful, or an error if the operation fails.
func QueryArmorData(ctx context.Context, params ArmorQueryParams) (Armor, error) {
	var armorResult Armor
	var err error

	if params.ID != 0 {
		armorResult, err = getArmorByID(ctx, params.ID)
	} else if params.Name != "" {
		var id int
		id, err = getArmorIDByName(ctx, params.Name)
		if err != nil {
			return armorResult, err
		}
		armorResult, err = getArmorByID(ctx, id)
	} else {
		err = fmt.Errorf("no name or id provided for armor query")
	}

	return armorResult, err
}
