package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"fmt"
	"math/rand/v2"
	"sort"
	"time"
)

// EncounterDirector is the centralized combat engine ("The DM").
type EncounterDirector struct {
	// RNG
	rng  *rand.Rand
	seed core.Seed
	// Dice Engine
	RollManager *roll_manager.RollManager

	// AI and Rules
	AIDirector  *AIDirector
	Adjudicator *Adjudicator

	// Feature Handlers
	featureHandlers map[core.SpecialAbility]FeatureHandler

	// Combatants
	Actors            map[int]*actor.Actor
	LegendaryActorIDs []int
	HasLair           bool
	Statistics        *EncounterStatistics

	// Initiative results
	InitiativeResults map[int]int

	// Turn Management
	TurnOrder    []int
	CurrentRound int

	// Config
	SimOptions *core.SimulationOptions

	// Logging
	EventContext *events.EventContext
	CombatLog    []events.TimelineEvent
}

// NewEncounterDirector initializes a new director with a shared RNG source.
func NewEncounterDirector(seed core.Seed, options *core.SimulationOptions) *EncounterDirector {
	rng := rand.New(rand.NewPCG(seed.Seed1, seed.Seed2))
	ed := &EncounterDirector{
		rng:               rng,
		seed:              seed,
		RollManager:       roll_manager.NewRollManager(rng),
		AIDirector:        NewAIDirector(rng),
		featureHandlers:   make(map[core.SpecialAbility]FeatureHandler),
		Actors:            make(map[int]*actor.Actor),
		InitiativeResults: make(map[int]int),
		SimOptions:        options,
		CurrentRound:      0,
		EventContext:      events.NewEventContext(),
		CombatLog:         make([]events.TimelineEvent, 0),
	}
	ed.Adjudicator = NewReferee(ed)
	ed.registerSRDFeatures()
	return ed
}

func (ed *EncounterDirector) registerSRDFeatures() {
	ed.featureHandlers[core.SpecAbilityAssassinate] = ed.HandleAssassinate
	ed.featureHandlers[core.SpecAbilityBerserk] = ed.HandleBerserk
	ed.featureHandlers[core.SpecAbilityBloodFrenzy] = ed.HandleBloodFrenzy
	//ed.featureHandlers[core.SpecAbilityConsumeLife] = ed.HandleConsumeLife
	ed.featureHandlers[core.SpecAbilityCorrosiveForm] = ed.HandleCorrosiveForm
	ed.featureHandlers[core.SpecAbilityDeathBurst] = ed.HandleDeathBurstAndThroes
	ed.featureHandlers[core.SpecAbilityDeathThroes] = ed.HandleDeathBurstAndThroes
	ed.featureHandlers[core.SpecAbilityDivineEminence] = ed.HandleSmiteLikeFeature
	ed.featureHandlers[core.SpecAbilityDivineSmite] = ed.HandleSmiteLikeFeature
	ed.featureHandlers[core.SpecAbilityEvasion] = ed.HandleEvasion
	ed.featureHandlers[core.SpecAbilityFireAura] = ed.HandleAura
	ed.featureHandlers[core.SpecAbilityFireForm] = ed.HandleAura
	ed.featureHandlers[core.SpecAbilityCunning] = ed.HandleCunning
	//ed.featureHandlers[core.SpecAbilityGibbering] = ed.HandleGibbering
	ed.featureHandlers[core.SpecAbilityHeatedBody] = ed.HandleMeleeTouchDamage
	ed.featureHandlers[core.SpecAbilityLegendaryResistance] = ed.HandleLegendaryResistance
	ed.featureHandlers[core.SpecAbilityAbsorption] = ed.HandleAbsorption
	ed.featureHandlers[core.SpecAbilityLimitedMagicImmunity] = ed.HandleLimitedMagicImmunity
	ed.featureHandlers[core.SpecAbilityMagicResistance] = ed.HandleMagicResistance
	ed.featureHandlers[core.SpecAbilityMagicWeapons] = ed.HandleMagicWeapons
	ed.featureHandlers[core.SpecAbilityMartialAdvantage] = ed.HandleMartialAdvantage
	ed.featureHandlers[core.SpecAbilityPackTactics] = ed.HandlePackTactics
	ed.featureHandlers[core.SpecAbilityReckless] = ed.HandleReckless
	ed.featureHandlers[core.SpecAbilityReflectiveCarapace] = nil
	ed.featureHandlers[core.SpecAbilityRegeneration] = ed.HandleRegeneration
	ed.featureHandlers[core.SpecAbilityRelentless] = ed.HandleRelentless
	ed.featureHandlers[core.SpecAbilityRelentlessEndurance] = ed.HandleRelentless
	ed.featureHandlers[core.SpecAbilitySneakAttack] = ed.HandleSneakAttack
	ed.featureHandlers[core.SpecAbilityUndeadFortitude] = ed.HandleUndeadFortitude
	ed.featureHandlers[core.SpecAbilitySecondWind] = ed.HandleSecondWind
	ed.featureHandlers[core.SpecAbilityRageStrengthSave] = ed.HandleSaveAdvantage
	ed.featureHandlers[core.SpecAbilityDangerSense] = ed.HandleSaveAdvantage
	ed.featureHandlers[core.SpecAbilitySlipperyMind] = ed.HandleSaveAdvantage
	ed.featureHandlers[core.SpecAbilityDeflectMissiles] = ed.HandleDeflectMissiles
	ed.featureHandlers[core.SpecAbilityFightingStyleArchery] = ed.HandleFightingStyle
	ed.featureHandlers[core.SpecAbilityFightingStyleDuel] = ed.HandleFightingStyle
	ed.featureHandlers[core.SpecAbilityFightingStyleGWF] = ed.HandleFightingStyle
	ed.featureHandlers[core.SpecAbilityFightingStyleTWF] = ed.HandleFightingStyle
	ed.featureHandlers[core.SpecAbilityDwarvenResilience] = ed.HandleSaveAdvantage
	ed.featureHandlers[core.SpecAbilityHalflingLucky] = ed.HandleHalflingLucky
	ed.featureHandlers[core.SpecAbilityIndomitable] = ed.HandleIndomitable
	ed.featureHandlers[core.SpecAbilityImprovedDivineSmite] = ed.HandleImprovedDivineSmite
	ed.featureHandlers[core.SpecAbilityLayOnHands] = ed.HandleLayOnHands
	ed.featureHandlers[core.SpecAbilitySavageAttacks] = ed.HandleSavageAttacks
	ed.featureHandlers[core.SpecAbilityBrutalCritical] = ed.HandleBrutalCritical
	ed.featureHandlers[core.SpecAbilityRelentlessRage] = ed.HandleRelentlessRage
	ed.featureHandlers[core.SpecAbilityRageExtraDamage] = ed.HandleRageExtraDamage
	ed.featureHandlers[core.SpecAbilityUncannyDodge] = ed.HandleUncannyDodge
	ed.featureHandlers[core.SpecAbilityElusive] = ed.HandleElusive
}

// AddActor adds a hydrated actor to the simulation and assigns a unique InstanceID.
func (ed *EncounterDirector) AddActor(a *actor.Actor) {
	// Simple auto-increment for InstanceID
	instanceID := len(ed.Actors) + 1
	a.InstanceID = instanceID
	ed.Actors[instanceID] = a

	// Account for lair specific needs
	if ed.HasLair && a.ActorType == core.ActorTypeLair {
		// TODO: Log this error, there can only be one lair actor
		return
	}
	if a.ActorType == core.ActorTypeLair {
		ed.HasLair = true
		// This should be handled in the config; safeguard to ensure proper targeting
		if a.Side != core.SideMonsters {
			a.Side = core.SideMonsters
		}
	}

	// Maintain a separate list of legendary actors
	if a.Metadata.IsLegendary {
		ed.LegendaryActorIDs = append(ed.LegendaryActorIDs, instanceID)
		sort.Ints(ed.LegendaryActorIDs)
	}
}

// SetupEncounter performs Round 0 operations.
func (ed *EncounterDirector) SetupEncounter() {
	ed.CurrentRound = 0
	ed.EventContext.GenerateSequenceID()
	ed.EventContext.GenerateCurrentID()

	ed.LogEvent(events.EventCombatStart, nil, nil)
	ed.RollInitiative()
	ed.SortTurnOrder()
	ed.Statistics = NewEncounterStatistics(ed.Actors)

	// Dispatch HookOnCombatStart
	ed.dispatchHooks(nil, core.HookOnCombatStart, nil)
}

// RollInitiative rolls initiative for all actors currently in the director.
func (ed *EncounterDirector) RollInitiative() {
	ed.LogEvent(events.EventInitiative, nil, nil)

	// Sort IDs for deterministic RNG consumption
	ids := make([]int, 0, len(ed.Actors))
	for id := range ed.Actors {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		a := ed.Actors[id]
		// Lair actors always have initiative 20 losing all ties
		if a.ActorType == core.ActorTypeLair {
			ed.InitiativeResults[id] = 20
			ed.LogEvent(events.EventInitiative, a, map[string]interface{}{
				"total":   20,
				"is_lair": true,
			})
			continue
		}

		dexMod := a.Abilities.GetAbilityModifier(core.AbilityDexterity)

		opts := roll_manager.RollOptions{
			RollType: core.DiceRollInitiative,
			Modifier: dexMod,
		}

		// Check for initiative-affecting features
		if a.HasFeature(core.SpecAbilityFeralInstinct) {
			opts.Advantage = core.RollAdvantage
		}

		res := ed.RollManager.RollD20(opts)
		ed.InitiativeResults[id] = res.Total
		ed.LogEvent(events.EventInitiative, a, map[string]interface{}{
			"roll": res,
		})
	}
}

// SortTurnOrder sorts actors based on initiative results.
func (ed *EncounterDirector) SortTurnOrder() {
	// Use Actor InstanceIDs for sorting to handle duplicate base Actor types properly
	ids := make([]int, 0, len(ed.Actors))
	for id := range ed.Actors {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	sort.Slice(ids, func(i, j int) bool {
		instanceID_I := ids[i]
		instanceID_J := ids[j]

		initI := ed.InitiativeResults[instanceID_I]
		initJ := ed.InitiativeResults[instanceID_J]

		if initI != initJ {
			return initI > initJ
		}

		// Tie-break: Lair always loses
		if ed.Actors[instanceID_I].ActorType == core.ActorTypeLair {
			return false
		}
		if ed.Actors[instanceID_J].ActorType == core.ActorTypeLair {
			return true
		}

		// Tie-break: Dexterity score
		dexI := ed.Actors[instanceID_I].Abilities.GetScore(core.AbilityDexterity)
		dexJ := ed.Actors[instanceID_J].Abilities.GetScore(core.AbilityDexterity)
		if dexI != dexJ {
			return dexI > dexJ
		}

		// Final tie-break: Player Characters over Monsters
		typeI := ed.Actors[instanceID_I].ActorType
		typeJ := ed.Actors[instanceID_J].ActorType
		if typeI != typeJ {
			return typeI == core.ActorTypeCharacter
		}

		return instanceID_I < instanceID_J // Consistent fallback using unique instance ID
	})

	ed.TurnOrder = ids
}

// AdvanceRound increments the round counter.
func (ed *EncounterDirector) AdvanceRound() {
	ed.CurrentRound++
}

// SimulateRound iterates through the turn order and processes each actor's turn.
func (ed *EncounterDirector) SimulateRound() (core.VictoryStatus, error) {
	ed.AdvanceRound()
	ed.processRoundStartEvents()

	victory := core.VictoryStatusNone

	for _, actorID := range ed.TurnOrder {

		if !ed.SimOptions.DebugEnableCharacterTurns && ed.Actors[actorID].IsCharacter() {
			continue
		}

		if !ed.SimOptions.DebugEnableMonsterTurns && ed.Actors[actorID].IsMonster() {
			continue
		}

		if !ed.SimOptions.DebugEnableLairTurns && ed.Actors[actorID].IsLair() {
			continue
		}

		a, ok := ed.Actors[actorID]
		if !ok {
			continue
		}

		// Skip dead or incapacitated actors
		if a.StateManager.CurrentHP <= 0 {
			continue
		}

		err := ed.ExecuteTurn(actorID)
		if err != nil {
			return victory, err
		}

		victory = ed.checkVictoryConditions()
		if victory != core.VictoryStatusNone {
			ed.LogEvent(events.EventVictory, nil, map[string]interface{}{
				"winner": victory,
				"rounds": ed.CurrentRound,
			})
			return victory, nil
		}

		err = ed.ExecuteLegendaryActions(actorID)
		if err != nil {
			return victory, err
		}
	}

	victory = ed.checkVictoryConditions()
	if victory != core.VictoryStatusNone {
		ed.LogEvent(events.EventVictory, nil, map[string]interface{}{
			"winner": victory,
			"rounds": ed.CurrentRound,
		})
	}

	return victory, nil
}

// ExecuteLegendaryActions handles the execution of legendary actions for valid actors during an encounter turn.
// It prioritizes available actions and enforces restrictions based on conditions such as health and action limits.
func (ed *EncounterDirector) ExecuteLegendaryActions(currentTurnActorID int) error {
	if ed.LegendaryActorIDs == nil {
		return nil
	}

	// If there's more than one legendary actor we should prioritize using different actions
	// instead of always using the first one.
	var selectedLegendaryActor *actor.Actor
	for _, id := range ed.LegendaryActorIDs {
		a, ok := ed.Actors[id]
		if !ok {
			continue
		}
		// Legendary actors can't act after their turn or if they're dead
		if currentTurnActorID == id ||
			a.StateManager.CurrentHP <= 0 ||
			!a.StateManager.CanActConditions() {
			continue
		}
		if a.StateManager.LegendaryActionUsedCount == 0 {
			selectedLegendaryActor = a
			break
		}
	}
	if selectedLegendaryActor == nil {
		for _, id := range ed.LegendaryActorIDs {
			a, ok := ed.Actors[id]
			if !ok {
				continue
			}
			// Legendary actors can't act after their turn or if they're dead
			if currentTurnActorID == id ||
				a.StateManager.CurrentHP <= 0 ||
				!a.StateManager.CanActConditions() {
				continue
			}

			if a.StateManager.LegendaryActionUsedCount >= a.StateManager.MaxLegendaryActions {
				continue
			}

			for _, act := range a.Actions {
				if act.Cost.ActivationType == core.ActLegendary &&
					a.StateManager.RemainingLegendaryActionCount() >= act.Cost.Value {
					selectedLegendaryActor = a
					break
				}
			}
			if selectedLegendaryActor != nil {
				break
			}
		}
	}

	// Choose the Legendary Action to use
	if selectedLegendaryActor == nil { // No available actions for legendary actors
		return nil
	}

	intents := ed.AIDirector.SelectAction(selectedLegendaryActor, core.DecisionLegendary, ed)
	if len(intents) == 0 { // No available actions for legendary actors
		return nil
	}
	// We shouldn't have more than one legendary action but safeguard
	if len(intents) > 1 {
		intents = intents[:1]
	}

	id := ed.LogEvent(events.EventLegendaryAction, selectedLegendaryActor, nil)
	ed.EventContext.PushParent(id)
	defer ed.EventContext.PopParent()

	ed.LogEvent(events.EventDecisionStart, selectedLegendaryActor, map[string]interface{}{
		"decision": core.DecisionLegendary,
		"action":   intents[0].Action.Name,
	})

	err := ed.Adjudicator.ResolveAction(selectedLegendaryActor, intents[0])
	if err != nil {
		return err
	}

	return nil
}

// ExecuteTurn handles the lifecycle of a single actor's turn.
func (ed *EncounterDirector) ExecuteTurn(actorID int) error {
	a, ok := ed.Actors[actorID]
	if !ok {
		return nil
	}

	// 1. Turn Start Events
	id := ed.LogEvent(events.EventTurnStart, a, nil)
	ed.EventContext.PushParent(id)
	defer ed.EventContext.PopParent()

	ed.processTurnStart(a)

	// 1a. PC Death Saves (if at 0 HP)
	if a.StateManager.CurrentHP <= 0 && a.IsCharacter() && a.StateManager.HealthState != core.HealthStateDead {
		ed.processDeathSaves(a)
		// If stabilized or still unconscious, turn ends
		if a.StateManager.CurrentHP <= 0 || a.StateManager.Conditions.Has(core.ConditionUnconscious) {
			ed.processTurnEnd(a)
			return nil
		}
	}

	// 2. Action Selection (AIDirector)
	decision := ed.AIDirector.SelectActionDecision(a, ed)

	// Parent 1: Decision Scope
	did := ed.LogEvent(events.EventDecisionStart, a, map[string]interface{}{
		"decision": decision,
	})
	ed.EventContext.PushParent(did)

	// Intents holds a slice of selected actions and intended targets.
	intents := ed.AIDirector.SelectAction(a, decision, ed)
	if len(intents) == 0 {
		ed.EventContext.PopParent()
		ed.processTurnEnd(a)
		return nil // No action taken
	}

	// 3. Action Resolution (Adjudicator)
	// Each intent represents a top-level decision (like "Use Action" or "Use Bonus Action")
	// The Adjudicator will handle the breakdown of Multiattack or Extra Attack internally
	// to ensure targeting is decided per individual strike if needed.
	for _, intent := range intents {
		err := ed.Adjudicator.ResolveAction(a, intent)
		if err != nil {
			ed.EventContext.PopParent()
			return err
		}
	}

	ed.EventContext.PopParent()

	// 4. Turn End Events
	ed.processTurnEnd(a)

	// Advance Sequence ID for the next turn
	ed.EventContext.GenerateSequenceID()

	return nil
}

func (ed *EncounterDirector) processTurnStart(a *actor.Actor) {
	a.StateManager.HasTakenTurnThisCombat = true
	a.StateManager.ResetStateForTurnStart()

	// Recharge monster actions
	for _, action := range a.Actions {
		if action.RechargeValue > 0 {
			if a.StateManager.Resource[action.Name] == 0 {
				rechargeRoll := ed.RollManager.RollDie(core.D6)

				success := rechargeRoll >= action.RechargeValue
				ed.LogEvent(events.EventRecharge, a, map[string]interface{}{
					"action":  action.Name,
					"roll":    rechargeRoll,
					"success": success,
				})

				if success {
					a.StateManager.Resource[action.Name] = 1
				}
			}
		}
	}

	// Turn start hooks
	ed.dispatchHooks(a, core.HookOnTurnStart, nil)
}

func (ed *EncounterDirector) processTurnEnd(a *actor.Actor) {
	ed.LogEvent(events.EventTurnEnd, a, nil)

	// Turn end hooks
	ed.dispatchHooks(a, core.HookOnTurnEnd, nil)

	a.StateManager.HasTakenTurnThisCombat = true
}

// ExportTimeline returns the current combat log as a slice of TimelineEvents.
func (ed *EncounterDirector) ExportTimeline() []events.TimelineEvent {
	return ed.CombatLog
}

func (ed *EncounterDirector) processDeathSaves(a *actor.Actor) {
	if a.StateManager.Conditions.Has(core.ConditionStable) {
		return
	}

	roll := ed.RollManager.RollDie(core.D20)

	if roll == 20 {
		// Gain 1 HP and become conscious
		hpRes := a.StateManager.ModifyHP(1, false, true)
		ed.LogEvent(events.EventHPModified, a, map[string]interface{}{
			"result": hpRes,
			"note":   "Natural 20 on Death Save",
		})
		a.StateManager.Conditions.Remove(core.ConditionUnconscious)
		a.StateManager.Conditions.Remove(core.ConditionProne)
		a.StateManager.Conditions.Remove(core.ConditionStable)
		if ed.Statistics != nil {
			ed.Statistics.ResetDeathSaveStats(a.InstanceID)
		}
		ed.LogEvent(events.EventDeathSave, a, map[string]interface{}{
			"roll":      roll,
			"success":   true,
			"is_nat_20": true,
		})
		return
	}

	success := roll >= 10
	isNat1 := roll == 1

	if isNat1 {
		a.StateManager.DeathSaveFailures += 2
		if ed.Statistics != nil {
			ed.Statistics.DeathSave(a.InstanceID, false)
			ed.Statistics.DeathSave(a.InstanceID, false)
		}
	} else if success {
		a.StateManager.DeathSaveSuccesses++
		if ed.Statistics != nil {
			ed.Statistics.DeathSave(a.InstanceID, true)
		}
	} else {
		a.StateManager.DeathSaveFailures++
		if ed.Statistics != nil {
			ed.Statistics.DeathSave(a.InstanceID, false)
		}
	}

	ed.LogEvent(events.EventDeathSave, a, map[string]interface{}{
		"roll":      roll,
		"success":   success,
		"is_nat_1":  isNat1,
		"failures":  a.StateManager.DeathSaveFailures,
		"successes": a.StateManager.DeathSaveSuccesses,
	})

	// Check if stabilized
	if a.StateManager.DeathSaveSuccesses >= 3 {
		// Stabilized at 0 HP
		a.StateManager.DeathSaveSuccesses = 0
		a.StateManager.DeathSaveFailures = 0
		a.StateManager.Conditions.Add(core.ConditionStable)
		if ed.Statistics != nil {
			ed.Statistics.ResetDeathSaveStats(a.InstanceID)
		}
		// In a full sim, we might add a 'Stable' condition or just leave at 0 HP.
		// For now, we just reset saves. They remain unconscious until healed.
	}

	// Re-evaluate health state (might have died)
	a.StateManager.HealthState = a.StateManager.GetHealthState(true)
}

func (ed *EncounterDirector) processRoundStartEvents() {
	for _, a := range ed.Actors {
		a.StateManager.ResetStateForRoundStart()
	}
}

func (ed *EncounterDirector) dispatchHooks(a *actor.Actor, hook core.HookType, ctx *FeatureContext) {
	if a == nil {
		return
	}
	if ed.SimOptions != nil && !ed.SimOptions.EnableSpecialAbilities {
		return
	}

	for _, f := range a.Features {
		if f.Hooks[hook] {
			err := ed.resolveFeatureHook(a, f, hook, ctx)
			if err != nil {
				// Log the error but continue the simulation to prevent total failure
				fmt.Printf("Feature Error: Actor %d (%s), Feature %s, Hook %v: %v\n",
					a.InstanceID, a.Name, f.Name, hook, err)
			}
		}
	}
}

func (ed *EncounterDirector) resolveFeatureHook(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	ed.LogEvent(events.EventFeatureTrigger, a, map[string]interface{}{
		"feature": f.Name,
		"hook":    hook,
	})

	if handler, ok := ed.featureHandlers[f.Name]; ok {
		return handler(a, f, hook, ctx)
	}
	return nil
}

// GetEnemyTargets returns a map of actors that are enemies of the given actor.
func (ed *EncounterDirector) GetEnemyTargets(a *actor.Actor) map[int]*actor.Actor {
	enemies := make(map[int]*actor.Actor)
	for id, other := range ed.Actors {
		if other.StateManager.CurrentHP <= 0 {
			continue
		}
		// Exclude Lair actors as they are environment/neutral
		if other.ActorType == core.ActorTypeLair {
			continue
		}
		// Basic team logic: different sides
		if a.Side != other.Side {
			enemies[id] = other
		}
	}
	return enemies
}

// GetAllyTargets returns a map of actors that are allies of the given actor.
func (ed *EncounterDirector) GetAllyTargets(a *actor.Actor) map[int]*actor.Actor {
	allies := make(map[int]*actor.Actor)
	for id, other := range ed.Actors {
		if other.StateManager.CurrentHP <= 0 {
			continue
		}
		if a.Side == other.Side {
			allies[id] = other
		}
	}
	return allies
}

// CalculateAvgEnemyDamage returns the average damage the enemy team can deal per turn.
func (ed *EncounterDirector) CalculateAvgEnemyDamage(a *actor.Actor) int {
	enemies := ed.GetEnemyTargets(a)
	totalAvg := 0
	for _, enemy := range enemies {
		for _, act := range enemy.Actions {
			for _, db := range act.DiceBlock {
				avg, _ := core.GetAverageRoll(db.NumberOfDice, db.Die, db.Modifier)
				totalAvg += avg
			}
		}
	}
	if len(enemies) == 0 {
		return 0
	}
	return totalAvg / len(enemies)
}

// LogEvent adds an event to the combat log with current context.
func (ed *EncounterDirector) LogEvent(eventType events.EventType, a *actor.Actor, data interface{}) string {
	eventID := core.NewUUIDv7()
	timelineEvent := events.TimelineEvent{
		Timestamp: time.Now(),
		ID:        eventID,
		ParentID:  ed.EventContext.GetParentID(),
		Round:     ed.CurrentRound,
		Type:      eventType,
		Data:      data,
	}

	if a != nil {
		timelineEvent.Actor = &events.ActorInfo{
			Name:       a.Name,
			InstanceID: a.InstanceID,
			Type:       a.ActorType,
			Side:       a.Side,
		}
	}

	ed.CombatLog = append(ed.CombatLog, timelineEvent)
	return eventID
}

// checkVictoryConditions evaluates the state of actors to determine if victory conditions are met and returns the result.
func (ed *EncounterDirector) checkVictoryConditions() core.VictoryStatus {
	var cAlive, mAlive bool
	for _, a := range ed.Actors {
		if a.ActorType == core.ActorTypeLair {
			continue
		}
		isDown := a.StateManager.CurrentHP <= 0
		if !isDown {
			if a.Side == core.SideCharacters {
				cAlive = true
			} else if a.Side == core.SideMonsters {
				mAlive = true
			}
		}
		if cAlive && mAlive {
			return core.VictoryStatusNone
		}
	}

	if cAlive && !mAlive {
		return core.VictoryStatusCharacters
	}
	if mAlive && !cAlive {
		return core.VictoryStatusMonsters
	}
	if !mAlive && !cAlive {
		return core.VictoryStatusDraw
	}

	return core.VictoryStatusNone
}
