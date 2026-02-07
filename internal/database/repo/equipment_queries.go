package repo

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/equipment"
	"fmt"
	"strconv"

	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

func GetAllWeaponData(ctx context.Context) ([]equipment.Equipment, error) {
	stmt := SELECT(
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
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var results []equipment.Equipment
	for rows.Next() {
		var w equipment.Equipment
		w.Weapon = &equipment.Weapon{}
		err = rows.Scan(
			&w.Name,
			&w.Weapon.Properties.IsVersatile,
			&w.Weapon.Properties.IsFinesse,
			&w.Weapon.Properties.IsTwoHanded,
			&w.Weapon.Properties.IsHeavy,
			&w.Weapon.Properties.IsLight,
			&w.Weapon.Properties.IsThrown,
			&w.Weapon.Properties.IsRanged,
			&w.Weapon.Properties.IsOnlyRanged,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		results = append(results, w)
	}

	return results, nil
}

func GetAllArmorData(ctx context.Context) ([]equipment.Equipment, error) {
	stmt := SELECT(
		EquipmentArmor.Name,
		EquipmentArmor.ArmorClass,
		EquipmentArmor.DexBonus,
		EquipmentArmor.MaxBonus,
		EquipmentArmor.MinimumStr,
	).FROM(EquipmentArmor)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var results []equipment.Equipment
	for rows.Next() {
		var a equipment.Equipment
		a.Armor = &equipment.Armor{}
		err = rows.Scan(
			&a.Name,
			&a.Armor.ArmorClass,
			&a.Armor.DexBonus,
			&a.Armor.MaxBonus,
			&a.Armor.MinimumStr,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		results = append(results, a)
	}
	return results, nil
}

func HydrateArmorData(ctx context.Context, armorID string) (*equipment.Equipment, error) {
	var err error
	_, err = uuid.Parse(armorID)
	if err == nil {
		return nil, nil
	} // armorID is a valid uuid -> hydrate from custom table

	id, err := strconv.Atoi(armorID)
	if err != nil {
		return nil, fmt.Errorf("unable to parse armor id: %w", err)
	}
	if id <= 0 || id > HighestSRDArmorID {
		return nil, fmt.Errorf("invalid armor id: %d", id)
	}

	var armor equipment.Equipment
	armor.Armor = &equipment.Armor{}

	stmt := SELECT(
		EquipmentArmor.Name,
		EquipmentArmor.ArmorClass,
		EquipmentArmor.DexBonus,
		EquipmentArmor.MaxBonus,
		EquipmentArmor.MinimumStr,
	).FROM(EquipmentArmor).WHERE(EquipmentArmor.ID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	err = row.Scan(
		&armor.Name,
		&armor.Armor.ArmorClass,
		&armor.Armor.DexBonus,
		&armor.Armor.MaxBonus,
		&armor.Armor.MinimumStr,
	)
	if err != nil {
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	armor.ID = core.MakeID(armorID)
	armor.IsCustom = false
	armor.Type = equipment.EquipmentTypeArmor

	return &armor, nil
}

func HydrateWeaponData(ctx context.Context, weaponID string) (*equipment.Equipment, error) {
	var err error
	_, err = uuid.Parse(weaponID)
	if err == nil {
		return nil, nil
	} // weaponID is a valid uuid -> hydrate from custom table

	id, err := strconv.Atoi(weaponID)
	if err != nil {
		return nil, fmt.Errorf("unable to parse weapon id: %w", err)
	}
	if id <= 0 || id > HighestSRDWeaponID {
		return nil, fmt.Errorf("invalid weapon id: %d", id)
	}

	stmt := SELECT(
		EquipmentWeapons.Name,
		EquipmentWeapons.IsVersatile,
		EquipmentWeapons.IsFinesse,
		EquipmentWeapons.IsTwoHanded,
		EquipmentWeapons.IsHeavy,
		EquipmentWeapons.IsLight,
		EquipmentWeapons.IsThrown,
		EquipmentWeapons.IsRanged,
		EquipmentWeapons.IsOnlyRanged,
	).FROM(EquipmentWeapons).WHERE(EquipmentWeapons.ID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	var weapon equipment.Equipment
	weapon.Weapon = &equipment.Weapon{}
	err = row.Scan(
		&weapon.Name,
		&weapon.Weapon.Properties.IsVersatile,
		&weapon.Weapon.Properties.IsFinesse,
		&weapon.Weapon.Properties.IsTwoHanded,
		&weapon.Weapon.Properties.IsHeavy,
		&weapon.Weapon.Properties.IsLight,
		&weapon.Weapon.Properties.IsThrown,
		&weapon.Weapon.Properties.IsRanged,
		&weapon.Weapon.Properties.IsOnlyRanged,
	)
	if err != nil {
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	weapon.ID = core.MakeID(weaponID)
	weapon.IsCustom = false
	weapon.Type = equipment.EquipmentTypeWeapon

	// Fetch damage blocks
	stmtDmg := SELECT(
		EquipmentWeaponsDamageBlocks.NumberOfDice,
		EquipmentWeaponsDamageBlocks.Die,
		EquipmentWeaponsDamageBlocks.DmgType,
	).FROM(EquipmentWeaponsDamageBlocks).WHERE(EquipmentWeaponsDamageBlocks.WeaponID.EQ(Int(int64(id))))

	qDmg, aDmg := stmtDmg.Sql()
	rows, err := database.Query(ctx, qDmg, aDmg...)
	if err != nil {
		return nil, fmt.Errorf("error getting damage blocks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var n, d int
		var dt string
		err = rows.Scan(&n, &d, &dt)
		if err != nil {
			return nil, fmt.Errorf("error scanning damage block: %w", err)
		}
		dice, _ := core.MakeDiceType(d)
		dmgType, _ := core.MakeDamageType(dt)
		weapon.Weapon.DamageBlocks = append(weapon.Weapon.DamageBlocks, core.DiceBlock{
			NumberOfDice: n,
			Die:          dice,
			DamageType:   dmgType,
		})
	}

	return &weapon, nil
}

func HydrateEquipment(ctx context.Context, cfg *actor.ActorConfig) ([]equipment.Equipment, error) {
	var results []equipment.Equipment

	// SRD Equipment
	for _, eCfg := range cfg.EquipmentConfigs {
		if eCfg.Type == equipment.EquipmentTypeWeapon {
			w, err := HydrateWeaponData(ctx, eCfg.ID)
			if err == nil && w != nil {
				results = append(results, *w)
				continue
			}
		} else {
			a, err := HydrateArmorData(ctx, eCfg.ID)
			if err == nil && a != nil {
				results = append(results, *a)
				continue
			}
		}
	}

	// Custom Equipment
	results = append(results, cfg.CustomEquipment...)

	return results, nil
}
