package monster

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/entity_configuration"
	"fmt"
	"math/rand/v2"
)

type Monster struct {
	MonsterBase
	Info                *core.CombatantInfo
	EntityStateManager  *entity_state_manager.EntityStateManager
	SpellCastingManager *spellcasting_manager.SpellcastingManager
	RollManager         *roll_manager.RollManager
	AI                  *MonsterAI
	ActionManager       *MonsterActionManager
	Configuration       entity_configuration.EntityConfiguration
	Seed                core.Seed
	RNG                 *rand.Rand
	EventListener       func(event interface{})
}

type MonsterBase struct {
	ID                  int
	InstanceID          int
	Name                string
	Size                string
	Type                MonsterType
	AC                  int
	ProficiencyBonus    int
	CR                  float64
	ApiURL              string
	IsLegendary         bool
	IsSpellcaster       bool
	IsInnateSpellcaster bool
	AbilityScores       core.AbilityScores
	AbilityScoreProf    core.AbilityScoresProficiencies
	HP                  core.HPConfig
	SpecialAbilities    SpecialAbilities
}

type MonsterQueryParams struct {
	Name []string
	ID   []int
}

func NewMonster(ctx context.Context, config MonsterConfig) (*Monster, error) {
	var seed core.Seed
	if config.Seed.Seed1 == 0 {
		seed.Seed1 = rand.Uint64()
	}
	if config.Seed.Seed2 == 0 {
		seed.Seed2 = rand.Uint64()
	}

	monster := Monster{
		MonsterBase:         config.Base,
		EntityStateManager:  &entity_state_manager.EntityStateManager{},
		SpellCastingManager: &spellcasting_manager.SpellcastingManager{},
		RollManager:         &roll_manager.RollManager{},
		AI:                  &MonsterAI{},
		ActionManager:       &MonsterActionManager{},
		Configuration:       config.EntityConfiguration,
		Seed:                seed,
		RNG:                 rand.New(rand.NewPCG(seed.Seed1, seed.Seed2)),
	}

	// Initialize managers
	var err error
	// Roll manager
	monster.RollManager = initializeRollManager(&monster)

	// ESM
	esmConfig := entity_state_manager.EntityStateConfig{
		AttackCount:         1,
		Conditions:          core.NewEntityConditions(),
		RelentlessThreshold: monster.SpecialAbilities.RelentlessThreshold,
		HasUndeadFortitude:  monster.SpecialAbilities.UndeadFortitude,
	}
	if monster.IsLegendary {
		esmConfig.MaxLegendaryActions = 3
	}

	monster.EntityStateManager, err = initalizeEntityStateManager(&monster, esmConfig)
	if err != nil {
		return nil, err
	}
	// Only override the initialized resistances if the config provides a non-nil map.
	// This avoids clobbering the default NewDamageResistances() with nil when a monster
	// has no DB-defined resistances or a query returns no rows for that monster.
	if config.Resistances != nil {
		monster.EntityStateManager.SetResistances(config.Resistances)
	}

	// Spellcasting Manager
	if monster.MonsterBase.IsSpellcaster {
		monster.SpellCastingManager, err = initializeSpellcastingManager(ctx, &monster, config.spellcastingConfig)
		if err != nil {
			return nil, err
		}
	}

	// Action manager
	mamConfig := &MAMConfig{
		Actions:          config.Actions,
		Multiattacks:     config.Multiattacks,
		LegendaryActions: config.LegendaryActions,
		SpecialAbilities: config.SpecialAbilities,
	}
	monster.ActionManager = initializeActionManager(&monster, mamConfig)

	// Set up HP SimOptions and monster hp
	monster.HP.HPSetMethod = config.HPSetMethod

	// AI
	monster.AI = NewMonsterAI(&monster, config.UtilityWeights)

	err = monster.setHP(monster.HP)
	if err != nil {
		return nil, err
	}

	return &monster, nil
}

// NewMonsterWithRNG initializes a Monster using the provided RNG without creating a new PCG.
func NewMonsterWithRNG(ctx context.Context, config MonsterConfig, rng *rand.Rand) (*Monster, error) {
	monster := Monster{
		MonsterBase:         config.Base,
		EntityStateManager:  &entity_state_manager.EntityStateManager{},
		SpellCastingManager: &spellcasting_manager.SpellcastingManager{},
		RollManager:         &roll_manager.RollManager{},
		AI:                  &MonsterAI{},
		ActionManager:       &MonsterActionManager{},
		Configuration:       config.EntityConfiguration,
		Seed:                config.Seed,
		RNG:                 rng,
	}

	// Initialize managers
	var err error
	// Roll manager
	monster.RollManager = initializeRollManager(&monster)

	// ESM
	esmConfig := entity_state_manager.EntityStateConfig{
		AttackCount:         1,
		Conditions:          core.NewEntityConditions(),
		RelentlessThreshold: monster.SpecialAbilities.RelentlessThreshold,
	}
	if monster.IsLegendary {
		esmConfig.MaxLegendaryActions = 3
	}

	monster.EntityStateManager, err = initalizeEntityStateManager(&monster, esmConfig)
	if err != nil {
		return nil, err
	}
	// Only override initialized resistances if a non-nil map is provided
	if config.Resistances != nil {
		monster.EntityStateManager.SetResistances(config.Resistances)
	}

	// Spellcasting Manager
	if monster.MonsterBase.IsSpellcaster {
		monster.SpellCastingManager, err = initializeSpellcastingManager(ctx, &monster, config.spellcastingConfig)
		if err != nil {
			return nil, err
		}
	}

	// Action manager
	mamConfig := &MAMConfig{
		Actions:          config.Actions,
		Multiattacks:     config.Multiattacks,
		LegendaryActions: config.LegendaryActions,
		SpecialAbilities: config.SpecialAbilities,
	}
	monster.ActionManager = initializeActionManager(&monster, mamConfig)

	// Set up HP SimOptions and monster hp
	monster.HP.HPSetMethod = config.HPSetMethod

	// AI
	monster.AI = NewMonsterAI(&monster, config.UtilityWeights)

	err = monster.setHP(monster.HP)
	if err != nil {
		return nil, err
	}

	return &monster, nil
}

func initializeRollManager(m *Monster) *roll_manager.RollManager {
	rm := roll_manager.NewRollManager(m, roll_manager.RerollAbilities{})
	return rm
}

func initalizeEntityStateManager(m *Monster, config entity_state_manager.EntityStateConfig) (*entity_state_manager.EntityStateManager, error) {
	esm, err := entity_state_manager.NewEntityStateManager(m, config)
	if err != nil {
		return nil, err
	}

	return esm, nil
}

func initializeSpellcastingManager(ctx context.Context, m *Monster, config MonsterSpellcastingConfig) (*spellcasting_manager.SpellcastingManager, error) {
	casterType := core.CasterMonsterTrueCaster
	if m.IsInnateSpellcaster {
		casterType = core.CasterMonsterInnate
	}
	// Upcast decision is deferred to combat-time via CombatContext options
	sm := spellcasting_manager.NewSpellcastingManager(m, m.RollManager, casterType, config.CastingLevel, config.SpellSlots, config.SpellSlots, config.AttackModifier)
	sm.SetAbility(config.Ability)
	sm.SetSaveDC(config.SaveDC)

	if casterType == core.CasterMonsterInnate {
		err := sm.AddKnownInnateSpells(config.InnateSpells)
		if err != nil {
			return nil, err
		}
	} else {
		err := sm.AddKnownSpells(config.LeveledSpells)
		if err != nil {
			return nil, err
		}
	}

	return sm, nil
}

func initializeActionManager(m *Monster, config *MAMConfig) *MonsterActionManager {
	mam := NewMonsterActionManager(m, m.RollManager, config)
	return mam
}

func (m *Monster) setHP(config core.HPConfig) error {
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
		events.LogDiceRollEvent(m, &hpRoll, m.EventListener)
		m.EntityStateManager.SetHPValues(hp)

		return nil
	case core.HPSetAverage:
		hp := entity_state_manager.HPValues{
			CurrentHP: config.HPAverage,
			MaxHP:     config.HPAverage,
			TempHP:    0,
			HitDie:    config.HitDie,
		}
		hpRoll := roll_manager.RollResult{
			DiceRollType:   core.DiceRollHPAvgUsedMonster,
			Die:            config.HitDie,
			FinalRollValue: config.HPAverage,
			Total:          config.HPAverage,
		}
		events.LogDiceRollEvent(m, &hpRoll, m.EventListener)
		m.EntityStateManager.SetHPValues(hp)

		return nil
	case core.HPSetRoll:
		hpRoll, err := m.RollManager.RollHP(config)
		if err != nil {
			return err
		}
		if hpRoll.Total <= 0 {
			hpRoll.Total = 1
			hpRoll.FinalRollValue = 1
		}
		hp := entity_state_manager.HPValues{
			CurrentHP: hpRoll.Total,
			MaxHP:     hpRoll.Total,
			TempHP:    0,
			HitDie:    hpRoll.Die,
		}
		m.EntityStateManager.SetHPValues(hp)
		return nil
	default:
		return fmt.Errorf("invalid HP set method: %v", config.HPSetMethod)
	}
}

func (m *Monster) GetEventListener() func(event interface{}) {
	return m.EventListener
}

func (m *Monster) SetEventListener(listener func(event interface{})) {
	m.EventListener = listener
}

func (m *Monster) GetState() interface{} {
	return m.EntityStateManager
}
func (m *Monster) GetName() string { return m.Name }
func (m *Monster) GetAbilityScores() core.AbilityScores {
	return m.AbilityScores
}
func (m *Monster) GetHPStatus() core.HPStatus {
	return m.EntityStateManager.GetHPStatus()
}
func (m *Monster) GetHitDie() core.DiceType   { return m.EntityStateManager.GetHitDie() }
func (m *Monster) GetAC() int                 { return m.AC }
func (m *Monster) GetLevel() float64          { return m.CR }
func (m *Monster) GetHPConfig() core.HPConfig { return m.HP }
func (m *Monster) SetHP(method core.HPSetMethod, value int) error {
	return m.setHP(core.HPConfig{HPSetMethod: method, Value: value})
}
func (m *Monster) IsUnconscious() bool  { return m.EntityStateManager.GetIsUnconscious() }
func (m *Monster) GetClassID() uint8    { return 0 }
func (m *Monster) IsDead() bool         { return m.EntityStateManager.GetIsDead() }
func (m *Monster) IsCharacter() bool    { return false }
func (m *Monster) IsMonster() bool      { return true }
func (m *Monster) GetIsLegendary() bool { return m.MonsterBase.IsLegendary }
func (m *Monster) GetRNG() *rand.Rand   { return m.RNG }
func (m *Monster) GetID() int {
	return m.MonsterBase.ID
}

func (m *Monster) GetInstanceID() int {
	return m.MonsterBase.InstanceID
}

func (m *Monster) SetInstanceID(id int) {
	m.MonsterBase.InstanceID = id
}

func (m *Monster) InitializeHP() error { return m.setHP(m.HP) }
func (m *Monster) IsSpellcaster() bool { return m.MonsterBase.IsSpellcaster }
func (m *Monster) IsHealer() bool      { return m.SpellCastingManager.HasHealingSpells() }
func (m *Monster) GetTargetPriority() core.TargetPriority {
	return m.EntityStateManager.GetTargetPrioritization()
}
func (m *Monster) SetTargetPriority(priority core.TargetPriority) {
	m.EntityStateManager.SetTargetPrioritization(priority)
}
func (m *Monster) ModifyHP(value int, isTemp bool, tempStacking bool, allowMassiveDamage bool, damageType core.DamageType, isCritical bool) (core.HPModificationResult, error) {
	return m.EntityStateManager.ModifyHP(value, isTemp, tempStacking, allowMassiveDamage, damageType, isCritical)
}

func (m *Monster) CanTakeActions() bool { return m.EntityStateManager.CanTakeActions() }
func (m *Monster) GetConditions() core.EntityConditions {
	return m.EntityStateManager.GetConditions()
}

func (m *Monster) GetType() string {
	return m.MonsterBase.Type.String()
}

func (m *Monster) IsConcentrating() bool {
	return m.EntityStateManager.IsConcentrating()
}

func (m *Monster) BreakConcentration() {
	m.EntityStateManager.BreakConcentration()
}

func (m *Monster) SetConcentrating(val bool, spellName string) {
	m.EntityStateManager.SetConcentrating(val, spellName)
}

func (m *Monster) Regenerate() {
	if m.AI.GetCombatContext() == nil {
		return
	}

	ctx := m.AI.GetCombatContext()

	if ctx.Opt().EnableSpecialAbilities && m.SpecialAbilities.RegenerationValue > 0 && !m.IsDead() {
		res, err := m.EntityStateManager.ModifyHP(m.SpecialAbilities.RegenerationValue, false, false, true, core.DamageNone, false)
		if err != nil {
			return
		}
		events.LogSpecialAbilityEvent(m, "Regeneration", fmt.Sprintf("%s uses Regeneration.", m.Name), "", m.SpecialAbilities.RegenerationValue, m.EventListener)
		events.LogHPModifiedEvent(m, m, res, m.EventListener)
	}
}

func (m *Monster) GetHasTakenTurnInCombat() bool {
	return m.EntityStateManager.GetHasTakenTurnInCombat()
}

var _ core.Entity = &Monster{}
