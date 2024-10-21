package weapon

import (
	"context"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"
	. "github.com/go-jet/jet/v2/postgres"
)

func getWeaponIDByName(ctx context.Context, name string) (int, error) {
	var id int
	stmt := SELECT(
		EquipmentWeapons.ID,
	).
		FROM(EquipmentWeapons).
		WHERE(EquipmentWeapons.Name.EQ(String(name)))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return id, fmt.Errorf("error getting weapon id by name: %w", err)
	}
	err = row.Scan(&id)
	if err != nil {
		return id, fmt.Errorf("error scanning weapon id by name: %w", err)
	}

	return id, nil
}

func getWeaponByID(ctx context.Context, id int) (Weapon, error) {
	var weaponResult Weapon
	stmt := SELECT(
		EquipmentWeapons.Name,
		EquipmentWeapons.IsVersatile,
		EquipmentWeapons.IsFinesse,
		EquipmentWeaponsDamageBlocks.NumberOfDice,
		EquipmentWeaponsDamageBlocks.Die,
		EquipmentWeaponsDamageBlocks.DmgType,
	).FROM(
		EquipmentWeapons.
			LEFT_JOIN(EquipmentWeaponsDamageBlocks, EquipmentWeapons.ID.EQ(EquipmentWeaponsDamageBlocks.WeaponID))).
		WHERE(
			EquipmentWeapons.ID.EQ(Int(int64(id))),
		)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return weaponResult, fmt.Errorf("error getting weaponResult by id: %w", err)
	}
	err = row.Scan(&weaponResult.Name, &weaponResult.IsVersatile, &weaponResult.IsFinesse, &weaponResult.NumberOfDice,
		&weaponResult.Die, &weaponResult.DamageType)
	if err != nil {
		return weaponResult, fmt.Errorf("error scanning weaponResult by id: %w", err)
	}

	return weaponResult, nil
}

func QueryWeaponData(ctx context.Context, params WeaponQueryParams) (Weapon, error) {
	var weaponResult Weapon
	var err error

	if params.ID != 0 {
		weaponResult, err = getWeaponByID(ctx, params.ID)
		weaponResult.IsRanged = isRangedWeapon(params.ID)
	} else if params.Name != "" {
		var id int
		id, err = getWeaponIDByName(ctx, params.Name)
		weaponResult, err = getWeaponByID(ctx, id)
		weaponResult.IsRanged = isRangedWeapon(id)
	} else {
		err = fmt.Errorf("no name or id provided for weapon query")
	}

	return weaponResult, err
}
