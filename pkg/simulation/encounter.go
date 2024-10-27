package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/simulation/events"
	"fmt"
	"sort"
)

type Encounter struct {
	sim           *Simulation
	Party         []*character.Character
	Monsters      []*monster.Monster
	CombatTracker []shared.Combatant
	Options       Options
	CurrentRound  int
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
		initiative, err := shared.InitiativeRoll(m.AbilityScores.Dexterity)
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
		initiative, err := shared.InitiativeRoll(p.AbilityScores.Dexterity)
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

func (e *Encounter) SimulateRound() {
	for _, entity := range e.CombatTracker {
		switch creature := entity.Creature.(type) {
		case *character.Character:
			if creature.IsUnconscious() {
				continue // Skip if the character is unconscious
			}
			e.handleCharacterTurn(creature)
		case *monster.Monster:
			if creature.IsUnconscious() {
				continue // Skip if the monster is unconscious
			}
			//e.handleMonsterTurn(creature) // TODO: Implement monster logic
		default:
			fmt.Printf("Unknown creature type %T\n", creature)
		}
	}
}

func (e *Encounter) handleCharacterTurn(character *character.Character) {
	actionType, err := e.ChooseCharacterActionType(character)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch actionType {
	case shared.ATHeal:
		e.performCharacterHealAction(character)
		return
	case shared.ATRanged:
		e.performCharacterRangedAttack(character)
	case shared.ATMelee:
		e.performCharacterMeleeAttack(character)
	case shared.ATSpell:
		// Implement spell logic
		return
	case shared.ATNoAction:
		fallthrough
	default:
		// No action to be taken
		return
	}
}

func (e *Encounter) performCharacterRangedAttack(character *character.Character) {
	target, _ := e.chooseDamageTarget(character)
	if monsterTarget, ok := target.(*monster.Monster); ok {
		_, err := character.MakeWeaponAttack(monsterTarget, "ranged")
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("Target is not a monster\n")
	}
}

func (e *Encounter) performCharacterMeleeAttack(character *character.Character) {
	target, _ := e.chooseDamageTarget(character)
	if monsterTarget, ok := target.(*monster.Monster); ok {
		// TODO: Add secondary slot
		_, err := character.MakeWeaponAttack(monsterTarget, "primary")
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("Target is not a monster\n")
	}
}

func (e *Encounter) performCharacterHealAction(c *character.Character) {
	target, _ := e.chooseHealTarget(c)
	healingSpell, err := e.chooseBestHealingSpell(c, target)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = c.MakeSpellHeal(target, healingSpell)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (e *Encounter) handleMonsterTurn(monster *monster.Monster) {
	// TODO: Implement monster turn logic
}
