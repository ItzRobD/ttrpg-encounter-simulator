package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/events"
	"fmt"
)

type AttackHandler struct{}

func (a *AttackHandler) HandleEvent(event events.CombatEvent) {
	if event.EventType == events.AttackEvent {
		fmt.Printf("[Round %d] <Attack> %s attacks %s with %s. %d to hit. Hit: %t\n",
			event.Round, event.Actor, event.Target, event.Attack, event.Value, event.Hit)
	}
}

type SpellAttack struct{}

func (a *SpellAttack) HandleEvent(event events.CombatEvent) {
	if event.EventType == events.SpellAttack {
		fmt.Printf("[Round %d] <Spell Attack> %s attacks %s with %s. %d to hit. Hit: %t\n",
			event.Round, event.Actor, event.Target, event.Attack, event.Value, event.Hit)

	}
}

type SpellDC struct{}

func (a *SpellDC) HandleEvent(event events.CombatEvent) {
	if event.EventType == events.SpellAttack {
		fmt.Printf("[Round %d] <Spell DC Attack> %s attacks %s with %s. DC is: %d. Saving throw: %d. Hit: %t\n",
			event.Round, event.Actor, event.Target, event.Attack, event.Value, event.SavingThrow, event.Hit)

	}
}

type DamageHandler struct{}

func (d *DamageHandler) HandleEvent(event events.CombatEvent) {
	if event.EventType == events.DamageEvent {
		fmt.Printf("[Round %d] <Damage> %s Does %d damage to %s.\n",
			event.Round, event.Actor, event.Value, event.Target)
	}
}

type HealHandler struct{}

func (h *HealHandler) HandleEvent(event events.CombatEvent) {
	if event.EventType == events.HealEvent {
		fmt.Printf("[Round %d] <Heal> %s heals %s for %d health.\n",
			event.Round, event.Actor, event.Target, event.Value)
	}
}

type DeathHandler struct{}

func (d *DeathHandler) HandleEvent(event events.CombatEvent) {
	if event.EventType == events.DeathEvent {
		fmt.Printf("[Round %d] <Death> %s has died.\n", event.Round, event.Target)
	}
}

type UnconsciousHandler struct{}

func (u *UnconsciousHandler) HandleEvent(event events.CombatEvent) {
	if event.EventType == events.UnconsciousEvent {
		fmt.Printf("[Round %d] <Unconscious> %s is unconscious.\n", event.Round, event.Actor)
	}
}

type RollHandler struct{}

func (r *RollHandler) HandleEvent(event events.CombatEvent) {
	if event.EventType == events.RollEvent {
		fmt.Printf("[Round %d] <Roll> %s rolls %d, rolls: %v.\n", event.Round, event.Actor, event.Value, event.Rolls)
	}
}

type HPRollHandler struct{}

func (hp *HPRollHandler) HandleEvent(event events.CombatEvent) {
	if event.EventType == events.HPRollEvent {
		fmt.Printf("[Round %d] HP <Roll> %s rolls %d, rolls: %v, amount to add: %d\n", event.Round, event.Actor, event.Value, event.Rolls, event.Added)
	}
}
