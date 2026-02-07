package monster

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/entity_configuration"
	"dnd5e-encounter-simulator-backend/pkg_old/spells"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type MonsterSummary struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	CR            float64 `json:"cr"`
	Type          string  `json:"type"`
	Size          string  `json:"size"`
	AC            int     `json:"ac"`
	IsLegendary   bool    `json:"is_legendary"`
	IsSpellcaster bool    `json:"is_spellcaster"`
	IsCustom      bool    `json:"is_custom"`
}
type MonsterConfig struct {
	MonsterBase         `json:",inline"`
	Actions             map[int]Action                           `json:"actions"`
	Multiattacks        map[int][]Multiattack                    `json:"multiattacks"`
	LegendaryActions    map[int]LegendaryAction                  `json:"legendary_actions"`
	Resistances         core.DamageResistances                   `json:"resistances"`
	DamageBreakers      []core.ResistBreaker                     `json:"damage_breakers"`
	SpellcastingConfig  MonsterSpellcastingConfig                `json:"spellcasting"`
	HPMethod            core.HPMethodType                        `json:"hp_method"`
	UtilityWeights      *core.UtilityWeights                     `json:"utility_weights"`
	Seed                core.Seed                                `json:"seed"`
	EntityConfiguration entity_configuration.EntityConfiguration `json:"entity_configuration"`
}

func (mc MonsterConfig) MarshalJSON() ([]byte, error) {
	type Alias MonsterConfig
	aux := struct {
		Alias
		Actions          []Action          `json:"actions"`
		LegendaryActions []LegendaryAction `json:"legendary_actions"`
	}{
		Alias: Alias(mc),
	}

	// Map to slice for Actions
	if mc.Actions != nil {
		aux.Actions = make([]Action, 0, len(mc.Actions))
		for _, v := range mc.Actions {
			aux.Actions = append(aux.Actions, v)
		}
	} else {
		aux.Actions = []Action{}
	}

	// Map to slice for Legendary Actions
	if mc.LegendaryActions != nil {
		aux.LegendaryActions = make([]LegendaryAction, 0, len(mc.LegendaryActions))
		for _, v := range mc.LegendaryActions {
			aux.LegendaryActions = append(aux.LegendaryActions, v)
		}
	} else {
		aux.LegendaryActions = []LegendaryAction{}
	}

	return json.Marshal(aux)
}

type MonsterDetailsResponse struct {
	Data MonsterConfig `json:"data"`
}

type MonsterSpellcastingConfig struct {
	MonsterID      int                  `json:"monster_id"`
	CastingLevel   int                  `json:"casting_level"`
	Ability        core.Ability         `json:"ability"`
	AttackModifier int                  `json:"attack_modifier"`
	SaveDC         int                  `json:"save_dc"`
	LeveledSpells  []spells.Spell       `json:"leveled_spells"`
	InnateSpells   []spells.InnateSpell `json:"innate_spells"`
	SpellSlots     spells.SpellSlots    `json:"spell_slots"`
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
	Assassinate               bool            `json:"assassinate"`
	BerserkThreshold          int             `json:"berserk_threshold"`
	BloodFrenzy               bool            `json:"blood_frenzy"`
	ConsumeLifeDie            core.DiceType   `json:"consume_life_die"`
	CorrosiveFormNumDice      int             `json:"corrosive_form_num_dice"`
	DeathBurstNumDice         int             `json:"death_burst_num_dice"`
	DeathBurstDamageType      core.DamageType `json:"death_burst_damage_type"`
	DeathBurstDC              int             `json:"death_burst_dc"`
	DeathThroesNumDice        int             `json:"death_throes_num_dice"`
	DeathThroesDC             int             `json:"death_throes_dc"`
	DivineEminenceNumDice     int             `json:"divine_eminence_num_dice"`
	Evasion                   bool            `json:"evasion"`
	FireAuraNumDice           int             `json:"fire_aura_num_dice"`
	FireForm                  bool            `json:"fire_form"`
	Gibbering                 bool            `json:"gibbering"`
	GnomeCunning              bool            `json:"gnome_cunning"`
	HeatedBodyNumDice         int             `json:"heated_body_num_dice"`
	LegendaryResistanceCount  int             `json:"legendary_resistance_count"`
	LightningAbsorption       bool            `json:"lightning_absorption"`
	LimitedMagicImmunityLevel int             `json:"limited_magic_immunity_level"`
	MagicResistance           bool            `json:"magic_resistance"`
	MagicWeapons              bool            `json:"magic_weapons"`
	MartialAdvantageNumDice   int             `json:"martial_advantage_num_dice"`
	PackTactics               bool            `json:"pack_tactics"`
	Reckless                  bool            `json:"reckless"`
	ReflectiveCarapace        bool            `json:"reflective_carapace"`
	RegenerationValue         int             `json:"regeneration_value"`
	RelentlessThreshold       int             `json:"relentless_threshold"`
	SneakAttackNumDice        int             `json:"sneak_attack_num_dice"`
	UndeadFortitude           bool            `json:"undead_fortitude"`
}

type SpecialAbilityValues struct {
	value int
	dc    int
}

func (s *SpecialAbilities) AddSpecialAbility(name string, values SpecialAbilityValues, dt core.DamageType) error {
	c := cases.Title(language.English)
	tcName := c.String(name)
	switch tcName {
	case SpecAbilityAssassinate:
		s.Assassinate = true
	case SpecAbilityBerserk:
		if values.value < 1 {
			return fmt.Errorf("berserk threshold must be greater than 0")
		}
		s.BerserkThreshold = values.value
	case SpecAbilityBloodFrenzy:
		s.BloodFrenzy = true
	case SpecAbilityConsumeLife:
		die, err := core.MakeDiceType(values.value)
		if err != nil {
			return err
		}
		s.ConsumeLifeDie = die
	case SpecAbilityCorrosiveForm:
		if values.value < 1 {
			return fmt.Errorf("corrosive form dice must be greater than 0")
		}
		s.CorrosiveFormNumDice = values.value
	case SpecAbilityDeathBurst:
		if values.value < 1 {
			return fmt.Errorf("death burst dice must be greater than 0")
		}
		if s.DeathThroesNumDice > 0 {
			return fmt.Errorf("death burst and death throes are exclusive")
		}
		if values.dc <= 0 {
			return fmt.Errorf("death burst DC must be greater than 0")
		}
		s.DeathBurstDC = values.dc
		s.DeathBurstNumDice = values.value
		s.DeathBurstDamageType = dt
	case SpecAbilityDeathThroes:
		if values.value < 1 {
			return fmt.Errorf("death throes dice must be greater than 0")
		}
		if s.DeathBurstNumDice > 0 {
			return fmt.Errorf("death burst and death throes are exclusive")
		}
		if values.dc <= 0 {
			return fmt.Errorf("death throes DC must be greater than 0")
		}
		s.DeathThroesDC = values.dc
		s.DeathThroesNumDice = values.value
	case SpecAbilityDivineEminence:
		if values.value < 1 {
			return fmt.Errorf("divine eminence dice must be greater than 0")
		}
		s.DivineEminenceNumDice = values.value
	case SpecAbilityEvasion:
		s.Evasion = true
	case SpecAbilityFireAura:
		if values.value < 1 {
			return fmt.Errorf("fire aura dice must be greater than 0")
		}
		s.FireAuraNumDice = values.value
	case SpecAbilityFireForm:
		s.FireForm = true
	case SpecAbilityGibbering:
		s.Gibbering = true
	case SpecAbilityGnomeCunning:
		s.GnomeCunning = true
	case SpecAbilityHeatedBody:
		if values.value < 1 {
			return fmt.Errorf("heated body dice must be greater than 0")
		}
		s.HeatedBodyNumDice = values.value
	case SpecAbilityLegendaryResistance:
		if values.value < 1 {
			return fmt.Errorf("legendary resistance count must be greater than 0")
		}
		s.LegendaryResistanceCount = values.value
	case SpecAbilityLightningAbsorption:
		s.LightningAbsorption = true
	case SpecAbilityLimitedMagicImmunity:
		if values.value < 1 || values.value > 9 {
			return fmt.Errorf("limited magic immunity level must be between 1 and 9")
		}
		s.LimitedMagicImmunityLevel = values.value
	case SpecAbilityMagicResistance:
		s.MagicResistance = true
	case SpecAbilityMagicWeapons:
		s.MagicWeapons = true
	case SpecAbilityMartialAdvantage:
		if values.value < 1 {
			return fmt.Errorf("martial advantage dice must be greater than 0")
		}
		s.MartialAdvantageNumDice = values.value
	case SpecAbilityPackTactics:
		s.PackTactics = true
	case SpecAbilityReckless:
		s.Reckless = true
	case SpecAbilityReflectiveCarapace:
		s.ReflectiveCarapace = true
	case SpecAbilityRegeneration:
		if values.value < 1 {
			// If regeneration value is 0 or less, we skip adding it rather than failing.
			// This handles cases where malformed data might exist in the database.
			return nil
		}
		s.RegenerationValue = values.value
	case SpecAbilityRelentless:
		if values.value < 1 {
			return fmt.Errorf("relentless threshold must be greater than 0")
		}
		s.RelentlessThreshold = values.value
	case SpecAbilitySneakAttack:
		if values.value < 1 {
			return fmt.Errorf("sneak attack dice must be greater than 0")
		}
		s.SneakAttackNumDice = values.value
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
