package martial_attacks

// DEPRECATED
//
//import (
//	"dnd5e-encounter-simulator-backend/pkg/core"
//	"dnd5e-encounter-simulator-backend/pkg/core/events"
//	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
//)
//
//type AttackRequest struct {
//	AttackData
//	Modifiers   AttackModifiers
//	Advantage   core.AdvantageType
//	AttackCount int
//}
//type AttackData struct {
//	Name              string
//	NumberOfDice      int
//	Die               int
//	AttackModifier    int
//	DamageModifier    int
//	DamageType        string
//	IsVersatileAttack bool
//}
//
//type AttackModifiers struct {
//	BonusAttackRoll int // Flat bonus, ie magic weapons
//	BonusDamageRoll int // Flat bonus, ie magic weapons, rage, hexblade curse
//
//	ShouldApplyDamageMod bool // Off hand attacks, TWF
//
//	PowerAttack bool // GWM / Sharpshooter (-5 attack, +10 damage) // TODO: Implement logic for this choice
//
//	ImprovedCritical bool // Crits on 19 and 20, Hexblade, Champion
//
//	RerollOnesAndTwos bool // GWF
//	// TODO: GWF Creates an extra attack
//
//	HalflingLucky bool // Reroll 1s on Attacks, Saves
//}
//
//type AttackResult struct {
//	ActorName     string
//	TargetName    string
//	AttackName    string
//	AttackCount   int
//	IsSuccess         bool
//	IsCriticalHit bool
//	AttackTotal   int
//	AttackRoll    int
//	Damage        int
//	DamageRolls   []int
//	DamageType    string
//}
//
//func (r *AttackResult) GetActorName() string   { return r.ActorName }
//func (r *AttackResult) GetTargetName() string  { return r.TargetName }
//func (r *AttackResult) GetAttackName() string  { return r.AttackName }
//func (r *AttackResult) GetAttackCount() int    { return r.AttackCount }
//func (r *AttackResult) GetIsHit() bool         { return r.IsSuccess }
//func (r *AttackResult) GetIsCriticalHit() bool { return r.IsCriticalHit }
//func (r *AttackResult) GetAttackTotal() int    { return r.AttackTotal }
//func (r *AttackResult) GetAttackRoll() int     { return r.AttackRoll }
//func (r *AttackResult) GetDamage() int         { return r.Damage }
//func (r *AttackResult) GetDamageRolls() []int  { return r.DamageRolls }
//func (r *AttackResult) GetDamageType() string  { return r.DamageType }
//
//func calculateDamage(ad AttackData, isCritical bool, modifiers AttackModifiers, options core.SimulationOptions) (int, []int, error) {
//	dmgMod := ad.DamageModifier
//	if !modifiers.ShouldApplyDamageMod {
//		dmgMod = 0
//	}
//	dmgMod += modifiers.BonusDamageRoll
//	if modifiers.PowerAttack {
//		dmgMod += 10
//	}
//
//	var total int
//	var rolls []int
//	var err error
//
//	if isCritical {
//		total, rolls, err = core.CalculateDamageCriticalHit(ad.NumberOfDice, ad.Die, dmgMod, options.UseImprovedCriticals)
//		if err != nil {
//			return 0, nil, err
//		}
//	} else {
//		total, rolls, err = core.DiceRollWithModifier(ad.NumberOfDice, ad.Die, dmgMod)
//		if err != nil {
//			return 0, nil, err
//		}
//	}
//	// TODO: Log dice rolls
//	if modifiers.RerollOnesAndTwos {
//		for i, roll := range rolls {
//			if roll <= 2 {
//				_, newRolls, err := core.RollDice(1, ad.Die)
//				if err != nil {
//					return 0, nil, err
//				}
//				total = total + newRolls[0] - roll
//				rolls[i] = newRolls[0]
//				// TODO: Log new roll replacement, maybe keep an array of rerolls to log easier
//			}
//		}
//	}
//
//	return total, rolls, err
//}
//
//// MakeMartialAttack rolls an attack, determines if it hits the target, and calculates damage if applicable.
//// Returns a boolean indicating a hit, an integer for damage dealt, and an error if any issues occurred.
//// Note: attacker and target can be mutated - must be a pointer
//func MakeMartialAttack(attacker core.Entity, target core.Entity, req *AttackRequest, options core.SimulationOptions) ([]AttackResult, error) {
//	var results []AttackResult
//
//	for i := 0; i < req.AttackCount; i++ {
//		var res AttackResult
//		// Make the req roll
//		attackMod := req.AttackData.AttackModifier + req.Modifiers.BonusAttackRoll
//
//		if req.Modifiers.PowerAttack {
//			attackMod -= 5
//		}
//
//		options := roll_manager.RollOptions{
//			Advantage:         0,
//			Modifier:          0,
//			RerollAbilities:   roll_manager.RerollAbilities{},
//			CriticalThreshold: 0,
//			TreatOnesAsTwos:   false,
//			RollType:          "",
//			RollContext:       "",
//			TargetValue:       0,
//		}
//
//		var attackTotal int
//		var attackRoll int
//		attackTotal, attackRoll, err := core.AttackRoll(attackMod, req.Advantage)
//		if err != nil {
//			return nil, err
//		}
//
//		// TODO: Note that halfling lucky does not use advantage again, re-roll lower
//		if attackRoll == 1 && req.Modifiers.HalflingLucky {
//			attackTotal, attackRoll, err = core.AttackRoll(attackMod, req.Advantage)
//			if err != nil {
//				return nil, err
//			}
//		}
//
//		res.ActorName = attacker.GetName()
//		res.TargetName = target.GetName()
//		res.AttackName = req.Name
//		res.AttackTotal = attackTotal
//		res.AttackRoll = attackRoll
//		res.AttackCount = i
//
//		events.LogDiceRollEvent(attacker, attackTotal, []int{attackRoll}, core.DiceRollAttack, attackMod, attacker.GetEventListener())
//
//		// Check if the req hits
//		critThreshold := 20
//		if req.Modifiers.ImprovedCritical {
//			critThreshold = 19
//		}
//		isCrit := core.IsCriticalHit(attackRoll, critThreshold)
//		if (isCrit || core.DoesAttackHit(attackTotal, target.GetAC())) && attackRoll != 1 {
//			res.IsSuccess = true
//			res.IsCriticalHit = isCrit
//
//			events.LogMeleeAttackEvent(attacker, target, &res, attacker.GetEventListener())
//
//			damage, rolls, err2 := calculateDamage(req.AttackData, isCrit, req.Modifiers, options)
//			if err2 != nil {
//				return nil, err2
//			}
//			res.Damage = damage
//			res.DamageRolls = rolls
//			res.DamageType = req.AttackData.DamageType
//			//events.LogDiceRollEvent(attacker, res.Damage, res.DamageRolls, shared.DiceRollDamage, req.AttackData.DamageModifier, attacker.GetEventListener())
//		} else {
//			// Log a miss
//			events.LogMeleeAttackEvent(attacker, target, &res, attacker.GetEventListener())
//		}
//
//		// Add each req to the results slice
//		results = append(results, res)
//	}
//
//	return results, nil
//}
