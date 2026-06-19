package character

import (
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"fmt"
)

// getAbilityScore returns the score for the specified ability of the character. Defaults to 0 if the ability is not found.
func (c *Character) getAbilityScore(ability core.Ability) int {
	switch ability {
	case core.AbilityStrength:
		return c.AbilityScores.Strength
	case core.AbilityDexterity:
		return c.AbilityScores.Dexterity
	case core.AbilityConstitution:
		return c.AbilityScores.Constitution
	case core.AbilityIntelligence:
		return c.AbilityScores.Intelligence
	case core.AbilityWisdom:
		return c.AbilityScores.Wisdom
	case core.AbilityCharisma:
		return c.AbilityScores.Charisma
	default:
		return 0
	}
}

// getAbilityScoreModifier calculates the ability score modifier for a given ability based on the character's ability scores.
// Returns the modifier as an integer or an error if the ability is invalid.
func (c *Character) getAbilityScoreModifier(ability core.Ability) (int, error) {
	var abilityMod int
	var err error
	switch ability {
	case core.AbilityStrength:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Strength)
	case core.AbilityDexterity:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Dexterity)
	case core.AbilityConstitution:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Constitution)
	case core.AbilityIntelligence:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Intelligence)
	case core.AbilityWisdom:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Wisdom)
	case core.AbilityCharisma:
		abilityMod, err = core.GetAbilityScoreModifier(c.AbilityScores.Charisma)
	default:
		abilityMod = 0
		err = fmt.Errorf("invalid ability provided: %s", ability)
	}
	if err != nil {
		return 0, err
	}
	return abilityMod, nil
}

// getIsProficientInAbility checks if the character is proficient in the specified ability and returns true if proficient.
func (c *Character) getIsProficientInAbility(ability core.Ability) bool {
	switch ability {
	case core.AbilityStrength:
		return c.AbilityScoreProf.Strength
	case core.AbilityDexterity:
		return c.AbilityScoreProf.Dexterity
	case core.AbilityConstitution:
		return c.AbilityScoreProf.Constitution
	case core.AbilityIntelligence:
		return c.AbilityScoreProf.Intelligence
	case core.AbilityWisdom:
		return c.AbilityScoreProf.Wisdom
	case core.AbilityCharisma:
		return c.AbilityScoreProf.Charisma
	default:
		return false
	}
}

// getSavingThrowBonus calculates the saving throw bonus from ability modifiers and proficiency based on character level.
func (c *Character) getSavingThrowBonus(ability core.Ability) (int, error) {
	var pb int
	var mod int
	var err error

	mod, err = c.getAbilityScoreModifier(ability)
	pb, err = core.GetCharacterProficiencyBonus(c.Level)
	if err != nil {
		return 0, err
	}

	if c.getIsProficientInAbility(ability) {
		return pb + mod, nil
	}
	return mod, nil
}
