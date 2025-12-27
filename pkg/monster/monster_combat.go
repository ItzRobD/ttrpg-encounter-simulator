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
	isRanged := false
	if len(adList) > 0 {
		isRanged = adList[0].IsRangedWeapon
	}
	// Compute final advantage using unified core helper
	computedAdv := core.DetermineAttackAdvantageForEntities(m, target, isRanged, core.RollNormal)

	attackOptions := core.AttackOptions{
		Advantage:            computedAdv,
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

func (m *Monster) createSpellCastRequest(target core.Entity, spellchoice core.SpellChoice, adv core.AdvantageType, simOptions *core.SimulationOptions) (*spellcasting_manager.SpellCastRequest, error) {
	spellcastData, err := m.createSpellAttackData(spellchoice)
	if err != nil {
		return nil, err
	}

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
func (m *Monster) MakeSavingThrow(ability core.Ability, targetValue int, isSpell bool, damageType core.DamageType) (core.RollResult, error) {
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
		events.LogDiceRollEvent(m, &result, m.EventListener)
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
	baseAdv := m.EntityStateManager.GetSavingThrowAdvantage(ability) // This should get the default advantage of the character
	condAdv := core.DetermineSaveAdvantageFromConditions(m.GetConditions(), ability)
	opts.Advantage = core.GetFinalAdvantageType([]core.AdvantageType{baseAdv, condAdv})
	opts.Modifier = mod
	opts.RollType = core.DiceRollSavingThrow
	opts.TargetValue = targetValue

	res, err := m.RollManager.RollSavingThrow(opts)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Monster) UpdateAICombatContext(ctx *core.CombatContext) error {
	m.AI.UpdateCombatContext(ctx)
	if m.SpellCastingManager != nil {
		m.SpellCastingManager.SetSimulationOptions(ctx.Opt())
	}
	return nil
}
