package armor

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"
)

func getArmorByName(ctx context.Context, name string) (Armor, error) {
	var armorResult Armor
	query := `
		SELECT
		    name, 
		    armor_class,
		    dex_bonus,
		    max_bonus, 
		    minimum_str 
		FROM equipment_armor WHERE name ILIKE $1`

	row, err := database.QueryRow(ctx, query, name)
	if err != nil {
		return armorResult, fmt.Errorf("error getting armorResult by name: %w", err)
	}
	err = row.Scan(&armorResult.Name, &armorResult.ArmorClass, &armorResult.DexBonus, &armorResult.MaxBonus, &armorResult.MinimumStr)
	if err != nil {
		return armorResult, fmt.Errorf("error scanning armorResult by name: %w", err)
	}

	return armorResult, nil
}

func getArmorByID(ctx context.Context, id int) (Armor, error) {
	var armorResult Armor
	query := `
		SELECT
		    name, 
		    armor_class,
		    dex_bonus,
		    max_bonus, 
		    minimum_str 
		FROM equipment_armor WHERE id ILIKE $1`

	row, err := database.QueryRow(ctx, query, id)
	if err != nil {
		return armorResult, fmt.Errorf("error getting armorResult by id: %w", err)
	}
	err = row.Scan(&armorResult.Name, &armorResult.ArmorClass, &armorResult.DexBonus, &armorResult.MaxBonus, &armorResult.MinimumStr)
	if err != nil {
		return armorResult, fmt.Errorf("error scanning armorResult by id: %w", err)
	}

	return armorResult, nil
}

func QueryArmorData(ctx context.Context, params ArmorQueryParams) (Armor, error) {
	var armorResult Armor
	var err error

	if params.ID != 0 {
		armorResult, err = getArmorByID(ctx, params.ID)
	} else if params.Name != "" {
		armorResult, err = getArmorByName(ctx, params.Name)
	} else {
		err = fmt.Errorf("no name or id provided for armor query")
	}

	return armorResult, err
}
