package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/monster"
)

func (e *Encounter) AddMonster(m *monster.Monster) error {
	m.EventListener = func(event interface{}) {
		if evt, ok := event.(events.CombatEvent); ok {
			e.sim.LogEvent(evt)
		}
	}
	if e.Options.UseMonsterHPAverage {
		hp, _, err := m.DetermineMonsterHP(true)
		if err != nil {
			return err
		}
		m.HP.HP = hp
	} else {
		hp, _, err := m.DetermineMonsterHP(false)
		if err != nil {
			return err
		}
		m.HP.HP = hp
	}
	e.Monsters = append(e.Monsters, m)
	return nil
}

func (e *Encounter) handleMonsterTurn(monster *monster.Monster) {
	// TODO: Implement monster turn logic
}
