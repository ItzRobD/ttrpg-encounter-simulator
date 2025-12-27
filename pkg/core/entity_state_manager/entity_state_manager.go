package entity_state_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"fmt"
	"sort"
)

type HPModificationResult struct {
	ModificationValue int
	OriginalHP        int
	OriginalTempHP    int
	NewHP             int
	NewTempHP         int
	DidHealHP         bool
	DidHealTempHP     bool
	DidTempDamage     bool
	DidHPDamage       bool
	IsUnconscious     bool
	IsMaxHealth       bool
}

func (hpmr HPModificationResult) GetModificationValue() int { return hpmr.ModificationValue }
func (hpmr HPModificationResult) GetOriginalHP() int        { return hpmr.OriginalHP }
func (hpmr HPModificationResult) GetOriginalTempHP() int    { return hpmr.OriginalTempHP }
func (hpmr HPModificationResult) GetNewHP() int             { return hpmr.NewHP }
func (hpmr HPModificationResult) GetNewTempHP() int         { return hpmr.NewTempHP }
func (hpmr HPModificationResult) GetDidHealHP() bool        { return hpmr.DidHealHP }
func (hpmr HPModificationResult) GetDidHealTempHP() bool    { return hpmr.DidHealTempHP }
func (hpmr HPModificationResult) GetDidTempDamage() bool    { return hpmr.DidTempDamage }
func (hpmr HPModificationResult) GetDidHPDamage() bool      { return hpmr.DidHPDamage }
func (hpmr HPModificationResult) GetIsUnconscious() bool    { return hpmr.IsUnconscious }
func (hpmr HPModificationResult) GetIsMaxHealth() bool      { return hpmr.IsMaxHealth }

type EntityStateConfig struct {
	CurrentHP            int
	MaxHP                int
	TempHP               int
	MaxLegendaryActions  int
	AttackCount          int
	Conditions           core.EntityConditions
	ActionPreference     core.ActionPreference
	VersatilePreference  core.VersatileWeaponPreference
	TargetPrioritization core.TargetPriority
	SpellcastingPriority core.SpellPriority
	InitiativeAdvantage  core.AdvantageType
	Resistances          core.DamageResistances
	InitiativeBonus      int
	// Class specifics variables
	BarbarianRelentlessUses  int
	BarbarianIsRaging        bool
	FighterIndomitableUses   int
	PaladinLayingOnHandsPool int
}

type EntityStateManager struct {
	Parent core.Entity
	// HP Management
	CurrentHP int
	MaxHP     int
	TempHP    int
	HitDie    core.DiceType

	// Action Economy
	HasUsedAction            bool
	HasUsedBonusAction       bool
	HasUsedReaction          bool
	LegendaryActionPoints    int
	LegendaryActionPointsMax int
	NumberOfAttacks          int
	RechargeActionStatus     map[int]bool // Key: Action index; Value: IsAvailable
	DBBreathWeaponUsed       bool

	// Conditions
	Conditions            core.EntityConditions
	DeathSaves            core.DeathSaves
	IsStable              bool
	IsDead                bool
	IsRecklesslyAttacking bool

	Initiative int

	// Preferences
	ActionPreference          core.ActionPreference
	VersatileWeaponPreference core.VersatileWeaponPreference
	TargetPrioritization      core.TargetPriority
	SpellcastingPriority      core.SpellPriority

	// Bonuses
	InitiativeAdvantage  core.AdvantageType
	InitiativeBonus      int
	Resistances          core.DamageResistances
	SavingThrowAdvantage map[core.Ability]core.AdvantageType

	// Class specific variables
	BarbarianHasRelentlessRage bool
	BarbarianRelentlessUses    int
	BarbarianIsRaging          bool
	FighterIndomitableUses     int
	PaladinLayingOnHandsPool   int

	// Race specific variables
	HalfOrcHasSavageAttacks          bool
	HalfOrcHasRelentlessEnduranceUse bool
}

func (esm *EntityStateManager) ExpendAction() {
	esm.HasUsedAction = true
}

func (esm *EntityStateManager) ExpendBonusAction() {
	esm.HasUsedBonusAction = true
}

func (esm *EntityStateManager) ExpendReaction() {
	esm.HasUsedReaction = true
}

func (esm *EntityStateManager) ReplenishAction() {
	esm.HasUsedAction = false
}

func (esm *EntityStateManager) ReplenishBonusAction() {
	esm.HasUsedBonusAction = false
}

func (esm *EntityStateManager) ReplenishReaction() {
	esm.HasUsedReaction = false
}

func (esm *EntityStateManager) RefreshActions() {
	esm.HasUsedAction = false
	esm.HasUsedBonusAction = false
	esm.HasUsedReaction = false
	esm.LegendaryActionPoints = esm.LegendaryActionPointsMax
}

func (esm *EntityStateManager) CanTakeActions() bool {
	// Conditions that prevent ALL actions
	if esm.Conditions.Has(core.ConditionIncapacitated) ||
		esm.Conditions.Has(core.ConditionStunned) ||
		esm.Conditions.Has(core.ConditionParalyzed) ||
		esm.Conditions.Has(core.ConditionPetrified) ||
		esm.Conditions.Has(core.ConditionUnconscious) ||
		esm.IsDead {
		return false
	}

	// Check if any action economy is available
	return !esm.HasUsedAction || !esm.HasUsedBonusAction || esm.LegendaryActionPoints > 0
}

func (esm *EntityStateManager) ExpendLegendaryActionPoints(value int) error {
	if value > esm.LegendaryActionPoints {
		return fmt.Errorf("cannot expend more legendary action points than available")
	}
	esm.LegendaryActionPoints -= value
	return nil
}

func (esm *EntityStateManager) ReplenishLegendaryActionPoints(value int) {
	esm.LegendaryActionPoints = max(esm.LegendaryActionPoints+value, esm.LegendaryActionPointsMax)
}

func (esm *EntityStateManager) GetLegendaryActionPoints() int {
	return esm.LegendaryActionPoints
}

func (esm *EntityStateManager) HasLegendaryActionPointsRemaining() bool {
	return esm.LegendaryActionPoints > 0
}

func (esm *EntityStateManager) GetNumberOfAttacks() int {
	return esm.NumberOfAttacks
}

func (esm *EntityStateManager) SetNumberOfExtraAttacks(value int) {
	esm.NumberOfAttacks = value
}

func (esm *EntityStateManager) SetActionPreference(p core.ActionPreference) {
	esm.ActionPreference = p
}

func (esm *EntityStateManager) GetActionPreference() core.ActionPreference {
	return esm.ActionPreference
}

func (esm *EntityStateManager) SetVersatileWeaponPreference(p core.VersatileWeaponPreference) {
	esm.VersatileWeaponPreference = p
}

func (esm *EntityStateManager) GetVersatileWeaponPreference() core.VersatileWeaponPreference {
	return esm.VersatileWeaponPreference
}

func (esm *EntityStateManager) SetTargetPrioritization(p core.TargetPriority) {
	esm.TargetPrioritization = p
}

func (esm *EntityStateManager) GetTargetPrioritization() core.TargetPriority {
	return esm.TargetPrioritization
}

func (esm *EntityStateManager) SetSpellcastingPriority(p core.SpellPriority) {
	esm.SpellcastingPriority = p
}

func (esm *EntityStateManager) GetSpellcastingPriority() core.SpellPriority {
	return esm.SpellcastingPriority
}

func (esm *EntityStateManager) SetInitiativeAdvantage(a core.AdvantageType) {
	esm.InitiativeAdvantage = a
}

func (esm *EntityStateManager) SetInitiative(value int) { esm.Initiative = value }

func (esm *EntityStateManager) GetInitiative() int { return esm.Initiative }

func (esm *EntityStateManager) GetInitiativeAdvantage() core.AdvantageType {
	return esm.InitiativeAdvantage
}

func (esm *EntityStateManager) SetInitiativeBonus(b int) {
	esm.InitiativeBonus = b
}

func (esm *EntityStateManager) GetInitiativeBonus() int {
	return esm.InitiativeBonus
}

// Conditions functions

func (esm *EntityStateManager) AddCondition(c core.Condition) {
	// Special handling for unconscious: also add prone condition
	if c == core.ConditionUnconscious {
		esm.Conditions.Add(core.ConditionUnconscious)
		esm.Conditions.Add(core.ConditionProne)
	} else {
		esm.Conditions.Add(c)
	}
}

func (esm *EntityStateManager) RemoveCondition(c core.Condition) {
	esm.Conditions.Remove(c)
}

func (esm *EntityStateManager) HasCondition(c core.Condition) bool {
	return esm.Conditions.Has(c)
}

func (esm *EntityStateManager) GetConditions() core.EntityConditions {
	return esm.Conditions
}

func (esm *EntityStateManager) GetActiveConditions() []core.Condition {
	return esm.Conditions.GetActive()
}

func (esm *EntityStateManager) ResetConditions() {
	esm.Conditions.Clear()
}

func (esm *EntityStateManager) SetUnconscious(isUnconscious bool) {
	if isUnconscious {
		// Directly add conditions to avoid circular call
		esm.Conditions.Add(core.ConditionUnconscious)
		esm.Conditions.Add(core.ConditionProne)
	} else {
		esm.Conditions.Remove(core.ConditionUnconscious)
	}
}

func (esm *EntityStateManager) GetActiveIncapacitatingConditions() []core.Condition {
	var incapacitating []core.Condition

	incapacitatingList := []core.Condition{
		core.ConditionIncapacitated,
		core.ConditionStunned,
		core.ConditionParalyzed,
		core.ConditionPetrified,
		core.ConditionUnconscious,
	}

	for _, condition := range incapacitatingList {
		if esm.Conditions.Has(condition) {
			incapacitating = append(incapacitating, condition)
		}
	}

	return incapacitating
}

// GetIsUnconscious determines if the entity is unconscious based on its conditions or current health points.
func (esm *EntityStateManager) GetIsUnconscious() bool {
	return esm.Conditions.Has(core.ConditionUnconscious) || esm.CurrentHP <= 0
}

func (esm *EntityStateManager) GetIsStable() bool {
	return esm.IsStable
}

func (esm *EntityStateManager) GetIsDead() bool { return esm.IsDead }

// HP Functions

func (esm *EntityStateManager) ResetHP() {
	esm.CurrentHP = esm.MaxHP
	esm.TempHP = 0
}

func (esm *EntityStateManager) SetHPValues(hp HPValues) {
	esm.CurrentHP = hp.GetHP()
	esm.MaxHP = hp.GetMaxHP()
	esm.TempHP = hp.GetTempHP()
	esm.HitDie = hp.GetHitDie()
}

func (esm *EntityStateManager) GetHitDie() core.DiceType {
	return esm.HitDie
}

func (esm *EntityStateManager) SetHitDie(d core.DiceType) { esm.HitDie = d }

func (esm *EntityStateManager) GetCurrentHP() int {
	return esm.CurrentHP
}

func (esm *EntityStateManager) GetMaxHP() int {
	return esm.MaxHP
}

func (esm *EntityStateManager) GetTempHP() int {
	return esm.TempHP
}

func (esm *EntityStateManager) GetIsMaxHealth() bool {
	return esm.CurrentHP == esm.MaxHP
}

func (esm *EntityStateManager) GetTotalHP() int {
	return esm.CurrentHP + esm.TempHP
}

func (esm *EntityStateManager) ModifyHP(value int, isTemp bool, tempStacking bool) (HPModificationResult, error) {
	res := HPModificationResult{
		ModificationValue: value,
		OriginalHP:        esm.CurrentHP,
		OriginalTempHP:    esm.TempHP,
		NewHP:             esm.CurrentHP,
		NewTempHP:         esm.TempHP,
		DidHealHP:         false,
		DidHealTempHP:     false,
		DidTempDamage:     false,
		DidHPDamage:       false,
		IsUnconscious:     false,
		IsMaxHealth:       false,
	}
	if isTemp { // This should only be true if we are adding to temp hp
		if value < 0 {
			res.DidHealHP = res.NewHP > res.OriginalHP
			res.DidHealTempHP = res.NewTempHP > res.OriginalTempHP
			res.DidTempDamage = res.NewTempHP < res.OriginalTempHP
			res.DidHPDamage = res.NewHP < res.OriginalHP
			res.IsUnconscious = esm.CurrentHP == 0
			res.IsMaxHealth = esm.CurrentHP == esm.MaxHP
			// This should not happen logically, return
			return res, fmt.Errorf("cannot specifically target temp hp with a negative value")
		} else {
			if tempStacking {
				esm.TempHP += value
				res.NewTempHP = esm.TempHP
			} else {
				esm.TempHP = max(esm.TempHP, value)
			}
		}
	} else {
		if value < 0 {
			dmg := -value // positive magnitude
			if esm.TempHP > 0 {
				if esm.TempHP >= dmg {
					esm.TempHP -= dmg // exact or partial absorption
				} else {
					overflow := dmg - esm.TempHP
					esm.TempHP = 0
					esm.CurrentHP -= overflow
				}
			} else {
				esm.CurrentHP -= dmg
			}
		} else { // Healing
			esm.CurrentHP = min(esm.CurrentHP+value, esm.MaxHP)
		}
		res.NewHP = esm.CurrentHP
		res.NewTempHP = esm.TempHP
	}
	res.DidHealHP = res.NewHP > res.OriginalHP
	res.DidHealTempHP = res.NewTempHP > res.OriginalTempHP
	res.DidTempDamage = res.NewTempHP < res.OriginalTempHP
	res.DidHPDamage = res.NewHP < res.OriginalHP
	res.IsUnconscious = esm.CurrentHP <= 0
	res.IsMaxHealth = esm.CurrentHP == esm.MaxHP

	// Handle 0 HP logic
	if res.IsUnconscious {
		// Monsters die instantly at 0 HP (standard D&D 5e rules)
		// Player characters go unconscious and make death saves
		if esm.Parent.IsMonster() {
			esm.Kill()
		} else {
			// Relentless Endurance (Half-Orc)
			// Trigger only if we were not already at 0 HP and we are not killed outright (massive damage)
			if esm.HalfOrcHasRelentlessEnduranceUse && res.OriginalHP > 0 && !esm.CheckMassiveDamage() {
				esm.SetUnconscious(false)
				esm.CurrentHP = 1
				esm.HalfOrcHasRelentlessEnduranceUse = false
				res.IsUnconscious = false
				res.NewHP = esm.CurrentHP
				events.LogCombatEventMessage(esm.Parent, "Relentless Endurance expended. New HP set to 1.", esm.Parent.GetEventListener())
			} else if esm.BarbarianHasRelentlessRage && esm.BarbarianIsRaging && res.OriginalHP > 0 && !esm.CheckMassiveDamage() {
				// Barbarian Relentless Rage
				// Make DC 10 + (uses * 5) Con saving throw
				dc := 10 + (esm.GetBarbarianRelentlessUses() * 5)
				saveResult, err := esm.Parent.MakeSavingThrow(core.AbilityConstitution, dc, false, core.DamageNone)
				if err != nil {
					return res, err
				}
				if saveResult.GetIsSuccess() {
					esm.SetUnconscious(false)
					esm.CurrentHP = 1
					res.IsUnconscious = false
					res.NewHP = esm.CurrentHP
					esm.IncrementBarbarianRelentlessUses()
					events.LogCombatEventMessage(esm.Parent, "Relentless Rage expended. New HP set to 1.", esm.Parent.GetEventListener())
				}
			} else {
				esm.SetUnconscious(true)
			}
		}
	} else if res.OriginalHP <= 0 && res.NewHP > 0 {
		// Character was healed from 0 or stable
		esm.SetUnconscious(false)
		esm.DeathSaves.Reset()
		esm.IsStable = false
	}

	return res, nil
}

// CheckMassiveDamage determines if the entity has taken damage exceeding its maximum HP in the negative range.
// Returns true if the entity's current HP is less than or equal to the negative value of its maximum HP.
func (esm *EntityStateManager) CheckMassiveDamage() bool {
	if esm.CurrentHP <= -esm.MaxHP {
		return true
	}
	return false
}

// Recharge Actions

// GetRechargeActionStatus returns the current recharge status for all actions as a map where keys are action indexes.
func (esm *EntityStateManager) GetRechargeActionStatus() map[int]bool {
	return esm.RechargeActionStatus
}

// GetRechargeActionStatusAtIndex checks if the recharge action at the specified index is available.
func (esm *EntityStateManager) GetRechargeActionStatusAtIndex(index int) bool {
	return esm.RechargeActionStatus[index]
}

// GetExpendedRechargeActionsIndex returns a list of indexes for recharge actions that are currently expended (unavailable).
func (esm *EntityStateManager) GetExpendedRechargeActionsIndex() []int {
	var result []int
	for i, v := range esm.RechargeActionStatus {
		if !v {
			result = append(result, i)
		}
	}
	// Sort to ensure deterministic processing order
	sort.Ints(result)
	return result
}

// ExpendRechargeAction sets the recharge action at the specified index to unavailable (false).
func (esm *EntityStateManager) ExpendRechargeAction(index int) {
	esm.RechargeActionStatus[index] = false
}

// RechargeRechargeAction sets the recharge action at the specified index to available (true).
func (esm *EntityStateManager) RechargeRechargeAction(index int) {
	esm.RechargeActionStatus[index] = true
}

// ResetAllRechargeActions resets all recharge actions to available by setting their status to true.
func (esm *EntityStateManager) ResetAllRechargeActions() {
	for i := 0; i < len(esm.RechargeActionStatus); i++ {
		esm.RechargeActionStatus[i] = true
	}
}

func (esm *EntityStateManager) SetDBBreathWeaponUsed(val bool) {
	esm.DBBreathWeaponUsed = val
}

func (esm *EntityStateManager) GetDBBreathWeaponUsed() bool {
	return esm.DBBreathWeaponUsed
}

func (esm *EntityStateManager) AddRechargeAction(index int) {
	if esm.RechargeActionStatus == nil {
		esm.RechargeActionStatus = make(map[int]bool)
	}
	esm.RechargeActionStatus[index] = true
}

// Death Saves

// ApplyDeathSavingThrowResult processes the result of a death saving throw and updates the entity's state accordingly.
func (esm *EntityStateManager) ApplyDeathSavingThrowResult(result core.RollResult) error {
	if result.GetDiceRollType() != core.DiceRollDeathSavingThrow {
		return fmt.Errorf("non-Kill saving throw result passed to ApplyDeathSavingThrowResult")
	}

	if result.GetIsCritical() {
		// Character revives with 1hp
		esm.Revive(1)
	} else if result.GetIsNaturalOne() {
		// Two Death Failures
		esm.DeathSaves.AddFailure(true)
	} else {
		if result.GetIsSuccess() {
			esm.DeathSaves.AddSuccess()
		} else {
			esm.DeathSaves.AddFailure(false)
		}
	}

	status := esm.DeathSaves.Evaluate()
	switch status {
	case core.DeathSaveSuccess:
		esm.SetStable(true)
	case core.DeathSaveFailure:
		esm.Kill()
	case core.DeathSaveNone:
		break
	default:
		return fmt.Errorf("unknown Kill save status: %v", status)
	}

	return nil
}

func (esm *EntityStateManager) TakeDamageWhileUnconscious(isCrit bool) {
	if isCrit {
		esm.DeathSaves.AddFailure(true)
	} else {
		esm.DeathSaves.AddFailure(false)
	}

	if esm.DeathSaves.Evaluate() == core.DeathSaveFailure {
		esm.Kill()
	}
}

func (esm *EntityStateManager) SetStable(stable bool) {
	esm.IsStable = stable
	esm.DeathSaves.Reset()
}

// Revive attempts to revive an entity by setting its HP and resetting relevant states if it is currently not alive.
// Returns an error if the entity is already alive.
func (esm *EntityStateManager) Revive(hpValue int) error {
	if esm.CurrentHP > 0 {
		return fmt.Errorf("cannot revive an entity that is already alive")
	}

	esm.CurrentHP = hpValue
	esm.IsStable = false
	esm.DeathSaves.Reset()
	esm.RemoveCondition(core.ConditionUnconscious)

	return nil
}

func (esm *EntityStateManager) Kill() {
	esm.IsDead = true
	esm.IsStable = false
	esm.Conditions.Clear()
}

func (esm *EntityStateManager) GetSavingThrowAdvantage(ability core.Ability) core.AdvantageType {
	return esm.SavingThrowAdvantage[ability]
}

func (esm *EntityStateManager) SetHasSavingThrowAdvantage(ability core.Ability, adv core.AdvantageType) {
	esm.SavingThrowAdvantage[ability] = adv
}

// Resistances

func (esm *EntityStateManager) AddResistance(dt core.DamageType, rt core.ResistanceType, rb []core.ResistBreaker) {
	// Fallback if resistances is not initialized
	if esm.Resistances == nil {
		esm.Resistances = core.NewDamageResistances()
	}

	esm.Resistances.SetResistance(dt, rt, rb)
}

func (esm *EntityStateManager) RemoveResistance(dt core.DamageType) error {
	if esm.Resistances == nil {
		return fmt.Errorf("resistances not initialized")
	}

	esm.Resistances.ResetResistance(dt)
	return nil
}

func (esm *EntityStateManager) GetResistances() core.DamageResistances {
	return esm.Resistances
}

func (esm *EntityStateManager) GetResistance(dt core.DamageType) (core.DamageResistance, error) {
	if esm.Resistances == nil {
		return core.DamageResistance{}, fmt.Errorf("resistances not initialized for parent: %v", esm.Parent.GetName())
	}
	return esm.Resistances.GetResistance(dt), nil
}

// Class-specific functions
func (esm *EntityStateManager) ResetBarbarianRelentlessUses() {
	esm.BarbarianRelentlessUses = 0
}

func (esm *EntityStateManager) GetBarbarianRelentlessUses() int {
	return esm.BarbarianRelentlessUses
}

func (esm *EntityStateManager) SetBarbarianIsRaging(val bool) { esm.BarbarianIsRaging = val }

func (esm *EntityStateManager) GetBarbarianIsRaging() bool { return esm.BarbarianIsRaging }

func (esm *EntityStateManager) SetIsRecklesslyAttacking(val bool) {
	esm.IsRecklesslyAttacking = val
}

func (esm *EntityStateManager) GetIsRecklesslyAttacking() bool {
	return esm.IsRecklesslyAttacking
}

func (esm *EntityStateManager) SetBarbarianRelentlessRage(val bool) {
	esm.BarbarianHasRelentlessRage = val
}

func (esm *EntityStateManager) IncrementBarbarianRelentlessUses() {
	esm.BarbarianRelentlessUses++
}

func (esm *EntityStateManager) SetFighterIndomitableUses(val int) {
	esm.FighterIndomitableUses = val
}

func (esm *EntityStateManager) GetFighterIndomitableUses() int {
	return esm.FighterIndomitableUses
}

func (esm *EntityStateManager) SetPaladinLayingOnHandsPool(val int) {
	esm.PaladinLayingOnHandsPool = val
}

func (esm *EntityStateManager) ExpendFighterIndomitableUses() {
	esm.FighterIndomitableUses--
}

func (esm *EntityStateManager) GetPaladinLayingOnHandsPool() int {
	return esm.PaladinLayingOnHandsPool
}

// ModifyPaladinLayingOnHandsPool adjusts the Paladin's Lay on Hands pool by the specified value.
// Negative value to expend points
func (esm *EntityStateManager) ModifyPaladinLayingOnHandsPool(val int) {
	esm.PaladinLayingOnHandsPool += val
}

// NewEntityStateManager initializes and returns a new EntityStateManager based on the provided parent entity and configuration.
// Returns an error if the configuration contains invalid values.
func NewEntityStateManager(parent core.Entity, config EntityStateConfig) (*EntityStateManager, error) {
	// Handle mistakes
	if config.MaxHP < 0 {
		return nil, fmt.Errorf("max HP must not be negative")
	}
	if config.CurrentHP > config.MaxHP {
		config.CurrentHP = config.MaxHP
	}
	if config.CurrentHP < 0 {
		config.CurrentHP = 0
	}
	if config.TempHP < 0 {
		config.TempHP = 0
	}
	if config.MaxLegendaryActions < 0 {
		config.MaxLegendaryActions = 0
	}
	if config.AttackCount < 0 {
		config.AttackCount = 0
	}
	if config.Conditions == nil {
		config.Conditions = core.NewEntityConditions()
	}
	if config.Resistances == nil {
		config.Resistances = core.NewDamageResistances()
	}

	return &EntityStateManager{
		Parent:                    parent,
		CurrentHP:                 config.CurrentHP,
		MaxHP:                     config.MaxHP,
		TempHP:                    config.TempHP,
		HasUsedAction:             false,
		HasUsedBonusAction:        false,
		LegendaryActionPoints:     config.MaxLegendaryActions,
		LegendaryActionPointsMax:  config.MaxLegendaryActions,
		NumberOfAttacks:           config.AttackCount,
		Conditions:                config.Conditions,
		Resistances:               config.Resistances,
		ActionPreference:          config.ActionPreference,
		VersatileWeaponPreference: config.VersatilePreference,
		TargetPrioritization:      config.TargetPrioritization,
		SpellcastingPriority:      config.SpellcastingPriority,
		InitiativeAdvantage:       config.InitiativeAdvantage,
		InitiativeBonus:           config.InitiativeBonus,
		DeathSaves:                core.NewDeathSaves(),
		PaladinLayingOnHandsPool:  config.PaladinLayingOnHandsPool,
		BarbarianRelentlessUses:   config.BarbarianRelentlessUses,
		BarbarianIsRaging:         config.BarbarianIsRaging,
		FighterIndomitableUses:    config.FighterIndomitableUses,
	}, nil
}
