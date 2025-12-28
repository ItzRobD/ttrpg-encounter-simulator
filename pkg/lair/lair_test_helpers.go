package lair

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"math/rand/v2"
)

// testRollResult minimally implements core.RollResult for saving throw stubs.
type testRollResult struct {
	diceType    core.DiceRollType
	final       int
	rolls       []int
	mod         int
	total       int
	adv         string
	orig        []int
	rerolled    bool
	crit        bool
	nat1        bool
	success     bool
	targetValue int
}

func (r testRollResult) GetDiceRollType() core.DiceRollType        { return r.diceType }
func (r testRollResult) GetNumberOfDice() int                      { return len(r.rolls) }
func (r testRollResult) GetDiceType() string                       { return "d20" }
func (r testRollResult) GetFinalRollValue() int                    { return r.final }
func (r testRollResult) GetFinalRolls() []int                      { return r.rolls }
func (r testRollResult) GetModifier() int                          { return r.mod }
func (r testRollResult) GetTotal() int                             { return r.total }
func (r testRollResult) GetAdvantage() string                      { return r.adv }
func (r testRollResult) GetOriginalRolls() []int                   { return r.orig }
func (r testRollResult) GetRerollEvents() []map[string]interface{} { return nil }
func (r testRollResult) GetWasRerolled() bool                      { return r.rerolled }
func (r testRollResult) GetIsCritical() bool                       { return r.crit }
func (r testRollResult) GetIsNaturalOne() bool                     { return r.nat1 }
func (r testRollResult) GetIsSuccess() bool                        { return r.success }
func (r testRollResult) GetTargetValue() int                       { return r.targetValue }

// testEntity provides a minimal core.Entity for lair tests.
type testEntity struct {
	name      string
	isChar    bool
	isMon     bool
	ac        int
	succeedST bool
	rng       *rand.Rand
}

func newChar(name string, ac int, succeedST bool) *testEntity {
	return &testEntity{name: name, isChar: true, isMon: false, ac: ac, succeedST: succeedST, rng: rand.New(rand.NewPCG(3, 4))}
}
func newMon(name string, ac int, succeedST bool) *testEntity {
	return &testEntity{name: name, isChar: false, isMon: true, ac: ac, succeedST: succeedST, rng: rand.New(rand.NewPCG(5, 6))}
}

// core.Entity methods (only what's needed by lair code paths)
func (e *testEntity) GetClassID() uint8                         { return 0 }
func (e *testEntity) IsDead() bool                              { return false }
func (e *testEntity) IsUnconscious() bool                       { return false }
func (e *testEntity) GetHPStatus() core.HPStatus                { return core.NewHPStatusStub() }
func (e *testEntity) GetName() string                           { return e.name }
func (e *testEntity) GetAC() int                                { return e.ac }
func (e *testEntity) GetEventListener() func(event interface{}) { return nil }
func (e *testEntity) SetEventListener(func(event interface{}))  {}
func (e *testEntity) GetLevel() float64                         { return 0 }
func (e *testEntity) GetHitDie() core.DiceType                  { return core.D8 }
func (e *testEntity) GetCasterLevel() int                       { return 0 }
func (e *testEntity) MakeSavingThrow(core.Ability, int, bool, core.DamageType) (core.RollResult, error) {
	// Return a deterministic pass/fail without rolling
	if e.succeedST {
		return testRollResult{diceType: core.DiceRollSavingThrow, final: 10, rolls: []int{10}, total: 20, success: true}, nil
	}
	return testRollResult{diceType: core.DiceRollSavingThrow, final: 1, rolls: []int{1}, total: 1, success: false}, nil
}
func (e *testEntity) GetSpellSaveDC(*core.Ability) (int, error)         { return 0, nil }
func (e *testEntity) GetAbilityScores() core.AbilityScores              { return core.AbilityScores{} }
func (e *testEntity) GetAbilityScore(core.Ability) int                  { return 10 }
func (e *testEntity) GetAbilityScoreModifier(core.Ability) (int, error) { return 0, nil }
func (e *testEntity) GetSavingThrowBonus(core.Ability) (int, error)     { return 0, nil }
func (e *testEntity) IsCharacter() bool                                 { return e.isChar }
func (e *testEntity) IsMonster() bool                                   { return e.isMon }
func (e *testEntity) GetIsLegendary() bool                              { return false }
func (e *testEntity) GetHPConfig() core.HPConfig                        { return core.HPConfig{} }
func (e *testEntity) GetState() interface{}                             { return nil }
func (e *testEntity) RollInitiative() (int, error)                      { return 10, nil }
func (e *testEntity) InitializeHP() error                               { return nil }
func (e *testEntity) IsSpellcaster() bool                               { return false }
func (e *testEntity) IsHealer() bool                                    { return false }
func (e *testEntity) GetTargetPriority() core.TargetPriority            { return core.PrioritizeLowestMaxHP }
func (e *testEntity) SetTargetPriority(core.TargetPriority)             {}
func (e *testEntity) ChooseSpellByHealingEfficiency(int) (*core.SpellChoice, error) {
	return nil, nil
}
func (e *testEntity) ChooseDamageSpellByPriority(core.SpellPriority) (*core.SpellChoice, error) {
	return nil, nil
}
func (e *testEntity) GetHealingSpellCount() int                                     { return 0 }
func (e *testEntity) GetDamageSpellCount() int                                      { return 0 }
func (e *testEntity) GetRNG() *rand.Rand                                            { return e.rng }
func (e *testEntity) GetAIRequest(int, core.AIRequestType) (*core.AIRequest, error) { return nil, nil }
func (e *testEntity) ExecuteAIRequest(*core.AIRequest) (*core.ActionOutcome, error) { return nil, nil }
func (e *testEntity) UpdateAICombatContext(*core.CombatContext) error               { return nil }
func (e *testEntity) ModifyHP(int, bool, bool) (core.HPModificationResult, error)   { return nil, nil }
func (e *testEntity) RefreshLegendaryActions()                                      {}
func (e *testEntity) CanTakeActions() bool                                          { return true }
func (e *testEntity) ProcessTurn(int, core.TurnType) (*core.TurnResult, *core.AIRequest, error) {
	return &core.TurnResult{TurnStatuses: map[core.TurnStatus]bool{core.TurnActionReady: true}}, &core.AIRequest{}, nil
}
func (e *testEntity) GetConditions() core.EntityConditions { return core.NewEntityConditions() }

func (e *testEntity) GetType() string {
	if e.isMon {
		return "Dragon"
	}
	return "Humanoid"
}
