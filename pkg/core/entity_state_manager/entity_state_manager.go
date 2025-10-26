package entity_state_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
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
	InitiativeBonus      int
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
	LegendaryActionPoints    int
	LegendaryActionPointsMax int
	NumberOfAttacks          int
	RechargeActionStatus     map[int]bool // Key: Action index; Value: IsAvailable

	// Conditions
	Conditions core.EntityConditions
	DeathSaves core.DeathSaves
	Initiative int

	// Preferences
	ActionPreference          core.ActionPreference
	VersatileWeaponPreference core.VersatileWeaponPreference
	TargetPrioritization      core.TargetPriority
	SpellcastingPriority      core.SpellPriority

	// Bonuses
	InitiativeAdvantage core.AdvantageType
	InitiativeBonus     int
	Resistances         core.DamageResistances
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

func (esm *EntityStateManager) SetHitDie(d core.DiceType) {
	esm.HitDie = d
}

func (esm *EntityStateManager) GetCurrentHP() int {
	return esm.CurrentHP
}

func (esm *EntityStateManager) GetMaxHP() int {
	return esm.MaxHP
}

func (esm *EntityStateManager) GetTempHP() int {
	return esm.TempHP
}

func (esm *EntityStateManager) GetIsUnconscious() bool {
	return esm.CurrentHP <= 0
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
		if value < 0 { // We are doing damage
			// Subtract from temp hp first
			if esm.TempHP > 0 {
				dmg := value
				if esm.TempHP > -value { // Arithmatic operators are backwards as value is negative
					esm.TempHP += value
				} else {
					diff := dmg + esm.TempHP // Overflow damage
					esm.TempHP = 0
					esm.CurrentHP = min(esm.CurrentHP+diff, 0)
				}
			} else {
				esm.CurrentHP += value // Direct HP Damage
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

	// TODO: Should this be handled within the manager during hp changes?
	if res.IsUnconscious {
		esm.AddCondition(core.ConditionUnconscious)
	}

	return res, nil
}

func (esm *EntityStateManager) ExpendAction() {
	esm.HasUsedAction = true
}

func (esm *EntityStateManager) ExpendBonusAction() {
	esm.HasUsedBonusAction = true
}

func (esm *EntityStateManager) ReplenishAction() {
	esm.HasUsedAction = false
}

func (esm *EntityStateManager) ReplenishBonusAction() {
	esm.HasUsedBonusAction = false
}

func (esm *EntityStateManager) CanTakeActions() bool {
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

func (esm *EntityStateManager) GetNumberOfAttacks() int {
	return esm.NumberOfAttacks
}

func (esm *EntityStateManager) SetNumberOfExtraAttacks(value int) {
	esm.NumberOfAttacks = value
}

func (esm *EntityStateManager) AddCondition(c core.Condition) {
	esm.Conditions[c] = true
}

func (esm *EntityStateManager) RemoveCondition(c core.Condition) {
	esm.Conditions[c] = false
}

func (esm *EntityStateManager) HasCondition(c core.Condition) bool {
	return esm.Conditions[c]
}

func (esm *EntityStateManager) GetConditions() core.EntityConditions {
	return esm.Conditions
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

func (esm *EntityStateManager) ResetActions() {
	esm.HasUsedAction = false
	esm.HasUsedBonusAction = false
	esm.LegendaryActionPoints = esm.LegendaryActionPointsMax
}

func (esm *EntityStateManager) ResetConditions() {
	esm.Conditions = core.NewEntityConditions()
}

func (esm *EntityStateManager) ResetHP() {
	esm.CurrentHP = esm.MaxHP
	esm.TempHP = 0
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

// AddRechargeAction adds a recharge action at the specified index and initializes the RechargeActionStatus map if nil.
func (esm *EntityStateManager) AddRechargeAction(index int) {
	if esm.RechargeActionStatus == nil {
		esm.RechargeActionStatus = make(map[int]bool)
	}
	esm.RechargeActionStatus[index] = true
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
		ActionPreference:          config.ActionPreference,
		VersatileWeaponPreference: config.VersatilePreference,
		TargetPrioritization:      config.TargetPrioritization,
		SpellcastingPriority:      config.SpellcastingPriority,
		InitiativeAdvantage:       config.InitiativeAdvantage,
		InitiativeBonus:           config.InitiativeBonus,
	}, nil
}
