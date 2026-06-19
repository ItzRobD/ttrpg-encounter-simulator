package actor

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"strings"
)

// ProcessFeatures applies logic-based modifications to the actor based on their features,
// class, and race. This handles things that are difficult to represent purely in a database.
func (a *Actor) ProcessFeatures() {
	// 1. Handle Racial Features
	a.processRacialFeatures()

	// 2. Handle Class Features
	a.processClassFeatures()

	// 3. Handle Passive/Static Modifiers (like AC calculations)
	a.calculateAC()
}

func (a *Actor) processRacialFeatures() {
	switch core.RaceID(a.Metadata.RaceID) {
	case core.Dwarf:
		if a.HasFeature(core.SpecAbilityDwarvenResilience) {
			a.Resistances.SetResistanceType(core.DamagePoison, core.ResistanceResistant)
		}
	case core.Dragonborn:
		// Dragonborn Resistance
		if a.HasFeature(core.SpecAbilityDraconicResistance) {
			for _, f := range a.Features {
				if f.Name == core.SpecAbilityDraconicResistance {
					if len(f.Data.DamageType) > 0 {
						a.Resistances.SetResistanceType(f.Data.DamageType[0], core.ResistanceResistant)
					}
				}
			}
		}

		// Dragonborn Breath Weapon
		for _, f := range a.Features {
			if f.Name == core.SpecAbilityBreathWeapon {
				dmgType := core.DamageFire // Default
				if len(f.Data.DamageType) > 0 {
					dmgType = f.Data.DamageType[0]
				}

				a.AddAction(core.Action{
					Name:       string(core.SpecAbilityBreathWeapon),
					ActionType: core.ATAction,
					Cost:       core.ActionCost{ActivationType: core.ActAction, Value: 1},
					DiceBlock: []core.DiceBlock{
						{
							NumberOfDice: f.Data.NumberOfDice,
							Die:          f.Data.Die,
							DamageType:   dmgType,
						},
					},
					HasDC:         true,
					DCSaveDC:      8 + a.Abilities.GetAbilityModifier(core.AbilityConstitution) + a.ProficiencyBonus,
					DCAbility:     core.AbilityConstitution,
					RechargeValue: 7, // Recharges on 7+ (impossible on D6), marks as limited resource
					IsAOE:         true,
				})
				// Initialize as charged
				a.StateManager.Resource[string(core.SpecAbilityBreathWeapon)] = 1
			}
		}
	case core.Tiefling:
		if a.HasFeature(core.SpecAbilityHellishResistance) {
			a.Resistances.SetResistanceType(core.DamageFire, core.ResistanceResistant)
		}
	}
}

func (a *Actor) processClassFeatures() {
	// Resource Initializations
	for _, f := range a.Features {
		// Indomitable
		if f.Name == core.SpecAbilityIndomitable {
			if _, ok := a.StateManager.Resource[string(core.SpecAbilityIndomitable)]; !ok {
				a.StateManager.Resource[string(core.SpecAbilityIndomitable)] = f.Data.Value
			}
		}

		// Lay on Hands
		if f.Name == core.SpecAbilityLayOnHands {
			// Pool is 5 * level
			level := a.Metadata.Level
			if level == 0 && a.Metadata.CR != 0 {
				level = int(a.Metadata.CR)
			}
			pool := level * 5
			a.StateManager.Resource[string(core.SpecAbilityLayOnHands)] = pool
		}

		// Legendary Resistance
		if f.Name == core.SpecAbilityLegendaryResistance {
			if _, ok := a.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)]; !ok {
				a.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] = f.Data.Value
			}
		}

		// Second Wind
		if f.Name == core.SpecAbilitySecondWind {
			if _, ok := a.StateManager.Resource[string(core.SpecAbilitySecondWind)]; !ok {
				a.StateManager.Resource[string(core.SpecAbilitySecondWind)] = 1
			}
		}

		// Arcane Recovery
		if f.Name == core.SpecAbilityArcaneRecovery {
			if _, ok := a.StateManager.Resource[string(core.SpecAbilityArcaneRecovery)]; !ok {
				a.StateManager.Resource[string(core.SpecAbilityArcaneRecovery)] = 1
			}
		}

		// TODO: I added short rests -> need to look through everything that uses short rests, features are complete
		// TODO: Eldritch blast uses multiple targets at higher levels
		// TODO: Druid wild shape
		// TODO: Monk ki points can now be added
		// TODO: Add below to ED
		// Action Surge
		if f.Name == core.SpecAbilityActionSurge {
			if _, ok := a.StateManager.Resource[string(core.SpecAbilityActionSurge)]; !ok {
				a.StateManager.Resource[string(core.SpecAbilityActionSurge)] = f.Data.Value
			}
		}

		// Stroke of Luck
		if f.Name == core.SpecAbilityStrokeOfLuck {
			if _, ok := a.StateManager.Resource[string(core.SpecAbilityStrokeOfLuck)]; !ok {
				a.StateManager.Resource[string(core.SpecAbilityStrokeOfLuck)] = 1
			}
		}
	}
}

func (a *Actor) calculateAC() {
	baseAC := 10 + a.Abilities.GetAbilityModifier(core.AbilityDexterity)

	// Check for Armor
	armor := a.Equipment.GetItem(equipment_manager.EquipmentSlotArmor)
	shield := a.Equipment.GetItem(equipment_manager.EquipmentSlotShield)

	if armor != nil && armor.Armor != nil {
		baseAC = armor.Armor.ArmorClass
		if armor.Armor.DexBonus {
			dexMod := a.Abilities.GetAbilityModifier(core.AbilityDexterity)
			if armor.Armor.MaxBonus && dexMod > 2 {
				dexMod = 2
			}
			baseAC += dexMod
		}
		baseAC += armor.Armor.Modifier
	} else {
		// Unarmored Defense
		if core.ClassID(a.Metadata.ClassID) == core.Barbarian {
			// Barbarian: 10 + Dex + Con
			baseAC = 10 + a.Abilities.GetAbilityModifier(core.AbilityDexterity) + a.Abilities.GetAbilityModifier(core.AbilityConstitution)
		} else if core.ClassID(a.Metadata.ClassID) == core.Monk && shield == nil {
			// Monk: 10 + Dex + Wis
			baseAC = 10 + a.Abilities.GetAbilityModifier(core.AbilityDexterity) + a.Abilities.GetAbilityModifier(core.AbilityWisdom)
		}
	}

	if shield != nil && shield.Armor != nil {
		baseAC += shield.Armor.ArmorClass
	}

	// Fighting Style: Defense (+1 AC while wearing armor)
	if armor != nil && a.HasFeature(core.SpecAbilityFightingStyleDef) {
		baseAC += 1
	}

	a.AC = baseAC
}

func (a *Actor) HasFeature(ability core.SpecialAbility) bool {
	for _, f := range a.Features {
		fName := string(f.Name)
		targetName := string(ability)
		if strings.EqualFold(fName, targetName) || strings.Contains(strings.ToLower(fName), strings.ToLower(targetName)) {
			return true
		}
	}
	return false
}

func (a *Actor) AddAction(action core.Action) {
	// Check if action already exists to avoid duplicates
	for _, existing := range a.Actions {
		if existing.Name == action.Name {
			return
		}
	}
	a.Actions = append(a.Actions, action)
}
