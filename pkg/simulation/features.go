package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"fmt"
	"sort"
	"strings"
)

// FeatureHandler is a function that resolves the logic for a specific feature.
type FeatureHandler func(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error

// FeatureContext provides additional context for feature resolution
type FeatureContext struct {
	Target        *actor.Actor
	AttackContext *AttackContext
	SaveContext   *SaveContext
	DamageContext *DamageContext
}

type AttackContext struct {
	AttackRoll       *roll_manager.RollOptions
	DamageRoll       *roll_manager.RollOptions
	IsCritical       *bool
	Action           *core.Action
	WeaponModifiers  *core.WeaponModifiers
	WeaponProperties *core.WeaponProperties
}

type DamageContext struct {
	DamageValue   *int
	DamageType    core.DamageType
	IsAbsorption  bool
	AbsorbedValue int
}

// SaveContext provides additional context for saving throw resolution
type SaveContext struct {
	Target      *actor.Actor
	Options     *roll_manager.RollOptions
	SaveSuccess bool
	IsPostRoll  bool
}

func (ed *EncounterDirector) validateFeatureDice(f core.Feature) error {
	if f.Data.NumberOfDice <= 0 {
		return fmt.Errorf("invalid number of dice for %s: %d", f.Name, f.Data.NumberOfDice)
	}
	if f.Data.Die == 0 || !f.Data.Die.IsValid() {
		return fmt.Errorf("invalid die for %s: %d", f.Name, f.Data.Die)
	}
	return nil
}

func (ed *EncounterDirector) validateFeatureValue(f core.Feature) error {
	if f.Data.Value <= 0 {
		return fmt.Errorf("invalid value for %s: %d", f.Name, f.Data.Value)
	}
	return nil
}

func (ed *EncounterDirector) validateFeatureDC(f core.Feature) error {
	if f.Data.DC <= 0 {
		return fmt.Errorf("invalid DC for %s: %d", f.Name, f.Data.DC)
	}
	if f.Data.Ability == "" {
		return fmt.Errorf("invalid ability for %s", f.Name)
	}
	if f.Data.DCOnSuccess == "" {
		return fmt.Errorf("invalid DC on success for %s", f.Name)
	}
	if f.Data.DamageType == nil || f.Data.DamageType[0] == core.DamageNone {
		return fmt.Errorf("invalid damage type for %s", f.Name)
	}
	return nil
}

func (ed *EncounterDirector) validateFeatureDamage(f core.Feature) error {
	if f.Data.DamageType == nil || f.Data.DamageType[0] == core.DamageNone {
		return fmt.Errorf("invalid damage type for %s", f.Name)
	}
	return nil
}

// Handlers for feature resolution, returns an error for log purposes

func (ed *EncounterDirector) HandleAssassinate(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.Target == nil || ctx.AttackContext == nil || ctx.AttackContext.AttackRoll == nil {
		return fmt.Errorf("invalid context for assassination")
	}

	if hook == core.HookOnSelfAttack {
		// Advantage against any creature that hasn't taken a turn
		if !ctx.Target.StateManager.HasTakenTurnThisCombat {
			if ctx.AttackContext.AttackRoll != nil {
				ctx.AttackContext.AttackRoll.AdvantageCount++
				ctx.AttackContext.AttackRoll.Advantage = ctx.AttackContext.AttackRoll.CalculateAdvantage()
			}
		}
	}
	return nil
}

func (ed *EncounterDirector) HandleBerserk(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if hook != core.HookOnTurnStart {
		return fmt.Errorf("invalid hook for berserk: %s", hook)
	}

	if err := ed.validateFeatureValue(f); err != nil {
		return err
	}
	if f.Data.Die == 0 || !f.Data.Die.IsValid() {
		return fmt.Errorf("invalid berserk die: %d", f.Data.Die)
	}

	res := ed.RollManager.RollDie(f.Data.Die)
	if res == f.Data.Value {
		a.StateManager.Conditions.Add(core.ConditionBerserk)
	}

	return nil
}

func (ed *EncounterDirector) HandleBloodFrenzy(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.Target == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for blood frenzy")
	}

	if hook == core.HookOnTargetInjured {
		if ctx.AttackContext.Action.ActionType != core.ATMelee {
			return fmt.Errorf("blood frenzy is only valid on melee attacks; action type: %s", ctx.AttackContext.Action.ActionType)
		}

		if ctx.Target.StateManager.GetHealthState(ctx.Target.IsCharacter()) != core.HealthStateHealthy &&
			ctx.Target.StateManager.GetHealthState(ctx.Target.IsCharacter()) != core.HealthStateDead &&
			ctx.Target.StateManager.GetHealthState(ctx.Target.IsCharacter()) != core.HealthStateUnconscious {
			if ctx.AttackContext.AttackRoll != nil {
				ctx.AttackContext.AttackRoll.AdvantageCount++
				ctx.AttackContext.AttackRoll.Advantage = ctx.AttackContext.AttackRoll.CalculateAdvantage()
			}
		}
	}

	return nil
}

func (ed *EncounterDirector) HandleMeleeTouchDamage(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.Target == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for %s", strings.ToLower(string(f.Name)))
	}

	if err := ed.validateFeatureDice(f); err != nil {
		return err
	}

	if hook == core.HookOnSelfHit {
		if ctx.AttackContext.Action.ActionType == core.ATMelee {
			opts := roll_manager.RollOptions{
				RollType: core.DiceRollDamage,
			}

			dmgRollRes := ed.RollManager.RollDice(f.Data.NumberOfDice, f.Data.Die, opts)
			dmgValue := dmgRollRes.Total

			attacker := ctx.Target
			attacker.StateManager.ModifyHP(-dmgValue, false, attacker.IsCharacter())

			// Update Stats
			ed.Statistics.AddDamage(ctx.Target.InstanceID, dmgValue)
			ed.Adjudicator.UpdateHealTargetsAfterDamage(ctx.Target)
		}
	}
	return nil
}

func (ed *EncounterDirector) HandleCorrosiveForm(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	return ed.HandleMeleeTouchDamage(a, f, hook, ctx)
}

func (ed *EncounterDirector) HandleDeathBurstAndThroes(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.Target == nil {
		return fmt.Errorf("invalid context for %s", strings.ToLower(string(f.Name)))
	}

	if err := ed.validateFeatureDice(f); err != nil {
		return err
	}

	if err := ed.validateFeatureDC(f); err != nil {
		return err
	}

	if hook == core.HookOnSelfDeath {
		// Determine targets
		initialTargetID := ctx.Target.InstanceID
		targetIDs := ed.Adjudicator.IdentifyAOETargets(a, initialTargetID, ed.SimOptions.MonsterDeathEffectsHitAllies)

		fAction := core.Action{
			Name:       string(f.Name),
			ActionType: core.ATFeature,
			DiceBlock: []core.DiceBlock{
				{
					NumberOfDice: f.Data.NumberOfDice,
					Die:          f.Data.Die,
					DamageType:   f.Data.DamageType[0],
				},
			},
			HasDC:       true,
			DCSaveDC:    f.Data.DC,
			DCAbility:   f.Data.Ability,
			DCOnSuccess: f.Data.DCOnSuccess,
		}

		for _, id := range targetIDs {
			target := ed.Actors[id]
			saveRes, _ := ed.Adjudicator.ResolveSavingThrow(a, target, &fAction)

			err := ed.Adjudicator.resolveDamage(a, target, &fAction, saveRes, false)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ed *EncounterDirector) HandleSmiteLikeFeature(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.Target == nil || ctx.AttackContext == nil || ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for %s", strings.ToLower(string(f.Name)))
	}

	if a == nil {
		return fmt.Errorf("invalid actor for %s", strings.ToLower(string(f.Name)))
	}
	if hook == core.HookOnSelfHit {
		if f.Name == core.SpecAbilityDivineEminence && a.StateManager.BonusActionUsedCount != 0 {
			return nil
		}

		// Only on melee attacks
		if ctx.AttackContext.Action.ActionType != core.ATMelee {
			return nil
		}

		// Check if we should smite based on sim options if it's a Paladin Smite
		if f.Name == core.SpecAbilityDivineSmite {
			if ed.SimOptions != nil && !ed.SimOptions.PaladinAlwaysSmite {
				return nil
			}
		}

		// Check for spell slots
		if len(a.StateManager.CurrentSlots) == 0 {
			return nil
		}

		// Find the best spell slot to use
		selectedLevel := -1
		useHighest := false
		if ed.SimOptions != nil && ed.SimOptions.PaladinUseHighestSmiteSlot {
			useHighest = true
		}

		levels := make([]int, 0, len(a.StateManager.CurrentSlots))
		for lvl, count := range a.StateManager.CurrentSlots {
			if count > 0 && lvl > 0 {
				levels = append(levels, lvl)
			}
		}

		if len(levels) == 0 {
			return nil
		}

		sort.Ints(levels)

		// Decide level based on AI
		if ed.SimOptions != nil && ed.SimOptions.UseWeightedAI {
			// If weighted AI, use potency to decide
			if ed.AIDirector.ShouldExpendHighResource(a, ctx.Target) {
				selectedLevel = levels[len(levels)-1] // Highest
			} else {
				selectedLevel = levels[0] // Lowest
			}
		} else {
			if useHighest {
				selectedLevel = levels[len(levels)-1]
			} else {
				selectedLevel = levels[0]
			}
		}

		// Expend the slot
		a.StateManager.CurrentSlots[selectedLevel]--
		// Divine eminence requires a bonus action
		if f.Name == core.SpecAbilityDivineEminence {
			a.StateManager.BonusActionUsedCount++
		}

		// Calculate damage
		numDice := f.Data.NumberOfDice
		// Scaling: +1 die per level above 1st
		if selectedLevel > 1 {
			numDice += selectedLevel - 1
		}

		// Bonus damage for specific types (e.g. Divine Smite vs Undead/Fiend)
		if f.Data.BonusTargetTypes != nil && len(f.Data.BonusTargetTypes) > 0 {
			targetType := ctx.Target.Metadata.MonsterType
			for _, bt := range f.Data.BonusTargetTypes {
				if bt == targetType {
					numDice++
					break
				}
			}
		}

		opts := roll_manager.RollOptions{
			RollType: core.DiceRollDamage,
		}

		dmgRollRes := ed.RollManager.RollDice(numDice, f.Data.Die, opts)
		dmgValue := dmgRollRes.Total

		// Apply damage to target
		ctx.Target.StateManager.ModifyHP(-dmgValue, false, ctx.Target.IsCharacter())

		// Update Stats
		ed.Statistics.AddDamage(a.InstanceID, dmgValue)
		ed.Adjudicator.UpdateHealTargetsAfterDamage(ctx.Target)

		ed.LogEvent(events.EventFeatureTrigger, a, map[string]interface{}{
			"feature": f.Name,
			"hook":    hook,
			"level":   selectedLevel,
			"damage":  dmgValue,
		})
	}

	return nil
}

func (ed *EncounterDirector) HandleEvasion(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.SaveContext == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for evasion")
	}

	if hook != core.HookOnSelfSavingThrow {
		return fmt.Errorf("invalid hook for evasion: %s", hook)
	}

	if !ctx.SaveContext.IsPostRoll {
		return nil
	}

	if !ctx.AttackContext.Action.HasDC ||
		ctx.AttackContext.Action.DCAbility != core.AbilityDexterity ||
		(ctx.AttackContext.Action.DCOnSuccess != core.DCOnSuccessHalf && ctx.AttackContext.Action.DCOnSuccess != core.DCOnSuccessNone) {
		return nil // Evasion doesn't apply to this action
	}

	// Apply Evasion logic to damage calculation if we are in the damage context
	if ctx.DamageContext != nil && ctx.DamageContext.DamageValue != nil {
		if ctx.SaveContext.SaveSuccess {
			// On success, take no damage
			*ctx.DamageContext.DamageValue = 0
		} else {
			// On failure, take only half damage.
			*ctx.DamageContext.DamageValue /= 2
		}
		// Since we've handled the damage mitigation for the save here,
		// we mark it as "Other" so the Adjudicator doesn't apply it again.
		ctx.AttackContext.Action.DCOnSuccess = core.DCOnSuccessOther
	}

	return nil
}

func (ed *EncounterDirector) HandleAura(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if err := ed.validateFeatureDice(f); err != nil {
		return err
	}
	if err := ed.validateFeatureDamage(f); err != nil {
		return err
	}

	if hook == core.HookOnTurnStart {
		targetIDs := ed.Adjudicator.IdentifyAOETargets(a, 0, false)

		fAction := core.Action{
			Name:       string(f.Name),
			ActionType: core.ATFeature,
			DiceBlock: []core.DiceBlock{
				{
					NumberOfDice: f.Data.NumberOfDice,
					Die:          f.Data.Die,
					DamageType:   f.Data.DamageType[0],
				},
			},
		}

		for _, id := range targetIDs {
			target := ed.Actors[id]
			err := ed.Adjudicator.resolveDamage(a, target, &fAction, false, false)
			if err != nil {
				return err
			}
		}
	}

	if hook == core.HookOnSelfHit {
		return ed.HandleMeleeTouchDamage(a, f, hook, ctx)
	}

	return nil
}

func (ed *EncounterDirector) HandleCunning(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.SaveContext == nil ||
		ctx.SaveContext.Options == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for cunning")
	}

	if hook != core.HookOnSelfSavingThrowAgainstMagic {
		return fmt.Errorf("invalid hook for cunning: %s", hook)
	}

	if !ctx.AttackContext.Action.HasDC {
		return fmt.Errorf("action does not require a saving throw for cunning")
	}

	if ctx.AttackContext.Action.DCAbility == core.AbilityCharisma ||
		ctx.AttackContext.Action.DCAbility == core.AbilityWisdom ||
		ctx.AttackContext.Action.DCAbility == core.AbilityIntelligence {
		ctx.SaveContext.Options.Advantage = core.RollAdvantage
	}

	return nil
}

func (ed *EncounterDirector) HandleLegendaryResistance(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.SaveContext == nil {
		return fmt.Errorf("invalid context for legendary resistance")
	}

	if hook != core.HookOnSelfSavingThrow {
		return fmt.Errorf("invalid hook for legendary resistance: %s", hook)
	}

	if !ctx.SaveContext.IsPostRoll {
		return nil
	}

	if ctx.SaveContext.SaveSuccess {
		return nil // No need to use it
	}

	// Check for uses in StateManager
	uses := a.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)]
	if uses <= 0 {
		return nil // Out of uses
	}

	ctx.SaveContext.SaveSuccess = true
	// Consume a use
	a.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)]--

	// Sync with feature data for persistence
	for i, feat := range a.Features {
		if feat.Name == f.Name {
			a.Features[i].Data.Value = a.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)]
			break
		}
	}

	return nil
}

func (ed *EncounterDirector) HandleAbsorption(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.DamageContext == nil {
		return fmt.Errorf("invalid context for absorption")
	}

	if hook != core.HookOnSelfDamageTaken {
		return nil // Absorption only triggers on damage taken
	}

	if f.Data.DamageType == nil || len(f.Data.DamageType) == 0 {
		return fmt.Errorf("invalid damage type for absorption: %s", f.Name)
	}

	// Check if the damage type matches
	matched := false
	for _, dt := range f.Data.DamageType {
		if dt == ctx.DamageContext.DamageType {
			matched = true
			break
		}
	}

	if !matched {
		return nil
	}

	// Absorb damage: negate it and mark it for healing
	absorbedValue := *ctx.DamageContext.DamageValue
	*ctx.DamageContext.DamageValue = 0
	ctx.DamageContext.IsAbsorption = true
	ctx.DamageContext.AbsorbedValue = absorbedValue

	return nil
}

func (ed *EncounterDirector) HandleLimitedMagicImmunity(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.AttackContext == nil || ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for limited magic immunity")
	}

	action := ctx.AttackContext.Action
	if action.ActionType != core.ATSpell {
		return nil
	}

	// Spells typically have ID in format "spell_id:level"
	idParts := strings.Split(string(action.ID), ":")
	if len(idParts) >= 2 {
		var spellLevel int
		_, err := fmt.Sscanf(idParts[len(idParts)-1], "%d", &spellLevel)
		if err == nil {
			if spellLevel <= f.Data.Value {
				// Immunity applies
				if hook == core.HookOnSelfSavingThrow || hook == core.HookOnSelfSavingThrowAgainstMagic {
					if ctx.SaveContext != nil {
						ctx.SaveContext.SaveSuccess = true
					}
				}

				if hook == core.HookOnSelfDamageTaken {
					if ctx.DamageContext != nil && ctx.DamageContext.DamageValue != nil {
						*ctx.DamageContext.DamageValue = 0
					}
				}
				return nil
			}
		}
	}

	// Advantage on saving throws against all other spells and magical effects
	if hook == core.HookOnSelfSavingThrowAgainstMagic {
		if ctx.SaveContext != nil && ctx.SaveContext.Options != nil {
			ctx.SaveContext.Options.AdvantageCount++
			ctx.SaveContext.Options.Advantage = ctx.SaveContext.Options.CalculateAdvantage()
		}
	}

	return nil
}

func (ed *EncounterDirector) HandleMagicResistance(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.SaveContext == nil ||
		ctx.SaveContext.Options == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for magic resistance")
	}

	if hook != core.HookOnSelfSavingThrowAgainstMagic {
		return fmt.Errorf("invalid hook for magic resistance: %s", hook)
	}

	if ctx.AttackContext.Action.ActionType == core.ATSpell {
		ctx.SaveContext.Options.AdvantageCount++
		ctx.SaveContext.Options.Advantage = ctx.SaveContext.Options.CalculateAdvantage()
	}

	return nil
}

func (ed *EncounterDirector) HandleMagicWeapons(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil ||
		ctx.AttackContext.WeaponModifiers == nil {
		return fmt.Errorf("invalid context for magic weapons")
	}

	if hook != core.HookOnSelfAttack {
		return fmt.Errorf("invalid hook for magic weapons: %s", hook)
	}

	if ctx.AttackContext.Action.IsWeaponAttack() {
		ctx.AttackContext.WeaponModifiers.IsMagic = true
	}

	return nil
}

func (ed *EncounterDirector) HandleMartialAdvantage(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.Target == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for martial advantage")
	}

	if err := ed.validateFeatureDice(f); err != nil {
		return err
	}

	if hook != core.HookOnSelfHit {
		return nil
	}

	// Must be a weapon attack and only used once per turn
	if !ctx.AttackContext.Action.IsWeaponAttack() ||
		a.StateManager.OncePerTurnUsed[string(f.Name)] {
		return nil
	}

	// Requires a nearby ally; no distance, does ally exist
	allies := ed.GetAllyTargets(a)
	foundValidAlly := false
	for _, ally := range allies {
		if ally.InstanceID == a.InstanceID {
			continue
		}
		if ally.StateManager.CanActConditions() {
			foundValidAlly = true
			break
		}
	}

	// No allies; can't apply feature
	if !foundValidAlly {
		return nil
	}

	// Use the same damage type as the first block of the action
	dmgType := core.DamageNone
	if len(ctx.AttackContext.Action.DiceBlock) > 0 {
		dmgType = ctx.AttackContext.Action.DiceBlock[0].DamageType
	}

	fAction := core.Action{
		Name:       string(f.Name),
		ActionType: core.ATFeature,
		DiceBlock: []core.DiceBlock{
			{
				NumberOfDice: f.Data.NumberOfDice,
				Die:          f.Data.Die,
				DamageType:   dmgType,
			},
		},
	}

	// Trigger feature
	ed.LogEvent(events.EventFeatureTrigger, a, map[string]interface{}{
		"feature": f.Name,
		"hook":    hook,
	})

	err := ed.Adjudicator.resolveDamage(a, ctx.Target, &fAction, false, *ctx.AttackContext.IsCritical)
	if err != nil {
		return err
	}

	a.StateManager.OncePerTurnUsed[string(f.Name)] = true

	return nil
}

func (ed *EncounterDirector) HandlePackTactics(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil {
		return fmt.Errorf("invalid context for pack tactics")
	}

	if hook != core.HookOnSelfAttack {
		return nil
	}

	// Requires a nearby ally; no distance, does ally exist
	allies := ed.GetAllyTargets(a)
	foundValidAlly := false
	for _, ally := range allies {
		if ally.InstanceID == a.InstanceID {
			continue
		}
		if ally.StateManager.CanActConditions() {
			foundValidAlly = true
			break
		}
	}

	// No allies; can't apply feature
	if !foundValidAlly {
		return nil
	}

	if ctx.AttackContext.AttackRoll != nil {
		ctx.AttackContext.AttackRoll.AdvantageCount++
		ctx.AttackContext.AttackRoll.Advantage = ctx.AttackContext.AttackRoll.CalculateAdvantage()
	}

	return nil
}

func (ed *EncounterDirector) HandleReckless(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if hook != core.HookOnTurnStart {
		return nil
	}

	a.StateManager.Conditions.Add(core.ConditionReckless)

	return nil
}

// Reflective carapace

func (ed *EncounterDirector) HandleRegeneration(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if hook != core.HookOnTurnStart {
		return fmt.Errorf("invalid hook for regeneration: %s", hook)
	}

	if err := ed.validateFeatureValue(f); err != nil {
		return err
	}

	if a.StateManager.CurrentHP >= a.StateManager.MaxHP {
		return nil
	}

	regenAmt := f.Data.Value
	hpRes := a.StateManager.ModifyHP(regenAmt, false, a.IsCharacter())
	ed.LogEvent(events.EventHPModified, a, map[string]interface{}{
		"result": hpRes,
		"note":   "Regeneration",
	})

	// Handle restoration of consciousness if it was at 0 HP
	if a.StateManager.CurrentHP > 0 && a.StateManager.Conditions.Has(core.ConditionUnconscious) {
		a.StateManager.Conditions.Remove(core.ConditionUnconscious)
		a.StateManager.Conditions.Remove(core.ConditionProne)
		a.StateManager.Conditions.Remove(core.ConditionStable)
		if ed.Statistics != nil {
			ed.Statistics.ResetDeathSaveStats(a.InstanceID)
		}
	}

	return nil
}

func (ed *EncounterDirector) HandleRelentless(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.DamageContext == nil || ctx.DamageContext.DamageValue == nil {
		return fmt.Errorf("invalid context for relentless")
	}

	if hook != core.HookOnSelfDamageTaken {
		return nil
	}

	damage := *ctx.DamageContext.DamageValue
	if damage <= 0 {
		return nil
	}

	// If the damage would reduce the actor to 0 HP
	if a.StateManager.CurrentHP-damage <= 0 {
		if f.Name == core.SpecAbilityRelentlessEndurance {
			// Once per combat (day) - we use OncePerTurnUsed but it's not reset for this feature
			if a.StateManager.OncePerTurnUsed[string(f.Name)] {
				return nil
			}
			// Reduce to 1 hit point instead
			*ctx.DamageContext.DamageValue = a.StateManager.CurrentHP - 1
			a.StateManager.OncePerTurnUsed[string(f.Name)] = true
		} else {
			// Standard Relentless has a threshold
			value := 0
			if err := ed.validateFeatureValue(f); err == nil {
				value = f.Data.Value
			}

			if damage <= value {
				// Reduce to 1 hit point instead
				*ctx.DamageContext.DamageValue = a.StateManager.CurrentHP - 1
			}
		}
	}

	return nil
}

func (ed *EncounterDirector) HandleSneakAttack(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.Target == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil ||
		ctx.AttackContext.AttackRoll == nil {
		return fmt.Errorf("invalid context for sneak attack")
	}

	if err := ed.validateFeatureDice(f); err != nil {
		return err
	}

	if hook != core.HookOnSelfHit {
		return nil
	}

	// Must be a weapon attack and only used once per turn
	if !ctx.AttackContext.Action.IsWeaponAttack() ||
		a.StateManager.OncePerTurnUsed[string(f.Name)] {
		return nil
	}

	hasAdvantage := ctx.AttackContext.AttackRoll.Advantage == core.RollAdvantage
	hasDisadvantage := ctx.AttackContext.AttackRoll.Advantage == core.RollDisadvantage

	// Can't sneak attack with disadvantage
	if hasDisadvantage {
		return nil
	}

	// If we don't have advantage, but an ally is nearby and we don't have disadvantage
	foundValidAlly := false
	if !hasDisadvantage {
		allies := ed.GetAllyTargets(a)
		for _, ally := range allies {
			if ally.InstanceID == a.InstanceID {
				continue
			}
			if ally.StateManager.CanActConditions() {
				foundValidAlly = true
				break
			}
		}
	}

	if !hasAdvantage && !foundValidAlly {
		return nil
	}

	// Use the same damage type as the first block of the action
	dmgType := core.DamageNone
	if len(ctx.AttackContext.Action.DiceBlock) > 0 {
		dmgType = ctx.AttackContext.Action.DiceBlock[0].DamageType
	}

	fAction := core.Action{
		Name:       string(f.Name),
		ActionType: core.ATFeature,
		DiceBlock: []core.DiceBlock{
			{
				NumberOfDice: f.Data.NumberOfDice,
				Die:          f.Data.Die,
				DamageType:   dmgType,
			},
		},
	}

	// Trigger feature
	ed.LogEvent(events.EventFeatureTrigger, a, map[string]interface{}{
		"feature": f.Name,
		"hook":    hook,
	})

	err := ed.Adjudicator.resolveDamage(a, ctx.Target, &fAction, false, *ctx.AttackContext.IsCritical)
	if err != nil {
		return err
	}

	a.StateManager.OncePerTurnUsed[string(f.Name)] = true

	return nil
}

func (ed *EncounterDirector) HandleUndeadFortitude(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.DamageContext == nil || ctx.DamageContext.DamageValue == nil || ctx.AttackContext == nil {
		return fmt.Errorf("invalid context for undead fortitude")
	}

	if hook != core.HookOnSelfDamageTaken {
		return nil
	}

	damage := *ctx.DamageContext.DamageValue
	if damage <= 0 {
		return nil
	}

	// Rule: If damage reduces the actor to 0 hit points
	if a.StateManager.CurrentHP-damage > 0 {
		return nil
	}

	// Rule: unless the damage is of a specified type (usually Radiant) or from a critical hit.
	// Check Damage Type
	if f.Data.DamageType != nil && len(f.Data.DamageType) > 0 {
		for _, dt := range f.Data.DamageType {
			if dt == ctx.DamageContext.DamageType {
				return nil
			}
		}
	}

	// Check Critical Hit
	if ctx.AttackContext.IsCritical != nil && *ctx.AttackContext.IsCritical {
		return nil
	}

	// Rule: it must make a Constitution saving throw with a DC of 5 + the damage taken.
	saveAction := core.Action{
		Name:      string(f.Name),
		HasDC:     true,
		DCSaveDC:  5 + damage,
		DCAbility: core.AbilityConstitution,
	}

	saveSuccess, _ := ed.Adjudicator.ResolveSavingThrow(a, a, &saveAction)

	// Rule: On a success, the actor drops to 1 hit point instead.
	if saveSuccess {
		// damage = currentHP - 1
		*ctx.DamageContext.DamageValue = a.StateManager.CurrentHP - 1
	}

	return nil
}

func (ed *EncounterDirector) HandleSaveAdvantage(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.SaveContext == nil ||
		ctx.SaveContext.Options == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for save advantage feature %s", f.Name)
	}

	if hook != core.HookOnSelfSavingThrow {
		return nil
	}

	if !ctx.AttackContext.Action.HasDC {
		return nil
	}

	shouldGrant := false
	switch f.Name {
	case core.SpecAbilityRageStrengthSave:
		if a.StateManager.IsRaging {
			shouldGrant = ctx.AttackContext.Action.DCAbility == core.AbilityStrength
		}
	case core.SpecAbilityDangerSense:
		shouldGrant = ctx.AttackContext.Action.DCAbility == core.AbilityDexterity
	case core.SpecAbilitySlipperyMind:
		shouldGrant = ctx.AttackContext.Action.DCAbility == core.AbilityWisdom
	case core.SpecAbilityDwarvenResilience:
		if ctx.AttackContext.Action.DiceBlock[0].DamageType == core.DamagePoison {
			shouldGrant = true
		}
	}

	if shouldGrant {
		ctx.SaveContext.Options.AdvantageCount++
		ctx.SaveContext.Options.Advantage = ctx.SaveContext.Options.CalculateAdvantage()
	}

	return nil
}

func (ed *EncounterDirector) HandleIndomitable(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.SaveContext == nil || ctx.SaveContext.Options == nil {
		return fmt.Errorf("invalid context for indomitable")
	}

	if hook != core.HookOnSelfSavingThrow {
		return nil
	}

	if !ctx.SaveContext.IsPostRoll {
		return nil
	}

	// Rule: You can reroll saving throws if you fail.
	if ctx.SaveContext.SaveSuccess {
		return nil
	}

	// Check for uses in StateManager
	uses := a.StateManager.Resource[string(core.SpecAbilityIndomitable)]
	if uses <= 0 {
		return nil
	}

	// Reroll the saving throw
	// We need to perform the roll again and update SaveSuccess in the context
	rollRes := ed.RollManager.RollD20(*ctx.SaveContext.Options)
	if rollRes == nil {
		return fmt.Errorf("failed to reroll saving throw for indomitable")
	}

	if rollRes.Total >= ctx.AttackContext.Action.DCSaveDC {
		ctx.SaveContext.SaveSuccess = true
	}

	// Consume a use
	a.StateManager.Resource[string(core.SpecAbilityIndomitable)]--

	// Sync with feature data for persistence if needed
	for i, feat := range a.Features {
		if feat.Name == f.Name {
			a.Features[i].Data.Value = a.StateManager.Resource[string(core.SpecAbilityIndomitable)]
			break
		}
	}

	ed.LogEvent(events.EventFeatureTrigger, a, map[string]interface{}{
		"feature": f.Name,
		"hook":    hook,
		"success": ctx.SaveContext.SaveSuccess,
	})

	return nil
}

func (ed *EncounterDirector) HandleLayOnHands(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if hook != core.HookOnTurnStart {
		return nil
	}

	// Requires an action
	if a.StateManager.ActionUsedCount > 0 {
		return nil
	}

	// Check pool
	pool := a.StateManager.Resource[string(core.SpecAbilityLayOnHands)]
	if pool <= 0 {
		return nil
	}

	// Find target that needs healing
	allies := ed.GetAllyTargets(a)
	var target *actor.Actor
	maxMissing := 0

	for _, ally := range allies {
		if ally.StateManager.GetHealthState(ally.IsCharacter()) == core.HealthStateDead {
			continue
		}
		missing := ally.StateManager.MaxHP - ally.StateManager.CurrentHP
		if missing > maxMissing {
			maxMissing = missing
			target = ally
		}
	}

	if target == nil || maxMissing == 0 {
		return nil
	}

	// Determine how much to heal
	healAmt := maxMissing
	if healAmt > pool {
		healAmt = pool
	}

	// Create a temporary action to use the centralized healing resolution
	fAction := core.Action{
		Name:        string(f.Name),
		ActionType:  core.ATFeature,
		AverageHeal: healAmt,
	}

	err := ed.Adjudicator.executeHealing(a, target.InstanceID, &fAction, ed)
	if err != nil {
		return err
	}

	// Consume from pool
	a.StateManager.Resource[string(core.SpecAbilityLayOnHands)] -= healAmt
	a.StateManager.ActionUsedCount++

	ed.LogEvent(events.EventFeatureTrigger, a, map[string]interface{}{
		"feature": f.Name,
		"target":  target.Name,
		"heal":    healAmt,
	})

	return nil
}

func (ed *EncounterDirector) HandleImprovedDivineSmite(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.DamageContext == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil ||
		ctx.AttackContext.IsCritical == nil {
		return fmt.Errorf("invalid context for improved divine smite")
	}

	if err := ed.validateFeatureDice(f); err != nil {
		return err
	}

	if hook != core.HookOnSelfHit {
		return nil
	}

	if !ctx.AttackContext.Action.IsMeleeAttack() {
		return nil
	}

	fAction := core.Action{
		Name:       string(f.Name),
		ActionType: core.ATFeature,
		DiceBlock: []core.DiceBlock{
			{
				NumberOfDice: f.Data.NumberOfDice,
				Die:          f.Data.Die,
				DamageType:   ctx.DamageContext.DamageType,
			},
		},
	}

	// Trigger feature
	ed.LogEvent(events.EventFeatureTrigger, a, map[string]interface{}{
		"feature": f.Name,
		"hook":    hook,
	})

	err := ed.Adjudicator.resolveDamage(a, ctx.Target, &fAction, false, *ctx.AttackContext.IsCritical)
	if err != nil {
		return err
	}

	return nil
}

func (ed *EncounterDirector) HandleDeflectMissiles(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.DamageContext == nil ||
		ctx.DamageContext.DamageValue == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for deflect missiles")
	}

	if err := ed.validateFeatureDice(f); err != nil {
		return err
	}

	if hook != core.HookOnSelfDamageTaken {
		return nil
	}

	// Rule: When you are hit with a ranged attack
	if ctx.AttackContext.Action.ActionType != core.ATRanged {
		return nil
	}

	// Requires a reaction
	if a.StateManager.ReactionUsedCount > 0 {
		return nil
	}

	// Rule: the damage is reduced by 1d10 + Dexterity modifier + Monk level.
	opts := roll_manager.RollOptions{
		RollType: core.DiceRollFeature,
	}

	reductionRoll := ed.RollManager.RollDice(f.Data.NumberOfDice, f.Data.Die, opts)
	dexMod := a.Abilities.GetAbilityModifier(core.AbilityDexterity)

	// No multiclassing yet; use actor level/CR
	level := a.Metadata.Level
	if level == 0 && a.Metadata.CR != 0 {
		level = int(a.Metadata.CR)
	}

	reduction := reductionRoll.Total + dexMod + level

	currentDamage := *ctx.DamageContext.DamageValue
	if reduction > currentDamage {
		reduction = currentDamage
	}

	*ctx.DamageContext.DamageValue -= reduction
	a.StateManager.ReactionUsedCount++

	ed.LogEvent(events.EventFeatureTrigger, a, map[string]interface{}{
		"feature":   f.Name,
		"reduction": reduction,
	})

	return nil
}

func (ed *EncounterDirector) HandleSecondWind(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if err := ed.validateFeatureDice(f); err != nil {
		return err
	}

	if hook != core.HookOnTurnStart {
		return nil
	}

	// Only use Second Wind if wounded (below 80% HP)
	if float64(a.StateManager.CurrentHP) >= float64(a.StateManager.MaxHP)*0.8 {
		return nil
	}

	// Check for resource
	if a.StateManager.Resource[string(core.SpecAbilitySecondWind)] <= 0 {
		return nil
	}

	// Requires a bonus action
	if a.StateManager.BonusActionUsedCount > 0 {
		return nil
	}

	// No multiclassing yet; use actor level/CR
	level := a.Metadata.Level
	if level == 0 && a.Metadata.CR != 0 {
		level = int(a.Metadata.CR)
	}

	// Create a temporary action to use the centralized healing resolution
	fAction := core.Action{
		Name:       string(f.Name),
		ActionType: core.ATFeature,
		DiceBlock: []core.DiceBlock{
			{
				NumberOfDice: f.Data.NumberOfDice,
				Die:          f.Data.Die,
				Modifier:     level,
			},
		},
	}

	err := ed.Adjudicator.executeHealing(a, a.InstanceID, &fAction, ed)
	if err != nil {
		return err
	}

	a.StateManager.BonusActionUsedCount++
	a.StateManager.Resource[string(core.SpecAbilitySecondWind)]--

	ed.LogEvent(events.EventFeatureTrigger, a, map[string]interface{}{
		"feature": f.Name,
	})

	return nil
}

func (ed *EncounterDirector) HandleFightingStyle(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.AttackContext == nil || ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for fighting style %s", f.Name)
	}

	props := ctx.AttackContext.WeaponProperties
	if props == nil {
		return nil // Not a weapon attack with properties
	}

	switch f.Name {
	case core.SpecAbilityFightingStyleArchery:
		if hook == core.HookOnSelfAttack {
			if props.IsRanged {
				if ctx.AttackContext.AttackRoll != nil {
					ctx.AttackContext.AttackRoll.Modifier += 2
				}
			}
		}

	case core.SpecAbilityFightingStyleDuel:
		if hook == core.HookOnSelfHit || hook == core.HookOnSelfAttack {
			// Rule: wielding a melee weapon in one hand and no other weapons.
			// This typically means primary is a melee weapon, and secondary is empty (or a shield).
			if !props.IsRanged && !props.IsTwoHanded {
				// Check secondary slot
				secondary := a.Equipment.GetItem("secondary")
				if secondary == nil || secondary.Type == "shield" {
					if ctx.AttackContext.WeaponModifiers != nil {
						ctx.AttackContext.WeaponModifiers.DamageBonus += 2
					}
				}
			}
		}

	case core.SpecAbilityFightingStyleGWF:
		if hook == core.HookOnSelfAttack {
			// Rule: wielding a melee weapon with two hands
			if !props.IsRanged && props.IsTwoHanded {
				if ctx.AttackContext.DamageRoll != nil {
					ctx.AttackContext.DamageRoll.RerollThreshold = 2
				}
			}
		}

	case core.SpecAbilityFightingStyleTWF:
		if hook == core.HookOnSelfAttack || hook == core.HookOnSelfHit {
			// Rule: wielding a weapon in each hand, you add your ability modifier to the damage of the offhand attack.
			if ctx.AttackContext.Action.ActionType == core.ATOffhand {
				// In our engine, ATOffhand attacks might already have the modifier removed if it's following standard rules,
				// or we need to ensure it's added here.
				// Based on common implementation, we might need to find the correct modifier.
				if ctx.AttackContext.WeaponModifiers != nil {
					ability := core.AbilityStrength
					if props.IsFinesse {
						str := a.Abilities.GetScore(core.AbilityStrength)
						dex := a.Abilities.GetScore(core.AbilityDexterity)
						if dex > str {
							ability = core.AbilityDexterity
						}
					}
					ctx.AttackContext.WeaponModifiers.DamageBonus += a.Abilities.GetAbilityModifier(ability)
				}
			}
		}
	}

	return nil
}

func (ed *EncounterDirector) HandleHalflingLucky(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil {
		return fmt.Errorf("invalid context for halfling lucky")
	}

	switch hook {
	case core.HookOnSelfAttack:
		if ctx.AttackContext == nil || ctx.AttackContext.AttackRoll == nil {
			return fmt.Errorf("invalid context for halfling lucky")
		}
		ctx.AttackContext.AttackRoll.RerollThreshold = 1
	case core.HookOnSelfSavingThrow:
		if ctx.SaveContext == nil || ctx.SaveContext.Options == nil {
			return fmt.Errorf("invalid context for halfling lucky")
		}
		ctx.SaveContext.Options.RerollThreshold = 1
	}

	return nil
}

func (ed *EncounterDirector) HandleSavageAttacks(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.Target == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil ||
		ctx.AttackContext.IsCritical == nil {
		return fmt.Errorf("invalid context for savage attacks")
	}

	if hook != core.HookOnSelfHit {
		return nil
	}

	// Rule: When you score a critical hit with a melee weapon attack
	if !*ctx.AttackContext.IsCritical {
		return nil
	}

	if ctx.AttackContext.Action.ActionType != core.ATMelee {
		return nil
	}

	// Identify the weapon's damage die.
	// Savage Attacks: "you can roll one of the weapon’s damage dice one additional time"
	if len(ctx.AttackContext.Action.DiceBlock) == 0 {
		return nil
	}

	// We'll take the first die from the first dice block as the "weapon's damage die"
	die := ctx.AttackContext.Action.DiceBlock[0].Die
	dmgType := ctx.AttackContext.Action.DiceBlock[0].DamageType

	fAction := core.Action{
		Name:       string(f.Name),
		ActionType: core.ATFeature,
		DiceBlock: []core.DiceBlock{
			{
				NumberOfDice: 1,
				Die:          die,
				DamageType:   dmgType,
			},
		},
	}

	// This additional die is NOT doubled by the critical hit itself (it's "extra damage of the critical hit")
	err := ed.Adjudicator.resolveDamage(a, ctx.Target, &fAction, false, false)
	if err != nil {
		return err
	}

	return nil
}

func (ed *EncounterDirector) HandleBrutalCritical(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil ||
		ctx.Target == nil ||
		ctx.AttackContext == nil ||
		ctx.AttackContext.Action == nil ||
		ctx.AttackContext.IsCritical == nil {
		return fmt.Errorf("invalid context for brutal critical")
	}

	if hook != core.HookOnSelfHit {
		return nil
	}

	// Rule: When you score a critical hit with a melee attack
	if !*ctx.AttackContext.IsCritical {
		return nil
	}

	// Note: Brutal Critical specifies "melee attack"
	if !ctx.AttackContext.Action.IsWeaponAttack() || ctx.AttackContext.Action.ActionType != core.ATMelee {
		return nil
	}

	if err := ed.validateFeatureDice(f); err != nil {
		return err
	}

	// Identify the weapon's damage die.
	if len(ctx.AttackContext.Action.DiceBlock) == 0 {
		return nil
	}

	// Use the first die from the first dice block
	die := ctx.AttackContext.Action.DiceBlock[0].Die
	dmgType := ctx.AttackContext.Action.DiceBlock[0].DamageType

	fAction := core.Action{
		Name:       string(f.Name),
		ActionType: core.ATFeature,
		DiceBlock: []core.DiceBlock{
			{
				NumberOfDice: f.Data.NumberOfDice,
				Die:          die,
				DamageType:   dmgType,
			},
		},
	}

	// This additional damage is NOT doubled by the critical hit itself
	err := ed.Adjudicator.resolveDamage(a, ctx.Target, &fAction, false, false)
	if err != nil {
		return err
	}

	return nil
}

func (ed *EncounterDirector) HandleRelentlessRage(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.DamageContext == nil || ctx.DamageContext.DamageValue == nil {
		return fmt.Errorf("invalid context for relentless rage")
	}

	if hook != core.HookOnSelfDamageTaken {
		return nil
	}

	// Only while raging
	if !a.StateManager.IsRaging {
		return nil
	}

	damage := *ctx.DamageContext.DamageValue
	if damage <= 0 {
		return nil
	}

	// If damage reduces the actor to 0 hp
	if a.StateManager.CurrentHP-damage <= 0 {
		// DC starts at 10 and increases by 5 for every use
		uses := a.StateManager.Resource[string(core.SpecAbilityRelentlessRage)]
		dc := 10 + (uses * 5)

		saveAction := core.Action{
			Name:      string(f.Name),
			HasDC:     true,
			DCSaveDC:  dc,
			DCAbility: core.AbilityConstitution,
		}

		saveSuccess, _ := ed.Adjudicator.ResolveSavingThrow(a, a, &saveAction)

		if saveSuccess {
			// If you succeed, you drop to 1 hit point instead.
			*ctx.DamageContext.DamageValue = a.StateManager.CurrentHP - 1
			a.StateManager.Resource[string(core.SpecAbilityRelentlessRage)]++
		}
	}

	return nil
}

func (ed *EncounterDirector) HandleRageExtraDamage(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.AttackContext == nil || ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for rage extra damage")
	}

	if hook != core.HookOnSelfAttack {
		return nil
	}

	// Only while raging
	if !a.StateManager.IsRaging {
		return nil
	}

	// Rule: melee weapon attack using Strength
	if ctx.AttackContext.Action.ActionType == core.ATMelee && ctx.AttackContext.Action.IsWeaponAttack() {
		if ctx.AttackContext.WeaponModifiers != nil {
			ctx.AttackContext.WeaponModifiers.DamageBonus += f.Data.Modifier
		}
	}

	return nil
}

func (ed *EncounterDirector) HandleUncannyDodge(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.DamageContext == nil || ctx.DamageContext.DamageValue == nil || ctx.AttackContext == nil || ctx.AttackContext.Action == nil {
		return fmt.Errorf("invalid context for uncanny dodge")
	}

	if hook != core.HookOnSelfDamageTaken {
		return nil
	}

	// Rule: When an attacker that you can see hits you with an attack
	// In this simulation, we assume you can see the attacker.
	if !ctx.AttackContext.Action.IsWeaponAttack() && ctx.AttackContext.Action.ActionType != core.ATSpell {
		// D&D 5e: "Uncanny Dodge ... when an attacker that you can see hits you with an attack"
		// This includes spell attacks if they hit.
		// However, it usually doesn't include saving throws (that's Evasion).
		return nil
	}

	// If it's a saving throw based action, Uncanny Dodge doesn't apply
	if ctx.AttackContext.Action.HasDC {
		return nil
	}

	// Requires a reaction
	if a.StateManager.ReactionUsedCount > 0 {
		return nil
	}

	// Rule: halve the attack's damage against you
	damage := *ctx.DamageContext.DamageValue
	if damage > 0 {
		*ctx.DamageContext.DamageValue = damage / 2
		a.StateManager.ReactionUsedCount++
	}

	return nil
}

func (ed *EncounterDirector) HandleElusive(a *actor.Actor, f core.Feature, hook core.HookType, ctx *FeatureContext) error {
	if ctx == nil || ctx.AttackContext == nil || ctx.AttackContext.AttackRoll == nil {
		return fmt.Errorf("invalid context for elusive")
	}

	if hook != core.HookOnSelfAttack {
		return nil
	}

	// Rule: No attack roll has advantage against you while you aren't incapacitated.
	if !a.StateManager.CanActConditions() {
		return nil
	}

	if ctx.AttackContext.AttackRoll.Advantage == core.RollAdvantage {
		ctx.AttackContext.AttackRoll.Advantage = core.RollNormal
	}

	return nil
}
