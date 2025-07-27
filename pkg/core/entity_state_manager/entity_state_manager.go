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
	NumberOfExtraAttacks     int

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

func (esm *EntityStateManager) GetNumberOfExtraAttacks() int {
	return esm.NumberOfExtraAttacks
}

func (esm *EntityStateManager) SetNumberOfExtraAttacks(value int) {
	esm.NumberOfExtraAttacks = value
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
		NumberOfExtraAttacks:      config.AttackCount,
		Conditions:                config.Conditions,
		ActionPreference:          config.ActionPreference,
		VersatileWeaponPreference: config.VersatilePreference,
		TargetPrioritization:      config.TargetPrioritization,
		SpellcastingPriority:      config.SpellcastingPriority,
		InitiativeAdvantage:       config.InitiativeAdvantage,
		InitiativeBonus:           config.InitiativeBonus,
	}, nil
}
