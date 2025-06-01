package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
)

type Options struct {
	UseMonsterHPAverage     bool
	CanMonstersCrit         bool
	CanPlayersCrit          bool
	HasIncreasedCrits       bool
	AllowPlayerHeals        bool
	AllowMonsterHeals       bool
	TargetPriority          shared.Prioritization
	ActionPreference        shared.ActionPreference
	AOEHitsAllEnemies       bool
	PlayerHealThresholdPct  int
	MonsterHealThresholdPct int
}

type Simulation struct {
	Encounter  Encounter
	simLog     []events.CombatEvent
	dispatcher *events.EventDispatcher
}

func New(options Options) Simulation {
	dispatcher := events.NewEventDispatcher()
	dispatcher.RegisterHandler(&events.AttackHandler{})
	dispatcher.RegisterHandler(&events.SpellAttackHandler{})
	dispatcher.RegisterHandler(&events.SpellDCHandler{})
	dispatcher.RegisterHandler(&events.HealHandler{})
	dispatcher.RegisterHandler(&events.DeathHandler{})
	dispatcher.RegisterHandler(&events.DamageHandler{})
	dispatcher.RegisterHandler(&events.UnconsciousHandler{})
	dispatcher.RegisterHandler(&events.RollHandler{})
	dispatcher.RegisterHandler(&events.HPRollHandler{})
	dispatcher.RegisterHandler(&events.ActionChoiceHandler{})
	dispatcher.RegisterHandler(&events.SpellChoiceHandler{})
	dispatcher.RegisterHandler(&events.HPModifiedHandler{})

	var s Simulation
	s.Encounter.Options = options
	s.dispatcher = dispatcher
	s.Encounter.sim = &s
	return s
}

func (s *Simulation) LogEvent(e events.CombatEvent) {
	e.SetRound(s.Encounter.CurrentRound)
	s.simLog = append(s.simLog, e)
	s.dispatcher.DispatchEvent(e)
}

func (s *Simulation) PrintSimulationLog() {
	fmt.Println("Simulation Log")
	for _, e := range s.simLog {
		fmt.Printf("%+v\n", e)
	}
}

func (s *Simulation) Simulate(maximumRounds int) error {
	s.Encounter.CurrentRound = 1
	err := s.Encounter.SetupCombatTracker()
	if err != nil {
		return err
	}
	s.Encounter.PrintCombatTracker()
	//s.Encounter.PrintEncounterMembers()

	for _, c := range s.Encounter.Party {
		c.EventListener = func(event interface{}) {
			if evt, ok := event.(events.CombatEvent); ok {
				s.LogEvent(evt)
			}
		}
	}
	//for _, m := range s.Encounter.Monsters {
	//	m.EventListener = func(event events.CombatEvent) {
	//		s.LogEvent(event)
	//	}
	//}

	for s.Encounter.CurrentRound <= maximumRounds {
		s.Encounter.SimulateRound()
		s.Encounter.CurrentRound++
	}

	return nil
}
