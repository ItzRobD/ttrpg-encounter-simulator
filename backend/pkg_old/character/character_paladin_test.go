package character

import (
	"dnd5e-encounter-simulator-backend/pkg_old/classes"
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"dnd5e-encounter-simulator-backend/pkg_old/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg_old/core/testhelpers"
	"dnd5e-encounter-simulator-backend/pkg_old/monster"
	"dnd5e-encounter-simulator-backend/pkg_old/spells"
	"dnd5e-encounter-simulator-backend/pkg_old/weapon"
	"testing"
)

type mockTarget struct {
	testhelpers.EmEntity
	targetType string
	ac         int
}

func (t mockTarget) GetEntityType() core.EntityType {
	//TODO implement me
	panic("implement me")
}

func (t mockTarget) Regenerate() {
	//TODO implement me
	panic("implement me")
}

func (t mockTarget) GetType() string {
	return t.targetType
}

func (t mockTarget) GetAC() int {
	return t.ac
}

func (t mockTarget) MakeSavingThrow(core.Ability, int, bool, core.DamageType, *core.SimulationOptions) (core.RollResult, error) {
	return nil, nil
}

func TestPaladin_DivineSmite(t *testing.T) {
	as := core.AbilityScores{Strength: 16, Charisma: 14}
	lvl := uint8(2)

	// Setup Paladin with Divine Smite
	char := newTestCharacter(t, as, lvl)
	char.Class = classes.Class{
		ID:   classes.Paladin,
		Name: "Paladin",
		ClassFeatures: classes.ClassFeatures{
			PaladinFeatures: &classes.PaladinFeatures{
				HasDivineSmite: true,
			},
		},
	}

	// Setup Spell Slots (2 slots at level 1)
	maxSlots := spells.SpellSlots{1: 2}
	currSlots := spells.SpellSlots{1: 2}
	char.SpellCastingManager = initializeSpellcastingManagerForTest(char, currSlots, maxSlots)

	// Setup Weapon
	longsword := &weapon.Weapon{
		Name: "Longsword",
		DamageBlocks: []core.DamageBlock{
			{NumberOfDice: 1, Die: core.D8, DamageType: core.DamageSlashing},
		},
	}
	char.EquipmentManager.SetWeapon(core.WSPrimary, longsword, true)

	t.Run("Basic Divine Smite", func(t *testing.T) {
		tgt := mockTarget{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), targetType: "Humanoid", ac: 0}
		simOpts := &core.SimulationOptions{
			EnableClassFeatures: true,
			PaladinAlwaysSmite:  true,
			CanCharactersCrit:   true,
		}

		req := &core.AIRequest{
			ActionType: core.ATMelee,
			WeaponSlot: core.WSPrimary,
			Target:     tgt,
			TargetID:   1,
			SimOptions: simOpts,
			Actor:      char,
		}

		// Reset slots for this sub-test
		char.SpellCastingManager.RecoverSpellSlotToMax(1)

		outcome, err := char.ExecuteAIRequest(req)
		if err != nil {
			t.Fatalf("ExecuteAIRequest failed: %v", err)
		}

		// Expect 2 effects: weapon damage and divine smite damage
		if len(outcome.Effects) != 2 {
			t.Errorf("Expected 2 effects, got %d", len(outcome.Effects))
		}

		// Verify Divine Smite effect
		foundSmite := false
		for _, e := range outcome.Effects {
			if e.DamageType == core.DamageRadiant {
				foundSmite = true
				// 1st level slot = 2d8 (avg 9)
				if e.Value < 2 || e.Value > 16 {
					t.Errorf("Divine Smite damage out of range: %d", e.Value)
				}
			}
		}
		if !foundSmite {
			t.Errorf("Divine Smite effect not found")
		}

		// Verify slot spent
		if char.SpellCastingManager.HasSpellSlotsAtLevel(1) && char.SpellCastingManager.GetStatus().CurrentSlots[1] != 1 {
			t.Errorf("Expected 1 slot remaining, got %d", char.SpellCastingManager.GetStatus().CurrentSlots[1])
		}
	})

	t.Run("Divine Smite scaling with slot level", func(t *testing.T) {
		// Level 5 Paladin has 2nd level slots
		maxSlots := spells.SpellSlots{1: 4, 2: 2}
		currSlots := spells.SpellSlots{1: 4, 2: 2}
		char.SpellCastingManager = initializeSpellcastingManagerForTest(char, currSlots, maxSlots)

		tgt := mockTarget{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), targetType: "Humanoid", ac: 0}
		simOpts := &core.SimulationOptions{
			EnableClassFeatures:        true,
			PaladinAlwaysSmite:         true,
			PaladinUseHighestSmiteSlot: true,
			CanCharactersCrit:          true,
		}

		req := &core.AIRequest{
			ActionType: core.ATMelee,
			WeaponSlot: core.WSPrimary,
			Target:     tgt,
			TargetID:   1,
			SimOptions: simOpts,
			Actor:      char,
		}

		outcome, err := char.ExecuteAIRequest(req)
		if err != nil {
			t.Fatalf("ExecuteAIRequest failed: %v", err)
		}

		// Verify 2nd level slot used (3d8)
		foundSmite := false
		for _, e := range outcome.Effects {
			if e.DamageType == core.DamageRadiant {
				foundSmite = true
				// 2nd level slot = 3d8 (avg 13.5)
				if e.Value < 3 || e.Value > 24 {
					t.Errorf("2nd level Divine Smite damage out of range: %d", e.Value)
				}
			}
		}
		if !foundSmite {
			t.Errorf("Divine Smite effect not found")
		}

		if char.SpellCastingManager.GetStatus().CurrentSlots[2] != 1 {
			t.Errorf("Expected 1 2nd-level slot remaining, got %d", char.SpellCastingManager.GetStatus().CurrentSlots[2])
		}
	})

	t.Run("Divine Smite bonus vs Undead", func(t *testing.T) {
		char.SpellCastingManager.RecoverSpellSlotToMax(1)
		tgt := mockTarget{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), targetType: monster.MTUndead, ac: 0}
		simOpts := &core.SimulationOptions{
			EnableClassFeatures: true,
			PaladinAlwaysSmite:  true,
			CanCharactersCrit:   true,
		}

		req := &core.AIRequest{
			ActionType: core.ATMelee,
			WeaponSlot: core.WSPrimary,
			Target:     tgt,
			TargetID:   1,
			SimOptions: simOpts,
			Actor:      char,
		}

		outcome, err := char.ExecuteAIRequest(req)
		if err != nil {
			t.Fatalf("ExecuteAIRequest failed: %v", err)
		}

		for _, e := range outcome.Effects {
			if e.DamageType == core.DamageRadiant {
				// 1st level vs Undead = 3d8 (avg 13.5)
				if e.Value < 3 || e.Value > 24 {
					t.Errorf("Undead bonus Divine Smite damage out of range: %d", e.Value)
				}
			}
		}
	})

	t.Run("Improved Divine Smite", func(t *testing.T) {
		char.Class.ClassFeatures.PaladinFeatures.HasImprovedDivineSmite = true
		tgt := mockTarget{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), targetType: "Humanoid", ac: 0}
		simOpts := &core.SimulationOptions{
			EnableClassFeatures: true,
			PaladinAlwaysSmite:  false, // Disable regular smite to isolate improved
		}

		req := &core.AIRequest{
			ActionType: core.ATMelee,
			WeaponSlot: core.WSPrimary,
			Target:     tgt,
			TargetID:   1,
			SimOptions: simOpts,
			Actor:      char,
		}

		// Ensure primary weapon hit by setting high bonus or target 0 AC
		outcome, err := char.ExecuteAIRequest(req)
		if err != nil {
			t.Fatalf("ExecuteAIRequest failed: %v", err)
		}

		// Improved Divine Smite adds +1d8 radiant to the attack itself
		// Actually, in the current implementation, it's added to AttackData, so it's part of the weapon damage effect?
		// No, MartialAttackManager processes all AttackData entries.
		// Wait, let's check pkg_old/character/character_combat.go
		// adSlice = append(adSlice, core.AttackData{...})
		// So ProcessAttackRequest returns multiple results if there's multiple AttackData.
		// And ExecuteAIRequest iterates results and adds an effect for each hit.

		radiantHits := 0
		for _, e := range outcome.Effects {
			if e.DamageType == core.DamageRadiant {
				radiantHits++
				if e.Value < 1 || e.Value > 8 {
					t.Errorf("Improved Divine Smite damage out of range: %d", e.Value)
				}
			}
		}
		if radiantHits != 1 {
			t.Errorf("Expected 1 radiant damage effect from Improved Divine Smite, got %d", radiantHits)
		}
	})

	t.Run("Divine Smite Critical - Standard", func(t *testing.T) {
		char.SpellCastingManager.RecoverSpellSlotToMax(1)
		tgt := mockTarget{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), targetType: "Humanoid", ac: 0}
		simOpts := &core.SimulationOptions{
			EnableClassFeatures: true,
			CanCharactersCrit:   true,
			UseImprovedCritical: false,
		}

		// Directly test resolveDivineSmite to ensure crit logic is correct
		effect := char.resolveDivineSmite(tgt, true, simOpts)
		if effect == nil {
			t.Fatalf("resolveDivineSmite returned nil")
		}

		// 1st level slot = 2d8. Crit = 4d8. Range [4, 32].
		if effect.Value < 4 || effect.Value > 32 {
			t.Errorf("Standard crit Divine Smite damage out of range: %d", effect.Value)
		}
	})

	t.Run("Divine Smite Critical - Improved", func(t *testing.T) {
		char.SpellCastingManager.RecoverSpellSlotToMax(1)
		tgt := mockTarget{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil), targetType: "Humanoid", ac: 0}
		simOpts := &core.SimulationOptions{
			EnableClassFeatures: true,
			CanCharactersCrit:   true,
			UseImprovedCritical: true,
		}

		effect := char.resolveDivineSmite(tgt, true, simOpts)
		if effect == nil {
			t.Fatalf("resolveDivineSmite returned nil")
		}

		// 1st level slot = 2d8. Improved Crit = 2d8 + 16 (max of 2d8). Range [18, 32].
		if effect.Value < 18 || effect.Value > 32 {
			t.Errorf("Improved crit Divine Smite damage out of range: %d", effect.Value)
		}
	})
}

// Helper to initialize spellcasting manager without needing full DB context
func initializeSpellcastingManagerForTest(c *Character, current, max spells.SpellSlots) *spellcasting_manager.SpellcastingManager {
	// Paladin is a character caster
	return spellcasting_manager.NewSpellcastingManager(
		c,
		c.RollManager,
		core.CasterCharacter, // Correct type from pkg_old/core/spell_types.go
		int(c.Level),
		current,
		max,
		2, // Mod value
	)
}
