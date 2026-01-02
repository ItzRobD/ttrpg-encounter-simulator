package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"strings"
)

type MonsterConfig struct {
	Base               MonsterBase
	Actions            map[int]Action
	Multiattacks       map[int][]Multiattack
	LegendaryActions   map[int]LegendaryAction
	SpecialAbilities   SpecialAbilities
	Resistances        core.DamageResistances
	DamageBreakers     []core.ResistBreaker
	spellcastingConfig MonsterSpellcastingConfig
	HPSetMethod        core.HPSetMethod
	Seed               core.Seed
}

type MonsterSpellcastingConfig struct {
	MonsterID      int
	CastingLevel   int
	Ability        core.Ability
	AttackModifier int
	SaveDC         int
	LeveledSpells  []spells.Spell
	InnateSpells   []spells.InnateSpell
	SpellSlots     spells.SpellSlots
}

type MonsterType string

const (
	MTAberration  = "Aberration"
	MTBeast       = "Beast"
	MTCelestial   = "Celestial"
	MTConstruct   = "Construct"
	MTDragon      = "Dragon"
	MTElemental   = "Elemental"
	MTFey         = "Fey"
	MTFiend       = "Fiend"
	MTGiant       = "Giant"
	MTHumanoid    = "Humanoid"
	MTMonstrosity = "Monstrosity"
	MTOoze        = "Ooze"
	MTPlant       = "Plant"
	MTUndead      = "Undead"
	MTUnknown     = "Unknown"
)

func (mt MonsterType) String() string {
	return string(mt)
}

func MakeMonsterType(s string) MonsterType {
	switch strings.ToLower(s) {
	case "aberration":
		return MTAberration
	case "beast":
		return MTBeast
	case "celestial":
		return MTCelestial
	case "construct":
		return MTConstruct
	case "dragon":
		return MTDragon
	case "elemental":
		return MTElemental
	case "fey":
		return MTFey
	case "fiend":
		return MTFiend
	case "giant":
		return MTGiant
	case "humanoid":
		return MTHumanoid
	case "monstrosity":
		return MTMonstrosity
	case "ooze":
		return MTOoze
	case "plant":
		return MTPlant
	case "undead":
		return MTUndead
	default:
		return MTUnknown
	}
}

type SpecialAbilities struct {
	Assassinate               bool
	BerserkThreshold          int
	BloodFrenzy               bool
	ConsumeLifeDie            core.DiceType // 3dN healing
	CorrosiveFormNumDice      int           // Nd8 damage
	DeathBurstNumDice         int           // Nd8 damage
	DeathBurstDamageType      core.DamageType
	DeathThroesNumDice        int // Nd6 damage
	DivineEminence            bool
	Evasion                   bool
	FireAuraNumDice           int // Nd6 damage
	FireForm                  bool
	Gibbering                 bool
	GnomeCunning              bool
	HeatedBodyNumDice         int // Nd10 damage
	LegendaryResistanceCount  int
	LightningAbsorption       bool
	LimitedMagicImmunityLevel int // no effect <= level
	MagicResistance           bool
	MagicWeapons              bool
	MartialAdvantageNumDice   int // Nd6 damage def 2d6
	PackTactics               bool
	Reckless                  bool
	ReflectiveCarapace        bool
	RegenerationValue         int // Recover this value at start of turn
	RelentlessThreshold       int // Drop to 1 if damage is less <= threshold
	SneakAttackNumDice        int // Nd6 sneak attack dice
	UndeadFortitude           bool
}

func (s *SpecialAbilities) AddSpecialAbility(name string, value int, dt core.DamageType) error {
	switch strings.ToTitle(name) {
	case SpecAbilityAssassinate:
		s.Assassinate = true
	case SpecAbilityBerserk:
		if value < 1 {
			return fmt.Errorf("berserk threshold must be greater than 0")
		}
		s.BerserkThreshold = value
	case SpecAbilityBloodFrenzy:
		s.BloodFrenzy = true
	case SpecAbilityConsumeLife:
		die, err := core.MakeDiceType(value)
		if err != nil {
			return err
		}
		s.ConsumeLifeDie = die
	case SpecAbilityCorrosiveForm:
		if value < 1 {
			return fmt.Errorf("corrosive form dice must be greater than 0")
		}
		s.CorrosiveFormNumDice = value
	case SpecAbilityDeathBurst:
		if value < 1 {
			return fmt.Errorf("death burst dice must be greater than 0")
		}
		s.DeathBurstNumDice = value
		s.DeathBurstDamageType = dt
	case SpecAbilityDeathThroes:
		if value < 1 {
			return fmt.Errorf("death throes dice must be greater than 0")
		}
		s.DeathThroesNumDice = value
	case SpecAbilityDivineEminence:
		s.DivineEminence = true
	case SpecAbilityEvasion:
		s.Evasion = true
	case SpecAbilityFireAura:
		if value < 1 {
			return fmt.Errorf("fire aura dice must be greater than 0")
		}
		s.FireAuraNumDice = value
	case SpecAbilityFireForm:
		s.FireForm = true
	case SpecAbilityGibbering:
		s.Gibbering = true
	case SpecAbilityGnomeCunning:
		s.GnomeCunning = true
	case SpecAbilityHeatedBody:
		if value < 1 {
			return fmt.Errorf("heated body dice must be greater than 0")
		}
		s.HeatedBodyNumDice = value
	case SpecAbilityLegendaryResistance:
		if value < 1 {
			return fmt.Errorf("legendary resistance count must be greater than 0")
		}
		s.LegendaryResistanceCount = value
	case SpecAbilityLightningAbsorption:
		s.LightningAbsorption = true
	case SpecAbilityLimitedMagicImmunity:
		if value < 1 || value > 9 {
			return fmt.Errorf("limited magic immunity level must be between 1 and 9")
		}
		s.LimitedMagicImmunityLevel = value
	case SpecAbilityMagicResistance:
		s.MagicResistance = true
	case SpecAbilityMagicWeapons:
		s.MagicWeapons = true
	case SpecAbilityMartialAdvantage:
		if value < 1 {
			return fmt.Errorf("martial advantage dice must be greater than 0")
		}
		s.MartialAdvantageNumDice = value
	case SpecAbilityPackTactics:
		s.PackTactics = true
	case SpecAbilityReckless:
		s.Reckless = true
	case SpecAbilityReflectiveCarapace:
		s.ReflectiveCarapace = true
	case SpecAbilityRegeneration:
		if value < 1 {
			return fmt.Errorf("regeneration value must be greater than 0")
		}
		s.RegenerationValue = value
	case SpecAbilityRelentless:
		if value < 1 {
			return fmt.Errorf("relentless threshold must be greater than 0")
		}
		s.RelentlessThreshold = value
	case SpecAbilitySneakAttack:
		if value < 1 {
			return fmt.Errorf("sneak attack dice must be greater than 0")
		}
		s.SneakAttackNumDice = value
	case SpecAbilityUndeadFortitude:
		s.UndeadFortitude = true
	default:
		return fmt.Errorf("invalid special ability: %s", name)
	}

	return nil
}

type SpecialAbility string

const (
	SpecAbilityAssassinate          = "Assassinate"
	SpecAbilityBerserk              = "Berserk"
	SpecAbilityBloodFrenzy          = "Blood Frenzy"
	SpecAbilityConsumeLife          = "Consume Life"
	SpecAbilityCorrosiveForm        = "Corrosive Form"
	SpecAbilityDeathBurst           = "Death Burst"
	SpecAbilityDeathThroes          = "Death Throes"
	SpecAbilityDivineEminence       = "Divine Eminence"
	SpecAbilityEvasion              = "Evasion"
	SpecAbilityFireAura             = "Fire Aura"
	SpecAbilityFireForm             = "Fire Form"
	SpecAbilityGibbering            = "Gibbering"
	SpecAbilityGnomeCunning         = "Gnome Cunning"
	SpecAbilityHeatedBody           = "Heated Body"
	SpecAbilityLegendaryResistance  = "Legendary Resistance"
	SpecAbilityLightningAbsorption  = "Lightning Absorption"
	SpecAbilityLimitedMagicImmunity = "Limited Magic Immunity"
	SpecAbilityMagicResistance      = "Magic Resistance"
	SpecAbilityMagicWeapons         = "Magic Weapons"
	SpecAbilityMartialAdvantage     = "Martial Advantage"
	SpecAbilityPackTactics          = "Pack Tactics"
	SpecAbilityReckless             = "Reckless"
	SpecAbilityReflectiveCarapace   = "Reflective Carapace"
	SpecAbilityRegeneration         = "Regeneration"
	SpecAbilityRelentless           = "Relentless"
	SpecAbilitySneakAttack          = "Sneak Attack (1/Turn)"
	SpecAbilityUndeadFortitude      = "Undead Fortitude"
)
