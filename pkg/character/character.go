package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/martial_attack_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/entity_configuration"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
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
	SpellcastingManager  *spellcasting_manager.SpellcastingManager
	MartialAttackManager *martial_attack_manager.MartialAttackManager
	RollManager          *roll_manager.RollManager
	Configuration        entity_configuration.EntityConfiguration
	EventListener        func(event interface{})
}

// New initializes and returns a new Character with the specified parameters or an error if validation fails.
func New(ctx context.Context, name string, classID classes.ClassID, level uint8, asConfig core.AbilityScoresConfig) (Character, error) {
	if classID < 0 || classID > 13 {
		return Character{}, fmt.Errorf("invalid class id during character initialization: %d", classID)
	}
	if level < 0 || level > 20 {
		return Character{}, fmt.Errorf("invalid level during character initialization, must be in range 1-20: %d", level)
	}
	if name == "" {
		name = "Unnamed Character"
	}

	var params classes.ClassQueryParams
	params.ID = classID
	params.Level = level
	c, err := classes.QueryClassData(ctx, params)
	if err != nil {
		return Character{}, err
	}

	// Set initial values for character
	char := Character{
		Name:                 name,
		Class:                c,
		Level:                level,
		AbilityScores:        asConfig.AbilityScores,
		AbilityScoreProf:     asConfig.Proficiencies,
		EntityState:          &entity_state_manager.EntityStateManager{},
		EquipmentManager:     &equipment_manager.EquipmentManager{},
		SpellcastingManager:  &spellcasting_manager.SpellcastingManager{},
		MartialAttackManager: &martial_attack_manager.MartialAttackManager{},
		RollManager:          &roll_manager.RollManager{},
		Configuration:        entity_configuration.EntityConfiguration{},
	}

	// Initialize managers
	// Roll Manager
	char.RollManager = initializeRollManager(ctx, &char, &char.Configuration)

	// Entity State Manager
	config := entity_state_manager.EntityStateConfig{
		AttackCount: c.AttackCount,
		Conditions:  core.NewEntityConditions(),
	}
	esm, err := initalizeEntityStateManager(ctx, &char, &config)
	if err != nil {
		return Character{}, err
	}
	char.EntityState = esm

	// Equipment Manager
	char.EquipmentManager, err = equipment_manager.NewEquipmentManager(&char)
	if err != nil {
		return Character{}, err
	}

	// Spellcasting Manager
	char.SpellcastingManager, err = initializeSpellcastingManager(ctx, &char)
	if err != nil {
		return Character{}, err
	}

	// Martial Attack Manager
	char.MartialAttackManager = martial_attack_manager.NewMartialAttackManager(&char, char.RollManager)

	return char, nil
}

func initializeRollManager(ctx context.Context, c *Character, eConfig *entity_configuration.EntityConfiguration) *roll_manager.RollManager {
	rm := roll_manager.NewRollManager(c, eConfig.CombatFeatures.ReRollAbilities)
	return rm
}

func initalizeEntityStateManager(ctx context.Context, c *Character, config *entity_state_manager.EntityStateConfig) (*entity_state_manager.EntityStateManager, error) {
	esm, err := entity_state_manager.NewEntityStateManager(c, *config)
	if err != nil {
		return nil, err
	}
	return esm, nil
}

// initializeSpellcastingManager initializes and configures a SpellcastingManager for the provided character.
// It retrieves the spell slots and usable spell IDs based on the character's level and class.
// Returns a pointer to the SpellcastingManager and an error if any issue occurs during initialization.
func initializeSpellcastingManager(ctx context.Context, c *Character) (*spellcasting_manager.SpellcastingManager, error) {
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

	sm.SetUsableSpellIDs(availableSpellIDs)

	return sm, nil
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

// CreateWeaponAttackData generates an AttackData object for a given weapon slot, considering proficiency and versatility.
// slot specifies the weapon slot to retrieve the weapon from.
// useVersatile indicates whether to use the weapon in versatile mode, if applicable.
// Returns the constructed AttackData and an error if any issue occurs in retrieving or calculating weapon properties.
func (c *Character) CreateWeaponAttackData(slot core.WeaponSlot, useVersatile bool) (martial_attack_manager.AttackData, error) {
	w, err := c.EquipmentManager.GetWeaponFromSlot(slot)
	if err != nil {
		return martial_attack_manager.AttackData{}, err
	}

	prof := c.EquipmentManager.GetIsProficientWithSlot(slot)
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
func (c *Character) CreateAttackRequest(target core.Entity, slot core.WeaponSlot, useVersatile bool, advantage core.AdvantageType, simulationOptions core.SimulationOptions) (*martial_attack_manager.AttackRequest, error) {
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
func (c *Character) CreateSpellCastRequest(spellChoice core.SpellChoice, options spellcasting_manager.SpellOptions, simOptions core.SimulationOptions) (*spellcasting_manager.SpellCastRequest, error) {
	spellcastData, err := c.CreateSpellAttackData(spellChoice)
	if err != nil {
		return nil, err
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

	if c.GetIsProficientInAbility(ability) {
		return pb + mod, nil
	}
	return mod, nil
}

// Interface Required Functions

func (c *Character) IsCharacter() bool { return true }
func (c *Character) IsMonster() bool   { return false }
func (c *Character) GetEventListener() func(event interface{}) {
	return c.EventListener
}
func (c *Character) IsUnconscious() bool                  { return c.EntityState.GetIsUnconscious() }
func (c *Character) GetHPStatus() core.HPStatus           { return c.EntityState.GetHPStatus() }
func (c *Character) GetName() string                      { return c.Name }
func (c *Character) GetAbilityScores() core.AbilityScores { return c.AbilityScores }
func (c *Character) GetLevel() interface{}                { return c.Level }
func (c *Character) GetCasterLevel() uint8                { return c.Level }
func (c *Character) GetAC() int                           { return c.EquipmentManager.GetAC() }
func (c *Character) GetAbilityScore(a core.Ability) int   { return c.getAbilityScore(a) }
func (c *Character) GetAbilityScoreModifier(a core.Ability) (int, error) {
	return c.getAbilityScoreModifier(a)
}
func (c *Character) GetSavingThrowBonus(a core.Ability) (int, error) { return c.getSavingThrowBonus(a) }

var _ core.Entity = &Character{}
