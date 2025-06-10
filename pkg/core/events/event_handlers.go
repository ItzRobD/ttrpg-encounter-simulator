package events

import (
	"fmt"
)

type EventHandler interface {
	HandleEvent(event CombatEvent)
}

type AttackHandler struct{}

func (h *AttackHandler) HandleEvent(event CombatEvent) {
	if attackEvent, ok := event.(*MeleeAttackEvent); ok {
		var s string
		if attackEvent.CriticalHit {
			s = fmt.Sprintf("[Round %d] <Critical Hit> %s attacks %s with %s. %d to hit. Success: %t\n",
				attackEvent.GetRound(),
				attackEvent.GetActor(),
				attackEvent.Target,
				attackEvent.AttackName,
				attackEvent.AttackRoll,
				attackEvent.Success)
		} else {
			s = fmt.Sprintf("[Round %d] <Attack> %s attacks %s with %s. %d to hit, %d + %d. Success: %t\n",
				attackEvent.GetRound(),
				attackEvent.GetActor(),
				attackEvent.Target,
				attackEvent.AttackName,
				attackEvent.AttackTotal,
				attackEvent.AttackRoll,
				attackEvent.AttackModifier,
				attackEvent.Success)
		}
		fmt.Printf(s)
	}
}

type ActionChoiceHandler struct{}

func (h *ActionChoiceHandler) HandleEvent(event CombatEvent) {
	if actionChoiceEvent, ok := event.(*ActionChoiceEvent); ok {
		fmt.Printf("[Round %d] <Action Choice> %s chooses %s as its action.\n",
			actionChoiceEvent.GetRound(),
			actionChoiceEvent.GetActor(),
			actionChoiceEvent.ActionChoice)
	}
}

type SpellChoiceHandler struct{}

func (h *SpellChoiceHandler) HandleEvent(event CombatEvent) {
	if spellChoiceEvent, ok := event.(*SpellChoiceEvent); ok {
		castFormula, err := spellChoiceEvent.SpellChoice.GetClosestFormulaToLevel(spellChoiceEvent.CastLevel)
		if err != nil {
			fmt.Printf("Error getting formula for spell %s at level %d: %v\n", spellChoiceEvent.SpellChoice.Name, spellChoiceEvent.CastLevel, err)
			return
		}

		fmt.Printf("[Round %d] %s chooses to cast %s at level %d. Formula: %dd%d + %d. Damage type: %s\n",
			spellChoiceEvent.GetRound(),
			spellChoiceEvent.GetActor(),
			spellChoiceEvent.SpellChoice.Name,
			spellChoiceEvent.CastLevel,
			castFormula.NumberOfDice,
			castFormula.Die,
			castFormula.AmountToAdd,
			castFormula.DamageType)
	}
}

type SpellAttackHandler struct{}

func (h *SpellAttackHandler) HandleEvent(event CombatEvent) {
	if spellAttackEvent, ok := event.(*SpellAttackEvent); ok {
		fmt.Printf("[Round %d] <Spell Attack> %s attacks %s with %s. %d to hit, %d + %d. Success: %t\n",
			spellAttackEvent.GetRound(),
			spellAttackEvent.GetActor(),
			spellAttackEvent.Target,
			spellAttackEvent.SpellChoice.Name,
			spellAttackEvent.AttackTotal,
			spellAttackEvent.AttackRoll,
			spellAttackEvent.AttackModifier,
			spellAttackEvent.Success)
	}
}

type SpellDCHandler struct{}

func (h *SpellDCHandler) HandleEvent(event CombatEvent) {
	if spellDCEvent, ok := event.(*SpellDCEvent); ok {
		fmt.Printf("[Round %d] <Spell DC Attack> %s attacks %s with %s. DC is: %d. Saving throw: %d. Success: %t\n",
			spellDCEvent.GetRound(),
			spellDCEvent.GetActor(),
			spellDCEvent.Target,
			spellDCEvent.SpellChoice.Name,
			spellDCEvent.DC,
			spellDCEvent.SavingThrow,
			spellDCEvent.Success)
	}
}

type DamageHandler struct{}

func (h *DamageHandler) HandleEvent(event CombatEvent) {
	if damageEvent, ok := event.(*DamageEvent); ok {
		fmt.Printf("[Round %d] <Damage> %s Does %d damage to %s.\n",
			damageEvent.GetRound(),
			damageEvent.GetActor(),
			damageEvent.Amount,
			damageEvent.Target)
	}
}

type HealHandler struct{}

func (h *HealHandler) HandleEvent(event CombatEvent) {
	if healEvent, ok := event.(*HealEvent); ok {
		fmt.Printf("[Round %d] <Heal> %s heals %s for %d hp, rolls: %v\n",
			healEvent.GetRound(),
			healEvent.GetActor(),
			healEvent.Target,
			healEvent.Amount,
			healEvent.Rolls)
	}
}

type DeathHandler struct{}

func (h *DeathHandler) HandleEvent(event CombatEvent) {
	if deathEvent, ok := event.(*DeathEvent); ok {
		fmt.Printf("[Round %d] <Death> %s has died.\n",
			deathEvent.GetRound(),
			deathEvent.GetActor())
	}
}

type HPModifiedHandler struct{}

func (h *HPModifiedHandler) HandleEvent(event CombatEvent) {
	if hpModifiedEvent, ok := event.(*HPModifiedEvent); ok {
		fmt.Printf("[Round %d] <HP Modified> %s had hp modified by %+d. Previous hp: %d, Current hp: %d\n",
			hpModifiedEvent.GetRound(),
			hpModifiedEvent.GetActor(),
			hpModifiedEvent.Amount,
			hpModifiedEvent.PreviousHP,
			hpModifiedEvent.CurrentHP)
	}

}

type UnconsciousHandler struct{}

func (h *UnconsciousHandler) HandleEvent(event CombatEvent) {
	if unconsciousEvent, ok := event.(*UnconsciousEvent); ok {
		fmt.Printf("[Round %d] <Unconscious> %s is unconscious.\n",
			unconsciousEvent.GetRound(),
			unconsciousEvent.GetActor())
	}
}

type RollHandler struct{}

func (h *RollHandler) HandleEvent(event CombatEvent) {
	if rollEvent, ok := event.(*DiceRollEvent); ok {
		fmt.Printf("[Round %d] <Roll> %s rolls for %s. Result: %d, rolls: %v, modifier: %d.\n",
			rollEvent.GetRound(),
			rollEvent.GetActor(),
			rollEvent.RollType,
			rollEvent.Value,
			rollEvent.Rolls,
			rollEvent.Modifier)
	}
}

type HPRollHandler struct{}

func (h *HPRollHandler) HandleEvent(event CombatEvent) {
	if hpRollEvent, ok := event.(*HPRollEvent); ok {
		fmt.Printf("[Round %d] <HP Roll> %s rolls %d, rolls: %v, amount to add: %d\n",
			hpRollEvent.GetRound(),
			hpRollEvent.GetActor(),
			hpRollEvent.Value,
			hpRollEvent.Rolls,
			hpRollEvent.Modifier)
	}
}

type TargetChoiceHandler struct{}

func (e *TargetChoiceEvent) HandleEvent(event CombatEvent) {
	if targetChoiceEvent, ok := event.(*TargetChoiceEvent); ok {
		fmt.Printf("[Round %d] <Target Choice> %s chooses %s as their target.\n",
			targetChoiceEvent.GetRound(),
			targetChoiceEvent.GetActor(),
			targetChoiceEvent.Target)
	}
}
