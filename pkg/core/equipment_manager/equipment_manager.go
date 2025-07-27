package equipment_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/armor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
)

type EquipmentManager struct {
	parent core.Entity

	// Equipped Items
	Armor            armor.Armor
	Weapons          map[core.WeaponSlot]*WeaponSlotData
	WeaponAttackData map[core.WeaponSlot]WeaponAttackData
}

type WeaponSlotData struct {
	Weapon       *weapon.Weapon
	IsProficient bool
}

type WeaponAttackData struct {
	Normal    core.AttackData
	Versatile *core.AttackData
}

func NewEquipmentManager(parent core.Entity) (*EquipmentManager, error) {
	if parent == nil {
		return nil, fmt.Errorf("parent cannot be nil")
	}
	return &EquipmentManager{parent: parent}, nil
}

func (em *EquipmentManager) GetParent() core.Entity {
	return em.parent
}

func (em *EquipmentManager) SetArmor(a armor.Armor) {
	em.Armor = a
}

func (em *EquipmentManager) GetArmor() armor.Armor {
	return em.Armor
}

// GetAC returns the Armor Class (AC) of the equipped armor in the EquipmentManager.
func (em *EquipmentManager) GetAC() int {
	return em.Armor.ArmorClass
}

// SetWeapon equips a weapon in the specified slot and sets proficiency status for the EquipmentManager.
func (em *EquipmentManager) SetWeapon(slot core.WeaponSlot, w *weapon.Weapon, isProficient bool) error {
	ws := em.Weapons[slot]
	ws.Weapon = w
	ws.IsProficient = isProficient
	em.Weapons[slot] = ws

	err := em.computeAttackDataForSlot(slot)
	if err != nil {
		delete(em.Weapons, slot)
		return fmt.Errorf("failed to compute attack data for weapon: %w", err)
	}

	return nil
}

// GetWeaponFromSlot retrieves the weapon equipped in the specified slot.
// Returns the weapon and nil if equipped, or nil and an error if no weapon is found in the slot.
func (em *EquipmentManager) GetWeaponFromSlot(slot core.WeaponSlot) (*weapon.Weapon, error) {
	if em.Weapons[slot].Weapon == nil {
		return nil, fmt.Errorf("weapon not equipped in slot %s", slot)
	}
	return em.Weapons[slot].Weapon, nil
}

// GetIsProficientWithSlot checks if the character is proficient with the weapon equipped in the specified weapon slot.
func (em *EquipmentManager) GetIsProficientWithSlot(slot core.WeaponSlot) bool {
	return em.Weapons[slot].IsProficient
}

// SetWeaponProficiencyBySlot sets the proficiency status for the weapon in the specified slot.
func (em *EquipmentManager) SetWeaponProficiencyBySlot(slot core.WeaponSlot, isProficient bool) {
	ws := em.Weapons[slot]
	ws.IsProficient = isProficient
	em.Weapons[slot] = ws
}

// GetWeaponAttackData retrieves the attack data for a weapon in a specified slot, optionally using versatile attack data.
// Returns the retrieved attack data and any error if the slot has no data or if any issues occur.
func (em *EquipmentManager) GetWeaponAttackData(slot core.WeaponSlot, useVersatile bool) (core.AttackData, error) {
	ad, exists := em.WeaponAttackData[slot]
	if !exists {
		return core.AttackData{}, fmt.Errorf("no attack data for weapon slot %w", slot)
	}

	if useVersatile && ad.Versatile != nil {
		return *ad.Versatile, nil
	}

	return ad.Normal, nil
}

// CanUseVersatile checks whether the weapon in the specified slot can utilize versatile attack options.
func (em *EquipmentManager) CanUseVersatile(slot core.WeaponSlot) bool {
	if options, exists := em.WeaponAttackData[slot]; exists {
		return options.Versatile != nil
	}
	return false
}

// GetAvailableWeaponSlots returns a slice of all available weapon slots that have weapon attack data in the EquipmentManager.
func (em *EquipmentManager) GetAvailableWeaponSlots() []core.WeaponSlot {
	slots := make([]core.WeaponSlot, 0, len(em.WeaponAttackData))
	for slot := range em.WeaponAttackData {
		slots = append(slots, slot)
	}
	return slots
}

// computeAttackDataForSlot calculates and sets the attack data for a given weapon slot in the EquipmentManager.
// Returns an error if the weapon does not exist in the slot or if any calculation-related operation fails.
func (em *EquipmentManager) computeAttackDataForSlot(slot core.WeaponSlot) error {
	w, exists := em.Weapons[slot]
	if !exists {
		return fmt.Errorf("no w in slot %v", slot)
	}

	prof := em.GetIsProficientWithSlot(slot)

	as := em.parent.GetAbilityScores()
	attackMod, err := w.Weapon.GetAttackModifier(&as, uint8(em.parent.GetLevel()), prof)
	if err != nil {
		return err
	}

	damageMod, err := w.Weapon.GetWeaponModifier(&as)
	if err != nil {
		return err
	}

	normal := core.AttackData{
		Name:              w.Weapon.Name,
		NumberOfDice:      w.Weapon.NumberOfDice,
		Die:               w.Weapon.Die,
		AttackModifier:    attackMod,
		DamageModifier:    damageMod,
		DamageType:        w.Weapon.DamageType,
		IsVersatileAttack: false,
	}
	weaponData := WeaponAttackData{Normal: normal}

	if w.Weapon.IsVersatile {
		versatile := core.AttackData{
			Name:              w.Weapon.Name,
			NumberOfDice:      w.Weapon.NumberOfDice,
			Die:               w.Weapon.Die + 2,
			AttackModifier:    attackMod,
			DamageModifier:    damageMod,
			DamageType:        w.Weapon.DamageType,
			IsVersatileAttack: true,
		}
		weaponData.Versatile = &versatile
	}

	em.WeaponAttackData[slot] = weaponData

	return nil
}
