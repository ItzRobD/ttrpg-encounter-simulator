package character

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/martial_attack_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/testhelpers"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"math/rand/v2"
	"testing"
)

// target with controllable AC using the lightweight entity fake
type targetZeroAC struct{ testhelpers.EmEntity }

func (t targetZeroAC) GetAC() int { return 0 }
func (t targetZeroAC) MakeSavingThrow(ability core.Ability, targetValue int, isSpell bool, damageType core.DamageType, simOptions *core.SimulationOptions) (core.RollResult, error) {
	return nil, nil
}

// build a minimal Character with seeded RNG and initialized managers (no DB)
func newTestCharacter(t *testing.T, as core.AbilityScores, lvl uint8) *Character {
	t.Helper()
	ch := &Character{
		Name:          "Tester",
		Level:         lvl,
		AbilityScores: as,
		Seed:          core.Seed{Seed1: 11, Seed2: 22},
		RNG:           rand.New(rand.NewPCG(11, 22)),
	}

	// Roll manager
	ch.RollManager = roll_manager.NewRollManager(ch, roll_manager.RerollAbilities{})

	// Entity state manager (set 1 attack by default so character attacks produce results)
	esm, err := entity_state_manager.NewEntityStateManager(ch, entity_state_manager.EntityStateConfig{
		CurrentHP:   10,
		MaxHP:       10,
		AttackCount: 1,
		Conditions:  core.NewEntityConditions(),
	})
	if err != nil {
		t.Fatalf("NewEntityStateManager: %v", err)
	}
	ch.EntityStateManager = esm

	// Equipment manager
	em, err := equipment_manager.NewEquipmentManager(ch)
	if err != nil {
		t.Fatalf("NewEquipmentManager: %v", err)
	}
	ch.EquipmentManager = em

	// Martial manager
	ch.MartialAttackManager = martial_attack_manager.NewMartialAttackManager(ch, ch.RollManager)

	return ch
}

func TestRollInitiative_WritesState(t *testing.T) {
	ch := newTestCharacter(t, core.AbilityScores{Dexterity: 14}, 5)
	total, err := ch.RollInitiative()
	if err != nil {
		t.Fatalf("RollInitiative error: %v", err)
	}
	if total != ch.EntityStateManager.GetInitiative() {
		t.Errorf("initiative mismatch: got=%d state=%d", total, ch.EntityStateManager.GetInitiative())
	}
}

func TestInitializeHP_Variants_ValueAverageRoll(t *testing.T) {
	ch := newTestCharacter(t, core.AbilityScores{Constitution: 14}, 5)

	// capture events
	var eventsSeen int
	ch.SetEventListener(func(e interface{}) { eventsSeen++ })

	// Value
	ch.HPConfig = core.HPConfig{HPSetMethod: core.HPSetValue, Value: 15, HitDie: core.D8}
	if err := ch.InitializeHP(); err != nil {
		t.Fatalf("InitializeHP value: %v", err)
	}
	if ch.EntityStateManager.GetCurrentHP() != 15 || ch.EntityStateManager.GetMaxHP() != 15 {
		t.Errorf("value HP failed: hp=%d max=%d", ch.EntityStateManager.GetCurrentHP(), ch.EntityStateManager.GetMaxHP())
	}

	// Average
	ch.HPConfig = core.HPConfig{HPSetMethod: core.HPSetAverage, HPAverage: 12, HitDie: core.D8}
	if err := ch.InitializeHP(); err != nil {
		t.Fatalf("InitializeHP average: %v", err)
	}
	if ch.EntityStateManager.GetCurrentHP() != 12 || ch.EntityStateManager.GetMaxHP() != 12 {
		t.Errorf("average HP failed: hp=%d max=%d", ch.EntityStateManager.GetCurrentHP(), ch.EntityStateManager.GetMaxHP())
	}

	// Roll path (characters use level to determine dice; RollHP requires NumberOfDice > 0 for validation)
	ch.HPConfig = core.HPConfig{HPSetMethod: core.HPSetRoll, HitDie: core.D8, NumberOfDice: 1, Modifier: 0}
	if err := ch.InitializeHP(); err != nil {
		t.Fatalf("InitializeHP roll: %v", err)
	}
	if ch.EntityStateManager.GetCurrentHP() <= 0 || ch.EntityStateManager.GetMaxHP() <= 0 {
		t.Errorf("roll HP failed: hp=%d max=%d", ch.EntityStateManager.GetCurrentHP(), ch.EntityStateManager.GetMaxHP())
	}

	if eventsSeen == 0 {
		t.Errorf("expected at least one event to be emitted on HP initialization")
	}
}

func TestExecuteAIRequest_MeleeProducesDamage(t *testing.T) {
	// STR 16 (mod 3), prof bonus at level 5 = +3
	ch := newTestCharacter(t, core.AbilityScores{Strength: 16, Dexterity: 14}, 5)

	longsword := &weapon.Weapon{
		Name:         "Longsword",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamageSlashing,
		Properties: weapon.Properties{
			IsRanged: false,
		},
	}
	if err := ch.EquipmentManager.SetWeapon(core.WSPrimary, longsword, true); err != nil {
		t.Fatalf("SetWeapon: %v", err)
	}

	tgt := targetZeroAC{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil)}
	req := &core.AIRequest{
		ActionType:   core.ATMelee,
		WeaponSlot:   core.WSPrimary,
		Advantage:    core.RollNormal,
		UseVersatile: false,
		SimOptions:   &core.SimulationOptions{}, // avoid nil deref in CreateAttackRequest
		Target:       tgt,
		TargetID:     0,
		ActorID:      1,
	}

	out, err := ch.ExecuteAIRequest(req)
	if err != nil {
		t.Fatalf("ExecuteAIRequest(melee) error: %v", err)
	}
	if out.ActionType != core.ATMelee {
		t.Fatalf("unexpected action type: %v", out.ActionType)
	}
	if len(out.Effects) != 1 {
		t.Fatalf("expected 1 effect (one attack), got %d", len(out.Effects))
	}
	if out.Effects[0].Type != core.EffectDamage || out.Effects[0].Value <= 0 {
		t.Errorf("expected damage effect with value > 0, got %+v", out.Effects[0])
	}
}

func TestCreateAttackRequest_PropagatesOptions(t *testing.T) {
	ch := newTestCharacter(t, core.AbilityScores{Strength: 16, Dexterity: 14}, 5)
	rapier := &weapon.Weapon{
		Name:         "Rapier",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamagePiercing,
		Properties: weapon.Properties{
			IsRanged:  false,
			IsFinesse: true,
		},
	}
	if err := ch.EquipmentManager.SetWeapon(core.WSPrimary, rapier, true); err != nil {
		t.Fatalf("SetWeapon: %v", err)
	}
	tgt := targetZeroAC{EmEntity: testhelpers.NewEmEntity(1, core.AbilityScores{}, nil)}

	req, err := ch.CreateAttackRequest(tgt, core.WSPrimary, false, &core.SimulationOptions{UseImprovedCriticals: true})
	if err != nil {
		t.Fatalf("CreateAttackRequest: %v", err)
	}

	if !req.AttackOptions.ImprovedCritical {
		t.Errorf("improved critical not propagated")
	}
	if len(req.AttackData) != 1 || req.AttackData[0].Name != rapier.Name {
		t.Errorf("attack data mismatch: %+v", req.AttackData)
	}
}

func TestCreateWeaponAttackData_Modifiers(t *testing.T) {
	// STR 16 (+3), DEX 14 (+2), level 5 (prof +3)
	ch := newTestCharacter(t, core.AbilityScores{Strength: 16, Dexterity: 14}, 5)

	// Non-finesse melee should use STR
	mace := &weapon.Weapon{Name: "Mace", NumberOfDice: 1, Die: core.D6, DamageType: core.DamageBludgeoning, Properties: weapon.Properties{IsRanged: false}}
	if err := ch.EquipmentManager.SetWeapon(core.WSPrimary, mace, true); err != nil {
		t.Fatalf("SetWeapon: %v", err)
	}
	ad, err := ch.CreateWeaponAttackData(core.WSPrimary, false)
	if err != nil {
		t.Fatalf("CreateWeaponAttackData: %v", err)
	}
	if ad.AttackModifier != 6 { // 3 STR + 3 prof
		t.Errorf("attack mod=%d want 6", ad.AttackModifier)
	}
	if ad.DamageModifier != 3 {
		t.Errorf("damage mod=%d want 3", ad.DamageModifier)
	}

	// Finesse melee should use DEX if higher
	rapier := &weapon.Weapon{Name: "Rapier", NumberOfDice: 1, Die: core.D8, DamageType: core.DamagePiercing, Properties: weapon.Properties{IsRanged: false, IsFinesse: true}}
	if err := ch.EquipmentManager.SetWeapon(core.WSSecondary, rapier, true); err != nil {
		t.Fatalf("SetWeapon secondary: %v", err)
	}
	ad2, err := ch.CreateWeaponAttackData(core.WSSecondary, false)
	if err != nil {
		t.Fatalf("CreateWeaponAttackData finesse: %v", err)
	}
	if ad2.AttackModifier != 6 { // best of STR(+3) or DEX(+2) + prof(+3)
		t.Errorf("finesse attack mod=%d want 6", ad2.AttackModifier)
	}
	if ad2.DamageModifier != 3 { // best of STR(+3) or DEX(+2)
		t.Errorf("finesse damage mod=%d want 3", ad2.DamageModifier)
	}
}

func TestCreateWeaponAttackData_MagicalModifiers(t *testing.T) {
	// Build a character with STR 16 (+3) and Level 5 (Prof +3)
	ch := newTestCharacter(t, core.AbilityScores{Strength: 16, Dexterity: 14}, 5)

	sword := &weapon.Weapon{
		Name:         "Magic Sword",
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamageSlashing,
		Properties:   weapon.Properties{IsRanged: false},
	}
	sword.SetModifiers(weapon.Modifiers{
		IsMagic:     true,
		AttackBonus: 1,
		DamageBonus: 1,
	})

	if err := ch.EquipmentManager.SetWeapon(core.WSPrimary, sword, true); err != nil {
		t.Fatalf("SetWeapon: %v", err)
	}

	ad, err := ch.CreateWeaponAttackData(core.WSPrimary, false)
	if err != nil {
		t.Fatalf("CreateWeaponAttackData: %v", err)
	}

	// Normal: 3 (STR) + 3 (Prof) = 6
	// With +1 weapon: 6 + 1 = 7
	if ad.AttackModifier != 7 {
		t.Errorf("AttackModifier = %d, want 7", ad.AttackModifier)
	}
	// Normal damage: 3 (STR)
	// With +1 weapon: 3 + 1 = 4
	if ad.DamageModifier != 4 {
		t.Errorf("DamageModifier = %d, want 4", ad.DamageModifier)
	}

	foundMagic := false
	for _, rb := range ad.ResistBreakers {
		if rb == core.ResistBreakerMagic {
			foundMagic = true
			break
		}
	}
	if !foundMagic {
		t.Error("expected magic resist breaker")
	}
}

func TestTargetPriority_RoundTrip(t *testing.T) {
	ch := newTestCharacter(t, core.AbilityScores{}, 1)
	ch.SetTargetPriority(core.PrioritizeMostDamaged)
	if got := ch.GetTargetPriority(); got != core.PrioritizeMostDamaged {
		t.Errorf("target priority mismatch: got=%v want=%v", got, core.PrioritizeMostDamaged)
	}
}
