package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/events"
	"dnd5e-encounter-simulator-backend/pkg/rolling"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
)

type Monster struct {
	MonsterBase
	DamageModifiers  []MonsterDamageModifier
	ResistBreakers   []shared.DamageBreaker
	Actions          []MonsterAction
	Multiattacks     []MonsterMultiattack
	LegendaryActions []LegendaryAction
	SpecialAbilities []SpecialAbility
	Spellcasting     MSpellcasting
	EventListener    func(events.CombatEvent)
}

type MonsterBase struct {
	ID                  int
	Name                string
	Size                string
	Type                string
	AC                  int
	ProficiencyBonus    int
	CR                  float64
	ApiURL              string
	IsLegendary         bool
	IsSpellcaster       bool
	IsInnateSpellcaster bool
	AbilityScores       shared.AbilityScores
	HP                  shared.MonsterHP
	SaveProficiencies   shared.SaveProficiencies
}

type MonsterDamageModifier struct {
	DamageType   string
	ModifierType string
}

type MonsterAction struct {
	ActionID      int
	Name          string
	RechargeValue int
	HasDC         bool // Used to determine if embedded struct is of value
	Index         int
	NumberOfDice  int
	Die           int
	AmountToAdd   int
	AttackBonus   int
	DamageType    string
	MonsterActionDC
}

type MonsterActionDC struct {
	Ability   string
	OnSuccess string
	DC        int
}

type MonsterMultiattack struct {
	ActionID    int
	AttackCount int
	IsOption    bool
	OptionIndex int
}

type LegendaryAction struct {
	Cost int
	MonsterAction
}

type SpecialAbility struct {
	Name        string
	UsageCount  int
	Description string
}

type MSpellcasting struct {
	CastingLevel   int
	Ability        string
	AttackModifier int
	SaveDC         int
	InnateSpells   []InnateSpell
	SC             StandardSC
}

type InnateSpell struct {
	Spell      spells.Spell
	TimePerDay int
}

type StandardSC struct {
	Spells        []spells.Spell
	SpellSlots    map[int]int // Current available spell slots
	MaxSpellSlots map[int]int // Max spell slots - do not change
}

type MonsterQueryParams struct {
	Name string
	ID   int
}

func (m *Monster) DetermineMonsterHP(useAverage bool) (int, int, error) {
	if !useAverage {
		toAdd := m.HP.AmountToAdd
		s, rolls, err := rolling.RollDice(m.HP.NumberOfDice, m.HP.Die)
		if err != nil {
			return 0, toAdd, fmt.Errorf("error rolling hp dice: %w", err)
		}
		if m.EventListener != nil {
			event := events.CombatEvent{
				EventType: events.HPRollEvent,
				Actor:     m.Name,
				Value:     s + toAdd,
				Rolls:     rolls,
				Added:     toAdd,
			}
			m.EventListener(event)
		}
		return s + toAdd, toAdd, nil
	} else {
		event := events.CombatEvent{
			EventType: events.HPRollEvent,
			Actor:     m.Name,
			Value:     m.HP.HPAverage,
			Rolls:     []int{m.HP.HPAverage},
		}
		m.EventListener(event)
		return m.HP.HPAverage, 0, nil
	}
}

func (m *Monster) ModifyHP(amount int) {
	m.HP.HP += amount
	if m.HP.HP > m.HP.MaxHP {
		m.HP.HP = m.HP.MaxHP
	}
	if m.HP.HP < 0 {
		m.HP.HP = 0
	}
}

func (m *Monster) IsUnconscious() bool {
	if m.HP.HP <= 0 {
		if m.EventListener != nil {
			event := events.CombatEvent{
				EventType: events.UnconsciousEvent,
				Actor:     m.Name,
			}
			m.EventListener(event)
		}
		return true
	}
	return false
}

func (m *Monster) GetName() string {
	return m.Name
}

func (m *Monster) GetCurrentHP() int {
	return m.HP.HP
}

func (m *Monster) GetMaxHP() int {
	return m.HP.MaxHP
}

var _ shared.Entity = &Monster{}

// TODO: Add New function?
