package core

import "fmt"

type DiceRollType string

const (
	DiceRollGeneral      DiceRollType = "general"
	DiceRollAttack       DiceRollType = "attack"
	DiceRollDamage       DiceRollType = "damage"
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

type DCOnSuccess string

const (
	DCOnSuccessNone DCOnSuccess = "none"
	DCOnSuccessHalf DCOnSuccess = "half"
)

func (dcs DCOnSuccess) String() string {
	return string(dcs)
}

type SpellType string

const (
	STDamage  SpellType = "damage"
	STHealing SpellType = "healing"
)

func (st SpellType) String() string {
	return string(st)
}

type CasterType string

const (
	CasterCharacter         CasterType = "character"
	CasterMonsterInnate     CasterType = "innate_monster"
	CasterMonsterTrueCaster CasterType = "spellcaster_monster"
)

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
