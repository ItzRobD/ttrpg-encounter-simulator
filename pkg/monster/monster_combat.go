package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
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

	m.EntityState.SetInitiative(res.Total)

	return res.Total, nil
}

func (m *Monster) createAttackRequest(target core.Entity, actionIndex int, actionType core.ActionType, adv core.AdvantageType, simulationOptions *core.SimulationOptions) (*core.AttackRequest, error) {
	if !isValidMonsterActionType(actionType) {
		return nil, fmt.Errorf("invalid action type for monster attack request")
	}

	// TODO: Handle these
	if simulationOptions == nil {
		// Defensive default to avoid nil dereference; callers should normally pass non-nil
		simulationOptions = &core.SimulationOptions{}
	}
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
		SpellcastingModifier: 0, // TODO: Will this ever be used for anything?
	}, nil
}

func (m *Monster) createSpellCastRequest(target core.Entity, spellchoice core.SpellChoice, adv core.AdvantageType, simOptions *core.SimulationOptions) (*spellcasting_manager.SpellCastRequest, error) {
	spellcastData, err := m.createSpellAttackData(spellchoice)
	if err != nil {
		return nil, err
	}

	// TODO: Handle the creation of these options dynamically
	options := spellcasting_manager.SpellOptions{
		Advantage:            adv,
		BonusToAttackRoll:    0,
		BonusToDamageRoll:    0,
		ShouldApplyDamageMod: false,
		ImprovedCritical:     false,
		TreatOnesAsTwos:      false,
	}

	return &spellcasting_manager.SpellCastRequest{
		SpellCastData:     spellcastData,
		SpellOptions:      options,
		SimulationOptions: simOptions,
		Target:            target,
	}, nil
}

func isValidMonsterActionType(actionType core.ActionType) bool {
	return actionType == core.ATMonsterAction ||
		actionType == core.ATLegendaryAction ||
		actionType == core.ATMonsterSpecial ||
		actionType == core.ATMonsterMultiattack
}

// MakeSavingThrow calculates a saving throw for the given ability and returns the total roll result or an error.
func (m *Monster) MakeSavingThrow(ability core.Ability, targetValue int) (core.RollResult, error) {
	// roll dice, add save
	mod, err := m.GetSavingThrowBonus(ability)
	if err != nil {
		return nil, err
	}

	opts := roll_manager.NewRollOptions()
	// TODO: Will monsters ever have advantage? Features apply this
	opts.Modifier = mod
	opts.RollType = core.DiceRollSavingThrow
	opts.TargetValue = targetValue

	res, err := m.RollManager.RollSavingThrow(ability, opts)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Monster) UpdateAICombatContext(ctx *core.CombatContext) error {
	m.AI.UpdateCombatContext(ctx)
	return nil
}
