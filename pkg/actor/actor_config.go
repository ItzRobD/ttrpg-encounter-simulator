package actor

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/equipment"
	"dnd5e-encounter-simulator-backend/pkg/spells"
)

type ActorConfig struct {
	// Base data
	ID        string         `json:"ID"`
	Name      string         `json:"name"`
	Side      core.Side      `json:"side"`
	ActorType core.ActorType `json:"actor_type"`
	IsCustom  bool           `json:"is_custom"`

	// Raw Stats
	Abilities core.Abilities `json:"abilities"`
	HPConfig  core.HPConfig  `json:"hp_config"`
	AC        int            `json:"ac"`

	Metadata Metadata `json:"metadata"`

	// SRD/Saved Equipment/Spells
	EquipmentConfigs []equipment.EquipmentConfig `json:"equipment_configs,omitempty"`
	KnownSpellIDs    []int                       `json:"known_spell_ids,omitempty"`

	// Unsaved Custom Equipment/Spells
	CustomEquipment []equipment.Equipment `json:"custom_equipment,omitempty"`
	CustomSpells    []spells.Spell        `json:"custom_spells,omitempty"`

	// Actions
	Actions      []core.Action `json:"actions,omitempty"`
	SpellActions []core.Action `json:"spell_actions,omitempty"`

	// Resistances
	Resistances core.DamageResistances `json:"resistances,omitempty"`

	// Spellcasting
	Spellcasting *MonsterSpellcastingConfig `json:"spellcasting,omitempty"`

	// Features
	Features []core.Feature `json:"features,omitempty"`

	// AI and Behavior
	Behavior BehaviorConfig `json:"behavior,omitempty"`
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
