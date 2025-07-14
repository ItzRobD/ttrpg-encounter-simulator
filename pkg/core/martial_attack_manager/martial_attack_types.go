package martial_attack_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
)

type AttackOptions struct {
	Advantage            core.AdvantageType
	NumberOfAttacks      int
	BonusToAttackRoll    int  // Flat bonus, ie magic weapons
	BonusToDamageRoll    int  // Flat bonus, ie magic weapons, rage, hexblade curse
	ShouldApplyDamageMod bool // Off hand attacks, TWF
	PowerAttack          bool // GWM / Sharpshooter (-5 attack, +10 damage) // TODO: Implement logic for this choice
	ImprovedCritical     bool // Crits on 19 and 20, Hexblade, Champion
	RerollOnesAndTwos    bool // GWF
	// TODO: GWF Creates an extra attack
}

func (ao AttackOptions) GetAdvantage() core.AdvantageType { return ao.Advantage }
func (ao AttackOptions) GetNumberOfAttacks() int          { return ao.NumberOfAttacks }
func (ao AttackOptions) GetBonusToAttackRoll() int        { return ao.BonusToAttackRoll }
func (ao AttackOptions) GetBonusToDamageRoll() int        { return ao.BonusToDamageRoll }
func (ao AttackOptions) GetShouldApplyDamageMod() bool    { return ao.ShouldApplyDamageMod }
func (ao AttackOptions) GetIsPowerAttack() bool           { return ao.PowerAttack }
func (ao AttackOptions) GetIsImprovedCritical() bool      { return ao.ImprovedCritical }
func (ao AttackOptions) GetShouldRerollOnesAndTwos() bool { return ao.RerollOnesAndTwos }

type AttackData struct {
	Name              string
	NumberOfDice      int
	Die               core.DiceType
	AttackModifier    int // Added to attack roll. Character: Proficiency + Ability Mod; Monster: To Hit Bonus
	DamageModifier    int
	DamageType        core.DamageType
	IsVersatileAttack bool
}

func (ad AttackData) GetAttackName() string      { return ad.Name }
func (ad AttackData) GetNumberOfDice() int       { return ad.NumberOfDice }
func (ad AttackData) GetDie() core.DiceType      { return ad.Die }
func (ad AttackData) GetAttackModifier() int     { return ad.AttackModifier }
func (ad AttackData) GetDamageModifier() int     { return ad.DamageModifier }
func (ad AttackData) GetDamageType() string      { return ad.DamageType.String() }
func (ad AttackData) GetIsVersatileAttack() bool { return ad.IsVersatileAttack }

type AttackRequest struct {
	AttackData        AttackData
	AttackOptions     AttackOptions
	SimulationOptions core.SimulationOptions
	Target            core.Entity
}

func (ar *AttackRequest) GetAttackData() core.AttackData               { return ar.AttackData }
func (ar *AttackRequest) GetAttackOptions() core.AttackOptions         { return ar.AttackOptions }
func (ar *AttackRequest) GetSimulationOptions() core.SimulationOptions { return ar.SimulationOptions }
func (ar *AttackRequest) GetTarget() core.Entity                       { return ar.Target }

type AttackResult struct {
	ActorName     string
	TargetName    string
	AttackName    string
	AttackCount   int
	IsHit         bool
	IsCriticalHit bool
	AttackTotal   int
	AttackRoll    int
	DamageRoll    core.RollResult
	DamageType    core.DamageType
}

func (r AttackResult) GetActorName() string             { return r.ActorName }
func (r AttackResult) GetTargetName() string            { return r.TargetName }
func (r AttackResult) GetAttackName() string            { return r.AttackName }
func (r AttackResult) GetAttackCount() int              { return r.AttackCount }
func (r AttackResult) GetIsHit() bool                   { return r.IsHit }
func (r AttackResult) GetIsCriticalHit() bool           { return r.IsCriticalHit }
func (r AttackResult) GetAttackTotal() int              { return r.AttackTotal }
func (r AttackResult) GetAttackRoll() int               { return r.AttackRoll }
func (r AttackResult) GetDamageResult() core.RollResult { return r.DamageRoll }
func (r AttackResult) GetDamageType() core.DamageType   { return r.DamageType }
