package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
)

type SpellCastRequest struct {
	SpellCastData     SpellCastData
	SpellOptions      SpellOptions
	SimulationOptions *core.SimulationOptions
	Target            core.Entity
}

func (scr SpellCastRequest) GetSpellCastData() core.SpellCastData { return scr.SpellCastData }
func (scr SpellCastRequest) GetSpellOptions() core.SpellOptions   { return scr.SpellOptions }
func (scr SpellCastRequest) GetSimulationOptions() *core.SimulationOptions {
	return scr.SimulationOptions
}
func (scr SpellCastRequest) GetTarget() core.Entity { return scr.Target }

type SpellOptions struct {
	Advantage            core.AdvantageType
	BonusToAttackRoll    int
	BonusToDamageRoll    int
	ShouldApplyDamageMod bool
	ImprovedCritical     bool
	TreatOnesAsTwos      bool // Elemental Adept
}

func (so SpellOptions) GetAdvantage() core.AdvantageType { return so.Advantage }
func (so SpellOptions) GetBonusToAttackRoll() int        { return so.BonusToAttackRoll }
func (so SpellOptions) GetBonusToDamageRoll() int        { return so.BonusToDamageRoll }
func (so SpellOptions) GetShouldApplyDamageMod() bool    { return so.ShouldApplyDamageMod }
func (so SpellOptions) GetIsImprovedCritical() bool      { return so.ImprovedCritical }
func (so SpellOptions) GetTreatOnesAsTwos() bool         { return so.TreatOnesAsTwos }

type SpellCastData struct {
	SpellChoice          core.SpellChoice
	AttackModifier       int // Attack Bonus
	SpellcastingModifier int // This is the caster's spellcast ability modifier
}

func (scd SpellCastData) GetSpellChoice() core.SpellChoice { return scd.SpellChoice }
func (scd SpellCastData) GetAttackModifier() int           { return scd.AttackModifier }
func (scd SpellCastData) GetSpellcastingModifier() int     { return scd.SpellcastingModifier }

type SpellResult struct {
	ActorName        string
	TargetName       string
	SpellName        string
	SpellLevel       int
	SpellTotalValue  int // Damage or heal amount
	AttackRoll       int
	AttackTotal      int
	IsSuccess        bool
	IsCriticalHit    bool
	HasDC            bool
	TargetDCValue    int
	SpellSaveAbility core.Ability
	SpellSaveEffect  core.DCOnSuccess
	SpellSaveRolls   []int
	SpellSaveTotal   int
	SpellSaveSuccess bool
	ValueRoll        core.RollResult
	DamageType       core.DamageType
	IsConcentration  bool
	IsAOE            bool
}

func (r *SpellResult) GetActorName() string                 { return r.ActorName }
func (r *SpellResult) GetTargetName() string                { return r.TargetName }
func (r *SpellResult) GetSpellName() string                 { return r.SpellName }
func (r *SpellResult) GetSpellLevel() int                   { return r.SpellLevel }
func (r *SpellResult) GetSpellTotalValue() int              { return r.SpellTotalValue }
func (r *SpellResult) GetAttackRoll() int                   { return r.AttackRoll }
func (r *SpellResult) GetAttackTotal() int                  { return r.AttackTotal }
func (r *SpellResult) GetIsHit() bool                       { return r.IsSuccess }
func (r *SpellResult) GetIsCriticalHit() bool               { return r.IsCriticalHit }
func (r *SpellResult) GetHasDC() bool                       { return r.HasDC }
func (r *SpellResult) GetTargetDCValue() int                { return r.TargetDCValue }
func (r *SpellResult) GetSpellSaveAbility() core.Ability    { return r.SpellSaveAbility }
func (r *SpellResult) GetSpellSaveEffect() core.DCOnSuccess { return r.SpellSaveEffect }
func (r *SpellResult) GetSpellSaveRolls() []int             { return r.SpellSaveRolls }
func (r *SpellResult) GetSpellSaveTotal() int               { return r.SpellSaveTotal }
func (r *SpellResult) GetSpellSaveSuccess() bool            { return r.SpellSaveSuccess }
func (r *SpellResult) GetValueResult() core.RollResult      { return r.ValueRoll }
func (r *SpellResult) GetDamageType() core.DamageType       { return r.DamageType }
func (r *SpellResult) GetIsConcentration() bool             { return r.IsConcentration }
