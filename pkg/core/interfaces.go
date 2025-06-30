package core

//type AttackData interface {
//	GetActorName() string
//	GetTargetName() string
//	GetAttackName() string
//	GetAttackCount() int
//	IsHit() bool
//	CriticalHit() bool
//	GetAttackTotal() int
//	GetAttackRoll() int
//	GetDamage() int
//	GetDamageRolls() []int
//	GetDamageType() string
//}
//
//type AttackModifiers interface {
//	GetBonusAttackRoll() int
//	GetBonusDamageRoll() int
//	GetShouldApplyDamageMod() bool
//	GetPowerAttack() bool
//	GetImprovedCritical() bool
//	GetTreatOnesAsTwos() bool
//	GetRerollOnesAndTwos() bool
//	GetHalflingLucky() bool
//}
//
//type AttackRequestData interface {
//	GetAttackData() AttackData
//	GetModifiers() AttackModifiers
//	GetAdvantage() AdvantageType
//	GetAttackCount() int
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

type SpellResultData interface {
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
	GetSpellSaveAbility() Ability
	GetSpellSaveRoll() int
	GetSpellSaveTotal() bool
	GetSpellSaveSuccess() bool
	GetDamage() int
	GetDamageRolls() []int
	GetDamageType() string
}
