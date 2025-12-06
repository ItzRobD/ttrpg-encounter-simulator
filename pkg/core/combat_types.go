package core

import (
	"fmt"
	"strings"
)

type VictoryStatus string

const (
	VictoryStatusNone       VictoryStatus = "none"
	VictoryStatusCharacters VictoryStatus = "characters"
	VictoryStatusMonsters   VictoryStatus = "monsters"
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
	Average           int
}

func (ad AttackData) GetAttackName() string      { return ad.Name }
func (ad AttackData) GetNumberOfDice() int       { return ad.NumberOfDice }
func (ad AttackData) GetDie() DiceType           { return ad.Die }
func (ad AttackData) GetAttackModifier() int     { return ad.AttackModifier }
func (ad AttackData) GetDamageModifier() int     { return ad.DamageModifier }
func (ad AttackData) GetDamageType() string      { return ad.DamageType.String() }
func (ad AttackData) GetIsVersatileAttack() bool { return ad.IsVersatileAttack }

type AttackRequest struct {
	AttackData        []AttackData
	AttackOptions     AttackOptions
	SimulationOptions *SimulationOptions
	Target            Entity
}

func (ar *AttackRequest) GetAttackData() []AttackData              { return ar.AttackData }
func (ar *AttackRequest) GetAttackOptions() AttackOptions          { return ar.AttackOptions }
func (ar *AttackRequest) GetSimulationOptions() *SimulationOptions { return ar.SimulationOptions }
func (ar *AttackRequest) GetTarget() Entity                        { return ar.Target }

type AttackResult struct {
	ActorName     string
	TargetName    string
	AttackName    string
	AttackCount   int
	TargetValue   int
	IsHit         bool
	IsCriticalHit bool
	AttackTotal   int
	AttackRoll    int
	DamageRoll    RollResult
	DamageType    DamageType
}

func (r AttackResult) GetActorName() string        { return r.ActorName }
func (r AttackResult) GetTargetName() string       { return r.TargetName }
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
	Advantage            AdvantageType
	NumberOfAttacks      int
	BonusToAttackRoll    int  // Flat bonus, ie magic weapons
	BonusToDamageRoll    int  // Flat bonus, ie magic weapons, rage, hexblade curse
	ShouldApplyDamageMod bool // Off hand attacks, TWF
	PowerAttack          bool // GWM / Sharpshooter (-5 attack, +10 damage) // TODO: Implement logic for this choice
	ImprovedCritical     bool // Crits on 19 and 20, Hexblade, Champion
	RerollOnesAndTwos    bool // GWF
	// TODO: GWF Creates an extra attack
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
	CombatantInfo      map[int]*CombatantInfo
	LegendaryCreatures map[int]bool

	// Lookup lists
	CharactersInNeedOfHealing []int
	MonstersInNeedOfHealing   []int
	DeadCombatants            []int

	// Turn Tracking
	TurnOrder        []int
	CurrentTurnIndex int
	CurrentRound     int
	ActingEntityID   int

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

type CombatantInfo struct {
	Combatant *Combatant

	// Current state
	State *CombatantState

	// Historical data
	Statistics *CombatStatistics

	// Capability/threat info
	UsedCapabilities *CombatantCapabilities
}

type CombatantState struct {
	CurrentHP        int
	MaxHP            int
	HealThreshold    int
	Concentration    *ConcentrationInfo
	HasActedThisTurn bool
	Conditions       EntityConditions
}

type ConcentrationInfo struct {
	IsConcentrating bool
	SpellName       *string
	AffectedTargets []int // All targets affected by the concentration
	Duration        *int
	RoundsRemaining *int
	RoundStarted    *int
}

type CombatStatistics struct {
	TotalDamageDealt       int
	TotalHealingDone       int
	DamageByRound          []int
	HealingByRound         []int
	LastDamageDealt        int
	LastHealingDone        int
	AverageDamagePerRound  float64
	AverageHealingPerRound float64

	// Attack patterns
	AttacksMade   int
	AttacksHit    int
	AttacksMissed int
	CriticalHits  int

	// Defensive stats
	TimesDamaged         int
	TimesHealed          int
	TotalDamageTaken     int
	TotalHealingReceived int
	DeathSaveSuccesses   int
	DeathSaveFailures    int
}

type CombatantCapabilities struct {
	KnownSpells []string

	// Tactical flags
	HasUsedCounterspell     bool
	HasUsedHealingSpells    bool
	HasUsedAOE              bool
	HasUsedRangedAttack     bool
	HasUsedMeleeAttack      bool
	HasUsedLegendaryActions bool
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
