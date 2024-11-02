package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/simulation/events"
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
		s, rolls, err := shared.RollDice(m.HP.NumberOfDice, m.HP.Die)
		if err != nil {
			return 0, toAdd, fmt.Errorf("error rolling hp dice: %w", err)
		}

		events.LogHPRollEvent(m, s, rolls, toAdd, m.EventListener)
		return s + toAdd, toAdd, nil
	} else {
		events.LogHPRollEvent(m, m.HP.HPAverage, []int{m.HP.HPAverage}, 0, m.EventListener)
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

//func (m *Monster) SavingThrow(ability string) (int, error) {
//	// roll dice, add save
//	var mod int
//	var err error
//	switch ability {
//	case spells.SpellDCStrength:
//		if m.SaveProficiencies.Strength == 0 {
//			mod, err = shared.GetAbilityScoreModifier(m.AbilityScores.Strength)
//			if err != nil {
//				return 0, err
//			}
//		} else {
//			mod = m.SaveProficiencies.Strength
//		}
//	case spells.SpellDCDexterity:
//		if m.SaveProficiencies.Dexterity == 0 {
//			mod, err = shared.GetAbilityScoreModifier(m.AbilityScores.Dexterity)
//			if err != nil {
//				return 0, err
//			}
//		} else {
//			mod = m.SaveProficiencies.Dexterity
//		}
//	case spells.SpellDCConstitution:
//		if m.SaveProficiencies.Constitution == 0 {
//			mod, err = shared.GetAbilityScoreModifier(m.AbilityScores.Constitution)
//			if err != nil {
//				return 0, err
//			}
//		} else {
//			mod = m.SaveProficiencies.Constitution
//		}
//	case spells.SpellDCIntelligence:
//		if m.SaveProficiencies.Intelligence == 0 {
//			mod, err = shared.GetAbilityScoreModifier(m.AbilityScores.Intelligence)
//			if err != nil {
//				return 0, err
//			}
//		} else {
//			mod = m.SaveProficiencies.Intelligence
//		}
//	case spells.SpellDCWisdom:
//		if m.SaveProficiencies.Wisdom == 0 {
//			mod, err = shared.GetAbilityScoreModifier(m.AbilityScores.Wisdom)
//			if err != nil {
//				return 0, err
//			}
//		} else {
//			mod = m.SaveProficiencies.Wisdom
//		}
//	case spells.SpellDCCharisma:
//		if m.SaveProficiencies.Charisma == 0 {
//			mod, err = shared.GetAbilityScoreModifier(m.AbilityScores.Charisma)
//			if err != nil {
//				return 0, err
//			}
//		} else {
//			mod = m.SaveProficiencies.Charisma
//		}
//	default:
//		return 0, fmt.Errorf("invalid ability: %s", ability)
//	}
//
//	roll, rolls, err := shared.RollDice(1, 20)
//	save := roll + mod
//	m.logRollEvent(roll, rolls, mod)
//
//	return save, nil
//}

func (m *Monster) GetName() string {
	return m.Name
}

func (m *Monster) GetCurrentHP() int {
	return m.HP.HP
}

func (m *Monster) GetCurrentHPPct() int {
	hpPct := int(float64(m.HP.HP) / float64(m.HP.MaxHP) * 100)
	return hpPct
}

func (m *Monster) GetMaxHP() int {
	return m.HP.MaxHP
}

func (m *Monster) GetCR() float64 { return m.CR }

func (m *Monster) GetAC() int { return m.AC }

func (m *Monster) IsUnconscious() bool {
	if m.HP.HP <= 0 {
		if m.EventListener != nil {
			event := events.CombatEvent{
				EventType: events.ETUnconsciousEvent,
				Actor:     m.Name,
			}
			m.EventListener(event)
		}
		return true
	}
	return false
}

func (m *Monster) GetEventListener() func(events.CombatEvent) {
	return m.EventListener
}

var _ shared.Entity = &Monster{}

// TODO: Add New function?
