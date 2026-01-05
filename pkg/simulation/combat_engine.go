package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"fmt"
	"slices"
	"sort"
)

type CombatEngine struct {
	CurrentRound     int
	CurrentTurnIndex int
	Combatants       map[int]*core.Combatant
	TurnOrder        []int
	CombatContext    *core.CombatContext
	EventContext     *core.EventContext
	SimOptions       *core.SimulationOptions
	tieBreakRolls    map[int]int
}

func NewCombatEngine(simOptions *core.SimulationOptions) *CombatEngine {
	return &CombatEngine{
		CurrentRound:     0,
		CurrentTurnIndex: 0,
		Combatants:       make(map[int]*core.Combatant),
		TurnOrder:        nil,
		CombatContext:    nil,
		EventContext:     nil,
		SimOptions:       simOptions,
	}
}

func (ce *CombatEngine) AddCombatant(c *core.Combatant) {
	if ce.Combatants == nil {
		ce.Combatants = make(map[int]*core.Combatant)
	}

	id := len(ce.Combatants)
	c.GetEntity().SetInstanceID(id)
	ce.Combatants[id] = c
}

func (ce *CombatEngine) getSortedCombatantIDs() []int {
	ids := make([]int, 0, len(ce.Combatants))
	for id := range ce.Combatants {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}

// SetupCombat initializes the combat by resetting the current round, rolling initiatives, and updating combatants and the tracker.
// Returns an error if combatants are missing or if an issue occurs during initiative rolling or tracker setup.
func (ce *CombatEngine) SetupCombat() error {
	ce.CurrentRound = 0

	if len(ce.Combatants) <= 0 {
		return fmt.Errorf("combatants list is empty")
	}

	// Pre-allocate tie-break storage
	if ce.tieBreakRolls == nil {
		ce.tieBreakRolls = make(map[int]int)
	} else {
		for k := range ce.tieBreakRolls {
			delete(ce.tieBreakRolls, k)
		}
	}

	// Sort combatant IDs to ensure deterministic roll order
	ids := ce.getSortedCombatantIDs()

	for _, id := range ids {
		c := ce.Combatants[id]
		entity := c.GetEntity()

		initiative, err := entity.RollInitiative()
		if err != nil {
			return err
		}

		ce.Combatants[id].Initiative = initiative

		// Precompute tie-break roll: 1d20; lair auto-loses ties (store 0)
		if c.IsLair {
			ce.tieBreakRolls[id] = 0
		} else if rng := entity.GetRNG(); rng != nil {
			ce.tieBreakRolls[id] = rng.IntN(20) + 1
		} else {
			ce.tieBreakRolls[id] = 10 // neutral default
		}
	}

	return ce.setupCombatTracker()
}

// setupCombatTracker initializes and sorts the combat tracker based on initiative, dexterity, and id order of combatants.
func (ce *CombatEngine) setupCombatTracker() error {
	ce.TurnOrder = ce.getSortedCombatantIDs()

	sort.Slice(ce.TurnOrder, func(i, j int) bool {
		idxI := ce.TurnOrder[i]
		idxJ := ce.TurnOrder[j]

		initI := ce.Combatants[idxI].GetInitiative()
		initJ := ce.Combatants[idxJ].GetInitiative()

		if initI != initJ {
			return initI > initJ
		}

		// Lair loses ties at equal initiative
		if ce.Combatants[idxI].IsLair != ce.Combatants[idxJ].IsLair {
			// Non-lair wins tie
			return !ce.Combatants[idxI].IsLair
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

		// If dexterity is the same, roll a tie-breaker (precomputed d20)
		tbI := ce.tieBreakRolls[idxI]
		tbJ := ce.tieBreakRolls[idxJ]
		if tbI != tbJ {
			return tbI > tbJ
		}

		// Final tie-breaker: id
		return idxI < idxJ
	})

	return nil
}

func (ce *CombatEngine) rollInitiativeForAllCombatants() error {
	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		c := ce.Combatants[id]
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
func (ce *CombatEngine) RunCombat(maxRounds int) (core.VictoryStatus, error) {
	if len(ce.Combatants) == 0 {
		return core.VictoryStatusNone, fmt.Errorf("no combatants added to combat engine")
	}

	if len(ce.TurnOrder) == 0 {
		return core.VictoryStatusNone, fmt.Errorf("no combatants in tracker")
	}

	ce.initializeCombatContext()
	ce.initializeEventContext()
	// Actions should occur starting at Round 1.
	// Round 0 is reserved for setup logs like initiative and HP initialization.
	for round := 1; round <= maxRounds; round++ {
		ce.CurrentRound = round
		victory, err := ce.SimulateRound()
		if err != nil {
			return core.VictoryStatusNone, err
		}

		if victory != core.VictoryStatusNone {
			return victory, nil
		}
	}

	return core.VictoryStatusNone, nil
}

// SimulateRound executes a single round of combat, processing AI decisions and actions for all combatants in the tracker.
// It also manages legendary actions, ensuring legendary creatures can act within the constraints of the combat rules.
// Returns an error if any part of the simulation encounters a failure.
func (ce *CombatEngine) SimulateRound() (core.VictoryStatus, error) {
	err := ce.roundStartEvents()
	if err != nil {
		return core.VictoryStatusNone, fmt.Errorf("failed to execute round start events: %v", err)
	}

	for _, combatantID := range ce.TurnOrder {
		// Lair takes a special turn at initiative 20 when enabled
		if ce.Combatants[combatantID].IsLair {
			if ce.SimOptions != nil && ce.SimOptions.AllowLairActions {
				// Update context for lair and execute its turn like any other
				ce.updateCombatContext(combatantID)
				// Ensure the lair entity receives the fresh combat context before acting
				_ = ce.Combatants[combatantID].GetEntity().UpdateAICombatContext(ce.CombatContext)
				// Ensure the lair entity receives the fresh event context before acting
				ce.Combatants[combatantID].GetEntity().PushEventContext(ce.EventContext)

				status, aiReq, combatErr := ce.Combatants[combatantID].GetEntity().ProcessTurn(combatantID, core.TurnTypeNormal)
				if combatErr != nil {
					return core.VictoryStatusNone, combatErr
				}
				if ce.shouldSkipCombatantTurn(status) {
					// Even when skipping, check if victory has already occurred (e.g., everyone else is down)
					if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
						return v, nil
					}
					continue
				}
				if err2 := ce.ProcessAIRequest(aiReq); err2 != nil {
					return core.VictoryStatusNone, err2
				}
				// Immediately check victory after lair action
				if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
					return v, nil
				}
			}
			continue
		}

		combatError := ce.turnStartEvents(combatantID)
		if combatError != nil {
			return core.VictoryStatusNone, fmt.Errorf("failed to execute turn start events for combatant %d: %v", combatantID, combatError)
		}
		ce.updateCombatContext(combatantID)
		// Start-of-turn victory check (e.g., all enemies fell earlier in round)
		if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
			return v, nil
		}

		for {
			// Event ctx lifecycle - executeTurn updates parent ID for the action id
			status, aiReq, turnError := ce.executeTurn(combatantID)
			if turnError != nil {
				return core.VictoryStatusNone, turnError
			}

			// Record death saves if any were made
			ce.recordDeathSaves(combatantID, status)

			// Handle non-acting statuses or nil action request
			if ce.shouldSkipCombatantTurn(status) || aiReq == nil {

				// Even if no action taken, check victory
				if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
					return v, nil
				}
				break
			}

			// Execute the actions in the turn request
			turnError = ce.ProcessAIRequest(aiReq)
			if turnError != nil {
				return core.VictoryStatusNone, turnError
			}

			// Immediately check victory after each action
			if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
				return v, nil
			}

			// Re-update context after each action to reflect changes (e.g. someone might now need healing)
			ce.updateCombatContext(combatantID)
		}

		combatError = ce.turnEndEvents(combatantID)
		if combatError != nil {
			return core.VictoryStatusNone, fmt.Errorf("failed to execute turn end events for combatant %d: %v", combatantID, combatError)
		}

		// Update state and statistics at end of turn
		if c, exists := ce.Combatants[combatantID]; exists && !c.IsLair {
			c.Info.UpdateState()
			c.Info.Statistics.TurnsSinceLastHeal++
		}
	}

	err = ce.roundEndEvents()
	if err != nil {
		return core.VictoryStatusNone, fmt.Errorf("failed to execute round end events: %v", err)
	}

	// Check victory conditions at the end of each round
	victory := ce.checkVictoryCondition()
	if victory != core.VictoryStatusNone {
		return victory, nil
	}

	return core.VictoryStatusNone, nil
}

func (ce *CombatEngine) PrintCombatTracker() {
	order := 0
	for _, index := range ce.TurnOrder {
		order++
		combatant := ce.Combatants[index]
		if combatant.IsLair {
			fmt.Printf("Order Index: %d - initiative: %d - Name: Lair\n", order, combatant.GetInitiative())
		} else {
			fmt.Printf("Order Index: %d - initiative: %d - Name: %s\n", order, combatant.GetInitiative(), combatant.GetEntity().GetName())
		}
	}
}

// Debug function
func (ce *CombatEngine) PrintCombatants() {
	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		c := ce.Combatants[id]
		if c.IsLair {
			fmt.Println("Name: Lair")
			continue
		}
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

func (ce *CombatEngine) refreshLegendaryActions(combatantID int) {
	if len(ce.CombatContext.LegendaryCreatures) == 0 {
		return
	}

	if ce.Combatants[combatantID].GetEntity().IsMonster() && ce.Combatants[combatantID].GetEntity().GetIsLegendary() {
		ce.Combatants[combatantID].GetEntity().RefreshLegendaryActions()
	}
}

// roundStartEvents triggers necessary actions or states at the start of a combat round, including updates and ability rolls.
func (ce *CombatEngine) roundStartEvents() error {
	err := ce.rollRechargeAbilities()
	if err != nil {
		return err
	}
	// TODO: Condition duration tracking
	return nil
}

func (ce *CombatEngine) rollRechargeAbilities() error {
	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		c := ce.Combatants[id]
		if c.IsLair {
			// Let lair roll recharge for its actions (if any)
			// Type assert to known lair type without importing lair here; use interface probing
			if lr, ok := c.GetEntity().(interface{ GetActionManager() interface{} }); ok {
				// Best effort: reflect expected type methods
				if lam, ok2 := lr.GetActionManager().(interface{ RollRechargeActions() }); ok2 {
					lam.RollRechargeActions()
				}
			}
			continue
		}

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
	ce.EventContext.GenerateSequenceID()
	combatant := ce.Combatants[combatantID]
	entity := combatant.GetEntity()

	entity.Regenerate()

	// Track that the entity has taken a turn in this combat
	if entity.IsCharacter() {
		c := entity.(*character.Character)
		c.EntityStateManager.SetHasTakenTurnInCombat(true)
	} else if entity.IsMonster() {
		m := entity.(*monster.Monster)
		m.EntityStateManager.SetHasTakenTurnInCombat(true)
	}

	// General Turn Start Events (Refreshing actions for both characters and monsters)
	if !entity.IsDead() && !entity.IsUnconscious() {
		if c, ok := entity.(*character.Character); ok {
			c.EntityStateManager.RefreshActions()
		} else if m, ok := entity.(*monster.Monster); ok {
			m.EntityStateManager.RefreshActions()
		}
	}

	return nil
}

func (ce *CombatEngine) executeTurn(combatantID int) (*core.TurnResult, *core.AIRequest, error) {
	combatant := ce.Combatants[combatantID]
	entity := combatant.GetEntity()

	// Update Combatant's AI ctx
	err := entity.UpdateAICombatContext(ce.CombatContext)
	if err != nil {
		return nil, nil, err
	}

	// Update Combatant's Event ctx
	ce.EventContext.GenerateParentID() // Create a new parent ID to act as the action ID
	entity.PushEventContext(ce.EventContext)

	// Tell entity to process its turn
	// This will determine if it can act. If not, log; if so create ai request
	// Pass the ai request back to CE for processing
	// ce should understand why an ai request is empty
	// return TurnStatus, AIRequest, Error?

	turnResult, aiReq, err := entity.ProcessTurn(combatantID, core.TurnTypeNormal)
	if err != nil {
		return nil, nil, err
	}

	return turnResult, aiReq, nil
}

func (ce *CombatEngine) shouldSkipCombatantTurn(status *core.TurnResult) bool {
	if status.TurnStatuses[core.TurnActionReady] ||
		status.TurnStatuses[core.TurnRevived] ||
		status.TurnStatuses[core.TurnLegendaryReady] {
		return false
	}
	return true
}

func (ce *CombatEngine) turnEndEvents(currentCombatantID int) error {
	// TODO: End of turn events:
	//		Condition duration
	//		Saving throws for conditions
	//		Temp HP Expiration?

	err := ce.executeLegendaryAction(currentCombatantID)
	if err != nil {
		return fmt.Errorf("error processing turn end events: %v", err)
	}

	return nil
}

// TODO: Add an option to execute only one legendary action per turn to not overwhelm
func (ce *CombatEngine) executeLegendaryAction(actingCombatantID int) error {
	if len(ce.CombatContext.LegendaryCreatures) == 0 {
		return nil
	}
	// Build sorted list of legendary creatures (by initiative, excluding current actor)
	legCreatureIDs := make([]int, 0)
	for id := range ce.CombatContext.LegendaryCreatures {
		if id != actingCombatantID {
			legCreatureIDs = append(legCreatureIDs, id)
		}
	}

	// Sort by initiative (descending), then by dex, then by id
	slices.SortFunc(legCreatureIDs, func(a, b int) int {
		initA := ce.Combatants[a].GetInitiative()
		initB := ce.Combatants[b].GetInitiative()

		// Sort by initiative (descending)
		if initA != initB {
			return initB - initA
		}

		// Tie-breaker: dexterity
		entityA := ce.Combatants[a].GetEntity()
		entityB := ce.Combatants[b].GetEntity()

		dexA, _ := entityA.GetAbilityScoreModifier(core.AbilityDexterity)
		dexB, _ := entityB.GetAbilityScoreModifier(core.AbilityDexterity)

		if dexA != dexB {
			return dexB - dexA
		}

		// Final tie-breaker: id
		return a - b
	})

	// Iterate through sorted legendary creatures
	for _, legID := range legCreatureIDs {
		entity := ce.Combatants[legID].GetEntity()

		// Request legendary action - entity checks if it has points available
		status, legAIReq, err := entity.ProcessTurn(legID, core.TurnTypeLegendary)
		if err != nil {
			return fmt.Errorf("failed to get legendary AI request for combatant %d: %v", legID, err)
		}

		// TODO: Update to use statuses - we did with character but not monster yet
		// Check if creature is out of legendary action points
		// Assuming your AIRequest or a new field indicates this
		if status.TurnStatuses[core.TurnLegendaryUnavailable] {
			// Out of points, try the next entity in the list
			continue
		}
		if status.TurnStatuses[core.TurnLegendaryReady] {

			// Process the legendary action
			err = ce.ProcessAIRequest(legAIReq)
			if err != nil {
				return fmt.Errorf("failed to process legendary action for combatant %d: %v", legID, err)
			}

			// Victory may occur due to this legendary action; short-circuit further legendary actions
			if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
				return nil
			}

			// If LimitedLegendaryActions is true, only one creature acts per turn
			if ce.SimOptions.LimitedLegendaryActions {
				break
			}
		}
	}

	return nil
}

func (ce *CombatEngine) roundEndEvents() error {
	// TODO: End of round events:
	//		Condition duration
	//		Saving throws for conditions
	//		Temp HP Expiration?
	//		Ongoing damage?
	// 		Regeneration?
	return nil
}
