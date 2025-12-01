package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"fmt"
	"math"
	"slices"
	"sort"
)

type CombatEngine struct {
	CurrentRound     int
	CurrentTurnIndex int
	Combatants       map[int]*core.Combatant
	TurnOrder        []int
	CombatContext    *core.CombatContext
	SimOptions       *core.SimulationOptions
}

func NewCombatEngine(simOptions *core.SimulationOptions) *CombatEngine {
	return &CombatEngine{
		CurrentRound:     0,
		CurrentTurnIndex: 0,
		Combatants:       make(map[int]*core.Combatant),
		TurnOrder:        nil,
		CombatContext:    nil,
		SimOptions:       simOptions,
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
	target, exists := ce.Combatants[outcome.TargetID]
	if !exists {
		return fmt.Errorf("target entity not found in combat")
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

		actor, exists := ce.Combatants[outcome.ActorID]
		if !exists {
			return fmt.Errorf("actor entity not found in combat")
		}
		entity := actor.GetEntity()

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
		// Skip lair combatants
		if c.IsLair {
			continue
		}

		entity := c.GetEntity()

		initiative, err := entity.RollInitiative()
		if err != nil {
			return err
		}

		ce.Combatants[id].Initiative = initiative
	}

	return ce.setupCombatTracker()
}

// setupCombatTracker initializes and sorts the combat tracker based on initiative, dexterity, and ID order of combatants.
func (ce *CombatEngine) setupCombatTracker() error {
	ce.TurnOrder = make([]int, 0, len(ce.Combatants))
	for id := range ce.Combatants {
		ce.TurnOrder = append(ce.TurnOrder, id)
	}

	sort.Slice(ce.TurnOrder, func(i, j int) bool {
		idxI := ce.TurnOrder[i]
		idxJ := ce.TurnOrder[j]

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
	for i, id := range ce.TurnOrder {
		if ce.Combatants[id].GetInitiative() >= 20 {
			insertIdx = i + 1
		} else {
			break
		}
	}

	lairCombatant := core.Combatant{
		Entity:     nil,
		Initiative: 20,
		IsLair:     true,
	}
	lairID := -1 // Lair will never be targeted
	ce.Combatants[lairID] = &lairCombatant
	ce.TurnOrder = append(ce.TurnOrder[:insertIdx], append([]int{lairID}, ce.TurnOrder[insertIdx:]...)...)
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
func (ce *CombatEngine) RunCombat(maxRounds int) (core.VictoryStatus, error) {
	if len(ce.Combatants) == 0 {
		return core.VictoryStatusNone, fmt.Errorf("no combatants added to combat engine")
	}

	if len(ce.TurnOrder) == 0 {
		return core.VictoryStatusNone, fmt.Errorf("no combatants in tracker")
	}

	ce.initializeCombatContext()
	for round := ce.CurrentRound; round <= maxRounds; round++ {
		victory, err := ce.SimulateRound()
		if err != nil {
			return core.VictoryStatusNone, err
		}

		if victory != core.VictoryStatusNone {
			return victory, nil
		}
		ce.CurrentRound++
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
		// Skip lair combatants - they don't take normal turns
		if ce.Combatants[combatantID].IsLair {
			// TODO: Handle lair actions here
			continue
		}

		combatError := ce.turnStartEvents(combatantID)
		if combatError != nil {
			return core.VictoryStatusNone, fmt.Errorf("failed to execute turn start events for combatant %d: %v", combatantID, combatError)
		}
		ce.updateCombatContext(combatantID)
		status, aiReq, combatError := ce.executeTurn(combatantID)
		if combatError != nil {
			return core.VictoryStatusNone, combatError
		}

		// Handle non-acting statuses
		if ce.shouldSkipCombatantTurn(status) {
			continue
		}

		// Execute the actions in the turn request
		combatError = ce.ProcessAIRequest(aiReq)
		if combatError != nil {
			return core.VictoryStatusNone, combatError
		}

		combatError = ce.turnEndEvents(combatantID)
		if combatError != nil {
			return core.VictoryStatusNone, fmt.Errorf("failed to execute turn end events for combatant %d: %v", combatantID, combatError)
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
	for _, index := range ce.TurnOrder {
		order++
		combatant := ce.Combatants[index]
		if combatant.IsLair {
			fmt.Printf("Order Index: %d - Initiative: %d - Name: Lair\n", order, combatant.GetInitiative())
		} else {
			fmt.Printf("Order Index: %d - Initiative: %d - Name: %s\n", order, combatant.GetInitiative(), combatant.GetEntity().GetName())
		}
	}
}

// Debug function
func (ce *CombatEngine) PrintCombatants() {
	for _, c := range ce.Combatants {
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

// initializeCombatContext initializes the combat context by setting up combatants, rounds, and relevant configuration values.
func (ce *CombatEngine) initializeCombatContext() {
	if ce.CombatContext == nil {
		ce.CombatContext = core.NewCombatContext(*ce.SimOptions)
	}
	ce.CombatContext.CurrentRound = ce.CurrentRound
	ce.CombatContext.TurnOrder = ce.TurnOrder

	ce.CombatContext.CombatantInfo = make(map[int]*core.CombatantInfo)
	for id, combatant := range ce.Combatants {
		// Skip lair combatants (they have no entity)
		if combatant.IsLair {
			continue
		}

		ce.CombatContext.CombatantInfo[id] = combatant.Info

		// Track legendary creatures
		if combatant.Entity.IsMonster() && combatant.Entity.GetIsLegendary() {
			ce.CombatContext.LegendaryCreatures[id] = true
		}
	}
}

func (ce *CombatEngine) updateCombatContext(actorID int) {
	ce.CombatContext.CurrentRound = ce.CurrentRound
	ce.CombatContext.ActingEntityID = actorID

	ce.CombatContext.CharactersInNeedOfHealing, ce.CombatContext.MonstersInNeedOfHealing = ce.calculateEntitiesNeedingHealing()
	ce.CombatContext.DeadCombatants = ce.getDeadCombatantIDs()

	// Update state for all combatants
	for id, _ := range ce.Combatants {
		if info, exists := ce.CombatContext.CombatantInfo[id]; exists {
			info.UpdateState()
		}
	}
}

func (ce *CombatEngine) calculateEntitiesNeedingHealing() ([]int, []int) {
	charNeedHealing := make([]int, 0)
	monNeedHealing := make([]int, 0)

	for id, combatant := range ce.Combatants {
		// Skip lair combatants
		if combatant.IsLair {
			continue
		}

		entity := combatant.GetEntity()

		// Calculate HP percentage
		var threshold int
		if entity.IsCharacter() {
			threshold = ce.CombatContext.Options.CharacterHealThresholdPct
		} else {
			threshold = ce.CombatContext.Options.MonsterHealThresholdPct
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

func (ce *CombatEngine) getDeadCombatantIDs() []int {
	deadCombatants := make([]int, 0)

	for id, combatant := range ce.Combatants {
		// Skip lair combatants
		if combatant.IsLair {
			continue
		}

		entity := combatant.GetEntity()

		if entity.IsDead() {
			deadCombatants = append(deadCombatants, id)
		}
	}

	return deadCombatants
}

// refreshLegendaryActions resets the legendary action count for all legendary creatures in the combat context.
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
	for _, c := range ce.Combatants {
		// Skip lair combatants
		if c.IsLair {
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
	combatant := ce.Combatants[combatantID]
	entity := combatant.GetEntity()

	// Character Specific Events
	if entity.IsCharacter() {
		c, ok := entity.(*character.Character)
		if !ok {
			return fmt.Errorf("entity is character but type assertion failed")
		}

		if !c.EntityState.IsDead && !c.EntityState.GetIsUnconscious() {
			c.EntityState.RefreshActions()
		}
	}

	// Monster Specific Events
	if entity.IsMonster() {
		ce.refreshLegendaryActions(combatantID)
	}

	return nil
}

func (ce *CombatEngine) executeTurn(combatantID int) (*core.TurnResult, *core.AIRequest, error) {
	combatant := ce.Combatants[combatantID]
	entity := combatant.GetEntity()

	// Update Combatant's AI Context
	err := entity.UpdateAICombatContext(ce.CombatContext)
	if err != nil {
		return nil, nil, err
	}

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

	// Sort by initiative (descending), then by dex, then by ID
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

		// Final tie-breaker: ID
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

func (ce *CombatEngine) checkVictoryCondition() core.VictoryStatus {
	var aliveCharacters, aliveMonsters bool

	for id := range ce.Combatants {
		// Skip lair combatants
		if ce.Combatants[id].IsLair {
			continue
		}

		entity := ce.Combatants[id].GetEntity()
		if entity.IsDead() {
			continue
		}
		if entity.IsCharacter() {
			aliveCharacters = true
		} else if entity.IsMonster() {
			aliveMonsters = true
		}

		if aliveCharacters && aliveMonsters {
			return core.VictoryStatusNone
		}
	}

	if !aliveCharacters {
		return core.VictoryStatusMonsters
	}
	if !aliveMonsters {
		return core.VictoryStatusCharacters
	}
	return core.VictoryStatusNone
}
