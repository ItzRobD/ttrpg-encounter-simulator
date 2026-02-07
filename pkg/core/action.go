package core

// ActionType defines a category of actions, such as melee, ranged, spells, or special monster abilities.
type ActionType string

const (
	ATMelee       ActionType = "melee"
	ATOffhand     ActionType = "offhand"
	ATRanged      ActionType = "range"
	ATAction      ActionType = "monster_action"
	ATMultiAttack ActionType = "monster_multiattack"
	ATRecharge    ActionType = "monster_recharge"
	ATLegendary   ActionType = "monster_legendary"
	ATSpell       ActionType = "spell"
	ATFeature     ActionType = "feature"
)

// ActivationType represents the type of action activation required, such as action, bonus, reaction, or legendary.
type ActivationType string

const (
	ActAction    ActivationType = "action"
	ActBonus     ActivationType = "bonus"
	ActReaction  ActivationType = "reaction"
	ActLegendary ActivationType = "legendary"
)

// Multiattack represents a set of multiple attacks associated with a specific action and the number of times it is used.
type Multiattack struct {
	ActionID ID  `json:"action_id"`
	Count    int `json:"count"`
}

// ActionCost represents the cost required to perform an action, including activation type and value.
type ActionCost struct {
	ActivationType ActivationType `json:"activation_activation_type"`
	Value          int            `json:"value"`
}

// Action represents a predefined action that an actor can perform in the game, including attacks, abilities, or spells.
type Action struct {
	// Base Action Data
	ID         ID         `json:"id"`
	Name       string     `json:"name"`
	ActionType ActionType `json:"action_type"`
	Cost       ActionCost `json:"cost"`

	// Attack Data
	AttackBonus     int              `json:"attack_bonus"`
	DiceBlock       []DiceBlock      `json:"dice_block"`
	IsAutoHit       bool             `json:"is_auto_hit"`
	WeaponModifiers *WeaponModifiers `json:"weapon_modifiers,omitempty"`

	Multiattack []Multiattack `json:"multiattack,omitempty"`

	// Saving Throw Data
	HasDC       bool        `json:"has_dc,omitempty"`
	DCSaveDC    int         `json:"dc,omitempty"`
	DCAbility   Ability     `json:"dc_ability,omitempty"`
	DCOnSuccess DCOnSuccess `json:"dc_on_success,omitempty"`

	// Precalculated Stats for AI
	AverageDamage int `json:"average_damage,omitempty"`
	AverageHeal   int `json:"average_heal,omitempty"`

	// Properties
	WeaponProperties *WeaponProperties `json:"weapon_properties,omitempty"`

	// Constraints
	RechargeValue int  `json:"recharge_value,omitempty"`
	CastLevel     int  `json:"cast_level,omitempty"`
	IsInnate      bool `json:"is_innate,omitempty"`
	IsAOE         bool `json:"is_aoe,omitempty"`
}

func (a *Action) IsWeaponAttack() bool {
	switch a.ActionType {
	case ATMelee, ATRanged, ATOffhand:
		return true
	}

	return false
}

func (a *Action) IsMeleeAttack() bool {
	switch a.ActionType {
	case ATMelee, ATOffhand:
		return true
	}

	return false
}
