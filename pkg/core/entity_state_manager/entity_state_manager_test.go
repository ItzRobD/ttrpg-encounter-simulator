package entity_state_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/testhelpers"
	"testing"
)

// minimal roll result fake for death saves
type dsRoll struct {
	crit    bool
	nat1    bool
	success bool
}

func (r dsRoll) GetDiceRollType() core.DiceRollType        { return core.DiceRollDeathSavingThrow }
func (r dsRoll) GetNumberOfDice() int                      { return 1 }
func (r dsRoll) GetDiceType() string                       { return "d20" }
func (r dsRoll) GetFinalRollValue() int                    { return 0 }
func (r dsRoll) GetFinalRolls() []int                      { return nil }
func (r dsRoll) GetModifier() int                          { return 0 }
func (r dsRoll) GetTotal() int                             { return 0 }
func (r dsRoll) GetAdvantage() string                      { return "none" }
func (r dsRoll) GetOriginalRolls() []int                   { return nil }
func (r dsRoll) GetRerollEvents() []map[string]interface{} { return nil }
func (r dsRoll) GetWasRerolled() bool                      { return false }
func (r dsRoll) GetIsCritical() bool                       { return r.crit }
func (r dsRoll) GetIsNaturalOne() bool                     { return r.nat1 }
func (r dsRoll) GetIsSuccess() bool                        { return r.success }
func (r dsRoll) GetTargetValue() int                       { return 10 }

func TestNewEntityStateManager_ClampsAndDefaults(t *testing.T) {
	parent := testhelpers.NewEmEntity(5, core.AbilityScores{}, nil)

	// Negative MaxHP should error per implementation
	_, err := NewEntityStateManager(parent, EntityStateConfig{MaxHP: -10})
	if err == nil {
		t.Fatalf("expected error when MaxHP < 0")
	}

	// Valid MaxHP with negative CurrentHP/TempHP should clamp
	cfg := EntityStateConfig{
		CurrentHP:           -5,
		MaxHP:               10,
		TempHP:              -3,
		AttackCount:         -1,
		MaxLegendaryActions: -2,
		Conditions:          nil,
	}
	esm, err := NewEntityStateManager(parent, cfg)
	if err != nil {
		t.Fatalf("NewEntityStateManager error: %v", err)
	}
	if esm.MaxHP != 10 {
		t.Errorf("MaxHP = %d, want 10", esm.MaxHP)
	}
	if esm.CurrentHP != 0 {
		t.Errorf("CurrentHP = %d, want 0 (clamped)", esm.CurrentHP)
	}
	if esm.TempHP != 0 {
		t.Errorf("TempHP = %d, want 0 (clamped)", esm.TempHP)
	}
	if esm.NumberOfAttacks != 0 {
		t.Errorf("NumberOfAttacks = %d, want 0", esm.NumberOfAttacks)
	}
	if esm.LegendaryActionPoints != 0 || esm.LegendaryActionPointsMax != 0 {
		t.Errorf("Legendary points = (%d,%d), want (0,0)", esm.LegendaryActionPoints, esm.LegendaryActionPointsMax)
	}
	if esm.Conditions == nil {
		t.Errorf("Conditions should be initialized")
	}
}

func TestActionEconomyAndRefresh(t *testing.T) {
	// Use 0 legendary so expending all actions makes CanTakeActions false
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(5, core.AbilityScores{}, nil), EntityStateConfig{MaxLegendaryActions: 0})
	if !esm.CanTakeActions() {
		t.Errorf("expected can take actions at start")
	}
	esm.ExpendAction()
	esm.ExpendBonusAction()
	esm.ExpendReaction()
	if esm.CanTakeActions() {
		t.Errorf("expected cannot take actions after expending all (no legendary spent yet)")
	}
	esm.RefreshActions()
	if !esm.CanTakeActions() {
		t.Errorf("expected can take actions after refresh")
	}
	if esm.LegendaryActionPoints != esm.LegendaryActionPointsMax {
		t.Errorf("legendary not reset on refresh")
	}
}

func TestLegendaryActionPoints(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(5, core.AbilityScores{}, nil), EntityStateConfig{MaxLegendaryActions: 3})
	if err := esm.ExpendLegendaryActionPoints(2); err != nil {
		t.Fatalf("expend LAP: %v", err)
	}
	if esm.GetLegendaryActionPoints() != 1 {
		t.Errorf("LAP = %d, want 1", esm.GetLegendaryActionPoints())
	}
	esm.ReplenishLegendaryActionPoints(1)
	if esm.GetLegendaryActionPoints() < esm.LegendaryActionPointsMax {
		t.Errorf("LAP should be >= max after replenish, got %d", esm.GetLegendaryActionPoints())
	}
}

func TestConditions_UnconsciousAddsProne(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{})
	esm.AddCondition(core.ConditionUnconscious)
	if !esm.HasCondition(core.ConditionUnconscious) {
		t.Errorf("expected unconscious condition")
	}
	if !esm.HasCondition(core.ConditionProne) {
		t.Errorf("expected prone when unconscious added")
	}
}

func TestUnconsciousFromHP(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{MaxHP: 10, CurrentHP: 0})
	if !esm.GetIsUnconscious() {
		t.Errorf("expected unconscious at 0 HP")
	}
}

func TestHP_Modify_TempAndHealing(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{MaxHP: 10, CurrentHP: 10, TempHP: 5})
	// Damage that consumes only temp
	res, err := esm.ModifyHP(-3, false, false)
	if err != nil {
		t.Fatalf("ModifyHP damage: %v", err)
	}
	if esm.TempHP != 2 || esm.CurrentHP != 10 {
		t.Errorf("after -3 dmg: hp=%d temp=%d, want 10/2", esm.CurrentHP, esm.TempHP)
	}
	if !res.DidTempDamage || res.DidHPDamage {
		t.Errorf("expected only temp damage flags")
	}

	// Temp HP add without stacking uses max
	res, err = esm.ModifyHP(8, true, false)
	if err != nil {
		t.Fatalf("ModifyHP temp add: %v", err)
	}
	if esm.TempHP != 8 {
		t.Errorf("temp after add max(2,8) = %d, want 8", esm.TempHP)
	}

	// Healing caps at max
	esm.CurrentHP = 7
	res, err = esm.ModifyHP(10, false, false)
	if err != nil {
		t.Fatalf("ModifyHP heal: %v", err)
	}
	if esm.CurrentHP != 10 || !res.DidHealHP {
		t.Errorf("healing cap failed: hp=%d", esm.CurrentHP)
	}

	// isTemp=true with negative should error
	if _, err := esm.ModifyHP(-1, true, false); err == nil {
		t.Errorf("expected error when applying negative temp hp")
	}
}

func TestCheckMassiveDamage(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{MaxHP: 12, CurrentHP: 0})
	// Constructor clamps CurrentHP to >= 0; set negative explicitly to simulate state after outside change
	esm.CurrentHP = -12
	if !esm.CheckMassiveDamage() {
		t.Errorf("expected massive damage true at -MaxHP")
	}
}

func TestRechargeActions(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{})
	esm.AddRechargeAction(0)
	esm.AddRechargeAction(1)
	if !esm.GetRechargeActionStatusAtIndex(0) || !esm.GetRechargeActionStatusAtIndex(1) {
		t.Fatalf("recharge actions should be available after add")
	}
	esm.ExpendRechargeAction(0)
	if esm.GetRechargeActionStatusAtIndex(0) {
		t.Errorf("index 0 should be expended")
	}
	expended := esm.GetExpendedRechargeActionsIndex()
	if len(expended) != 1 || expended[0] != 0 {
		t.Errorf("expected expended [0], got %v", expended)
	}
	esm.RechargeRechargeAction(0)
	esm.ResetAllRechargeActions()
	if !esm.GetRechargeActionStatusAtIndex(0) || !esm.GetRechargeActionStatusAtIndex(1) {
		t.Errorf("reset should make all available")
	}
}

func TestDeathSaves_CriticalRevive(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{MaxHP: 10, CurrentHP: 0})
	if err := esm.ApplyDeathSavingThrowResult(dsRoll{crit: true}); err != nil {
		t.Fatalf("ApplyDeathSavingThrowResult: %v", err)
	}
	if esm.CurrentHP != 1 {
		t.Errorf("crit should revive to 1 HP, got %d", esm.CurrentHP)
	}
	if esm.HasCondition(core.ConditionUnconscious) {
		t.Errorf("revive should remove unconscious condition")
	}
}

func TestModifyHP_KillsMonsterAtZero(t *testing.T) {
	// Monster should die at 0 HP when damage reduces to 0 or below
	monster := testhelpers.NewEmMonster(1, core.AbilityScores{})
	esm, _ := NewEntityStateManager(monster, EntityStateConfig{MaxHP: 10, CurrentHP: 1})
	_, _ = esm.ModifyHP(-5, false, false)
	if !esm.GetIsDead() {
		t.Errorf("monster should be dead after dropping to 0 or below")
	}
}

func TestPreferencesAndInitiativeRoundTrip(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(5, core.AbilityScores{}, nil), EntityStateConfig{})

	// Initiative and bonuses
	esm.SetInitiative(12)
	esm.SetInitiativeBonus(3)
	esm.SetInitiativeAdvantage(core.RollAdvantage)

	if esm.GetInitiative() != 12 {
		t.Errorf("initiative=%d want 12", esm.GetInitiative())
	}
	if esm.GetInitiativeBonus() != 3 {
		t.Errorf("initiative bonus=%d want 3", esm.GetInitiativeBonus())
	}
	if esm.GetInitiativeAdvantage() != core.RollAdvantage {
		t.Errorf("initiative adv=%v want Advantage", esm.GetInitiativeAdvantage())
	}

	// Preferences
	esm.SetActionPreference(core.APPreferMelee)
	esm.SetVersatileWeaponPreference(core.VWPPreferVersatile)
	esm.SetTargetPrioritization(core.PrioritizeLowestHealth)
	esm.SetSpellcastingPriority(core.SPHighestLevel)

	if esm.GetActionPreference() != core.APPreferMelee {
		t.Errorf("action pref mismatch")
	}
	if esm.GetVersatileWeaponPreference() != core.VWPPreferVersatile {
		t.Errorf("versatile pref mismatch")
	}
	if esm.GetTargetPrioritization() != core.PrioritizeLowestHealth {
		t.Errorf("target priority mismatch")
	}
	if esm.GetSpellcastingPriority() != core.SPHighestLevel {
		t.Errorf("spell priority mismatch")
	}
}

func TestConditionsHelpersAndIncapacitation(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{})

	// Add and remove conditions
	esm.AddCondition(core.ConditionPoisoned)
	if !esm.HasCondition(core.ConditionPoisoned) {
		t.Fatalf("expected poisoned")
	}
	esm.RemoveCondition(core.ConditionPoisoned)
	if esm.HasCondition(core.ConditionPoisoned) {
		t.Fatalf("poisoned should be removed")
	}

	// Reset clears all
	esm.AddCondition(core.ConditionGrappled)
	esm.AddCondition(core.ConditionFrightened)
	if len(esm.GetActiveConditions()) == 0 {
		t.Fatalf("expected active conditions")
	}
	esm.ResetConditions()
	if len(esm.GetActiveConditions()) != 0 {
		t.Fatalf("expected no active conditions after reset")
	}

	// Incapacitating conditions disable actions
	incapacitating := []core.Condition{
		core.ConditionIncapacitated,
		core.ConditionStunned,
		core.ConditionParalyzed,
		core.ConditionPetrified,
		core.ConditionUnconscious,
	}
	for _, c := range incapacitating {
		esm.ResetConditions()
		esm.RefreshActions()
		esm.AddCondition(c)
		if esm.CanTakeActions() {
			t.Errorf("CanTakeActions should be false with condition %v", c)
		}
	}

	// GetActiveIncapacitatingConditions returns the set we expect
	esm.ResetConditions()
	esm.AddCondition(core.ConditionStunned)
	esm.AddCondition(core.ConditionPoisoned) // non-incapacitating
	inc := esm.GetActiveIncapacitatingConditions()
	if len(inc) != 1 || inc[0] != core.ConditionStunned {
		t.Errorf("incapacitating list = %v want [Stunned]", inc)
	}
}

func TestHPHelpersAndStatus(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{MaxHP: 20, CurrentHP: 5, TempHP: 3})

	// Hit die set/get
	esm.SetHitDie(core.D10)
	if esm.GetHitDie() != core.D10 {
		t.Errorf("hit die mismatch")
	}

	// Totals & status
	if esm.GetTotalHP() != 8 {
		t.Errorf("total hp=%d want 8", esm.GetTotalHP())
	}
	if esm.GetIsMaxHealth() {
		t.Errorf("should not be max health")
	}

	// ResetHP sets to max and clears temp
	esm.ResetHP()
	if esm.GetCurrentHP() != 20 || esm.GetTempHP() != 0 {
		t.Errorf("after reset: hp=%d temp=%d want 20/0", esm.GetCurrentHP(), esm.GetTempHP())
	}
	if !esm.GetIsMaxHealth() {
		t.Errorf("should be max health after reset")
	}

	// SetHPValues from HPValues and check GetHPStatus HPPct
	hp := HPValues{CurrentHP: 10, MaxHP: 20, TempHP: 5, HitDie: core.D8}
	esm.SetHPValues(hp)
	st := esm.GetHPStatus()
	if st.GetHP() != 10 || st.GetMaxHP() != 20 || st.GetTempHP() != 5 || st.GetHitDie() != core.D8 {
		t.Fatalf("hp status mismatch: %+v", st)
	}
	// 10/20 -> 50%
	if st.GetHPPct() != 50 {
		t.Errorf("HPPct=%d want 50", st.GetHPPct())
	}

	// Getters direct
	if esm.GetCurrentHP() != 10 || esm.GetMaxHP() != 20 || esm.GetTempHP() != 5 {
		t.Errorf("getters mismatch")
	}
}

func TestModifyHP_Edges_TempStackingAndOverflow(t *testing.T) {
	// Start with some temp and HP
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{MaxHP: 10, CurrentHP: 10, TempHP: 3})

	// Temp stacking true: adds
	if _, err := esm.ModifyHP(5, true, true); err != nil {
		t.Fatalf("temp stacking add err: %v", err)
	}
	if esm.TempHP != 8 {
		t.Errorf("temp after stacking add=%d want 8", esm.TempHP)
	}

	// Exact overflow boundary: temp exactly consumed (no HP change)
	if _, err := esm.ModifyHP(-8, false, false); err != nil {
		t.Fatalf("damage err: %v", err)
	}
	if esm.TempHP != 0 || esm.CurrentHP != 10 {
		t.Errorf("after -8 into temp: hp=%d temp=%d want 10/0", esm.CurrentHP, esm.TempHP)
	}

	// Damage below zero HP
	if _, err := esm.ModifyHP(-15, false, false); err != nil {
		t.Fatalf("big damage err: %v", err)
	}
	if esm.CurrentHP >= 0 {
		t.Errorf("hp should be negative or zero after big damage, got %d", esm.CurrentHP)
	}
	if !esm.GetIsUnconscious() {
		t.Errorf("should be unconscious at <=0 HP")
	}
}

func TestDeathSaves_Progressions(t *testing.T) {
	// Three successes -> stable
	esm1, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{MaxHP: 10, CurrentHP: 0})
	for i := 0; i < 3; i++ {
		if err := esm1.ApplyDeathSavingThrowResult(dsRoll{success: true}); err != nil {
			t.Fatalf("apply success: %v", err)
		}
	}
	if !esm1.GetIsStable() || esm1.GetIsDead() {
		t.Errorf("expected stable and not dead; stable=%v dead=%v", esm1.GetIsStable(), esm1.GetIsDead())
	}

	// Natural one (double failure) + one more failure -> death
	esm2, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{MaxHP: 10, CurrentHP: 0})
	if err := esm2.ApplyDeathSavingThrowResult(dsRoll{nat1: true}); err != nil {
		t.Fatalf("apply nat1: %v", err)
	}
	if err := esm2.ApplyDeathSavingThrowResult(dsRoll{success: false}); err != nil {
		t.Fatalf("apply failure: %v", err)
	}
	if !esm2.GetIsDead() {
		t.Errorf("expected dead after nat1 + failure")
	}
}

func TestTakeDamageWhileUnconscious_Progression(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{MaxHP: 10, CurrentHP: 0})
	// two non-crit failures then one more should kill
	esm.TakeDamageWhileUnconscious(false)
	esm.TakeDamageWhileUnconscious(false)
	if esm.GetIsDead() {
		t.Fatalf("should not be dead yet")
	}
	esm.TakeDamageWhileUnconscious(false)
	if !esm.GetIsDead() {
		t.Errorf("expected dead after 3 failures while unconscious")
	}
}

func TestRevive_ErrorWhenAlive(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), EntityStateConfig{MaxHP: 10, CurrentHP: 5})
	if err := esm.Revive(1); err == nil {
		t.Errorf("expected error when reviving an alive entity")
	}
}

func TestLegendaryPoints_ErrorAndRemaining(t *testing.T) {
	esm, _ := NewEntityStateManager(testhelpers.NewEmEntity(5, core.AbilityScores{}, nil), EntityStateConfig{MaxLegendaryActions: 1})
	if !esm.HasLegendaryActionPointsRemaining() {
		t.Fatalf("expected to have points")
	}
	// overspend should error
	if err := esm.ExpendLegendaryActionPoints(2); err == nil {
		t.Errorf("expected error on overspend")
	}
	// spend exactly available
	if err := esm.ExpendLegendaryActionPoints(1); err != nil {
		t.Fatalf("spend err: %v", err)
	}
	if esm.HasLegendaryActionPointsRemaining() {
		t.Errorf("should have no points remaining")
	}
}
