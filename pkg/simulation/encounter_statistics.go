package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
)

type CombatStatistics struct {
	TotalDamageDealt       int         `json:"total_damage_dealt"`
	TotalHealingDone       int         `json:"total_healing_done"`
	DamageByRound          map[int]int `json:"damage_by_round"`
	HealingByRound         map[int]int `json:"healing_by_round"`
	LastDamageDealt        int         `json:"last_damage_dealt"`
	LastHealingDone        int         `json:"last_healing_done"`
	AverageDamagePerRound  float64     `json:"average_damage_per_round"`
	AverageHealingPerRound float64     `json:"average_healing_per_round"`

	// Attack patterns
	AttacksMade    int `json:"attacks_made"`
	AttacksHit     int `json:"attacks_hit"`
	AttacksMissed  int `json:"attacks_missed"`
	HealingActions int `json:"healing_actions"`
	CriticalHits   int `json:"critical_hits"`

	// Defensive stats
	TimesDamaged         int `json:"times_damaged"`
	TimesHealed          int `json:"times_healed"`
	TotalDamageTaken     int `json:"total_damage_taken"`
	TotalHealingReceived int `json:"total_healing_received"`
	DeathSaveSuccesses   int `json:"death_save_successes"`
	DeathSaveFailures    int `json:"death_save_failures"`

	// Premium AI Tracking
	LastAttackerID int `json:"last_attacker_id"` // id of the last entity to deal damage to this combatant
}

func NewCombatStatistics() *CombatStatistics {
	return &CombatStatistics{
		DamageByRound:  make(map[int]int),
		HealingByRound: make(map[int]int),
	}
}

type EncounterStatistics struct {
	actors                map[int]*actor.Actor
	statistics            map[int]*CombatStatistics
	totalRounds           int
	NeedsHealing          []int // Slice of Actor InstanceIDs currently below threshold
	NeedsEmergencyHealing []int // Slice of Actor InstanceIDs currently below emergency threshold
}

func NewEncounterStatistics(actors map[int]*actor.Actor) *EncounterStatistics {
	ca := &EncounterStatistics{
		actors:                actors,
		statistics:            make(map[int]*CombatStatistics, len(actors)),
		totalRounds:           1,
		NeedsHealing:          make([]int, 0),
		NeedsEmergencyHealing: make([]int, 0),
	}

	ca.InitCombatStatistics()
	return ca
}

func (es *EncounterStatistics) InitCombatStatistics() {
	for id, _ := range es.actors {
		es.statistics[id] = NewCombatStatistics()
	}
}

func (es *EncounterStatistics) IncrementRound() {
	es.totalRounds++
}

// MarkNeedsHealing adds an actor to the NeedsHealing registry if not already present.
func (es *EncounterStatistics) MarkNeedsHealing(actorID int) {
	for _, id := range es.NeedsHealing {
		if id == actorID {
			return
		}
	}
	es.NeedsHealing = append(es.NeedsHealing, actorID)
}

// ClearNeedsHealing removes an actor from the NeedsHealing registry.
func (es *EncounterStatistics) ClearNeedsHealing(actorID int) {
	for i, id := range es.NeedsHealing {
		if id == actorID {
			es.NeedsHealing = append(es.NeedsHealing[:i], es.NeedsHealing[i+1:]...)
			return
		}
	}
}

// GetAverageHealingNeeded calculates the average amount of healing needed for actors on a specific side.
func (es *EncounterStatistics) GetAverageHealingNeededForSide(side core.Side) int {
	if len(es.NeedsHealing) == 0 {
		return 0
	}

	totalMissingHP := 0
	for _, id := range es.NeedsHealing {
		if es.actors[id].Side != side {
			continue
		}
		actor := es.actors[id]
		missing := actor.StateManager.MaxHP - actor.StateManager.CurrentHP
		totalMissingHP += missing
	}

	return totalMissingHP / len(es.NeedsHealing)
}

// GetHPDiffValuePerSide returns a map of actor IDs to the difference between their current HP and max HP.
func (es *EncounterStatistics) GetHPDiffValuePerSide(side core.Side) map[int]int {
	hpDiff := make(map[int]int)

	// Combine NeedsHealing and NeedsEmergencyHealing to ensure all targets are checked
	uniqueIDs := make(map[int]bool)
	for _, id := range es.NeedsHealing {
		uniqueIDs[id] = true
	}
	for _, id := range es.NeedsEmergencyHealing {
		uniqueIDs[id] = true
	}

	if len(uniqueIDs) == 0 {
		return nil
	}

	for id := range uniqueIDs {
		if es.actors[id].Side != side {
			continue
		}
		actor := es.actors[id]
		hpDiff[id] = actor.StateManager.MaxHP - actor.StateManager.CurrentHP
	}
	return hpDiff
}

// MarkNeedsEmergencyHealing adds an actor to the NeedsEmergencyHealing registry if not already present.
func (es *EncounterStatistics) MarkNeedsEmergencyHealing(actorID int) {
	for _, id := range es.NeedsEmergencyHealing {
		if id == actorID {
			return
		}
	}
	es.NeedsEmergencyHealing = append(es.NeedsEmergencyHealing, actorID)
}

// ClearNeedsEmergencyHealing removes an actor from the NeedsEmergencyHealing registry.
func (es *EncounterStatistics) ClearNeedsEmergencyHealing(actorID int) {
	for i, id := range es.NeedsEmergencyHealing {
		if id == actorID {
			es.NeedsEmergencyHealing = append(es.NeedsEmergencyHealing[:i], es.NeedsEmergencyHealing[i+1:]...)
			return
		}
	}
}

// AddAttack records an attack event between two actors, updating their combat statistics.
func (es *EncounterStatistics) AddAttack(attackerID int, targetID int, hit bool, crit bool, damage int) {
	if _, ok := es.actors[attackerID]; !ok {
		return
	}
	es.statistics[attackerID].AttacksMade++
	es.statistics[attackerID].LastDamageDealt = damage
	if hit {
		es.statistics[attackerID].AttacksHit++
		es.statistics[targetID].TotalDamageTaken += damage
		es.statistics[targetID].TimesDamaged++
		es.statistics[targetID].LastAttackerID = attackerID
		es.statistics[attackerID].DamageByRound[es.totalRounds] += damage
		es.statistics[attackerID].TotalDamageDealt += damage
		if crit {
			es.statistics[attackerID].CriticalHits++
		}
	} else {
		es.statistics[attackerID].AttacksMissed++
	}
	es.statistics[attackerID].AverageDamagePerRound = float64(es.statistics[attackerID].TotalDamageDealt) / float64(es.totalRounds)
}

// AddDamage updates the target's combat statistics by adding the specified damage and incrementing TimesDamaged.
// Example usage: Corrosive form; Adds damage but is not the direct result of an attack action
func (es *EncounterStatistics) AddDamage(targetID int, damage int) {
	es.statistics[targetID].TotalDamageTaken += damage
	es.statistics[targetID].TimesDamaged++
}

// AddHeal records a healing event between two actors, updating their combat statistics.
func (es *EncounterStatistics) AddHeal(healerID int, targetID int, value int) {
	if _, ok := es.actors[healerID]; !ok {
		return
	}

	es.statistics[healerID].HealingActions++
	es.statistics[healerID].LastHealingDone = value
	es.statistics[healerID].HealingByRound[es.totalRounds] += value
	es.statistics[healerID].TotalHealingDone += value
	es.statistics[healerID].AverageHealingPerRound = float64(es.statistics[healerID].TotalHealingDone) / float64(es.totalRounds)

	es.statistics[targetID].TimesHealed++
	es.statistics[targetID].TotalHealingReceived += value
}

func (es *EncounterStatistics) DeathSave(actorID int, success bool) {
	if _, ok := es.actors[actorID]; !ok {
		return
	}
	if success {
		es.statistics[actorID].DeathSaveSuccesses++
	} else {
		es.statistics[actorID].DeathSaveFailures++
	}
}

func (es *EncounterStatistics) ResetDeathSaveStats(actorID int) {
	if stat, ok := es.statistics[actorID]; ok {
		stat.DeathSaveSuccesses = 0
		stat.DeathSaveFailures = 0
	}
}

func (es *EncounterStatistics) GetCombatant(id int) *CombatStatistics {
	if _, ok := es.statistics[id]; !ok {
		return nil
	}
	return es.statistics[id]
}

// GetHighestSingleDamageActorID returns the ID of the actor with the highest single instance of damage dealt of the specified type.
func (es *EncounterStatistics) GetHighestSingleDamageActorID(requestedType core.ActorType) int {
	return getHighestActorID(es, requestedType, func(stat *CombatStatistics) int {
		return stat.LastDamageDealt
	})
}

// GetHighestSingleHealingActorID returns the ID of the actor with the highest single instance of healing done of the specified type.
func (es *EncounterStatistics) GetHighestSingleHealingActorID(requestedType core.ActorType) int {
	return getHighestActorID(es, requestedType, func(stat *CombatStatistics) int {
		return stat.TotalHealingDone
	})
}

// GetHighestTotalDamageActorID returns the ID of the actor with the highest total damage dealt of the specified type.
func (es *EncounterStatistics) GetHighestTotalDamageActorID(requestedType core.ActorType) int {
	return getHighestActorID(es, requestedType, func(stat *CombatStatistics) int {
		return stat.TotalDamageTaken
	})
}

// GetHighestTotalHealingActorID returns the ID of the actor with the highest total healing received of the specified type.
func (es *EncounterStatistics) GetHighestTotalHealingActorID(requestedType core.ActorType) int {
	return getHighestActorID(es, requestedType, func(stat *CombatStatistics) int {
		return stat.TotalHealingReceived
	})
}

// GetHighestAverageDamageActorID returns the ID of the actor with the highest average damage dealt of the specified type.
func (es *EncounterStatistics) GetHighestAverageDamageActorID(requestedType core.ActorType) int {
	return getHighestActorID(es, requestedType, func(stat *CombatStatistics) float64 {
		return stat.AverageDamagePerRound
	})
}

// GetHighestAverageHealingActorID returns the ID of the actor with the highest average healing done of the specified type.
func (es *EncounterStatistics) GetHighestAverageHealingActorID(requestedType core.ActorType) int {
	return getHighestActorID(es, requestedType, func(stat *CombatStatistics) float64 {
		return stat.AverageHealingPerRound
	})
}

// GetLastAttackerID returns the ID of the last entity to deal damage to the specified actor.
func (es *EncounterStatistics) GetLastAttackerID(id int) int {
	return es.statistics[id].LastAttackerID
}

// GetHitRatePct returns the percentage of attacks that hit the specified actor.
func (es *EncounterStatistics) GetHitRatePct(id int) int {
	return es.statistics[id].AttacksHit * 100 / es.statistics[id].AttacksMade
}

// compareFunc is a helper type for extracting the value we want to compare
type compareValue interface {
	int | float64
}

func getHighestActorID[T compareValue](
	es *EncounterStatistics,
	requestedType core.ActorType,
	getValue func(*CombatStatistics) T,
) int {
	var maxVal T
	bestID := 0
	first := true

	for id, stat := range es.statistics {
		// Filter by ActorType if requested
		if requestedType != core.ActorTypeNone && es.actors[id].ActorType != requestedType {
			continue
		}

		currentVal := getValue(stat)
		if first || currentVal > maxVal {
			maxVal = currentVal
			bestID = id
			first = false
		}
	}
	return bestID
}
