package weapon

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	"testing"
)

func TestGetAttackModifier_Logic(t *testing.T) {
	longsword := Weapon{
		Name:         "Longsword",
		IsFinesse:    false,
		IsVersatile:  true,
		NumberOfDice: 1,
		Die:          core.D8,
		DamageType:   core.DamageSlashing,
		IsRanged:     false,
	}

	tests := []struct {
		name    string
		as      core.AbilityScores
		clvl    uint8
		prof    bool
		wantMod int
		wantErr bool
	}{
		{
			name: "level 4 proficient with 18 STR",
			as: core.AbilityScores{
				Strength:     18,
				Dexterity:    14,
				Constitution: 16,
				Intelligence: 10,
				Wisdom:       12,
				Charisma:     10,
			},
			clvl:    4,
			prof:    true,
			wantMod: 6,
			wantErr: false,
		},
		{
			name: "level 4 NOT proficient with 18 STR",
			as: core.AbilityScores{
				Strength:     18,
				Dexterity:    14,
				Constitution: 16,
				Intelligence: 10,
				Wisdom:       12,
				Charisma:     10,
			},
			clvl:    4,
			prof:    false,
			wantMod: 4,
			wantErr: false,
		},
		{
			name: "invalid ability score - high",
			as: core.AbilityScores{
				Strength:     35,
				Dexterity:    14,
				Constitution: 16,
				Intelligence: 10,
				Wisdom:       12,
				Charisma:     10,
			},
			clvl:    4,
			prof:    true,
			wantMod: 0,
			wantErr: true,
		},
		{
			name: "invalid ability score - low",
			as: core.AbilityScores{
				Strength:     0,
				Dexterity:    14,
				Constitution: 16,
				Intelligence: 10,
				Wisdom:       12,
				Charisma:     10,
			},
			clvl:    4,
			prof:    true,
			wantMod: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := longsword.GetAttackModifier(&tt.as, tt.clvl, tt.prof)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAttackModifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.wantMod {
				t.Errorf("GetAttackModifier() = %d, want %d", got, tt.wantMod)
			}
		})
	}
}

func TestGetAttackModifier_DB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	dbErr := database.InitDb(&database.InitOpts{EnvPath: "../../.env"})

	if dbErr != nil {
		fmt.Println(dbErr)
		return
	}
	defer database.CloseDb()

	// Define weapon test cases
	weaponTests := []struct {
		id      int
		name    string
		usesSTR bool // false = finesse (uses DEX)
	}{
		{22, "Longsword", true},
		{35, "Longbow", false}, // Ranged but uses DEX
		{23, "Rapier", false},  // Finesse
		{2, "Dagger", false},   // Finesse
	}

	// Define stat scenarios
	statTests := []struct {
		name string
		as   core.AbilityScores
		clvl uint8
		prof bool
	}{
		{
			name: "level 4 - high STR, low DEX, proficient",
			as:   core.AbilityScores{Strength: 18, Dexterity: 10},
			clvl: 4,
			prof: true,
		},
		{
			name: "level 4 - low STR, high DEX, proficient",
			as:   core.AbilityScores{Strength: 10, Dexterity: 18},
			clvl: 4,
			prof: true,
		},
		{
			name: "level 4 - not proficient",
			as:   core.AbilityScores{Strength: 18, Dexterity: 14},
			clvl: 4,
			prof: false,
		},
		{
			name: "level 14 - high STR, low DEX, proficient",
			as:   core.AbilityScores{Strength: 18, Dexterity: 10},
			clvl: 14,
			prof: true,
		},
		{
			name: "level 14 - low STR, high DEX, proficient",
			as:   core.AbilityScores{Strength: 10, Dexterity: 18},
			clvl: 14,
			prof: true,
		},
		{
			name: "level 14 - not proficient",
			as:   core.AbilityScores{Strength: 18, Dexterity: 14},
			clvl: 14,
			prof: false,
		},
	}

	for _, wt := range weaponTests {
		weapon := getTestWeapon(t, wt.id)

		for _, st := range statTests {
			testName := fmt.Sprintf("%s_%s", wt.name, st.name)
			t.Run(testName, func(t *testing.T) {
				got, err := weapon.GetAttackModifier(&st.as, st.clvl, st.prof)
				if err != nil {
					t.Fatalf("GetAttackModifier() error = %v", err)
				}

				// Calculate expected value based on weapon properties
				var abilityMod int
				var modErr error
				if weapon.IsFinesse || weapon.IsRanged {
					// Finesse/ranged weapons use DEX
					abilityMod, modErr = core.GetAbilityScoreModifier(st.as.Dexterity)
				} else {
					// Melee weapons use STR
					abilityMod, modErr = core.GetAbilityScoreModifier(st.as.Strength)
				}

				if modErr != nil {
					t.Fatalf("Failed to calculate ability mod: %v", modErr)
				}

				expectedMod := abilityMod
				if st.prof {
					pb, _ := core.GetCharacterProficiencyBonus(st.clvl)
					expectedMod += pb
				}

				if got != expectedMod {
					t.Errorf("%s with %s: got %d, want %d (ability:%d, prof:%v)",
						weapon.Name, st.name, got, expectedMod, abilityMod, st.prof)
				}
			})
		}
	}
}

func getTestWeapon(t *testing.T, id int) Weapon {
	t.Helper()
	ctx := context.Background()
	weapon, err := QueryWeaponData(ctx, WeaponQueryParams{ID: id})
	if err != nil {
		t.Fatalf("Failed to get weapon %d: %v", id, err)
	}
	return weapon
}
