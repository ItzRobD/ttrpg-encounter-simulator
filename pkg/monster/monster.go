package monster

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"errors"
	"fmt"
	"math/rand/v2"
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

	// Moving hp setup to during simulation
	//err = monster.setHP(monster.HP)
	//if err != nil {
	//	return nil, err
	//}

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
		if hpRoll.Total <= 0 {
			hpRoll.Total = 1
			hpRoll.FinalRollValue = 1
		}
		if err != nil {
			return err
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

func (m *Monster) createAttackRequest(target core.Entity, actionIndex int, actionType core.ActionType, adv core.AdvantageType, simulationOptions *core.SimulationOptions) (*core.AttackRequest, error) {
	if !isValidMonsterActionType(actionType) {
		return nil, fmt.Errorf("invalid action type for monster attack request")
	}

	// TODO: Handle these
	attackOptions := core.AttackOptions{
		Advantage:            adv,
		ShouldApplyDamageMod: true,
		ImprovedCritical:     simulationOptions.UseImprovedCriticals,
	}

	return &core.AttackRequest{
		AttackData:        m.ActionManager.GetAttackDataFromIndex(actionIndex, actionType),
		AttackOptions:     attackOptions,
		SimulationOptions: simulationOptions,
		Target:            target,
	}, nil

}

func (m *Monster) createSpellAttackData(spellChoice core.SpellChoice) (spellcasting_manager.SpellCastData, error) {
	spellBonus := m.SpellCastingManager.GetSpellcastModifierValue()

	return spellcasting_manager.SpellCastData{
		SpellChoice:          spellChoice,
		AttackModifier:       spellBonus,
		SpellcastingModifier: 0, // TODO: Will this ever be used for anything?
	}, nil
}

func (m *Monster) createSpellCastRequest(spellchoice core.SpellChoice, adv core.AdvantageType, simOptions *core.SimulationOptions) (*spellcasting_manager.SpellCastRequest, error) {
	spellcastData, err := m.createSpellAttackData(spellchoice)
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

func isValidMonsterActionType(actionType core.ActionType) bool {
	return actionType == core.ATMonsterAction ||
		actionType == core.ATLegendaryAction ||
		actionType == core.ATMonsterSpecial ||
		actionType == core.ATLegendaryAction ||
		actionType == core.ATMonsterMultiattack
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

func (m *Monster) SetEventListener(listener func(event interface{})) {
	m.EventListener = listener
}

func (m *Monster) GetState() interface{} {
	return m.EntityState
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

func (m *Monster) GetAC() int { return m.AC }

func (m *Monster) GetLevel() float64 { return m.CR }

func (m *Monster) GetHPConfig() core.HPConfig { return m.HP }

func (m *Monster) SetHP(method core.HPSetMethod, value int) error {
	return m.setHP(core.HPConfig{HPSetMethod: method, Value: value})
}

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

func (m *Monster) RollInitiative() (int, error) {
	var err error
	opts := roll_manager.NewRollOptions()
	opts.Modifier, err = m.GetAbilityScoreModifier(core.AbilityDexterity)
	if err != nil {
		return 0, err
	}

	res, err := m.RollManager.RollInitiative(opts)
	if err != nil {
		return 0, err
	}

	m.EntityState.SetInitiative(res.Total)

	return res.Total, nil
}

func (m *Monster) IsCharacter() bool { return false }
func (m *Monster) IsMonster() bool   { return true }

func (m *Monster) GetRNG() *rand.Rand  { return m.RNG }
func (m *Monster) GetID() int          { return m.ID }
func (m *Monster) InitializeHP() error { return m.setHP(m.HP) }
func (m *Monster) IsSpellcaster() bool { return m.MonsterBase.IsSpellcaster }
func (m *Monster) IsHealer() bool      { return m.SpellCastingManager.HasHealingSpells() }
func (m *Monster) GetTargetPriority() core.TargetPriority {
	return m.EntityState.TargetPrioritization
}
func (m *Monster) SetTargetPriority(priority core.TargetPriority) {
	m.EntityState.TargetPrioritization = priority
}
func (m *Monster) ModifyHP(value int, isTemp bool, tempStacking bool) (core.HPModificationResult, error) {
	return m.EntityState.ModifyHP(value, isTemp, tempStacking)
}
func (m *Monster) ChooseSpellByHealingEfficiency(targetValue int) (*core.SpellChoice, error) {
	choice, err := m.SpellCastingManager.GetMostEfficientHealingSpell(targetValue)
	if err != nil {
		return nil, err
	}
	return choice, nil
}
func (m *Monster) ChooseDamageSpellByPriority(p core.SpellPriority) (*core.SpellChoice, error) {
	return m.SpellCastingManager.ChooseSpellByPriority(core.STDamage, p)
}
func (m *Monster) GetHealingSpellCount() int {
	return m.SpellCastingManager.GetHealingSpellCount()
}
func (m *Monster) GetDamageSpellCount() int {
	return m.SpellCastingManager.GetDamageSpellCount()
}

func (m *Monster) UpdateAICombatContext(ctx *core.CombatContext) error {
	m.AI.UpdateCombatContext(ctx)
	return nil
}

func (m *Monster) GetAIRequest(actorID int, t core.AIRequestType) (*core.AIRequest, error) {
	var req *core.AIRequest
	var err error
	switch t {
	case core.AIReqChooseAction:
		req, err = m.AI.createMonsterDamageActionRequest()
		if err != nil {
			return nil, err
		}
	default:
		return req, fmt.Errorf("invalid AI request type: %s", t)
	}

	// TODO: Logging

	req.ActorID = actorID

	return req, nil
}

func (m *Monster) ExecuteAIRequest(req *core.AIRequest) (*core.ActionOutcome, error) {
	switch req.ActionType {
	case core.ATMonsterAction, core.ATMonsterMultiattack, core.ATMonsterSpecial:
		attackReq, err := m.createAttackRequest(req.Target, req.ActionIndex, req.ActionType, req.Advantage, req.SimOptions)
		if err != nil {
			return nil, err
		}

		results, err := m.ActionManager.ProcessAttackRequest(attackReq)
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
		scReq, err := m.createSpellCastRequest(*req.SpellChoice, req.Advantage, req.SimOptions)
		if err != nil {
			return nil, err
		}

		res, err := m.SpellCastingManager.CastSpell(scReq)
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
	default:
		return nil, fmt.Errorf("invalid action type: %s", req.ActionType)
	}

	return nil, errors.New("invalid action type")
}

var _ core.Entity = &Monster{}
