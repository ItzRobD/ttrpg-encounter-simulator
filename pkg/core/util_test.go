package core

import "testing"

// Test your actual GetAbilityScoreModifier function
func TestGetAbilityScoreModifier(t *testing.T) {
	tests := []struct {
		name    string
		score   int
		wantMod int
		wantErr bool
	}{
		// Valid scores from your main.go (Frank and Jack have 18 STR)
		{"strength 18", 18, 4, false},
		{"dexterity 14", 14, 2, false},
		{"constitution 16", 16, 3, false},
		{"intelligence 10", 10, 0, false},
		{"wisdom 12", 12, 1, false},

		// Edge cases
		{"minimum valid score", 1, -5, false},
		{"maximum valid score", 30, 10, false},
		{"average score", 10, 0, false},

		// Invalid scores
		{"score too low", 0, 0, true},
		{"score too high", 31, 0, true},
		{"negative score", -5, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMod, gotErr := GetAbilityScoreModifier(tt.score)

			// Check if error expectation matches
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetAbilityScoreModifier(%d) error = %v, wantErr %v",
					tt.score, gotErr, tt.wantErr)
				return
			}

			// If we expected an error, we're done
			if tt.wantErr {
				return
			}

			// Check if modifier matches
			if gotMod != tt.wantMod {
				t.Errorf("GetAbilityScoreModifier(%d) = %d, want %d",
					tt.score, gotMod, tt.wantMod)
			}
		})
	}
}

// Test proficiency bonus calculation
func TestGetCharacterProficiencyBonus(t *testing.T) {
	tests := []struct {
		name    string
		level   uint8
		want    int
		wantErr bool
	}{
		{"level 1", 1, 2, false},
		{"level 4 (Frank and Jack)", 4, 2, false},
		{"level 5 (tier 2)", 5, 3, false},
		{"level 9 (tier 3)", 9, 4, false},
		{"level 20 (max)", 20, 6, false},
		{"level 0 (invalid)", 0, 0, true},
		{"level 21 (invalid)", 21, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCharacterProficiencyBonus(tt.level)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCharacterProficiencyBonus(%d) error = %v, wantErr %v",
					tt.level, err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("GetCharacterProficiencyBonus(%d) = %d, want %d",
					tt.level, got, tt.want)
			}
		})
	}
}

func TestGetMonsterProficiencyBonus(t *testing.T) {
	tests := []struct {
		name    string
		cr      float64
		want    int
		wantErr bool
	}{
		{"cr 1/8", 1.0 / 8, 2, false},
		{"cr 2", 2, 2, false},
		{"cr 15", 15, 5, false},
		{"cr 30", 30, 9, false},
		{"cr 40", 40, 0, true},
		{"negative cr", -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMonsterProficiencyBonus(tt.cr)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetMonsterProficiencyBonus(%f) error = %v, wantErr %v",
					tt.cr, got, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("GetMonsterProficiencyBonus(%f) = %d, want %d",
					tt.cr, got, tt.want)
			}
		})
	}
}
