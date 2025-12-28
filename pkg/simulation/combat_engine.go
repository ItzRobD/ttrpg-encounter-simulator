package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
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
	tieBreakRolls    map[int]int
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
	case core.ATHeal, core.ATMonsterHeal:
		return ce.executeHeal(req)
		//case core.ATUnarmed:
		//	return ce.executeUnarmedAttack(req)
	case core.ATDragonbornBreathWeapon:
		outcome, err := req.Actor.ExecuteAIRequest(req)
		if err != nil {
			return err
		}
		return ce.processActionResults(req.Actor, outcome)
	case core.ATMonsterAction:
		return ce.executeMonsterAction(req)
	case core.ATMonsterMultiattack:
		return ce.executeMonsterMultiattack(req)
	case core.ATLegendaryAction:
		return ce.executeMonsterLegendaryAction(req)
	case core.ATLairAction:
		// Lair actions are executed by the Lair entity but follow the same
		// generic "actor executes request, engine processes effects" path.
		outcome, err := req.Actor.ExecuteAIRequest(req)
		if err != nil {
			return err
		}
		return ce.processActionResults(req.Actor, outcome)
	default:
		return fmt.Errorf("unknown action type: %v", req.ActionType)
	}

}

func (ce *CombatEngine) attachOptionsToAIRequest(aiReq *core.AIRequest) {
	aiReq.SimOptions = ce.SimOptions
}

func (ce *CombatEngine) executeWeaponAttack(aiReq *core.AIRequest) error {
	// If weapon slot is not specified, use primary slot
	if aiReq.WeaponSlot == "" {
		aiReq.WeaponSlot = core.WSPrimary
	}

	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	actionErr := ce.processActionResults(aiReq.Actor, outcome)
	if actionErr != nil {
		return actionErr
	}

	// get actor state manager
	actorESM, ok := aiReq.Actor.GetState().(*entity_state_manager.EntityStateManager)
	if !ok || actorESM == nil {
		return fmt.Errorf("actor state manager is nil or wrong type")
	}

	if !actorESM.GetHasUsedBonusAction() {
		offhandReq, ohErr := aiReq.Actor.GetAIRequest(aiReq.ActorID, core.AIReqOffhandAttack)
		if ohErr != nil {
			return ohErr
		}

		if offhandReq != nil {
			ohOutcome, ohOutcomeErr := aiReq.Actor.ExecuteAIRequest(offhandReq)
			if ohOutcomeErr != nil {
				return ohOutcomeErr
			}
			if ohResError := ce.processActionResults(aiReq.Actor, ohOutcome); ohResError != nil {
				return ohResError
			}
			actorESM.ExpendBonusAction()
		}
	}

	return nil
}

func (ce *CombatEngine) executeSpellCast(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}
	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeHeal(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}
	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeMonsterAction(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeMonsterMultiattack(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) executeMonsterLegendaryAction(aiReq *core.AIRequest) error {
	outcome, err := aiReq.Actor.ExecuteAIRequest(aiReq)
	if err != nil {
		return err
	}

	return ce.processActionResults(aiReq.Actor, outcome)
}

func (ce *CombatEngine) processActionResults(actor core.Entity, outcome *core.ActionOutcome) error {
	// Handle concentration start if the action is a concentration spell
	if outcome.IsConcentration {
		actorCombatant, exists := ce.Combatants[outcome.ActorID]
		if exists {
			// Determine targets - for now we just use the primary target of the action
			// In the future, this could be expanded for AOE concentration spells
			targets := []int{outcome.TargetID}
			// Duration is usually 10 rounds (1 minute) for most combat spells, but we should ideally get it from the spell
			// For now, default to 10 rounds as a placeholder if not specified.
			// Most concentration spells in this system are likely 1 minute.
			duration := 10

			// We need the current round
			currentRound := ce.CurrentRound

			actorCombatant.Info.StartConcentration(outcome.SpellName, targets, duration, currentRound)
		}
	}

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
				ce.applyDeflectMissiles(target.GetEntity(), &effect)
				ce.applyUncannyDodgeToEffect(target.GetEntity(), &effect)
				ce.applyEvasionToEffect(target.GetEntity(), &effect)
				res, rErr := ce.computeDamageValueAfterResistances(
					target.GetEntity(),
					effect.DamageType,
					effect.ResistBreakers,
					-effect.Value)
				if rErr != nil {
					return rErr
				}
				events.LogDamageModifiedEvent(actor, target.GetEntity(), res, actor.GetEventListener())
				hpModResult, err = target.GetEntity().ModifyHP(res.FinalValue, false, false)
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

			// Log after each effect's HP modification for clarity
			events.LogHPModifiedEvent(actor, target.GetEntity(), hpModResult, actor.GetEventListener())

			// Handle concentration check if triggered
			if hpModResult.GetTriggeredConcentrationCheck() {
				damageTaken := hpModResult.GetDamageTaken()
				dc := max(10, damageTaken/2)
				saveResult, err := target.GetEntity().MakeSavingThrow(core.AbilityConstitution, dc, false, core.DamageNone)
				if err != nil {
					return fmt.Errorf("failed to make concentration check: %v", err)
				}

				if !saveResult.GetIsSuccess() {
					target.Info.BreakConcentration()
					events.LogCombatEventMessage(target.GetEntity(), "Failed concentration check. Concentration broken.", target.GetEntity().GetEventListener())
				} else {
					events.LogCombatEventMessage(target.GetEntity(), "Succeeded concentration check. Concentration maintained.", target.GetEntity().GetEventListener())
				}
			}

			// Persist any state changes immediately
			ce.Combatants[outcome.TargetID] = target

			// If target is down or dead, check victory then stop processing remaining effects
			if target.GetEntity().IsDead() || target.GetEntity().IsUnconscious() {
				if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
					return nil
				}
				break
			}

			// Early victory check after each effect
			if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
				return nil
			}
		}

		//actor, exists := ce.Combatants[outcome.ActorID]
		//if !exists {
		//	return fmt.Errorf("actor entity not found in combat")
		//}
		//entity := actor.GetEntity()

		// Final state persist (already persisted per-effect) left for safety
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

// setupCombatTracker initializes and sorts the combat tracker based on initiative, dexterity, and ID order of combatants.
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

		// Final tie-breaker: ID
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
		status, aiReq, combatError := ce.executeTurn(combatantID)
		if combatError != nil {
			return core.VictoryStatusNone, combatError
		}

		// Handle non-acting statuses or nil action request
		if ce.shouldSkipCombatantTurn(status) || aiReq == nil {
			// If no action is taken (e.g., no valid targets), check victory now to avoid spinning
			if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
				return v, nil
			}
			continue
		}

		// Execute the actions in the turn request
		combatError = ce.ProcessAIRequest(aiReq)
		if combatError != nil {
			return core.VictoryStatusNone, combatError
		}

		// Immediately check victory after each action
		if v := ce.checkVictoryCondition(); v != core.VictoryStatusNone {
			return v, nil
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
			fmt.Printf("Order Index: %d - initiative: %d - Name: Lair\n", order, combatant.GetInitiative())
		} else {
			fmt.Printf("Order Index: %d - initiative: %d - Name: %s\n", order, combatant.GetInitiative(), combatant.GetEntity().GetName())
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
		ce.CombatContext = core.NewCombatContext(ce.SimOptions)
	}
	ce.CombatContext.CurrentRound = ce.CurrentRound
	ce.CombatContext.TurnOrder = ce.TurnOrder

	ce.CombatContext.CombatantInfo = make(map[int]*core.CombatantInfo)

	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		combatant := ce.Combatants[id]
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
	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		if info, exists := ce.CombatContext.CombatantInfo[id]; exists {
			info.UpdateState()
		}
	}
}

// calculateEntitiesNeedingHealing identifies entities needing healing and returns their IDs grouped as characters and monsters.
// It evaluates thresholds based on entity type and excludes lair combatants and unconscious entities from consideration.
// Returns a slice of character IDs and a slice of monster IDs.
func (ce *CombatEngine) calculateEntitiesNeedingHealing() ([]int, []int) {
	charNeedHealing := make([]int, 0)
	monNeedHealing := make([]int, 0)

	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		combatant := ce.Combatants[id]
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

		// Entity needs healing if below threshold and not dead
		if entity.GetHPStatus().GetHPPct() <= threshold && !entity.IsDead() {
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

	ids := ce.getSortedCombatantIDs()
	for _, id := range ids {
		combatant := ce.Combatants[id]
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
	combatant := ce.Combatants[combatantID]
	entity := combatant.GetEntity()

	// Character Specific Events
	if entity.IsCharacter() {
		c, ok := entity.(*character.Character)
		if !ok {
			return fmt.Errorf("entity is character but type assertion failed")
		}

		if !c.EntityStateManager.GetIsDead() && !c.EntityStateManager.GetIsUnconscious() {
			c.EntityStateManager.RefreshActions()
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

func (ce *CombatEngine) checkVictoryCondition() core.VictoryStatus {
	var aliveCharacters, aliveMonsters bool

	for id := range ce.Combatants {
		// Skip lair combatants
		if ce.Combatants[id].IsLair {
			continue
		}

		entity := ce.Combatants[id].GetEntity()
		// Treat unconscious as not alive for victory purposes
		if entity.IsDead() || entity.IsUnconscious() {
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

func (ce *CombatEngine) computeDamageValueAfterResistances(target core.Entity, dt core.DamageType, b []core.ResistBreaker, value int) (core.DamageModificationResult, error) {
	if target == nil {
		return core.DamageModificationResult{}, fmt.Errorf("target is nil")
	}

	targetESM, ok := target.GetState().(*entity_state_manager.EntityStateManager)
	if !ok || targetESM == nil {
		return core.DamageModificationResult{}, fmt.Errorf("target state manager is nil or wrong type")
	}

	targetResistances := targetESM.GetResistances()
	isPetrified := targetESM.GetConditions().Has(core.ConditionPetrified)
	if isPetrified {
		targetResistances = core.GetConditionEffects(core.ConditionPetrified).TemporaryResistance
	}

	result := core.DamageModificationResult{
		OriginalValue:  value,
		FinalValue:     value,
		ResistanceType: core.ResistanceNone,
	}

	if targetResistances == nil {
		return result, fmt.Errorf("target resistances are nil")
	}

	// Safe lookup: default to ResistanceNone when key is missing
	resistance := targetResistances.GetResistanceType(dt)
	result.ResistanceType = resistance

	// Resistance can only be broken if the attacker actually provides at least one breaker
	// and those breakers satisfy the target's breaker requirements for this damage type.
	brokenRes := len(b) > 0 && targetResistances.DamageTypeContainsAllBreakers(dt, b)
	result.ResistanceBroken = brokenRes
	result.ResistanceType = resistance

	switch resistance {
	case core.ResistanceNone:
		break
	case core.ResistanceVulnerable:
		if !brokenRes {
			result.FinalValue *= 2
		}
	case core.ResistanceResistant:
		if !brokenRes {
			result.FinalValue /= 2
		}
	case core.ResistanceImmune:
		if !brokenRes {
			result.FinalValue = 0
		}
	default:
		// Provide richer diagnostics to help identify unexpected values
		return core.DamageModificationResult{}, fmt.Errorf(
			"unknown resistance type for %s: damageType=%s, rawType=%q",
			target.GetName(), dt.String(), resistance,
		)
	}

	result.WasModified = result.FinalValue != value

	return result, nil
}

// applyEvasionToEffect applies the evasion feature effects for rogues and monks, modifying the effect value based on saving throws.
func (ce *CombatEngine) applyEvasionToEffect(target core.Entity, effect *core.Effect) {
	targetChar, ok := target.(*character.Character)
	if !ok {
		return // Ignore non-character targets
	}

	// Require a valid saving throw context and that it is a Dexterity save
	if effect == nil || effect.SaveCtx == nil || effect.SaveCtx.Ability != core.AbilityDexterity {
		return
	}

	// Only apply when class features are enabled (if options provided)
	if ce.SimOptions != nil && !ce.SimOptions.EnableClassFeatures {
		return
	}

	// Only Rogues/Monks have access to Evasion in this model
	if !(targetChar.Class.ID == classes.Rogue || targetChar.Class.ID == classes.Monk) {
		return // Ignore non-rogue/monk targets
	}
	hasEvasion := false

	switch targetChar.Class.ID {
	case classes.Rogue:
		hasEvasion = targetChar.Class.ClassFeatures.RogueFeatures.HasEvasion
	case classes.Monk:
		hasEvasion = targetChar.Class.ClassFeatures.MonkFeatures.HasEvasion
	default:
		return // Ignore non-rogue/monk targets
	}

	if !hasEvasion {
		return // Don't apply evasion if target doesn't have access
	}

	if effect.SaveCtx.Success && effect.SaveCtx.OnSuccess == core.DCOnSuccessHalf {
		effect.Value = 0 // Feature reduces damage to zero
	} else if !effect.SaveCtx.Success {
		effect.Value /= 2
	}
}

func (ce *CombatEngine) applyUncannyDodgeToEffect(target core.Entity, effect *core.Effect) {
	targetChar, ok := target.(*character.Character)
	if !ok {
		return // Ignore non-character targets
	}

	// Require a valid saving throw context and that it is a Dexterity save
	if effect == nil {
		return
	}

	// Only apply when class features are enabled (if options provided)
	if ce.SimOptions != nil && !ce.SimOptions.EnableClassFeatures {
		return
	}

	if targetChar.Class.ID != classes.Rogue {
		return
	}

	if targetChar.Class.ClassFeatures.RogueFeatures.HasUncannyDodge &&
		!targetChar.EntityStateManager.GetHasUsedReaction() {
		effect.Value /= 2
		targetChar.EntityStateManager.ExpendReaction()
		return
	}
}

func (ce *CombatEngine) applyDeflectMissiles(target core.Entity, effect *core.Effect) {
	targetChar, ok := target.(*character.Character)
	if !ok {
		return // Ignore non-character targets
	}

	// Require a valid saving throw context and that it is a Dexterity save
	if effect == nil {
		return
	}

	// Only apply when class features are enabled (if options provided)
	if ce.SimOptions != nil && !ce.SimOptions.EnableClassFeatures {
		return
	}

	if targetChar.Class.ID != classes.Monk {
		return
	}

	if targetChar.Class.ClassFeatures.MonkFeatures != nil &&
		targetChar.Class.ClassFeatures.MonkFeatures.HasDeflectMissiles &&
		effect.AttackCtx.IsRanged {
		dexMod, err := targetChar.GetAbilityScoreModifier(core.AbilityDexterity)
		if err != nil {
			return
		}
		roll := targetChar.RollManager.RollDie(core.D10)
		effect.Value = int(math.Max(0, float64(effect.Value)-float64(dexMod)-float64(roll)-float64(targetChar.Level)))
		targetChar.EntityStateManager.ExpendReaction()
		return
	}
}
