package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/events"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"fmt"
)

func (m *Monster) RollInitiative() (int, error) {
	var err error
	opts := roll_manager.NewRollOptions()
	opts.Modifier, err = m.GetAbilityScoreModifier(core.AbilityDexterity)
	if err != nil {
		return 0, err
	}

	res, err := m.RollManager.RollInitiative(opts)
	if err != nil {
		return 0, err
	}

	m.EntityStateManager.SetInitiative(res.Total)

	return res.Total, nil
}

func (m *Monster) createAttackRequest(target core.Entity, actionIndex int, actionType core.ActionType, simulationOptions *core.SimulationOptions) (*core.AttackRequest, error) {
	if !isValidMonsterActionType(actionType) {
		return nil, fmt.Errorf("invalid action type for monster attack request")
	}

	if simulationOptions == nil {
		// Defensive default to avoid nil dereference; callers should normally pass non-nil
		simulationOptions = &core.SimulationOptions{}
	}

	// Determine ranged vs melee for condition rules from the first attack data
	adList := m.ActionManager.GetAttackDataFromIndex(actionIndex, actionType)
	if len(adList) == 0 {
		return nil, fmt.Errorf("no attack data found for action index %d and type %v", actionIndex, actionType)
	}
	isRanged := adList[0].IsRangedWeapon
	// Compute final advantage using unified core helper
	advSlice := make([]core.AdvantageType, 0)
	computedAdv := core.DetermineAttackAdvantageForEntities(m, target, isRanged, core.RollNormal)
	advSlice = append(advSlice, computedAdv)
	if m.hasPackTacticsAdvantage() {
		advSlice = append(advSlice, core.RollAdvantage)
	}
	if m.hasBloodFrenzyAdvantage(target) {
		advSlice = append(advSlice, core.RollAdvantage)
	}
	if m.hasRecklessAdvantage() {
		advSlice = append(advSlice, core.RollAdvantage)
	}
	if m.hasAssassinateAdvantage(target) {
		advSlice = append(advSlice, core.RollAdvantage)
	}

	adv := core.GetFinalAdvantageType(advSlice)

	attackOptions := core.AttackOptions{
		Advantage:            adv,
		ShouldApplyDamageMod: true,
		ImprovedCritical:     simulationOptions.UseImprovedCriticals,
	}

	return &core.AttackRequest{
		AttackData:        m.ActionManager.GetAttackDataFromIndex(actionIndex, actionType),
		AttackOptions:     attackOptions,
		SimulationOptions: simulationOptions,
		Target:            target,
	}, nil
}

func (m *Monster) createSpellAttackData(spellChoice core.SpellChoice) (spellcasting_manager.SpellCastData, error) {
	spellBonus := m.SpellCastingManager.GetSpellcastModifierValue()
	return spellcasting_manager.SpellCastData{
		SpellChoice:          spellChoice,
		AttackModifier:       spellBonus,
		SpellcastingModifier: m.SpellCastingManager.GetSpellcastModifierValue(),
	}, nil
}

func (m *Monster) createSpellCastRequest(target core.Entity, spellchoice core.SpellChoice, simOptions *core.SimulationOptions) (*spellcasting_manager.SpellCastRequest, error) {
	spellcastData, err := m.createSpellAttackData(spellchoice)
	if err != nil {
		return nil, err
	}

	// Compute final advantage using unified core helper
	advSlice := make([]core.AdvantageType, 0)
	computedAdv := core.DetermineAttackAdvantageForEntities(m, target, !spellcastData.SpellChoice.GetSpell().GetIsTouch(), core.RollNormal)
	advSlice = append(advSlice, computedAdv)
	if m.hasPackTacticsAdvantage() {
		advSlice = append(advSlice, core.RollAdvantage)
	}
	if m.hasAssassinateAdvantage(target) {
		advSlice = append(advSlice, core.RollAdvantage)
	}
	adv := core.GetFinalAdvantageType(advSlice)

	options := spellcasting_manager.SpellOptions{
		Advantage:            adv,
		BonusToAttackRoll:    0,     // future: features/auras
		BonusToDamageRoll:    0,     // future: features/auras
		ShouldApplyDamageMod: false, // RollSpellValue already handles spell modifiers
		ImprovedCritical:     spellcastData.SpellChoice.Spell.GetSpellType() == core.STDamage && (simOptions != nil && simOptions.UseImprovedCriticals),
		TreatOnesAsTwos:      false, // future: features
	}

	return &spellcasting_manager.SpellCastRequest{
		SpellCastData:     spellcastData,
		SpellOptions:      options,
		SimulationOptions: simOptions,
		Target:            target,
	}, nil
}

func (m *Monster) CreateHealRequest(target core.Entity) (*core.HealRequest, error) {
	// 1. Determine how much healing the target actually needs
	hpNeeded := target.GetHPStatus().GetHPDifference()
	if hpNeeded <= 0 {
		return nil, fmt.Errorf("target does not require healing")
	}

	// 2. Monsters only heal via spells for now
	if m.SpellCastingManager.HasHealingSpells() {
		choice, err := m.SpellCastingManager.GetMostEfficientHealingSpell(hpNeeded)
		if err == nil && choice != nil {
			return &core.HealRequest{
				Source:      core.HealSourceSpell,
				Target:      target,
				SpellChoice: choice,
			}, nil
		}
	}

	return nil, fmt.Errorf("no healing resources available")
}

func isValidMonsterActionType(actionType core.ActionType) bool {
	return actionType == core.ATMonsterAction ||
		actionType == core.ATLegendaryAction ||
		actionType == core.ATMonsterSpecial ||
		actionType == core.ATMonsterMultiattack
}

// MakeSavingThrow calculates a saving throw for the given ability and returns the total roll result or an error.
func (m *Monster) MakeSavingThrow(ability core.Ability, targetValue int, isSpell bool, damageType core.DamageType, simOptions *core.SimulationOptions) (core.RollResult, error) {
	activeConditions := m.GetConditions().GetActive()
	isStrDexSave := ability == core.AbilityStrength || ability == core.AbilityDexterity

	autoFailResult := func() core.RollResult {
		result := roll_manager.RollResult{
			DiceRollType:   core.DiceRollSavingThrow,
			NumberOfDice:   0,
			Die:            0,
			FinalRollValue: 0,
			FinalRolls:     []int{0},
			Modifier:       0,
			Total:          0,
			Advantage:      core.RollNormal,
			IsSuccess:      false,
			TargetValue:    targetValue,
		}
		if m.EntityStateManager.GetLegendaryResistanceUses() > 0 {
			m.useLegendaryResistance(&result)
		}
		m.LogEvent(events.ETRollEvent, &result)
		return result
	}

	if isStrDexSave && len(activeConditions) > 0 {
		for _, cond := range activeConditions {
			effect := core.GetConditionEffects(cond)
			if effect.AutoFailStrDexSave {
				return autoFailResult(), nil
			}
		}
	}

	mod, err := m.GetSavingThrowBonus(ability)
	if err != nil {
		return nil, err
	}

	opts := roll_manager.NewRollOptions()
	advSlice := make([]core.AdvantageType, 0)
	baseAdv := m.EntityStateManager.GetSavingThrowAdvantage(ability) // This should get the default advantage of the character
	condAdv := core.DetermineSaveAdvantageFromConditions(m.GetConditions(), ability)
	advSlice = append(advSlice, baseAdv, condAdv)

	// Special abilities
	if simOptions != nil && simOptions.EnableSpecialAbilities {
		if m.SpecialAbilities.MagicResistance && isSpell {
			advSlice = append(advSlice, core.RollAdvantage)
		}
		if m.SpecialAbilities.GnomeCunning && isSpell &&
			(ability == core.AbilityWisdom || ability == core.AbilityCharisma || ability == core.AbilityIntelligence) {
			advSlice = append(advSlice, core.RollAdvantage)
		}
	} else if simOptions == nil {
		// Fallback for when options are not provided (e.g. tests or simple calls)
		if m.SpecialAbilities.MagicResistance && isSpell {
			advSlice = append(advSlice, core.RollAdvantage)
		}
	}

	opts.Advantage = core.GetFinalAdvantageType(advSlice)
	opts.Modifier = mod
	opts.RollType = core.DiceRollSavingThrow
	opts.TargetValue = targetValue

	res, err := m.RollManager.RollSavingThrow(opts)
	if err != nil {
		return nil, err
	}

	if !res.IsSuccess && m.EntityStateManager.GetLegendaryResistanceUses() > 0 {
		m.useLegendaryResistance(res)
	}

	return res, nil
}

func (m *Monster) GetAttackBonus() int {
	bestBonus := 0

	// Check actions from ActionManager
	if m.ActionManager != nil {
		for _, action := range m.ActionManager.Actions {
			if action.AttackBonus > bestBonus {
				bestBonus = action.AttackBonus
			}
		}
	}

	// Check spellcasting
	if m.MonsterBase.IsSpellcaster && m.SpellCastingManager != nil {
		bonus := m.SpellCastingManager.GetAttackModifier()
		if bonus > bestBonus {
			bestBonus = bonus
		}
	}

	return bestBonus
}

func (m *Monster) UpdateAICombatContext(ctx *core.CombatContext) error {
	m.AI.UpdateCombatContext(ctx)
	if m.SpellCastingManager != nil {
		m.SpellCastingManager.SetSimulationOptions(ctx.Opt())
	}
	if info, ok := ctx.CombatantInfo[m.MonsterBase.InstanceID]; ok {
		m.Info = info
	}
	return nil
}

func (m *Monster) PushEventContext(ctx *core.EventContext) {
	m.AI.UpdateEventContext(ctx)
}

func (m *Monster) GetCurrentEventContext() *core.EventContext {
	return m.AI.eventCtx
}

// useLegendaryResistance allows a Monster to succeed a failed roll by modifying the RollResult and expending a use.
func (m *Monster) useLegendaryResistance(res *roll_manager.RollResult) {
	if res.IsSuccess || m.EntityStateManager.GetLegendaryResistanceUses() <= 0 {
		return // nothing to change if success or no uses left
	}

	res.IsSuccess = true
	res.WasRerolled = true
	m.EntityStateManager.ExpendLegendaryResistanceUse()

	events.LogCombatEventMessage(m.GetCurrentEventContext(), m, fmt.Sprintf("%s uses Legendary Resistance to succeed on the saving throw!", m.GetName()), m.EventListener)
}

func (m *Monster) hasPackTacticsAdvantage() bool {
	if !m.SpecialAbilities.PackTactics {
		return false
	}

	if m.AI.GetCombatContext() == nil {
		return false
	}

	ctx := m.AI.GetCombatContext()

	return ((ctx.ConsciousMonsterCount - 1) > 0) && ctx.Opt().EnableSpecialAbilities && m.SpecialAbilities.PackTactics
}

func (m *Monster) hasBloodFrenzyAdvantage(target core.Entity) bool {
	if !m.SpecialAbilities.BloodFrenzy {
		return false
	}

	if m.AI.GetCombatContext() == nil {
		return false
	}

	ctx := m.AI.GetCombatContext()

	if target.GetHPStatus().GetHPDifference() > 0 && ctx.Opt().EnableSpecialAbilities && m.SpecialAbilities.BloodFrenzy {
		return true
	}

	return false
}

func (m *Monster) hasRecklessAdvantage() bool {
	if !m.SpecialAbilities.Reckless {
		return false
	}

	if m.AI.GetCombatContext() == nil {
		return false
	}

	ctx := m.AI.GetCombatContext()

	return ctx.Opt().EnableSpecialAbilities && m.SpecialAbilities.Reckless
}

func (m *Monster) hasAssassinateAdvantage(target core.Entity) bool {
	if !m.SpecialAbilities.Assassinate {
		return false
	}

	if m.AI.GetCombatContext() == nil {
		return false
	}

	ctx := m.AI.GetCombatContext()
	if !ctx.Opt().EnableSpecialAbilities {
		return false
	}

	// Assassinate: advantage on attack rolls against any creature that hasn't taken a turn in the combat yet.
	return !target.GetHasTakenTurnInCombat()
}

func (m *Monster) resolveMartialAdvantage(isCritical bool, simOptions *core.SimulationOptions) *core.Effect {
	if m.SpecialAbilities.MartialAdvantageNumDice <= 0 {
		return nil
	}

	if m.EntityStateManager.GetHasUsedMartialAdvantage() {
		return nil
	}

	if m.AI.GetCombatContext() == nil {
		return nil
	}

	ctx := m.AI.GetCombatContext()

	// Condition: "The hobgoblin can deal an extra 7 (2d6) damage to a creature it hits with a weapon attack
	// if that creature is within 5 feet of an ally of the hobgoblin that isn't incapacitated."
	// Simplified: Check if there's at least one other conscious monster in the combat.
	if (ctx.ConsciousMonsterCount - 1) <= 0 {
		return nil
	}

	numDice := m.SpecialAbilities.MartialAdvantageNumDice
	opts := roll_manager.NewRollOptions()
	opts.RollType = core.DiceRollDamage

	var res *roll_manager.RollResult
	var err error

	if isCritical && simOptions.UseImprovedCriticals {
		total, rolls := m.RollManager.RollExtraMaxDice(numDice, core.D6)
		res = &roll_manager.RollResult{
			DiceRollType:   core.DiceRollDamage,
			NumberOfDice:   len(rolls),
			Die:            core.D6,
			FinalRollValue: total,
			FinalRolls:     rolls,
			Total:          total,
		}
	} else {
		if isCritical {
			numDice *= 2
		}
		res, err = m.RollManager.RollDice(numDice, core.D6, opts)
	}

	if err != nil || res == nil {
		return nil
	}

	m.EntityStateManager.SetHasUsedMartialAdvantage(true)
	events.LogSpecialAbilityEvent(m.GetCurrentEventContext(), m, "Martial Advantage", fmt.Sprintf("%s deals extra damage from Martial Advantage!", m.GetName()), "", res.Total, m.EventListener)

	return &core.Effect{
		Type:       core.EffectDamage,
		Value:      res.Total,
		BaseValue:  res.Total,
		DamageType: core.DamageSlashing, // Extra damage is usually of the same type as the attack, simplified to Slashing
	}
}

func (m *Monster) resolveDivineEminence(isCritical bool, simOptions *core.SimulationOptions) *core.Effect {
	if m.SpecialAbilities.DivineEminenceNumDice <= 0 {
		return nil
	}

	if !m.EntityStateManager.GetIsDivineEminenceActive() {
		return nil
	}

	// Divine Eminence: As a bonus action, the priest can cause its melee weapon attacks to magically deal
	// an extra 10 (3d6) radiant damage to a target on a hit.
	numDice := m.EntityStateManager.GetDivineEminenceDice()
	if numDice <= 0 {
		numDice = m.SpecialAbilities.DivineEminenceNumDice
	}

	opts := roll_manager.NewRollOptions()
	opts.RollType = core.DiceRollDamage

	var res *roll_manager.RollResult
	var err error

	if isCritical && simOptions.UseImprovedCriticals {
		total, rolls := m.RollManager.RollExtraMaxDice(numDice, core.D6)
		res = &roll_manager.RollResult{
			DiceRollType:   core.DiceRollDamage,
			NumberOfDice:   len(rolls),
			Die:            core.D6,
			FinalRollValue: total,
			FinalRolls:     rolls,
			Total:          total,
		}
	} else {
		if isCritical {
			numDice *= 2
		}
		res, err = m.RollManager.RollDice(numDice, core.D6, opts)
	}

	if err != nil || res == nil {
		return nil
	}

	events.LogSpecialAbilityEvent(m.GetCurrentEventContext(), m, "Divine Eminence", fmt.Sprintf("%s deals extra radiant damage from Divine Eminence!", m.GetName()), "", res.Total, m.EventListener)

	return &core.Effect{
		Type:       core.EffectDamage,
		Value:      res.Total,
		BaseValue:  res.Total,
		DamageType: core.DamageRadiant,
	}
}

func (m *Monster) resolveSneakAttack(params core.SneakAttackParams, simOptions *core.SimulationOptions) *core.Effect {
	if m.SpecialAbilities.SneakAttackNumDice <= 0 {
		return nil
	}

	if params.IsSpell || params.IsRanged {
		return nil
	}

	if m.EntityStateManager.GetHasUsedSneakAttack() {
		return nil
	}

	// Sneak Attack: if you have advantage then perform sneak attack
	// or if within 5 feet (simulated by AlwaysUseSneakAttack flag)
	canUse := params.Advantage == core.RollAdvantage || (simOptions != nil && simOptions.AlwaysUseSneakAttack)

	if !canUse {
		return nil
	}

	numDice := m.SpecialAbilities.SneakAttackNumDice
	opts := roll_manager.NewRollOptions()
	opts.RollType = core.DiceRollDamage

	var res *roll_manager.RollResult
	var err error

	if params.IsCritical && simOptions.UseImprovedCriticals {
		total, rolls := m.RollManager.RollExtraMaxDice(numDice, core.D6)
		res = &roll_manager.RollResult{
			DiceRollType:   core.DiceRollDamage,
			NumberOfDice:   len(rolls),
			Die:            core.D6,
			FinalRollValue: total,
			FinalRolls:     rolls,
			Total:          total,
		}
	} else {
		if params.IsCritical {
			numDice *= 2
		}
		res, err = m.RollManager.RollDice(numDice, core.D6, opts)
	}

	if err != nil || res == nil {
		return nil
	}

	m.EntityStateManager.SetHasUsedSneakAttack(true)
	events.LogSpecialAbilityEvent(m.GetCurrentEventContext(), m, "Sneak Attack", fmt.Sprintf("%s deals extra damage from Sneak Attack!", m.GetName()), "", res.Total, m.EventListener)

	return &core.Effect{
		Type:       core.EffectDamage,
		Value:      res.Total,
		BaseValue:  res.Total,
		DamageType: params.DamageType,
	}
}
