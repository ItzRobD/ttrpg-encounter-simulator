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
	Armor       armor.Armor
	WeaponSlots map[core.WeaponSlot]WeaponSlotData
}

type WeaponSlotData struct {
	Weapon       *weapon.Weapon
	IsProficient bool
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
func (em *EquipmentManager) SetWeapon(slot core.WeaponSlot, w *weapon.Weapon, isProficient bool) {
	ws := em.WeaponSlots[slot]
	ws.Weapon = w
	ws.IsProficient = isProficient
	em.WeaponSlots[slot] = ws
}

// GetWeaponFromSlot retrieves the weapon equipped in the specified slot.
// Returns the weapon and nil if equipped, or nil and an error if no weapon is found in the slot.
func (em *EquipmentManager) GetWeaponFromSlot(slot core.WeaponSlot) (*weapon.Weapon, error) {
	if em.WeaponSlots[slot].Weapon == nil {
		return nil, fmt.Errorf("weapon not equipped in slot %s", slot)
	}
	return em.WeaponSlots[slot].Weapon, nil
}

// GetIsProficientWithSlot checks if the character is proficient with the weapon equipped in the specified weapon slot.
func (em *EquipmentManager) GetIsProficientWithSlot(slot core.WeaponSlot) bool {
	return em.WeaponSlots[slot].IsProficient
}

// SetWeaponProficiencyBySlot sets the proficiency status for the weapon in the specified slot.
func (em *EquipmentManager) SetWeaponProficiencyBySlot(slot core.WeaponSlot, isProficient bool) {
	ws := em.WeaponSlots[slot]
	ws.IsProficient = isProficient
	em.WeaponSlots[slot] = ws
}

func NewEquipmentManager(parent core.Entity) (*EquipmentManager, error) {
	if parent == nil {
		return nil, fmt.Errorf("parent cannot be nil")
	}
	return &EquipmentManager{parent: parent}, nil
}
