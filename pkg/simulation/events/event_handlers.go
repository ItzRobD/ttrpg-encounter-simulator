package events

import (
	"fmt"
)

type ActionChoiceHandler struct{}

func (h *ActionChoiceHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETActionChoiceEvent {
		fmt.Printf("[Round %d] %s chooses %s as its action.\n",
			event.Round, event.Actor, event.ActionChoice)
	}
}

type SpellChoiceHandler struct{}

func (h *SpellChoiceHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETSpellChoiceEvent {
		fmt.Printf("[Round %d] %s chooses to cast %s at level %d. Formula: %dd%d + %d. Damage type: %s\n",
			event.Round, event.Actor, event.SpellChoice.Name, event.SpellChoice.CastFormula.CastLevel,
			event.SpellChoice.CastFormula.NumberOfDice, event.SpellChoice.CastFormula.Die,
			event.SpellChoice.CastFormula.AmountToAdd, event.SpellChoice.CastFormula.DamageType)
	}
}

type AttackHandler struct{}

func (h *AttackHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETAttackEvent {
		fmt.Printf("[Round %d] <Attack> %s attacks %s with %s. %d to hit. Hit: %t\n",
			event.Round, event.Actor, event.Target, event.Attack, event.Value, event.Hit)
	}
}

type SpellAttackHandler struct{}

func (h *SpellAttackHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETSpellAttack {
		fmt.Printf("[Round %d] <Spell Attack> %s attacks %s with %s. %d to hit. Hit: %t\n",
			event.Round, event.Actor, event.Target, event.Attack, event.Value, event.Hit)

	}
}

type SpellDCHandler struct{}

func (h *SpellDCHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETSpellAttack {
		fmt.Printf("[Round %d] <Spell DC Attack> %s attacks %s with %s. DC is: %d. Saving throw: %d. Hit: %t\n",
			event.Round, event.Actor, event.Target, event.Attack, event.Value, event.SavingThrow, event.Hit)

	}
}

type DamageHandler struct{}

func (h *DamageHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETDamageEvent {
		fmt.Printf("[Round %d] <Damage> %s Does %d damage to %s.\n",
			event.Round, event.Actor, event.Value, event.Target)
	}
}

type HealHandler struct{}

func (h *HealHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETHealEvent {
		fmt.Printf("[Round %d] <Heal> %s heals %s for %d hp, rolls: %v\n",
			event.Round, event.Actor, event.Target, event.Value, event.Rolls)
	}
}

type DeathHandler struct{}

func (h *DeathHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETDeathEvent {
		fmt.Printf("[Round %d] <Death> %s has died.\n", event.Round, event.Target)
	}
}

type HPModifiedHandler struct{}

func (h *HPModifiedHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETHPModifiedEvent {
		fmt.Printf("[Round %d] <HP Modified> %s had hp modified by %+d. Previous hp: %d, Current hp: %d\n", event.Round, event.Actor, event.Value, event.PreviousHP, event.CurrentHP)
	}
}

type UnconsciousHandler struct{}

func (h *UnconsciousHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETUnconsciousEvent {
		fmt.Printf("[Round %d] <Unconscious> %s is unconscious.\n", event.Round, event.Actor)
	}
}

type RollHandler struct{}

func (h *RollHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETRollEvent {
		fmt.Printf("[Round %d] <Roll> %s rolls %d, rolls: %v.\n", event.Round, event.Actor, event.Value, event.Rolls)
	}
}

type HPRollHandler struct{}

func (h *HPRollHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETHPRollEvent {
		fmt.Printf("[Round %d] HP <Roll> %s rolls %d, rolls: %v, amount to add: %d\n", event.Round, event.Actor, event.Value, event.Rolls, event.Added)
	}
}
