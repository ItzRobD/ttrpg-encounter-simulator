package core

//type AttackData interface {
//	GetAttackName() string
//	GetNumberOfDice() int
//	GetDie() DiceType
//	GetAttackModifier() int
//	GetDamageModifier() int
//	GetDamageType() string
//	GetIsVersatileAttack() bool
//}

type AttackResultData interface {
	GetActorName() string
	GetTargetName() string
	GetAttackName() string
	GetAttackCount() int
	GetIsHit() bool
	GetIsCriticalHit() bool
	GetAttackTotal() int
	GetAttackRoll() int
	GetDamage() int
	GetDamageRolls() []int
	GetDamageType() string
}

type SpellResult interface {
	GetActorName() string
	GetTargetName() string
	GetSpellName() string
	GetSpellLevel() int
	GetSpellTotalValue() int
	GetAttackRoll() int
	GetAttackTotal() int
	GetIsHit() bool
	GetIsCriticalHit() bool
	GetHasDC() bool
	GetTargetDCValue() int
	GetSpellSaveAbility() Ability
	GetSpellSaveEffect() DCOnSuccess
	GetSpellSaveRolls() []int
	GetSpellSaveTotal() int
	GetSpellSaveSuccess() bool
	GetValueResult() RollResult
	GetDamageType() DamageType
}

type SpellCastData interface {
	GetSpellChoice() SpellChoice
	GetAttackModifier() int
	GetSpellcastingModifier() int
}

type Spell interface {
	GetID() int
	GetName() string
	GetDescription() string
	GetIsConcentration() bool
	GetCastingTime() CastingTime
	GetIsRitual() bool
	GetLevel() int
	GetSpellType() SpellType
	GetIsAOE() bool
	GetHasDC() bool
	GetApiURL() string
	GetLevelType() string
	GetSpellDC() SpellDC
	GetFormulas() map[int]CastFormula
}

type SpellDC interface {
	GetAbility() Ability
	GetOnSuccess() DCOnSuccess
}

type SpellOptions interface {
	GetAdvantage() AdvantageType
	GetBonusToAttackRoll() int
	GetBonusToDamageRoll() int
	GetShouldApplyDamageMod() bool
	GetIsImprovedCritical() bool
	GetTreatOnesAsTwos() bool
}

type SpellCastRequest interface {
	GetSpellCastData() SpellCastData
	GetSpellOptions() SpellOptions
	GetSimulationOptions() *SimulationOptions
	GetTarget() Entity
}

type RollResult interface {
	GetDiceRollType() DiceRollType
	GetNumberOfDice() int
	GetDiceType() string
	GetFinalRollValue() int
	GetFinalRolls() []int
	GetModifier() int
	GetTotal() int
	GetAdvantage() string // Note: returns type string

	GetOriginalRolls() []int
	GetRerollEvents() []map[string]interface{}
	GetWasRerolled() bool

	GetIsCritical() bool
	GetIsNaturalOne() bool
	GetIsSuccess() bool
	GetTargetValue() int
}

type HPStatus interface {
	GetHP() int
	GetMaxHP() int
	GetTempHP() int
	GetHPPct() int
	GetHPDifference() int
	GetHitDie() DiceType
}

type ActionOutcomeData interface {
	GetActionType() ActionType
	GetTargetID() int
	GetActorID() int
	GetEffects() []Effect
}

type EffectData interface {
	GetType() EffectType
	GetValue() int
	GetDamageType() DamageType
	GetCondition() *Condition
}

type HPModificationResult interface {
	GetModificationValue() int
	GetOriginalHP() int
	GetOriginalTempHP() int
	GetNewHP() int
	GetNewTempHP() int
	GetDidHealHP() bool
	GetDidHealTempHP() bool
	GetDidTempDamage() bool
	GetDidHPDamage() bool
	GetIsUnconscious() bool
	GetIsMaxHealth() bool
}
