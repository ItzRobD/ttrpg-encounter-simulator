package core

type TargetType string

const (
	TTDamage  TargetType = "damage"
	TTHealing TargetType = "healing"
	TTBuff    TargetType = "buff"
)

type UtilityWeights struct {
	// ActionWeights maps ActivationType to a base utility multiplier
	ActionWeights map[ActivationType]float64 `json:"action_weights"`

	// TargetFactorWeights define the influence of various tactical and emotional factors
	TargetFactorWeights TargetFactorWeights `json:"target_factor_weights"`

	// ResourceExpenditureWeight influences spell slot level and Smite usage
	ResourceExpenditureWeight float64 `json:"resource_expenditure_weight"`
}

type TargetFactorWeights struct {
	HighThreat         float64 `json:"high_threat"`
	TargetPotency      float64 `json:"target_potency"`
	TargetHitability   float64 `json:"target_hitability"`
	Vengeance          float64 `json:"vengeance"`
	LowHP              float64 `json:"low_hp"`
	CasterPriority     float64 `json:"caster_priority"`
	ConcentrationBreak float64 `json:"concentration_break"`
	ElitePriority      float64 `json:"elite_priority"`
	EmergencyHeal      float64 `json:"emergency_heal"`
}

// CalculateHitabilityFactor returns a value between 0 and 1 representing the probability of hitting a target.
func CalculateHitabilityFactor(targetAC int, actorAttackBonus int) float64 {
	needed := targetAC - actorAttackBonus
	if needed <= 2 {
		return 19.0 / 20.0 // 95% chance
	}
	if needed >= 20 {
		return 1.0 / 20.0 // 5% chance
	}
	return float64(21-needed) / 20.0
}

// CalculatePotencyFactor returns a heuristic value representing how dangerous a target is inherently.
func CalculatePotencyFactor(targetAC int, targetAttackBonus int, avgDamage float64, maxDamage float64) float64 {
	// Normalizes a 'Boss' level threat (20 AC, +10 Bonus, 40 Turn Damage, 60 Burst Damage) to ~1.0.
	// We weight turn damage higher than single-hit burst damage for sustained threat assessment.
	return ((float64(targetAC) / 20.0) +
		(float64(targetAttackBonus) / 10.0) +
		(avgDamage / 40.0) +
		(maxDamage / 60.0)) / 4.0
}

// CalculateHPFactor returns a utility weight (0.0 to 1.0) based on HP visibility and current health.
func CalculateHPFactor(hp int, maxHP int, mode HPVisibilityMode) float64 {
	if maxHP <= 0 {
		return 0
	}
	pct := float64(hp) / float64(maxHP)

	switch mode {
	case HPVisible:
		// Linear inverse: 1.0 at 0% HP, 0.0 at 100% HP
		return 1.0 - pct
	case HPPercentage:
		// Binary: 1.0 if bloodied (<= 50%), 0.0 otherwise
		if pct <= 0.5 {
			return 1.0
		}
		return 0.0
	case HPStatusHidden:
		// HP is ignored in the calculation
		return 0.0
	default:
		return 0.0
	}
}

// CalculateEmergencyHealFactor returns a multiplier for healing utility when an ally is in danger.
func CalculateEmergencyHealFactor(hp int, avgEnemyDamage int) float64 {
	if avgEnemyDamage <= 0 || hp >= avgEnemyDamage {
		return 0.0
	}
	return 1.0 - (float64(hp) / float64(avgEnemyDamage))
}
