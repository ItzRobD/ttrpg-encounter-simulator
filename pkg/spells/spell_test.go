package spells

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	"testing"
)

// TestGetClosestFormulaToLevel_Logic tests the formula selection logic without database
func TestGetClosestFormulaToLevel_Logic(t *testing.T) {
	// Create a test spell with multiple formulas (like Fireball)
	fireball := Spell{
		Name:  "Fireball",
		Level: 3,
		Formulas: map[int]core.CastFormula{
			3: {CastLevel: 3, NumberOfDice: 8, Die: core.D6, AmountToAdd: 0},  // 8d6
			4: {CastLevel: 4, NumberOfDice: 9, Die: core.D6, AmountToAdd: 0},  // 9d6
			5: {CastLevel: 5, NumberOfDice: 10, Die: core.D6, AmountToAdd: 0}, // 10d6
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
			formula, err := fireball.GetClosestFormulaToLevel(tt.castLevel)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetClosestFormulaToLevel(%d) error = %v, wantErr %v", tt.castLevel, err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return // Expected error, test passed
			}

			if formula.NumberOfDice != tt.wantNumDice {
				t.Errorf("GetClosestFormulaToLevel(%d) dice = %d, want %d", tt.castLevel, formula.NumberOfDice, tt.wantNumDice)
			}

			if formula.CastLevel != tt.wantFormulaLvl {
				t.Errorf("GetClosestFormulaToLevel(%d) formula level = %d, want %d", tt.castLevel, formula.CastLevel, tt.wantFormulaLvl)
			}
		})
	}
}

// TestGetAverageDamageAtLevel_Logic tests damage calculation without database
func TestGetAverageDamageAtLevel_Logic(t *testing.T) {
	// Fireball: 8d6 at level 3, +1d6 per level above
	fireball := Spell{
		Name:  "Fireball",
		Level: 3,
		Formulas: map[int]core.CastFormula{
			3: {CastLevel: 3, NumberOfDice: 8, Die: core.D6, AmountToAdd: 0, UseSpellmod: false},
			4: {CastLevel: 4, NumberOfDice: 9, Die: core.D6, AmountToAdd: 0, UseSpellmod: false},
		},
	}

	// Magic Missile: uses spell modifier
	magicMissile := Spell{
		Name:  "Magic Missile",
		Level: 1,
		Formulas: map[int]core.CastFormula{
			1: {CastLevel: 1, NumberOfDice: 1, Die: core.D4, AmountToAdd: 1, UseSpellmod: true},
		},
	}

	tests := []struct {
		name        string
		spell       Spell
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
			avg, formula, err := tt.spell.GetAverageDamageAtLevel(tt.castLevel, tt.spellModDmg)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAverageDamageAtLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if avg != tt.wantAvg {
				t.Errorf("%s: avg = %d, want %d (dice:%dd%d, mod:%d)",
					tt.name, avg, tt.wantAvg, formula.NumberOfDice, formula.Die, tt.spellModDmg)
			}
		})
	}
}

// TestGetFormulaForCantrip_Logic tests cantrip scaling without database
func TestGetFormulaForCantrip_Logic(t *testing.T) {
	// Fire Bolt scales at levels 5, 11, 17
	fireBolt := Spell{
		Name:  "Fire Bolt",
		Level: 0, // Cantrip
		Formulas: map[int]core.CastFormula{
			1:  {CastLevel: 1, NumberOfDice: 1, Die: core.D10, AmountToAdd: 0},
			5:  {CastLevel: 5, NumberOfDice: 2, Die: core.D10, AmountToAdd: 0},
			11: {CastLevel: 11, NumberOfDice: 3, Die: core.D10, AmountToAdd: 0},
			17: {CastLevel: 17, NumberOfDice: 4, Die: core.D10, AmountToAdd: 0},
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
			formula, err := fireBolt.GetFormulaForCantrip(tt.casterLevel)
			if err != nil {
				t.Fatalf("GetFormulaForCantrip(%d) error = %v", tt.casterLevel, err)
			}

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
	fireBolt := Spell{
		Name:  "Fire Bolt",
		Level: 0,
		Formulas: map[int]core.CastFormula{
			1:  {CastLevel: 1, NumberOfDice: 1, Die: core.D10, AmountToAdd: 0, UseSpellmod: false},
			5:  {CastLevel: 5, NumberOfDice: 2, Die: core.D10, AmountToAdd: 0, UseSpellmod: false},
			11: {CastLevel: 11, NumberOfDice: 3, Die: core.D10, AmountToAdd: 0, UseSpellmod: false},
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
		params       SpellQueryParams
		expectSpells int
		checkSpell   func(*testing.T, map[int]Spell)
	}{
		{
			name:         "Query Fireball by ID",
			params:       SpellQueryParams{ID: []int{119}}, // Fireball ID (adjust if different in your DB)
			expectSpells: 1,
			checkSpell: func(t *testing.T, spells map[int]Spell) {
				for _, spell := range spells {
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
			name:         "Query multiple spells by ID",
			params:       SpellQueryParams{ID: []int{119, 2}}, // Fireball + Acid Splash
			expectSpells: 2,
			checkSpell: func(t *testing.T, spells map[int]Spell) {
				if len(spells) != 2 {
					t.Errorf("Expected 2 spells, got %d", len(spells))
				}
			},
		},
		{
			name:         "Query spell by name",
			params:       SpellQueryParams{Name: []string{"Fireball"}},
			expectSpells: 1,
			checkSpell: func(t *testing.T, spells map[int]Spell) {
				for _, spell := range spells {
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
			spells, err := QuerySpellData(ctx, tt.params)
			if err != nil {
				t.Fatalf("QuerySpellData() error = %v", err)
			}

			if len(spells) != tt.expectSpells {
				t.Errorf("Expected %d spells, got %d", tt.expectSpells, len(spells))
			}

			if tt.checkSpell != nil {
				tt.checkSpell(t, spells)
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
		// Adjust these IDs based on your database
		{119, "Fireball", 3, 0, 25, 32},   // 8d6 = 28 avg (25-32 range)
		{119, "Fireball", 5, 0, 32, 40},   // 10d6 = 35 avg
		{2, "Acid Splash", 1, 0, 3, 3},    // Cantrip: 1d6 = 3.5 -> 3 avg (level 1 caster)
		{2, "Acid Splash", 5, 0, 6, 7},    // Cantrip: 2d6 at level 5 (scales)
		{2, "Acid Splash", 11, 0, 10, 11}, // Cantrip: 3d6 at level 11
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
func getTestSpell(t *testing.T, id int) Spell {
	t.Helper()
	ctx := context.Background()
	spells, err := QuerySpellData(ctx, SpellQueryParams{ID: []int{id}})
	if err != nil {
		t.Fatalf("Failed to get spell %d: %v", id, err)
	}
	if len(spells) != 1 {
		t.Fatalf("Expected 1 spell, got %d", len(spells))
	}
	for _, spell := range spells {
		return spell
	}
	t.Fatal("No spell returned")
	return Spell{}
}
