package core

import (
	"fmt"
	"strings"
)

type SpellType string

const (
	STDamage  SpellType = "damage"
	STHealing SpellType = "healing"
)

func (st SpellType) String() string {
	return string(st)
}

func MakeSpellType(s string) (SpellType, error) {
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
	CasterMonsterInnate     CasterType = "monster_innate"
	CasterMonsterTrueCaster CasterType = "monster_spellcaster"
)

func (ct CasterType) String() string {
	return string(ct)
}

func MakeCasterType(s string) (CasterType, error) {
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
	CastingTimeRitual   CastingTime = "ritual"
)

func (ct CastingTime) String() string {
	return string(ct)
}

func MakeCastingTime(s string) (CastingTime, error) {
	switch strings.ToLower(s) {
	case "action":
		return CastingTimeAction, nil
	case "bonus":
		return CastingTimeBonus, nil
	case "reaction":
		return CastingTimeReaction, nil
	case "instant":
		return CastingTimeInstant, nil
	case "ritual":
		return CastingTimeRitual, nil
	default:
		return CastingTimeAction, fmt.Errorf("invalid casting time")
	}
}

type CastFormula struct {
	CastLevel    int        `json:"cast_level"`
	NumberOfDice int        `json:"number_of_dice"`
	Die          DiceType   `json:"die"`
	AmountToAdd  int        `json:"amount_to_add"`
	UseSpellmod  bool       `json:"use_spellmod"` // UseSpellmod specifies whether the spell modifier should be added to the calculated damage.
	DamageType   DamageType `json:"damage_type"`
	AverageValue int        `json:"average_value"`
}

func (f CastFormula) GetCastLevel() int         { return f.CastLevel }
func (f CastFormula) GetNumberOfDice() int      { return f.NumberOfDice }
func (f CastFormula) GetDie() DiceType          { return f.Die }
func (f CastFormula) GetAmountToAdd() int       { return f.AmountToAdd }
func (f CastFormula) GetUseSpellModifier() bool { return f.UseSpellmod }
func (f CastFormula) GetDamageType() DamageType { return f.DamageType }
func (f CastFormula) GetAverageValue() int      { return f.AverageValue }

type DCOnSuccess string

const (
	DCOnSuccessNone  DCOnSuccess = "none"
	DCOnSuccessHalf  DCOnSuccess = "half"
	DCOnSuccessOther DCOnSuccess = "other"
)

func (dcs DCOnSuccess) String() string {
	return string(dcs)
}

func MakeDCOnSuccess(s string) (DCOnSuccess, error) {
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
