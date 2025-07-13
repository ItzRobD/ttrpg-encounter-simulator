package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"errors"
	"fmt"
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
	SpellChoice    spells.SpellChoice
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
	SpellSaveRolls   []int
	SpellSaveTotal   int
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
func (r *SpellResult) GetSpellSaveRolls() []int          { return r.SpellSaveRolls }
func (r *SpellResult) GetSpellSaveTotal() int            { return r.SpellSaveTotal }
func (r *SpellResult) GetSpellSaveSuccess() bool         { return r.SpellSaveSuccess }
func (r *SpellResult) GetDamage() int                    { return r.Damage }
func (r *SpellResult) GetDamageRolls() []int             { return r.DamageRolls }
func (r *SpellResult) GetDamageType() string             { return r.DamageType }

func (s *SpellcastingManager) CastSpell(target core.Entity, spell *spells.SpellChoice) {
	switch spell.Spell.SpellType {
	case spells.STDamage:
		// Damage
	case spells.STHealing:
		// Healing

	}
}

func calculateDamage(req *SpellCastRequest, isCritical bool, options core.Options) (int, []int, error) {
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

func (s *SpellcastingManager) castDamageSpell(target core.Entity, req *SpellCastRequest, options core.Options) (*SpellResult, error) {
	var res SpellResult
	switch req.AttackData.SpellChoice.Spell.HasDC {
	case true:
		res.HasDC = true

		a, err := core.GetNormalizedAbility(req.AttackData.SpellChoice.Spell.SpellDC.Ability)
		if err != nil {
			return nil, err
		}
		saveTotal, saveRolls, err := target.MakeSavingThrow(a)
		if err != nil {
			return nil, err
		}

		res.SpellSaveTotal = saveTotal
		res.SpellSaveRolls = saveRolls

		if s.IsSaveSuccessful(saveTotal, s.parent.GetSpellSaveDC(a)) {
			res.SpellSaveSuccess = true
		} else {
			res.SpellSaveSuccess = false
		}

		// Calculate damage
		var damage int
		var damageRolls []int
		damage, damageRolls, err = calculateDamage(req, false, options)
		if err != nil {
			return nil, err
		}

		if res.SpellSaveSuccess {
			if req.AttackData.SpellChoice.Spell.SpellDC.OnSuccess == "none" {
				res.SpellTotalValue = 0
				res.Damage = 0
				res.DamageRolls = damageRolls
			} else if req.AttackData.SpellChoice.Spell.SpellDC.OnSuccess == "half" {
				res.SpellTotalValue = damage / 2
				res.Damage = damage / 2
				res.DamageRolls = damageRolls
			} else {
				fmt.Println("Other DC On success not implemented")
			}
		} else {
			res.SpellTotalValue = damage
			res.Damage = damage
			res.DamageRolls = damageRolls
		}
	case false:
		res.HasDC = false

		var attackTotal int
		var attackRolls []int
		attackTotal, attackRolls, err := core.AttackRoll(req.AttackData.AttackModifier, req.Advantage)
		if err != nil {
			return nil, err
		}

		// TODO: Note that halfling lucky does not use advantage again, re-roll lower
		if attackRolls == 1 && req.Modifiers.HalflingLucky {
			if req.Advantage == core.RollNormal {
				attackTotal, attackRolls, err = core.AttackRoll(req.AttackData.AttackModifier, req.Advantage)
				if err != nil {
					return nil, err
				}
			} else {

			}
		}

		events.LogDiceRollEvent(s.parent, attackTotal, []int{attackRolls}, core.DiceRollAttack, req.AttackData.AttackModifier, s.parent.GetEventListener())
		res.AttackRoll = attackRolls
		res.AttackTotal = attackTotal

		critThreshold := 20
		isCrit := core.IsCriticalHit(attackRolls, critThreshold)
		if (isCrit || core.DoesAttackHit(attackRolls, target.GetAC())) && attackRolls != 1 {
			res.IsHit = true
			res.IsCriticalHit = isCrit

			events.LogSpellAttackEvent(s.parent, target, &res, s.parent.GetEventListener())

			damage, rolls, err2 := calculateDamage(req, isCrit, options)
			if err2 != nil {
				return nil, err2
			}
			res.Damage = damage
			res.SpellTotalValue = damage
			res.DamageRolls = rolls
			res.DamageType = req.AttackData.SpellChoice.Formula.DamageType

		}
	default:
		return nil, errors.New("invalid spell")
	}

	res.ActorName = s.parent.GetName()
	res.TargetName = target.GetName()
	res.SpellName = req.AttackData.SpellChoice.Spell.Name
	res.SpellLevel = req.AttackData.SpellChoice.Spell.Level

	return &res, nil
}

func (s *SpellcastingManager) castHealingSpell(target core.Entity, spell *spells.SpellChoice) error {
	var result SpellResult

}

func (s *SpellcastingManager) IsSaveSuccessful(roll int, dc int) bool {
	return roll >= dc
}
