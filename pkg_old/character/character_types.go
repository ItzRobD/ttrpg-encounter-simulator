package character

import (
	"dnd5e-encounter-simulator-backend/pkg_old/classes"
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/entity_configuration"
	"dnd5e-encounter-simulator-backend/pkg_old/races"
	"dnd5e-encounter-simulator-backend/pkg_old/spells"
	"dnd5e-encounter-simulator-backend/pkg_old/weapon"
)

type CharacterSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RaceID   int    `json:"race_id"`
	ClassID  int    `json:"class_id"`
	Level    uint8  `json:"level"`
	IsCustom bool   `json:"is_custom"`
}
type CharacterConfig struct {
	ID                  string                                   `json:"id"`
	Name                string                                   `json:"name"`
	Type                string                                   `json:"type"`
	Size                string                                   `json:"size"`
	CR                  float64                                  `json:"cr"`
	AC                  int                                      `json:"ac"`
	RaceID              races.RaceID                             `json:"race_id"`
	DragonbornColor     *races.DragonbornColor                   `json:"dragonborn_color"`
	ClassID             classes.ClassID                          `json:"class_id"`
	Level               uint8                                    `json:"level"`
	AsConfig            core.AbilityScoresConfig                 `json:"as_config"`
	HPConfig            core.HPConfig                            `json:"hp"`
	Seed                core.Seed                                `json:"seed"`
	Equipment           EquipmentConfig                          `json:"equipment"`
	Resistances         core.DamageResistances                   `json:"resistances"`
	FightingStyles      []classes.FightingStyle                  `json:"fighting_styles"`
	KnownSpells         []int                                    `json:"known_spells"`
	UtilityWeights      *core.UtilityWeights                     `json:"utility_weights"`
	EntityConfiguration entity_configuration.EntityConfiguration `json:"entity_configuration"`
	InstanceID          int                                      `json:"instance_id"`
	IsCustom            bool                                     `json:"is_custom"`
}

type ArmorSummary struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ArmorClass int    `json:"ac"`
	MinimumStr int    `json:"minimum_strength"`
}

type WeaponSlotConfig struct {
	WeaponID     int                   `json:"weapon_id"`
	IsProficient bool                  `json:"is_proficient"`
	Modifiers    *weapon.Modifiers     `json:"modifiers"`
	WeaponData   *weapon.WeaponSummary `json:"weapon_data,omitempty"`
}

// EquipmentConfig defines configuration for a character's equipment including armor and weapon slot mapping.
type EquipmentConfig struct {
	ArmorID           int                `json:"armor_id"`
	ArmorData         *ArmorSummary      `json:"armor_data,omitempty"`
	ShieldData        *ArmorSummary      `json:"shield_data,omitempty"`
	HasShieldEquipped bool               `json:"has_shield_equipped"`
	PrimarySlot       []WeaponSlotConfig `json:"primary_slot"`
	SecondarySlot     []WeaponSlotConfig `json:"secondary_slot"`
	RangedSlot        []WeaponSlotConfig `json:"ranged_slot"`
}

type CharacterSpellcastingConfig struct {
	CastingLevel   int                  `json:"casting_level"`
	Ability        core.Ability         `json:"ability"`
	AttackModifier int                  `json:"attack_modifier"`
	SaveDC         int                  `json:"save_dc"`
	SpellSlots     spells.SpellSlots    `json:"spell_slots"`
	KnownSpells    []spells.Spell       `json:"known_spells"`
	InnateSpells   []spells.InnateSpell `json:"innate_spells"`
}

type CharacterDetailsResponse struct {
	Data CharacterConfigWithSpellcasting `json:"data"`
}

type CharacterConfigWithSpellcasting struct {
	CharacterConfig `json:",inline"`
	Spellcasting    *CharacterSpellcastingConfig `json:"spellcasting"`
}
