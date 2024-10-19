package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/armor"
	"dnd5e-encounter-simulator-backend/pkg/class"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
)

type Character struct {
	Name          string
	Class         class.Class
	Level         int
	AbilityScores shared.AbilityScores
	HP            shared.PlayerHP
	Eq            Equipment
	KnownSpells   []spells.Spell
}

type Equipment struct {
	Armor     armor.Armor   `json:"armor"`
	Primary   weapon.Weapon `json:"primary"`
	Secondary weapon.Weapon `json:"secondary"`
	Ranged    weapon.Weapon `json:"ranged"`
}

func New(ctx context.Context, name string, classID int, level int, abilityScores shared.AbilityScores, hp shared.PlayerHP) (Character, error) {
	if classID < 0 || classID > 13 {
		return Character{}, fmt.Errorf("invalid class id during character initialization: %d", classID)
	}
	if level < 0 || level > 20 {
		return Character{}, fmt.Errorf("invalid level during character initialization, must be in range 1-20: %d", level)
	}
	if name == "" {
		name = "Unnamed Character"
	}

	var params class.ClassQueryParams
	params.ID = classID
	c, err := class.QueryClassData(ctx, params)
	if err != nil {
		return Character{}, err
	}

	return Character{
		Name:          name,
		Class:         c,
		Level:         level,
		AbilityScores: abilityScores,
		HP:            hp,
		Eq:            Equipment{},
		KnownSpells:   []spells.Spell{},
	}, nil
}

func (c *Character) AddSRDArmor(ctx context.Context, armorID int) error {
	if armorID < 0 || armorID > 13 {
		return fmt.Errorf("invalid armor id: %d", armorID)
	}
	var params armor.ArmorQueryParams
	params.ID = armorID
	a, err := armor.QueryArmorData(ctx, params)
	if err != nil {
		return fmt.Errorf("error querying armor data: %w", err)
	}

	c.Eq.Armor = a
	return nil
}

func (c *Character) AddCustomArmor(name string, armorClass int, dexBonus bool, maxBonus bool, minimumStr int) error {
	a, err := armor.New(name, armorClass, dexBonus, maxBonus, minimumStr)
	if err != nil {
		return err
	}

	c.Eq.Armor = a
	return nil
}

func (c *Character) AddSRDWeapon(ctx context.Context, weaponID int, slot string) error {
	if weaponID < 0 || weaponID > 35 {
		return fmt.Errorf("invalid weapon id: %d", weaponID)
	}
	if slot != "primary" && slot != "secondary" && slot != "ranged" {
		return fmt.Errorf("invalid slot identifier provided: %s", slot)
	}

	var params weapon.WeaponQueryParams
	params.ID = weaponID
	w, err := weapon.QueryWeaponData(ctx, params)
	if err != nil {
		return fmt.Errorf("error querying weapon data: %w", err)
	}

	err = c.addWeapon(w, slot)
	if err != nil {
		return err
	}

	return nil
}

func (c *Character) AddCustomWeapon(name string, isVersatile bool, numberOfDice int, die int, damageType string, isRanged bool, slot string) error {
	w, err := weapon.New(name, isVersatile, numberOfDice, die, damageType, isRanged)
	if err != nil {
		return err
	}

	err = c.addWeapon(w, slot)
	if err != nil {
		return err
	}

	return nil
}

func (c *Character) addWeapon(w weapon.Weapon, slot string) error {
	if !w.IsRanged && slot == "ranged" {
		return fmt.Errorf("cannot assign non ranged weapon to ranged slot")
	}
	switch slot {
	case "primary":
		c.Eq.Primary = w
		return nil
	case "secondary":
		c.Eq.Secondary = w
		return nil
	case "ranged":
		c.Eq.Ranged = w
		return nil
	default:
		return fmt.Errorf("invalid slot identifier provided: %s", slot)
	}
}

func spellExists(spells []spells.Spell, spellID int) bool {
	for _, s := range spells {
		if s.ID == spellID {
			return true
		}
	}
	return false
}

func isValidSpell(c *Character, spellID int) bool {
	if spellID < 0 || spellID > 319 {
		return false
	}

	if spellExists(c.Class.Spellcasting.ClassHealingSpells, spellID) {
		return true
	}
	if spellExists(c.Class.Spellcasting.ClassDamageSpells, spellID) {
		return true
	}

	return false
}

func (c *Character) AddKnownSpell(ctx context.Context, spellID int) error {
	if !isValidSpell(c, spellID) {
		return fmt.Errorf("spell id invalid or not available to class: %d", spellID)
	}
	var params spells.SpellQueryParams
	params.ID = spellID
	// TODO: This might cause an issue: if the spell is a cantrip or a leveled spell lower than the character level
	params.Level = c.Level
	s, err := spells.QuerySpellData(ctx, params)
	if err != nil {
		return fmt.Errorf("error querying spell data: %w", err)
	}
	c.KnownSpells = append(c.KnownSpells, s)
	return nil
}
