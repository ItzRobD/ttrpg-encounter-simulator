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
	"dnd5e-encounter-simulator-backend/pkg/spells"
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

// Interface Required Functions
func (c *Character) IsCharacter() bool                          { return true }
func (c *Character) IsMonster() bool                            { return false }
func (c *Character) GetIsLegendary() bool                       { return false }
func (c *Character) RefreshLegendaryActions()                   { return }
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

func (c *Character) CanTakeActions() bool { return c.EntityState.CanTakeActions() }

var _ core.Entity = &Character{}
