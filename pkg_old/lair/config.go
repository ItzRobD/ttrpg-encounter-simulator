package lair

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"fmt"
	"math/rand/v2"
	"strings"
)

// LairActionMode defines how an action is resolved.
// Attack: d20 attack vs AC with damage roll on hit.
// DC: saving throw by target(s); on success apply OnSuccess effect.
type LairActionMode string

const (
	LAMAttack LairActionMode = "attack"
	LAMDC     LairActionMode = "dc"
)

func (m LairActionMode) String() string { return string(m) }

func NewLairActionMode(s string) (LairActionMode, error) {
	switch strings.ToLower(s) {
	case "attack":
		return LAMAttack, nil
	case "dc":
		return LAMDC, nil
	default:
		return LAMAttack, fmt.Errorf("invalid lair action mode")
	}
}

// TargetSide declares which side this action will target.
// "characters" will target player characters; "monsters" will target monsters.
type TargetSide string

const (
	TargetCharacters TargetSide = "characters"
	TargetMonsters   TargetSide = "monsters"
)

func (ts TargetSide) String() string { return string(ts) }

func NewTargetSide(s string) (TargetSide, error) {
	switch strings.ToLower(s) {
	case "characters":
		return TargetCharacters, nil
	case "monsters":
		return TargetMonsters, nil
	default:
		return TargetCharacters, fmt.Errorf("invalid target side")
	}
}

// LairActionInput defines one lair action from API/config.
// For Mode==attack, use AttackBonus and damage dice.
// For Mode==dc, use DCAbility+DC and damage dice, with OnSuccess behavior.
type LairActionInput struct {
	Name         string         `json:"name"`
	Mode         LairActionMode `json:"mode"`
	TargetSide   TargetSide     `json:"target_side"`   // per-action side
	TargetPolicy string         `json:"target_policy"` // maps to core.TargetPriority
	IsAOE        bool           `json:"is_aoe"`
	Recharge     int            `json:"recharge"` // 0=none, else value 2-6

	// Attack mode
	AttackBonus int `json:"attack_bonus"`

	// DC mode
	DCAbility core.Ability     `json:"dc_ability"`
	DCValue   int              `json:"dc_value"`
	OnSuccess core.DCOnSuccess `json:"on_success"`

	// Damage formula (used for both modes)
	NumberOfDice int             `json:"number_of_dice"`
	Die          core.DiceType   `json:"die"`
	AmountToAdd  int             `json:"amount_to_add"`
	DamageType   core.DamageType `json:"damage_type"`
}

// LairConfig bundles the lair entity request.
type LairConfig struct {
	Enabled    bool              `json:"enabled"`
	Name       string            `json:"name"`
	Initiative int               `json:"initiative"` // defaults to 20 if zero
	Actions    []LairActionInput `json:"actions"`
}

// NewLairFromConfig constructs a Lair from config and loads all actions.
func NewLairFromConfig(cfg *LairConfig, rng *rand.Rand) (*Lair, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, fmt.Errorf("lair config not enabled")
	}
	if rng == nil {
		return nil, fmt.Errorf("lair RNG must be provided by SimulationManager; got nil")
	}
	l := NewLair(cfg.Name, rng)
	if cfg.Initiative <= 0 {
		// The engine will use the provided initiative on the combatant, not this value.
		// Lair.RollInitiative returns 20 by default, but we preserve provided initiative at combatant level.
	}

	// Load actions
	for i, a := range cfg.Actions {
		if a.Mode == "" {
			a.Mode = LAMAttack
		}
		// Build internal lair action
		lam := LairAction{
			Name:         a.Name,
			Mode:         a.Mode,
			TargetSide:   a.TargetSide,
			TargetPolicy: core.NewPrioritization(strings.ToLower(a.TargetPolicy)),
			IsAOE:        a.IsAOE,
			Recharge:     a.Recharge,
			AttackData: core.AttackData{
				Name: a.Name,
				DamageBlocks: []core.DamageBlock{
					{
						NumberOfDice: a.NumberOfDice,
						Die:          a.Die,
						DamageType:   a.DamageType,
						Modifier:     a.AmountToAdd,
					},
				},
				AttackModifier: a.AttackBonus,
				DamageModifier: 0,
			},
			DCAbility: a.DCAbility,
			DCValue:   a.DCValue,
			OnSuccess: a.OnSuccess,
		}
		l.actionManager.AddLairAction(i, lam)
	}
	return l, nil
}
