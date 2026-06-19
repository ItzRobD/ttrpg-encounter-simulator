package spells_test

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/internal/database/repo"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"testing"
)

// TestGetClosestFormulaToLevel_Logic tests the formula selection logic without database
func TestGetClosestFormulaToLevel_Logic(t *testing.T) {
	// Create a test spell with multiple formulas (like Fireball)
	fireball := spells.Spell{
		Name:  "Fireball",
		Level: 3,
		Formulas: map[int][]core.CastFormula{
			3: {{CastLevel: 3, NumberOfDice: 8, Die: core.D6, AmountToAdd: 0}},  // 8d6
			4: {{CastLevel: 4, NumberOfDice: 9, Die: core.D6, AmountToAdd: 0}},  // 9d6
			5: {{CastLevel: 5, NumberOfDice: 10, Die: core.D6, AmountToAdd: 0}}, // 10d6
		},
	}

	tests := []struct {
		name           string
		castLevel      int
		wantNumDice    int
		wantFormulaLvl int
		wantErr        bool
	}{
		{"cast at minimum level", 3, 8, 3, false},
		{"cast at level 4", 4, 9, 4, false},
		{"cast at level 5", 5, 10, 5, false},
		{"upcast to level 6 (uses level 5 formula)", 6, 10, 5, false},
		{"upcast to level 9 (uses level 5 formula)", 9, 10, 5, false},
		{"below minimum level", 2, 0, 0, true},
		{"level 0 (cantrips use different logic)", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formulas, err := fireball.GetClosestFormulaToLevel(tt.castLevel)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetClosestFormulaToLevel(%d) error = %v, wantErr %v", tt.castLevel, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return // Expected error, test passed
			}

			formula := formulas[0]
			if formula.NumberOfDice != tt.wantNumDice {
				t.Errorf("GetClosestFormulaToLevel(%d) dice = %d, want %d", tt.castLevel, formula.NumberOfDice, tt.wantNumDice)
			}

			if formula.CastLevel != tt.wantFormulaLvl {
				t.Errorf("GetClosestFormulaToLevel(%d) formula level = %d, want %d", tt.castLevel, formula.CastLevel, tt.wantFormulaLvl)
			}
		})
	}
}

func TestGetHighestAverageAmount(t *testing.T) {
	tests := []struct {
		name     string
		formulas map[int][]core.CastFormula
		want     int
	}{
		{
			name:     "No formulas",
			formulas: nil,
			want:     0,
		},
		{
			name: "Single formula",
			formulas: map[int][]core.CastFormula{
				1: {{AverageValue: 10}},
			},
			want: 10,
		},
		{
			name: "Multiple levels",
			formulas: map[int][]core.CastFormula{
				1: {{AverageValue: 10}},
				2: {{AverageValue: 20}},
			},
			want: 20,
		},
		{
			name: "Multiple formulas at highest level",
			formulas: map[int][]core.CastFormula{
				1: {{AverageValue: 10}},
				2: {
					{AverageValue: 15},
					{AverageValue: 10},
				},
			},
			want: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &spells.Spell{Formulas: tt.formulas}
			if got := s.GetHighestAverageAmount(); got != tt.want {
				t.Errorf("GetHighestAverageAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetAverageDamageAtLevel_Logic tests damage calculation without database
func TestGetAverageDamageAtLevel_Logic(t *testing.T) {
	// Fireball: 8d6 at level 3, +1d6 per level above
	fireball := spells.Spell{
		Name:  "Fireball",
		Level: 3,
		Formulas: map[int][]core.CastFormula{
			3: {{CastLevel: 3, NumberOfDice: 8, Die: core.D6, AmountToAdd: 0, UseSpellmod: false}},
			4: {{CastLevel: 4, NumberOfDice: 9, Die: core.D6, AmountToAdd: 0, UseSpellmod: false}},
		},
	}

	// Magic Missile: uses spell modifier
	magicMissile := spells.Spell{
		Name:  "Magic Missile",
		Level: 1,
		Formulas: map[int][]core.CastFormula{
			1: {{CastLevel: 1, NumberOfDice: 1, Die: core.D4, AmountToAdd: 1, UseSpellmod: true}},
		},
	}

	tests := []struct {
		name        string
		spell       spells.Spell
		castLevel   int
		spellModDmg int
		wantAvg     int
		wantErr     bool
	}{
		// Fireball tests (d6 avg = 3.5, floor to 3)
		{"Fireball level 3", fireball, 3, 0, 28, false},                    // 8 * 3.5 = 28
		{"Fireball level 4", fireball, 4, 4, 31, false},                    // 9 * 3.5 = 31.5 -> 31
		{"Fireball upcasted to 5 (uses lvl 4)", fireball, 5, 0, 31, false}, // Uses closest formula
		{"Fireball below min level", fireball, 2, 0, 0, true},

		// Magic Missile tests (d4 avg = 2.5, floor to 2)
		{"Magic Missile no modifier", magicMissile, 1, 0, 3, false}, // 1*2.5 + 1 = 3.5 -> 3
		{"Magic Missile +4 modifier", magicMissile, 1, 4, 7, false}, // 1*2.5 + 1 + 4 = 7.5 -> 7
		{"Magic Missile +5 modifier", magicMissile, 1, 5, 8, false}, // 1*2.5 + 1 + 5 = 8.5 -> 8
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avg, formulas, err := tt.spell.GetAverageDamageAtLevel(tt.castLevel, tt.spellModDmg)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAverageDamageAtLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if avg != tt.wantAvg {
				formula := formulas[0]
				t.Errorf("%s: avg = %d, want %d (dice:%dd%d, mod:%d)",
					tt.name, avg, tt.wantAvg, formula.NumberOfDice, formula.Die, tt.spellModDmg)
			}
		})
	}
}

// TestGetFormulaForCantrip_Logic tests cantrip scaling without database
func TestGetFormulaForCantrip_Logic(t *testing.T) {
	// Fire Bolt scales at levels 5, 11, 17
	fireBolt := spells.Spell{
		Name:  "Fire Bolt",
		Level: 0, // Cantrip
		Formulas: map[int][]core.CastFormula{
			1:  {{CastLevel: 1, NumberOfDice: 1, Die: core.D10, AmountToAdd: 0}},
			5:  {{CastLevel: 5, NumberOfDice: 2, Die: core.D10, AmountToAdd: 0}},
			11: {{CastLevel: 11, NumberOfDice: 3, Die: core.D10, AmountToAdd: 0}},
			17: {{CastLevel: 17, NumberOfDice: 4, Die: core.D10, AmountToAdd: 0}},
		},
	}

	tests := []struct {
		name        string
		casterLevel int
		wantDice    int
		wantDie     core.DiceType
	}{
		{"level 1 character", 1, 1, core.D10},
		{"level 4 character", 4, 1, core.D10},
		{"level 5 character", 5, 2, core.D10},
		{"level 10 character", 10, 2, core.D10},
		{"level 11 character", 11, 3, core.D10},
		{"level 16 character", 16, 3, core.D10},
		{"level 17 character", 17, 4, core.D10},
		{"level 20 character", 20, 4, core.D10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formulas, err := fireBolt.GetFormulaForCantrip(tt.casterLevel)
			if err != nil {
				t.Fatalf("GetFormulaForCantrip(%d) error = %v", tt.casterLevel, err)
			}

			formula := formulas[0]
			if formula.NumberOfDice != tt.wantDice {
				t.Errorf("level %d: dice = %d, want %d", tt.casterLevel, formula.NumberOfDice, tt.wantDice)
			}

			if formula.Die != tt.wantDie {
				t.Errorf("level %d: die = %v, want %v", tt.casterLevel, formula.Die, tt.wantDie)
			}
		})
	}
}

// TestGetAverageDamageCantrip_Logic tests cantrip damage calculation
func TestGetAverageDamageCantrip_Logic(t *testing.T) {
	fireBolt := spells.Spell{
		Name:  "Fire Bolt",
		Level: 0,
		Formulas: map[int][]core.CastFormula{
			1:  {{CastLevel: 1, NumberOfDice: 1, Die: core.D10, AmountToAdd: 0, UseSpellmod: false}},
			5:  {{CastLevel: 5, NumberOfDice: 2, Die: core.D10, AmountToAdd: 0, UseSpellmod: false}},
			11: {{CastLevel: 11, NumberOfDice: 3, Die: core.D10, AmountToAdd: 0, UseSpellmod: false}},
		},
	}

	tests := []struct {
		name        string
		casterLevel int
		spellModDmg int
		wantAvg     int
	}{
		{"level 1 wizard", 1, 0, 5},    // 1d10 avg = 5.5 -> 5
		{"level 5 wizard", 5, 0, 11},   // 2d10 avg = 11
		{"level 11 wizard", 11, 0, 16}, // 3d10 avg = 16.5 -> 16
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avg, _, err := fireBolt.GetAverageDamageCantrip(tt.casterLevel, tt.spellModDmg)
			if err != nil {
				t.Fatalf("GetAverageDamageCantrip() error = %v", err)
			}

			if avg != tt.wantAvg {
				t.Errorf("level %d: avg = %d, want %d", tt.casterLevel, avg, tt.wantAvg)
			}
		})
	}
}

// Integration tests with database
func TestQuerySpellData_DB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	dbErr := database.InitDb(&database.InitOpts{EnvPath: "../../.env"})
	if dbErr != nil {
		t.Fatalf("Failed to init database: %v", dbErr)
	}
	defer database.CloseDb()

	tests := []struct {
		name         string
		params       spells.SpellQueryParams
		expectSpells int
		checkSpell   func(*testing.T, map[core.ID]spells.Spell)
	}{
		{
			name:         "Query Fireball by id",
			params:       spells.SpellQueryParams{ID: []int{119}}, // Fireball id (adjust if different in your DB)
			expectSpells: 1,
			checkSpell: func(t *testing.T, spellMap map[core.ID]spells.Spell) {
				for _, spell := range spellMap {
					if spell.Name != "Fireball" {
						t.Errorf("Expected Fireball, got %s", spell.Name)
					}
					if spell.Level != 3 {
						t.Errorf("Fireball should be level 3, got %d", spell.Level)
					}
					if !spell.IsAOE {
						t.Error("Fireball should be AOE")
					}
				}
			},
		},
		{
			name:         "Query spell by name",
			params:       spells.SpellQueryParams{Name: []string{"Fireball"}},
			expectSpells: 1,
			checkSpell: func(t *testing.T, spellMap map[core.ID]spells.Spell) {
				for _, spell := range spellMap {
					if spell.Name != "Fireball" {
						t.Errorf("Expected Fireball, got %s", spell.Name)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			spellMap, err := repo.QuerySpellData(ctx, tt.params)
			if err != nil {
				t.Fatalf("QuerySpellData() error = %v", err)
			}

			if len(spellMap) != tt.expectSpells {
				t.Errorf("Expected %d spells, got %d", tt.expectSpells, len(spellMap))
			}

			if tt.checkSpell != nil {
				tt.checkSpell(t, spellMap)
			}
		})
	}
}

// TestRealSpellDamageCalculations tests damage calculations with real spell data from DB
func TestRealSpellDamageCalculations_DB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	dbErr := database.InitDb(&database.InitOpts{EnvPath: "../../.env"})
	if dbErr != nil {
		t.Fatalf("Failed to init database: %v", dbErr)
	}
	defer database.CloseDb()

	// Test known spells with their expected damage
	tests := []struct {
		spellID     int
		spellName   string
		castLevel   int // For leveled spells: cast level; For cantrips: caster level
		spellModDmg int
		minDamage   int // Expected minimum reasonable damage
		maxDamage   int // Expected maximum reasonable damage
	}{
		{
			spellID:     119,
			spellName:   "Fireball",
			castLevel:   3,
			spellModDmg: 0,
			minDamage:   25,
			maxDamage:   32,
		},
		{
			spellID:     2,
			spellName:   "Acid Splash",
			castLevel:   1,
			spellModDmg: 0,
			minDamage:   3,
			maxDamage:   3,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_level_%d", tt.spellName, tt.castLevel), func(t *testing.T) {
			spell := getTestSpell(t, tt.spellID)

			if spell.Name != tt.spellName {
				t.Fatalf("Expected %s, got %s", tt.spellName, spell.Name)
			}

			var avg int
			var err error

			if spell.Level == 0 {
				// Cantrip
				avg, _, err = spell.GetAverageDamageCantrip(tt.castLevel, tt.spellModDmg)
			} else {
				// Leveled spell
				avg, _, err = spell.GetAverageDamageAtLevel(tt.castLevel, tt.spellModDmg)
			}

			if err != nil {
				t.Fatalf("Damage calculation error: %v", err)
			}

			if avg < tt.minDamage || avg > tt.maxDamage {
				t.Errorf("%s at level %d: avg damage %d outside expected range [%d-%d]",
					spell.Name, tt.castLevel, avg, tt.minDamage, tt.maxDamage)
			}

			t.Logf("%s (level %d): average damage = %d", spell.Name, tt.castLevel, avg)
		})
	}
}

// Helper function to get a spell from DB for testing
func getTestSpell(t *testing.T, id int) spells.Spell {
	t.Helper()
	ctx := context.Background()
	spellMap, err := repo.QuerySpellData(ctx, spells.SpellQueryParams{ID: []int{id}})
	if err != nil {
		t.Fatalf("Failed to get spell %d: %v", id, err)
	}
	if len(spellMap) != 1 {
		t.Fatalf("Expected 1 spell, got %d", len(spellMap))
	}
	for _, spell := range spellMap {
		return spell
	}
	t.Fatal("No spell returned")
	return spells.Spell{}
}
