package actor

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/equipment"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"math"
)

func (a *Actor) GetID() core.ID {
	return a.ID
}

func (a *Actor) GetInstanceID() int {
	return a.InstanceID
}

func (a *Actor) SetInstanceID(id int) {
	a.InstanceID = id
}

func (a *Actor) GetEntityType() core.ActorType {
	return a.ActorType
}

func (a *Actor) IsCharacter() bool {
	return a.ActorType == core.ActorTypeCharacter
}

func (a *Actor) IsMonster() bool {
	return a.ActorType == core.ActorTypeMonster
}

func (a *Actor) IsLegendary() bool {
	if a.ActorType == core.ActorTypeMonster {
		return a.Metadata.IsLegendary
	}
	return false
}

func (a *Actor) IsSpellcaster() bool {
	if a.ActorType == core.ActorTypeLair {
		return false
	}

	return a.Metadata.SpellcasterMetadata.IsSpellcaster ||
		a.Metadata.SpellcasterMetadata.IsInnateCaster ||
		a.SpellManager.GetSpellCount() > 0
}

func (a *Actor) IsHealer() bool {
	if a.SpellManager.GetHealingSpellCount() > 0 {
		return true
	}

	for _, act := range a.Actions {
		if act.AverageHeal > 0 {
			if act.Name == string(core.SpecAbilitySecondWind) {
				continue
			}
			return true
		}
	}

	if a.HasFeature(core.SpecAbilityLayOnHands) {
		return true
	}

	return false
}

func (a *Actor) IsLair() bool {
	return a.ActorType == core.ActorTypeLair
}

func (a *Actor) GetIsCustom() bool {
	return a.IsCustom
}

func (a *Actor) GetLevelOrCR() float64 {
	switch a.ActorType {
	case core.ActorTypeCharacter:
		return float64(a.Metadata.Level)
	case core.ActorTypeMonster:
		return a.Metadata.CR
	default:
		return 0.0
	}
}

// GetProficiencyBonus calculates and returns the actor's proficiency bonus based on their level or challenge rating.
func (a *Actor) GetProficiencyBonus() int {
	n := a.GetLevelOrCR()
	return int(math.Ceil(n/4)) + 1
}

func (a *Actor) GetHitDice() map[core.DiceType]int {
	return a.HPConfig.HitDice
}

// GetAbilityProficiencyBonus returns the proficiency bonus for the given ability if the actor is proficient in it, otherwise 0.
func (a *Actor) GetAbilityProficiencyBonus(ability core.Ability) int {
	if a.Abilities.GetIsProficientInAbility(ability) {
		return a.GetProficiencyBonus()
	}
	return 0
}

// GetSpellSaveDC returns the spell save DC for the given ability.
func (a *Actor) GetSpellSaveDC(ability core.Ability) int {
	return 8 + a.GetAbilityProficiencyBonus(ability) + a.Abilities.GetAbilityModifier(ability)
}

// GetSavingThrowBonus returns the saving throw bonus for the given ability.
func (a *Actor) GetSavingThrowBonus(ability core.Ability) int {
	return a.Abilities.GetAbilityModifier(ability) + a.GetAbilityProficiencyBonus(ability)
}

// GetAttackBonus returns the attack bonus for the given ability and proficiency status.
func (a *Actor) GetAttackBonus(ability core.Ability, proficient bool) int {
	if proficient {
		return a.GetAbilityProficiencyBonus(ability) + a.Abilities.GetAbilityModifier(ability)
	}
	return a.Abilities.GetAbilityModifier(ability)
}

func (a *Actor) UpdateActionsFromEquipment() {
	for slot, e := range a.Equipment.Items {
		action := a.MakeWeaponActionFromEquipment(&e, slot)
		if action != nil {
			a.Actions = append(a.Actions, *action)
		}
	}
}

func (a *Actor) UpdateActionsFromSpells() {
	var actions []core.Action
	for _, spellList := range a.SpellManager.HealingSpells {
		for _, spell := range spellList {
			actions = append(actions, a.MakeActionsFromSpell(spell)...)
		}
	}

	for _, spellList := range a.SpellManager.DamageSpells {
		for _, spell := range spellList {
			actions = append(actions, a.MakeActionsFromSpell(spell)...)
		}
	}

	for _, innateList := range a.SpellManager.HealingSpellsInnate {
		for _, innate := range innateList {
			spellActions := a.MakeActionsFromSpell(&innate.Spell)
			for i := range spellActions {
				spellActions[i].IsInnate = true
			}
			actions = append(actions, spellActions...)
		}
	}

	for _, innateList := range a.SpellManager.DamageSpellsInnate {
		for _, innate := range innateList {
			spellActions := a.MakeActionsFromSpell(&innate.Spell)
			for i := range spellActions {
				spellActions[i].IsInnate = true
			}
			actions = append(actions, spellActions...)
		}
	}

	a.SpellActions = actions
}

func (a *Actor) MakeWeaponActionFromEquipment(e *equipment.Equipment, slot equipment_manager.EquipmentSlot) *core.Action {
	if e.Weapon == nil {
		return nil
	}
	w := e.Weapon
	p := w.Properties

	var t core.ActionType
	if w.Properties.IsOnlyRanged {
		t = core.ATRanged
	} else {
		t = core.ATMelee
	}

	var c core.ActionCost
	if slot == equipment_manager.EquipmentSlotSecondary {
		c = core.ActionCost{
			ActivationType: core.ActBonus,
			Value:          1,
		}
	} else {
		c = core.ActionCost{
			ActivationType: core.ActAction,
			Value:          1,
		}
	}

	ability := core.AbilityStrength
	isProficient := true // Weapons in equipment are usually proficient for PCs, or handled by GetAttackBonus
	if p.IsRanged {
		ability = core.AbilityDexterity
	} else if p.IsFinesse {
		str := a.Abilities.GetAbilityModifier(core.AbilityStrength)
		dex := a.Abilities.GetAbilityModifier(core.AbilityDexterity)
		if dex > str {
			ability = core.AbilityDexterity
		}
	}
	ab := a.GetAttackBonus(ability, isProficient)
	dmgMod := a.Abilities.GetAbilityModifier(ability)

	// Add damage bonus from weapon modifiers
	if w.Modifiers.DamageBonus != 0 {
		dmgMod += w.Modifiers.DamageBonus
	}

	// Apply modifier to damage blocks
	damageBlocks := make([]core.DiceBlock, len(w.DamageBlocks))
	copy(damageBlocks, w.DamageBlocks)
	for i := range damageBlocks {
		damageBlocks[i].Modifier = dmgMod
	}

	slotID := 0
	switch slot {
	case equipment_manager.EquipmentSlotPrimary:
		slotID = 1
	case equipment_manager.EquipmentSlotSecondary:
		slotID = 2
	case equipment_manager.EquipmentSlotRanged:
		slotID = 3
	}

	return &core.Action{
		ID:               core.MakeID(fmt.Sprintf("%s:%d", e.ID.String(), slotID)),
		Name:             e.Name,
		ActionType:       t,
		Cost:             c,
		AttackBonus:      ab,
		DiceBlock:        damageBlocks,
		AverageDamage:    a.calculateActionAverageDamage(damageBlocks),
		WeaponModifiers:  &w.Modifiers,
		WeaponProperties: &w.Properties,
	}
}

func (a *Actor) calculateActionAverageDamage(dice []core.DiceBlock) int {
	total := 0
	for _, db := range dice {
		avg, _ := core.GetAverageRoll(db.NumberOfDice, db.Die, db.Modifier)
		total += avg
	}
	return total
}

func (a *Actor) MakeActionsFromSpell(s *spells.Spell) []core.Action {
	var actions []core.Action

	spellMod := a.Abilities.GetAbilityModifier(a.SpellManager.SpellcastingAbility) + a.ProficiencyBonus

	// Handle Cantrips (Level 0)
	if s.Level == 0 {
		level := a.Metadata.Level
		if level == 0 && a.Metadata.CR != 0 {
			level = int(a.Metadata.CR)
		}
		if level == 0 {
			level = 1
		}

		formulaSlice, err := s.GetFormulaForCantrip(level)
		if err != nil {
			fmt.Printf("failed to get formula for cantrip %s: %v\n", s.Name, err)
			return nil
		}

		action := a.createSpellAction(s, formulaSlice, 0, spellMod)
		actions = append(actions, action)
		return actions
	}

	// Handle Leveled Spells
	for lvl, formulaSlice := range s.Formulas {
		// Only generate actions for levels the actor has slots for
		if count, ok := a.StateManager.MaxSlots[lvl]; !ok || count == 0 {
			continue
		}

		action := a.createSpellAction(s, formulaSlice, lvl, spellMod)
		actions = append(actions, action)
	}

	return actions
}

func (a *Actor) createSpellAction(s *spells.Spell, formulas []core.CastFormula, castLvl int, spellMod int) core.Action {
	var damageBlocks []core.DiceBlock
	var totalAvg int

	for _, f := range formulas {
		modifier := f.AmountToAdd
		if f.UseSpellmod {
			modifier += spellMod
		}

		db := core.MakeDiceBlock(f.NumberOfDice, f.Die, f.DamageType, modifier)
		damageBlocks = append(damageBlocks, db)

		avg, _ := core.GetAverageRoll(f.NumberOfDice, f.Die, modifier)
		totalAvg += avg
	}

	costType := core.ActAction
	switch s.CastingTime {
	case core.CastingTimeBonus:
		costType = core.ActBonus
	case core.CastingTimeReaction:
		costType = core.ActReaction
	}

	action := core.Action{
		ID:         core.MakeID(fmt.Sprintf("%s:%d", s.ID.String(), castLvl)),
		Name:       s.Name,
		ActionType: core.ATSpell,
		Cost: core.ActionCost{
			ActivationType: costType,
			Value:          1,
		},
		DiceBlock:     damageBlocks,
		AttackBonus:   spellMod,
		IsAutoHit:     s.IsAutoHit,
		HasDC:         s.HasDC,
		DCSaveDC:      a.GetSpellSaveDC(a.SpellManager.SpellcastingAbility),
		DCAbility:     s.SpellDC.Ability,
		CastLevel:     castLvl,
		AverageDamage: 0,
		AverageHeal:   0,
	}

	if castLvl > 0 {
		action.Name = fmt.Sprintf("%s (%d)", s.Name, castLvl)
	}

	if s.SpellType == core.STHealing {
		action.AverageHeal = totalAvg
	}
	if s.SpellType == core.STDamage {
		action.AverageDamage = totalAvg
	}

	return action
}

func (a *Actor) UpdateOffensiveValues() {
	var maxSingleHit float64
	var bestTurnDamage float64
	actionDmgMap := make(map[core.ID]float64)

	// A single action per turn
	for _, act := range a.Actions {
		avg := float64(act.AverageDamage)
		actionDmgMap[act.ID] = avg
		if avg > maxSingleHit {
			maxSingleHit = avg
		}
		// Initialize bestTurnDamage with the strongest single action
		if avg > bestTurnDamage {
			bestTurnDamage = avg
		}
	}

	// Multiattacks - used instead of single actions
	for _, act := range a.Actions {
		if act.Multiattack != nil {
			turnDmg := 0.0
			for _, ma := range act.Multiattack {
				turnDmg += actionDmgMap[ma.ActionID] * float64(ma.Count)
			}
			if turnDmg > bestTurnDamage {
				bestTurnDamage = turnDmg
			}
		}
	}

	// Spells - used instead of single action
	for _, act := range a.SpellActions {
		if act.CastLevel > 0 {
			if act.IsInnate {
				if a.StateManager.InnateCurrent[act.Name] <= 0 {
					continue
				}
			} else if a.StateManager.CurrentSlots[act.CastLevel] <= 0 {
				continue
			}
		}
		avg := float64(act.AverageDamage)
		if avg > maxSingleHit {
			maxSingleHit = avg
		}
		if avg > bestTurnDamage {
			bestTurnDamage = avg
		}
	}

	a.Metadata.HighestOffensiveValue = maxSingleHit
	a.Metadata.AverageOffensiveValue = bestTurnDamage
}
