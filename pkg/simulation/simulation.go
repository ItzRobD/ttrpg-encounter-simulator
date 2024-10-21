package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/events"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/rolling"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"fmt"
	"math/rand/v2"
	"sort"
)

type Prioritization int

const (
	NoPriority Prioritization = iota
	PrioritizeLowestHealth
	PrioritizeMostDamaged
	PrioritizeSpellcasting
	PrioritizeHealer
	PrioritizeHighestCR
	PrioritizeLowestCR
	PrioritizeHighestMaxHP
	PrioritizeLowestMaxHP
)

type Options struct {
	UseMonsterHPAverage     bool
	CanMonstersCrit         bool
	CanPlayersCrit          bool
	HasIncreasedCrits       bool
	AllowPlayerHeals        bool
	AllowMonsterHeals       bool
	Prioritization          Prioritization
	AOEHitsAllEnemies       bool
	PlayerHealThresholdPct  int
	MonsterHealThresholdPct int
}

type Encounter struct {
	sim           *Simulation
	Party         []*character.Character
	Monsters      []*monster.Monster
	CombatTracker []shared.Combatant
	Options       Options
	CurrentRound  int
}

type Simulation struct {
	Encounter  Encounter
	simLog     []events.CombatEvent
	dispatcher *EventDispatcher
}

func New(options Options) Simulation {
	dispatcher := NewEventDispatcher()
	dispatcher.RegisterListener(&AttackHandler{})
	dispatcher.RegisterListener(&HealHandler{})
	dispatcher.RegisterListener(&DeathHandler{})
	dispatcher.RegisterListener(&DamageHandler{})
	dispatcher.RegisterListener(&UnconsciousHandler{})
	dispatcher.RegisterListener(&RollHandler{})
	dispatcher.RegisterListener(&HPRollHandler{})

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

func (e *Encounter) PrintCombatTracker() {
	fmt.Println("Combat Tracker")
	for _, c := range e.CombatTracker {
		fmt.Printf("Initiative: %d, Name: %s\n", c.InitiativeScore, c.Creature.GetName())
	}
}

func (e *Encounter) AddCombatant(c shared.Combatant) error {
	if c.InitiativeScore <= 0 {
		return fmt.Errorf("initiative score must be greater than zero")
	}
	if c.Creature == nil {
		return fmt.Errorf("creature cannot be nil")
	}
	e.CombatTracker = append(e.CombatTracker, c)
	return nil
}

func (e *Encounter) AddPartyMember(c *character.Character) {
	e.Party = append(e.Party, c)
}

func (e *Encounter) AddMonster(m *monster.Monster) error {
	m.EventListener = func(event events.CombatEvent) {
		e.sim.LogEvent(event)
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

func (e *Encounter) SetupCombatTracker() error {
	if e.CombatTracker == nil {
		e.CombatTracker = []shared.Combatant{}
	}

	for _, m := range e.Monsters {
		initiative, err := rolling.InitiativeRoll(m.AbilityScores.Dexterity)
		if err != nil {
			return err
		}
		err = e.AddCombatant(shared.Combatant{
			InitiativeScore: initiative,
			Creature:        m,
		})
		if err != nil {
			return err
		}
	}
	for _, p := range e.Party {
		initiative, err := rolling.InitiativeRoll(p.AbilityScores.Dexterity)
		if err != nil {
			return err
		}
		err = e.AddCombatant(shared.Combatant{
			InitiativeScore: initiative,
			Creature:        p,
		})
		if err != nil {
			return err
		}
	}

	sort.Slice(e.CombatTracker, func(i, j int) bool {
		return e.CombatTracker[i].InitiativeScore > e.CombatTracker[j].InitiativeScore
	})

	return nil
}

func (s *Simulation) Simulate() error {
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

	for s.Encounter.CurrentRound <= 5 {
		s.Encounter.SimulateRound()
		s.Encounter.CurrentRound++
	}
	s.Encounter.PrintCombatTracker()

	return nil
}

func (e *Encounter) SimulateRound() {
	for _, entity := range e.CombatTracker {
		switch c := entity.Creature.(type) {
		case *character.Character:
			if c.IsUnconscious() {
				// Unconscious logic if needed
			}
			target, _ := e.ChooseTarget(entity.Creature)
			if m, ok := target.(*monster.Monster); ok {
				_, err := c.MakeAttack(m, "primary")
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Printf("Target is not a monster\n")
			}

		case *monster.Monster:
			if c.IsUnconscious() {
				// Unconscious logic if needed
			}
		default:
			fmt.Printf("Unknown creature type %T\n", c)
		}
	}
}

func (e *Encounter) ChooseTarget(actor shared.Entity) (shared.Entity, error) {
	var target shared.Entity
	switch actor.(type) {
	case *character.Character:
		// TODO: Add logic to determine targeting prioritization
		// Prioritize Spellcasting Monster
		// Prioritize Highest CR
		// Prioritize Lowest CR
		// Prioritize damaged
		// prioritize highest vs lowest max hp
		switch e.Options.Prioritization {
		case NoPriority:
			fmt.Println("No Prioritization")
			var targets []shared.Entity
			for _, entity := range e.CombatTracker {
				switch m := entity.Creature.(type) {
				case *monster.Monster:
					targets = append(targets, m)
				default:
					continue
				}
			}
			if len(targets) != 0 {
				t := rand.IntN(len(targets))
				target = targets[t]
			} else {
				return nil, fmt.Errorf("no targets available")
			}
		case PrioritizeMostDamaged:
			for _, entity := range e.CombatTracker {
				switch m := entity.Creature.(type) {
				case *monster.Monster:
					if target == nil {
						target = m
						continue
					}
					if m.GetMaxHP()-m.GetCurrentHP() > target.GetMaxHP()-target.GetCurrentHP() {
						target = m
					}
				}
			}
		case PrioritizeLowestHealth:
			for _, entity := range e.CombatTracker {
				switch m := entity.Creature.(type) {
				case *monster.Monster:
					if target == nil {
						target = m
						continue
					}
					fmt.Printf("Target: %d, Current: %d\n", target.GetCurrentHP(), m.GetCurrentHP())
					if m.GetCurrentHP() < target.GetCurrentHP() {
						target = m
					}
				default:
					continue
				}
			}
		default:
			panic("unhandled default case")
		}
	case *monster.Monster:
	default:
		fmt.Printf("Unknown creature type %T\n", actor)
	}

	return target, nil
}
