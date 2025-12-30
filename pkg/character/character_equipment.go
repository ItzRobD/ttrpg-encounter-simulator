package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/armor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
)

func (c *Character) setupEquipmentFromConfig(ctx context.Context, config EquipmentConfig) error {
	// Handle armor
	if config.ArmorID > 0 {
		a, err := armor.QueryArmorData(ctx, armor.ArmorQueryParams{ID: config.ArmorID})
		if err != nil {
			return fmt.Errorf("failed to get armor ID %d: %w", config.ArmorID, err)
		}
		c.EquipmentManager.SetArmor(a)
	}

	// Handle primary slot weapons
	for _, wConfig := range config.PrimarySlot {
		w, err := weapon.QueryWeaponData(ctx, weapon.WeaponQueryParams{ID: wConfig.WeaponID})
		if err != nil {
			return fmt.Errorf("failed to get w ID %d for primary slot: %w", wConfig.WeaponID, err)
		}

		if wConfig.Modifiers != nil {
			w.SetModifiers(*wConfig.Modifiers)
		}

		err = c.EquipmentManager.SetWeapon(core.WSPrimary, &w, wConfig.IsProficient)
		if err != nil {
			return fmt.Errorf("failed to set primary w: %w", err)
		}
	}

	// Handle secondary slot weapons
	for _, wConfig := range config.SecondarySlot {
		w, err := weapon.QueryWeaponData(ctx, weapon.WeaponQueryParams{ID: wConfig.WeaponID})
		if err != nil {
			return fmt.Errorf("failed to get w ID %d for secondary slot: %w", wConfig.WeaponID, err)
		}

		if wConfig.Modifiers != nil {
			w.SetModifiers(*wConfig.Modifiers)
		}

		err = c.EquipmentManager.SetWeapon(core.WSSecondary, &w, wConfig.IsProficient)
		if err != nil {
			return fmt.Errorf("failed to set secondary w: %w", err)
		}
	}

	// Handle ranged slot weapons
	for _, wConfig := range config.RangedSlot {
		w, err := weapon.QueryWeaponData(ctx, weapon.WeaponQueryParams{ID: wConfig.WeaponID})
		if err != nil {
			return fmt.Errorf("failed to get w ID %d for ranged slot: %w", wConfig.WeaponID, err)
		}

		if wConfig.Modifiers != nil {
			w.SetModifiers(*wConfig.Modifiers)
		}

		err = c.EquipmentManager.SetWeapon(core.WSRanged, &w, wConfig.IsProficient)
		if err != nil {
			return fmt.Errorf("failed to set ranged w: %w", err)
		}
	}

	return nil
}
