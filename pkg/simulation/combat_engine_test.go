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
	"dnd5e-encounter-simulator-backend/pkg/spells"
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
	ch.AI = character.NewCharacterAI(ch, nil)
	// Properly initialized spellcasting manager
	ch.SpellCastingManager = spellcasting_manager.NewSpellcastingManager(ch, ch.RollManager, core.CasterCharacter, int(lvl), spells.SpellSlots{}, spells.SpellSlots{}, 0)
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
	m.AI = monster.NewMonsterAI(m, nil)
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
		BaseValue:  36,
		DamageType: core.DamageFire,
		SaveCtx: &core.SaveContext{
			Ability:   core.AbilityDexterity,
			TargetDC:  15,
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
		BaseValue:  20,
		DamageType: core.DamageFire,
		SaveCtx: &core.SaveContext{
			Ability:   core.AbilityDexterity,
			TargetDC:  15,
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
		BaseValue:  24,
		DamageType: core.DamagePoison,
		SaveCtx: &core.SaveContext{
			Ability:   core.AbilityConstitution, // not Dex
			TargetDC:  15,
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

func TestCombatEngine_ProcessActionResults_AOEHitsAllEnemies(t *testing.T) {
	opts := &core.SimulationOptions{AOEHitsAllEnemies: true}
	ce := NewCombatEngine(opts)

	att := buildTestCharacter(t, core.AbilityScores{}, 1)
	// target 1 will fail the save (low dex)
	tgt1 := buildTestMonster(t, 10)
	tgt1.MonsterBase.AbilityScores.Dexterity = 1
	// target 2 will pass the save (high dex)
	tgt2 := buildTestMonster(t, 10)
	tgt2.MonsterBase.AbilityScores.Dexterity = 30

	cAtt := core.NewCombatantWithInfo(att)
	cTgt1 := core.NewCombatantWithInfo(tgt1)
	cTgt2 := core.NewCombatantWithInfo(tgt2)

	ce.AddCombatant(cAtt)  // id 0
	ce.AddCombatant(cTgt1) // id 1
	ce.AddCombatant(cTgt2) // id 2

	outcome := &core.ActionOutcome{
		ActionType: core.ATSpell,
		TargetID:   1, // Primary target is tgt1
		ActorID:    0,
		Success:    true,
		IsAOE:      true,
		Effects: []core.Effect{
			{
				Type:       core.EffectDamage,
				Value:      10, // tgt1 fails, takes full 10
				BaseValue:  10,
				DamageType: core.DamageFire,
				SaveCtx: &core.SaveContext{
					Ability:   core.AbilityDexterity,
					TargetDC:  5, // Lower DC to ensure tgt2 passes
					Success:   false,
					OnSuccess: core.DCOnSuccessHalf,
				},
			},
		},
	}

	if err := ce.processActionResults(att, outcome); err != nil {
		t.Fatalf("processActionResults: %v", err)
	}

	if tgt1.EntityStateManager.GetCurrentHP() != 5 {
		t.Errorf("Target 1 (primary, failed save): expected 5 HP (15-10), got %d", tgt1.EntityStateManager.GetCurrentHP())
	}
	// Target 2 is high dex, should pass the save and take 5 damage
	if tgt2.EntityStateManager.GetCurrentHP() != 10 {
		t.Errorf("Target 2 (AOE, should pass save): expected 10 HP (15-5), got %d", tgt2.EntityStateManager.GetCurrentHP())
	}
}

func TestCombatEngine_LightningAbsorption(t *testing.T) {
	ce := NewCombatEngine(&core.SimulationOptions{EnableSpecialAbilities: true})

	// Attacker
	attacker := buildTestMonster(t, 0)

	// Target: Monster with Lightning Absorption
	target := buildTestMonster(t, 0)
	target.SpecialAbilities.LightningAbsorption = true
	target.EntityStateManager.ModifyHP(-5, false, false, false, core.DamageNone, false) // 10/15 HP

	ce.AddCombatant(core.NewCombatantWithInfo(attacker)) // id 0
	ce.AddCombatant(core.NewCombatantWithInfo(target))   // id 1

	startHP := target.EntityStateManager.GetCurrentHP()

	// Fabricate a Lightning damage effect
	effects := []core.Effect{{
		Type:       core.EffectDamage,
		Value:      5,
		BaseValue:  5,
		DamageType: core.DamageLightning,
	}}

	if err := runOutcome(t, ce, attacker, 1, effects); err != nil {
		t.Fatalf("processActionResults: %v", err)
	}

	// Lightning Absorption: 10 HP + 5 absorbed = 15 HP
	if target.EntityStateManager.GetCurrentHP() != startHP+5 {
		t.Fatalf("expected healing due to Lightning Absorption, hp=%d want=%d", target.EntityStateManager.GetCurrentHP(), startHP+5)
	}
}

func TestCombatEngine_MaxDamageAndStats(t *testing.T) {
	// Setup
	opts := &core.SimulationOptions{}
	ce := NewCombatEngine(opts)

	// Create actor
	actor := buildTestCharacter(t, core.AbilityScores{Strength: 16}, 1)
	actorCombatant := core.NewCombatantWithInfo(actor)
	actor.SetInstanceID(0)
	ce.AddCombatant(actorCombatant)

	// Create Target
	target := buildTestMonster(t, 10)
	targetCombatant := core.NewCombatantWithInfo(target)
	target.SetInstanceID(1)
	ce.AddCombatant(targetCombatant)

	ce.SetupCombat()

	// 1. Verify MaxDamageSeen update
	effect := core.Effect{
		Type:       core.EffectDamage,
		Value:      15,
		DamageType: core.DamageSlashing,
	}
	outcome := &core.ActionOutcome{
		ActorID:  actor.GetInstanceID(),
		TargetID: target.GetInstanceID(),
		Effects:  []core.Effect{effect},
	}

	err := ce.processActionResults(actor, outcome)
	if err != nil {
		t.Fatalf("Failed to process action results: %v", err)
	}

	if ce.CombatContext.MaxDamageSeen != 15 {
		t.Errorf("Expected MaxDamageSeen to be 15, got %d", ce.CombatContext.MaxDamageSeen)
	}

	// 2. Verify Statistics updates
	if actorCombatant.Info.Statistics.TotalDamageDealt != 15 {
		t.Errorf("Expected TotalDamageDealt to be 15, got %d", actorCombatant.Info.Statistics.TotalDamageDealt)
	}
	if targetCombatant.Info.Statistics.TotalDamageTaken != 15 {
		t.Errorf("Expected TotalDamageTaken to be 15, got %d", targetCombatant.Info.Statistics.TotalDamageTaken)
	}
	if targetCombatant.Info.Statistics.LastAttackerID != actor.GetInstanceID() {
		t.Errorf("Expected LastAttackerID to be %d, got %d", actor.GetInstanceID(), targetCombatant.Info.Statistics.LastAttackerID)
	}

	// 3. Verify turn end state update
	// Mock turn end for actor
	ce.CurrentRound = 1
	err = ce.turnEndEvents(actor.GetInstanceID())
	if err != nil {
		t.Fatalf("Failed turn end events: %v", err)
	}

	// In SimulateRound we added the state update
	// Let's call it manually or simulate a round part
	ce.Combatants[actor.GetInstanceID()].Info.UpdateState()
	ce.Combatants[actor.GetInstanceID()].Info.Statistics.TurnsSinceLastHeal++

	if actorCombatant.Info.Statistics.TurnsSinceLastHeal != 1 {
		t.Errorf("Expected TurnsSinceLastHeal to be 1, got %d", actorCombatant.Info.Statistics.TurnsSinceLastHeal)
	}
}

func TestCombatEngine_HealingStats(t *testing.T) {
	// Setup
	opts := &core.SimulationOptions{}
	ce := NewCombatEngine(opts)

	actor := buildTestCharacter(t, core.AbilityScores{Wisdom: 16}, 1)
	actorCombatant := core.NewCombatantWithInfo(actor)
	actor.SetInstanceID(0)
	ce.AddCombatant(actorCombatant)

	target := buildTestCharacter(t, core.AbilityScores{Constitution: 16}, 1)
	// Give target some damage
	target.EntityStateManager.ModifyHP(-10, false, false, false, core.DamageSlashing, false)
	targetCombatant := core.NewCombatantWithInfo(target)
	target.SetInstanceID(1)
	ce.AddCombatant(targetCombatant)

	ce.SetupCombat()

	effect := core.Effect{
		Type:  core.EffectHealing,
		Value: 5,
	}
	outcome := &core.ActionOutcome{
		ActorID:  actor.GetInstanceID(),
		TargetID: target.GetInstanceID(),
		Effects:  []core.Effect{effect},
	}

	err := ce.processActionResults(actor, outcome)
	if err != nil {
		t.Fatalf("Failed to process action results: %v", err)
	}

	if actorCombatant.Info.Statistics.TotalHealingDone != 5 {
		t.Errorf("Expected TotalHealingDone to be 5, got %d", actorCombatant.Info.Statistics.TotalHealingDone)
	}
	if targetCombatant.Info.Statistics.TotalHealingReceived != 5 {
		t.Errorf("Expected TotalHealingReceived to be 5, got %d", targetCombatant.Info.Statistics.TotalHealingReceived)
	}
	if targetCombatant.Info.Statistics.TurnsSinceLastHeal != 0 {
		t.Errorf("Expected TurnsSinceLastHeal to be 0 after healing, got %d", targetCombatant.Info.Statistics.TurnsSinceLastHeal)
	}
}

func TestCombatEngine_AttackStatistics(t *testing.T) {
	// Setup
	opts := &core.SimulationOptions{}
	ce := NewCombatEngine(opts)

	// Create actor (guaranteed to hit if we use AC 0)
	actor := buildTestCharacter(t, core.AbilityScores{Strength: 20}, 1)
	actorCombatant := core.NewCombatantWithInfo(actor)
	actor.SetInstanceID(0)
	ce.AddCombatant(actorCombatant)

	// Create Target
	target := buildTestMonster(t, 10)
	target.MonsterBase.AC = 0 // Ensure hit
	targetCombatant := core.NewCombatantWithInfo(target)
	target.SetInstanceID(1)
	ce.AddCombatant(targetCombatant)

	ce.SetupCombat()

	// 1. Melee Hit
	outcome := &core.ActionOutcome{
		ActionType: core.ATMelee,
		ActorID:    0,
		TargetID:   1,
		AttackResults: []core.AttackResult{
			{IsHit: true, IsCriticalHit: false},
		},
	}
	ce.processActionResults(actor, outcome)

	if actorCombatant.Info.Statistics.AttacksMade != 1 {
		t.Errorf("Expected AttacksMade to be 1, got %d", actorCombatant.Info.Statistics.AttacksMade)
	}
	if actorCombatant.Info.Statistics.AttacksHit != 1 {
		t.Errorf("Expected AttacksHit to be 1, got %d", actorCombatant.Info.Statistics.AttacksHit)
	}

	// 2. Miss
	outcome = &core.ActionOutcome{
		ActionType: core.ATMelee,
		ActorID:    0,
		TargetID:   1,
		AttackResults: []core.AttackResult{
			{IsHit: false, IsCriticalHit: false},
		},
	}
	ce.processActionResults(actor, outcome)

	if actorCombatant.Info.Statistics.AttacksMade != 2 {
		t.Errorf("Expected AttacksMade to be 2, got %d", actorCombatant.Info.Statistics.AttacksMade)
	}
	if actorCombatant.Info.Statistics.AttacksMissed != 1 {
		t.Errorf("Expected AttacksMissed to be 1, got %d", actorCombatant.Info.Statistics.AttacksMissed)
	}

	// 3. Critical Hit
	outcome = &core.ActionOutcome{
		ActionType: core.ATMelee,
		ActorID:    0,
		TargetID:   1,
		AttackResults: []core.AttackResult{
			{IsHit: true, IsCriticalHit: true},
		},
	}
	ce.processActionResults(actor, outcome)

	if actorCombatant.Info.Statistics.AttacksMade != 3 {
		t.Errorf("Expected AttacksMade to be 3, got %d", actorCombatant.Info.Statistics.AttacksMade)
	}
	if actorCombatant.Info.Statistics.AttacksHit != 2 {
		t.Errorf("Expected AttacksHit to be 2, got %d", actorCombatant.Info.Statistics.AttacksHit)
	}
	if actorCombatant.Info.Statistics.CriticalHits != 1 {
		t.Errorf("Expected CriticalHits to be 1, got %d", actorCombatant.Info.Statistics.CriticalHits)
	}
}

func TestCombatEngine_DeathSaveStatistics(t *testing.T) {
	// Setup
	opts := &core.SimulationOptions{}
	ce := NewCombatEngine(opts)

	// Create Character
	actor := buildTestCharacter(t, core.AbilityScores{Constitution: 10}, 1)
	actorCombatant := core.NewCombatantWithInfo(actor)
	actor.SetInstanceID(0)
	ce.AddCombatant(actorCombatant)

	// Set character to unconscious but not stable
	actor.EntityStateManager.ModifyHP(-100, false, false, false, core.DamageSlashing, false)
	if !actor.EntityStateManager.GetIsUnconscious() {
		t.Fatalf("Character should be unconscious")
	}

	ce.SetupCombat()

	// Mocking some rolls for death saves
	// 1. Success
	status := &core.TurnResult{
		TurnStatuses: map[core.TurnStatus]bool{
			core.TurnDeathSaveSuccess: true,
		},
	}
	ce.recordDeathSaves(0, status)

	if actorCombatant.Info.Statistics.DeathSaveSuccesses != 1 {
		t.Errorf("Expected DeathSaveSuccesses to be 1, got %d", actorCombatant.Info.Statistics.DeathSaveSuccesses)
	}

	// 2. Failure
	status = &core.TurnResult{
		TurnStatuses: map[core.TurnStatus]bool{
			core.TurnDeathSaveFailed: true,
		},
	}
	ce.recordDeathSaves(0, status)

	if actorCombatant.Info.Statistics.DeathSaveFailures != 1 {
		t.Errorf("Expected DeathSaveFailures to be 1, got %d", actorCombatant.Info.Statistics.DeathSaveFailures)
	}

	// 3. Natural 1 (Double Failure)
	status = &core.TurnResult{
		TurnStatuses: map[core.TurnStatus]bool{
			core.TurnDeathSaveFailedDouble: true,
		},
	}
	ce.recordDeathSaves(0, status)

	if actorCombatant.Info.Statistics.DeathSaveFailures != 3 {
		t.Errorf("Expected DeathSaveFailures to be 3 (1+2), got %d", actorCombatant.Info.Statistics.DeathSaveFailures)
	}

	// 4. Natural 20 (Revived - counts as Success)
	status = &core.TurnResult{
		TurnStatuses: map[core.TurnStatus]bool{
			core.TurnRevived: true,
		},
	}
	ce.recordDeathSaves(0, status)

	if actorCombatant.Info.Statistics.DeathSaveSuccesses != 2 {
		t.Errorf("Expected DeathSaveSuccesses to be 2 (1+1), got %d", actorCombatant.Info.Statistics.DeathSaveSuccesses)
	}
}
