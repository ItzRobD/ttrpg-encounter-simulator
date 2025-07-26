package monster

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/monster_action_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"fmt"
)

// TODO: Action manager should be complete
//	Next is to clean up monster spellcasting -> SpellCastingManager
//	Configure entity state -> db queries fore resistances, we didn't add any resistatnace
//	functionaility for characters. That also needs to be completed

type Monster struct {
	MonsterBase
	EntityState         *entity_state_manager.EntityStateManager
	SpellCastingManager *spellcasting_manager.SpellcastingManager
	RollManager         *roll_manager.RollManager
	ActionManager       *monster_action_manager.MonsterActionManager
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
	HP                  MonsterHP
}

type MonsterHP struct {
	HPAverage    int
	NumberOfDice int
	Die          core.DiceType
	AmountToAdd  int
}

type MonsterQueryParams struct {
	Name []string
	ID   []int
}

func NewMonster(ctx context.Context, config MonsterConfig) (*Monster, error) {
	monster := &Monster{
		MonsterBase:         config.Base,
		EntityState:         &entity_state_manager.EntityStateManager{},
		SpellCastingManager: &spellcasting_manager.SpellcastingManager{},
		RollManager:         &roll_manager.RollManager{},
		ActionManager:       &monster_action_manager.MonsterActionManager{},
	}

	// Initialize managers
	var err error
	// Roll manager
	monster.RollManager = initializeRollManager(monster)

	// ESM
	esmConfig := entity_state_manager.EntityStateConfig{
		AttackCount: 1,
		Conditions:  core.NewEntityConditions(),
	}
	if monster.IsLegendary {
		esmConfig.MaxLegendaryActions = 3
	}

	monster.EntityState, err = initalizeEntityStateManager(monster, esmConfig)
	if err != nil {
		return nil, err
	}

	// Spellcasting Manager
	monster.SpellCastingManager, err = initializeSpellcastingManager(ctx, monster, config.spellcastingConfig)
	if err != nil {
		return nil, err
	}

	// Action manager
	mamConfig := &monster_action_manager.MAMConfig{
		Actions:          config.Actions,
		Multiattacks:     config.Multiattacks,
		LegendaryActions: config.LegendaryActions,
		SpecialAbilities: config.SpecialAbilities,
	}
	monster.ActionManager = initializeActionManager(monster, mamConfig)

	return monster, nil
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
		sm.AddKnownInnateSpells(config.InnateSpells)
	} else {
		sm.AddKnownSpells(config.LeveledSpells)
	}

	return sm, nil
}

func initializeActionManager(m *Monster, config *monster_action_manager.MAMConfig) *monster_action_manager.MonsterActionManager {
	mam := monster_action_manager.NewMonsterActionManager(m, m.RollManager, config)
	return mam
}

func (m *Monster) setHP(method core.HPSetMethod, value int) error {
	switch method {
	case core.HPSetValue:
		hp := entity_state_manager.HPValues{
			CurrentHP: value,
			MaxHP:     value,
			TempHP:    0,
			HitDie:    m.HP.Die,
		}
		hpRoll := roll_manager.RollResult{
			DiceRollType:   core.DiceRollHPValueUsed,
			FinalRollValue: value,
			Total:          value,
		}
		events.LogDiceRollEvent(m, &hpRoll, m.EventListener)
		m.EntityState.SetHPValues(hp)

		return nil
	case core.HPSetAverage:
		hp := entity_state_manager.HPValues{
			CurrentHP: m.HP.HPAverage,
			MaxHP:     m.HP.HPAverage,
			TempHP:    0,
			HitDie:    m.HP.Die,
		}
		hpRoll := roll_manager.RollResult{
			DiceRollType:   core.DiceRollHPAvgUsedMonster,
			Die:            m.HP.Die,
			FinalRollValue: m.HP.HPAverage,
			Total:          m.HP.HPAverage,
		}
		events.LogDiceRollEvent(m, &hpRoll, m.EventListener)
		m.EntityState.SetHPValues(hp)

		return nil
	case core.HPSetRoll:
		m.RollManager.RollHP() //TODO: This function needs to be added for monsters
		return nil
	default:
		return fmt.Errorf("invalid HP set method: %s", m)
	}
}

// MakeSavingThrow calculates a saving throw for the given ability and returns the total roll result or an error.
func (m *Monster) MakeSavingThrow(ability core.Ability, targetValue int) (core.RollResult, error) {
	// roll dice, add save
	mod, err := m.GetSavingThrowBonus(ability)
	if err != nil {
		return nil, err
	}

	opts := roll_manager.RollOptions{
		Advantage:         core.RollNormal, // TODO: Will monsters ever have advantage? Features apply this
		Modifier:          mod,
		CriticalThreshold: 0,     // Not relevant
		TreatOnesAsTwos:   false, // Not relevant
		RollType:          core.DiceRollSavingThrow,
		TargetValue:       targetValue,
	}

	res, err := m.RollManager.RollSavingThrow(ability, opts)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Monster) GetEventListener() func(event interface{}) {
	return m.EventListener
}

func (m *Monster) GetName() string {
	return m.Name
}

func (m *Monster) GetAbilityScores() core.AbilityScores {
	return m.AbilityScores
}

func (m *Monster) GetHPStatus() core.HPStatus {
	return m.EntityState.GetHPStatus()
}

func (m *Monster) GetHitDie() core.DiceType { return m.EntityState.GetHitDie() }

func (m *Monster) GetCR() float64 { return m.CR }

func (m *Monster) GetAC() int { return m.AC }

func (m *Monster) GetLevel() interface{} { return m.CR }

func (m *Monster) GetCasterLevel() int { return m.SpellCastingManager.GetCasterLevel() }

//func (m *Monster) GetSpellSaveDC() int { return m.SpellCastingManager.GetSaveDC() }

func (m *Monster) GetSpellSaveDC(ability *core.Ability) (int, error) {
	return m.SpellCastingManager.GetSaveDC(), nil
}

func (m *Monster) IsUnconscious() bool { return m.EntityState.GetIsUnconscious() }

// GetAbilityScore returns the score for the specified ability of the monster. Defaults to 0 if the ability is not found.
func (m *Monster) GetAbilityScore(ability core.Ability) int {
	switch ability {
	case core.AbilityStrength:
		return m.AbilityScores.Strength
	case core.AbilityDexterity:
		return m.AbilityScores.Dexterity
	case core.AbilityConstitution:
		return m.AbilityScores.Constitution
	case core.AbilityIntelligence:
		return m.AbilityScores.Intelligence
	case core.AbilityWisdom:
		return m.AbilityScores.Wisdom
	case core.AbilityCharisma:
		return m.AbilityScores.Charisma
	default:
		return 0
	}
}

func (m *Monster) GetAbilityScoreModifier(ability core.Ability) (int, error) {
	var abilityMod int
	var err error
	switch ability {
	case core.AbilityStrength:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Strength)
	case core.AbilityDexterity:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Dexterity)
	case core.AbilityConstitution:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Constitution)
	case core.AbilityIntelligence:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Intelligence)
	case core.AbilityWisdom:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Wisdom)
	case core.AbilityCharisma:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Charisma)
	default:
		abilityMod = 0
		err = fmt.Errorf("invalid ability provided: %s", ability)
	}
	if err != nil {
		return 0, err
	}
	return abilityMod, nil
}

func (m *Monster) getIsProficientInAbility(ability core.Ability) bool {
	switch ability {
	case core.AbilityStrength:
		return m.AbilityScoreProf.Strength
	case core.AbilityDexterity:
		return m.AbilityScoreProf.Dexterity
	case core.AbilityConstitution:
		return m.AbilityScoreProf.Constitution
	case core.AbilityIntelligence:
		return m.AbilityScoreProf.Intelligence
	case core.AbilityWisdom:
		return m.AbilityScoreProf.Wisdom
	case core.AbilityCharisma:
		return m.AbilityScoreProf.Charisma
	default:
		return false
	}
}

func (m *Monster) GetSavingThrowBonus(ability core.Ability) (int, error) {
	var pb int
	var mod int
	var err error

	mod, err = m.GetAbilityScoreModifier(ability)
	pb, err = core.GetMonsterProficiencyBonus(m.CR)
	if err != nil {
		return 0, err
	}

	if m.getIsProficientInAbility(ability) {
		return pb + mod, nil
	}
	return mod, nil
}

func (m *Monster) IsCharacter() bool { return false }
func (m *Monster) IsMonster() bool   { return true }

var _ core.Entity = &Monster{}
