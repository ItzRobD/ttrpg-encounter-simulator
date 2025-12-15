package character

import (
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/testhelpers"
	"testing"
)

// Helper to build a minimal Character for style application tests without DB
func makeCharWithStyles(styles ...classes.FightingStyle) *Character {
	c := &Character{}
	c.Class.FightingStyles = append([]classes.FightingStyle{}, styles...)
	parent := testhelpers.NewEmEntity(1, core.AbilityScores{}, nil)
	em, _ := equipment_manager.NewEquipmentManager(parent)
	c.EquipmentManager = em
	return c
}

func Test_TWF_OffhandDamageModifier_TogglesWithStyle(t *testing.T) {
	// Offhand defaults to no ability mod
	c := makeCharWithStyles() // no styles

	ad := core.AttackData{}
	opts := core.AttackOptions{ShouldApplyDamageMod: false}

	// Simulate offhand context: slot = WSSecondary
	c.applyFightingStyles(&ad, &opts, false, false, false, core.WSSecondary)
	if opts.GetShouldApplyDamageMod() {
		t.Fatalf("expected offhand ShouldApplyDamageMod=false without TWF, got true")
	}

	// With TWF style, offhand should apply ability mod
	c = makeCharWithStyles(classes.StyleTWF)
	ad = core.AttackData{}
	opts = core.AttackOptions{ShouldApplyDamageMod: false}
	c.applyFightingStyles(&ad, &opts, false, false, false, core.WSSecondary)
	if !opts.GetShouldApplyDamageMod() {
		t.Fatalf("expected offhand ShouldApplyDamageMod=true with TWF, got false")
	}
}

func Test_Dueling_AddsPlus2Damage_NoShield_NotVersatileTwoHanded(t *testing.T) {
	c := makeCharWithStyles(classes.StyleDueling)

	// Ensure no shield and not using versatile two-handed
	c.EquipmentManager.HasShieldEquipped = false

	baseDMG := 3
	ad := core.AttackData{DamageModifier: baseDMG, IsVersatileAttack: false}
	opts := core.AttackOptions{}

	c.applyFightingStyles(&ad, &opts, false, false, false, core.WSPrimary)
	if ad.DamageModifier != baseDMG+2 {
		t.Fatalf("expected dueling to add +2 damage, got %d (base %d)", ad.DamageModifier, baseDMG)
	}

	// Negative: with shield, no bonus
	c = makeCharWithStyles(classes.StyleDueling)
	c.EquipmentManager.HasShieldEquipped = true
	ad = core.AttackData{DamageModifier: baseDMG, IsVersatileAttack: false}
	opts = core.AttackOptions{}
	c.applyFightingStyles(&ad, &opts, false, false, false, core.WSPrimary)
	if ad.DamageModifier != baseDMG {
		t.Fatalf("expected dueling no bonus when shield equipped, got %d (base %d)", ad.DamageModifier, baseDMG)
	}

	// Negative: versatile (two-handed usage), no bonus
	c = makeCharWithStyles(classes.StyleDueling)
	c.EquipmentManager.HasShieldEquipped = false
	ad = core.AttackData{DamageModifier: baseDMG, IsVersatileAttack: true}
	opts = core.AttackOptions{}
	c.applyFightingStyles(&ad, &opts, false, true, true, core.WSPrimary)
	if ad.DamageModifier != baseDMG {
		t.Fatalf("expected dueling no bonus when versatile/two-handed, got %d (base %d)", ad.DamageModifier, baseDMG)
	}
}

func Test_Archery_AddsPlus2ToHit_ForRangedWeaponsOnly(t *testing.T) {
	c := makeCharWithStyles(classes.StyleArchery)

	// Ranged weapon path
	baseAtk := 5
	ad := core.AttackData{AttackModifier: baseAtk, IsRangedWeapon: true}
	opts := core.AttackOptions{}
	c.applyFightingStyles(&ad, &opts, true, false, false, core.WSRanged)
	if ad.AttackModifier != baseAtk+2 {
		t.Fatalf("expected archery to add +2 to attack for ranged, got %d (base %d)", ad.AttackModifier, baseAtk)
	}

	// Melee weapon should not get bonus
	ad = core.AttackData{AttackModifier: baseAtk, IsRangedWeapon: false}
	opts = core.AttackOptions{}
	c.applyFightingStyles(&ad, &opts, false, false, false, core.WSPrimary)
	if ad.AttackModifier != baseAtk {
		t.Fatalf("expected no archery bonus for melee, got %d (base %d)", ad.AttackModifier, baseAtk)
	}
}

func Test_GWF_SetsRerollFlag_WhenTwoHandedOrVersatile(t *testing.T) {
	c := makeCharWithStyles(classes.StyleGWF)

	// Versatile used two-handed
	ad := core.AttackData{}
	opts := core.AttackOptions{}
	c.applyFightingStyles(&ad, &opts, false, true, false, core.WSPrimary)
	if !opts.GetTreatOnesAsTwos() { // maps to RerollOnesAndTwos flag
		t.Fatalf("expected GWF to enable reroll flag for versatile two-handed attacks")
	}

	// Two-handed weapon
	ad = core.AttackData{}
	opts = core.AttackOptions{}
	c.applyFightingStyles(&ad, &opts, false, false, true, core.WSPrimary)
	if !opts.GetTreatOnesAsTwos() {
		t.Fatalf("expected GWF to enable reroll flag for two-handed attacks")
	}

	// Negative: one-handed non-versatile
	ad = core.AttackData{}
	opts = core.AttackOptions{}
	c.applyFightingStyles(&ad, &opts, false, false, false, core.WSPrimary)
	if opts.GetTreatOnesAsTwos() {
		t.Fatalf("did not expect GWF reroll flag for one-handed non-versatile attacks")
	}
}
