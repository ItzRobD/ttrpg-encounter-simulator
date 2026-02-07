package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg_old/armor"
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/weapon"
	"fmt"
)

func (c *Character) setupEquipmentFromConfig(ctx context.Context, config EquipmentConfig) error {
	// Handle armor
	if config.ArmorID > 0 {
		a, err := armor.QueryArmorData(ctx, armor.ArmorQueryParams{ID: config.ArmorID})
		if err != nil {
			return fmt.Errorf("failed to get armor id %d: %w", config.ArmorID, err)
		}
		c.EquipmentManager.SetArmor(a)
	}

	// Handle shield
	if config.HasShieldEquipped {
		// In D&D 5e, a standard shield provides +2 AC.
		// We query for a standard shield (ID 2 usually, but let's see if we should query by ID or just set a default)
		// For now, let's stick to the ID provided in the discussion if available, or query for a shield.
		// If ArmorID for shield is not explicitly in config, we assume ID 2 for a standard shield.
		shieldID := 2
		s, err := armor.QueryArmorData(ctx, armor.ArmorQueryParams{ID: shieldID})
		if err != nil {
			// Fallback to a basic shield if query fails
			s = armor.Armor{ID: shieldID, Name: "Shield", ArmorClass: 2}
		}
		c.EquipmentManager.SetShield(s)
	}

	// Handle primary slot weapons
	for _, wConfig := range config.PrimarySlot {
		w, err := weapon.QueryWeaponData(ctx, weapon.WeaponQueryParams{ID: wConfig.WeaponID})
		if err != nil {
			return fmt.Errorf("failed to get w id %d for primary slot: %w", wConfig.WeaponID, err)
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
	// Note: If a shield is equipped, secondary weapons usually can't be used (handled in EquipmentManager)
	for _, wConfig := range config.SecondarySlot {
		w, err := weapon.QueryWeaponData(ctx, weapon.WeaponQueryParams{ID: wConfig.WeaponID})
		if err != nil {
			return fmt.Errorf("failed to get w id %d for secondary slot: %w", wConfig.WeaponID, err)
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
			return fmt.Errorf("failed to get w id %d for ranged slot: %w", wConfig.WeaponID, err)
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
