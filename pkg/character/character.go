package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/armor"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/martial_attack_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/entity_configuration"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
)

// Character represents a player or NPC with attributes like name, class, level, and various managers for gameplay systems.
type Character struct {
	Name                 string
	Class                classes.Class
	Level                uint8
	AbilityScores        core.AbilityScores
	AbilityScoreProf     core.AbilityScoresProficiencies
	EntityState          *entity_state_manager.EntityStateManager
	EquipmentManager     *equipment_manager.EquipmentManager
	SpellCastingManager  *spellcasting_manager.SpellcastingManager
	MartialAttackManager *martial_attack_manager.MartialAttackManager
	RollManager          *roll_manager.RollManager
	AI                   *CharacterAI
	Configuration        entity_configuration.EntityConfiguration
	HPConfig             core.HPConfig
	Seed                 core.Seed
	RNG                  *rand.Rand
	EventListener        func(event interface{})
}

type CharacterConfig struct {
	Name      string
	ClassID   classes.ClassID
	Level     uint8
	AsConfig  core.AbilityScoresConfig
	HPMethod  core.HPSetMethod
	HPValue   int
	Seed      core.Seed
	Equipment EquipmentConfig
}

type EquipmentConfig struct {
	ArmorID       int
	PrimarySlot   map[int]bool
	SecondarySlot map[int]bool
	RangedSlot    map[int]bool
}

// NewCharacter initializes and returns a new Character with the specified parameters or an error if validation fails.
func NewCharacter(ctx context.Context, charConfig CharacterConfig) (*Character, error) {
	if charConfig.ClassID < 0 || charConfig.ClassID > 13 {
		return nil, fmt.Errorf("invalid classData id during character initialization: %d", charConfig.ClassID)
	}
	if charConfig.Level < 0 || charConfig.Level > 20 {
		return nil, fmt.Errorf("invalid level during character initialization, must be in range 1-20: %d", charConfig.Level)
	}
	if charConfig.Name == "" {
		charConfig.Name = "Unnamed Character"
	}

	var params classes.ClassQueryParams
	params.ID = charConfig.ClassID
	params.Level = charConfig.Level
	classData, err := classes.QueryClassData(ctx, params)
	if err != nil {
		return nil, err
	}

	var seed core.Seed
	if charConfig.Seed.Seed1 == 0 {
		seed.Seed1 = rand.Uint64()
	}
	if charConfig.Seed.Seed2 == 0 {
		seed.Seed2 = rand.Uint64()
	}

	// Set initial values for character
	char := Character{
		Name:                 charConfig.Name,
		Class:                classData,
		Level:                charConfig.Level,
		AbilityScores:        charConfig.AsConfig.AbilityScores,
		AbilityScoreProf:     charConfig.AsConfig.Proficiencies,
		EntityState:          &entity_state_manager.EntityStateManager{},
		EquipmentManager:     &equipment_manager.EquipmentManager{},
		SpellCastingManager:  &spellcasting_manager.SpellcastingManager{},
		MartialAttackManager: &martial_attack_manager.MartialAttackManager{},
		RollManager:          &roll_manager.RollManager{},
		AI:                   &CharacterAI{},
		Configuration:        entity_configuration.EntityConfiguration{}, // TODO: this isn't being set up
		HPConfig:             core.HPConfig{HPSetMethod: charConfig.HPMethod, Value: charConfig.HPValue},
		Seed:                 seed,
		RNG:                  rand.New(rand.NewPCG(seed.Seed1, seed.Seed2)),
	}

	// Initialize managers
	// Roll Manager
	char.RollManager = initializeRollManager(&char, &char.Configuration)

	// AI
	char.AI = NewCharacterAI(&char)
	// Entity State Manager
	// TODO: Implement resistances for characters
	esmConfig := entity_state_manager.EntityStateConfig{
		AttackCount: classData.AttackCount,
		Conditions:  core.NewEntityConditions(),
	}
	esm, err := initializeEntityStateManager(&char, &esmConfig)
	if err != nil {
		return nil, err
	}
	char.EntityState = esm

	// Equipment Manager
	char.EquipmentManager, err = equipment_manager.NewEquipmentManager(&char)
	if err != nil {
		return nil, err
	}

	// Add equipment from config
	err = char.setupEquipmentFromConfig(ctx, charConfig.Equipment)
	if err != nil {
		return nil, err
	}

	// Spellcasting Manager
	char.SpellCastingManager, err = initializeSpellcastingManager(ctx, &char)
	if err != nil {
		fmt.Printf("Error initializing spellcasting manager: %v\n", err)
		return nil, err
	}

	// Martial Attack Manager
	char.MartialAttackManager = martial_attack_manager.NewMartialAttackManager(&char, char.RollManager)

	// Set up HP SimOptions and character hp
	char.HPConfig.HitDie = char.GetHitDie()
	char.HPConfig.NumberOfDice = int(char.Level - 1)
	modifier, _ := char.getAbilityScoreModifier(core.AbilityConstitution)
	char.HPConfig.HPAverage = int(math.Round(float64(char.GetHitDie().Int())+float64(char.HPConfig.NumberOfDice)*char.Class.HitDie.Avg()) + float64(int(char.Level)*modifier))
	char.HPConfig.Modifier = modifier

	// Moving HP Setup to during simulation
	//err = char.setHP(char.HPConfig)
	//if err != nil {
	//	return nil, err
	//}

	return &char, nil
}

func initializeRollManager(c *Character, eConfig *entity_configuration.EntityConfiguration) *roll_manager.RollManager {
	rm := roll_manager.NewRollManager(c, eConfig.CombatFeatures.ReRollAbilities)
	return rm
}

func initializeEntityStateManager(c *Character, config *entity_state_manager.EntityStateConfig) (*entity_state_manager.EntityStateManager, error) {
	esm, err := entity_state_manager.NewEntityStateManager(c, *config)
	if err != nil {
		return nil, err
	}
	return esm, nil
}

// initializeSpellcastingManager initializes and returns a SpellCastingManager for the given character and context.
// It configures spell slots, casting options, usable spells, and other spell-related properties.
// Returns an error if data retrieval or initialization fails.
func initializeSpellcastingManager(ctx context.Context, c *Character) (*spellcasting_manager.SpellcastingManager, error) {
	if c.Class.ID == classes.Barbarian {
		return &spellcasting_manager.SpellcastingManager{}, nil
	}
	slots, err := classes.GetSpellSlotsByLevelAndClassID(ctx, c.Level, c.Class.ID.Int())
	if err != nil {
		return nil, err
	}

	spellModValue, err := c.GetSpellBonus(true)

	// TODO: Sim options can be passed via context for simplification
	canUpcast := ctx.Value("CanUpcast").(bool)
	sm := spellcasting_manager.NewSpellcastingManager(c, c.RollManager, core.CasterCharacter, int(c.Level), slots, slots, canUpcast, spellModValue)
	if err != nil {
		return nil, err
	}

	availableSpellIDs, err := spells.GetUsableSpellIDsByClassID(ctx, c.Class.ID.Int())
	if err != nil {
		return nil, err
	}

	if len(availableSpellIDs) > 0 {
		spellMap, err := spells.QuerySpellData(ctx, spells.SpellQueryParams{ID: availableSpellIDs})
		if err != nil {
			return nil, err
		}

		err = sm.AddKnownSpellsFromMap(spellMap)
		if err != nil {
			return nil, err
		}
	}

	return sm, nil
}

// setHP sets the character's hit points (HP) using the specified method and value.
// Arguments:
// - m: The HP set method, which determines whether to set HP by value, average, or rolling.
// - value: The value used to set HP when the method is HPSetValue.
// Returns an error if an invalid method is provided or internal operations fail.
func (c *Character) setHP(config core.HPConfig) error {
	switch config.HPSetMethod {
	case core.HPSetValue:
		hp := entity_state_manager.HPValues{
			CurrentHP: config.Value,
			MaxHP:     config.Value,
			TempHP:    0,
			HitDie:    config.HitDie,
		}
		hpRoll := roll_manager.RollResult{
			DiceRollType:   core.DiceRollHPValueUsed,
			FinalRollValue: config.Value,
			Total:          config.Value,
		}
		events.LogDiceRollEvent(c, &hpRoll, c.EventListener)
		c.EntityState.SetHPValues(hp)

		return nil
	case core.HPSetAverage:
		hp := entity_state_manager.HPValues{
			CurrentHP: config.HPAverage,
			MaxHP:     config.HPAverage,
			TempHP:    0,
			HitDie:    config.HitDie,
		}
		hpRoll := roll_manager.RollResult{
			DiceRollType:   core.DiceRollHPAvgUsed,
			NumberOfDice:   config.NumberOfDice,
			Die:            config.HitDie,
			FinalRollValue: config.HPAverage,
			Modifier:       config.Modifier,
			Total:          config.HPAverage,
		}
		events.LogDiceRollEvent(c, &hpRoll, c.EventListener)
		c.EntityState.SetHPValues(hp)

		return nil
	case core.HPSetRoll:
		hpRoll, err := c.RollManager.RollHP(config)
		if err != nil {
			return err
		}
		hp := entity_state_manager.HPValues{
			CurrentHP: hpRoll.Total,
			MaxHP:     hpRoll.Total,
			TempHP:    0,
			HitDie:    hpRoll.Die,
		}
		c.EntityState.SetHPValues(hp)
		return nil
	default:
		return fmt.Errorf("invalid HP set method: %v", config.HPSetMethod)
	}
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

// GetSpellSaveDC calculates the spell save DC for the character based on the given ability modifier and a base value of 8.
func (c *Character) GetSpellSaveDC(ability *core.Ability) (int, error) {
	if ability == nil {
		return 8, fmt.Errorf("ability cannot be nil")
	}
	var abilityMod int
	var err error
	switch *ability {
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
	return 8 + abilityMod, nil
}

// CreateWeaponAttackData generates an AttackData object for a given weapon slot, considering proficiency and versatility.
// slot specifies the weapon slot to retrieve the weapon from.
// useVersatile indicates whether to use the weapon in versatile mode, if applicable.
// Returns the constructed AttackData and an error if any issue occurs in retrieving or calculating weapon properties.
func (c *Character) CreateWeaponAttackData(slot core.WeaponSlot, useVersatile bool) (core.AttackData, error) {
	w, err := c.EquipmentManager.GetWeaponFromSlot(slot)
	if err != nil {
		return core.AttackData{}, err
	}

	prof := c.EquipmentManager.GetIsProficientWithSlot(slot)

	attackMod, err := w.GetAttackModifier(&c.AbilityScores, c.Level, prof)
	if err != nil {
		return core.AttackData{}, err
	}

	damageMod, err := w.GetWeaponModifier(&c.AbilityScores)
	if err != nil {
		return core.AttackData{}, err
	}

	die := w.Die
	var v bool
	if useVersatile && w.IsVersatile {
		die = w.Die + 2
		v = true
	}

	return core.AttackData{
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
func (c *Character) CreateAttackRequest(target core.Entity, slot core.WeaponSlot, adv core.AdvantageType, useVersatile bool, simulationOptions *core.SimulationOptions) (*martial_attack_manager.AttackRequest, error) {
	attackData, err := c.EquipmentManager.GetWeaponAttackData(slot, useVersatile)
	if err != nil {
		return nil, err
	}

	// TODO: This will have to be handled internally by other functions to get the values of each of these
	//		Will have to account for character feats
	attackOptions := martial_attack_manager.AttackOptions{
		NumberOfAttacks:      c.EntityState.GetNumberOfAttacks(),
		BonusToAttackRoll:    0,
		BonusToDamageRoll:    0,
		ShouldApplyDamageMod: true,
		PowerAttack:          false,
		ImprovedCritical:     false,
		RerollOnesAndTwos:    false,
		Advantage:            adv,
	}

	return &martial_attack_manager.AttackRequest{
		AttackData:        attackData,
		AttackOptions:     attackOptions,
		SimulationOptions: simulationOptions,
		Target:            target,
	}, nil
}

// CreateSpellAttackData creates and returns the data for a spell attack, including attack and spell modifiers.
// It takes a SpellChoice as input and computes the necessary modifiers for the attack.
// Returns a SpellCastData struct and an error if any calculation fails.
func (c *Character) CreateSpellAttackData(spellChoice core.SpellChoice) (spellcasting_manager.SpellCastData, error) {
	// TODO: We always add proficiency, determine if we need to account for different armor/scrolls - unlikely
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
func (c *Character) CreateSpellCastRequest(spellChoice core.SpellChoice, adv core.AdvantageType, simOptions *core.SimulationOptions) (*spellcasting_manager.SpellCastRequest, error) {
	spellcastData, err := c.CreateSpellAttackData(spellChoice)
	if err != nil {
		return nil, err
	}

	// TODO: Handle the creation of these options dynamically
	options := spellcasting_manager.SpellOptions{
		Advantage:            adv,
		BonusToAttackRoll:    0,
		BonusToDamageRoll:    0,
		ShouldApplyDamageMod: false,
		ImprovedCritical:     false,
		TreatOnesAsTwos:      false,
	}

	return &spellcasting_manager.SpellCastRequest{
		SpellCastData:     spellcastData,
		SpellOptions:      options,
		SimulationOptions: simOptions,
		Target:            nil,
	}, nil
}

// MakeSavingThrow calculates a saving throw roll using the specified ability and returns the result, rolls, and an error if any.
func (c *Character) MakeSavingThrow(ability core.Ability, targetValue int) (core.RollResult, error) {
	mod, err := c.getSavingThrowBonus(ability)
	if err != nil {
		return nil, err
	}

	opts := roll_manager.RollOptions{
		Advantage:         core.RollNormal, // TODO: Determining advantage needs to be handled ie racial traits
		Modifier:          mod,
		CriticalThreshold: 0,     // Not relevant
		TreatOnesAsTwos:   false, // Not relevant
		RollType:          core.DiceRollSavingThrow,
		TargetValue:       targetValue,
	}

	res, err := c.RollManager.RollSavingThrow(ability, opts)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// getAbilityScore returns the score for the specified ability of the character. Defaults to 0 if the ability is not found.
func (c *Character) getAbilityScore(ability core.Ability) int {
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

// getAbilityScoreModifier calculates the ability score modifier for a given ability based on the character's ability scores.
// Returns the modifier as an integer or an error if the ability is invalid.
func (c *Character) getAbilityScoreModifier(ability core.Ability) (int, error) {
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

// getIsProficientInAbility checks if the character is proficient in the specified ability and returns true if proficient.
func (c *Character) getIsProficientInAbility(ability core.Ability) bool {
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

// getSavingThrowBonus calculates the saving throw bonus from ability modifiers and proficiency based on character level.
func (c *Character) getSavingThrowBonus(ability core.Ability) (int, error) {
	var pb int
	var mod int
	var err error

	mod, err = c.getAbilityScoreModifier(ability)
	pb, err = core.GetCharacterProficiencyBonus(c.Level)
	if err != nil {
		return 0, err
	}

	if c.getIsProficientInAbility(ability) {
		return pb + mod, nil
	}
	return mod, nil
}

func (c *Character) setupEquipmentFromConfig(ctx context.Context, config EquipmentConfig) error {
	// Handle armor
	if config.ArmorID > 0 {
		a, err := armor.QueryArmorData(ctx, armor.ArmorQueryParams{ID: config.ArmorID})
		if err != nil {
			return fmt.Errorf("failed to get armor ID %d: %w", config.ArmorID, err)
		}
		c.EquipmentManager.SetArmor(a)
	}

	// Handle primary slot weapons
	for weaponID, isProficient := range config.PrimarySlot {
		w, err := weapon.QueryWeaponData(ctx, weapon.WeaponQueryParams{ID: weaponID})
		if err != nil {
			return fmt.Errorf("failed to get w ID %d for primary slot: %w", weaponID, err)
		}

		err = c.EquipmentManager.SetWeapon(core.WSPrimary, &w, isProficient)
		if err != nil {
			return fmt.Errorf("failed to set primary w: %w", err)
		}
	}

	// Handle secondary slot weapons
	for weaponID, isProficient := range config.SecondarySlot {
		w, err := weapon.QueryWeaponData(ctx, weapon.WeaponQueryParams{ID: weaponID})
		if err != nil {
			return fmt.Errorf("failed to get w ID %d for secondary slot: %w", weaponID, err)
		}

		err = c.EquipmentManager.SetWeapon(core.WSSecondary, &w, isProficient)
		if err != nil {
			return fmt.Errorf("failed to set secondary w: %w", err)
		}
	}

	// Handle ranged slot weapons
	for weaponID, isProficient := range config.RangedSlot {
		w, err := weapon.QueryWeaponData(ctx, weapon.WeaponQueryParams{ID: weaponID})
		if err != nil {
			return fmt.Errorf("failed to get w ID %d for ranged slot: %w", weaponID, err)
		}

		err = c.EquipmentManager.SetWeapon(core.WSRanged, &w, isProficient)
		if err != nil {
			return fmt.Errorf("failed to set ranged w: %w", err)
		}
	}

	return nil

}

func (c *Character) RollInitiative() (int, error) {
	var err error
	opts := roll_manager.NewRollOptions()
	opts.Modifier, err = c.getAbilityScoreModifier(core.AbilityDexterity)
	if err != nil {
		return 0, err
	}
	// TODO: Handle chaaracter feats such as Alert for +5

	res, err := c.RollManager.RollInitiative(opts)
	if err != nil {
		return 0, err
	}

	c.EntityState.SetInitiative(res.Total)

	return res.Total, nil
}

// Interface Required Functions

func (c *Character) IsCharacter() bool                          { return true }
func (c *Character) IsMonster() bool                            { return false }
func (c *Character) GetEventListener() func(event interface{})  { return c.EventListener }
func (c *Character) SetEventListener(f func(event interface{})) { c.EventListener = f }
func (c *Character) IsUnconscious() bool                        { return c.EntityState.GetIsUnconscious() }
func (c *Character) GetHPStatus() core.HPStatus                 { return c.EntityState.GetHPStatus() }
func (c *Character) GetName() string                            { return c.Name }
func (c *Character) GetAbilityScores() core.AbilityScores       { return c.AbilityScores }
func (c *Character) GetLevel() float64                          { return float64(c.Level) }
func (c *Character) GetHPConfig() core.HPConfig                 { return c.HPConfig }
func (c *Character) GetCasterLevel() int                        { return int(c.Level) }
func (c *Character) GetAC() int                                 { return c.EquipmentManager.GetAC() }
func (c *Character) GetAbilityScore(a core.Ability) int         { return c.getAbilityScore(a) }
func (c *Character) GetAbilityScoreModifier(a core.Ability) (int, error) {
	return c.getAbilityScoreModifier(a)
}
func (c *Character) GetSavingThrowBonus(a core.Ability) (int, error) { return c.getSavingThrowBonus(a) }
func (c *Character) GetHitDie() core.DiceType                        { return c.Class.HitDie }
func (c *Character) GetState() interface{}                           { return c.EntityState }
func (c *Character) InitializeHP() error                             { return c.setHP(c.HPConfig) }
func (c *Character) IsSpellcaster() bool                             { return c.SpellCastingManager.HasAnyKnownSpells() }
func (c *Character) IsHealer() bool                                  { return c.SpellCastingManager.HasHealingSpells() }
func (c *Character) GetRNG() *rand.Rand                              { return c.RNG }
func (c *Character) GetTargetPriority() core.TargetPriority {
	return c.EntityState.GetTargetPrioritization()
}
func (c *Character) SetTargetPriority(p core.TargetPriority) {
	c.EntityState.SetTargetPrioritization(p)
}
func (c *Character) ModifyHP(value int, isTemp bool, tempStacking bool) (core.HPModificationResult, error) {
	return c.EntityState.ModifyHP(value, isTemp, tempStacking)
}
func (c *Character) ChooseSpellByHealingEfficiency(targetValue int) (*core.SpellChoice, error) {
	choice, err := c.SpellCastingManager.GetMostEfficientHealingSpell(targetValue)
	if err != nil {
		return nil, err
	}
	return choice, nil
}
func (c *Character) ChooseDamageSpellByPriority(p core.SpellPriority) (*core.SpellChoice, error) {
	return c.SpellCastingManager.ChooseSpellByPriority(core.STDamage, p)
}
func (c *Character) GetHealingSpellCount() int {
	return c.SpellCastingManager.GetHealingSpellCount()
}
func (c *Character) GetDamageSpellCount() int {
	return c.SpellCastingManager.GetDamageSpellCount()
}

func (c *Character) UpdateAICombatContext(ctx *core.CombatContext) error {
	c.AI.UpdateCombatContext(ctx)
	return nil
}

func (c *Character) GetAIRequest(actorID int, t core.AIRequestType) (*core.AIRequest, error) {
	var req *core.AIRequest
	var err error
	switch t {
	case core.AIReqChooseAction:
		req, err = c.AI.chooseCharacterAction()
		if err != nil {
			return nil, err
		}
	default:
		return req, fmt.Errorf("invalid AI request type: %s", t)
	}

	events.LogCharacterActionChoiceEvent(c, req.ActionType, c.EventListener)

	req.ActorID = actorID

	return req, nil
}

func (c *Character) ExecuteAIRequest(req *core.AIRequest) (*core.ActionOutcome, error) {
	switch req.ActionType {
	case core.ATMelee, core.ATRanged:
		attackReq, err := c.CreateAttackRequest(req.Target, req.WeaponSlot, req.Advantage, req.UseVersatile, req.SimOptions)
		if err != nil {
			return nil, err
		}

		results, err := c.MartialAttackManager.ProcessAttackRequest(attackReq)
		if err != nil {
			return nil, err
		}

		var effects []core.Effect
		for _, res := range results {
			if res.GetIsHit() {
				effects = append(effects, core.Effect{
					Type:       core.EffectDamage,
					Value:      res.GetDamageResult().GetTotal(),
					DamageType: res.GetDamageType(),
				})
			}
		}

		return &core.ActionOutcome{
			ActionType: req.ActionType,
			TargetID:   req.TargetID,
			ActorID:    req.ActorID,
			Effects:    effects,
		}, nil
	case core.ATSpell:
		scReq, err := c.CreateSpellCastRequest(*req.SpellChoice, req.Advantage, req.SimOptions)
		if err != nil {
			return nil, err
		}

		res, err := c.SpellCastingManager.CastSpell(scReq)
		if err != nil {
			return nil, err
		}

		var effects []core.Effect
		if res.GetIsHit() {
			if req.SpellChoice.Spell.GetSpellType() == core.STDamage {
				effects = append(effects, core.Effect{
					Type:       core.EffectDamage,
					Value:      res.GetSpellTotalValue(),
					DamageType: res.GetDamageType(),
				})
			} else if req.SpellChoice.Spell.GetSpellType() == core.STHealing {
				effects = append(effects, core.Effect{
					Type:  core.EffectHealing,
					Value: res.GetSpellTotalValue(),
				})
			}
		}

		return &core.ActionOutcome{
			ActionType: req.ActionType,
			TargetID:   req.TargetID,
			ActorID:    req.ActorID,
			Effects:    effects,
		}, nil
	case core.ATHeal:
		return nil, errors.New("not implemented")
	}
	return nil, errors.New("invalid action type")
}

var _ core.Entity = &Character{}
