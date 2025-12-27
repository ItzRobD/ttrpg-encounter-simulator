package events

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
)

type EventHandler interface {
	HandleEvent(event CombatEvent)
}

type UniversalEventHandler struct{}

func (h *UniversalEventHandler) HandleEvent(event CombatEvent) {
	switch e := event.(type) {
	case *MeleeAttackEvent:
		h.handleMeleeAttack(e)
	case *ActionChoiceEvent:
		h.handleActionChoice(e)
	case *SpellChoiceEvent:
		h.handleSpellChoice(e)
	case *SpellAttackEvent:
		h.handleSpellAttack(e)
	case *SpellDCEvent:
		h.handleSpellDC(e)
	case *DamageEvent:
		h.handleDamage(e)
	case *HealEvent:
		h.handleHeal(e)
	case *DeathEvent:
		h.handleDeath(e)
	case *HPModifiedEvent:
		h.handleHPModified(e)
	case *DragonbornBreathWeaponEvent:
		h.handleDragonbornBreathWeapon(e)
	case *UnconsciousEvent:
		h.handleUnconscious(e)
	case *DamageModifiedEvent:
		h.handleDamageModified(e)
	case *DiceRollEvent:
		h.handleDiceRoll(e)
	case *HPRollEvent:
		h.handleHPRoll(e)
	case *TargetChoiceEvent:
		h.handleTargetChoice(e)
	case *SavingThrowEvent:
		h.handleSavingThrow(e)
	case *CombatEventMessage:
		h.handleCombatMessage(e)
	default:
		fmt.Printf("Unknown event type: %T\n", e)
	}

}

func (h *UniversalEventHandler) handleMeleeAttack(e *MeleeAttackEvent) {
	var s string
	if e.CriticalHit {
		s = fmt.Sprintf("[Round %d] <Martial Critical Hit> Attack %d - %s attacks %s with %s. %d to hit, %d + %d. Target: %d. Success: %t. Damage: %d %s\n",
			e.GetRound(), e.AttackCount, e.GetActor(), e.Target, e.AttackName, e.AttackRoll, e.AttackRoll, e.AttackModifier, e.TargetValue, e.Success, e.DamageTotal, e.DamageType)
	} else if !e.Success {
		s = fmt.Sprintf("[Round %d] <Martial Miss> Attack %d - %s attacks %s with %s. %d to hit, %d + %d. Target: %d. Success: %t\n",
			e.GetRound(), e.AttackCount, e.GetActor(), e.Target, e.AttackName, e.AttackTotal, e.AttackRoll, e.AttackModifier, e.TargetValue, e.Success)
	} else {
		s = fmt.Sprintf("[Round %d] <Martial Attack> Attack %d - %s attacks %s with %s. %d to hit, %d + %d. Target: %d. Success: %t. Damage: %d %s\n",
			e.GetRound(), e.AttackCount, e.GetActor(), e.Target, e.AttackName, e.AttackTotal, e.AttackRoll, e.AttackModifier, e.TargetValue, e.Success, e.DamageTotal, e.DamageType)
	}
	fmt.Print(s)
}

func (h *UniversalEventHandler) handleDragonbornBreathWeapon(e *DragonbornBreathWeaponEvent) {
	s := fmt.Sprintf("[Round %d] <Dragonborn Breath Weapon> %s uses breath weapon against %s. DC: %d, %s save: %d. Success: %t. Damage: %d %s\n",
		e.GetRound(), e.GetActor(), e.Target, e.DC, e.SaveAbility, e.SavingThrowResult, e.SavingThrowSuccess, e.DamageTotal, e.DamageType)
	fmt.Print(s)
}

func (h *UniversalEventHandler) handleActionChoice(e *ActionChoiceEvent) {
	fmt.Printf("[Round %d] <Action Choice> %s chooses %s as its action.\n",
		e.GetRound(),
		e.GetActor(),
		e.ActionChoice)
}

func (h *UniversalEventHandler) handleSpellChoice(e *SpellChoiceEvent) {
	fmt.Printf("[Round %d] %s chooses to cast %s at level %d. Formula: %dd%d + %d. Damage type: %s\n",
		e.GetRound(),
		e.GetActor(),
		e.SpellChoice.Spell.GetName(),
		e.SpellChoice.Formula.CastLevel,
		e.SpellChoice.Formula.NumberOfDice,
		e.SpellChoice.Formula.Die,
		e.SpellChoice.Formula.AmountToAdd,
		e.SpellChoice.Formula.DamageType)
}

func (h *UniversalEventHandler) handleSpellAttack(e *SpellAttackEvent) {
	var s string
	switch e.HasDC {
	case true:
		s = fmt.Sprintf("[Round %d] <Spell DC Attack> %s attacks %s with %s at level %d. DC is: %d. %s Saving throw: %d. Save success: %t.",
			e.GetRound(),
			e.GetActor(),
			e.Target,
			e.SpellName,
			e.SpellLevel,
			e.DCValue,
			e.DCAbility,
			e.SavingThrowResult,
			e.SavingThrowSuccess)
		if e.SavingThrowSuccess {
			if e.SaveEffect == "half" {
				s += fmt.Sprintf(" Half damage is dealt: %d %s.\n",
					e.DamageTotal,
					e.DamageType)
			} else if e.SaveEffect == "none" {
				s += fmt.Sprintf(" No damage is dealt.\n")
			}
		} else {
			s += fmt.Sprintf(" Damage: %d %s\n",
				e.DamageTotal,
				e.DamageType)
		}
	case false:
		if e.CriticalHit {
			s = fmt.Sprintf("[Round %d] <Spell Critical Hit> %s attacks %s with %s at level %d. %d to hit, %d + %d. Success: %t. Damage: %d %s\n",
				e.GetRound(), e.GetActor(), e.Target, e.SpellName, e.SpellLevel, e.AttackRoll, e.AttackRoll, e.AttackModifier, e.Success, e.DamageTotal, e.DamageType)
		} else if !e.Success {
			s = fmt.Sprintf("[Round %d] <Spell Miss> %s attacks %s with %s at level %d. %d to hit, %d + %d. Success: %t\n",
				e.GetRound(), e.GetActor(), e.Target, e.SpellName, e.SpellLevel, e.AttackTotal, e.AttackRoll, e.AttackModifier, e.Success)
		} else {
			s = fmt.Sprintf("[Round %d] <Spell Attack> %s attacks %s with %s at level %d. %d to hit, %d + %d. Success: %t. Damage: %d %s\n",
				e.GetRound(), e.GetActor(), e.Target, e.SpellName, e.SpellLevel, e.AttackTotal, e.AttackRoll, e.AttackModifier, e.Success, e.DamageTotal, e.DamageType)
		}
	}

	fmt.Print(s)
}

func (h *UniversalEventHandler) handleSpellDC(e *SpellDCEvent) {
	fmt.Printf("[Round %d] <Spell DC Attack> %s attacks %s with %s. DC is: %d. Saving throw: %d. Success: %t\n",
		e.GetRound(),
		e.GetActor(),
		e.Target,
		e.SpellChoice.Name,
		e.DC,
		e.SavingThrow,
		e.Success)
}

func (h *UniversalEventHandler) handleDamage(e *DamageEvent) {
	fmt.Printf("[Round %d] <Damage> %s Does %d damage to %s. Rolls: %v\n",
		e.GetRound(),
		e.GetActor(),
		e.Amount,
		e.Target,
		e.Rolls)
}

func (h *UniversalEventHandler) handleHeal(e *HealEvent) {
	var s string
	switch e.IsSpell {
	case true:
		s = fmt.Sprintf("[Round %d] <Heal> %s heals %s using %s at level %d for %d hp, rolls: %v.\n",
			e.GetRound(),
			e.GetActor(),
			e.Target,
			e.Name,
			e.SpellLevel,
			e.HealTotal,
			e.HealRolls)
	case false:
		s = fmt.Sprintf("[Round %d] <Heal> %s heals %s using %s for %d hp, rolls: %v.\n",
			e.GetRound(),
			e.GetActor(),
			e.Target,
			e.Name,
			e.HealTotal,
			e.HealRolls)
	}

	if e.Name == "Lay on Hands" {
		s = fmt.Sprintf("[Round %d] <Heal> %s uses Lay on Hands on %s for %d hp.\n",
			e.GetRound(),
			e.GetActor(),
			e.Target,
			e.HealTotal)
	}

	fmt.Print(s)
}

func (h *UniversalEventHandler) handleDeath(e *DeathEvent) {
	fmt.Printf("[Round %d] <Death> %s has died.\n",
		e.GetRound(),
		e.GetActor())
}

func (h *UniversalEventHandler) handleHPModified(e *HPModifiedEvent) {
	fmt.Printf("[Round %d] <HP Modified> %s modified %s's hp by %d. New hp: %d, IsUnconscious: %t\n",
		e.GetRound(),
		e.GetActor(),
		e.SubjectName,
		e.ModificationValue,
		e.NewHP,
		e.IsUnconscious)
}

func (h *UniversalEventHandler) handleDamageModified(e *DamageModifiedEvent) {
	fmt.Printf("[Round %d] <Damage Modified> %s on %s. Original: %d, Final: %d, Modified: %t, Resistance: %s, Broken: %t\n",
		e.GetRound(),
		e.GetActor(),
		e.SubjectName,
		e.OriginalValue,
		e.FinalValue,
		e.WasModified,
		e.ResistanceType,
		e.ResistanceBroken)
}

func (h *UniversalEventHandler) handleUnconscious(e *UnconsciousEvent) {
	fmt.Printf("[Round %d] <Unconscious> %s is unconscious.\n",
		e.GetRound(),
		e.GetActor())
}

func (h *UniversalEventHandler) handleDiceRoll(e *DiceRollEvent) {
	var s string
	switch e.RollType {
	case core.DiceRollGeneral:
		s = fmt.Sprintf("[Round %d] <Roll> %s rolls for %s. Dice: %dd%s, Total: %d, Final rolls: %v, Advantage: %s, Modifier: %d.",
			e.GetRound(),
			e.GetActor(),
			e.RollType,
			e.NumberOfDice,
			e.Die,
			e.Total,
			e.FinalRolls,
			e.Advantage,
			e.Modifier)
	case core.DiceRollInitiative:
		s = fmt.Sprintf("[Round %d] <Initiative> %s rolls for %s. Dice: %dd%s, Total: %d, Final rolls: %v, Advantage: %s, Modifier: %d.",
			e.GetRound(),
			e.GetActor(),
			e.RollType,
			e.NumberOfDice,
			e.Die,
			e.Total,
			e.FinalRolls,
			e.Advantage,
			e.Modifier)
	case core.DiceRollSavingThrow:
		s = fmt.Sprintf("[Round %d] <Saving Throw> %s rolls for %s. Dice: %dd%s, Total: %d, Final rolls: %v, Advantage: %s, Modifier: %d, DC: %d, Success: %t.",
			e.GetRound(),
			e.GetActor(),
			e.RollType,
			e.NumberOfDice,
			e.Die,
			e.Total,
			e.FinalRolls,
			e.Advantage,
			e.Modifier,
			e.TargetValue,
			e.IsSuccess)
		if e.FinalRolls[0] == 0 {
			s += ", AutoFailure: True."
		}
	case core.DiceRollAbilityCheck:
		s = fmt.Sprintf("[Round %d] <AbilityUsed Check> %s rolls for %s. Dice: %dd%s, Total: %d, Final rolls: %v, Advantage: %s, Modifier: %d, Target Value: %d, Success: %t.",
			e.GetRound(),
			e.GetActor(),
			e.RollType,
			e.NumberOfDice,
			e.Die,
			e.Total,
			e.FinalRolls,
			e.Advantage,
			e.Modifier,
			e.TargetValue,
			e.IsSuccess)
	case core.DiceRollHP:
		s = fmt.Sprintf("[Round %d] <HP roll> %s rolls for %s. Dice: %dd%s, Final rolls: %v, Modifier: %d, Total: %d.",
			e.GetRound(),
			e.GetActor(),
			e.RollType,
			e.NumberOfDice,
			e.Die,
			e.FinalRolls,
			e.Modifier,
			e.Total)
	case core.DiceRollHPAvgUsed:
		s = fmt.Sprintf("[Round %d] <HP roll> %s used average values for hp. Dice: %dd%s, Total: %d, Modifier: %d,",
			e.GetRound(),
			e.GetActor(),
			e.NumberOfDice,
			e.Die,
			e.Total,
			e.Modifier)
	case core.DiceRollHPValueUsed:
		s = fmt.Sprintf("[Round %d] <HP roll> %s used direct values for hp. Total: %d.",
			e.GetRound(),
			e.GetActor(),
			e.Total)
	case core.DiceRollAttack:
		s = fmt.Sprintf("[Round %d] <Attack Roll> %s rolls for %s. Dice: %dd%s, Total: %d, Final rolls: %v, Advantage: %s, Modifier: %d, Target Value: %d, Success: %t.",
			e.GetRound(),
			e.GetActor(),
			e.RollType,
			e.NumberOfDice,
			e.Die,
			e.Total,
			e.FinalRolls,
			e.Advantage,
			e.Modifier,
			e.TargetValue,
			e.IsSuccess)
	case core.DiceRollDamage:
		// Clarify that Total includes the modifier; show dice subtotal separately
		s = fmt.Sprintf("[Round %d] <Damage Roll> %s rolls for %s. Dice: %dd%s, DiceTotal: %d, Final Rolls: %v, Modifier: %d, Total: %d.",
			e.GetRound(),
			e.GetActor(),
			e.RollType,
			e.NumberOfDice,
			e.Die,
			e.FinalRollValue,
			e.FinalRolls,
			e.Modifier,
			e.Total)
	case core.DiceRollHealing:
		s = fmt.Sprintf("[Round %d] <Healing Roll> %s rolls for %s. Dice: %dd%s, DiceTotal: %d, Final Rolls: %v, Modifier: %d, Total: %d.",
			e.GetRound(),
			e.GetActor(),
			e.RollType,
			e.NumberOfDice,
			e.Die,
			e.FinalRollValue,
			e.FinalRolls,
			e.Modifier,
			e.Total)
	case core.DiceRollRecharge:
		s = fmt.Sprintf("[Round %d] <Recharge roll> %s rolls to recharge ability: %s. Total: %d. Succcess: %t.",
			e.GetRound(),
			e.GetActor(),
			e.Name,
			e.Total,
			e.IsSuccess)
	}

	if e.WasRerolled {
		s += fmt.Sprintf(" Dice were rerolled, original rolls: %v \n",
			e.OriginalRolls)
	} else {
		s += fmt.Sprintf("\n")
	}

	fmt.Print(s)
}

func (h *UniversalEventHandler) handleHPRoll(e *HPRollEvent) {
	fmt.Printf("[Round %d] <HP Roll> %s rolls %d, rolls: %v, amount to add: %d\n",
		e.GetRound(),
		e.GetActor(),
		e.Value,
		e.Rolls,
		e.Modifier)
}

func (h *UniversalEventHandler) handleTargetChoice(e *TargetChoiceEvent) {
	fmt.Printf("[Round %d] <Target Choice> %s chooses %s as their target.\n",
		e.GetRound(),
		e.GetActor(),
		e.Target)
}

func (h *UniversalEventHandler) handleSavingThrow(e *SavingThrowEvent) {
	fmt.Printf("[Round %d] <Saving Throw> %s rolls saving throw. Result: %d, roll: %d, modifier: %d, success: %t\n",
		e.GetRound(),
		e.GetActor(),
		e.Result,
		e.Roll,
		e.Modifier,
		e.Success)
}

func (h *UniversalEventHandler) handleCombatMessage(e *CombatEventMessage) {
	fmt.Printf("[Round %d] <Combat Message> %s\n",
		e.GetRound(),
		e.Message)
}
