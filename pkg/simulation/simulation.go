package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/simulation/events"
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
	dispatcher.RegisterListener(&events.AttackHandler{})
	dispatcher.RegisterListener(&events.SpellAttackHandler{})
	dispatcher.RegisterListener(&events.SpellDCHandler{})
	dispatcher.RegisterListener(&events.HealHandler{})
	dispatcher.RegisterListener(&events.DeathHandler{})
	dispatcher.RegisterListener(&events.DamageHandler{})
	dispatcher.RegisterListener(&events.UnconsciousHandler{})
	dispatcher.RegisterListener(&events.RollHandler{})
	dispatcher.RegisterListener(&events.HPRollHandler{})
	dispatcher.RegisterListener(&events.ActionChoiceHandler{})
	dispatcher.RegisterListener(&events.SpellChoiceHandler{})
	dispatcher.RegisterListener(&events.HPModifiedHandler{})

	var s Simulation
	s.Encounter.Options = options
	s.dispatcher = dispatcher
	s.Encounter.sim = &s
	return s
}

func (s *Simulation) LogEvent(e events.CombatEvent) {
	e.Round = s.Encounter.CurrentRound
	s.simLog = append(s.simLog, e)
	s.dispatcher.DispatchEvent(e)
}

func (s *Simulation) PrintSimulationLog() {
	fmt.Println("Simulation Log")
	for _, e := range s.simLog {
		fmt.Printf("%+v\n", e)
	}
}

func (s *Simulation) Simulate(maxximumRounds int) error {
	s.Encounter.CurrentRound = 1
	err := s.Encounter.SetupCombatTracker()
	if err != nil {
		return err
	}

	for _, c := range s.Encounter.Party {
		c.EventListener = func(event events.CombatEvent) {
			s.LogEvent(event)
		}
	}
	//for _, m := range s.Encounter.Monsters {
	//	m.EventListener = func(event events.CombatEvent) {
	//		s.LogEvent(event)
	//	}
	//}

	for s.Encounter.CurrentRound <= maxximumRounds {
		s.Encounter.SimulateRound()
		s.Encounter.CurrentRound++
	}
	s.Encounter.PrintCombatTracker()

	return nil
}
