package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/testhelpers"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"testing"
)

func TestExecuteAIRequest_MultiAttackCountEffects(t *testing.T) {
	ch := newTestCharacter(t, core.AbilityScores{Strength: 16}, 5)
	// Two attacks per action
	ch.EntityState.SetNumberOfExtraAttacks(2)

	sword := &weapon.Weapon{Name: "Sword", NumberOfDice: 1, Die: core.D6, DamageType: core.DamageSlashing, IsMelee: true}
	if err := ch.EquipmentManager.SetWeapon(core.WSPrimary, sword, true); err != nil {
		t.Fatalf("SetWeapon: %v", err)
	}

	tgt := targetZeroAC{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil)}
	req := &core.AIRequest{ActionType: core.ATMelee, WeaponSlot: core.WSPrimary, Advantage: core.RollNormal, UseVersatile: false, SimOptions: &core.SimulationOptions{}, Target: tgt}
	out, err := ch.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest multi: %v", err)
	}
	if len(out.Effects) != 2 {
		t.Fatalf("expected 2 effects for two attacks, got %d", len(out.Effects))
	}
}

func TestExecuteAIRequest_RangedProducesDamage(t *testing.T) {
	ch := newTestCharacter(t, core.AbilityScores{Dexterity: 16}, 5)
	bow := &weapon.Weapon{Name: "Shortbow", NumberOfDice: 1, Die: core.D6, DamageType: core.DamagePiercing, IsMelee: false, IsRanged: true, IsFinesse: false}
	if err := ch.EquipmentManager.SetWeapon(core.WSRanged, bow, true); err != nil {
		t.Fatalf("SetWeapon ranged: %v", err)
	}
	tgt := targetZeroAC{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil)}
	req := &core.AIRequest{ActionType: core.ATRanged, WeaponSlot: core.WSRanged, Advantage: core.RollNormal, UseVersatile: false, SimOptions: &core.SimulationOptions{}, Target: tgt}
	out, err := ch.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest ranged: %v", err)
	}
	if len(out.Effects) != 1 || out.Effects[0].Type != core.EffectDamage || out.Effects[0].Value <= 0 {
		t.Errorf("expected single damage effect > 0, got %+v", out.Effects)
	}
}

func TestCreateWeaponAttackData_VersatileDie(t *testing.T) {
	ch := newTestCharacter(t, core.AbilityScores{Strength: 16}, 5)
	longsword := &weapon.Weapon{Name: "Longsword", NumberOfDice: 1, Die: core.D8, DamageType: core.DamageSlashing, IsMelee: true, IsVersatile: true}
	if err := ch.EquipmentManager.SetWeapon(core.WSPrimary, longsword, true); err != nil {
		t.Fatalf("SetWeapon: %v", err)
	}
	ad, err := ch.CreateWeaponAttackData(core.WSPrimary, true)
	if err != nil {
		t.Fatalf("CreateWeaponAttackData versatile: %v", err)
	}
	if !ad.IsVersatileAttack || ad.Die != core.D10 {
		t.Errorf("expected versatile D10, got versatile=%v die=%v", ad.IsVersatileAttack, ad.Die)
	}
}

func TestCharacter_GetAIRequest_InvalidType_Error(t *testing.T) {
	ch := newTestCharacter(t, core.AbilityScores{}, 1)
	if _, err := ch.GetAIRequest(1, core.AIRequestType(255)); err == nil {
		t.Fatalf("expected error for invalid AI request type")
	}
}

func TestMakeSavingThrow_ReturnsResult(t *testing.T) {
	ch := newTestCharacter(t, core.AbilityScores{Dexterity: 14}, 5)
	res, err := ch.MakeSavingThrow(core.AbilityDexterity, 10)
	if err != nil {
		t.Fatalf("MakeSavingThrow: %v", err)
	}
	if res == nil {
		t.Fatalf("nil result from MakeSavingThrow")
	}
	if res.GetDiceRollType() != core.DiceRollSavingThrow {
		t.Errorf("unexpected roll type: %v", res.GetDiceRollType())
	}
}
