package actor

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/equipment_manager"
	"dnd5e-encounter-simulator-backend/pkg/core/state_manager"
	"dnd5e-encounter-simulator-backend/pkg/equipment"
	"dnd5e-encounter-simulator-backend/pkg_old/classes"
	"dnd5e-encounter-simulator-backend/pkg_old/races"
	"testing"
)

func TestFeatureProcessor_StaticFeatures(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *Actor
		verify func(*testing.T, *Actor)
	}{
		{
			name: "Dwarven Resilience adds Poison Resistance",
			setup: func() *Actor {
				return &Actor{
					Metadata:    Metadata{RaceID: races.RaceID(core.Dwarf)},
					Features:    []core.Feature{{Name: core.SpecAbilityDwarvenResilience}},
					Resistances: core.NewDamageResistances(),
				}
			},
			verify: func(t *testing.T, a *Actor) {
				if a.Resistances.GetResistanceType(core.DamagePoison) != core.ResistanceResistant {
					t.Errorf("Expected poison resistance")
				}
			},
		},
		{
			name: "Hellish Resistance adds Fire Resistance",
			setup: func() *Actor {
				return &Actor{
					Metadata:    Metadata{RaceID: races.RaceID(core.Tiefling)},
					Features:    []core.Feature{{Name: core.SpecAbilityHellishResistance}},
					Resistances: core.NewDamageResistances(),
				}
			},
			verify: func(t *testing.T, a *Actor) {
				if a.Resistances.GetResistanceType(core.DamageFire) != core.ResistanceResistant {
					t.Errorf("Expected fire resistance")
				}
			},
		},
		{
			name: "Draconic Resistance adds configured Damage Type",
			setup: func() *Actor {
				return &Actor{
					Metadata: Metadata{RaceID: races.RaceID(core.Dragonborn)},
					Features: []core.Feature{
						{
							Name: core.SpecAbilityDraconicResistance,
							Data: core.FeatureData{DamageType: []core.DamageType{core.DamageAcid}},
						},
					},
					Resistances: core.NewDamageResistances(),
				}
			},
			verify: func(t *testing.T, a *Actor) {
				if a.Resistances.GetResistanceType(core.DamageAcid) != core.ResistanceResistant {
					t.Errorf("Expected acid resistance")
				}
			},
		},
		{
			name: "Fighting Style: Defense adds +1 AC while wearing armor",
			setup: func() *Actor {
				a := &Actor{
					Abilities: core.Abilities{AbilityScores: core.AbilityScores{Dexterity: 10}}, // +0
					Features:  []core.Feature{{Name: core.SpecAbilityFightingStyleDef}},
					Equipment: equipment_manager.NewEquipmentManager(),
				}
				// Mock wearing armor
				a.Equipment.AddItem(equipment_manager.EquipmentSlotArmor, equipment.Equipment{
					Type:  equipment.EquipmentTypeArmor,
					Armor: &equipment.Armor{ArmorClass: 10, DexBonus: true},
				})
				return a
			},
			verify: func(t *testing.T, a *Actor) {
				// Base AC 10 + 0 Dex + 1 style = 11
				if a.AC != 11 {
					t.Errorf("Expected AC 11, got %d", a.AC)
				}
			},
		},
		{
			name: "Lay on Hands initializes resource pool",
			setup: func() *Actor {
				return &Actor{
					Metadata:     Metadata{Level: 5, ClassID: classes.ClassID(core.Paladin)},
					Features:     []core.Feature{{Name: core.SpecAbilityLayOnHands}},
					StateManager: state_manager.StateManager{Resource: make(map[string]int)},
				}
			},
			verify: func(t *testing.T, a *Actor) {
				expected := 25 // 5 * level
				if a.StateManager.Resource[string(core.SpecAbilityLayOnHands)] != expected {
					t.Errorf("Expected Lay on Hands pool %d, got %d", expected, a.StateManager.Resource[string(core.SpecAbilityLayOnHands)])
				}
			},
		},
		{
			name: "Indomitable initializes resource count",
			setup: func() *Actor {
				return &Actor{
					Features: []core.Feature{
						{
							Name: core.SpecAbilityIndomitable,
							Data: core.FeatureData{Value: 2},
						},
					},
					StateManager: state_manager.StateManager{Resource: make(map[string]int)},
				}
			},
			verify: func(t *testing.T, a *Actor) {
				if a.StateManager.Resource[string(core.SpecAbilityIndomitable)] != 2 {
					t.Errorf("Expected Indomitable uses 2, got %d", a.StateManager.Resource[string(core.SpecAbilityIndomitable)])
				}
			},
		},
		{
			name: "Legendary Resistance initializes resource count",
			setup: func() *Actor {
				return &Actor{
					Features: []core.Feature{
						{
							Name: core.SpecAbilityLegendaryResistance,
							Data: core.FeatureData{Value: 3},
						},
					},
					StateManager: state_manager.StateManager{Resource: make(map[string]int)},
				}
			},
			verify: func(t *testing.T, a *Actor) {
				if a.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)] != 3 {
					t.Errorf("Expected Legendary Resistance uses 3, got %d", a.StateManager.Resource[string(core.SpecAbilityLegendaryResistance)])
				}
			},
		},
		{
			name: "Second Wind initializes resource count",
			setup: func() *Actor {
				return &Actor{
					Features:     []core.Feature{{Name: core.SpecAbilitySecondWind}},
					StateManager: state_manager.StateManager{Resource: make(map[string]int)},
				}
			},
			verify: func(t *testing.T, a *Actor) {
				if a.StateManager.Resource[string(core.SpecAbilitySecondWind)] != 1 {
					t.Errorf("Expected Second Wind uses 1, got %d", a.StateManager.Resource[string(core.SpecAbilitySecondWind)])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.setup()
			a.ProcessFeatures()
			tt.verify(t, a)
		})
	}
}
