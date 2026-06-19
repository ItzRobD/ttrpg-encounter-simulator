package equipment_manager

import (
	"dnd5e-encounter-simulator-backend/pkg_old/armor"
	"dnd5e-encounter-simulator-backend/pkg_old/classes"
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/weapon"
	"fmt"
	"sort"
)

type EquipmentManager struct {
	parent core.Entity

	// Equipped items
	Armor             armor.Armor
	Shield            armor.Armor
	HasShieldEquipped bool
	Weapons           map[core.WeaponSlot]*WeaponSlotData
	WeaponAttackData  map[core.WeaponSlot]WeaponAttackData
	HasDefenseStyle   bool
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
	return &EquipmentManager{
		parent:           parent,
		Weapons:          make(map[core.WeaponSlot]*WeaponSlotData),
		WeaponAttackData: make(map[core.WeaponSlot]WeaponAttackData),
	}, nil
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

// SetShield equips a shield in the EquipmentManager. Secondary weapon, if equipped, cannot be used while a shield is equipped.
func (em *EquipmentManager) SetShield(s armor.Armor) {
	em.Shield = s
	em.HasShieldEquipped = true
}

func (em *EquipmentManager) GetShield() armor.Armor {
	return em.Shield
}

func (em *EquipmentManager) GetHasShieldEquipped() bool {
	return em.HasShieldEquipped
}

func (em *EquipmentManager) SetHasDefenseStyle(val bool) {
	em.HasDefenseStyle = val
}

func (em *EquipmentManager) GetHasDefenseStyle() bool {
	return em.HasDefenseStyle
}

// GetAC returns the Armor ClassID (AC) of the equipped armor in the EquipmentManager.
func (em *EquipmentManager) GetAC() int {
	base := 10
	returnValue := base

	dexMod, err := em.parent.GetAbilityScoreModifier(core.AbilityDexterity)
	if err != nil {
		return -1
	}

	if em.Armor == (armor.Armor{}) {
		// Unarmored
		id := em.parent.GetClassID()
		if classes.ClassID(id) == classes.Monk && !em.HasShieldEquipped { // unarmored defense
			wisMod, err := em.parent.GetAbilityScoreModifier(core.AbilityWisdom)
			if err != nil {
				return -1
			}
			returnValue = base + dexMod + wisMod
		} else if classes.ClassID(id) == classes.Barbarian {
			conMod, err := em.parent.GetAbilityScoreModifier(core.AbilityConstitution)
			if err != nil {
				return -1
			}
			returnValue = base + dexMod + conMod
		} else {
			returnValue = base + dexMod
		}
	} else {
		// Armored
		if em.Armor.DexBonus {
			if em.Armor.MaxBonus { // medium armor
				if dexMod > 2 {
					dexMod = 2
				}
				returnValue = em.Armor.ArmorClass + dexMod
			} else { // light armor
				returnValue = em.Armor.ArmorClass + dexMod
			}
		} else { // heavy armor
			returnValue = em.Armor.ArmorClass
		}

		if em.HasDefenseStyle {
			returnValue += 1
		}
	}

	if em.HasShieldEquipped {
		returnValue += em.Shield.ArmorClass
	}

	return returnValue
}

// SetWeapon equips a weapon in the specified slot and sets proficiency status for the EquipmentManager.
func (em *EquipmentManager) SetWeapon(slot core.WeaponSlot, w *weapon.Weapon, isProficient bool) error {
	ws := em.Weapons[slot]
	if ws == nil {
		ws = &WeaponSlotData{} // Create new WeaponSlotData if it doesn't exist
	}

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
	if em.Weapons == nil {
		return nil, fmt.Errorf("weapon not equipped in slot %s", slot)
	}
	ws, ok := em.Weapons[slot]
	if !ok || ws == nil || ws.Weapon == nil {
		return nil, fmt.Errorf("weapon not equipped in slot %s", slot)
	}
	return ws.Weapon, nil
}

// GetIsProficientWithSlot checks if the character is proficient with the weapon equipped in the specified weapon slot.
func (em *EquipmentManager) GetIsProficientWithSlot(slot core.WeaponSlot) bool {
	if em.Weapons == nil {
		return false
	}
	ws, ok := em.Weapons[slot]
	if !ok || ws == nil {
		return false
	}
	return ws.IsProficient
}

// SetWeaponProficiencyBySlot sets the proficiency status for the weapon in the specified slot.
func (em *EquipmentManager) SetWeaponProficiencyBySlot(slot core.WeaponSlot, isProficient bool) {
	if em.Weapons == nil {
		em.Weapons = make(map[core.WeaponSlot]*WeaponSlotData)
	}
	ws := em.Weapons[slot]
	if ws == nil {
		ws = &WeaponSlotData{}
	}
	ws.IsProficient = isProficient
	em.Weapons[slot] = ws
}

// HasMeleeWeapon checks if a melee weapon is equipped in any weapon slot of the EquipmentManager.
func (em *EquipmentManager) HasMeleeWeapon() bool {
	if em.Weapons == nil {
		return false
	}

	for _, weaponSlot := range em.Weapons {
		if weaponSlot != nil && weaponSlot.Weapon != nil && !weaponSlot.Weapon.Properties.IsOnlyRanged {
			return true
		}
	}
	return false
}

func (em *EquipmentManager) HasRangedWeapon() bool {
	if em.Weapons == nil {
		return false
	}

	for _, weaponSlot := range em.Weapons {
		if weaponSlot != nil && weaponSlot.Weapon != nil && weaponSlot.Weapon.Properties.IsRanged {
			return true
		}
	}
	return false
}

// GetWeaponAttackData retrieves the attack data for a weapon in a specified slot, optionally using versatile attack data.
// Returns the retrieved attack data and any error if the slot has no data or if any issues occur.
func (em *EquipmentManager) GetWeaponAttackData(slot core.WeaponSlot, useVersatile bool) (core.AttackData, error) {
	ad, exists := em.WeaponAttackData[slot]
	if !exists {
		return core.AttackData{}, fmt.Errorf("no attack data for weapon slot %s", slot.String())
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
// The returned slots are sorted by their integer value to ensure deterministic iteration.
func (em *EquipmentManager) GetAvailableWeaponSlots() []core.WeaponSlot {
	slots := make([]core.WeaponSlot, 0, len(em.WeaponAttackData))
	for slot := range em.WeaponAttackData {
		slots = append(slots, slot)
	}

	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})

	return slots
}

// computeAttackDataForSlot calculates and sets the attack data for a given weapon slot in the EquipmentManager.
// Returns an error if the weapon does not exist in the slot or if any calculation-related operation fails.
func (em *EquipmentManager) computeAttackDataForSlot(slot core.WeaponSlot) error {
	w, exists := em.Weapons[slot]
	if !exists || w == nil {
		return fmt.Errorf("no weapon in slot %s", slot.String())
	}

	prof := em.GetIsProficientWithSlot(slot)

	as := em.parent.GetAbilityScores()
	attackMod, err := w.Weapon.GetAttackModifier(&as, uint8(em.parent.GetLevel()), prof)
	if err != nil {
		return err
	}

	damageMod, ability, err := w.Weapon.GetWeaponModifier(&as)
	if err != nil {
		return err
	}

	resistBreakers := make([]core.ResistBreaker, 0)
	resistBreakers = append(resistBreakers, w.Weapon.GetResistBreakers()...)

	normal := core.AttackData{
		Name:              w.Weapon.Name,
		DamageBlocks:      w.Weapon.DamageBlocks,
		AttackModifier:    attackMod + w.Weapon.GetAttackBonus(),
		AbilityUsed:       ability,
		DamageModifier:    damageMod + w.Weapon.GetDamageBonus(),
		ResistBreakers:    resistBreakers,
		IsVersatileAttack: false,
		IsRangedWeapon:    w.Weapon.Properties.IsRanged,
		IsTwoHandedWeapon: w.Weapon.Properties.IsTwoHanded,
		IsFinesseWeapon:   w.Weapon.Properties.IsFinesse,
		IsOnlyRanged:      w.Weapon.Properties.IsOnlyRanged,
		IsLightWeapon:     w.Weapon.Properties.IsLight,
		IsThrownWeapon:    w.Weapon.Properties.IsThrown,
		IsHeavyWeapon:     w.Weapon.Properties.IsHeavy,
		WeaponAttackBonus: w.Weapon.GetAttackBonus(),
		WeaponDamageBonus: w.Weapon.GetDamageBonus(),
	}
	weaponData := WeaponAttackData{Normal: normal}

	if w.Weapon.Properties.IsVersatile {
		vDamageBlocks := make([]core.DamageBlock, len(w.Weapon.DamageBlocks))
		copy(vDamageBlocks, w.Weapon.DamageBlocks)
		if len(vDamageBlocks) > 0 {
			vDamageBlocks[0].Die = vDamageBlocks[0].Die + 2
		}

		versatile := core.AttackData{
			Name:              w.Weapon.Name,
			DamageBlocks:      vDamageBlocks,
			AttackModifier:    attackMod + w.Weapon.GetAttackBonus(),
			DamageModifier:    damageMod + w.Weapon.GetDamageBonus(),
			ResistBreakers:    resistBreakers,
			IsVersatileAttack: true,
			WeaponAttackBonus: w.Weapon.GetAttackBonus(),
			WeaponDamageBonus: w.Weapon.GetDamageBonus(),
		}
		weaponData.Versatile = &versatile
	}

	em.WeaponAttackData[slot] = weaponData

	return nil
}
