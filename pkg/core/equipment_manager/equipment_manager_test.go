package equipment_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/equipment"
	"testing"
)

func TestEquipmentManager_AddItem(t *testing.T) {
	em := NewEquipmentManager()

	item := equipment.Equipment{
		Name: "Plate Armor",
		Type: equipment.EquipmentTypeArmor,
		Armor: &equipment.Armor{
			ArmorClass: 18,
		},
	}

	err := em.AddItem(EquipmentSlotArmor, item)
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}

	equipped := em.GetItem(EquipmentSlotArmor)
	if equipped == nil {
		t.Fatal("Expected item to be equipped")
	}

	if equipped.Name != "Plate Armor" {
		t.Errorf("Expected Plate Armor, got %s", equipped.Name)
	}
}

func TestEquipmentManager_InvalidItem(t *testing.T) {
	em := NewEquipmentManager()

	item := equipment.Equipment{
		Name: "Broken Armor",
		Type: equipment.EquipmentTypeArmor,
		Armor: &equipment.Armor{
			ArmorClass: 0, // Invalid
		},
	}

	err := em.AddItem(EquipmentSlotArmor, item)
	if err == nil {
		t.Error("Expected error for invalid armor class 0")
	}
}
