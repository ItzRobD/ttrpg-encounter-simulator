package monster

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
)

// GetAbilityScore returns the score for the specified ability of the monster. Defaults to 0 if the ability is not found.
func (m *Monster) GetAbilityScore(ability core.Ability) int {
	switch ability {
	case core.AbilityStrength:
		return m.AbilityScores.Strength
	case core.AbilityDexterity:
		return m.AbilityScores.Dexterity
	case core.AbilityConstitution:
		return m.AbilityScores.Constitution
	case core.AbilityIntelligence:
		return m.AbilityScores.Intelligence
	case core.AbilityWisdom:
		return m.AbilityScores.Wisdom
	case core.AbilityCharisma:
		return m.AbilityScores.Charisma
	default:
		return 0
	}
}

func (m *Monster) GetAbilityScoreModifier(ability core.Ability) (int, error) {
	var abilityMod int
	var err error
	switch ability {
	case core.AbilityStrength:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Strength)
	case core.AbilityDexterity:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Dexterity)
	case core.AbilityConstitution:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Constitution)
	case core.AbilityIntelligence:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Intelligence)
	case core.AbilityWisdom:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Wisdom)
	case core.AbilityCharisma:
		abilityMod, err = core.GetAbilityScoreModifier(m.AbilityScores.Charisma)
	default:
		abilityMod = 0
		err = fmt.Errorf("invalid ability provided: %s", ability)
	}
	if err != nil {
		return 0, err
	}
	return abilityMod, nil
}

func (m *Monster) getIsProficientInAbility(ability core.Ability) bool {
	switch ability {
	case core.AbilityStrength:
		return m.AbilityScoreProf.Strength
	case core.AbilityDexterity:
		return m.AbilityScoreProf.Dexterity
	case core.AbilityConstitution:
		return m.AbilityScoreProf.Constitution
	case core.AbilityIntelligence:
		return m.AbilityScoreProf.Intelligence
	case core.AbilityWisdom:
		return m.AbilityScoreProf.Wisdom
	case core.AbilityCharisma:
		return m.AbilityScoreProf.Charisma
	default:
		return false
	}
}

func (m *Monster) GetSavingThrowBonus(ability core.Ability) (int, error) {
	var pb int
	var mod int
	var err error

	mod, err = m.GetAbilityScoreModifier(ability)
	if err != nil {
		return 0, err
	}
	pb, err = core.GetMonsterProficiencyBonus(m.CR)
	if err != nil {
		return 0, err
	}

	if m.getIsProficientInAbility(ability) {
		return pb + mod, nil
	}
	return mod, nil
}
