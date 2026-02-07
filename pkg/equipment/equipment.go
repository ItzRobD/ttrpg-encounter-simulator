package equipment

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"errors"
)

type EquipmentType string

const (
	EquipmentTypeArmor  EquipmentType = "armor"
	EquipmentTypeWeapon EquipmentType = "weapon"
	EquipmentTypeShield EquipmentType = "shield"
)

type Equipment struct {
	ID       core.ID       `json:"id"`
	Name     string        `json:"name"`
	IsCustom bool          `json:"is_custom"`
	Type     EquipmentType `json:"type"`
	Armor    *Armor        `json:"armor,omitempty"`
	Weapon   *Weapon       `json:"weapon,omitempty"`
}

type Armor struct {
	ArmorClass int  `json:"ac"`
	DexBonus   bool `json:"dex_bonus"`   // DexBonus indicates whether the Dexterity bonus is applied to the armor class.
	MaxBonus   bool `json:"max_bonus"`   // MaxBonus indicates whether the armor class is limited to 2 maximum.
	MinimumStr int  `json:"minimum_str"` // MinimumStr indicates the minimum strength required to equip the armor.
	Modifier   int  `json:"modifier"`    // Modifier indicates the armor class modifier.
}

type Weapon struct {
	DamageBlocks []core.DiceBlock      `json:"damage_blocks"`
	Properties   core.WeaponProperties `json:"properties"`
	Modifiers    core.WeaponModifiers  `json:"modifiers"`
}

func (w *Weapon) GetAverageDamage() int {
	avgDamage := 0
	for _, block := range w.DamageBlocks {
		avg, _ := core.GetAverageRoll(block.NumberOfDice, block.Die, block.Modifier)
		avgDamage += avg
	}
	return avgDamage
}

func (w *Weapon) GetMaxDamage() int {
	maxDamage := 0
	for _, block := range w.DamageBlocks {
		maxDamage += block.NumberOfDice*block.Die.Max() + block.Modifier
	}
	return maxDamage
}

func (w *Weapon) GetMinDamage() int {
	minDamage := 0
	for _, block := range w.DamageBlocks {
		minDamage += block.NumberOfDice*block.Die.Min() + block.Modifier
	}
	return minDamage
}

func ValidateItem(item *Equipment) error {
	switch item.Type {
	case EquipmentTypeArmor:
		if item.Armor.ArmorClass <= 0 {
			return errors.New("armor class must be positive")
		}
		if item.Armor.MinimumStr < 0 || item.Armor.MinimumStr > 30 {
			return errors.New("minimum strength requirement must be between 0 and 30")
		}
	case EquipmentTypeWeapon:
		if len(item.Weapon.DamageBlocks) <= 0 {
			return errors.New("weapon must have at least one damage block")
		}
		if item.Weapon.Properties.IsOnlyRanged && !item.Weapon.Properties.IsRanged {
			return errors.New("weapon cannot be only ranged if it is not a ranged weapon")
		}
	}
	return nil
}
