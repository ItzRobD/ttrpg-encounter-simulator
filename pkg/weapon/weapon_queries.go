package weapon

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"
)

func getWeaponByName(ctx context.Context, name string) (Weapon, error) {
	var weaponResult Weapon
	query := `
		SELECT
			w.name,
			w.is_versatile,
			ewdb.number_of_dice,
			ewdb.die,
			ewdb.dmg_type
		FROM equipment_weapons w
		JOIN public.equipment_weapons_damage_blocks ewdb on w.id = ewdb.weapon_id
		WHERE w.name ILIKE $1`

	row, err := database.QueryRow(ctx, query, name)
	if err != nil {
		return weaponResult, fmt.Errorf("error getting weaponResult by name: %w", err)
	}
	err = row.Scan(&weaponResult.Name, &weaponResult.IsVersatile, &weaponResult.NumberOfDice,
		&weaponResult.Die, &weaponResult.DamageType)
	if err != nil {
		return weaponResult, fmt.Errorf("error scanning weaponResult by name: %w", err)
	}

	return weaponResult, nil
}

func getWeaponByID(ctx context.Context, id int) (Weapon, error) {
	var weaponResult Weapon
	query := `
		SELECT
			w.name,
			w.is_versatile,
			ewdb.number_of_dice,
			ewdb.die,
			ewdb.dmg_type
		FROM equipment_weapons w
		JOIN public.equipment_weapons_damage_blocks ewdb on w.id = ewdb.weapon_id
		WHERE w.id = $1`

	row, err := database.QueryRow(ctx, query, id)
	if err != nil {
		return weaponResult, fmt.Errorf("error getting weaponResult by id: %w", err)
	}
	err = row.Scan(&weaponResult.Name, &weaponResult.IsVersatile, &weaponResult.NumberOfDice,
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
	} else if params.Name != "" {
		weaponResult, err = getWeaponByName(ctx, params.Name)
	} else {
		err = fmt.Errorf("no name or id provided for weapon query")
	}

	return weaponResult, err
}
