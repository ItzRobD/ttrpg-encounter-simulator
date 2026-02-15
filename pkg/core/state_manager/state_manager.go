package state_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
)

type StateManager struct {
	// Base State
	CurrentHP   int                  `json:"current_hp"`
	MaxHP       int                  `json:"max_hp"`
	TempHP      int                  `json:"temp_hp"`
	Conditions  core.ActorConditions `json:"conditions"`
	HealthState core.HealthState     `json:"health_state"`

	// Action
	AttackCount              int `json:"attack_count"`
	MaxLegendaryActions      int `json:"max_legendary_actions"`
	ActionUsedCount          int `json:"action_used_count"`
	BonusActionUsedCount     int `json:"bonus_action_used_count"`
	ReactionUsedCount        int `json:"reaction_used_count"`
	LegendaryActionUsedCount int `json:"legendary_action_used_count"`

	// Spell Slots and Innate Uses
	MaxSlots      spells.SpellSlots `json:"max_slots"`
	CurrentSlots  spells.SpellSlots `json:"current_slots"`
	InnateMax     map[string]int    `json:"innate_max"`
	InnateCurrent map[string]int    `json:"innate_current"`

	// Class Resources/Flags
	IsRaging bool           `json:"is_raging"`
	Resource map[string]int `json:"resource"`

	// Simulation Tracking
	HasTakenTurnThisCombat bool            `json:"has_taken_turn_this_combat"`
	OncePerTurnUsed        map[string]bool `json:"once_per_turn_used"`

	// Death Saves
	DeathSaveSuccesses int `json:"death_save_successes"`
	DeathSaveFailures  int `json:"death_save_failures"`

	// Adventuring Day Resources
	MaxHitDice     map[core.DiceType]int `json:"max_hit_dice"`
	CurrentHitDice map[core.DiceType]int `json:"current_hit_dice"`

	// Recovery tracking
	ShortRestRecoveredFeatures []string `json:"short_rest_recovered_features"`
}

// ResetStateForRoundStart resets the counts for actions, bonus actions, reactions, and legendary actions to zero.
func (sm *StateManager) ResetStateForRoundStart() {
	sm.ActionUsedCount = 0
	sm.BonusActionUsedCount = 0
	sm.OncePerTurnUsed = make(map[string]bool)
}

func (sm *StateManager) ResetStateForTurnStart() {
	sm.LegendaryActionUsedCount = 0
	sm.ReactionUsedCount = 0
}

// ResetStateForNewEncounter prepares the actor for a new encounter by clearing temporary conditions
// and resetting action counters, while preserving HP, spell slots, and persistent resources.
func (sm *StateManager) ResetStateForNewEncounter() {
	// Reset action counts
	sm.ResetStateForRoundStart()
	sm.ResetStateForTurnStart()
	sm.HasTakenTurnThisCombat = false

	// Clear necessary conditions
	condToRemove := []core.Condition{
		core.ConditionGrappled,
		core.ConditionIncapacitated,
		core.ConditionParalyzed,
		core.ConditionPetrified,
		core.ConditionStunned,
		core.ConditionFrightened,
		core.ConditionCharmed,
		core.ConditionRestrained,
		core.ConditionBlinded,
		core.ConditionDeafened,
		core.ConditionPoisoned,
		core.ConditionReckless,
		core.ConditionBerserk,
	}

	for _, c := range condToRemove {
		sm.Conditions.Remove(c)
	}
	sm.Conditions.Remove(core.ConditionProne)
	sm.Conditions.Remove(core.ConditionUnconscious)
	sm.Conditions.Remove(core.ConditionStable)

	// Reset death saves
	sm.DeathSaveSuccesses = 0
	sm.DeathSaveFailures = 0

	// Reset attack count (per turn limit)
	sm.AttackCount = 0

	// Re-evaluate health state
	sm.HealthState = sm.GetHealthState(true) // Treat as PC for safe state eval
}

func (sm *StateManager) RemainingLegendaryActionCount() int {
	return sm.MaxLegendaryActions - sm.LegendaryActionUsedCount
}

// GetHealthState evaluates current and maximum HP to determine the health state.
func (sm *StateManager) GetHealthState(isPC bool) core.HealthState {
	if sm.CurrentHP <= 0 {
		if isPC {
			if sm.DeathSaveFailures >= 3 {
				return core.HealthStateDead
			}
			return core.HealthStateUnconscious
		}
		return core.HealthStateDead
	}
	if sm.CurrentHP >= sm.MaxHP {
		return core.HealthStateHealthy
	}

	pct := float64(sm.CurrentHP) / float64(sm.MaxHP)
	if pct >= 0.75 {
		return core.HealthStateWounded
	} else if pct >= 0.15 {
		return core.HealthStateBloody
	}

	return core.HealthStateCritical
}

func (sm *StateManager) ModifyHP(amount int, isTemp bool, isPC bool) core.HPModificationResult {
	res := core.HPModificationResult{
		OriginalHP:     sm.CurrentHP,
		OriginalTempHP: sm.TempHP,
	}

	if isTemp {
		if amount <= 0 {
			// We shouldn't be removing temp hp
			return res
		}
		// Temporary HP doesn't stack; the highest value stays
		if amount > sm.TempHP {
			sm.TempHP = amount
			res.DidHealTempHP = true
		}
		res.NewHP = sm.CurrentHP
		res.NewTempHP = sm.TempHP
		res.ModificationValue = amount
		return res
	}

	if amount > 0 {
		// Healing
		originalHP := sm.CurrentHP
		sm.CurrentHP += amount
		if sm.CurrentHP > sm.MaxHP {
			sm.CurrentHP = sm.MaxHP
			res.IsMaxHealth = true
		}
		res.DidHealHP = true
		res.ModificationValue = sm.CurrentHP - originalHP
		// Reset death saves on healing
		if sm.CurrentHP > 0 {
			sm.DeathSaveFailures = 0
			sm.DeathSaveSuccesses = 0
			sm.Conditions.Remove(core.ConditionStable)
		}
	} else if amount < 0 {
		// Damage
		damage := -amount
		res.ModificationValue = damage

		// Apply damage to temp hp first
		if sm.TempHP > 0 {
			res.DidTempDamage = true
			if damage <= sm.TempHP {
				sm.TempHP -= damage
				res.TempHPUsed = damage
				damage = 0
			} else {
				res.TempHPUsed = sm.TempHP
				damage -= sm.TempHP
				sm.TempHP = 0
			}
		}

		// Apply rollover damage to current hp
		if damage > 0 {
			res.DidHPDamage = true
			sm.CurrentHP -= damage
			if sm.CurrentHP < 0 {
				sm.CurrentHP = 0
			}
		}
	}

	// Update health state after modifying hp
	sm.HealthState = sm.GetHealthState(isPC)
	if sm.HealthState == core.HealthStateUnconscious || sm.HealthState == core.HealthStateDead {
		res.IsUnconscious = true
	}

	res.NewHP = sm.CurrentHP
	res.NewTempHP = sm.TempHP

	return res
}

// CanActConditions determines if the actor can act based on the absence of incapacitating conditions.
func (sm *StateManager) CanActConditions() bool {
	hasPreventingConditions := sm.Conditions.Has(core.ConditionIncapacitated) ||
		sm.Conditions.Has(core.ConditionPetrified) ||
		sm.Conditions.Has(core.ConditionStunned) ||
		sm.Conditions.Has(core.ConditionUnconscious)

	return !hasPreventingConditions
}
