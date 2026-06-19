package actor

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spell_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
)

type Actor struct {
	// Base data
	ID         core.ID        `json:"id"`
	InstanceID int            `json:"instance_id"`
	Name       string         `json:"name"`
	Side       core.Side      `json:"side"`
	ActorType  core.ActorType `json:"actor_type"`
	IsCustom   bool           `json:"is_custom"`

	// Shared data
	Abilities        core.Abilities `json:"abilities"`
	HPConfig         core.HPConfig  `json:"hp_config"`
	AC               int            `json:"ac"`
	ProficiencyBonus int            `json:"proficiency_bonus"`

	// Managers and State
	StateManager state_manager.StateManager         `json:"state_manager"`
	Equipment    equipment_manager.EquipmentManager `json:"equipment"`
	Resistances  core.DamageResistances             `json:"resistances"`
	SpellManager spell_manager.SpellManager         `json:"spell_manager"`

	// Actions
	Actions      []core.Action  `json:"actions,omitempty"`
	SpellActions []core.Action  `json:"spell_actions,omitempty"`
	Features     []core.Feature `json:"features,omitempty"`

	// AI and Behavior
	Behavior BehaviorConfig `json:"behavior"`

	// Type-specific
	Metadata Metadata `json:"metadata"`
}

type BehaviorConfig struct {
	ActionPreference          core.ActionPreference `json:"action_preference"`
	SecondaryActionPreference core.ActionPreference `json:"secondary_action_preference,omitempty"`
	TargetPriority            core.TargetPriority   `json:"target_priority"`
	SecondaryTargetPriority   core.TargetPriority   `json:"secondary_target_priority,omitempty"`
	Weights                   *core.UtilityWeights  `json:"weights,omitempty"`
}

type Metadata struct {
	Level int     `json:"level"`
	CR    float64 `json:"cr"`

	// Character
	ClassID              core.ClassID         `json:"class_id,omitempty"`
	RaceID               core.RaceID          `json:"race_id,omitempty"`
	DragonbornColor      core.DragonbornColor `json:"dragonborn_color,omitempty"`
	DragonbornDamageType core.DamageType      `json:"dragonborn_damage_type,omitempty"`

	// Monster
	MonsterSize         core.MonsterSize    `json:"size,omitempty"`
	MonsterType         core.MonsterType    `json:"type,omitempty"`
	IsLegendary         bool                `json:"is_legendary,omitempty"`
	MaxLegendaryActions int                 `json:"max_legendary_actions,omitempty"`
	SpellcasterMetadata SpellcasterMetadata `json:"spellcaster_metadata,omitempty"`

	// Precalculated Stats
	AverageOffensiveValue float64 `json:"average_offensive_value,omitempty"`
	HighestOffensiveValue float64 `json:"highest_offensive_value,omitempty"`
}

type SpellcasterMetadata struct {
	IsSpellcaster       bool         `json:"is_spellcaster,omitempty"`
	IsInnateCaster      bool         `json:"is_innate_caster,omitempty"`
	SpellcastingAbility core.Ability `json:"spellcasting_ability,omitempty"`
	SpellcastingLevel   int          `json:"spellcasting_level,omitempty"`
}

func (a *Actor) GetResistances() core.DamageResistances {
	// Create a new map to return to avoid modifying the base resistances
	res := make(core.DamageResistances)
	if a.Resistances != nil {
		for dt, r := range a.Resistances {
			res[dt] = r
		}
	} else {
		// Fallback to empty resistances if nil
		res = core.NewDamageResistances()
	}

	if core.ClassID(a.Metadata.ClassID) == core.Barbarian && a.StateManager.IsRaging {
		// Barbarians get physical resistance while raging if they have the feature
		if a.HasFeature(core.SpecAbilityRageResistance) {
			res.SetResistanceType(core.DamageSlashing, core.ResistanceResistant)
			res.SetResistanceType(core.DamageBludgeoning, core.ResistanceResistant)
			res.SetResistanceType(core.DamagePiercing, core.ResistanceResistant)
		}
	}
	return res
}

func (a *Actor) ToConfig() ActorConfig {
	cfg := ActorConfig{
		ID:          a.ID.String(),
		InstanceID:  a.InstanceID,
		Name:        a.Name,
		Side:        a.Side,
		ActorType:   a.ActorType,
		IsCustom:    a.IsCustom,
		Abilities:   a.Abilities,
		HPConfig:    a.HPConfig,
		AC:          a.AC,
		Metadata:    a.Metadata,
		Resistances: a.Resistances,
		Actions:     a.Actions,
		Features:    a.Features,
		Behavior:    a.Behavior,
	}

	// We don't necessarily need to populate EquipmentConfigs/KnownSpellIDs
	// because the hydrated Actor already has the Actions and SpellActions derived from them.
	// But we can include the custom ones if they exist.

	return cfg
}

func NewActorFromConfig(config *ActorConfig) (*Actor, error) {
	a := &Actor{
		ID:         core.MakeID(config.ID),
		InstanceID: config.InstanceID,
		Name:       config.Name,
		Side:       config.Side,
		ActorType:  config.ActorType,
		IsCustom:   config.IsCustom,

		Abilities: config.Abilities,
		HPConfig:  config.HPConfig,
		AC:        config.AC,

		Resistances: config.Resistances,
		Actions:     config.Actions,
		Metadata:    config.Metadata,

		StateManager: state_manager.StateManager{
			HealthState:         core.HealthStateHealthy,
			Conditions:          core.NewActorConditions(),
			Resource:            make(map[string]int),
			MaxSlots:            make(spells.SpellSlots),
			CurrentSlots:        make(spells.SpellSlots),
			InnateMax:           make(map[string]int),
			InnateCurrent:       make(map[string]int),
			OncePerTurnUsed:     make(map[string]bool),
			MaxLegendaryActions: config.Metadata.MaxLegendaryActions,
		},
		Equipment:    equipment_manager.NewEquipmentManager(),
		SpellManager: spell_manager.NewSpellManager(),

		Features: config.Features,
		Behavior: config.Behavior,
	}

	if a.Resistances == nil {
		a.Resistances = core.NewDamageResistances()
	}

	// Always calculate Proficiency Bonus from Metadata
	a.ProficiencyBonus = a.GetProficiencyBonus()

	// Initialize recharge actions as available
	for _, act := range a.Actions {
		if act.RechargeValue > 0 {
			a.StateManager.Resource[act.Name] = 1
		}
	}

	if config.Spellcasting != nil {
		a.StateManager.MaxSlots = config.Spellcasting.SpellSlots
		a.StateManager.CurrentSlots = make(spells.SpellSlots)
		for k, v := range config.Spellcasting.SpellSlots {
			a.StateManager.CurrentSlots[k] = v
		}

		for _, s := range config.Spellcasting.LeveledSpells {
			err := a.SpellManager.AddKnownSpell(&s)
			if err != nil {
				return nil, err
			}
		}
		for _, is := range config.Spellcasting.InnateSpells {
			err := a.SpellManager.AddKnownInnateSpell(&is)
			if err != nil {
				return nil, err
			}
			if is.MaxCastsPerDay > 0 {
				a.StateManager.InnateMax[is.Spell.Name] = is.MaxCastsPerDay
				a.StateManager.InnateCurrent[is.Spell.Name] = is.MaxCastsPerDay
			}
		}
	}

	for _, s := range config.CustomSpells {
		err := a.SpellManager.AddKnownSpell(&s)
		if err != nil {
			return nil, err
		}
	}

	// Initialize HP from Config as a fallback
	if a.StateManager.MaxHP == 0 {
		a.StateManager.MaxHP = config.HPConfig.HPAverage
		a.StateManager.CurrentHP = config.HPConfig.HPAverage
	}

	return a, nil
}
