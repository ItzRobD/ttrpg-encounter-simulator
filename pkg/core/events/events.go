package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"time"
)

type EventType string

const (
	ETAttackEvent                 EventType = "attack"
	ETSpellAttackEvent            EventType = "spellattack"
	ETSpellDCEvent                EventType = "spelldc"
	ETHealEvent                   EventType = "heal"
	ETDamageEvent                 EventType = "damage"
	ETDeathEvent                  EventType = "death"
	ETUnconsciousEvent            EventType = "unconscious"
	ETRollEvent                   EventType = "roll"
	ETHPRollEvent                 EventType = "hproll"
	ETActionChoiceEvent           EventType = "actionchoice"
	ETSpellChoiceEvent            EventType = "spellchoice"
	ETSpellSlotsEvent             EventType = "spellslots"
	ETHPModifiedEvent             EventType = "hpmodified"
	ETSavingThrowEvent            EventType = "savingthrow"
	ETTargetChoiceEvent           EventType = "targetchoice"
	ECombatEventMessage           EventType = "combatmessage"
	ETDamageModifiedEvent         EventType = "damagedmodified"
	ETDragonbornBreathWeaponEvent EventType = "dragonbornbreathweapon"
	ETSpecialAbilityEvent         EventType = "specialability"
)

type CombatEvent interface {
	GetRound() int
	GetActor() core.Entity
	GetActorName() string
	GetTimestamp() time.Time
	SetRound(int)
	SetActor(entity core.Entity)
	SetTimestamp(time.Time)
	GetID() string
	Context() *core.EventContext
	GetEventType() EventType
}

type BaseEvent struct {
	Round     int
	Timestamp time.Time
	ctx       *core.EventContext
	actor     core.Entity
	id        string
}

func (b *BaseEvent) GetRound() int {
	return b.Round
}
func (b *BaseEvent) GetTimestamp() time.Time {
	return b.Timestamp
}
func (b *BaseEvent) GetActor() core.Entity {
	return b.actor
}
func (b *BaseEvent) GetActorName() string {
	return b.actor.GetName()
}
func (b *BaseEvent) SetRound(round int) {
	b.Round = round
}
func (b *BaseEvent) SetTimestamp(timestamp time.Time) {
	b.Timestamp = timestamp
}
func (b *BaseEvent) SetActor(actor core.Entity) {
	b.actor = actor
}
func (b *BaseEvent) GetID() string {
	return b.id
}
func (b *BaseEvent) SetID(id string) {
	b.id = id
}
func (b *BaseEvent) MakeNewEventID() {
	b.id = core.NewUUIDv7()
}
func (b *BaseEvent) SetContext(ctx *core.EventContext) {
	b.ctx = ctx
}
func (b *BaseEvent) Context() *core.EventContext {
	return b.ctx
}

type MeleeAttackEvent struct {
	BaseEvent
	Target         string
	AttackName     string
	AttackCount    int
	AttackRoll     int
	AttackModifier int
	AttackTotal    int
	TargetValue    int
	Success        bool
	CriticalHit    bool
	DamageTotal    int
	DamageType     string
}

func (e *MeleeAttackEvent) GetEventType() EventType { return ETAttackEvent }

type DragonbornBreathWeaponEvent struct {
	BaseEvent
	Target             string
	DamageTotal        int
	DamageType         string
	DC                 int
	SaveAbility        string
	SavingThrowSuccess bool
	SavingThrowResult  int
}

func (e *DragonbornBreathWeaponEvent) GetEventType() EventType { return ETDragonbornBreathWeaponEvent }

type DecisionFactor string

const (
	FactorEmergencyHeal      DecisionFactor = "Ally Critical HP"
	FactorHighThreat         DecisionFactor = "High Damage Taken"
	FactorBloodiedTarget     DecisionFactor = "Enemy Bloodied"
	FactorDeterministicNoise DecisionFactor = "DM Whim (Noise)"
	FactorOptimalDamage      DecisionFactor = "Max Damage Output"
	FactorHighHitability     DecisionFactor = "Easy to Hit"
	FactorLowHitability      DecisionFactor = "Hard to Hit"
	FactorHighPotency        DecisionFactor = "Dangerous Target (High Stats)"
	FactorVengeance          DecisionFactor = "Vengeance (Last Attacker)"
	FactorConcentration      DecisionFactor = "Breaking Concentration"
	FactorEliteThreat        DecisionFactor = "Elite Enemy (Multiattack)"
	FactorPotencyNova        DecisionFactor = "High Resource (Elite Target)"
	FactorCriticalSmite      DecisionFactor = "Critical Hit Smite"
)

type ActionUtilityScore struct {
	ActionType core.ActionType
	TotalScore float64
	Factors    map[DecisionFactor]float64 // Factor name -> Weighted contribution
}

type ActionChoiceEvent struct {
	BaseEvent
	ActionChoice core.ActionType
	AllScores    []ActionUtilityScore       // Data for the UI to graph all options
	TopReasons   []DecisionFactor           // Human-readable top 3
	UtilityScore float64                    // Final winning score
	Factors      map[DecisionFactor]float64 // All factors
}

func (e *ActionChoiceEvent) GetEventType() EventType { return ETActionChoiceEvent }

//type SpellSlotsEvent struct {
//	BaseEvent
//	SpellSlots shared.SpellSlots
//}

//func (e *SpellSlotsEvent) GetEventType() EventType { return ETSpellSlotsEvent }

type SpellChoiceEvent struct {
	BaseEvent
	SpellChoice   *core.SpellChoice
	ManagerStatus *spells.SpellcastingManagerStatus
}

func (e *SpellChoiceEvent) GetEventType() EventType { return ETSpellChoiceEvent }

type SpellAttackEvent struct {
	BaseEvent
	Target             string
	SpellName          string
	SpellLevel         int
	AttackRoll         int
	AttackModifier     int
	AttackTotal        int
	Success            bool
	CriticalHit        bool
	DamageTotal        int
	DamageType         string
	HasDC              bool
	DCAbility          string
	SaveEffect         string
	DCValue            int
	SavingThrowSuccess bool
	SavingThrowResult  int
}

func (e *SpellAttackEvent) GetEventType() EventType { return ETSpellAttackEvent }

type SpellDCEvent struct {
	BaseEvent
	Target      string
	SpellChoice *spells.Spell
	DC          int
	SavingThrow int
	Success     bool
}

func (e *SpellDCEvent) GetEventType() EventType { return ETSpellDCEvent }

type DamageEvent struct {
	BaseEvent
	Target     string
	DamageType string
	Amount     int
	Rolls      []int
}

func (e *DamageEvent) GetEventType() EventType { return ETDamageEvent }

type HealEvent struct {
	BaseEvent
	Target     string
	Name       string
	IsSpell    bool
	SpellLevel int
	HealTotal  int
	HealRolls  []int
}

func (e *HealEvent) GetEventType() EventType { return ETHealEvent }

type DeathEvent struct {
	BaseEvent
}

func (e *DeathEvent) GetEventType() EventType { return ETDeathEvent }

type UnconsciousEvent struct {
	BaseEvent
}

func (e *UnconsciousEvent) GetEventType() EventType { return ETUnconsciousEvent }

type DamageModifiedEvent struct {
	BaseEvent
	SubjectName      string
	OriginalValue    int
	FinalValue       int
	WasModified      bool
	ResistanceType   core.ResistanceType
	ResistanceBroken bool
}

func (e *DamageModifiedEvent) GetEventType() EventType { return ETDamageModifiedEvent }

type HPModifiedEvent struct {
	BaseEvent
	SubjectName       string
	ModificationValue int
	OriginalHP        int
	OriginalTempHP    int
	NewHP             int
	NewTempHP         int
	DidHealHP         bool
	DidHealTempHP     bool
	DidTempDamage     bool
	DidHPDamage       bool
	IsUnconscious     bool
	IsMaxHealth       bool
}

func (e *HPModifiedEvent) GetEventType() EventType { return ETHPModifiedEvent }

type DiceRollEvent struct {
	BaseEvent
	RollType       core.DiceRollType
	NumberOfDice   int
	Die            string
	FinalRollValue int
	FinalRolls     []int
	Modifier       int
	Total          int
	Advantage      string

	// Reroll tracking
	OriginalRolls []int
	RerollEvents  []map[string]interface{}
	WasRerolled   bool

	// Special results
	IsCritical   bool
	IsNaturalOne bool
	IsSuccess    bool
	TargetValue  int
	Name         string // Used for recharges only
}

func (e *DiceRollEvent) GetEventType() EventType { return ETRollEvent }

type HPRollEvent struct {
	BaseEvent
	Value    int
	Rolls    []int
	Modifier int
}

func (e *HPRollEvent) GetEventType() EventType { return ETHPRollEvent }

type SavingThrowEvent struct {
	BaseEvent
	Result   int
	Roll     int
	Modifier int
	Success  bool
}

func (e *SavingThrowEvent) GetEventType() EventType { return ETSavingThrowEvent }

type TargetChoiceEvent struct {
	BaseEvent
	Target  string
	Score   float64
	Factors map[DecisionFactor]float64
}

func (e *TargetChoiceEvent) GetEventType() EventType { return ETTargetChoiceEvent }

type SpecialAbilityEvent struct {
	BaseEvent
	AbilityName string
	Description string
	Target      string
	Value       int
}

func (e *SpecialAbilityEvent) GetEventType() EventType { return ETSpecialAbilityEvent }

type CombatEventMessage struct {
	BaseEvent
	Actor   string
	Message string
}

func (e *CombatEventMessage) GetEventType() EventType { return ECombatEventMessage }

type CombatLogger interface {
	LogEvent(event CombatEvent)
}

type CombatListener interface {
	HandleEvent(event CombatEvent)
}
