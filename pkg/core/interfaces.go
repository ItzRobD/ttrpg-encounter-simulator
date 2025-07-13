package core

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
	GetSpellSaveRolls() []int
	GetSpellSaveTotal() int
	GetSpellSaveSuccess() bool
	GetDamage() int
	GetDamageRolls() []int
	GetDamageType() string
}

type DiceRollResultData interface {
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
}
