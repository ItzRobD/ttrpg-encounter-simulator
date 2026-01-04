package character

import (
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"testing"
)

func TestCharacter_SneakAttack(t *testing.T) {
	as := core.AbilityScores{Strength: 10, Dexterity: 16, Constitution: 14, Intelligence: 10, Wisdom: 10, Charisma: 10}
	c := newTestCharacter(t, as, 3) // Level 3 rogue has 2d6 sneak attack
	c.Class.ID = classes.Rogue
	c.Class.ClassFeatures.RogueFeatures = &classes.RogueFeatures{
		NumOfSneakAttackDice: 2,
	}

	// Equip a weapon so CreateAttackRequest doesn't fail
	w, _ := weapon.New("Dagger", false, true, 1, core.D4, core.DamagePiercing, weapon.Properties{IsFinesse: true})
	c.EquipmentManager.SetWeapon(core.WSPrimary, &w, true)

	target := targetZeroAC{}
	simOptions := &core.SimulationOptions{
		EnableClassFeatures:  true,
		AlwaysUseSneakAttack: true,
	}

	// 1. Sneak attack with AlwaysUseSneakAttack
	req := &core.AIRequest{
		ActionType: core.ATMelee,
		Target:     target,
		WeaponSlot: core.WSPrimary,
		SimOptions: simOptions,
	}

	_, err := c.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("failed to execute ai request: %v", err)
	}

	if !c.EntityStateManager.GetHasUsedSneakAttack() {
		t.Errorf("expected sneak attack to be used")
	}

	// 2. Sneak attack only once per turn
	c.EntityStateManager.SetHasUsedSneakAttack(true)
	_, _ = c.ExecuteAIRequest(req)

	c.EntityStateManager.RefreshActions()
	if c.EntityStateManager.GetHasUsedSneakAttack() {
		t.Errorf("expected sneak attack to be reset after RefreshActions")
	}

	// 3. Sneak attack with advantage
	simOptions.AlwaysUseSneakAttack = false
	// We need to force advantage. Advantage is determined by DetermineAttackAdvantageForEntities
	// which uses conditions.
	c.EntityStateManager.AddCondition(core.ConditionInvisible) // Being invisible gives advantage

	_, err = c.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("failed to execute ai request: %v", err)
	}

	if !c.EntityStateManager.GetHasUsedSneakAttack() {
		t.Errorf("expected sneak attack to be used with advantage")
	}

	// 4. No sneak attack with ranged attack
	c.EntityStateManager.RefreshActions()
	simOptions.AlwaysUseSneakAttack = true
	req.ActionType = core.ATRanged
	// Equip a ranged weapon
	rw, _ := weapon.New("Shortbow", false, false, 1, core.D6, core.DamagePiercing, weapon.Properties{IsRanged: true, IsOnlyRanged: true})
	c.EquipmentManager.SetWeapon(core.WSRanged, &rw, true)
	req.WeaponSlot = core.WSRanged

	_, err = c.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("failed to execute ai request: %v", err)
	}
	if c.EntityStateManager.GetHasUsedSneakAttack() {
		t.Errorf("expected NO sneak attack with ranged weapon")
	}
}
