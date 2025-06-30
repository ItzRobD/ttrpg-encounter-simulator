package monster

import (
	"context"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"strings"
)

type Monster struct {
	MonsterBase
	CombatState      core.CombatState
	EntityModifiers  core.EntityModifiers
	DamageModifiers  []MonsterDamageModifier
	ResistBreakers   []shared.DamageBreaker
	Actions          []MonsterAction
	Multiattacks     []MonsterMultiattack
	LegendaryActions []LegendaryAction
	SpecialAbilities []SpecialAbility
	Spellcasting     MSpellcasting
	EventListener    func(event interface{})
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
	AbilityScores       core.AbilityScores
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

func NewSRDMonster(ctx context.Context, params MonsterQueryParams, em core.EntityModifiers) (*Monster, error) {
	var err error
	base, err := QueryMonsterData(ctx, params)
	if err != nil {
		return nil, err
	}

	damageModifiers, err := getMonsterDamageModifiersByID(ctx, base.ID)
	if err != nil {
		return nil, err
	}

	resistBreakers, err := getMonsterResistBreakersByID(ctx, base.ID)
	if err != nil {
		return nil, err
	}

	actions, err := getMonsterActionsByID(ctx, base.ID)
	if err != nil {
		return nil, err
	}

	multiattacks, err := getMonsterMultiattacksByID(ctx, base.ID)
	if err != nil {
		return nil, err
	}

	specialAbilities, err := getMonsterSpecialAbilities(ctx, base.ID)
	if err != nil {
		return nil, err
	}

	legendaryActions, err := getMonsterLegendaryActionsByID(ctx, base.ID)
	if err != nil {
		fmt.Println(err)
	}

	spellcasting, err := getMonsterSpellcastingByID(ctx, base.ID)
	if err != nil {
		if !strings.Contains(err.Error(), "no rows in result set") {
			fmt.Println(err)
		}
	}

	cs := core.CombatState{
		CurrentHP: base.HP.MaxHP,
		MaxHP:     base.HP.MaxHP,
		TempHP:    0,

		HasUsedAction:         false,
		HasUsedBonusAction:    false,
		HasUsedReaction:       false,
		LegendaryActionPoints: 3,
	}

	monster := &Monster{
		MonsterBase:      base,
		CombatState:      cs,
		EntityModifiers:  em,
		DamageModifiers:  damageModifiers,
		ResistBreakers:   resistBreakers,
		Actions:          actions,
		Multiattacks:     multiattacks,
		LegendaryActions: legendaryActions,
		SpecialAbilities: specialAbilities,
		Spellcasting:     spellcasting,
	}

	return monster, nil
}

func (m *Monster) SetEntityModifiers(em core.EntityModifiers) {
	m.EntityModifiers = em
}

func (m *Monster) DetermineMonsterHP(useAverage bool) (int, int, error) {
	if !useAverage {
		toAdd := m.HP.AmountToAdd
		s, rolls, err := core.RollDice(m.HP.NumberOfDice, m.HP.Die)
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

func (m *Monster) ModifyHP(value int) {
	m.HP.HP += amount
	if m.HP.HP > m.HP.MaxHP {
		m.HP.HP = m.HP.MaxHP
	}
	if m.HP.HP < 0 {
		m.HP.HP = 0
	}
}

// GetSavingThrowRollResult calculates a saving throw for the given ability and returns the total roll result or an error.
func (m *Monster) GetSavingThrowRollResult(ability string) (int, error) {
	// roll dice, add save
	var mod int
	var err error
	switch ability {
	case spells.SpellDCStrength:
		if m.SaveProficiencies.Strength == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Strength)
			if err != nil {
				return 0, err
			}
		} else {
			mod = m.SaveProficiencies.Strength
		}
	case spells.SpellDCDexterity:
		if m.SaveProficiencies.Dexterity == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Dexterity)
			if err != nil {
				return 0, err
			}
		} else {
			mod = m.SaveProficiencies.Dexterity
		}
	case spells.SpellDCConstitution:
		if m.SaveProficiencies.Constitution == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Constitution)
			if err != nil {
				return 0, err
			}
		} else {
			mod = m.SaveProficiencies.Constitution
		}
	case spells.SpellDCIntelligence:
		if m.SaveProficiencies.Intelligence == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Intelligence)
			if err != nil {
				return 0, err
			}
		} else {
			mod = m.SaveProficiencies.Intelligence
		}
	case spells.SpellDCWisdom:
		if m.SaveProficiencies.Wisdom == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Wisdom)
			if err != nil {
				return 0, err
			}
		} else {
			mod = m.SaveProficiencies.Wisdom
		}
	case spells.SpellDCCharisma:
		if m.SaveProficiencies.Charisma == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Charisma)
			if err != nil {
				return 0, err
			}
		} else {
			mod = m.SaveProficiencies.Charisma
		}
	default:
		return 0, fmt.Errorf("invalid ability: %s", ability)
	}

	//roll, rolls, err := shared.RollDice(1, 20)
	roll, _, err := core.RollDice(1, 20)
	save := roll + mod
	//m.logRollEvent(roll, rolls, mod)

	return save, nil
}

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

func (m *Monster) GetLevel() interface{} { return m.CR }

func (m *Monster) GetCasterLevel() int { return m.Spellcasting.CastingLevel }

func (m *Monster) IsUnconscious() bool {
	if m.HP.HP <= 0 {
		if m.EventListener != nil {
			event := &events.UnconsciousEvent{}
			event.SetActor(m.Name)
			m.EventListener(event)
		}
		return true
	}
	return false
}

func (m *Monster) GetEventListener() func(event interface{}) {
	return m.EventListener
}

var _ core.Entity = &Monster{}

// TODO: Add function to create melee attacks from actions
