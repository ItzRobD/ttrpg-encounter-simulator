package state_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"testing"
)

func TestStateManager_Conditions(t *testing.T) {
	sm := StateManager{
		Conditions: core.NewActorConditions(),
	}

	if sm.Conditions.Has(core.ConditionBlinded) {
		t.Error("Should not be blinded initially")
	}

	sm.Conditions.Add(core.ConditionBlinded)
	if !sm.Conditions.Has(core.ConditionBlinded) {
		t.Error("Should be blinded after Add")
	}

	sm.Conditions.Remove(core.ConditionBlinded)
	if sm.Conditions.Has(core.ConditionBlinded) {
		t.Error("Should not be blinded after Remove")
	}
}

func TestStateManager_ModifyHP_Damage(t *testing.T) {
	sm := &StateManager{
		CurrentHP: 10,
		MaxHP:     20,
		TempHP:    5,
	}

	// Damage less than TempHP
	sm.ModifyHP(-3, false, false)
	if sm.TempHP != 2 || sm.CurrentHP != 10 {
		t.Errorf("Expected TempHP 2, CurrentHP 10; got TempHP %d, CurrentHP %d", sm.TempHP, sm.CurrentHP)
	}

	// Damage equal to TempHP
	sm.ModifyHP(-2, false, false)
	if sm.TempHP != 0 || sm.CurrentHP != 10 {
		t.Errorf("Expected TempHP 0, CurrentHP 10; got TempHP %d, CurrentHP %d", sm.TempHP, sm.CurrentHP)
	}

	// Damage more than TempHP (but none left)
	sm.ModifyHP(-5, false, false)
	if sm.TempHP != 0 || sm.CurrentHP != 5 {
		t.Errorf("Expected TempHP 0, CurrentHP 5; got TempHP %d, CurrentHP %d", sm.TempHP, sm.CurrentHP)
	}

	// Damage with TempHP
	sm.TempHP = 10
	sm.ModifyHP(-15, false, false)
	if sm.TempHP != 0 || sm.CurrentHP != 0 {
		t.Errorf("Expected TempHP 0, CurrentHP 0; got TempHP %d, CurrentHP %d", sm.TempHP, sm.CurrentHP)
	}
}

func TestStateManager_ModifyHP_Healing(t *testing.T) {
	sm := &StateManager{
		CurrentHP: 5,
		MaxHP:     20,
	}

	// Normal healing
	sm.ModifyHP(10, false, false)
	if sm.CurrentHP != 15 {
		t.Errorf("Expected CurrentHP 15; got %d", sm.CurrentHP)
	}

	// Overhealing
	sm.ModifyHP(10, false, false)
	if sm.CurrentHP != 20 {
		t.Errorf("Expected CurrentHP 20 (MaxHP); got %d", sm.CurrentHP)
	}
}

func TestStateManager_ModifyHP_TempHP(t *testing.T) {
	sm := &StateManager{
		TempHP: 5,
	}

	// New TempHP higher
	sm.ModifyHP(10, true, false)
	if sm.TempHP != 10 {
		t.Errorf("Expected TempHP 10; got %d", sm.TempHP)
	}

	// New TempHP lower
	sm.ModifyHP(5, true, false)
	if sm.TempHP != 10 {
		t.Errorf("Expected TempHP 10 (highest stays); got %d", sm.TempHP)
	}
}

func TestStateManager_ModifyHP_HealthState(t *testing.T) {
	sm := &StateManager{
		CurrentHP: 20,
		MaxHP:     20,
	}

	sm.ModifyHP(-5, false, false) // 15/20 = 75% -> Wounded
	if sm.HealthState != core.HealthStateWounded {
		t.Errorf("Expected HealthStateWounded; got %s", sm.HealthState)
	}

	sm.ModifyHP(-14, false, false) // 1/20 = 5% -> Critical
	if sm.HealthState != core.HealthStateCritical {
		t.Errorf("Expected HealthStateCritical; got %s", sm.HealthState)
	}

	sm.ModifyHP(-1, false, false) // 0/20 -> Dead
	if sm.HealthState != core.HealthStateDead {
		t.Errorf("Expected HealthStateDead; got %s", sm.HealthState)
	}
}
