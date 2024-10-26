package events

import (
	"fmt"
)

type AttackHandler struct{}

func (a *AttackHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETAttackEvent {
		fmt.Printf("[Round %d] <Attack> %s attacks %s with %s. %d to hit. Hit: %t\n",
			event.Round, event.Actor, event.Target, event.Attack, event.Value, event.Hit)
	}
}

type SpellAttack struct{}

func (a *SpellAttack) HandleEvent(event CombatEvent) {
	if event.EventType == ETSpellAttack {
		fmt.Printf("[Round %d] <Spell Attack> %s attacks %s with %s. %d to hit. Hit: %t\n",
			event.Round, event.Actor, event.Target, event.Attack, event.Value, event.Hit)

	}
}

type SpellDC struct{}

func (a *SpellDC) HandleEvent(event CombatEvent) {
	if event.EventType == ETSpellAttack {
		fmt.Printf("[Round %d] <Spell DC Attack> %s attacks %s with %s. DC is: %d. Saving throw: %d. Hit: %t\n",
			event.Round, event.Actor, event.Target, event.Attack, event.Value, event.SavingThrow, event.Hit)

	}
}

type DamageHandler struct{}

func (d *DamageHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETDamageEvent {
		fmt.Printf("[Round %d] <Damage> %s Does %d damage to %s.\n",
			event.Round, event.Actor, event.Value, event.Target)
	}
}

type HealHandler struct{}

func (h *HealHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETHealEvent {
		fmt.Printf("[Round %d] <Heal> %s heals %s for %d health.\n",
			event.Round, event.Actor, event.Target, event.Value)
	}
}

type DeathHandler struct{}

func (d *DeathHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETDeathEvent {
		fmt.Printf("[Round %d] <Death> %s has died.\n", event.Round, event.Target)
	}
}

type UnconsciousHandler struct{}

func (u *UnconsciousHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETUnconsciousEvent {
		fmt.Printf("[Round %d] <Unconscious> %s is unconscious.\n", event.Round, event.Actor)
	}
}

type RollHandler struct{}

func (r *RollHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETRollEvent {
		fmt.Printf("[Round %d] <Roll> %s rolls %d, rolls: %v.\n", event.Round, event.Actor, event.Value, event.Rolls)
	}
}

type HPRollHandler struct{}

func (hp *HPRollHandler) HandleEvent(event CombatEvent) {
	if event.EventType == ETHPRollEvent {
		fmt.Printf("[Round %d] HP <Roll> %s rolls %d, rolls: %v, amount to add: %d\n", event.Round, event.Actor, event.Value, event.Rolls, event.Added)
	}
}
