package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/armor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/martial_attack_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
)

// Character represents a player character in the game including its stats, abilities, equipment, and features.
// Name:  name of the character
// Class
// Level: character level
// AbilityScores: Struct containing the 6 ability scores
// HP: PlayerHP struct containing HP, MaxHP, Hit Die, and rolls
// Eq: Struct containing equippeed weapons and armor
// WeaponProficiency: Struct containing whether character is proficient in each weapon slot
// SpellcastingManager:
// MartialAttackManager:
// ActionPreference: represents character preference to perform melee/ranged/spell actions
// MeleePreference: represents character preference to use versatile weapon, not, no proference
// SpellPriority: represents character preference on spell type set by simulator
// EntityModifiers: Advantage/bonus on rolls, versatile attack and upcast options
// Feats: character assigned feats
// RogueFeatures: sneak attack and assassinate data
// NumberOfAttacks: represents the number of attacks a character can make (extra attack)
// EventListener: Registered for action logs
type Character struct {
	Name                 string
	Class                Class
	Level                int
	AbilityScores        core.AbilityScores
	AbilityScoreProf     core.AbilityScoresProficiencies
	HP                   shared.PlayerHP
	Eq                   Equipment
	WeaponProficiency    WeaponProficiencies
	SpellcastingManager  *spellcasting_manager.SpellcastingManager
	MartialAttackManager *martial_attack_manager.MartialAttackManager
	RollManager          *roll_manager.RollManager
	ActionPreference     core.ActionPreference
	MeleePreference      core.VersatileWeaponPreference
	SpellPriority        core.SpellPriority
	EntityModifiers      core.EntityModifiers
	Feats                CharacterFeats
	RerollAbilities      roll_manager.RerollAbilities
	RogueFeatures        RogueFeatures
	NumberOfAttacks      int
	EventListener        func(event interface{})
}

// TODO: New Managers/ideas
// 		C/M |Combat State Manager: HP Management, Action Management, Leg Actions, Conditions
//		C | Equipment
//		C | Proficiencies:

// Equipment represents a character's equipped armor and weapons, including primary, secondary, and ranged weapon slots.
type Equipment struct {
	Armor     armor.Armor   `json:"armor"`
	Primary   weapon.Weapon `json:"primary"`
	Secondary weapon.Weapon `json:"secondary"`
	Ranged    weapon.Weapon `json:"ranged"`
}

// WeaponProficiencies represents a character's proficiency in using primary, secondary, and ranged weapons.
type WeaponProficiencies struct {
	Primary   bool
	Secondary bool
	Ranged    bool
}

// RogueFeatures defines features specific to a rogue character, including sneak attack and assassinate capabilities.
type RogueFeatures struct {
	HasSneakAttack       bool
	HasAssassinate       bool
	NumOfSneakAttackDice int
}

// CharacterFeats represents the set of optional feats that a character may possess in the game.
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

func New(ctx context.Context, name string, classID int, level int, abilityScores core.AbilityScores, hp shared.PlayerHP, ap core.ActionPreference, sp core.SpellPriority, em core.EntityModifiers) (Character, error) {
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

	sm, err := initializeSpellcastingManager(ctx, &char, em.CanUpcast)
	if err != nil {
		return Character{}, err
	}

	char.SpellcastingManager = sm

	return char, nil
}

// initializeSpellcastingManager initializes and configures a SpellcastingManager for the provided character.
// It retrieves the spell slots and usable spell IDs based on the character's level and class.
// Returns a pointer to the SpellcastingManager and an error if any issue occurs during initialization.
func initializeSpellcastingManager(ctx context.Context, c *Character, canUpcast bool) (*spellcasting_manager.SpellcastingManager, error) {
	slots, err := GetSpellSlotsByLevelAndClassID(ctx, c.Level, c.Class.ID)
	if err != nil {
		return nil, err
	}

	spellModValue, err := c.GetSpellBonus(true)

	sm := spellcasting_manager.NewSpellcastingManager(c, core.CasterCharacter, c.Level, slots, slots, canUpcast, spellModValue)
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

// AddSRDArmor adds a piece of armor identified by armorID to the character's equipment.
// armorID must be a valid identifier within the allowed range (0 to 13).
// Returns an error if the armor ID is invalid or if fetching armor data fails.
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

// AddCustomArmor assigns a new custom armor to the character using the specified parameters and validates the input.
// Returns an error if the armor creation fails.
func (c *Character) AddCustomArmor(name string, armorClass int, dexBonus bool, maxBonus bool, minimumStr int) error {
	a, err := armor.New(name, armorClass, dexBonus, maxBonus, minimumStr)
	if err != nil {
		return err
	}

	c.Eq.Armor = a
	return nil
}

// AddSRDWeapon adds a weapon to the character's inventory using the provided weapon ID and slot.
// Returns an error if the weapon ID is invalid, slot is invalid, or data query fails.
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

// AddCustomWeapon adds a custom weapon to the character with specified properties and associates it with a specified slot.
// Returns an error if weapon creation or adding it to the character fails.
func (c *Character) AddCustomWeapon(name string, isVersatile bool, isFinesse bool, numberOfDice int, die core.DiceType, damageType core.DamageType, isRanged bool, slot string) error {
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

// SetWeaponProficiencies sets the character's primary, secondary, and ranged weapon proficiencies.
func (c *Character) SetWeaponProficiencies(p WeaponProficiencies) {
	c.WeaponProficiency.Primary = p.Primary
	c.WeaponProficiency.Secondary = p.Secondary
	c.WeaponProficiency.Ranged = p.Ranged
}

// SetEntityModifiers sets the EntityModifiers property, allowing customization of initiative, attacks, and casting options.
func (c *Character) SetEntityModifiers(em core.EntityModifiers) {
	c.EntityModifiers = em
}

// addWeapon assigns a weapon to a specified equipment slot of the character and returns an error if the operation fails.
// The slot must be one of "primary", "secondary", or "ranged". Non-ranged weapons cannot be assigned to the "ranged" slot.
// Returns an error if the slot is invalid or incompatible with the weapon type.
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

// IsUnconscious checks if the character's current HP is zero or below, triggering an unconscious event if applicable.
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

// GetSpellBonus calculates the spellcasting bonus based on the character's ability score and optionally adds proficiency bonus.
// Returns the calculated bonus or an error if the spellcasting modifier or level is invalid.
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

// PrefersSpells determines if the character's class prefers spells based on its class ID. Returns true if the class prefers spells.
func (c *Character) PrefersSpells() bool {
	if c.Class.ID == 11 || c.Class.ID == 12 || c.Class.ID == 13 {
		return true
	}
	return false
}

// PrefersRanged determines if the character prefers ranged attacks based on its class with ID 9 returning true.
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

func (c *Character) GetAbilityScores() core.AbilityScores {
	return c.AbilityScores
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

// GetSpellSaveDC calculates the spell save DC for the character based on the given ability modifier and a base value of 8.
func (c *Character) GetSpellSaveDC(ability core.Ability) int {
	var abilityMod int
	var err error
	switch ability {
	case core.AbilityStrength:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Strength)
	case core.AbilityDexterity:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Dexterity)
	case core.AbilityConstitution:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Constitution)
	case core.AbilityIntelligence:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Intelligence)
	case core.AbilityWisdom:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Wisdom)
	case core.AbilityCharisma:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Charisma)
	default:
		abilityMod = 0
		err = fmt.Errorf("invalid ability provided: %s", ability)
	}
	if err != nil {
		return 0
	}
	return 8 + abilityMod
}

// GetWeaponProficiencyFromSlot returns the weapon proficiency for the given weapon slot or an error if the slot is invalid.
func (c *Character) GetWeaponProficiencyFromSlot(slot core.WeaponSlot) (bool, error) {
	switch slot {
	case core.WSPrimary:
		return c.WeaponProficiency.Primary, nil
	case core.WSSecondary:
		return c.WeaponProficiency.Secondary, nil
	case core.WSRanged:
		return c.WeaponProficiency.Ranged, nil
	default:
		return false, fmt.Errorf("invalid slot identifier provided: %s", slot)
	}
}

// CreateWeaponAttackData generates an AttackData object for a given weapon slot, considering proficiency and versatility.
// slot specifies the weapon slot to retrieve the weapon from.
// useVersatile indicates whether to use the weapon in versatile mode, if applicable.
// Returns the constructed AttackData and an error if any issue occurs in retrieving or calculating weapon properties.
func (c *Character) CreateWeaponAttackData(slot core.WeaponSlot, useVersatile bool) (martial_attack_manager.AttackData, error) {
	w, err := c.getWeaponFromSlot(slot)
	if err != nil {
		return martial_attack_manager.AttackData{}, err
	}

	prof, err := c.GetWeaponProficiencyFromSlot(slot)
	if err != nil {
		return martial_attack_manager.AttackData{}, err
	}

	attackMod, err := w.GetAttackModifier(&c.AbilityScores, c.Level, prof)
	if err != nil {
		return martial_attack_manager.AttackData{}, err
	}

	damageMod, err := w.GetWeaponModifier(&c.AbilityScores)
	if err != nil {
		return martial_attack_manager.AttackData{}, err
	}

	die := w.Die
	var v bool
	if useVersatile && w.IsVersatile {
		die = w.Die + 2
		v = true
	}

	return martial_attack_manager.AttackData{
		Name:              w.Name,
		NumberOfDice:      w.NumberOfDice,
		Die:               die,
		AttackModifier:    attackMod,
		DamageModifier:    damageMod,
		DamageType:        w.DamageType,
		IsVersatileAttack: v,
	}, nil
}

// CreateAttackRequest generates an attack request with specific weapon data, modifiers, advantage type, and attack count.
func (c *Character) CreateAttackRequest(slot core.WeaponSlot, useVersatile bool, advantage core.AdvantageType, simulationOptions core.SimulationOptions) (*martial_attack_manager.AttackRequest, error) {
	attackData, err := c.CreateWeaponAttackData(slot, useVersatile)
	if err != nil {
		return nil, err
	}

	// TODO: This will have to be handled internally by other functions to get the values of each of these
	//		Will have to account for character feats
	attackOptions := martial_attack_manager.AttackOptions{
		BonusToAttackRoll:    0,
		BonusToDamageRoll:    0,
		ShouldApplyDamageMod: false,
		PowerAttack:          false,
		ImprovedCritical:     false,
		RerollOnesAndTwos:    false,
	}

	return &martial_attack_manager.AttackRequest{
		AttackData:        attackData,
		AttackOptions:     attackOptions,
		SimulationOptions: simulationOptions,
		// TODO: I removed target here. make sure this is being set by simulation
	}, nil
}

// CreateSpellAttackData creates and returns the data for a spell attack, including attack and spell modifiers.
// It takes a SpellChoice as input and computes the necessary modifiers for the attack.
// Returns a SpellCastData struct and an error if any calculation fails.
func (c *Character) CreateSpellAttackData(spellChoice spells.SpellChoice) (spellcasting_manager.SpellCastData, error) {
	spellBonus, err := c.GetSpellBonus(true)
	if err != nil {
		return spellcasting_manager.SpellCastData{}, err
	}

	spellMod, err := c.GetSpellBonus(false)
	if err != nil {
		return spellcasting_manager.SpellCastData{}, err
	}

	return spellcasting_manager.SpellCastData{
		SpellChoice:          spellChoice,
		AttackModifier:       spellBonus,
		SpellcastingModifier: spellMod,
	}, nil
}

// CreateSpellCastRequest generates a new SpellCastRequest based on the given spell choice and advantage type.
func (c *Character) CreateSpellCastRequest(spellChoice spells.SpellChoice, advantage core.AdvantageType) (*spellcasting_manager.SpellCastRequest, error) {
	attackData, err := c.CreateSpellAttackData(spellChoice)
	if err != nil {
		return nil, err
	}

	modifiers := spellcasting_manager.SpellOptions{
		TreatOnesAsTwos: false,
		HalflingLucky:   false,
	}
	// TODO: Since the rework of spellcasting, roll, and attack managers
	//		Need to now rework the creation of attack and spellcast requests for
	//		both mnster and character
	//		Afterwards look at the logic of turns, keep in mind logging has also changed
	//		for some things -> universal handler and also use of requests and results
	//
	// TODO: There may also be additional errors to clean up from these changes -> types etc
	//			Also ensure we are validating type creations

	return &spellcasting_manager.SpellCastRequest{
		AttackData: attackData,
		Modifiers:  modifiers,
		Advantage:  advantage,
	}, nil
}

// MakeSavingThrow calculates a saving throw roll using the specified ability and returns the result, rolls, and an error if any.
func (c *Character) MakeSavingThrow(ability core.Ability, targetValue int) (core.RollResult, error) {
	mod, err := c.GetSavingThrowBonus(ability)
	if err != nil {
		return nil, err
	}

	opts := roll_manager.RollOptions{
		Advantage:         core.RollNormal, // TODO: Determining advantage needs to be handled ie racial traits
		Modifier:          mod,
		CriticalThreshold: 0,     // Not relevant
		TreatOnesAsTwos:   false, // Not relevant
		RollType:          core.DiceRollSavingThrow,
		RollContext:       "Saving Throw",
		TargetValue:       targetValue,
	}

	res, err := c.RollManager.RollSavingThrow(ability, opts)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetAbilityScore returns the score for the specified ability of the character. Defaults to 0 if the ability is not found.
func (c *Character) GetAbilityScore(ability core.Ability) int {
	switch ability {
	case core.AbilityStrength:
		return c.AbilityScores.Strength
	case core.AbilityDexterity:
		return c.AbilityScores.Dexterity
	case core.AbilityConstitution:
		return c.AbilityScores.Constitution
	case core.AbilityIntelligence:
		return c.AbilityScores.Intelligence
	case core.AbilityWisdom:
		return c.AbilityScores.Wisdom
	case core.AbilityCharisma:
		return c.AbilityScores.Charisma
	default:
		return 0
	}
}

// GetAbilityScoreModifier calculates the ability score modifier for a given ability based on the character's ability scores.
// Returns the modifier as an integer or an error if the ability is invalid.
func (c *Character) GetAbilityScoreModifier(ability core.Ability) (int, error) {
	var abilityMod int
	var err error
	switch ability {
	case core.AbilityStrength:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Strength)
	case core.AbilityDexterity:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Dexterity)
	case core.AbilityConstitution:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Constitution)
	case core.AbilityIntelligence:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Intelligence)
	case core.AbilityWisdom:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Wisdom)
	case core.AbilityCharisma:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Charisma)
	default:
		abilityMod = 0
		err = fmt.Errorf("invalid ability provided: %s", ability)
	}
	if err != nil {
		return 0, err
	}
	return abilityMod, nil
}

// GetIsProficientInAbility checks if the character is proficient in the specified ability and returns true if proficient.
func (c *Character) GetIsProficientInAbility(ability core.Ability) bool {
	switch ability {
	case core.AbilityStrength:
		return c.AbilityScoreProf.Strength
	case core.AbilityDexterity:
		return c.AbilityScoreProf.Dexterity
	case core.AbilityConstitution:
		return c.AbilityScoreProf.Constitution
	case core.AbilityIntelligence:
		return c.AbilityScoreProf.Intelligence
	case core.AbilityWisdom:
		return c.AbilityScoreProf.Wisdom
	case core.AbilityCharisma:
		return c.AbilityScoreProf.Charisma
	default:
		return false
	}
}

// GetSavingThrowBonus calculates the saving throw bonus from ability modifiers and proficiency based on character level.
func (c *Character) GetSavingThrowBonus(ability core.Ability) (int, error) {
	var pb int
	var mod int
	var err error

	mod, err = c.GetAbilityScoreModifier(ability)
	pb, err = core.GetCharacterProficiencyBonus(c.Level)
	if err != nil {
		return 0, err
	}

	if c.GetIsProficientInAbility(ability) {
		return pb + mod, nil
	}
	return mod, nil
}

func (c *Character) IsCharacter() bool { return true }
func (c *Character) IsMonster() bool   { return false }

func (c *Character) GetEventListener() func(event interface{}) {
	return c.EventListener
}

var _ core.Entity = &Character{}
