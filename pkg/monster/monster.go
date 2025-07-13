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
	m.HP.HP += value
	if m.HP.HP > m.HP.MaxHP {
		m.HP.HP = m.HP.MaxHP
	}
	if m.HP.HP < 0 {
		m.HP.HP = 0
	}
}

// MakeSavingThrow calculates a saving throw for the given ability and returns the total roll result or an error.
func (m *Monster) MakeSavingThrow(ability core.Ability) (int, []int, error) {
	// roll dice, add save
	var mod int
	var err error
	switch ability {
	case core.AbilityStrength:
		if m.SaveProficiencies.Strength == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Strength)
			if err != nil {
				return 0, nil, err
			}
		} else {
			mod = m.SaveProficiencies.Strength
		}
	case core.AbilityDexterity:
		if m.SaveProficiencies.Dexterity == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Dexterity)
			if err != nil {
				return 0, nil, err
			}
		} else {
			mod = m.SaveProficiencies.Dexterity
		}
	case core.AbilityConstitution:
		if m.SaveProficiencies.Constitution == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Constitution)
			if err != nil {
				return 0, nil, err
			}
		} else {
			mod = m.SaveProficiencies.Constitution
		}
	case core.AbilityIntelligence:
		if m.SaveProficiencies.Intelligence == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Intelligence)
			if err != nil {
				return 0, nil, err
			}
		} else {
			mod = m.SaveProficiencies.Intelligence
		}
	case core.AbilityWisdom:
		if m.SaveProficiencies.Wisdom == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Wisdom)
			if err != nil {
				return 0, nil, err
			}
		} else {
			mod = m.SaveProficiencies.Wisdom
		}
	case core.AbilityCharisma:
		if m.SaveProficiencies.Charisma == 0 {
			mod, err = core.GetAbilityScoreModifier(m.AbilityScores.Charisma)
			if err != nil {
				return 0, nil, err
			}
		} else {
			mod = m.SaveProficiencies.Charisma
		}
	default:
		return 0, nil, fmt.Errorf("invalid ability: %s", ability)
	}

	roll, rolls, err := core.RollDice(1, 20)
	save := roll + mod

	return save, rolls, nil
}

func (m *Monster) GetName() string {
	return m.Name
}

func (m *Monster) GetAbilityScores() core.AbilityScores {
	return m.AbilityScores
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

func (m *Monster) GetSpellSaveDC(ability core.Ability) int {
	var abilityMod int
	var err error
	switch ability {
	case core.AbilityStrength:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Strength)
	case core.AbilityDexterity:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Dexterity)
	case core.AbilityConstitution:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Constitution)
	case core.AbilityIntelligence:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Intelligence)
	case core.AbilityWisdom:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Wisdom)
	case core.AbilityCharisma:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Charisma)
	default:
		abilityMod = 0
		err = fmt.Errorf("invalid ability provided: %s", ability)
	}
	if err != nil {
		return 0
	}
	return 8 + abilityMod
}

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

// GetAbilityScore returns the score for the specified ability of the monster. Defaults to 0 if the ability is not found.
func (m *Monster) GetAbilityScore(ability core.Ability) int {
	switch ability {
	case core.AbilityStrength:
		return m.AbilityScores.Strength
	case core.AbilityDexterity:
		return m.AbilityScores.Dexterity
	case core.AbilityConstitution:
		return m.AbilityScores.Constitution
	case core.AbilityIntelligence:
		return m.AbilityScores.Intelligence
	case core.AbilityWisdom:
		return m.AbilityScores.Wisdom
	case core.AbilityCharisma:
		return m.AbilityScores.Charisma
	default:
		return 0
	}
}

func (m *Monster) GetAbilityScoreModifier(ability core.Ability) (int, error) {
	var abilityMod int
	var err error
	switch ability {
	case core.AbilityStrength:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Strength)
	case core.AbilityDexterity:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Dexterity)
	case core.AbilityConstitution:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Constitution)
	case core.AbilityIntelligence:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Intelligence)
	case core.AbilityWisdom:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Wisdom)
	case core.AbilityCharisma:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Charisma)
	default:
		abilityMod = 0
		err = fmt.Errorf("invalid ability provided: %s", ability)
	}
	if err != nil {
		return 0, err
	}
	return abilityMod, nil
}

func (m *Monster) GetSavingThrowBonus(ability core.Ability) int {
	switch ability {
	case core.AbilityStrength:
		return m.SaveProficiencies.Strength
	case core.AbilityDexterity:
		return m.SaveProficiencies.Dexterity
	case core.AbilityConstitution:
		return m.SaveProficiencies.Constitution
	case core.AbilityIntelligence:
		return m.SaveProficiencies.Intelligence
	case core.AbilityWisdom:
		return m.SaveProficiencies.Wisdom
	case core.AbilityCharisma:
		return m.SaveProficiencies.Charisma
	default:
		return false
	}
}

func (m *Monster) GetEventListener() func(event interface{}) {
	return m.EventListener
}

var _ core.Entity = &Monster{}

// TODO: Add function to create melee attacks from actions
