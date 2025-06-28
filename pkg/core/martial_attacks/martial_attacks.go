package martial_attacks

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/shared"
)

type AttackRequest struct {
	AttackData
	Modifiers   AttackModifiers
	Advantage   shared.AdvantageType
	AttackCount int
}
type AttackData struct {
	Name              string
	NumberOfDice      int
	Die               int
	AttackModifier    int
	DamageModifier    int
	DamageType        string
	IsVersatileAttack bool
}

type AttackModifiers struct {
	BonusAttackRoll int // Flat bonus, ie magic weapons
	BonusDamageRoll int // Flat bonus, ie magic weapons, rage, hexblade curse

	ShouldApplyDamageMod bool // Off hand attacks, TWF

	PowerAttack bool // GWM / Sharpshooter (-5 attack, +10 damage) // TODO: Implement logic for this choice

	ImprovedCritical bool // Crits on 19 and 20, Hexblade, Champion
	TreatOnesAsTwos  bool // Elemental Adept

	RerollOnesAndTwos bool // GWF
	// TODO: GWF Creates an extra attack

	HalflingLucky bool // Reroll 1s on Attacks, Saves
}

type AttackResult struct {
	AttackerName  string
	TargetName    string
	AttackName    string
	AttackCount   int
	IsHit         bool
	IsCriticalHit bool
	AttackTotal   int
	AttackRoll    int
	Damage        int
	DamageRolls   []int
	DamageType    string
}

func DoesAttackHit(attackTotal int, ac int) bool {
	if attackTotal >= ac {
		return true
	}
	return false
}

func isCriticalHit(attackRoll int, critThreshold int) bool {
	if attackRoll >= critThreshold {
		return true
	}
	return false
}

func CalculateDamage(ad AttackData, isCritical bool, modifiers AttackModifiers, options core.Options) (int, []int, error) {
	dmgMod := ad.DamageModifier
	if !modifiers.ShouldApplyDamageMod {
		dmgMod = 0
	}
	dmgMod += modifiers.BonusDamageRoll
	if modifiers.PowerAttack {
		dmgMod += 10
	}

	var total int
	var rolls []int
	var err error

	if isCritical {
		total, rolls, err = shared.CalculateDamageCriticalHit(ad.NumberOfDice, ad.Die, dmgMod, options.UseImprovedCriticals)
		if err != nil {
			return 0, nil, err
		}
	} else {
		total, rolls, err = shared.DiceRollWithModifier(ad.NumberOfDice, ad.Die, dmgMod)
		if err != nil {
			return 0, nil, err
		}
	}
	// TODO: Log dice rolls
	if modifiers.RerollOnesAndTwos {
		for i, roll := range rolls {
			if roll <= 2 {
				_, newRolls, err := shared.RollDice(1, ad.Die)
				if err != nil {
					return 0, nil, err
				}
				total = total + newRolls[0] - roll
				rolls[i] = newRolls[0]
				// TODO: Log new roll replacement, maybe keep an array of rerolls to log easier
			}
		}
	}
	if modifiers.TreatOnesAsTwos {
		for i, roll := range rolls {
			if roll == 1 {
				rolls[i] = 2
				total += 1
				// TODO: Log adding an additional point of damage for the modifier
			}
		}
	}

	return total, rolls, err
}

// MakeMartialAttack rolls an attack, determines if it hits the target, and calculates damage if applicable.
// Returns a boolean indicating a hit, an integer for damage dealt, and an error if any issues occurred.
// Note: attacker and target can be mutated - must be a pointer
func MakeMartialAttack(attacker core.Entity, target core.Entity, attack AttackRequest, options core.Options) ([]AttackResult, error) {
	var results []AttackResult

	for i := 0; i < attack.AttackCount; i++ {
		var res AttackResult
		// Make the attack roll
		attackMod := attack.AttackData.AttackModifier + attack.Modifiers.BonusAttackRoll

		if attack.Modifiers.PowerAttack {
			attackMod -= 5
		}

		var attackTotal int
		var attackRoll int
		attackTotal, attackRoll, err := shared.AttackRoll(attackMod, attack.Advantage)
		if err != nil {
			return nil, err
		}

		if attackRoll == 1 && attack.Modifiers.HalflingLucky {
			attackTotal, attackRoll, err = shared.AttackRoll(attackMod, attack.Advantage)
			if err != nil {
				return nil, err
			}
		}

		res.AttackerName = attacker.GetName()
		res.TargetName = target.GetName()
		res.AttackName = attack.Name
		res.AttackTotal = attackTotal
		res.AttackRoll = attackRoll
		res.AttackCount = i

		events.LogDiceRollEvent(attacker, attackTotal, []int{attackRoll}, shared.DiceRollAttack, attackMod, attacker.GetEventListener())

		// Check if the attack hits
		critThreshold := 20
		if attack.Modifiers.ImprovedCritical {
			critThreshold = 19
		}
		isCrit := isCriticalHit(attackRoll, critThreshold)
		if (isCrit || DoesAttackHit(attackTotal, target.GetAC())) && attackRoll != 1 {
			res.IsHit = true
			res.IsCriticalHit = isCrit

			events.LogMeleeAttackEvent(attacker, target, res, attacker.GetEventListener())

			damage, rolls, err2 := CalculateDamage(attack.AttackData, isCrit, attack.Modifiers, options)
			if err2 != nil {
				return nil, err2
			}
			res.Damage = damage
			res.DamageRolls = rolls
			res.DamageType = attack.AttackData.DamageType
			//events.LogDiceRollEvent(attacker, res.Damage, res.DamageRolls, shared.DiceRollDamage, attack.AttackData.DamageModifier, attacker.GetEventListener())
		} else {
			// Log a miss
			events.LogMeleeAttackEvent(attacker, target, res, attacker.GetEventListener())
		}

		// Add each attack to the results slice
		results = append(results, res)
	}

	return results, nil
}
