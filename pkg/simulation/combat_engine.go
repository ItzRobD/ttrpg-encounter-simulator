package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"fmt"
	"math"
	"sort"
)

type CombatEngine struct {
	CurrentRound  int
	Combatants    map[int]*core.Combatant
	CombatTracker []int
	CombatContext *core.CombatContext
	SimOptions    *core.SimulationOptions
}

func NewCombatEngine(simOptions *core.SimulationOptions) *CombatEngine {
	return &CombatEngine{
		CurrentRound:  0,
		Combatants:    make(map[int]*core.Combatant),
		CombatTracker: nil,
		CombatContext: nil,
		SimOptions:    simOptions,
	}
}

func (ce *CombatEngine) ProcessAIRequest(req *core.AIRequest) error {
	ce.attachOptionsToAIRequest(req)
	switch req.ActionType {
	case core.ATMelee, core.ATRanged:
		return ce.executeWeaponAttack(req)
	case core.ATSpell:
		return ce.executeSpellCast(req)
	//case core.ATHeal:
	//	return ce.executeHeal(req)
	//case core.ATUnarmed:
	//	return ce.executeUnarmedAttack(req)
	case core.ATMonsterAction:
		return ce.executeMonsterAction(req)
	case core.ATMonsterMultiattack:
		return ce.executeMonsterMultiattack(req)
	case core.ATLegendaryAction:
		return ce.executeMonsterLegendaryAction(req)
	default:
		return fmt.Errorf("unknown action type: %v", req.ActionType)
	}

}

func (ce *CombatEngine) attachOptionsToAIRequest(aiReq *core.AIRequest) {
	aiReq.SimOptions = ce.SimOptions
}

func (ce *CombatEngine) executeWeaponAttack(aiReq *core.AIRequest) error {
	// TODO: Choose weapons slot | versatile
	aiReq.WeaponSlot = core.WSPrimary
	results, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	return ce.processActionResults(results)
}

func (ce *CombatEngine) executeSpellCast(aiReq *core.AIRequest) error {
	results, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}
	return ce.processActionResults(results)
}

func (ce *CombatEngine) executeMonsterAction(aiReq *core.AIRequest) error {
	results, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	return ce.processActionResults(results)
}

func (ce *CombatEngine) executeMonsterMultiattack(aiReq *core.AIRequest) error {
	results, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	return ce.processActionResults(results)
}

func (ce *CombatEngine) executeMonsterLegendaryAction(aiReq *core.AIRequest) error {
	results, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	return ce.processActionResults(results)
}

func (ce *CombatEngine) processActionResults(outcome *core.ActionOutcome) error {
	target, exists := ce.CombatContext.AllCombatants[outcome.TargetID]
	if !exists {
		return fmt.Errorf("target entity not found in combat context")
	}

	if len(outcome.Effects) > 0 {
		var hpModResult core.HPModificationResult
		var err error
		for _, effect := range outcome.Effects {
			switch effect.Type {
			case core.EffectDamage:
				v := -effect.Value
				hpModResult, err = target.GetEntity().ModifyHP(v, false, false)
				if err != nil {
					return fmt.Errorf("failed to modify target entity HP: %v", err)
				}
			case core.EffectHealing:
				v := math.Abs(float64(effect.Value))
				hpModResult, err = target.GetEntity().ModifyHP(int(v), false, false)
				if err != nil {
					return fmt.Errorf("failed to modify target entity HP: %v", err)
				}
			case core.EffectTempHP:
				v := math.Abs(float64(effect.Value))
				hpModResult, err = target.GetEntity().ModifyHP(int(v), true, false)
				if err != nil {
					return fmt.Errorf("failed to modify target entity HP: %v", err)
				}
			case core.EffectCondition:
				return fmt.Errorf("effects of type %v are not supported", core.EffectCondition)
			}
		}

		entity := ce.CombatContext.AllCombatants[outcome.ActorID].GetEntity()

		events.LogHPModifiedEvent(entity, target.GetEntity(), hpModResult, entity.GetEventListener())

		ce.Combatants[outcome.TargetID] = target
	}

	return nil
}

func (ce *CombatEngine) AddCombatant(c *core.Combatant) {
	if ce.Combatants == nil {
		ce.Combatants = make(map[int]*core.Combatant)
	}

	ce.Combatants[len(ce.Combatants)] = c
}

// SetupCombat initializes the combat by resetting the current round, rolling initiatives, and updating combatants and the tracker.
// Returns an error if combatants are missing or if an issue occurs during initiative rolling or tracker setup.
func (ce *CombatEngine) SetupCombat() error {
	ce.CurrentRound = 0

	if len(ce.Combatants) <= 0 {
		return fmt.Errorf("combatants list is empty")
	}

	for id, c := range ce.Combatants {
		entity := c.GetEntity()

		initiative, err := entity.RollInitiative()
		if err != nil {
			return err
		}

		updatedCombatant := core.NewCombatant(entity, initiative)
		ce.Combatants[id] = updatedCombatant
	}

	return ce.setupCombatTracker()
}

// setupCombatTracker initializes and sorts the combat tracker based on initiative, dexterity, and ID order of combatants.
func (ce *CombatEngine) setupCombatTracker() error {
	ce.CombatTracker = make([]int, 0, len(ce.Combatants))
	for id := range ce.Combatants {
		ce.CombatTracker = append(ce.CombatTracker, id)
	}

	sort.Slice(ce.CombatTracker, func(i, j int) bool {
		idxI := ce.CombatTracker[i]
		idxJ := ce.CombatTracker[j]

		initI := ce.Combatants[idxI].GetInitiative()
		initJ := ce.Combatants[idxJ].GetInitiative()

		if initI != initJ {
			return initI > initJ
		}

		// If initiative is the same, sort by dexterity modifier
		dexI, err := ce.Combatants[idxI].GetEntity().GetAbilityScoreModifier(core.AbilityDexterity)
		if err != nil {
			return false
		}
		dexJ, err := ce.Combatants[idxJ].GetEntity().GetAbilityScoreModifier(core.AbilityDexterity)
		if err != nil {
			return false
		}

		if dexI != dexJ {
			return dexI > dexJ
		}

		// If dexterity is the same, sort by ID
		return idxI < idxJ
	})

	ce.insertLairCombatant()
	return nil
}

// insertLairCombatant adds a dummy lair combatant with fixed initiative to the combat tracker, maintaining initiative order.
func (ce *CombatEngine) insertLairCombatant() {
	insertIdx := 0
	for i, id := range ce.CombatTracker {
		if ce.Combatants[id].GetInitiative() >= 20 {
			insertIdx = i + 1
		} else {
			break
		}
	}

	lairCombatant := core.Combatant{
		Entity:     nil,
		Initiative: 20,
		CanAct:     true,
		IsLair:     true,
	}
	lairID := -1 // Lair will never be targeted
	ce.Combatants[lairID] = &lairCombatant
	ce.CombatTracker = append(ce.CombatTracker[:insertIdx], append([]int{lairID}, ce.CombatTracker[insertIdx:]...)...)
}

func (ce *CombatEngine) rollInitiativeForAllCombatants() error {
	for id, c := range ce.Combatants {
		init, err := c.GetEntity().RollInitiative()
		if err != nil {
			return err
		}

		updatedCombatant := c
		updatedCombatant.Initiative = init
		ce.Combatants[id] = updatedCombatant
	}
	return nil
}

// RunCombat executes combat rounds for a maximum of maxRounds and handles the progression of the combat state.
// Returns an error if any issue occurs during round simulation.
func (ce *CombatEngine) RunCombat(maxRounds int) error {
	ce.initializeCombatContext()
	for round := ce.CurrentRound; round <= maxRounds; round++ {
		err := ce.SimulateRound()
		if err != nil {
			return err
		}
		ce.CurrentRound++
	}

	return nil
}

// SimulateRound executes a single round of combat, processing AI decisions and actions for all combatants in the tracker.
// It also manages legendary actions, ensuring legendary creatures can act within the constraints of the combat rules.
// Returns an error if any part of the simulation encounters a failure.
func (ce *CombatEngine) SimulateRound() error {
	legIDActedThisRound := make(map[int]bool)
	err := ce.roundStartEvents()
	if err != nil {
		return fmt.Errorf("failed to execute round start events: %v", err)
	}
	for _, combatantID := range ce.CombatTracker {
		ce.updateCombatContext(combatantID)
		combatant := ce.Combatants[combatantID]

		if !combatant.GetCanAct() {
			continue // Skip unconscious combatants
		}

		if combatant.GetEntity().IsMonster() {
			// Handle Monster Specific Actions

		} else if combatant.GetEntity().IsCharacter() {
			// Handle Character Specific Actions

		}

		// Update Combatant's AI Context
		err := combatant.GetEntity().UpdateAICombatContext(ce.CombatContext)
		if err != nil {
			return err
		}

		// Choose entity action
		aiReq, err := combatant.GetEntity().GetAIRequest(combatantID, core.AIReqChooseAction)
		if err != nil {
			return err
		}

		err = ce.ProcessAIRequest(aiReq)
		if err != nil {
			return err
		}

		if len(ce.CombatContext.LegendaryCreatures) > 0 {
			for _, id := range ce.CombatTracker {
				if ce.isLegendaryCreature(id) && !legIDActedThisRound[id] && id != combatantID {
					legAIReq, errL := ce.CombatContext.AllCombatants[id].GetEntity().GetAIRequest(id, core.AIReqLegendaryAction)
					if errL != nil {
						return errL
					}
					errL = ce.ProcessAIRequest(legAIReq)
					if errL != nil {
						return errL
					}

					legIDActedThisRound[id] = true

					if ce.SimOptions.LimitedLegendaryActions {
						break
					}
				}
			}
		}
		// TODO: Multi attack HP modification is only logging one event not for each action
	}
	return nil
}

func (ce *CombatEngine) isLegendaryCreature(id int) bool {
	if ce.CombatContext.LegendaryCreatures == nil {
		return false
	}
	_, exists := ce.CombatContext.LegendaryCreatures[id]
	return exists
}

// Debug function
func (ce *CombatEngine) PrintCombatTracker() {
	order := 0
	for _, index := range ce.CombatTracker {
		order++
		fmt.Printf("Order Index: %d - Initiative: %d - Name: %s\n", order, ce.Combatants[index].GetInitiative(), ce.Combatants[index].GetEntity().GetName())
	}
}

// Debug function
func (ce *CombatEngine) PrintCombatants() {
	for _, c := range ce.Combatants {
		fmt.Printf("Name: %s\n", c.GetEntity().GetName())
	}
}

func (ce *CombatEngine) processAttackResults(attackResults []core.AttackResult) error {
	for _, result := range attackResults {
		if result.GetIsHit() {
			fmt.Println(result)
		}
	}

	return nil
}

// initializeCombatContext initializes the combat context by setting up combatants, rounds, and relevant configuration values.
func (ce *CombatEngine) initializeCombatContext() {
	if ce.CombatContext == nil {
		ce.CombatContext = &core.CombatContext{}
	}
	if ce.CombatContext.AllCombatants == nil {
		ce.CombatContext.AllCombatants = make(map[int]*core.Combatant)
	}
	if ce.CombatContext.LegendaryCreatures == nil {
		ce.CombatContext.LegendaryCreatures = make(map[int]uint8)
	}

	ce.CombatContext.AllCombatants = ce.Combatants
	ce.CombatContext.CurrentRound = ce.CurrentRound

	if ce.SimOptions != nil {
		ce.CombatContext.AllowCharacterHeals = ce.SimOptions.AllowCharacterHeals
		ce.CombatContext.AllowMonsterHeals = ce.SimOptions.AllowMonsterHeals
		ce.CombatContext.AOEHitsAllEnemies = ce.SimOptions.AOEHitsAllEnemies
		ce.CombatContext.CharacterHealThresholdPct = ce.SimOptions.CharacterHealThresholdPct
		ce.CombatContext.MonsterHealThresholdPct = ce.SimOptions.MonsterHealThresholdPct
	}

	for id, combatant := range ce.Combatants {
		entity := combatant.GetEntity()
		if entity.IsMonster() && entity.GetIsLegendary() {
			ce.CombatContext.LegendaryCreatures[id] = 1
		}
	}
}

func (ce *CombatEngine) updateCombatContext(actorID int) {
	ce.CombatContext.AllCombatants = ce.Combatants
	ce.CombatContext.CurrentRound = ce.CurrentRound
	ce.CombatContext.ActingEntityID = actorID

	ce.CombatContext.CharactersInNeedOfHealing, ce.CombatContext.MonstersInNeedOfHealing = ce.calculateEntitiesNeedingHealing()

	for id, combatant := range ce.Combatants {
		entity := combatant.GetEntity()

		// Check if combatant is unconscious/defeated
		if entity.IsUnconscious() {
			// Mark as unable to act but keep in combat for potential revival
			temp := combatant
			temp.SetCanAct(false)
			ce.Combatants[id] = temp
			ce.CombatContext.AllCombatants[id] = temp
		}

	}

}

func (ce *CombatEngine) calculateEntitiesNeedingHealing() ([]int, []int) {
	var charNeedHealing []int
	var monNeedHealing []int

	for id, combatant := range ce.Combatants {
		entity := combatant.GetEntity()

		// Calculate HP percentage
		var threshold int
		if entity.IsCharacter() {
			threshold = ce.CombatContext.CharacterHealThresholdPct
		} else {
			threshold = ce.CombatContext.MonsterHealThresholdPct
		}

		// Entity needs healing if below threshold and not unconscious
		// TODO: Unconscious can get healed to no longer be unconscious
		if entity.GetHPStatus().GetHPPct() <= threshold && !entity.IsUnconscious() {
			if entity.IsCharacter() {
				charNeedHealing = append(charNeedHealing, id)
			} else {
				monNeedHealing = append(monNeedHealing, id)
			}
		}
	}

	return charNeedHealing, monNeedHealing
}

// refreshLegendaryActions resets the legendary action count for all legendary creatures in the combat context.
func (ce *CombatEngine) refreshLegendaryActions() {
	if len(ce.CombatContext.LegendaryCreatures) == 0 {
		return
	}

	for id, _ := range ce.CombatContext.LegendaryCreatures {
		ce.Combatants[id].GetEntity().RefreshLegendaryActions()
	}
}

// roundStartEvents triggers necessary actions or states at the start of a combat round, including updates and ability rolls.
func (ce *CombatEngine) roundStartEvents() error {
	ce.refreshLegendaryActions()
	err := ce.rollRechargeAbilities()
	if err != nil {
		return err
	}
	// TODO: Condition duration tracking
	return nil
}

func (ce *CombatEngine) rollRechargeAbilities() error {
	for _, c := range ce.Combatants {
		if c.GetEntity().IsMonster() {
			m, ok := c.GetEntity().(*monster.Monster)
			if !ok {
				return fmt.Errorf("entity is monster but type assertion failed")
			}

			m.ActionManager.RollRechargeActions()
		}
	}
	return nil
}

func (ce *CombatEngine) turnStartEvents(combatantID int) error {
	combatant := ce.Combatants[combatantID]
	entity := combatant.GetEntity()

	// Character Specific Events
	if entity.IsCharacter() {
		c, ok := entity.(*character.Character)
		if !ok {
			return fmt.Errorf("entity is character but type assertion failed")
		}

		// Death Saves
		if c.EntityState.GetIsUnconscious() && !c.EntityState.IsStable && !c.EntityState.IsDead {
			res, err := c.RollManager.RollDeathSavingThrow()
			if err != nil {
				return fmt.Errorf("failed to roll death saving throw: %v", err)
			}
			err = c.EntityState.ApplyDeathSavingThrowResult(res)
			if err != nil {
				return fmt.Errorf("failed to apply death saving throw result: %v", err)
			}
		}

		if !c.EntityState.IsDead && !c.EntityState.GetIsUnconscious() {
			c.EntityState.RefreshActions()
		}
	}

	return nil
}
