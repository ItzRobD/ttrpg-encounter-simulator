package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"errors"
)

type SpellCastRequest struct {
	AttackData SpellAttackData
	Modifiers  SpellModifiers
	Advantage  core.AdvantageType
}

type SpellModifiers struct {
	TreatOnesAsTwos bool // Elemental Adept
	HalflingLucky   bool // Reroll 1s on Attacks, Saves
}

type SpellAttackData struct {
	SpellChoice    SpellChoice
	AttackModifier int
	SpellModifier  int
}

type SpellResult struct {
	ActorName        string
	TargetName       string
	SpellName        string
	SpellLevel       int
	SpellTotalValue  int // Damage or heal amount
	AttackRoll       int
	AttackTotal      int
	IsHit            bool
	IsCriticalHit    bool
	HasDC            bool
	SpellSaveAbility core.Ability
	SpellSaveRoll    int
	SpellSaveTotal   bool
	SpellSaveSuccess bool
	Damage           int
	DamageRolls      []int
	DamageType       string
}

func (r *SpellResult) GetActorName() string              { return r.ActorName }
func (r *SpellResult) GetTargetName() string             { return r.TargetName }
func (r *SpellResult) GetSpellName() string              { return r.SpellName }
func (r *SpellResult) GetSpellLevel() int                { return r.SpellLevel }
func (r *SpellResult) GetSpellTotalValue() int           { return r.SpellTotalValue }
func (r *SpellResult) GetAttackRoll() int                { return r.AttackRoll }
func (r *SpellResult) GetAttackTotal() int               { return r.AttackTotal }
func (r *SpellResult) GetIsHit() bool                    { return r.IsHit }
func (r *SpellResult) GetIsCriticalHit() bool            { return r.IsCriticalHit }
func (r *SpellResult) GetHasDC() bool                    { return r.HasDC }
func (r *SpellResult) GetSpellSaveAbility() core.Ability { return r.SpellSaveAbility }
func (r *SpellResult) GetSpellSaveRoll() int             { return r.SpellSaveRoll }
func (r *SpellResult) GetSpellSaveTotal() bool           { return r.SpellSaveTotal }
func (r *SpellResult) GetSpellSaveSuccess() bool         { return r.SpellSaveSuccess }
func (r *SpellResult) GetDamage() int                    { return r.Damage }
func (r *SpellResult) GetDamageRolls() []int             { return r.DamageRolls }
func (r *SpellResult) GetDamageType() string             { return r.DamageType }

func (s *SpellcastingManager) CastSpell(target core.Entity, spell *SpellChoice) {
	switch spell.Spell.SpellType {
	case spells.STDamage:
		// Damage
	case spells.STHealing:
		// Healing

	}
}

func calculatedamage(req *SpellCastRequest, isCritical bool, options core.Options) (int, []int, error) {
	var dmgMod int
	if req.AttackData.SpellChoice.Formula.UseSpellmod {
		dmgMod = req.AttackData.SpellModifier
	} else {
		dmgMod = req.AttackData.SpellChoice.Formula.AmountToAdd
	}

	var total int
	var rolls []int
	var err error

	if isCritical {
		total, rolls, err = core.CalculateDamageCriticalHit(req.AttackData.SpellChoice.Formula.NumberOfDice, req.AttackData.SpellChoice.Formula.Die, dmgMod, options.UseImprovedCriticals)
		if err != nil {
			return 0, nil, err
		}
	} else {
		total, rolls, err = core.DiceRollWithModifier(req.AttackData.SpellChoice.Formula.NumberOfDice, req.AttackData.SpellChoice.Formula.Die, dmgMod)
		if err != nil {
			return 0, nil, err
		}
	}

	// TODO: Log dice rolls
	if req.Modifiers.TreatOnesAsTwos {
		for i, roll := range rolls {
			if roll == 1 {
				rolls[i] = 2
			}
			// TODO: Log new roll replacement
		}
	}

	return total, rolls, nil
}

func (s *SpellcastingManager) castDamageSpell(target core.Entity, req *SpellCastRequest, options core.Options) error {
	var res SpellResult
	switch req.AttackData.SpellChoice.Spell.HasDC {
	case true:
		res.HasDC = true
		// TODO: Handle DC saves for spells
	case false:
		res.HasDC = false

		var attackTotal int
		var attackRoll int
		attackTotal, attackRoll, err := core.AttackRoll(req.AttackData.AttackModifier, req.Advantage)
		if err != nil {
			return err
		}

		// TODO: Note that halfling lucky does not use advantage again, re-roll lower
		if attackRoll == 1 && req.Modifiers.HalflingLucky {
			attackTotal, attackRoll, err = core.AttackRoll(req.AttackData.AttackModifier, req.Advantage)
			if err != nil {
				return err
			}
		}

		events.LogDiceRollEvent(s.parent, attackTotal, []int{attackRoll}, core.DiceRollAttack, req.AttackData.AttackModifier, s.parent.GetEventListener())
		res.AttackRoll = attackRoll
		res.AttackTotal = attackTotal

		critThreshold := 20
		isCrit := core.IsCriticalHit(attackRoll, critThreshold)
		if (isCrit || core.DoesAttackHit(attackRoll, target.GetAC())) && attackRoll != 1 {
			res.IsHit = true
			res.IsCriticalHit = isCrit

			events.LogSpellAttackEvent(s.parent, target, &res, s.parent.GetEventListener())

			damage, rolls, err2 := calculatedamage(req, isCrit, options)
			if err2 != nil {
				return err2
			}
			res.Damage = damage
			res.SpellTotalValue = damage
			res.DamageRolls = rolls
			res.DamageType = req.AttackData.SpellChoice.Formula.DamageType

		}
	default:
		return errors.New("invalid spell")
	}

	res.ActorName = s.parent.GetName()
	res.TargetName = target.GetName()
	res.SpellName = req.AttackData.SpellChoice.Spell.Name
	res.SpellLevel = req.AttackData.SpellChoice.Spell.Level
}

func (s *SpellcastingManager) castHealingSpell(target core.Entity, spell *SpellChoice) error {
	var result SpellResult

}
