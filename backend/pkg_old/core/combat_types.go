package core

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type DamageModificationResult struct {
	OriginalValue    int            `json:"original_value"`
	FinalValue       int            `json:"final_value"`
	WasModified      bool           `json:"was_modified"`
	ResistanceType   ResistanceType `json:"resistance_type"` // None, Resistant, Vulnerable, Immune
	ResistanceBroken bool           `json:"resistance_broken"`
}

type VictoryStatus string

const (
	VictoryStatusNone       VictoryStatus = "none"
	VictoryStatusCharacters VictoryStatus = "characters"
	VictoryStatusMonsters   VictoryStatus = "monsters"
)

// TargetStatus indicates the result of a target selection attempt.
// TargetOK: a valid target id was chosen
// TargetNone: there were no valid targets to choose from
// TargetInvalidType: the selection strategy or inputs were invalid
type TargetStatus int

const (
	TargetOK TargetStatus = iota
	TargetNone
	TargetInvalidType
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
	DamageNone        DamageType = "none"
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
	// Applied when a Barbarian (or similar feature) uses Reckless Attack; everyone has advantage to hit them
	ConditionRecklessExposed Condition = "reckless_exposed"
)

func (c Condition) String() string {
	return string(c)
}

func NewEntityConditions() EntityConditions {
	return map[Condition]bool{
		ConditionBlinded:         false,
		ConditionCharmed:         false,
		ConditionDeafened:        false,
		ConditionFrightened:      false,
		ConditionGrappled:        false,
		ConditionIncapacitated:   false,
		ConditionInvisible:       false,
		ConditionParalyzed:       false,
		ConditionPetrified:       false,
		ConditionPoisoned:        false,
		ConditionProne:           false,
		ConditionRestrained:      false,
		ConditionStunned:         false,
		ConditionUnconscious:     false,
		ConditionRecklessExposed: false,
	}
}

func (ec EntityConditions) Add(c Condition) {
	ec[c] = true
}

func (ec EntityConditions) Remove(c Condition) {
	delete(ec, c)
}

func (ec EntityConditions) Has(c Condition) bool {
	return ec[c]
}

func (ec EntityConditions) Clear() {
	for c := range ec {
		ec[c] = false
	}
}

func (ec EntityConditions) GetActive() []Condition {
	var active []Condition
	for c := range ec {
		if ec[c] {
			active = append(active, c)
		}
	}
	return active
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
	case "reckless_exposed":
		return ConditionRecklessExposed, nil
	default:
		return ConditionNone, fmt.Errorf("invalid condition")
	}
}

type ConditionEffect struct {
	AutoFailStrDexSave  bool
	OutgoingAttackRoll  AdvantageType
	IncomingAttackRoll  AdvantageType
	SavingThrow         map[Ability]AdvantageType
	HasTempResistances  bool
	TemporaryResistance DamageResistances
}

func GetConditionEffects(c Condition) ConditionEffect {
	var e ConditionEffect
	e.TemporaryResistance = NewDamageResistances()
	// initialize map fields that might be assigned into
	e.SavingThrow = make(map[Ability]AdvantageType)
	switch c {
	case ConditionBlinded:
		e.OutgoingAttackRoll = RollDisadvantage
		e.IncomingAttackRoll = RollAdvantage
	case ConditionCharmed:
		break
	case ConditionDeafened:
		break
	case ConditionFrightened:
		e.OutgoingAttackRoll = RollDisadvantage
	case ConditionGrappled:
		break
	case ConditionIncapacitated:
		break
	case ConditionInvisible:
		e.OutgoingAttackRoll = RollAdvantage
		e.IncomingAttackRoll = RollDisadvantage
	case ConditionParalyzed:
		e.AutoFailStrDexSave = true
		e.IncomingAttackRoll = RollAdvantage
	case ConditionPetrified:
		e.AutoFailStrDexSave = true
		e.IncomingAttackRoll = RollAdvantage
		e.HasTempResistances = true
		e.TemporaryResistance = NewDamageResistancesAll(ResistanceResistant)
	case ConditionPoisoned:
		e.OutgoingAttackRoll = RollDisadvantage
	case ConditionProne:
		e.OutgoingAttackRoll = RollDisadvantage
		e.IncomingAttackRoll = RollNormal
	case ConditionRestrained:
		e.OutgoingAttackRoll = RollDisadvantage
		e.IncomingAttackRoll = RollAdvantage
		e.SavingThrow[AbilityDexterity] = RollDisadvantage
	case ConditionStunned:
		e.AutoFailStrDexSave = true
		e.IncomingAttackRoll = RollAdvantage
	case ConditionUnconscious:
		e.AutoFailStrDexSave = true
		e.IncomingAttackRoll = RollAdvantage
	case ConditionRecklessExposed:
		// All incoming attack rolls have advantage against this creature
		e.IncomingAttackRoll = RollAdvantage
	default:
		return ConditionEffect{}
	}

	return e
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
	Resistance ResistanceType  `json:"resistance"`
	Breakers   []ResistBreaker `json:"breakers"`
}

func NewEmptyDamageResistance() DamageResistance {
	return DamageResistance{
		Resistance: ResistanceNone,
		Breakers:   make([]ResistBreaker, 0),
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

func NewDamageResistancesAll(rt ResistanceType) DamageResistances {
	return map[DamageType]DamageResistance{
		DamageAcid:        NewDamageResistance(rt, nil),
		DamageCold:        NewDamageResistance(rt, nil),
		DamageFire:        NewDamageResistance(rt, nil),
		DamageForce:       NewDamageResistance(rt, nil),
		DamageLightning:   NewDamageResistance(rt, nil),
		DamageNecrotic:    NewDamageResistance(rt, nil),
		DamagePoison:      NewDamageResistance(rt, nil),
		DamagePsychic:     NewDamageResistance(rt, nil),
		DamageRadiant:     NewDamageResistance(rt, nil),
		DamageThunder:     NewDamageResistance(rt, nil),
		DamageSlashing:    NewDamageResistance(rt, nil),
		DamageBludgeoning: NewDamageResistance(rt, nil),
		DamagePiercing:    NewDamageResistance(rt, nil),
	}
}

func (dr DamageResistances) GetResistanceType(dt DamageType) ResistanceType {
	if rt, ok := dr[dt]; ok {
		return rt.Resistance
	}
	return ResistanceNone
}

func (dr DamageResistances) GetResistance(dt DamageType) DamageResistance {
	if res, ok := dr[dt]; ok {
		return res
	}
	// Return a safe default instead of zero-value so callers never see empty ResistanceType
	return NewEmptyDamageResistance()
}

func (dr DamageResistances) SetResistance(d DamageType, rt ResistanceType, rb []ResistBreaker) {
	dr[d] = NewDamageResistance(rt, rb)
}

func (dr DamageResistances) SetResistanceType(dt DamageType, rt ResistanceType) {
	res := dr[dt]
	res.Resistance = rt
	dr[dt] = res
}

func (dr DamageResistances) ResetResistance(dt DamageType) {
	dr[dt] = NewEmptyDamageResistance()
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

func (dr DamageResistances) DamageTypeContainsBreaker(dt DamageType, rb ResistBreaker) bool {
	if res, ok := dr[dt]; ok {
		for _, br := range res.Breakers {
			if br == rb {
				return true
			}
		}
	}
	return false
}

func (dr DamageResistances) DamageTypeContainsAllBreakers(dt DamageType, rb []ResistBreaker) bool {
	res, ok := dr[dt]
	if !ok {
		return false
	}

	for _, br := range rb {
		found := false
		for _, existingBreaker := range res.Breakers {
			if existingBreaker == br {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
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

type DamageBlock struct {
	NumberOfDice int        `json:"number_of_dice"`
	Die          DiceType   `json:"die"`
	DamageType   DamageType `json:"damage_type"`
	Modifier     int        `json:"modifier"` // Flat bonus for this specific block
}

type AttackData struct {
	Name              string          `json:"name"`
	DamageBlocks      []DamageBlock   `json:"damage_blocks"`
	AttackModifier    int             `json:"attack_modifier"`
	AbilityUsed       Ability         `json:"ability_used"`
	DamageModifier    int             `json:"damage_modifier"` // Global modifier (e.g. Strength)
	ResistBreakers    []ResistBreaker `json:"resist_breakers"`
	IsVersatileAttack bool            `json:"is_versatile_attack"`
	IsRangedWeapon    bool            `json:"is_ranged_weapon"`
	IsTwoHandedWeapon bool            `json:"is_two_handed_weapon"`
	IsFinesseWeapon   bool            `json:"is_finesse_weapon"`
	IsLightWeapon     bool            `json:"is_light_weapon"`
	IsThrownWeapon    bool            `json:"is_thrown_weapon"`
	IsOnlyRanged      bool            `json:"is_only_ranged"`
	IsHeavyWeapon     bool            `json:"is_heavy_weapon"`
	Average           int             `json:"average"`
	WeaponAttackBonus int             `json:"weapon_attack_bonus"`
	WeaponDamageBonus int             `json:"weapon_damage_bonus"`
}

func (ad AttackData) GetAttackName() string { return ad.Name }
func (ad AttackData) GetNumberOfDice() int {
	if len(ad.DamageBlocks) > 0 {
		return ad.DamageBlocks[0].NumberOfDice
	}
	return 0
}
func (ad AttackData) GetDie() DiceType {
	if len(ad.DamageBlocks) > 0 {
		return ad.DamageBlocks[0].Die
	}
	return D0
}
func (ad AttackData) GetAttackModifier() int { return ad.AttackModifier }
func (ad AttackData) GetDamageModifier() int { return ad.DamageModifier }
func (ad AttackData) GetDamageType() string {
	if len(ad.DamageBlocks) > 0 {
		return ad.DamageBlocks[0].DamageType.String()
	}
	return ""
}
func (ad AttackData) GetIsVersatileAttack() bool         { return ad.IsVersatileAttack }
func (ad AttackData) GetResistBreakers() []ResistBreaker { return ad.ResistBreakers }
func (ad AttackData) GetAverage() int                    { return ad.Average }

type AttackRequest struct {
	AttackData        []AttackData
	AttackOptions     AttackOptions
	SimulationOptions *SimulationOptions
	Target            Entity
}

type HealSource int

const (
	HealSourceSpell HealSource = iota
	HealSourceLayingOnHands
)

type HealRequest struct {
	Source       HealSource
	Target       Entity
	SpellChoice  *SpellChoice  // Used if Source == HealSourceSpell
	AbilityValue int           // Used if Source == HealSourceLayingOnHands
	Advantage    AdvantageType // For healing spells that might need it (rare)
}

type SneakAttackParams struct {
	IsCritical bool
	Advantage  AdvantageType
	DamageType DamageType
	IsRanged   bool
	IsSpell    bool
}

func (ar *AttackRequest) GetAttackData() []AttackData              { return ar.AttackData }
func (ar *AttackRequest) GetAttackOptions() AttackOptions          { return ar.AttackOptions }
func (ar *AttackRequest) GetSimulationOptions() *SimulationOptions { return ar.SimulationOptions }
func (ar *AttackRequest) GetTarget() Entity                        { return ar.Target }

type AttackResult struct {
	ActorName      string          `json:"actor_name"`
	TargetName     string          `json:"target_name"`
	Target         Entity          `json:"-"`
	AttackName     string          `json:"attack_name"`
	AttackCount    int             `json:"attack_count"`
	TargetValue    int             `json:"target_value"`
	IsHit          bool            `json:"is_hit"`
	IsCriticalHit  bool            `json:"is_critical_hit"`
	AttackTotal    int             `json:"attack_total"`
	AttackRoll     int             `json:"attack_roll"`
	DamageRoll     RollResult      `json:"damage_roll"`
	DamageType     DamageType      `json:"damage_type"`
	ResistBreakers []ResistBreaker `json:"resist_breakers"`
	IsRanged       bool            `json:"is_ranged"`
	AdvantageUsed  AdvantageType   `json:"advantage_used"`
}

func (r AttackResult) GetActorName() string        { return r.ActorName }
func (r AttackResult) GetTargetName() string       { return r.TargetName }
func (r AttackResult) GetTarget() Entity           { return r.Target }
func (r AttackResult) GetAttackName() string       { return r.AttackName }
func (r AttackResult) GetAttackCount() int         { return r.AttackCount }
func (r AttackResult) GetIsHit() bool              { return r.IsHit }
func (r AttackResult) GetIsCriticalHit() bool      { return r.IsCriticalHit }
func (r AttackResult) GetAttackTotal() int         { return r.AttackTotal }
func (r AttackResult) GetAttackRoll() int          { return r.AttackRoll }
func (r AttackResult) GetDamageResult() RollResult { return r.DamageRoll }
func (r AttackResult) GetDamageType() DamageType   { return r.DamageType }
func (r AttackResult) GetTargetValue() int         { return r.TargetValue }

type AttackOptions struct {
	Advantage            AdvantageType `json:"advantage"`
	NumberOfAttacks      int           `json:"number_of_attacks"`
	BonusToAttackRoll    int           `json:"bonus_to_attack_roll"`
	BonusToDamageRoll    int           `json:"bonus_to_damage_roll"`
	ShouldApplyDamageMod bool          `json:"should_apply_damage_mod"`
	PowerAttack          bool          `json:"power_attack"`
	ImprovedCritical     bool          `json:"improved_critical"`
	RerollOnesAndTwos    bool          `json:"reroll_ones_and_twos"`
	ExtraCritDice        int           `json:"extra_crit_dice"`
}

func (ao AttackOptions) GetAdvantage() AdvantageType   { return ao.Advantage }
func (ao AttackOptions) GetNumberOfAttacks() int       { return ao.NumberOfAttacks }
func (ao AttackOptions) GetBonusToAttackRoll() int     { return ao.BonusToAttackRoll }
func (ao AttackOptions) GetBonusToDamageRoll() int     { return ao.BonusToDamageRoll }
func (ao AttackOptions) GetShouldApplyDamageMod() bool { return ao.ShouldApplyDamageMod }
func (ao AttackOptions) GetIsPowerAttack() bool        { return ao.PowerAttack }
func (ao AttackOptions) GetIsImprovedCritical() bool   { return ao.ImprovedCritical }
func (ao AttackOptions) GetTreatOnesAsTwos() bool      { return ao.RerollOnesAndTwos }

type CombatContext struct {
	CombatantInfo           map[int]*CombatantInfo
	LegendaryCreatures      map[int]bool
	ConsciousCharacterCount int
	ConsciousMonsterCount   int

	// Lookup lists
	CharactersInNeedOfHealing []int
	MonstersInNeedOfHealing   []int
	DeadCombatants            []int

	// Turn Tracking
	TurnOrder        []int
	CurrentTurnIndex int
	CurrentRound     int
	ActingEntityID   int

	// Premium AI global tracking
	MaxDamageSeen int // Highest damage dealt in a single instance during this combat

	// Combat options
	Options *SimulationOptions
}

// NewCombatContext creates a new CombatContext. Pass the shared SimulationOptions pointer
// so all readers observe the same configuration during combat.
func NewCombatContext(options *SimulationOptions) *CombatContext {
	return &CombatContext{
		CombatantInfo:      make(map[int]*CombatantInfo),
		LegendaryCreatures: make(map[int]bool),
		Options:            options,
	}
}

// Opt returns a non-nil SimulationOptions pointer for safe access to combat options.
// If the context or options are nil, it returns a pointer to a zero-value SimulationOptions.
func (cc *CombatContext) Opt() *SimulationOptions {
	if cc == nil || cc.Options == nil {
		return &SimulationOptions{}
	}
	return cc.Options
}

type ActionOutcome struct {
	ActionType      ActionType     `json:"action_type"`
	TargetID        int            `json:"target_id"`
	ActorID         int            `json:"actor_id"`
	Success         bool           `json:"success"`
	Effects         []Effect       `json:"effects"`
	IsAOE           bool           `json:"is_aoe"`
	IsConcentration bool           `json:"is_concentration"`
	SpellName       string         `json:"spell_name"`
	AttackResults   []AttackResult `json:"attack_results"`
}

type Effect struct {
	Type           EffectType      `json:"type"`
	Value          int             `json:"value"`      // Calculated value (after primary target save if applicable)
	BaseValue      int             `json:"base_value"` // Base value (before any saves/resistances)
	DamageType     DamageType      `json:"damage_type"`
	ResistBreakers []ResistBreaker `json:"resist_breakers"`
	Condition      *Condition      `json:"condition"`
	SaveCtx        *SaveContext    `json:"save_ctx,omitempty"`
	AttackCtx      *AttackContext  `json:"attack_ctx,omitempty"`
	SpellCtx       *SpellContext   `json:"spell_ctx,omitempty"`
	SourceRollID   string          `json:"source_roll_id"`
}

type SpellContext struct {
	SpellLevel int `json:"spell_level"`
}

type SaveContext struct {
	Ability   Ability     `json:"ability"`
	TargetDC  int         `json:"target_dc"`
	Success   bool        `json:"success"`
	OnSuccess DCOnSuccess `json:"on_success"`
}

type AttackContext struct {
	IsRanged   bool `json:"is_ranged"`
	IsCritical bool `json:"is_critical"`
}

// EffectType represents the classification of an effect, such as damage, healing, condition, or temporary hit points.
type EffectType string

const (
	EffectDamage    EffectType = "damage"
	EffectHealing   EffectType = "healing"
	EffectCondition EffectType = "condition"
	EffectTempHP    EffectType = "temp_hp"
)

type CombatantInfo struct {
	Combatant *Combatant `json:"-"`

	// Current state
	State *CombatantState `json:"state"`

	// Historical data
	Statistics *CombatStatistics `json:"statistics"`

	// Capability/threat info
	UsedCapabilities *CombatantCapabilities `json:"used_capabilities"`
}

type CombatantState struct {
	CurrentHP        int                `json:"current_hp"`
	MaxHP            int                `json:"max_hp"`
	HealThreshold    int                `json:"heal_threshold"`
	Concentration    *ConcentrationInfo `json:"concentration"`
	HasActedThisTurn bool               `json:"has_acted_this_turn"`
	Conditions       EntityConditions   `json:"conditions"`
}

type ConcentrationInfo struct {
	IsConcentrating bool    `json:"is_concentrating"`
	SpellName       *string `json:"spell_name"`
	AffectedTargets []int   `json:"affected_targets"` // All targets affected by the concentration
	Duration        *int    `json:"duration"`
	RoundsRemaining *int    `json:"rounds_remaining"`
	RoundStarted    *int    `json:"round_started"`
}

type CombatStatistics struct {
	TotalDamageDealt       int     `json:"total_damage_dealt"`
	TotalHealingDone       int     `json:"total_healing_done"`
	DamageByRound          []int   `json:"damage_by_round"`
	HealingByRound         []int   `json:"healing_by_round"`
	LastDamageDealt        int     `json:"last_damage_dealt"`
	LastHealingDone        int     `json:"last_healing_done"`
	AverageDamagePerRound  float64 `json:"average_damage_per_round"`
	AverageHealingPerRound float64 `json:"average_healing_per_round"`

	// Attack patterns
	AttacksMade   int `json:"attacks_made"`
	AttacksHit    int `json:"attacks_hit"`
	AttacksMissed int `json:"attacks_missed"`
	CriticalHits  int `json:"critical_hits"`

	// Defensive stats
	TimesDamaged         int `json:"times_damaged"`
	TimesHealed          int `json:"times_healed"`
	TotalDamageTaken     int `json:"total_damage_taken"`
	TotalHealingReceived int `json:"total_healing_received"`
	DeathSaveSuccesses   int `json:"death_save_successes"`
	DeathSaveFailures    int `json:"death_save_failures"`

	// Premium AI Tracking
	LastAttackerID     int `json:"last_attacker_id"`      // id of the last entity to deal damage to this combatant
	TurnsSinceLastHeal int `json:"turns_since_last_heal"` // Turns since this entity last received healing
}

type CombatantCapabilities struct {
	KnownSpells []string `json:"known_spells"`

	// Tactical flags
	HasUsedCounterspell     bool `json:"has_used_counterspell"`
	HasUsedHealingSpells    bool `json:"has_used_healing_spells"`
	HasUsedAOE              bool `json:"has_used_aoe"`
	HasUsedRangedAttack     bool `json:"has_used_ranged_attack"`
	HasUsedMeleeAttack      bool `json:"has_used_melee_attack"`
	HasUsedLegendaryActions bool `json:"has_used_legendary_actions"`
}

// CombatantInfo methods

func NewCombatantInfo(combatant *Combatant) *CombatantInfo {
	return &CombatantInfo{
		Combatant: combatant,
		State:     &CombatantState{},
		Statistics: &CombatStatistics{
			AverageDamagePerRound:  0,
			AverageHealingPerRound: 0,
		},
		UsedCapabilities: &CombatantCapabilities{},
	}
}

// UpdateState refreshes the current state from the combatant entity
func (ci *CombatantInfo) UpdateState() {
	entity := ci.Combatant.Entity
	ci.State.CurrentHP = entity.GetHPStatus().GetHP()
	ci.State.MaxHP = entity.GetHPStatus().GetMaxHP()
	// Heal Threshold
	ci.State.Conditions = entity.GetConditions()

	// Sync concentration from entity state manager
	isConcentrating := entity.IsConcentrating()
	if ci.State.Concentration != nil {
		ci.State.Concentration.IsConcentrating = isConcentrating
		if !isConcentrating {
			ci.State.Concentration.SpellName = nil
			ci.State.Concentration.AffectedTargets = nil
			ci.State.Concentration.Duration = nil
			ci.State.Concentration.RoundsRemaining = nil
			ci.State.Concentration.RoundStarted = nil
		}
	} else if isConcentrating {
		// If entity says it's concentrating but we don't have info,
		// we might need to initialize it, but we lack details here.
		// For now, assume CombatEngine/StartConcentration handles the heavy lifting.
		ci.State.Concentration = &ConcentrationInfo{
			IsConcentrating: true,
		}
	}
}

// StartConcentration begins concentration on a spell/ability
func (ci *CombatantInfo) StartConcentration(spellName string, targets []int, duration int, currentRound int) {
	ci.State.Concentration = &ConcentrationInfo{
		IsConcentrating: true,
		SpellName:       &spellName,
		AffectedTargets: targets,
		Duration:        &duration,
		RoundsRemaining: &duration,
		RoundStarted:    &currentRound,
	}
	// Sync to entity
	ci.Combatant.Entity.SetConcentrating(true, spellName)
}

// BreakConcentration ends concentration (failed save, new concentration, etc.)
func (ci *CombatantInfo) BreakConcentration() {
	if ci.State.Concentration != nil {
		ci.State.Concentration.IsConcentrating = false
		ci.State.Concentration.SpellName = nil
		ci.State.Concentration.AffectedTargets = nil
		ci.State.Concentration.Duration = nil
		ci.State.Concentration.RoundsRemaining = nil
		ci.State.Concentration.RoundStarted = nil
	}
	// Sync to entity
	ci.Combatant.Entity.SetConcentrating(false, "")
}

// DecrementConcentrationRounds decrements rounds remaining
func (ci *CombatantInfo) DecrementConcentrationRounds() {
	if ci.State.Concentration != nil && ci.State.Concentration.IsConcentrating {
		*ci.State.Concentration.RoundsRemaining--
		if *ci.State.Concentration.RoundsRemaining <= 0 {
			ci.BreakConcentration()
		}
	}
}

// CombatStatistics methods

// RecordDamageDealt adds damage to statistics
func (cs *CombatStatistics) RecordDamageDealt(damage int, round int) {
	cs.TotalDamageDealt += damage
	cs.LastDamageDealt = damage

	// Ensure slice is large enough
	for len(cs.DamageByRound) <= round {
		cs.DamageByRound = append(cs.DamageByRound, 0)
	}
	cs.DamageByRound[round] += damage

	// Recalculate average
	cs.updateAverageDamage()
}

// RecordHealingDone adds healing to statistics
func (cs *CombatStatistics) RecordHealingDone(healing int, round int) {
	cs.TotalHealingDone += healing
	cs.LastHealingDone = healing

	// Ensure slice is large enough
	for len(cs.HealingByRound) <= round {
		cs.HealingByRound = append(cs.HealingByRound, 0)
	}
	cs.HealingByRound[round] += healing

	// Recalculate average
	cs.updateAverageHealing()
}

// RecordAttack logs an attack attempt
func (cs *CombatStatistics) RecordAttack(hit bool, critical bool) {
	cs.AttacksMade++
	if hit {
		cs.AttacksHit++
		if critical {
			cs.CriticalHits++
		}
	} else {
		cs.AttacksMissed++
	}
}

// RecordDamageTaken logs damage received
func (cs *CombatStatistics) RecordDamageTaken(value int) {
	cs.TotalDamageTaken += value
	cs.TimesDamaged++
}

// RecordHealingReceived logs healing received
func (cs *CombatStatistics) RecordHealingReceived(value int) {
	cs.TotalHealingReceived += value
	cs.TimesHealed++
}

// RecordDeathSave logs a death saving throw
func (cs *CombatStatistics) RecordDeathSave(success bool) {
	if success {
		cs.DeathSaveSuccesses++
	} else {
		cs.DeathSaveFailures++
	}
}

func (cs *CombatStatistics) updateAverageDamage() {
	roundsActive := len(cs.DamageByRound)
	if roundsActive > 0 {
		cs.AverageDamagePerRound = float64(cs.TotalDamageDealt) / float64(roundsActive)
	}
}

func (cs *CombatStatistics) updateAverageHealing() {
	roundsActive := len(cs.HealingByRound)
	if roundsActive > 0 {
		cs.AverageHealingPerRound = float64(cs.TotalHealingDone) / float64(roundsActive)
	}
}

// CombatantCapabilities methods

func (cc *CombatantCapabilities) AddKnownSpell(name string) {
	if cc.KnownSpells == nil {
		cc.KnownSpells = make([]string, 0)
	}
	cc.KnownSpells = append(cc.KnownSpells, name)
}

func (cc *CombatantCapabilities) UseCounterspell() {
	cc.HasUsedCounterspell = true
}

func (cc *CombatantCapabilities) UseHealingSpell() {
	cc.HasUsedHealingSpells = true
}

func (cc *CombatantCapabilities) UseAOE() {
	cc.HasUsedAOE = true
}

func (cc *CombatantCapabilities) UseMeleeAttack() {
	cc.HasUsedMeleeAttack = true
}

func (cc *CombatantCapabilities) UseRangedAttack() {
	cc.HasUsedRangedAttack = true
}

// UseLegendaryAction consumes a legendary action
func (cc *CombatantCapabilities) UseLegendaryAction() {
	cc.HasUsedLegendaryActions = true
}

type EventContext struct {
	sequenceID string
	parentID   string
	currentID  string
}

func NewEventContext() *EventContext {
	return &EventContext{
		sequenceID: "",
		parentID:   "",
		currentID:  "",
	}
}

func (ctx *EventContext) GetSequenceID() string { return ctx.sequenceID }
func (ctx *EventContext) GenerateSequenceID()   { ctx.sequenceID = NewUUIDv7() }
func (ctx *EventContext) GetParentID() string   { return ctx.parentID }
func (ctx *EventContext) SetParentID(id string) { ctx.parentID = id }
func (ctx *EventContext) GenerateParentID()     { ctx.parentID = NewUUIDv7() }
func (ctx *EventContext) GetCurrentID() string  { return ctx.currentID }
func (ctx *EventContext) GenerateCurrentID()    { ctx.currentID = NewUUIDv7() }
func (ctx *EventContext) AdvanceScope() {
	ctx.parentID = ctx.currentID
	ctx.GenerateCurrentID()
}

// NewUUIDv7 generates and returns a new UUID version 7 as a string.
func NewUUIDv7() string {
	u, _ := uuid.NewV7()
	return u.String()
}
