package core

import (
	"fmt"
	"strings"
)

type DiceRollType string

const (
	DiceRollGeneral      DiceRollType = "general"
	DiceRollAttack       DiceRollType = "attack"
	DiceRollDamage       DiceRollType = "damage"
	DiceRollHealing      DiceRollType = "healing"
	DiceRollInitiative   DiceRollType = "initiative"
	DiceRollAbilityCheck DiceRollType = "ability check"
	DiceRollSavingThrow  DiceRollType = "saving throw"
)

type DiceType int

const (
	D4   DiceType = 4
	D6   DiceType = 6
	D8   DiceType = 8
	D10  DiceType = 10
	D12  DiceType = 12
	D20  DiceType = 20
	D100 DiceType = 100
)

func (dt DiceType) String() string {
	return fmt.Sprintf("%d", int(dt))
}

func (dt DiceType) Int() int {
	return int(dt)
}

func (dt DiceType) Max() int {
	return int(dt)
}

func (dt DiceType) Min() int {
	return 1
}

func (dt DiceType) Avg() float64 {
	return (float64(dt) / 2) + 0.5
}

func (dt DiceType) IsValid() bool {
	switch dt {
	case D4, D6, D8, D10, D12, D20, D100:
		return true
	}
	return false
}

type AdvantageType int

const (
	RollNormal AdvantageType = iota
	RollAdvantage
	RollDisadvantage
)

func (at AdvantageType) String() string {
	switch at {
	case RollNormal:
		return "Normal"
	case RollAdvantage:
		return "Advantage"
	case RollDisadvantage:
		return "Disadvantage"
	default:
		return "invalid"
	}
}

type DamageType string

const (
	DamageAcid        DamageType = "acid"
	DamageCold        DamageType = "cold"
	DamageFire        DamageType = "fire"
	DamageForce       DamageType = "force"
	DamageLightning   DamageType = "lightning"
	DamageNecrotic    DamageType = "necrotic"
	DamagePoison      DamageType = "poison"
	DamagePsychic     DamageType = "psychic"
	DamageRadiant     DamageType = "radiant"
	DamageThunder     DamageType = "thunder"
	DamageSlashing    DamageType = "slashing"
	DamageBludgeoning DamageType = "bludgeoning"
	DamagePiercing    DamageType = "piercing"
)

func (dt DamageType) String() string {
	return string(dt)
}

func MakeDamageType(s string) (DamageType, error) {
	switch strings.ToLower(s) {
	case "acid":
		return DamageAcid, nil
	case "cold":
		return DamageCold, nil
	case "fire":
		return DamageFire, nil
	case "force":
		return DamageForce, nil
	case "lightning":
		return DamageLightning, nil
	case "necrotic":
		return DamageNecrotic, nil
	case "poison":
		return DamagePoison, nil
	case "psychic":
		return DamagePsychic, nil
	case "radiant":
		return DamageRadiant, nil
	case "thunder":
		return DamageThunder, nil
	case "slashing":
		return DamageSlashing, nil
	case "bludgeoning":
		return DamageBludgeoning, nil
	case "piercing":
		return DamagePiercing, nil
	default:
		return DamageAcid, fmt.Errorf("invalid damage type")
	}
}

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
	CasterMonsterInnate     CasterType = "innate_monster"
	CasterMonsterTrueCaster CasterType = "spellcaster_monster"
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
