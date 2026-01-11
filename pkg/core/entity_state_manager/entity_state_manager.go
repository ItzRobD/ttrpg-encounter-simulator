package entity_state_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"fmt"
	"sort"
)

type HPModificationResult struct {
	ModificationValue           int
	OriginalHP                  int
	OriginalTempHP              int
	NewHP                       int
	NewTempHP                   int
	DidHealHP                   bool
	DidHealTempHP               bool
	DidTempDamage               bool
	DidHPDamage                 bool
	IsUnconscious               bool
	IsMaxHealth                 bool
	TriggeredConcentrationCheck bool
	DamageTaken                 int
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
func (hpmr HPModificationResult) GetTriggeredConcentrationCheck() bool {
	return hpmr.TriggeredConcentrationCheck
}
func (hpmr HPModificationResult) GetDamageTaken() int { return hpmr.DamageTaken }
func (hpmr HPModificationResult) GetHealingReceived() int {
	if hpmr.DidHealHP {
		return hpmr.ModificationValue
	}
	return 0
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
	Resistances          core.DamageResistances
	InitiativeBonus      int
	// Class specifics variables
	BarbarianRelentlessUses  int
	BarbarianIsRaging        bool
	FighterIndomitableUses   int
	PaladinLayingOnHandsPool int
	RelentlessThreshold      int
	HasUndeadFortitude       bool
}

type EntityStateManager struct {
	Parent core.Entity
	// HP Management
	currentHP int
	maxHP     int
	tempHP    int
	hitDie    core.DiceType

	// Action Economy
	hasUsedAction            bool
	hasUsedBonusAction       bool
	hasUsedReaction          bool
	legendaryActionPoints    int
	legendaryActionPointsMax int
	numberOfAttacks          int
	rechargeActionStatus     map[int]bool // Key: Action index; Value: IsAvailable
	dbBreathWeaponUsed       bool

	// conditions
	conditions            core.EntityConditions
	deathSaves            core.DeathSaves
	isStable              bool
	isDead                bool
	isRecklesslyAttacking bool

	isConcentrating        bool
	concentratingSpellName string

	initiative int

	// Preferences
	actionPreference          core.ActionPreference
	versatileWeaponPreference core.VersatileWeaponPreference
	targetPrioritization      core.TargetPriority
	spellcastingPriority      core.SpellPriority

	// Bonuses
	initiativeAdvantage  core.AdvantageType
	initiativeBonus      int
	resistances          core.DamageResistances
	savingThrowAdvantage map[core.Ability]core.AdvantageType

	// Class specific variables
	barbarianHasRelentlessRage bool
	barbarianRelentlessUses    int
	BarbarianIsRaging          bool
	FighterIndomitableUses     int
	PaladinLayingOnHandsPool   int

	// Race specific variables
	HalfOrcHasSavageAttacks          bool
	HalfOrcHasRelentlessEnduranceUse bool

	// Monster special abilities
	LegendaryResistanceUses int
	isBerserking            bool
	relentlessThreshold     int
	hasUndeadFortitude      bool
	hasUsedMartialAdvantage bool
	isDivineEminenceActive  bool
	divineEminenceDice      int
	hasUsedSneakAttack      bool
	hasTakenTurnInCombat    bool
}

func (esm *EntityStateManager) GetHasUsedAction() bool {
	return esm.hasUsedAction
}

func (esm *EntityStateManager) GetHasUsedBonusAction() bool {
	return esm.hasUsedBonusAction
}

func (esm *EntityStateManager) GetHasUsedReaction() bool {
	return esm.hasUsedReaction
}

func (esm *EntityStateManager) SetHasUsedAction(val bool) {
	esm.hasUsedAction = val
}

func (esm *EntityStateManager) SetHasUsedBonusAction(val bool) {
	esm.hasUsedBonusAction = val
}

func (esm *EntityStateManager) SetHasUsedReaction(val bool) {
	esm.hasUsedReaction = val
}

func (esm *EntityStateManager) ExpendAction() {
	esm.hasUsedAction = true
}

func (esm *EntityStateManager) ExpendBonusAction() {
	esm.hasUsedBonusAction = true
}

func (esm *EntityStateManager) ExpendReaction() {
	esm.hasUsedReaction = true
}

func (esm *EntityStateManager) ReplenishAction() {
	esm.hasUsedAction = false
}

func (esm *EntityStateManager) ReplenishBonusAction() {
	esm.hasUsedBonusAction = false
}

func (esm *EntityStateManager) ReplenishReaction() {
	esm.hasUsedReaction = false
}

func (esm *EntityStateManager) RefreshActions() {
	esm.hasUsedAction = false
	esm.hasUsedBonusAction = false
	esm.hasUsedReaction = false
	esm.legendaryActionPoints = esm.legendaryActionPointsMax
	esm.hasUsedMartialAdvantage = false
	esm.isDivineEminenceActive = false
	esm.hasUsedSneakAttack = false
}

func (esm *EntityStateManager) CanTakeActions() bool {
	// conditions that prevent ALL actions
	if esm.conditions.Has(core.ConditionIncapacitated) ||
		esm.conditions.Has(core.ConditionStunned) ||
		esm.conditions.Has(core.ConditionParalyzed) ||
		esm.conditions.Has(core.ConditionPetrified) ||
		esm.conditions.Has(core.ConditionUnconscious) ||
		esm.isDead {
		return false
	}

	// Check if any action economy is available
	return !esm.hasUsedAction || !esm.hasUsedBonusAction || esm.legendaryActionPoints > 0
}

func (esm *EntityStateManager) ExpendLegendaryActionPoints(value int) error {
	if value > esm.legendaryActionPoints {
		return fmt.Errorf("cannot expend more legendary action points than available")
	}
	esm.legendaryActionPoints -= value
	return nil
}

func (esm *EntityStateManager) ReplenishLegendaryActionPoints(value int) {
	esm.legendaryActionPoints = max(esm.legendaryActionPoints+value, esm.legendaryActionPointsMax)
}

func (esm *EntityStateManager) GetLegendaryActionPoints() int {
	return esm.legendaryActionPoints
}

func (esm *EntityStateManager) GetLegendaryActionPointsMax() int {
	return esm.legendaryActionPointsMax
}

func (esm *EntityStateManager) HasLegendaryActionPointsRemaining() bool {
	return esm.legendaryActionPoints > 0
}

func (esm *EntityStateManager) GetNumberOfAttacks() int {
	return esm.numberOfAttacks
}

func (esm *EntityStateManager) SetNumberOfExtraAttacks(value int) {
	esm.numberOfAttacks = value
}

func (esm *EntityStateManager) SetActionPreference(p core.ActionPreference) {
	esm.actionPreference = p
}

func (esm *EntityStateManager) GetActionPreference() core.ActionPreference {
	return esm.actionPreference
}

func (esm *EntityStateManager) SetVersatileWeaponPreference(p core.VersatileWeaponPreference) {
	esm.versatileWeaponPreference = p
}

func (esm *EntityStateManager) GetVersatileWeaponPreference() core.VersatileWeaponPreference {
	return esm.versatileWeaponPreference
}

func (esm *EntityStateManager) SetTargetPrioritization(p core.TargetPriority) {
	esm.targetPrioritization = p
}

func (esm *EntityStateManager) GetTargetPrioritization() core.TargetPriority {
	return esm.targetPrioritization
}

func (esm *EntityStateManager) SetSpellcastingPriority(p core.SpellPriority) {
	esm.spellcastingPriority = p
}

func (esm *EntityStateManager) GetSpellcastingPriority() core.SpellPriority {
	return esm.spellcastingPriority
}

func (esm *EntityStateManager) SetInitiativeAdvantage(a core.AdvantageType) {
	esm.initiativeAdvantage = a
}

func (esm *EntityStateManager) SetInitiative(value int) { esm.initiative = value }

func (esm *EntityStateManager) GetInitiative() int { return esm.initiative }

func (esm *EntityStateManager) GetInitiativeAdvantage() core.AdvantageType {
	return esm.initiativeAdvantage
}

func (esm *EntityStateManager) SetInitiativeBonus(b int) {
	esm.initiativeBonus = b
}

func (esm *EntityStateManager) GetInitiativeBonus() int {
	return esm.initiativeBonus
}

// conditions functions

func (esm *EntityStateManager) AddCondition(c core.Condition) {
	// Special handling for unconscious: also add prone condition
	if c == core.ConditionUnconscious {
		esm.conditions.Add(core.ConditionUnconscious)
		esm.conditions.Add(core.ConditionProne)
		events.LogConditionEvent(esm.Parent.GetCurrentEventContext(), esm.Parent, core.ConditionUnconscious, true, esm.Parent.GetEventListener())
		events.LogConditionEvent(esm.Parent.GetCurrentEventContext(), esm.Parent, core.ConditionProne, true, esm.Parent.GetEventListener())
	} else {
		esm.conditions.Add(c)
		events.LogConditionEvent(esm.Parent.GetCurrentEventContext(), esm.Parent, c, true, esm.Parent.GetEventListener())
	}

	// Break concentration if incapacitated or other severe conditions
	if c == core.ConditionIncapacitated || c == core.ConditionStunned || c == core.ConditionParalyzed || c == core.ConditionPetrified || c == core.ConditionUnconscious {
		esm.BreakConcentration()
	}
}

func (esm *EntityStateManager) RemoveCondition(c core.Condition) {
	if esm.conditions.Has(c) {
		esm.conditions.Remove(c)
		events.LogConditionEvent(esm.Parent.GetCurrentEventContext(), esm.Parent, c, false, esm.Parent.GetEventListener())
	}
}

func (esm *EntityStateManager) HasCondition(c core.Condition) bool {
	return esm.conditions.Has(c)
}

func (esm *EntityStateManager) GetConditions() core.EntityConditions {
	return esm.conditions
}

func (esm *EntityStateManager) GetActiveConditions() []core.Condition {
	return esm.conditions.GetActive()
}

func (esm *EntityStateManager) ResetConditions() {
	for _, c := range esm.conditions.GetActive() {
		events.LogConditionEvent(esm.Parent.GetCurrentEventContext(), esm.Parent, c, false, esm.Parent.GetEventListener())
	}
	esm.conditions.Clear()
}

func (esm *EntityStateManager) SetUnconscious(isUnconscious bool) {
	if isUnconscious {
		// Directly add conditions to avoid circular call
		esm.conditions.Add(core.ConditionUnconscious)
		esm.conditions.Add(core.ConditionProne)
		events.LogConditionEvent(esm.Parent.GetCurrentEventContext(), esm.Parent, core.ConditionUnconscious, true, esm.Parent.GetEventListener())
		events.LogConditionEvent(esm.Parent.GetCurrentEventContext(), esm.Parent, core.ConditionProne, true, esm.Parent.GetEventListener())
		esm.BreakConcentration()
	} else {
		if esm.conditions.Has(core.ConditionUnconscious) {
			esm.conditions.Remove(core.ConditionUnconscious)
			events.LogConditionEvent(esm.Parent.GetCurrentEventContext(), esm.Parent, core.ConditionUnconscious, false, esm.Parent.GetEventListener())
		}
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
		if esm.conditions.Has(condition) {
			incapacitating = append(incapacitating, condition)
		}
	}

	return incapacitating
}

// GetIsUnconscious determines if the entity is unconscious based on its conditions or current health points.
func (esm *EntityStateManager) GetIsUnconscious() bool {
	return esm.conditions.Has(core.ConditionUnconscious) || esm.currentHP <= 0
}

func (esm *EntityStateManager) GetIsStable() bool {
	return esm.isStable
}

func (esm *EntityStateManager) GetIsDead() bool { return esm.isDead }

func (esm *EntityStateManager) GetHasUsedMartialAdvantage() bool {
	return esm.hasUsedMartialAdvantage
}

func (esm *EntityStateManager) SetHasUsedMartialAdvantage(val bool) {
	esm.hasUsedMartialAdvantage = val
}

func (esm *EntityStateManager) GetHasUsedSneakAttack() bool {
	return esm.hasUsedSneakAttack
}

func (esm *EntityStateManager) SetHasUsedSneakAttack(val bool) {
	esm.hasUsedSneakAttack = val
}

func (esm *EntityStateManager) GetHasTakenTurnInCombat() bool {
	return esm.hasTakenTurnInCombat
}

func (esm *EntityStateManager) SetHasTakenTurnInCombat(val bool) {
	esm.hasTakenTurnInCombat = val
}

func (esm *EntityStateManager) GetIsDivineEminenceActive() bool {
	return esm.isDivineEminenceActive
}

func (esm *EntityStateManager) SetDivineEminenceActive(val bool) {
	esm.isDivineEminenceActive = val
}

func (esm *EntityStateManager) GetDivineEminenceDice() int {
	return esm.divineEminenceDice
}

func (esm *EntityStateManager) SetDivineEminenceDice(val int) {
	esm.divineEminenceDice = val
}

func (esm *EntityStateManager) SetLegendaryActionPoints(val int) {
	esm.legendaryActionPoints = val
}

func (esm *EntityStateManager) SetLegendaryActionPointsMax(val int) {
	esm.legendaryActionPointsMax = val
}

func (esm *EntityStateManager) SetConditions(c core.EntityConditions) {
	esm.conditions = c
}

// HP Functions

func (esm *EntityStateManager) ResetHP() {
	esm.currentHP = esm.maxHP
	esm.tempHP = 0
}

func (esm *EntityStateManager) SetHPValues(hp HPValues) {
	esm.currentHP = hp.GetHP()
	esm.maxHP = hp.GetMaxHP()
	esm.tempHP = hp.GetTempHP()
	esm.hitDie = hp.GetHitDie()
}

func (esm *EntityStateManager) GetHitDie() core.DiceType {
	return esm.hitDie
}

func (esm *EntityStateManager) SetHitDie(d core.DiceType) { esm.hitDie = d }

func (esm *EntityStateManager) GetCurrentHP() int {
	return esm.currentHP
}

func (esm *EntityStateManager) GetMaxHP() int {
	return esm.maxHP
}

func (esm *EntityStateManager) GetTempHP() int {
	return esm.tempHP
}

func (esm *EntityStateManager) GetIsMaxHealth() bool {
	return esm.currentHP == esm.maxHP
}

func (esm *EntityStateManager) GetTotalHP() int {
	return esm.currentHP + esm.tempHP
}

func (esm *EntityStateManager) ModifyHP(value int, isTemp bool, tempStacking bool, allowMassiveDamage bool, damageType core.DamageType, isCritical bool) (HPModificationResult, error) {
	res := HPModificationResult{
		ModificationValue:           value,
		OriginalHP:                  esm.currentHP,
		OriginalTempHP:              esm.tempHP,
		NewHP:                       esm.currentHP,
		NewTempHP:                   esm.tempHP,
		DidHealHP:                   false,
		DidHealTempHP:               false,
		DidTempDamage:               false,
		DidHPDamage:                 false,
		IsUnconscious:               false,
		IsMaxHealth:                 false,
		TriggeredConcentrationCheck: false,
		DamageTaken:                 0,
	}
	if isTemp { // This should only be true if we are adding to temp hp
		if value < 0 {
			dmg := -value
			res.DamageTaken = dmg
			res.DidHealHP = res.NewHP > res.OriginalHP
			res.DidHealTempHP = res.NewTempHP > res.OriginalTempHP
			res.DidTempDamage = res.NewTempHP < res.OriginalTempHP
			res.DidHPDamage = res.NewHP < res.OriginalHP
			res.IsUnconscious = esm.currentHP == 0
			res.IsMaxHealth = esm.currentHP == esm.maxHP
			// This should not happen logically, return
			return res, fmt.Errorf("cannot specifically target temp hp with a negative value")
		} else {
			if tempStacking {
				esm.tempHP += value
				res.NewTempHP = esm.tempHP
			} else {
				esm.tempHP = max(esm.tempHP, value)
			}
		}
	} else {
		if value < 0 {
			dmg := -value // positive magnitude
			res.DamageTaken = dmg
			if esm.tempHP > 0 {
				if esm.tempHP >= dmg {
					esm.tempHP -= dmg // exact or partial absorption
				} else {
					overflow := dmg - esm.tempHP
					esm.tempHP = 0
					esm.currentHP -= overflow
				}
			} else {
				esm.currentHP -= dmg
			}

			// Trigger concentration check if damage was taken and concentrating
			if esm.isConcentrating && dmg > 0 {
				res.TriggeredConcentrationCheck = true
			}
		} else { // Healing
			esm.currentHP = min(esm.currentHP+value, esm.maxHP)
		}
		res.NewHP = esm.currentHP
		res.NewTempHP = esm.tempHP
	}
	res.DidHealHP = res.NewHP > res.OriginalHP
	res.DidHealTempHP = res.NewTempHP > res.OriginalTempHP
	res.DidTempDamage = res.NewTempHP < res.OriginalTempHP
	res.DidHPDamage = res.NewHP < res.OriginalHP
	res.IsUnconscious = esm.currentHP <= 0
	res.IsMaxHealth = esm.currentHP == esm.maxHP

	// Handle 0 HP logic
	if res.IsUnconscious {
		// Massive Damage Check: if remaining damage equals or exceeds max HP, it's instant death
		if allowMassiveDamage && esm.CheckMassiveDamage() {
			esm.Kill()
			res.IsUnconscious = false // Overridden by dead
			esm.Parent.LogEvent(events.ETSpecialAbilityEvent, &events.SpecialAbilityData{
				AbilityName: "Massive Damage",
				Description: "Killed by massive damage!",
				TargetName:  "",
				Value:       0,
			})
		} else if esm.Parent.IsMonster() {
			// Relentless (Monster)
			if esm.relentlessThreshold > 0 && res.OriginalHP > 0 && res.DamageTaken <= esm.relentlessThreshold {
				esm.SetUnconscious(false)
				esm.currentHP = 1
				res.IsUnconscious = false
				res.NewHP = esm.currentHP
				esm.Parent.LogEvent(events.ETSpecialAbilityEvent, &events.SpecialAbilityData{
					AbilityName: "Relentless",
					Description: fmt.Sprintf("Relentless triggered. New HP set to 1 (Damage taken: %d <= threshold: %d).", res.DamageTaken, esm.relentlessThreshold),
					TargetName:  "",
					Value:       1,
				})
			} else if esm.hasUndeadFortitude && res.OriginalHP > 0 && damageType != core.DamageRadiant && !isCritical {
				// Undead Fortitude
				// If damage reduces the zombie to 0 hit points, it must make a Constitution saving throw with a DC of 5 + the damage taken,
				// unless the damage is radiant or from a critical hit. On a success, the zombie drops to 1 hit point instead.
				dc := 5 + res.DamageTaken

				saveResult, err := esm.Parent.MakeSavingThrow(core.AbilityConstitution, dc, false, core.DamageNone, nil)
				if err != nil {
					return res, err
				}

				if saveResult.GetIsSuccess() {
					esm.SetUnconscious(false)
					esm.currentHP = 1
					res.IsUnconscious = false
					res.NewHP = esm.currentHP
					esm.Parent.LogEvent(events.ETSpecialAbilityEvent, &events.SpecialAbilityData{
						AbilityName: "Undead Fortitude",
						Description: fmt.Sprintf("Undead Fortitude triggered. New HP set to 1 (DC %d).", dc),
						TargetName:  "",
						Value:       1,
					})
				} else {
					esm.Kill()
				}
			} else {
				// Monsters die instantly at 0 HP (standard D&D 5e rules)
				esm.Kill()
			}
		} else {
			// Relentless Endurance (Half-Orc)
			// Trigger only if we were not already at 0 HP
			if esm.HalfOrcHasRelentlessEnduranceUse && res.OriginalHP > 0 {
				esm.SetUnconscious(false)
				esm.currentHP = 1
				esm.HalfOrcHasRelentlessEnduranceUse = false
				res.IsUnconscious = false
				res.NewHP = esm.currentHP
				esm.Parent.LogEvent(events.ETSpecialAbilityEvent, &events.SpecialAbilityData{
					AbilityName: "Relentless Endurance",
					Description: "Relentless Endurance expended. New HP set to 1.",
					TargetName:  "",
					Value:       1,
				})
			} else if esm.barbarianHasRelentlessRage && esm.BarbarianIsRaging && res.OriginalHP > 0 {
				// Barbarian Relentless Rage
				// Make DC 10 + (uses * 5) Con saving throw
				dc := 10 + (esm.GetBarbarianRelentlessUses() * 5)
				saveResult, err := esm.Parent.MakeSavingThrow(core.AbilityConstitution, dc, false, core.DamageNone, nil)
				if err != nil {
					return res, err
				}
				if saveResult.GetIsSuccess() {
					esm.SetUnconscious(false)
					esm.currentHP = 1
					res.IsUnconscious = false
					res.NewHP = esm.currentHP
					esm.IncrementBarbarianRelentlessUses()
					esm.Parent.LogEvent(events.ETSpecialAbilityEvent, &events.SpecialAbilityData{
						AbilityName: "Relentless Rage",
						Description: "Relentless Rage expended. New HP set to 1.",
						TargetName:  "",
						Value:       1,
					})
				}
			} else {
				esm.SetUnconscious(true)
			}
		}
	} else if res.OriginalHP <= 0 && res.NewHP > 0 {
		// Character was healed from 0 or stable
		esm.SetUnconscious(false)
		esm.deathSaves.Reset()
		esm.isStable = false
	}

	return res, nil
}

// CheckMassiveDamage determines if the entity has taken damage exceeding its maximum HP in the negative range.
// Returns true if the entity's current HP is less than or equal to the negative value of its maximum HP.
func (esm *EntityStateManager) CheckMassiveDamage() bool {
	if esm.currentHP <= -esm.maxHP {
		return true
	}
	return false
}

// Recharge Actions

// GetRechargeActionStatus returns the current recharge status for all actions as a map where keys are action indexes.
func (esm *EntityStateManager) GetRechargeActionStatus() map[int]bool {
	return esm.rechargeActionStatus
}

// GetRechargeActionStatusAtIndex checks if the recharge action at the specified index is available.
func (esm *EntityStateManager) GetRechargeActionStatusAtIndex(index int) bool {
	return esm.rechargeActionStatus[index]
}

// GetExpendedRechargeActionsIndex returns a list of indexes for recharge actions that are currently expended (unavailable).
func (esm *EntityStateManager) GetExpendedRechargeActionsIndex() []int {
	var result []int
	for i, v := range esm.rechargeActionStatus {
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
	esm.rechargeActionStatus[index] = false
}

// RechargeRechargeAction sets the recharge action at the specified index to available (true).
func (esm *EntityStateManager) RechargeRechargeAction(index int) {
	esm.rechargeActionStatus[index] = true
}

// ResetAllRechargeActions resets all recharge actions to available by setting their status to true.
func (esm *EntityStateManager) ResetAllRechargeActions() {
	for i := 0; i < len(esm.rechargeActionStatus); i++ {
		esm.rechargeActionStatus[i] = true
	}
}

func (esm *EntityStateManager) SetDBBreathWeaponUsed(val bool) {
	esm.dbBreathWeaponUsed = val
}

func (esm *EntityStateManager) GetDBBreathWeaponUsed() bool {
	return esm.dbBreathWeaponUsed
}

func (esm *EntityStateManager) AddRechargeAction(index int) {
	if esm.rechargeActionStatus == nil {
		esm.rechargeActionStatus = make(map[int]bool)
	}
	esm.rechargeActionStatus[index] = true
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
		esm.deathSaves.AddFailure(true)
	} else {
		if result.GetIsSuccess() {
			esm.deathSaves.AddSuccess()
		} else {
			esm.deathSaves.AddFailure(false)
		}
	}

	status := esm.deathSaves.Evaluate()
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
		esm.deathSaves.AddFailure(true)
	} else {
		esm.deathSaves.AddFailure(false)
	}

	if esm.deathSaves.Evaluate() == core.DeathSaveFailure {
		esm.Kill()
	}
}

func (esm *EntityStateManager) SetStable(stable bool) {
	esm.isStable = stable
	esm.deathSaves.Reset()
}

// Revive attempts to revive an entity by setting its HP and resetting relevant states if it is currently not alive.
// Returns an error if the entity is already alive.
func (esm *EntityStateManager) Revive(hpValue int) error {
	if esm.currentHP > 0 {
		return fmt.Errorf("cannot revive an entity that is already alive")
	}

	esm.currentHP = hpValue
	esm.isStable = false
	esm.deathSaves.Reset()
	esm.RemoveCondition(core.ConditionUnconscious)

	return nil
}

func (esm *EntityStateManager) Kill() {
	esm.isDead = true
	esm.isStable = false
	esm.conditions.Clear()
	if esm.isConcentrating {
		esm.BreakConcentration()
	}
}

func (esm *EntityStateManager) GetSavingThrowAdvantage(ability core.Ability) core.AdvantageType {
	return esm.savingThrowAdvantage[ability]
}

func (esm *EntityStateManager) SetHasSavingThrowAdvantage(ability core.Ability, adv core.AdvantageType) {
	esm.savingThrowAdvantage[ability] = adv
}

// resistances

func (esm *EntityStateManager) AddResistance(dt core.DamageType, rt core.ResistanceType, rb []core.ResistBreaker) {
	// Fallback if resistances is not initialized
	if esm.resistances == nil {
		esm.resistances = core.NewDamageResistances()
	}

	esm.resistances.SetResistance(dt, rt, rb)
}

func (esm *EntityStateManager) RemoveResistance(dt core.DamageType) error {
	if esm.resistances == nil {
		return fmt.Errorf("resistances not initialized")
	}

	esm.resistances.ResetResistance(dt)
	return nil
}

func (esm *EntityStateManager) SetResistances(r core.DamageResistances) {
	esm.resistances = r
}

func (esm *EntityStateManager) GetResistances() core.DamageResistances {
	return esm.resistances
}

func (esm *EntityStateManager) GetResistance(dt core.DamageType) (core.DamageResistance, error) {
	if esm.resistances == nil {
		return core.DamageResistance{}, fmt.Errorf("resistances not initialized for parent: %v", esm.Parent.GetName())
	}
	return esm.resistances.GetResistance(dt), nil
}

// Class-specific functions
func (esm *EntityStateManager) ResetBarbarianRelentlessUses() {
	esm.barbarianRelentlessUses = 0
}

func (esm *EntityStateManager) SetConcentrating(val bool, spellName string) {
	esm.isConcentrating = val
	esm.concentratingSpellName = spellName
}

func (esm *EntityStateManager) IsConcentrating() bool {
	return esm.isConcentrating
}

func (esm *EntityStateManager) GetConcentratingSpellName() string {
	return esm.concentratingSpellName
}

func (esm *EntityStateManager) BreakConcentration() {
	if esm.isConcentrating {
		esm.isConcentrating = false
		esm.concentratingSpellName = ""
		// Dispatch event? Maybe later if needed
	}
}

func (esm *EntityStateManager) GetBarbarianRelentlessUses() int {
	return esm.barbarianRelentlessUses
}

func (esm *EntityStateManager) SetBarbarianIsRaging(val bool) { esm.BarbarianIsRaging = val }

func (esm *EntityStateManager) GetBarbarianIsRaging() bool { return esm.BarbarianIsRaging }

func (esm *EntityStateManager) SetIsRecklesslyAttacking(val bool) {
	esm.isRecklesslyAttacking = val
}

func (esm *EntityStateManager) GetIsRecklesslyAttacking() bool {
	return esm.isRecklesslyAttacking
}

func (esm *EntityStateManager) SetBarbarianRelentlessRage(val bool) {
	esm.barbarianHasRelentlessRage = val
}

func (esm *EntityStateManager) IncrementBarbarianRelentlessUses() {
	esm.barbarianRelentlessUses++
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

func (esm *EntityStateManager) SetLegendaryResistanceUses(val int) {
	esm.LegendaryResistanceUses = val
}

func (esm *EntityStateManager) GetLegendaryResistanceUses() int {
	return esm.LegendaryResistanceUses
}

func (esm *EntityStateManager) ExpendLegendaryResistanceUse() {
	esm.LegendaryResistanceUses--
}

func (esm *EntityStateManager) SetIsBerserking(val bool) {
	esm.isBerserking = val
}

func (esm *EntityStateManager) GetIsBerserking() bool {
	return esm.isBerserking
}

func (esm *EntityStateManager) SetRelentlessThreshold(val int) {
	esm.relentlessThreshold = val
}

func (esm *EntityStateManager) GetRelentlessThreshold() int {
	return esm.relentlessThreshold
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
		currentHP:                 config.CurrentHP,
		maxHP:                     config.MaxHP,
		tempHP:                    config.TempHP,
		hasUsedAction:             false,
		hasUsedBonusAction:        false,
		legendaryActionPoints:     config.MaxLegendaryActions,
		legendaryActionPointsMax:  config.MaxLegendaryActions,
		numberOfAttacks:           config.AttackCount,
		conditions:                config.Conditions,
		resistances:               config.Resistances,
		actionPreference:          config.ActionPreference,
		versatileWeaponPreference: config.VersatilePreference,
		targetPrioritization:      config.TargetPrioritization,
		spellcastingPriority:      config.SpellcastingPriority,
		initiativeAdvantage:       config.InitiativeAdvantage,
		initiativeBonus:           config.InitiativeBonus,
		deathSaves:                core.NewDeathSaves(),
		PaladinLayingOnHandsPool:  config.PaladinLayingOnHandsPool,
		barbarianRelentlessUses:   config.BarbarianRelentlessUses,
		BarbarianIsRaging:         config.BarbarianIsRaging,
		FighterIndomitableUses:    config.FighterIndomitableUses,
		relentlessThreshold:       config.RelentlessThreshold,
		hasUndeadFortitude:        config.HasUndeadFortitude,
	}, nil
}
