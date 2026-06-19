package actor

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/spell_manager"
	"testing"
)

func TestIsHealer(t *testing.T) {
	t.Run("Spellcaster with healing spells is healer", func(t *testing.T) {
		a := &Actor{
			SpellManager: spell_manager.SpellManager{
				HealingSpellCount: 1,
			},
		}
		if !a.IsHealer() {
			t.Errorf("Expected actor with healing spells to be healer")
		}
	})

	t.Run("Actor with healing action is healer", func(t *testing.T) {
		a := &Actor{
			Actions: []core.Action{
				{Name: "Heal", AverageHeal: 10},
			},
		}
		if !a.IsHealer() {
			t.Errorf("Expected actor with healing action to be healer")
		}
	})

	t.Run("Paladin with Lay on Hands is healer", func(t *testing.T) {
		a := &Actor{
			Features: []core.Feature{
				{Name: core.SpecAbilityLayOnHands},
			},
		}
		if !a.IsHealer() {
			t.Errorf("Expected Paladin with Lay on Hands to be healer")
		}
	})

	t.Run("Actor without healing capabilities is not healer", func(t *testing.T) {
		a := &Actor{
			Actions: []core.Action{
				{Name: "Attack", AverageDamage: 10},
			},
		}
		if a.IsHealer() {
			t.Errorf("Expected actor without healing capabilities to not be healer")
		}
	})

	t.Run("Actor with Second Wind is NOT healer", func(t *testing.T) {
		a := &Actor{
			Features: []core.Feature{
				{Name: core.SpecAbilitySecondWind},
			},
		}
		if a.IsHealer() {
			t.Errorf("Expected actor with ONLY Second Wind to NOT be flagged as healer")
		}
	})

	t.Run("Actor with Second Wind Action is NOT healer", func(t *testing.T) {
		a := &Actor{
			Actions: []core.Action{
				{Name: string(core.SpecAbilitySecondWind), AverageHeal: 10},
			},
		}
		if a.IsHealer() {
			t.Errorf("Expected actor with Second Wind Action to NOT be flagged as healer")
		}
	})
}
