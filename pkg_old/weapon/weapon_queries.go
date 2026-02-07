package weapon

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"fmt"

	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	. "github.com/go-jet/jet/v2/postgres"
)

// getWeaponIDByName retrieves the id of a weapon based on its name from the database or returns an error if not found.
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

// getWeaponByID retrieves a weapon by its id from the database and returns the weapon or an error if any issues occur.
func getWeaponByID(ctx context.Context, id int) (Weapon, error) {
	var weaponResult Weapon
	var weaponProperties Properties
	stmt := SELECT(
		EquipmentWeapons.ID,
		EquipmentWeapons.Name,
		EquipmentWeapons.IsVersatile,
		EquipmentWeapons.IsFinesse,
		EquipmentWeapons.IsTwoHanded,
		EquipmentWeapons.IsHeavy,
		EquipmentWeapons.IsLight,
		EquipmentWeapons.IsThrown,
		EquipmentWeapons.IsRanged,
		EquipmentWeapons.IsOnlyRanged,
	).FROM(EquipmentWeapons).
		WHERE(EquipmentWeapons.ID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return weaponResult, fmt.Errorf("error getting weaponResult by id: %w", err)
	}
	err = row.Scan(&weaponResult.ID, &weaponResult.Name, &weaponProperties.IsVersatile, &weaponProperties.IsFinesse, &weaponProperties.IsTwoHanded,
		&weaponProperties.IsHeavy, &weaponProperties.IsLight, &weaponProperties.IsThrown, &weaponProperties.IsRanged, &weaponProperties.IsOnlyRanged)
	if err != nil {
		return weaponResult, fmt.Errorf("error scanning weaponResult by id: %w", err)
	}

	weaponResult.Properties = weaponProperties

	// Fetch damage blocks
	stmtDmg := SELECT(
		EquipmentWeaponsDamageBlocks.NumberOfDice,
		EquipmentWeaponsDamageBlocks.Die,
		EquipmentWeaponsDamageBlocks.DmgType,
	).FROM(EquipmentWeaponsDamageBlocks).
		WHERE(EquipmentWeaponsDamageBlocks.WeaponID.EQ(Int(int64(id))))

	qDmg, aDmg := stmtDmg.Sql()
	rows, err := database.Query(ctx, qDmg, aDmg...)
	if err != nil {
		return weaponResult, fmt.Errorf("error getting damage blocks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var n, d int
		var dt string
		rows.Scan(&n, &d, &dt)
		dice, _ := core.MakeDiceType(d)
		dmgType, _ := core.MakeDamageType(dt)
		weaponResult.DamageBlocks = append(weaponResult.DamageBlocks, core.DamageBlock{
			NumberOfDice: n,
			Die:          dice,
			DamageType:   dmgType,
		})
	}

	return weaponResult, nil
}

// QueryWeaponData retrieves detailed weapon data based on either weapon id or name provided in the query parameters.
// Returns a Weapon struct and an error if the query fails or neither name nor id is provided in the params.
func QueryWeaponData(ctx context.Context, params WeaponQueryParams) (Weapon, error) {
	var weaponResult Weapon
	var err error

	if params.ID != 0 {
		weaponResult, err = getWeaponByID(ctx, params.ID)
	} else if params.Name != "" {
		var id int
		id, err = getWeaponIDByName(ctx, params.Name)
		weaponResult, err = getWeaponByID(ctx, id)
	} else {
		err = fmt.Errorf("no name or id provided for weapon query")
	}

	return weaponResult, err
}

func GetWeaponSummaries(ctx context.Context) (map[int]WeaponSummary, error) {
	summaries := make(map[int]WeaponSummary)

	stmt := SELECT(
		EquipmentWeapons.ID,
		EquipmentWeapons.Name,
		EquipmentWeapons.IsVersatile,
		EquipmentWeapons.IsFinesse,
		EquipmentWeapons.IsTwoHanded,
		EquipmentWeapons.IsHeavy,
		EquipmentWeapons.IsLight,
		EquipmentWeapons.IsThrown,
		EquipmentWeapons.IsRanged,
		EquipmentWeapons.IsOnlyRanged,
	).FROM(EquipmentWeapons)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to query weapon summaries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var summary WeaponSummary
		err = rows.Scan(
			&summary.ID,
			&summary.Name,
			&summary.Properties.IsVersatile,
			&summary.Properties.IsFinesse,
			&summary.Properties.IsTwoHanded,
			&summary.Properties.IsHeavy,
			&summary.Properties.IsLight,
			&summary.Properties.IsThrown,
			&summary.Properties.IsRanged,
			&summary.Properties.IsOnlyRanged,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning weapon summaries: %w", err)
		}
		summaries[summary.ID] = summary
	}

	// Fetch all damage blocks
	stmtDmg := SELECT(
		EquipmentWeaponsDamageBlocks.WeaponID,
		EquipmentWeaponsDamageBlocks.NumberOfDice,
		EquipmentWeaponsDamageBlocks.Die,
		EquipmentWeaponsDamageBlocks.DmgType,
	).FROM(EquipmentWeaponsDamageBlocks)

	qDmg, aDmg := stmtDmg.Sql()
	rowsDmg, err := database.Query(ctx, qDmg, aDmg...)
	if err == nil {
		defer rowsDmg.Close()
		for rowsDmg.Next() {
			var wid, n, d int
			var dt string
			rowsDmg.Scan(&wid, &n, &d, &dt)
			if summary, ok := summaries[wid]; ok {
				dice, _ := core.MakeDiceType(d)
				dmgType, _ := core.MakeDamageType(dt)
				summary.DamageBlocks = append(summary.DamageBlocks, core.DamageBlock{
					NumberOfDice: n,
					Die:          dice,
					DamageType:   dmgType,
				})
				summaries[wid] = summary
			}
		}
	}

	return summaries, nil
}
