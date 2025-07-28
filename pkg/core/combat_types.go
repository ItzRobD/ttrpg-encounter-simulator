package core

import (
	"fmt"
	"strings"
)

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

type EntityConditions map[Condition]bool

type Condition string

const (
	ConditionNone          Condition = "none"
	ConditionBlinded       Condition = "blinded"
	ConditionCharmed       Condition = "charmed"
	ConditionDeafened      Condition = "deafened"
	ConditionFrightened    Condition = "frightened"
	ConditionGrappled      Condition = "grappled"
	ConditionIncapacitated Condition = "incapacitated"
	ConditionInvisible     Condition = "invisible"
	ConditionParalyzed     Condition = "paralyzed"
	ConditionPetrified     Condition = "petrified"
	ConditionPoisoned      Condition = "poisoned"
	ConditionProne         Condition = "prone"
	ConditionRestrained    Condition = "restrained"
	ConditionStunned       Condition = "stunned"
	ConditionUnconscious   Condition = "unconscious"
)

func (c Condition) String() string {
	return string(c)
}

func NewEntityConditions() EntityConditions {
	return map[Condition]bool{
		ConditionBlinded:       false,
		ConditionCharmed:       false,
		ConditionDeafened:      false,
		ConditionFrightened:    false,
		ConditionGrappled:      false,
		ConditionIncapacitated: false,
		ConditionInvisible:     false,
		ConditionParalyzed:     false,
		ConditionPetrified:     false,
		ConditionPoisoned:      false,
		ConditionProne:         false,
		ConditionRestrained:    false,
		ConditionStunned:       false,
		ConditionUnconscious:   false,
	}
}

func NewCondition(s string) (Condition, error) {
	switch strings.ToLower(s) {
	case "blinded":
		return ConditionBlinded, nil
	case "charmed":
		return ConditionCharmed, nil
	case "deafened":
		return ConditionDeafened, nil
	case "frightened":
		return ConditionFrightened, nil
	case "grappled":
		return ConditionGrappled, nil
	case "incapacitated":
		return ConditionIncapacitated, nil
	case "invisible":
		return ConditionInvisible, nil
	case "paralyzed":
		return ConditionParalyzed, nil
	case "petrified":
		return ConditionPetrified, nil
	case "poisoned":
		return ConditionPoisoned, nil
	case "prone":
		return ConditionProne, nil
	case "restrained":
		return ConditionRestrained, nil
	case "stunned":
		return ConditionStunned, nil
	case "unconscious":
		return ConditionUnconscious, nil
	default:
		return ConditionNone, fmt.Errorf("invalid condition")
	}
}

type ExhaustionLevel int

func MakeExhaustionLevel(i int) (ExhaustionLevel, error) {
	if i < 1 || i > 6 {
		return 0, fmt.Errorf("exhaustion level must be between 1 and 6")
	}
	return ExhaustionLevel(i), nil
}

func (el ExhaustionLevel) String() string {
	return fmt.Sprintf("%d", int(el))
}

func (el ExhaustionLevel) Value() int {
	return int(el)
}

type DamageResistances map[DamageType]DamageResistance

type DamageResistance struct {
	Resistance ResistanceType
	Breakers   []ResistBreaker
}

func NewEmptyDamageResistance() DamageResistance {
	return DamageResistance{
		Resistance: ResistanceNone,
	}
}

func NewDamageResistance(rt ResistanceType, rb []ResistBreaker) DamageResistance {
	return DamageResistance{
		Resistance: rt,
		Breakers:   rb,
	}
}

type ResistanceType string

const (
	ResistanceNone       ResistanceType = "none"
	ResistanceResistant  ResistanceType = "resist"
	ResistanceVulnerable ResistanceType = "vulnerable"
	ResistanceImmune     ResistanceType = "immune"
)

func (rt ResistanceType) String() string {
	return string(rt)
}

func MakeResistanceType(s string) (ResistanceType, error) {
	switch strings.ToLower(s) {
	case "none":
		return ResistanceNone, nil
	case "resistant":
		return ResistanceResistant, nil
	case "vulnerable":
		return ResistanceVulnerable, nil
	case "immune":
		return ResistanceImmune, nil
	default:
		return ResistanceNone, fmt.Errorf("invalid resistance type")
	}
}

func NewDamageResistances() DamageResistances {
	return map[DamageType]DamageResistance{
		DamageAcid:        NewEmptyDamageResistance(),
		DamageCold:        NewEmptyDamageResistance(),
		DamageFire:        NewEmptyDamageResistance(),
		DamageForce:       NewEmptyDamageResistance(),
		DamageLightning:   NewEmptyDamageResistance(),
		DamageNecrotic:    NewEmptyDamageResistance(),
		DamagePoison:      NewEmptyDamageResistance(),
		DamagePsychic:     NewEmptyDamageResistance(),
		DamageRadiant:     NewEmptyDamageResistance(),
		DamageThunder:     NewEmptyDamageResistance(),
		DamageSlashing:    NewEmptyDamageResistance(),
		DamageBludgeoning: NewEmptyDamageResistance(),
		DamagePiercing:    NewEmptyDamageResistance(),
	}
}

func (dr DamageResistances) GetResistanceType(dt DamageType) ResistanceType {
	if rt, ok := dr[dt]; ok {
		return rt.Resistance
	}
	return ResistanceNone
}

func (dr DamageResistances) SetResistanceType(dt DamageType, rt ResistanceType) {
	res := dr[dt]
	res.Resistance = rt
	dr[dt] = res
}

func (dr DamageResistances) SetPhysicalResistance(rt ResistanceType) {
	res := dr[DamageSlashing]
	res.Resistance = rt
	dr[DamageSlashing] = res
	res = dr[DamageBludgeoning]
	res.Resistance = rt
	dr[DamageBludgeoning] = res
	res = dr[DamagePiercing]
	res.Resistance = rt
	dr[DamagePiercing] = res
}

func (dr DamageResistances) AddBreaker(dt DamageType, rb ResistBreaker) {
	res := dr[dt]
	res.Breakers = append(res.Breakers, rb)
	dr[dt] = res
}

func (dr DamageResistances) RemoveBreaker(dt DamageType, rb ResistBreaker) {
	res := dr[dt]
	for i, br := range res.Breakers {
		if br == rb {
			res.Breakers = append(res.Breakers[:i], res.Breakers[i+1:]...)
			break
		}
	}
}

func (dr DamageResistances) GetBreakers(dt DamageType) []ResistBreaker {
	if res, ok := dr[dt]; ok {
		return res.Breakers
	}
	return nil
}

type ResistBreaker string

const (
	ResistBreakerNone           ResistBreaker = "none"
	ResistBreakerMagic          ResistBreaker = "magic"
	ResistBreakerSilvered       ResistBreaker = "silvered"
	ResistBreakerAdamantine     ResistBreaker = "adamantine"
	ResistBreakerColdForgedIron ResistBreaker = "cold forged iron"
)

func (rb ResistBreaker) String() string {
	return string(rb)
}

func MakeResistBreaker(s string) (ResistBreaker, error) {
	switch strings.ToLower(s) {
	case "none", "":
		return ResistBreakerNone, nil
	case "magic":
		return ResistBreakerMagic, nil
	case "silvered":
		return ResistBreakerSilvered, nil
	case "adamantine":
		return ResistBreakerAdamantine, nil
	case "cold forged iron":
		return ResistBreakerColdForgedIron, nil
	default:
		return ResistBreakerNone, fmt.Errorf("invalid resistance breaker")
	}
}

type AttackData struct {
	Name              string
	NumberOfDice      int
	Die               DiceType
	AttackModifier    int // Added to attack roll. Character: Proficiency + Ability Mod; Monster: To Hit Bonus
	DamageModifier    int
	DamageType        DamageType
	IsVersatileAttack bool
}

func (ad AttackData) GetAttackName() string      { return ad.Name }
func (ad AttackData) GetNumberOfDice() int       { return ad.NumberOfDice }
func (ad AttackData) GetDie() DiceType           { return ad.Die }
func (ad AttackData) GetAttackModifier() int     { return ad.AttackModifier }
func (ad AttackData) GetDamageModifier() int     { return ad.DamageModifier }
func (ad AttackData) GetDamageType() string      { return ad.DamageType.String() }
func (ad AttackData) GetIsVersatileAttack() bool { return ad.IsVersatileAttack }

type CombatContext struct {
	AllCombatants  map[int]Combatant
	NeedHealingIDs []int
	CurrentRound   int
	ActingEntityID int

	// Combat options
	AllowCharacterHeals       bool
	AllMonsterHeals           bool
	AOEHitsAllEnemies         bool
	CharacterHealThresholdPct int
	MonsterHealThresholdPct   int
}

type CombatContextConfig struct {
	AllCombatants  map[int]Combatant
	NeedHealingIDs []int
	CurrentRound   int
	ActingEntityID int
} // ???

func NewCombatContext(config CombatContextConfig) CombatContext {
	// TODO: Finish implementation of this
	return CombatContext{
		//AllCombatants:  combatants,
		//NeedHealingIDs: needHealing,
		//CurrentRound:   round,
	}
}

type ActionOutcome struct {
	ActionType ActionType
	TargetID   int
	ActorID    int
	Success    bool
	Effects    []Effect
}

type Effect struct {
	Type       EffectType
	Value      int
	DamageType DamageType
	Condition  *Condition
}

type EffectType string

const (
	EffectDamage    EffectType = "damage"
	EffectHealing   EffectType = "healing"
	EffectCondition EffectType = "condition"
	EffectTempHP    EffectType = "temp_hp"
)
