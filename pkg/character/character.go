package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/armor"
	"dnd5e-encounter-simulator-backend/pkg/class"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/combat"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
)

type Character struct {
	Name              string
	Class             class.Class
	Level             int
	AbilityScores     shared.AbilityScores
	HP                shared.PlayerHP
	Eq                Equipment
	WeaponProficiency WeaponProficiencies
	KnownSpells       []spells.Spell
	SpellSlots        shared.SpellSlots
	ActionPreference  shared.ActionPreference
	SpellPriority     shared.SpellPriority
	EntityModifiers   core.EntityModifiers
	EventListener     func(event interface{})
}

type Equipment struct {
	Armor     armor.Armor   `json:"armor"`
	Primary   weapon.Weapon `json:"primary"`
	Secondary weapon.Weapon `json:"secondary"`
	Ranged    weapon.Weapon `json:"ranged"`
}

type WeaponProficiencies struct {
	Primary   bool
	Secondary bool
	Ranged    bool
}

func New(ctx context.Context, name string, classID int, level int, abilityScores shared.AbilityScores, hp shared.PlayerHP, ap shared.ActionPreference, sp shared.SpellPriority, em core.EntityModifiers) (Character, error) {
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
		Name:             name,
		Class:            c,
		Level:            level,
		AbilityScores:    abilityScores,
		HP:               hp,
		Eq:               Equipment{},
		KnownSpells:      []spells.Spell{},
		SpellSlots:       c.Spellcasting.MaxSpellSlots[level],
		ActionPreference: ap,
		SpellPriority:    sp,
		EntityModifiers:  em,
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

func (c *Character) AddCustomWeapon(name string, isVersatile bool, isFinesse bool, numberOfDice int, die int, damageType string, isRanged bool, slot string) error {
	w, err := weapon.New(name, isVersatile, isFinesse, numberOfDice, die, damageType, isRanged)
	if err != nil {
		return err
	}

	err = c.addWeapon(w, slot)
	if err != nil {
		return err
	}

	return nil
}

func (c *Character) SetWeaponProficiencies(p WeaponProficiencies) {
	c.WeaponProficiency.Primary = p.Primary
	c.WeaponProficiency.Secondary = p.Secondary
	c.WeaponProficiency.Ranged = p.Ranged
}

func (c *Character) SetEntityModifiers(em core.EntityModifiers) {
	c.EntityModifiers = em
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

func (c *Character) IsUnconscious() bool {
	if c.HP.HP <= 0 {
		if c.EventListener != nil {
			event := &events.UnconsciousEvent{}
			event.SetActor(c.Name)
			c.EventListener(event)
		}
		return true
	}
	return false
}

func (c *Character) GetSpellBonus() (int, error) {
	var mod int
	var err error
	switch c.Class.SpellcastingMod {
	case "int":
		mod, err = shared.GetAbilityScoreModifier(c.AbilityScores.Intelligence)
	case "wis":
		mod, err = shared.GetAbilityScoreModifier(c.AbilityScores.Wisdom)
	case "cha":
		mod, err = shared.GetAbilityScoreModifier(c.AbilityScores.Charisma)
	case "":
		mod = 0
	default:
		err = fmt.Errorf("invalid spellcasting modifier: %s", c.Class.SpellcastingMod)
	}

	var pb int
	pb, err = shared.GetCharacterProficiencyBonus(c.Level)
	if err != nil {
		return 0, err
	}

	return mod + pb, nil
}

func (c *Character) PrefersSpells() bool {
	if c.Class.ID == 11 || c.Class.ID == 12 || c.Class.ID == 13 {
		return true
	}
	return false
}

func (c *Character) PrefersRanged() bool {
	if c.Class.ID == 9 {
		return true
	}
	return false
}

func (c *Character) GetSpellSlots() shared.SpellSlots {
	return c.SpellSlots
}

func (c *Character) GetName() string {
	return c.Name
}

func (c *Character) GetCurrentHP() int {
	return c.HP.HP
}

func (c *Character) GetMaxHP() int {
	return c.HP.MaxHP
}

func (c *Character) GetCurrentHPPct() int {
	hpPct := int(float64(c.HP.HP) / float64(c.HP.MaxHP) * 100)
	return hpPct
}

func (c *Character) GetAC() int {
	return c.Eq.Armor.ArmorClass
}

func (c *Character) GetWeaponProficiencyFromSlot(slot shared.WeaponSlot) (bool, error) {
	switch slot {
	case shared.WSPrimary:
		return c.WeaponProficiency.Primary, nil
	case shared.WSSecondary:
		return c.WeaponProficiency.Secondary, nil
	case shared.WSRanged:
		return c.WeaponProficiency.Ranged, nil
	default:
		return false, fmt.Errorf("invalid slot identifier provided: %s", slot)
	}
}

func (c *Character) CreateWeaponAttackInfo(slot shared.WeaponSlot) (combat.AttackInfo, error) {
	w, err := c.getWeaponFromSlot(slot)
	if err != nil {
		return combat.AttackInfo{}, err
	}

	prof, err := c.GetWeaponProficiencyFromSlot(slot)
	if err != nil {
		return combat.AttackInfo{}, err
	}

	attackMod, err := w.GetAttackModifier(&c.AbilityScores, c.Level, prof)
	if err != nil {
		return combat.AttackInfo{}, err
	}

	damageMod, err := w.GetWeaponModifier(&c.AbilityScores)
	if err != nil {
		return combat.AttackInfo{}, err
	}

	return combat.AttackInfo{
		Name:           w.Name,
		NumberOfDice:   w.NumberOfDice,
		Die:            w.Die,
		AttackModifier: attackMod,
		DamageModifier: damageMod,
		DamageType:     w.DamageType,
	}, nil
}

func (c *Character) GetEventListener() func(event interface{}) {
	return c.EventListener
}

var _ core.Entity = &Character{}
