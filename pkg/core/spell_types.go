package core

import (
	"fmt"
	"strings"
)

type DCOnSuccess string

const (
	DCOnSuccessNone  DCOnSuccess = "none"
	DCOnSuccessHalf  DCOnSuccess = "half"
	DCOnSuccessOther DCOnSuccess = "other"
)

func (dcs DCOnSuccess) String() string {
	return string(dcs)
}

func NewDCOnSuccess(s string) (DCOnSuccess, error) {
	switch strings.ToLower(s) {
	case "none":
		return DCOnSuccessNone, nil
	case "half":
		return DCOnSuccessHalf, nil
	case "other":
		return DCOnSuccessOther, nil
	default:
		return DCOnSuccessNone, fmt.Errorf("invalid DCOnSuccess")
	}
}

type SpellType string

const (
	STDamage  SpellType = "damage"
	STHealing SpellType = "healing"
)

func (st SpellType) String() string {
	return string(st)
}

func NewSpellType(s string) (SpellType, error) {
	switch strings.ToLower(s) {
	case "damage":
		return STDamage, nil
	case "healing":
		return STHealing, nil
	default:
		return STDamage, fmt.Errorf("invalid spell type")
	}
}

type CasterType string

const (
	CasterCharacter         CasterType = "character"
	CasterMonsterInnate     CasterType = "innate_monster"
	CasterMonsterTrueCaster CasterType = "spellcaster_monster"
)

func (ct CasterType) String() string {
	return string(ct)
}

func NewCasterType(s string) (CasterType, error) {
	switch strings.ToLower(s) {
	case "character":
		return CasterCharacter, nil
	case "innate_monster":
		return CasterMonsterInnate, nil
	case "spellcaster_monster":
		return CasterMonsterTrueCaster, nil
	default:
		return CasterCharacter, fmt.Errorf("invalid caster type")
	}
}

type CastingTime string

const (
	CastingTimeAction   CastingTime = "action"
	CastingTimeBonus    CastingTime = "bonus"
	CastingTimeReaction CastingTime = "reaction"
	CastingTimeInstant  CastingTime = "instant"
)

func (ct CastingTime) String() string {
	return string(ct)
}

func NewCastingTime(s string) (CastingTime, error) {
	switch strings.ToLower(s) {
	case "action":
		return CastingTimeAction, nil
	case "bonus":
		return CastingTimeBonus, nil
	case "reaction":
		return CastingTimeReaction, nil
	default:
		return CastingTimeAction, fmt.Errorf("invalid casting time")
	}
}

type CastFormula struct {
	CastLevel    int
	NumberOfDice int
	Die          DiceType
	AmountToAdd  int
	UseSpellmod  bool // UseSpellmod specifies whether the spell modifier should be added to the calculated damage.
	DamageType   DamageType
	AverageValue int
}

func (f CastFormula) GetCastLevel() int         { return f.CastLevel }
func (f CastFormula) GetNumberOfDice() int      { return f.NumberOfDice }
func (f CastFormula) GetDie() DiceType          { return f.Die }
func (f CastFormula) GetAmountToAdd() int       { return f.AmountToAdd }
func (f CastFormula) GetUseSpellModifier() bool { return f.UseSpellmod }
func (f CastFormula) GetDamageType() DamageType { return f.DamageType }
func (f CastFormula) GetAverageValue() int      { return f.AverageValue }

type SpellChoice struct {
	Spell   Spell
	Formula *CastFormula
}

func (sc SpellChoice) GetSpell() Spell {
	return sc.Spell
}

func (sc SpellChoice) GetFormula() *CastFormula {
	return sc.Formula
}
