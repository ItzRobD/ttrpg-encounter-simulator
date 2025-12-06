package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"fmt"
)

// GetSpellBonus calculates the spellcasting bonus based on the character's ability score and optionally adds proficiency bonus.
// Returns the calculated bonus or an error if the spellcasting modifier or level is invalid.
func (c *Character) GetSpellBonus(addProficiency bool) (int, error) {
	var mod int
	var err error
	switch c.Class.SpellcastingMod {
	case core.AbilityStrength:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Strength)
	case core.AbilityDexterity:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Dexterity)
	case core.AbilityConstitution:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Constitution)
	case core.AbilityIntelligence:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Intelligence)
	case core.AbilityWisdom:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Wisdom)
	case core.AbilityCharisma:
		mod, err = core.GetAbilityScoreModifier(c.AbilityScores.Charisma)
	default:
		mod = 0
		err = fmt.Errorf("invalid spellcasting modifier: %s", c.Class.SpellcastingMod)
	}

	if addProficiency {
		var pb int
		pb, err = core.GetCharacterProficiencyBonus(c.Level)
		if err != nil {
			return 0, err
		}

		return mod + pb, nil
	}

	return mod, nil
}

// GetSpellSaveDC calculates the spell save DC for the character based on the given ability modifier and a base value of 8.
func (c *Character) GetSpellSaveDC(ability *core.Ability) (int, error) {
	if ability == nil {
		return 8, fmt.Errorf("ability cannot be nil")
	}
	var abilityMod int
	var err error
	switch *ability {
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
	return 8 + abilityMod, nil
}

// CreateSpellAttackData creates and returns the data for a spell attack, including attack and spell modifiers.
// It takes a SpellChoice as input and computes the necessary modifiers for the attack.
// Returns a SpellCastData struct and an error if any calculation fails.
func (c *Character) CreateSpellAttackData(spellChoice core.SpellChoice) (spellcasting_manager.SpellCastData, error) {
	spellBonus, err := c.GetSpellBonus(true)
	if err != nil {
		return spellcasting_manager.SpellCastData{}, err
	}

	spellMod, err := c.GetSpellBonus(false)
	if err != nil {
		return spellcasting_manager.SpellCastData{}, err
	}

	return spellcasting_manager.SpellCastData{
		SpellChoice:          spellChoice,
		AttackModifier:       spellBonus,
		SpellcastingModifier: spellMod,
	}, nil
}

// CreateSpellCastRequest generates a new SpellCastRequest based on the given spell choice and advantage type.
func (c *Character) CreateSpellCastRequest(target core.Entity, spellChoice core.SpellChoice, adv core.AdvantageType, simOptions *core.SimulationOptions) (*spellcasting_manager.SpellCastRequest, error) {
	spellcastData, err := c.CreateSpellAttackData(spellChoice)
	if err != nil {
		return nil, err
	}

	// Build spell options dynamically (minimal — feats/features handled later)
	options := spellcasting_manager.SpellOptions{
		Advantage:            adv,
		BonusToAttackRoll:    0,
		BonusToDamageRoll:    0,
		ShouldApplyDamageMod: false, // RollSpellValue handles spell modifiers when applicable
		ImprovedCritical:     (spellChoice.Spell.GetSpellType() == core.STDamage) && (simOptions != nil && simOptions.UseImprovedCriticals),
		TreatOnesAsTwos:      false, // Elemental Adept and similar will be wired via features later
	}

	return &spellcasting_manager.SpellCastRequest{
		SpellCastData:     spellcastData,
		SpellOptions:      options,
		SimulationOptions: simOptions,
		Target:            target,
	}, nil
}

func (c *Character) ChooseSpellByHealingEfficiency(targetValue int) (*core.SpellChoice, error) {
	choice, err := c.SpellCastingManager.GetMostEfficientHealingSpell(targetValue)
	if err != nil {
		return nil, err
	}
	return choice, nil
}

func (c *Character) ChooseDamageSpellByPriority(p core.SpellPriority) (*core.SpellChoice, error) {
	return c.SpellCastingManager.ChooseSpellByPriority(core.STDamage, p)
}
