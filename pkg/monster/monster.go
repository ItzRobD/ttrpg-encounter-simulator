package monster

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"fmt"
	"math/rand/v2"
)

type Monster struct {
	MonsterBase
	EntityState         *entity_state_manager.EntityStateManager
	SpellCastingManager *spellcasting_manager.SpellcastingManager
	RollManager         *roll_manager.RollManager
	AI                  *MonsterAI
	ActionManager       *MonsterActionManager
	Seed                core.Seed
	RNG                 *rand.Rand
	EventListener       func(event interface{})
}

type MonsterBase struct {
	ID                  int
	Name                string
	Size                string
	Type                string
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
		EntityState:         &entity_state_manager.EntityStateManager{},
		SpellCastingManager: &spellcasting_manager.SpellcastingManager{},
		RollManager:         &roll_manager.RollManager{},
		AI:                  &MonsterAI{},
		ActionManager:       &MonsterActionManager{},
		Seed:                seed,
		RNG:                 rand.New(rand.NewPCG(seed.Seed1, seed.Seed2)),
	}

	// Initialize managers
	var err error
	// Roll manager
	monster.RollManager = initializeRollManager(&monster)

	// ESM
	esmConfig := entity_state_manager.EntityStateConfig{
		AttackCount: 1,
		Conditions:  core.NewEntityConditions(),
	}
	if monster.IsLegendary {
		esmConfig.MaxLegendaryActions = 3
	}

	monster.EntityState, err = initalizeEntityStateManager(&monster, esmConfig)
	if err != nil {
		return nil, err
	}
	monster.EntityState.Resistances = config.Resistances

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
	monster.AI = NewMonsterAI(&monster)

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
	// TODO: Sim options can be passed via context for simplification
	canUpcast := ctx.Value("CanUpcast").(bool)
	casterType := core.CasterMonsterTrueCaster
	if m.IsInnateSpellcaster {
		casterType = core.CasterMonsterInnate
	}
	sm := spellcasting_manager.NewSpellcastingManager(m, m.RollManager, casterType, config.CastingLevel, config.SpellSlots, config.SpellSlots, canUpcast, config.AttackModifier)
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
		m.EntityState.SetHPValues(hp)

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
		m.EntityState.SetHPValues(hp)

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
		m.EntityState.SetHPValues(hp)
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
	return m.EntityState
}
func (m *Monster) GetName() string { return m.Name }
func (m *Monster) GetAbilityScores() core.AbilityScores {
	return m.AbilityScores
}
func (m *Monster) GetHPStatus() core.HPStatus {
	return m.EntityState.GetHPStatus()
}
func (m *Monster) GetHitDie() core.DiceType   { return m.EntityState.GetHitDie() }
func (m *Monster) GetAC() int                 { return m.AC }
func (m *Monster) GetLevel() float64          { return m.CR }
func (m *Monster) GetHPConfig() core.HPConfig { return m.HP }
func (m *Monster) SetHP(method core.HPSetMethod, value int) error {
	return m.setHP(core.HPConfig{HPSetMethod: method, Value: value})
}
func (m *Monster) IsUnconscious() bool  { return m.EntityState.GetIsUnconscious() }
func (m *Monster) GetClassID() uint8    { return 0 }
func (m *Monster) IsDead() bool         { return m.EntityState.GetIsDead() }
func (m *Monster) IsCharacter() bool    { return false }
func (m *Monster) IsMonster() bool      { return true }
func (m *Monster) GetIsLegendary() bool { return m.MonsterBase.IsLegendary }
func (m *Monster) GetRNG() *rand.Rand   { return m.RNG }
func (m *Monster) GetID() int           { return m.ID }
func (m *Monster) InitializeHP() error  { return m.setHP(m.HP) }
func (m *Monster) IsSpellcaster() bool  { return m.MonsterBase.IsSpellcaster }
func (m *Monster) IsHealer() bool       { return m.SpellCastingManager.HasHealingSpells() }
func (m *Monster) GetTargetPriority() core.TargetPriority {
	return m.EntityState.TargetPrioritization
}
func (m *Monster) SetTargetPriority(priority core.TargetPriority) {
	m.EntityState.TargetPrioritization = priority
}
func (m *Monster) ModifyHP(value int, isTemp bool, tempStacking bool) (core.HPModificationResult, error) {
	return m.EntityState.ModifyHP(value, isTemp, tempStacking)
}

func (m *Monster) CanTakeActions() bool { return m.EntityState.CanTakeActions() }
func (m *Monster) GetConditions() core.EntityConditions {
	return m.EntityState.GetConditions()
}

var _ core.Entity = &Monster{}
