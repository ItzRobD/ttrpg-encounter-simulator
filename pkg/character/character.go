package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/martial_attack_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/entity_configuration"
	"dnd5e-encounter-simulator-backend/pkg/races"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"math"
	"math/rand/v2"
)

// Character represents a player or NPC with attributes like name, class, level, and various managers for gameplay systems.
type Character struct {
	Name                 string
	Race                 races.Race
	Class                classes.Class
	Level                uint8
	AbilityScores        core.AbilityScores
	AbilityScoreProf     core.AbilityScoresProficiencies
	EntityStateManager   *entity_state_manager.EntityStateManager
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
	Name                string
	RaceID              races.RaceID
	DragonbornColor     *races.DragonbornColor
	ClassID             classes.ClassID
	Level               uint8
	AsConfig            core.AbilityScoresConfig
	HPMethod            core.HPSetMethod
	HPValue             int
	Seed                core.Seed
	Equipment           EquipmentConfig
	Resistances         core.DamageResistances
	FightingStyles      []classes.FightingStyle
	EntityConfiguration entity_configuration.EntityConfiguration
}

// EquipmentConfig defines configuration for a character's equipment including armor and weapon slot mapping.
// Weapons slots are of map[weaponID]isProficient
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
	if charConfig.RaceID < 0 || charConfig.RaceID > 9 {
		return nil, fmt.Errorf("invalid raceData id during character initialization: %d", charConfig.RaceID)
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
	// Setup class-specific features
	classData.ClassFeatures.SetupFeatures(classData.ID, charConfig.Level)

	// Apply Fighting Styles via class API to ensure they are registered properly
	if len(charConfig.FightingStyles) > 0 {
		for _, fs := range charConfig.FightingStyles {
			classData.AddFightingStyle(fs)
		}
	}
	raceData, err := races.QueryRaceData(ctx,
		races.RaceQueryParams{
			ID:              charConfig.RaceID,
			Level:           charConfig.Level,
			DragonbornColor: charConfig.DragonbornColor})
	if err != nil {
		return nil, err
	}

	// Set initial values for character
	char := Character{
		Name:                 charConfig.Name,
		Class:                classData,
		Race:                 raceData,
		Level:                charConfig.Level,
		AbilityScores:        charConfig.AsConfig.AbilityScores,
		AbilityScoreProf:     charConfig.AsConfig.Proficiencies,
		EntityStateManager:   &entity_state_manager.EntityStateManager{},
		EquipmentManager:     &equipment_manager.EquipmentManager{},
		SpellCastingManager:  &spellcasting_manager.SpellcastingManager{},
		MartialAttackManager: &martial_attack_manager.MartialAttackManager{},
		RollManager:          &roll_manager.RollManager{},
		AI:                   &CharacterAI{},
		Configuration:        charConfig.EntityConfiguration,
		HPConfig:             core.HPConfig{HPSetMethod: charConfig.HPMethod, Value: charConfig.HPValue},
		Seed:                 charConfig.Seed,
		RNG:                  nil,
	}

	// Initialize managers
	// Roll Manager
	char.RollManager = initializeRollManager(&char, &char.Configuration)
	if char.Race.ID == races.Halfling {
		char.RollManager.RerollAbilities.HasHalflingLucky = true
	}
	// Explicitly apply any extra reroll abilities from configuration
	if char.Configuration.CombatFeatures.ReRollAbilities.HasGreatWeaponFighting {
		char.RollManager.RerollAbilities.HasGreatWeaponFighting = true
	}
	if char.Configuration.CombatFeatures.ReRollAbilities.HasElementalAdept {
		char.RollManager.RerollAbilities.HasElementalAdept = true
	}
	if char.Configuration.CombatFeatures.ReRollAbilities.HasElvenAccuracy {
		char.RollManager.RerollAbilities.HasElvenAccuracy = true
	}
	if char.Configuration.CombatFeatures.ReRollAbilities.HasIndomitable {
		char.RollManager.RerollAbilities.HasIndomitable = true
	}
	// AI
	char.AI = NewCharacterAI(&char)
	// Entity State Manager
	esmConfig := entity_state_manager.EntityStateConfig{
		AttackCount: classData.AttackCount,
		Conditions:  core.NewEntityConditions(),
	}

	switch classData.ID {
	case classes.Barbarian:
		esmConfig.BarbarianRelentlessUses = 0
	case classes.Fighter:
		esmConfig.FighterIndomitableUses = classData.ClassFeatures.FighterFeatures.IndomitableUses
	case classes.Paladin:
		esmConfig.PaladinLayingOnHandsPool = classData.ClassFeatures.PaladinFeatures.LayOnHandsPool
	default:
		break
	}

	char.EntityStateManager, err = initializeEntityStateManager(&char, &esmConfig)
	if err != nil {
		return nil, err
	}

	// Apply class features to esm
	switch classData.ID {
	case classes.Barbarian:
		features := char.Class.ClassFeatures.BarbarianFeatures
		if features.HasDangerSense {
			char.EntityStateManager.SetHasSavingThrowAdvantage(core.AbilityDexterity, core.RollAdvantage)
		}
		if features.HasFeralInstinct {
			char.EntityStateManager.SetInitiativeAdvantage(core.RollAdvantage)
		}
		if char.EntityStateManager.BarbarianIsRaging {
			char.EntityStateManager.SetHasSavingThrowAdvantage(core.AbilityStrength, core.RollAdvantage)
			char.EntityStateManager.AddResistance(core.DamageSlashing, core.ResistanceResistant, nil)
			char.EntityStateManager.AddResistance(core.DamagePiercing, core.ResistanceResistant, nil)
			char.EntityStateManager.AddResistance(core.DamageBludgeoning, core.ResistanceResistant, nil)
		}
		char.EntityStateManager.SetBarbarianRelentlessRage(features.HasRelentlessRage)
	case classes.Rogue:
		features := char.Class.ClassFeatures.RogueFeatures
		if features.HasSlipperyMind {
			char.EntityStateManager.SetHasSavingThrowAdvantage(core.AbilityWisdom, core.RollAdvantage)
		}
	}

	// Apply race features to esm
	switch raceData.ID {
	case races.Dwarf:
		char.EntityStateManager.AddResistance(core.DamagePoison, core.ResistanceResistant, nil)
	case races.Dragonborn:
		if char.Race.DragonbornFeatures != nil {
			char.EntityStateManager.AddResistance(char.Race.DragonbornFeatures.DamageType, core.ResistanceResistant, nil)
		}
	case races.HalfOrc:
		char.EntityStateManager.HalfOrcHasSavageAttacks = true
		char.EntityStateManager.HalfOrcHasRelentlessEnduranceUse = true
	case races.Tiefling:
		char.EntityStateManager.AddResistance(core.DamageFire, core.ResistanceResistant, nil)
	default:
		break
	}

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

	err = char.setHP(char.HPConfig)
	if err != nil {
		return nil, err
	}

	return &char, nil
}

// NewCharacterWithRNG initializes a Character using a provided RNG. This avoids creating a new PCG instance
// and ensures all randomness derives from the simulation manager's RNG.
func NewCharacterWithRNG(ctx context.Context, charConfig CharacterConfig, rng *rand.Rand) (*Character, error) {
	if charConfig.ClassID < 0 || charConfig.ClassID > 13 {
		return nil, fmt.Errorf("invalid classData id during character initialization: %d", charConfig.ClassID)
	}
	if charConfig.RaceID < 0 || charConfig.RaceID > 9 {
		return nil, fmt.Errorf("invalid raceData id during character initialization: %d", charConfig.RaceID)
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
	// Setup class-specific features
	classData.ClassFeatures.SetupFeatures(classData.ID, charConfig.Level)

	// Apply Fighting Styles via class API to ensure they are registered properly
	if len(charConfig.FightingStyles) > 0 {
		for _, fs := range charConfig.FightingStyles {
			classData.AddFightingStyle(fs)
		}
	}
	raceData, err := races.QueryRaceData(ctx,
		races.RaceQueryParams{
			ID:              charConfig.RaceID,
			Level:           charConfig.Level,
			DragonbornColor: charConfig.DragonbornColor})
	if err != nil {
		return nil, err
	}

	// Initialize character with provided RNG
	char := Character{
		Name:                 charConfig.Name,
		Class:                classData,
		Race:                 raceData,
		Level:                charConfig.Level,
		AbilityScores:        charConfig.AsConfig.AbilityScores,
		AbilityScoreProf:     charConfig.AsConfig.Proficiencies,
		EntityStateManager:   &entity_state_manager.EntityStateManager{},
		EquipmentManager:     &equipment_manager.EquipmentManager{},
		SpellCastingManager:  &spellcasting_manager.SpellcastingManager{},
		MartialAttackManager: &martial_attack_manager.MartialAttackManager{},
		RollManager:          &roll_manager.RollManager{},
		AI:                   &CharacterAI{},
		Configuration:        charConfig.EntityConfiguration,
		HPConfig:             core.HPConfig{HPSetMethod: charConfig.HPMethod, Value: charConfig.HPValue},
		Seed:                 charConfig.Seed,
		RNG:                  rng,
	}

	// Managers that depend on RNG
	char.RollManager = initializeRollManager(&char, &char.Configuration)
	if char.Race.ID == races.Halfling {
		char.RollManager.RerollAbilities.HasHalflingLucky = true
	}
	// Explicitly apply any extra reroll abilities from configuration
	if char.Configuration.CombatFeatures.ReRollAbilities.HasGreatWeaponFighting {
		char.RollManager.RerollAbilities.HasGreatWeaponFighting = true
	}
	if char.Configuration.CombatFeatures.ReRollAbilities.HasElementalAdept {
		char.RollManager.RerollAbilities.HasElementalAdept = true
	}
	if char.Configuration.CombatFeatures.ReRollAbilities.HasElvenAccuracy {
		char.RollManager.RerollAbilities.HasElvenAccuracy = true
	}
	if char.Configuration.CombatFeatures.ReRollAbilities.HasIndomitable {
		char.RollManager.RerollAbilities.HasIndomitable = true
	}
	char.AI = NewCharacterAI(&char)

	// Entity State Manager
	esmConfig := entity_state_manager.EntityStateConfig{
		AttackCount: classData.AttackCount,
		Conditions:  core.NewEntityConditions(),
	}

	switch classData.ID {
	case classes.Barbarian:
		esmConfig.BarbarianRelentlessUses = 0
	case classes.Fighter:
		esmConfig.FighterIndomitableUses = classData.ClassFeatures.FighterFeatures.IndomitableUses
	case classes.Paladin:
		esmConfig.PaladinLayingOnHandsPool = classData.ClassFeatures.PaladinFeatures.LayOnHandsPool
	default:
		break
	}

	char.EntityStateManager, err = initializeEntityStateManager(&char, &esmConfig)
	if err != nil {
		return nil, err
	}

	// Apply class features to esm
	switch classData.ID {
	case classes.Barbarian:
		features := char.Class.ClassFeatures.BarbarianFeatures
		if features.HasDangerSense {
			char.EntityStateManager.SetHasSavingThrowAdvantage(core.AbilityDexterity, core.RollAdvantage)
		}
		if features.HasFeralInstinct {
			char.EntityStateManager.SetInitiativeAdvantage(core.RollAdvantage)
		}
		if char.EntityStateManager.BarbarianIsRaging {
			char.EntityStateManager.SetHasSavingThrowAdvantage(core.AbilityStrength, core.RollAdvantage)
			char.EntityStateManager.AddResistance(core.DamageSlashing, core.ResistanceResistant, nil)
			char.EntityStateManager.AddResistance(core.DamagePiercing, core.ResistanceResistant, nil)
			char.EntityStateManager.AddResistance(core.DamageBludgeoning, core.ResistanceResistant, nil)
		}
		char.EntityStateManager.SetBarbarianRelentlessRage(features.HasRelentlessRage)
	case classes.Rogue:
		features := char.Class.ClassFeatures.RogueFeatures
		if features.HasSlipperyMind {
			char.EntityStateManager.SetHasSavingThrowAdvantage(core.AbilityWisdom, core.RollAdvantage)
		}
	}

	// Apply race features to esm
	switch raceData.ID {
	case races.Dwarf:
		char.EntityStateManager.AddResistance(core.DamagePoison, core.ResistanceResistant, nil)
	case races.Dragonborn:
		if char.Race.DragonbornFeatures != nil {
			char.EntityStateManager.AddResistance(char.Race.DragonbornFeatures.DamageType, core.ResistanceResistant, nil)
		}
	case races.HalfOrc:
		char.EntityStateManager.HalfOrcHasSavageAttacks = true
		char.EntityStateManager.HalfOrcHasRelentlessEnduranceUse = true
	case races.Tiefling:
		char.EntityStateManager.AddResistance(core.DamageFire, core.ResistanceResistant, nil)
	default:
		break
	}

	// Equipment Manager
	char.EquipmentManager, err = equipment_manager.NewEquipmentManager(&char)
	if err != nil {
		return nil, err
	}
	if err = char.setupEquipmentFromConfig(ctx, charConfig.Equipment); err != nil {
		return nil, err
	}

	// Spellcasting Manager
	char.SpellCastingManager, err = initializeSpellcastingManager(ctx, &char)
	if err != nil {
		return nil, err
	}

	// Martial Attack Manager
	char.MartialAttackManager = martial_attack_manager.NewMartialAttackManager(&char, char.RollManager)

	// HP configuration
	char.HPConfig.HitDie = char.GetHitDie()
	char.HPConfig.NumberOfDice = int(char.Level - 1)
	modifier, _ := char.getAbilityScoreModifier(core.AbilityConstitution)
	char.HPConfig.HPAverage = int(math.Round(float64(char.GetHitDie().Int())+float64(char.HPConfig.NumberOfDice)*char.Class.HitDie.Avg()) + float64(int(char.Level)*modifier))
	char.HPConfig.Modifier = modifier
	if err = char.setHP(char.HPConfig); err != nil {
		return nil, err
	}

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

	// Upcast decision is deferred to combat-time via CombatContext options
	sm := spellcasting_manager.NewSpellcastingManager(c, c.RollManager, core.CasterCharacter, int(c.Level), slots, slots, spellModValue)
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
		c.EntityStateManager.SetHPValues(hp)

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
		c.EntityStateManager.SetHPValues(hp)

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
		c.EntityStateManager.SetHPValues(hp)
		return nil
	default:
		return fmt.Errorf("invalid HP set method: %v", config.HPSetMethod)
	}
}

// Interface Required Functions
func (c *Character) IsCharacter() bool                          { return true }
func (c *Character) IsMonster() bool                            { return false }
func (c *Character) GetIsLegendary() bool                       { return false }
func (c *Character) RefreshLegendaryActions()                   { return }
func (c *Character) GetEventListener() func(event interface{})  { return c.EventListener }
func (c *Character) SetEventListener(f func(event interface{})) { c.EventListener = f }
func (c *Character) IsUnconscious() bool                        { return c.EntityStateManager.GetIsUnconscious() }
func (c *Character) GetClassID() uint8                          { return uint8(c.Class.ID) }
func (c *Character) IsDead() bool                               { return c.EntityStateManager.GetIsDead() }
func (c *Character) GetHPStatus() core.HPStatus                 { return c.EntityStateManager.GetHPStatus() }
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
func (c *Character) GetState() interface{}                           { return c.EntityStateManager }
func (c *Character) InitializeHP() error                             { return c.setHP(c.HPConfig) }
func (c *Character) IsSpellcaster() bool                             { return c.SpellCastingManager.HasAnyKnownSpells() }
func (c *Character) IsHealer() bool {
	return c.SpellCastingManager.HasHealingSpells() ||
		(c.Class.ID == classes.Paladin && c.EntityStateManager.GetPaladinLayingOnHandsPool() > 0)
}
func (c *Character) GetRNG() *rand.Rand { return c.RNG }
func (c *Character) GetTargetPriority() core.TargetPriority {
	return c.EntityStateManager.GetTargetPrioritization()
}
func (c *Character) SetTargetPriority(p core.TargetPriority) {
	c.EntityStateManager.SetTargetPrioritization(p)
}
func (c *Character) ModifyHP(value int, isTemp bool, tempStacking bool) (core.HPModificationResult, error) {
	return c.EntityStateManager.ModifyHP(value, isTemp, tempStacking)
}

func (c *Character) GetHealingSpellCount() int {
	return c.SpellCastingManager.GetHealingSpellCount()
}

func (c *Character) GetDamageSpellCount() int {
	return c.SpellCastingManager.GetDamageSpellCount()
}

func (c *Character) UpdateAICombatContext(ctx *core.CombatContext) error {
	c.AI.UpdateCombatContext(ctx)
	if c.SpellCastingManager != nil {
		c.SpellCastingManager.SetSimulationOptions(ctx.Opt())
	}
	return nil
}

func (c *Character) CanTakeActions() bool { return c.EntityStateManager.CanTakeActions() }
func (c *Character) GetConditions() core.EntityConditions {
	return c.EntityStateManager.GetConditions()
}

func (c *Character) GetType() string {
	return "Humanoid"
}

func (c *Character) HasElusive() bool {
	if c.Class.ID == classes.Rogue && c.Class.ClassFeatures.RogueFeatures != nil {
		return c.Class.ClassFeatures.RogueFeatures.HasElusive
	}

	return false
}

var _ core.Entity = &Character{}
