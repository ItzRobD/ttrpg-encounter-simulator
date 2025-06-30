package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/armor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/martial_attacks"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
)

type Character struct {
	Name                string
	Class               Class
	Level               int
	AbilityScores       core.AbilityScores
	HP                  shared.PlayerHP
	Eq                  Equipment
	WeaponProficiency   WeaponProficiencies
	SpellcastingManager *spellcasting_manager.SpellcastingManager
	ActionPreference    shared.ActionPreference
	MeleePreference     shared.MeleePreference
	SpellPriority       shared.SpellPriority
	EntityModifiers     core.EntityModifiers
	Feats               CharacterFeats
	RogueFeatures       RogueFeatures
	NumberOfAttacks     int
	EventListener       func(event interface{})
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

type RogueFeatures struct {
	HasSneakAttack       bool
	HasAssassinate       bool
	NumOfSneakAttackDice int
}

type CharacterFeats struct {
	TwoWeaponFighting bool // Add damage to offhand
	GreatWeaponMaster bool // Power Attack
	Sharpshooter      bool // Power Attack
	XBowExpert        bool // Bonus action hand crossbow attack
	ShieldMaster      bool // Better dex saves with a shield
	WarCaster         bool // Adv on conc saves
	DualWielder       bool // +1 AC While dual wielding, can use non light weapons
	// Crusher/Slasher/Piercer
	//Defensive duelist - reaction to add ac
	HeavyArmorMaster bool // If heavy armor, non magic phys damage reduced by 3
}

func New(ctx context.Context, name string, classID int, level int, abilityScores core.AbilityScores, hp shared.PlayerHP, ap shared.ActionPreference, sp shared.SpellPriority, em core.EntityModifiers) (Character, error) {
	if classID < 0 || classID > 13 {
		return Character{}, fmt.Errorf("invalid class id during character initialization: %d", classID)
	}
	if level < 0 || level > 20 {
		return Character{}, fmt.Errorf("invalid level during character initialization, must be in range 1-20: %d", level)
	}
	if name == "" {
		name = "Unnamed Character"
	}

	var params ClassQueryParams
	params.ID = classID
	c, err := QueryClassData(ctx, params)
	if err != nil {
		return Character{}, err
	}

	numAttacks := 1
	numAttacks, err = GetNumberOfAttacksFromLevelAndClass(ctx, level, classID)
	if err != nil {
		fmt.Println(fmt.Errorf("error getting extra attacks from level and class: %w", err))
	}

	var rFeatures RogueFeatures
	if c.ID == 10 {
		sneakAttackDice, errSA := GetNumberOfSneakAttackDiceFromLevel(ctx, level)
		if errSA != nil {
			sneakAttackDice = 0
			fmt.Println(fmt.Errorf("error getting sneak attack dice from level: %w", errSA))
		}

		rFeatures.HasSneakAttack = true
		rFeatures.NumOfSneakAttackDice = sneakAttackDice
	}

	char := Character{
		Name:          name,
		Class:         c,
		Level:         level,
		AbilityScores: abilityScores,
		HP:            hp,
		Eq:            Equipment{},
		//KnownSpells:   []spells.Spell{},
		//SpellSlots:       c.Spellcasting.MaxSpellSlots[level],
		ActionPreference:    ap,
		SpellPriority:       sp,
		EntityModifiers:     em,
		NumberOfAttacks:     numAttacks,
		Feats:               CharacterFeats{},
		RogueFeatures:       rFeatures,
		SpellcastingManager: &spellcasting_manager.SpellcastingManager{},
	}

	sm, err := initializeSpellcastingManager(ctx, &char)
	if err != nil {
		return Character{}, err
	}

	char.SpellcastingManager = sm

	return char, nil
}

func initializeSpellcastingManager(ctx context.Context, c *Character) (*spellcasting_manager.SpellcastingManager, error) {
	slots, err := GetSpellSlotsByLevelAndClassID(ctx, c.Level, c.Class.ID)
	if err != nil {
		return nil, err
	}

	spellModValue, err := c.GetSpellBonus()

	sm := spellcasting_manager.NewSpellcastingManager(c, spellcasting_manager.CasterCharacter, c.Level, slots, slots, false, spellModValue)
	if err != nil {
		return nil, err
	}

	availableSpellIDs, err := spells.GetUsableSpellIDsByClassID(ctx, c.Class.ID)
	if err != nil {
		return nil, err
	}

	sm.SetUsableSpellIDs(availableSpellIDs)

	return sm, nil
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

func (c *Character) GetSpellBonus(addProficiency bool) (int, error) {
	var mod int
	var err error
	switch c.Class.SpellcastingMod {
	case core.AbilityStrength:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Strength)
	case core.AbilityDexterity:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Dexterity)
	case core.AbilityConstitution:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Constitution)
	case core.AbilityIntelligence:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Intelligence)
	case core.AbilityWisdom:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Wisdom)
	case core.AbilityCharisma:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Charisma)
	default:
		mod = 0
		err = fmt.Errorf("invalid spellcasting modifier: %s", c.Class.SpellcastingMod)
	}

	if addProficiency {
		var pb int
		pb, err = core.GetCharacterProficiencyBonus(c.Level)
		if err != nil {
			return 0, err
		}

		return mod + pb, nil
	}

	return mod, nil
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

//func (c *Character) GetSpellSlots() shared.SpellSlots {
//	return c.SpellSlots
//}

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

func (c *Character) GetLevel() interface{} { return c.Level }

func (c *Character) GetCasterLevel() int { return c.Level }

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

func (c *Character) CreateWeaponAttackData(slot shared.WeaponSlot, useVersatile bool) (martial_attacks.AttackData, error) {
	w, err := c.getWeaponFromSlot(slot)
	if err != nil {
		return martial_attacks.AttackData{}, err
	}

	prof, err := c.GetWeaponProficiencyFromSlot(slot)
	if err != nil {
		return martial_attacks.AttackData{}, err
	}

	attackMod, err := w.GetAttackModifier(&c.AbilityScores, c.Level, prof)
	if err != nil {
		return martial_attacks.AttackData{}, err
	}

	damageMod, err := w.GetWeaponModifier(&c.AbilityScores)
	if err != nil {
		return martial_attacks.AttackData{}, err
	}

	die := w.Die
	var v bool
	if useVersatile && w.IsVersatile {
		die = w.Die + 2
		v = true
	}

	return martial_attacks.AttackData{
		Name:              w.Name,
		NumberOfDice:      w.NumberOfDice,
		Die:               die,
		AttackModifier:    attackMod,
		DamageModifier:    damageMod,
		DamageType:        w.DamageType,
		IsVersatileAttack: v,
	}, nil
}

func (c *Character) CreateAttackRequest(slot shared.WeaponSlot, useVersatile bool, advantage core.AdvantageType) (*martial_attacks.AttackRequest, error) {
	attackData, err := c.CreateWeaponAttackData(slot, useVersatile)
	if err != nil {
		return nil, err
	}

	// TODO: This will have to be handled internally by other functions to get the values of each of these
	//		Will have to account for character feats
	modifiers := martial_attacks.AttackModifiers{
		BonusAttackRoll:      0,
		BonusDamageRoll:      0,
		ShouldApplyDamageMod: false,
		PowerAttack:          false,
		ImprovedCritical:     false,
		RerollOnesAndTwos:    false,
		HalflingLucky:        false,
	}

	return &martial_attacks.AttackRequest{
		AttackData:  attackData,
		Modifiers:   modifiers,
		Advantage:   advantage,
		AttackCount: c.NumberOfAttacks,
	}, nil
}

func (c *Character) CreateSpellAttackData(spellChoice spellcasting_manager.SpellChoice) (spellcasting_manager.SpellAttackData, error) {
	spellBonus, err := c.GetSpellBonus(true)
	if err != nil {
		return nil, err
	}

	spellMod, err := c.GetSpellBonus(false)
	if err != nil {
		return nil, err
	}

	return spellcasting_manager.SpellAttackData{
		SpellChoice:    spellChoice,
		AttackModifier: spellBonus,
		SpellModifier:  spellMod,
	}, nil
}

func (c *Character) CreateSpellCastRequest(spellChoice spellcasting_manager.SpellChoice, advantage core.AdvantageType) (*spellcasting_manager.SpellcastRequest, error) {
	attackData, err := c.CreateSpellAttackData(spellChoice)
	if err != nil {
		return nil, err
	}

	modifiers := spellcasting_manager.SpellModifiers{
		TreatOnesAsTwos: false,
		HalflingLucky:   false,
	}

	return &spellcasting_manager.SpellCastRequest{
		AttackData: attackData,
		Modifiers:  modifiers,
		Advantage:  advantage,
	}, nil
}

func (c *Character) GetEventListener() func(event interface{}) {
	return c.EventListener
}

var _ core.Entity = &Character{}
