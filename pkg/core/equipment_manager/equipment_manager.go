package equipment_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/equipment"
)

type EquipmentSlot string

const (
	EquipmentSlotArmor     EquipmentSlot = "armor"
	EquipmentSlotShield    EquipmentSlot = "shield"
	EquipmentSlotPrimary   EquipmentSlot = "primary"
	EquipmentSlotSecondary EquipmentSlot = "secondary"
	EquipmentSlotRanged    EquipmentSlot = "ranged"
)

type EquipmentManager struct {
	Items map[EquipmentSlot]equipment.Equipment `json:"items"`
}

// NewEquipmentManager creates a new EquipmentManager.
func NewEquipmentManager() EquipmentManager {
	return EquipmentManager{
		Items: make(map[EquipmentSlot]equipment.Equipment),
	}
}

// AddItem adds an equipment item to the specified slot.
func (em *EquipmentManager) AddItem(slot EquipmentSlot, item equipment.Equipment) error {
	if err := equipment.ValidateItem(&item); err != nil {
		return err
	}

	em.Items[slot] = item
	return nil
}

// GetItem returns the equipment item in the specified slot, or nil if no item is equipped in that slot.
func (em *EquipmentManager) GetItem(slot EquipmentSlot) *equipment.Equipment {
	if item, exists := em.Items[slot]; exists {
		return &item
	}
	return nil
}
