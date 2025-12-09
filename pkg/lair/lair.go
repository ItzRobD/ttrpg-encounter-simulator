package lair

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"fmt"
	"math/rand/v2"
)

// Lair is a minimal actor that represents lair actions acting at initiative 20.
// It implements core.Entity with safe defaults and cannot be targeted/damaged.
type Lair struct {
	name          string
	rng           *rand.Rand
	rollManager   *roll_manager.RollManager
	actionManager *LairActionManager
	ai            *LairAI
	listener      func(event interface{})
	combatCtx     *core.CombatContext
}

func NewLair(name string, rng *rand.Rand) *Lair {
	if name == "" {
		name = "Lair"
	}
	// RNG must be provided by the simulation manager; lair no longer creates its own RNG.
	if rng == nil {
		panic("lair.NewLair requires a non-nil RNG provided by the SimulationManager")
	}
	l := &Lair{name: name, rng: rng}
	l.rollManager = roll_manager.NewRollManager(l, roll_manager.RerollAbilities{})
	l.actionManager = NewLairActionManager(l, l.rollManager)
	l.ai = NewLairAI(l)
	return l
}

// Wiring helpers
func (l *Lair) GetActionManager() *LairActionManager { return l.actionManager }
func (l *Lair) GetAI() *LairAI                       { return l.ai }

// core.Entity minimal implementation
func (l *Lair) IsCharacter() bool        { return false }
func (l *Lair) IsMonster() bool          { return false }
func (l *Lair) GetIsLegendary() bool     { return false }
func (l *Lair) RefreshLegendaryActions() { return }
func (l *Lair) GetClassID() uint8        { return 0 }

func (l *Lair) GetEventListener() func(event interface{})  { return l.listener }
func (l *Lair) SetEventListener(f func(event interface{})) { l.listener = f }

func (l *Lair) IsUnconscious() bool { return false }
func (l *Lair) IsDead() bool        { return false }

func (l *Lair) GetHPStatus() core.HPStatus { return core.NewHPStatusStub() }

func (l *Lair) GetName() string { return l.name }

func (l *Lair) GetAbilityScores() core.AbilityScores                { return core.AbilityScores{} }
func (l *Lair) GetAbilityScore(a core.Ability) int                  { return 10 }
func (l *Lair) GetAbilityScoreModifier(a core.Ability) (int, error) { return 0, nil }
func (l *Lair) GetSavingThrowBonus(a core.Ability) (int, error)     { return 0, nil }

func (l *Lair) GetHitDie() core.DiceType     { return core.D8 }
func (l *Lair) GetLevel() float64            { return 0 }
func (l *Lair) GetCasterLevel() int          { return 0 }
func (l *Lair) GetHPConfig() core.HPConfig   { return core.HPConfig{} }
func (l *Lair) GetState() interface{}        { return nil }
func (l *Lair) InitializeHP() error          { return nil }
func (l *Lair) RollInitiative() (int, error) { return 20, nil }
func (l *Lair) GetAC() int                   { return 10 }

func (l *Lair) IsSpellcaster() bool       { return false }
func (l *Lair) IsHealer() bool            { return false }
func (l *Lair) GetHealingSpellCount() int { return 0 }
func (l *Lair) GetDamageSpellCount() int  { return 0 }
func (l *Lair) ChooseSpellByHealingEfficiency(targetValue int) (*core.SpellChoice, error) {
	return nil, fmt.Errorf("lair cannot cast spells")
}
func (l *Lair) ChooseDamageSpellByPriority(p core.SpellPriority) (*core.SpellChoice, error) {
	return nil, fmt.Errorf("lair cannot cast spells")
}

func (l *Lair) GetRNG() *rand.Rand { return l.rng }

func (l *Lair) GetTargetPriority() core.TargetPriority  { return core.PrioritizeLowestMaxHP }
func (l *Lair) SetTargetPriority(p core.TargetPriority) {}

func (l *Lair) MakeSavingThrow(ability core.Ability, targetValue int) (core.RollResult, error) {
	return nil, fmt.Errorf("lair cannot make saving throws")
}

func (l *Lair) GetSpellSaveDC(ability *core.Ability) (int, error) {
	return 0, fmt.Errorf("lair has no spell save DC")
}

func (l *Lair) ModifyHP(value int, isTemp bool, tempStacking bool) (core.HPModificationResult, error) {
	return nil, fmt.Errorf("lair cannot be damaged or healed")
}

func (l *Lair) GetConditions() core.EntityConditions { return core.NewEntityConditions() }

// AI/context
func (l *Lair) UpdateAICombatContext(ctx *core.CombatContext) error {
	l.combatCtx = ctx
	l.ai.UpdateCombatContext(ctx)
	return nil
}

func (l *Lair) GetAIRequest(actorID int, t core.AIRequestType) (*core.AIRequest, error) {
	switch t {
	case core.AIReqNormalAction:
		req, err := l.ai.BuildLairActionRequest()
		if err != nil {
			return nil, err
		}
		req.ActorID = actorID
		return req, nil
	default:
		return nil, fmt.Errorf("invalid AI request type for lair: %v", t)
	}
}

func (l *Lair) ExecuteAIRequest(req *core.AIRequest) (*core.ActionOutcome, error) {
	if req.ActionType != core.ATLairAction {
		return nil, fmt.Errorf("lair only executes lair actions")
	}
	// Advanced execution supports attack/DC modes and AOE; the manager will apply
	// recharge logic and compute effects across all affected targets.
	_, effects, err := l.actionManager.ExecuteAdvanced(req.ActionIndex, req.Target)
	if err != nil {
		return nil, err
	}

	return &core.ActionOutcome{
		ActionType: req.ActionType,
		TargetID:   req.TargetID,
		ActorID:    req.ActorID,
		Effects:    effects,
		Success:    len(effects) > 0,
	}, nil
}

func (l *Lair) ProcessTurn(actorID int, turnType core.TurnType) (*core.TurnResult, *core.AIRequest, error) {
	if turnType != core.TurnTypeNormal {
		return nil, nil, fmt.Errorf("lair only acts on normal turns")
	}
	res := &core.TurnResult{TurnStatuses: make(map[core.TurnStatus]bool)}
	res.Conditions = nil
	// If no actions configured, skip turn
	if l.actionManager == nil || len(l.actionManager.Actions) == 0 {
		res.TurnStatuses[core.TurnIncapacitated] = true
		return res, nil, nil
	}

	aiReq, err := l.GetAIRequest(actorID, core.AIReqNormalAction)
	if err != nil {
		return nil, nil, err
	}
	res.TurnStatuses[core.TurnActionReady] = true
	return res, aiReq, nil
}

func (l *Lair) CanTakeActions() bool { return true }
