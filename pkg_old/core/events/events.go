package events

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/spells"
	"time"
)

type EventType string

const (
	ETAttackEvent                 EventType = "attack"
	ETSpellAttackEvent            EventType = "spell_attack"
	ETSpellDCEvent                EventType = "spell_dc"
	ETHealEvent                   EventType = "heal"
	ETDamageEvent                 EventType = "damage"
	ETDeathEvent                  EventType = "death"
	ETUnconsciousEvent            EventType = "unconscious"
	ETRollEvent                   EventType = "roll"
	ETHPRollEvent                 EventType = "hp_roll"
	ETActionChoiceEvent           EventType = "action_choice"
	ETSpellChoiceEvent            EventType = "spell_choice"
	ETSpellSlotsEvent             EventType = "spell_slots"
	ETHPModifiedEvent             EventType = "hp_modified"
	ETSavingThrowEvent            EventType = "saving_throw"
	ETTargetChoiceEvent           EventType = "target_choice"
	ECombatEventMessage           EventType = "combat_message"
	ETDamageModifiedEvent         EventType = "damage_modified"
	ETDragonbornBreathWeaponEvent EventType = "dragonborn_breath_weapon"
	ETSpecialAbilityEvent         EventType = "special_ability"
	ETRollInitiative              EventType = "initiative"
	ETConditionEvent              EventType = "condition"
	ETVictoryEvent                EventType = "victory"
	ETEquipmentEvent              EventType = "equipment"
	ETTurnStartEvent              EventType = "turn_start"
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
	SetID(id string)
	Context() *core.EventContext
	SetContext(ctx *core.EventContext)
	GetEventType() EventType
	MakeNewEventID()
}

type BaseEvent struct {
	Round     int                `json:"round"`
	Timestamp time.Time          `json:"timestamp"`
	ctx       *core.EventContext `json:"-"`
	actor     core.Entity        `json:"-"`
	ID        string             `json:"id"`
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
	return b.ID
}
func (b *BaseEvent) SetID(id string) {
	b.ID = id
}
func (b *BaseEvent) MakeNewEventID() {
	b.ID = core.NewUUIDv7()
}
func (b *BaseEvent) SetContext(ctx *core.EventContext) {
	b.ctx = ctx
}
func (b *BaseEvent) Context() *core.EventContext {
	return b.ctx
}

type MartialAttackEvent struct {
	BaseEvent
	Target         string        `json:"target"`
	target         core.Entity   `json:"-"`
	AttackName     string        `json:"attack_name"`
	AttackCount    int           `json:"attack_count"`
	AttackRoll     int           `json:"attack_roll"`
	AttackModifier int           `json:"attack_modifier"`
	AttackTotal    int           `json:"attack_total"`
	TargetValue    int           `json:"target_value"`
	Success        bool          `json:"success"`
	CriticalHit    bool          `json:"critical_hit"`
	DamageTotal    int           `json:"damage_total"`
	DamageType     string        `json:"damage_type"`
	DamageParts    string        `json:"damage_parts"`
	IsRanged       bool          `json:"is_ranged"`
	ActionDetail   *ActionDetail `json:"action_detail,omitempty"`
}

func (e *MartialAttackEvent) SetTargetEntity(target core.Entity) { e.target = target }
func (e *MartialAttackEvent) GetTargetEntity() core.Entity       { return e.target }

func (e *MartialAttackEvent) GetEventType() EventType { return ETAttackEvent }

type EquipmentEvent struct {
	BaseEvent
	Name         string   `json:"name"`
	NumberOfDice int      `json:"number_of_dice"`
	Die          string   `json:"die"`
	DamageType   string   `json:"damage_type"`
	AttackBonus  int      `json:"attack_bonus"`
	DamageBonus  int      `json:"damage_bonus"`
	IsRanged     bool     `json:"is_ranged"`
	Properties   []string `json:"properties"`
	Modifiers    []string `json:"modifiers"`
}

func (e *EquipmentEvent) GetEventType() EventType {
	if e.actor != nil && e.actor.IsMonster() {
		return "action_detail"
	}
	return ETEquipmentEvent
}

type DragonbornBreathWeaponEvent struct {
	BaseEvent
	Target             string      `json:"target"`
	target             core.Entity `json:"-"`
	DamageTotal        int         `json:"damage_total"`
	DamageType         string      `json:"damage_type"`
	DC                 int         `json:"dc"`
	SaveAbility        string      `json:"save_ability"`
	SavingThrowSuccess bool        `json:"saving_throw_success"`
	SavingThrowResult  int         `json:"saving_throw_result"`
}

func (e *DragonbornBreathWeaponEvent) SetTargetEntity(target core.Entity) { e.target = target }
func (e *DragonbornBreathWeaponEvent) GetTargetEntity() core.Entity       { return e.target }

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
	ActionType core.ActionType            `json:"action_type"`
	TotalScore float64                    `json:"total_score"`
	Factors    map[DecisionFactor]float64 `json:"factors"` // Factor name -> Weighted contribution
}

type ActionChoiceEvent struct {
	BaseEvent
	ActionChoice core.ActionType            `json:"action_choice"`
	AllScores    []ActionUtilityScore       `json:"all_scores"`    // Data for the UI to graph all options
	TopReasons   []DecisionFactor           `json:"top_reasons"`   // Human-readable top 3
	UtilityScore float64                    `json:"utility_score"` // Final winning score
	Factors      map[DecisionFactor]float64 `json:"factors"`       // All factors
}

func (e *ActionChoiceEvent) GetEventType() EventType { return ETActionChoiceEvent }

//type SpellSlotsEvent struct {
//	BaseEvent
//	SpellSlots shared.SpellSlots
//}

//func (e *SpellSlotsEvent) GetEventType() EventType { return ETSpellSlotsEvent }

type SpellChoiceEvent struct {
	BaseEvent
	SpellChoice   *core.SpellChoice                 `json:"spell_choice"`
	ManagerStatus *spells.SpellcastingManagerStatus `json:"manager_status"`
	Target        core.Entity                       `json:"-"`
}

func (e *SpellChoiceEvent) SetTargetEntity(target core.Entity) { e.Target = target }
func (e *SpellChoiceEvent) GetTargetEntity() core.Entity       { return e.Target }

func (e *SpellChoiceEvent) GetEventType() EventType { return ETSpellChoiceEvent }

type SpellAttackEvent struct {
	BaseEvent
	Target             string `json:"target"`
	target             core.Entity
	SpellName          string        `json:"spell_name"`
	SpellLevel         int           `json:"spell_level"`
	AttackRoll         int           `json:"attack_roll"`
	AttackModifier     int           `json:"attack_modifier"`
	AttackTotal        int           `json:"attack_total"`
	Success            bool          `json:"success"`
	CriticalHit        bool          `json:"critical_hit"`
	DamageTotal        int           `json:"damage_total"`
	DamageType         string        `json:"damage_type"`
	HasDC              bool          `json:"has_dc"`
	DCAbility          string        `json:"dc_ability"`
	SaveEffect         string        `json:"save_effect"`
	DCValue            int           `json:"dc_value"`
	SavingThrowSuccess bool          `json:"saving_throw_success"`
	SavingThrowResult  int           `json:"saving_throw_result"`
	ActionDetail       *ActionDetail `json:"action_detail,omitempty"`
}

func (e *SpellAttackEvent) SetTargetEntity(target core.Entity) { e.target = target }
func (e *SpellAttackEvent) GetTargetEntity() core.Entity       { return e.target }

func (e *SpellAttackEvent) GetEventType() EventType { return ETSpellAttackEvent }

type SpellDCEvent struct {
	BaseEvent
	Target      string `json:"target"`
	target      core.Entity
	SpellChoice *spells.Spell `json:"spell_choice"`
	DC          int           `json:"dc"`
	SavingThrow int           `json:"saving_throw"`
	Success     bool          `json:"success"`
}

func (e *SpellDCEvent) SetTargetEntity(target core.Entity) { e.target = target }
func (e *SpellDCEvent) GetTargetEntity() core.Entity       { return e.target }

func (e *SpellDCEvent) GetEventType() EventType { return ETSpellDCEvent }

type DamageEvent struct {
	BaseEvent
	Target     string `json:"target"`
	target     core.Entity
	DamageType string `json:"damage_type"`
	Amount     int    `json:"amount"`
	Rolls      []int  `json:"rolls"`
}

func (e *DamageEvent) SetTargetEntity(target core.Entity) { e.target = target }
func (e *DamageEvent) GetTargetEntity() core.Entity       { return e.target }

func (e *DamageEvent) GetEventType() EventType { return ETDamageEvent }

type HealEvent struct {
	BaseEvent
	Target     string `json:"target"`
	target     core.Entity
	Name       string `json:"name"`
	IsSpell    bool   `json:"is_spell"`
	SpellLevel int    `json:"spell_level"`
	HealTotal  int    `json:"heal_total"`
	HealRolls  []int  `json:"heal_rolls"`
}

func (e *HealEvent) SetTargetEntity(target core.Entity) { e.target = target }
func (e *HealEvent) GetTargetEntity() core.Entity       { return e.target }

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
	SubjectName      string `json:"subject_name"`
	Subject          core.Entity
	OriginalValue    int                 `json:"original_value"`
	FinalValue       int                 `json:"final_value"`
	DamageType       core.DamageType     `json:"damage_type"`
	WasModified      bool                `json:"was_modified"`
	ResistanceType   core.ResistanceType `json:"resistance_type"`
	ResistanceBroken bool                `json:"resistance_broken"`
	SourceRollID     string              `json:"source_roll_id"`
}

func (e *DamageModifiedEvent) SetSubjectEntity(subject core.Entity) { e.Subject = subject }
func (e *DamageModifiedEvent) GetSubjectEntity() core.Entity        { return e.Subject }

func (e *DamageModifiedEvent) GetEventType() EventType { return ETDamageModifiedEvent }

type HPModifiedEvent struct {
	BaseEvent
	SubjectName       string `json:"subject_name"`
	Subject           core.Entity
	ModificationValue int             `json:"modification_value"`
	OriginalHP        int             `json:"original_hp"`
	OriginalTempHP    int             `json:"original_temp_hp"`
	NewHP             int             `json:"new_hp"`
	NewTempHP         int             `json:"new_temp_hp"`
	DamageType        core.DamageType `json:"damage_type"`
	DidHealHP         bool            `json:"did_heal_hp"`
	DidHealTempHP     bool            `json:"did_heal_temp_hp"`
	DidTempDamage     bool            `json:"did_temp_damage"`
	DidHPDamage       bool            `json:"did_hp_damage"`
	IsUnconscious     bool            `json:"is_unconscious"`
	IsMaxHealth       bool            `json:"is_max_health"`
	SourceRollID      string          `json:"source_roll_id"`
}

func (e *HPModifiedEvent) SetSubjectEntity(subject core.Entity) { e.Subject = subject }
func (e *HPModifiedEvent) GetSubjectEntity() core.Entity        { return e.Subject }

func (e *HPModifiedEvent) GetEventType() EventType { return ETHPModifiedEvent }

type DiceRollEvent struct {
	BaseEvent
	RollType       core.DiceRollType        `json:"roll_type"`
	NumberOfDice   int                      `json:"number_of_dice"`
	Die            string                   `json:"die"`
	FinalRollValue int                      `json:"final_roll_value"`
	FinalRolls     []int                    `json:"final_rolls"`
	Modifier       int                      `json:"modifier"`
	Total          int                      `json:"total"`
	Advantage      string                   `json:"advantage"`
	OriginalRolls  []int                    `json:"original_rolls"`
	RerollEvents   []map[string]interface{} `json:"reroll_events"`
	WasRerolled    bool                     `json:"was_rerolled"`
	IsCritical     bool                     `json:"is_critical"`
	IsNaturalOne   bool                     `json:"is_natural_one"`
	IsSuccess      bool                     `json:"is_success"`
	TargetValue    int                      `json:"target_value"`
	DamageType     string                   `json:"damage_type"`
	Name           string                   `json:"name"` // Used for recharges only
	Target         core.Entity              `json:"-"`
}

func (e *DiceRollEvent) SetTargetEntity(target core.Entity) { e.Target = target }
func (e *DiceRollEvent) GetTargetEntity() core.Entity       { return e.Target }

func (e *DiceRollEvent) GetEventType() EventType {
	if e.RollType == core.DiceRollInitiative {
		return ETRollInitiative
	}
	return ETRollEvent
}

type HPRollEvent struct {
	BaseEvent
	Value    int   `json:"value"`
	Rolls    []int `json:"rolls"`
	Modifier int   `json:"modifier"`
}

func (e *HPRollEvent) GetEventType() EventType { return ETHPRollEvent }

type SavingThrowEvent struct {
	BaseEvent
	Result   int  `json:"result"`
	Roll     int  `json:"roll"`
	Modifier int  `json:"modifier"`
	Success  bool `json:"success"`
}

func (e *SavingThrowEvent) GetEventType() EventType { return ETSavingThrowEvent }

type TargetChoiceEvent struct {
	BaseEvent
	Target       string                     `json:"target"`
	TargetEntity core.Entity                `json:"-"`
	Score        float64                    `json:"score"`
	Factors      map[DecisionFactor]float64 `json:"factors"`
}

func (e *TargetChoiceEvent) SetTargetEntity(target core.Entity) { e.TargetEntity = target }
func (e *TargetChoiceEvent) GetTargetEntity() core.Entity       { return e.TargetEntity }

func (e *TargetChoiceEvent) GetEventType() EventType { return ETTargetChoiceEvent }

type SpecialAbilityEvent struct {
	BaseEvent
	AbilityName  string      `json:"ability_name"`
	Description  string      `json:"description"`
	Target       string      `json:"target"`
	TargetEntity core.Entity `json:"-"`
	Value        int         `json:"value"`
}

func (e *SpecialAbilityEvent) SetTargetEntity(target core.Entity) { e.TargetEntity = target }
func (e *SpecialAbilityEvent) GetTargetEntity() core.Entity       { return e.TargetEntity }

func (e *SpecialAbilityEvent) GetEventType() EventType { return ETSpecialAbilityEvent }

type CombatEventMessage struct {
	BaseEvent
	Actor   string `json:"actor_name"`
	Message string `json:"message"`
}

func (e *CombatEventMessage) GetEventType() EventType { return ECombatEventMessage }

type ConditionEvent struct {
	BaseEvent
	Condition core.Condition `json:"condition"`
	IsAdded   bool           `json:"is_added"`
}

func (e *ConditionEvent) GetEventType() EventType { return ETConditionEvent }

type WinningSide string

const (
	WinningSideCharacters WinningSide = "characters"
	WinningSideMonsters   WinningSide = "monsters"
	WinningSideNone       WinningSide = "none"
)

type VictoryEvent struct {
	BaseEvent
	WinningSide WinningSide `json:"winning_side"`
	Rounds      int         `json:"rounds"`
}

func (e *VictoryEvent) GetEventType() EventType { return ETVictoryEvent }

type TurnStartEvent struct {
	BaseEvent
}

func (e *TurnStartEvent) GetEventType() EventType { return ETTurnStartEvent }

// ActionDetail represents the specific action data (formerly emitted as a standalone
// EquipmentEvent for monsters). It is now embedded into attack/spell events.
type ActionDetail struct {
	Name         string   `json:"name"`
	NumberOfDice int      `json:"number_of_dice"`
	Die          string   `json:"die"`
	DamageType   string   `json:"damage_type"`
	AttackBonus  int      `json:"attack_bonus"`
	DamageBonus  int      `json:"damage_bonus"`
	IsRanged     bool     `json:"is_ranged"`
	Properties   []string `json:"properties,omitempty"`
	Modifiers    []string `json:"modifiers,omitempty"`
}

type CombatLogger interface {
	LogEvent(event CombatEvent)
}

type CombatListener interface {
	HandleEvent(event CombatEvent)
}
