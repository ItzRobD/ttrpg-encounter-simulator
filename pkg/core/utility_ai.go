package core

type HPVisibilityMode int

const (
	HPVisibilityWhite HPVisibilityMode = iota // Exact HP known
	HPVisibilityGray                          // Bloodied status only (below 50%)
	HPVisibilityBlack                         // No HP information
)

type UtilityWeights struct {
	// ActionWeights maps ActionType to a base utility multiplier
	ActionWeights map[ActionType]float64

	// TargetFactorWeights define the influence of various tactical and emotional factors
	TargetFactorWeights struct {
		HighThreat         float64 // Based on LastDamageDealt (The "Big Hit")
		TargetPotency      float64 // Based on Target's inherent stats (AC, Attack Bonus)
		TargetHitability   float64 // Based on Chance to Hit (AC vs actor Attack Bonus)
		Vengeance          float64 // Targeting the last entity that hit this actor
		LowHP              float64 // Based on visibility mode
		CasterPriority     float64 // Prioritizing spellcasters
		ConcentrationBreak float64 // Targets concentrating on spells
		ElitePriority      float64 // Targets with Multiattack/Legendary actions
		EmergencyHeal      float64 // Multiplier for allies near death
	}

	// ResourceExpenditureWeight influences spell slot level and Smite usage
	ResourceExpenditureWeight float64
}

// CalculateHitabilityFactor returns a value between 0 and 1 representing the probability of hitting a target.
func CalculateHitabilityFactor(targetAC int, actorAttackBonus int) float64 {
	// (21 - (TargetAC - ActorAttackBonus)) / 20
	// Represents the number of successful outcomes on a d20 roll (1-20).
	// Natural 1 (always miss) and Natural 20 (always hit) are accounted for.
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
func CalculatePotencyFactor(targetAC int, targetAttackBonus int) float64 {
	// ((Target.AC / 20.0) + (Target.AttackBonus / 10.0)) / 2.0
	// Normalizes a 'Boss' level threat (20 AC, +10 Bonus) to 1.0.
	return ((float64(targetAC) / 20.0) + (float64(targetAttackBonus) / 10.0)) / 2.0
}

// CalculateHPFactor returns a utility weight (0.0 to 1.0) based on HP visibility and current health.
func CalculateHPFactor(hp int, maxHP int, mode HPVisibilityMode) float64 {
	if maxHP <= 0 {
		return 0
	}
	pct := float64(hp) / float64(maxHP)

	switch mode {
	case HPVisibilityWhite:
		// Linear inverse: 1.0 at 0% HP, 0.0 at 100% HP
		return 1.0 - pct
	case HPVisibilityGray:
		// Binary: 1.0 if bloodied (<= 50%), 0.0 otherwise
		if pct <= 0.5 {
			return 1.0
		}
		return 0.0
	case HPVisibilityBlack:
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
	// Scale: 1.0 at 0 HP, 0.0 at avgEnemyDamage HP
	return 1.0 - (float64(hp) / float64(avgEnemyDamage))
}

// ShouldExpendHighResource determines if a high-level resource should be used based on target potency.
func ShouldExpendHighResource(targetPotency float64, resourceWeight float64) bool {
	// Combines inherent target danger with the actor's willingness to spend resources.
	return (targetPotency * resourceWeight) > 0.75
}
