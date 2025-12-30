package simulation

import (
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/classes"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/entity_state_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/martial_attack_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/roll_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/spellcasting_manager"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"math/rand/v2"
	"testing"
)

// build a minimal Character with seeded RNG and initialized managers (no DB)
func buildTestCharacter(t *testing.T, as core.AbilityScores, lvl uint8) *character.Character {
	t.Helper()
	ch := &character.Character{
		Name:          "Tester",
		Level:         lvl,
		AbilityScores: as,
		Seed:          core.Seed{Seed1: 101, Seed2: 202},
		RNG:           rand.New(rand.NewPCG(101, 202)),
		HPConfig:      core.HPConfig{HPSetMethod: core.HPSetValue, Value: 12, HitDie: core.D8},
	}
	// Roll manager
	ch.RollManager = roll_manager.NewRollManager(ch, roll_manager.RerollAbilities{})
	// ESM
	esm, err := entity_state_manager.NewEntityStateManager(ch, entity_state_manager.EntityStateConfig{
		CurrentHP:   12,
		MaxHP:       12,
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
	// AI
	ch.AI = character.NewCharacterAI(ch)
	// Minimal non-nil spellcasting manager to avoid nil deref in AI checks
	ch.SpellCastingManager = &spellcasting_manager.SpellcastingManager{}
	return ch
}

// equip a simple melee weapon in the desired slot
func equipSword(t *testing.T, ch *character.Character, slot core.WeaponSlot) {
	t.Helper()
	sword := &weapon.Weapon{Name: "Sword", NumberOfDice: 1, Die: core.D6, DamageType: core.DamageSlashing, Properties: weapon.Properties{IsRanged: false}}
	if err := ch.EquipmentManager.SetWeapon(slot, sword, true); err != nil {
		t.Fatalf("SetWeapon: %v", err)
	}
}

// build a minimal Monster as a target with given AC and seeded RNG
func buildTestMonster(t *testing.T, ac int) *monster.Monster {
	t.Helper()
	m := &monster.Monster{
		MonsterBase: monster.MonsterBase{
			Name:          "Dummy",
			AC:            ac,
			AbilityScores: core.AbilityScores{Dexterity: 12},
		},
		RNG: rand.New(rand.NewPCG(303, 404)),
	}
	m.RollManager = roll_manager.NewRollManager(m, roll_manager.RerollAbilities{})
	esm, err := entity_state_manager.NewEntityStateManager(m, entity_state_manager.EntityStateConfig{
		CurrentHP:  15,
		MaxHP:      15,
		Conditions: core.NewEntityConditions(),
	})
	if err != nil {
		t.Fatalf("NewEntityStateManager(m): %v", err)
	}
	m.EntityStateManager = esm
	// Ensure InitializeHP sets a positive value during simulations
	m.HP = core.HPConfig{HPSetMethod: core.HPSetValue, Value: 15, HitDie: core.D8}
	// Initialize action manager with a simple melee action so monster AI can act
	act := monster.Action{ActionID: 1, Name: "Claw", NumberOfDice: 1, Die: core.D6, AmountToAdd: 1, AttackBonus: 5, DamageType: core.DamageSlashing}
	mamCfg := &monster.MAMConfig{Actions: map[int]monster.Action{1: act}}
	m.ActionManager = monster.NewMonsterActionManager(m, m.RollManager, mamCfg)
	// AI
	m.AI = monster.NewMonsterAI(m)
	return m
}

func buildCombatants(att *character.Character, tgt *monster.Monster) (*core.Combatant, *core.Combatant) {
	cAtt := core.NewCombatantWithInfo(att)
	cTgt := core.NewCombatantWithInfo(tgt)
	return cAtt, cTgt
}

// helper to run a fabricated outcome through the engine against registered combatants
func runOutcome(t *testing.T, ce *CombatEngine, actor core.Entity, targetID int, effects []core.Effect) error {
	t.Helper()
	outcome := &core.ActionOutcome{
		ActionType: core.ATSpell, // type not important for processing
		TargetID:   targetID,
		ActorID:    0,
		Success:    true,
		Effects:    effects,
	}
	return ce.processActionResults(actor, outcome)
}

func TestEvasion_DexSaveSuccess_OnSuccessHalf_ReducesToZero(t *testing.T) {
	// Setup CE with class features enabled
	ce := NewCombatEngine(&core.SimulationOptions{EnableClassFeatures: true})

	// Attacker: monster (dummy)
	attacker := buildTestMonster(t, 0)

	// Target: Monk with Evasion
	target := buildTestCharacter(t, core.AbilityScores{Dexterity: 16}, 7)
	target.Class.ID = classes.Monk
	target.Class.ClassFeatures.MonkFeatures = &classes.MonkFeatures{HasEvasion: true}

	// Register combatants: attacker id 0, target id 1
	ce.AddCombatant(core.NewCombatantWithInfo(attacker))
	ce.AddCombatant(core.NewCombatantWithInfo(target))

	startHP := target.EntityStateManager.GetCurrentHP()

	// Fabricate a Dex save effect that would normally be half on success
	effects := []core.Effect{{
		Type:       core.EffectDamage,
		Value:      18, // already computed spell damage before resistances
		DamageType: core.DamageFire,
		SaveCtx: &core.SaveContext{
			Ability:   core.AbilityDexterity,
			Success:   true,
			OnSuccess: core.DCOnSuccessHalf,
		},
	}}

	if err := runOutcome(t, ce, attacker, 1, effects); err != nil {
		t.Fatalf("processActionResults: %v", err)
	}

	// Evasion: success on Dex save with half-on-success → 0 damage
	if target.EntityStateManager.GetCurrentHP() != startHP {
		t.Fatalf("expected no damage due to Evasion, hp=%d start=%d", target.EntityStateManager.GetCurrentHP(), startHP)
	}
}

func TestEvasion_DexSaveFailure_HalvesDamage(t *testing.T) {
	ce := NewCombatEngine(&core.SimulationOptions{EnableClassFeatures: true})
	attacker := buildTestMonster(t, 0)
	target := buildTestCharacter(t, core.AbilityScores{Dexterity: 16}, 7)
	target.Class.ID = classes.Rogue
	target.Class.ClassFeatures.RogueFeatures = &classes.RogueFeatures{HasEvasion: true}

	ce.AddCombatant(core.NewCombatantWithInfo(attacker)) // id 0
	ce.AddCombatant(core.NewCombatantWithInfo(target))   // id 1

	startHP := target.EntityStateManager.GetCurrentHP()

	effects := []core.Effect{{
		Type:       core.EffectDamage,
		Value:      20,
		DamageType: core.DamageFire,
		SaveCtx: &core.SaveContext{
			Ability:   core.AbilityDexterity,
			Success:   false,                // failed save
			OnSuccess: core.DCOnSuccessHalf, // this spell/effect halves on success
		},
	}}

	if err := runOutcome(t, ce, attacker, 1, effects); err != nil {
		t.Fatalf("processActionResults: %v", err)
	}

	// Evasion on failure → half damage (20 → 10)
	expectedHP := startHP - 10
	if target.EntityStateManager.GetCurrentHP() != expectedHP {
		t.Fatalf("expected hp %d, got %d", expectedHP, target.EntityStateManager.GetCurrentHP())
	}
}

func TestEvasion_IgnoresNonDexSave(t *testing.T) {
	ce := NewCombatEngine(&core.SimulationOptions{EnableClassFeatures: true})
	attacker := buildTestMonster(t, 0)
	target := buildTestCharacter(t, core.AbilityScores{Dexterity: 16, Constitution: 14}, 7)
	target.Class.ID = classes.Monk
	target.Class.ClassFeatures.MonkFeatures = &classes.MonkFeatures{HasEvasion: true}

	ce.AddCombatant(core.NewCombatantWithInfo(attacker)) // id 0
	ce.AddCombatant(core.NewCombatantWithInfo(target))   // id 1

	startHP := target.EntityStateManager.GetCurrentHP()

	effects := []core.Effect{{
		Type:       core.EffectDamage,
		Value:      12,
		DamageType: core.DamagePoison,
		SaveCtx: &core.SaveContext{
			Ability:   core.AbilityConstitution, // not Dex
			Success:   true,
			OnSuccess: core.DCOnSuccessHalf,
		},
	}}

	if err := runOutcome(t, ce, attacker, 1, effects); err != nil {
		t.Fatalf("processActionResults: %v", err)
	}

	// Evasion should not apply → full 12 damage applied (no resistances in test)
	expectedHP := startHP - 12
	if target.EntityStateManager.GetCurrentHP() != expectedHP {
		t.Fatalf("expected hp %d, got %d", expectedHP, target.EntityStateManager.GetCurrentHP())
	}
}

func TestEvasion_IgnoresAttackRollEffects(t *testing.T) {
	ce := NewCombatEngine(&core.SimulationOptions{EnableClassFeatures: true})
	attacker := buildTestMonster(t, 0)
	target := buildTestCharacter(t, core.AbilityScores{Dexterity: 16}, 7)
	target.Class.ID = classes.Rogue
	target.Class.ClassFeatures.RogueFeatures = &classes.RogueFeatures{HasEvasion: true}

	ce.AddCombatant(core.NewCombatantWithInfo(attacker)) // id 0
	ce.AddCombatant(core.NewCombatantWithInfo(target))   // id 1

	startHP := target.EntityStateManager.GetCurrentHP()

	// No SaveCtx => represents attack roll based damage
	effects := []core.Effect{{
		Type:       core.EffectDamage,
		Value:      7,
		DamageType: core.DamageForce,
		SaveCtx:    nil,
	}}

	if err := runOutcome(t, ce, attacker, 1, effects); err != nil {
		t.Fatalf("processActionResults: %v", err)
	}

	expectedHP := startHP - 7
	if target.EntityStateManager.GetCurrentHP() != expectedHP {
		t.Fatalf("expected hp %d, got %d", expectedHP, target.EntityStateManager.GetCurrentHP())
	}
}

func TestCombatEngine_ProcessAIRequest_MeleeProducesDamage(t *testing.T) {
	ce := NewCombatEngine(&core.SimulationOptions{})
	// attacker with STR 16 for reasonable modifiers
	attacker := buildTestCharacter(t, core.AbilityScores{Strength: 16, Dexterity: 14}, 5)
	equipSword(t, attacker, core.WSPrimary)
	// Equip offhand to avoid nil secondary lookups during bonus action attempt
	equipSword(t, attacker, core.WSSecondary)
	target := buildTestMonster(t, 0) // AC=0 guarantees hits

	cAtt, cTgt := buildCombatants(attacker, target)
	ce.AddCombatant(cAtt) // id 0
	ce.AddCombatant(cTgt) // id 1

	// Initialize combat tracker before processing
	if err := ce.SetupCombat(); err != nil {
		t.Fatalf("SetupCombat: %v", err)
	}

	// Manually provide a minimal combat context for direct ProcessAIRequest path
	ctx := core.NewCombatContext(&core.SimulationOptions{})
	ctx.TurnOrder = ce.TurnOrder
	ctx.CurrentRound = 1
	ctx.CombatantInfo = map[int]*core.CombatantInfo{
		0: core.NewCombatantInfo(cAtt),
		1: core.NewCombatantInfo(cTgt),
	}
	ce.CombatContext = ctx
	// Ensure the acting entity has the context
	_ = attacker.UpdateAICombatContext(ctx)

	req := &core.AIRequest{
		Actor:      attacker,
		ActorID:    0,
		Target:     target,
		TargetID:   1,
		ActionType: core.ATMelee,
		WeaponSlot: core.WSPrimary,
		Advantage:  core.RollNormal,
		SimOptions: &core.SimulationOptions{},
	}

	// Execute
	if err := ce.ProcessAIRequest(req); err != nil {
		t.Fatalf("ProcessAIRequest: %v", err)
	}
	// Expect target HP reduced below starting 15
	if target.GetHPStatus().GetHP() >= 15 {
		t.Errorf("expected damage applied to target, hp=%d", target.GetHPStatus().GetHP())
	}
}

func TestCombatEngine_SetupCombat_RollsAndSortsInitiative(t *testing.T) {
	ce := NewCombatEngine(&core.SimulationOptions{})
	a := buildTestCharacter(t, core.AbilityScores{Dexterity: 14}, 5)
	b := buildTestMonster(t, 10)
	cA, cB := buildCombatants(a, b)
	ce.AddCombatant(cA)
	ce.AddCombatant(cB)

	if err := ce.SetupCombat(); err != nil {
		t.Fatalf("SetupCombat: %v", err)
	}
	// Expect only the two combatants added
	if len(ce.TurnOrder) != 2 {
		t.Fatalf("TurnOrder len=%d want 2", len(ce.TurnOrder))
	}
	// Verify initiatives are assigned for non-lair entries and ordering for those two is descending
	// Find non-lair indices
	nonLair := make([]int, 0, 2)
	for _, id := range ce.TurnOrder {
		if id != -1 {
			nonLair = append(nonLair, id)
		}
	}
	if len(nonLair) != 2 {
		t.Fatalf("expected 2 non-lair combatants, got %d", len(nonLair))
	}
	initA := ce.Combatants[nonLair[0]].GetInitiative()
	initB := ce.Combatants[nonLair[1]].GetInitiative()
	if initA < initB {
		t.Errorf("non-lair order not sorted desc: %d < %d", initA, initB)
	}
}

func TestCombatEngine_ProcessActionResults_AppliesDamageAndHealing(t *testing.T) {
	ce := NewCombatEngine(&core.SimulationOptions{})
	a := buildTestCharacter(t, core.AbilityScores{Strength: 14}, 5)
	b := buildTestMonster(t, 10)
	cA, cB := buildCombatants(a, b)
	ce.AddCombatant(cA) // id 0
	ce.AddCombatant(cB) // id 1

	// Directly apply a damage outcome
	dmg := 5
	start := b.GetHPStatus().GetHP()
	out := &core.ActionOutcome{ActionType: core.ATDamage, ActorID: 0, TargetID: 1, Effects: []core.Effect{{Type: core.EffectDamage, Value: dmg}}}
	if err := ce.processActionResults(a, out); err != nil {
		t.Fatalf("processActionResults damage: %v", err)
	}
	if b.GetHPStatus().GetHP() != start-dmg {
		t.Errorf("expected hp=%d after %d dmg, got %d", start-dmg, dmg, b.GetHPStatus().GetHP())
	}

	// Now heal 3
	out = &core.ActionOutcome{ActionType: core.ATHeal, ActorID: 0, TargetID: 1, Effects: []core.Effect{{Type: core.EffectHealing, Value: 3}}}
	if err := ce.processActionResults(a, out); err != nil {
		t.Fatalf("processActionResults heal: %v", err)
	}
	if b.GetHPStatus().GetHP() != start-dmg+3 {
		t.Errorf("expected hp=%d after heal, got %d", start-dmg+3, b.GetHPStatus().GetHP())
	}
}
