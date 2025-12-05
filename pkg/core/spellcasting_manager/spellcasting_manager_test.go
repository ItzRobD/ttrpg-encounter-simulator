package spellcasting_manager

import (
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"testing"
)

// Test simple logic functions that don't need complex setup

func TestIsSaveSuccessful(t *testing.T) {
	// Create minimal manager (doesn't need most fields for this test)
	scm := &SpellcastingManager{}

	tests := []struct {
		name string
		roll int
		dc   int
		want bool
	}{
		{"exact match passes", 15, 15, true},
		{"one above passes", 16, 15, true},
		{"one below fails", 14, 15, false},
		{"nat 20 vs DC 25", 20, 25, false}, // RAW: nat 20 doesn't auto-succeed saves
		{"nat 1 vs DC 5", 1, 5, false},
		{"high roll high DC", 25, 30, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scm.IsSaveSuccessful(tt.roll, tt.dc)
			if got != tt.want {
				t.Errorf("IsSaveSuccessful(%d, %d) = %v, want %v", tt.roll, tt.dc, got, tt.want)
			}
		})
	}
}

func TestHasSpellSlotsAtLevel(t *testing.T) {
	tests := []struct {
		name  string
		slots spells.SpellSlots
		level int
		want  bool
	}{
		{
			name:  "has level 1 slots",
			slots: spells.SpellSlots{1: 4, 2: 3},
			level: 1,
			want:  true,
		},
		{
			name:  "no level 3 slots",
			slots: spells.SpellSlots{1: 4, 2: 3},
			level: 3,
			want:  false,
		},
		{
			name:  "used all level 1 slots",
			slots: spells.SpellSlots{1: 0, 2: 3},
			level: 1,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scm := &SpellcastingManager{
				currentSlots: tt.slots,
			}

			got := scm.HasSpellSlotsAtLevel(tt.level)
			if got != tt.want {
				t.Errorf("HasSpellSlotsAtLevel(%d) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestHasAnySpellSlots(t *testing.T) {
	tests := []struct {
		name  string
		slots spells.SpellSlots
		want  bool
	}{
		{
			name:  "has slots",
			slots: spells.SpellSlots{1: 4},
			want:  true,
		},
		{
			name:  "no slots at all",
			slots: spells.SpellSlots{},
			want:  false,
		},
		{
			name:  "has high level slots only",
			slots: spells.SpellSlots{9: 1},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scm := &SpellcastingManager{
				currentSlots: tt.slots,
			}

			got := scm.HasAnySpellSlots()
			if got != tt.want {
				t.Errorf("HasAnySpellSlots() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpendSpellSlot(t *testing.T) {
	tests := []struct {
		name         string
		initialSlots spells.SpellSlots
		slotLevel    int
		wantSlots    spells.SpellSlots
		wantErr      bool
	}{
		{
			name:         "expend level 1 slot",
			initialSlots: spells.SpellSlots{1: 4, 2: 3},
			slotLevel:    1,
			wantSlots:    spells.SpellSlots{1: 3, 2: 3},
			wantErr:      false,
		},
		{
			name:         "expend last slot",
			initialSlots: spells.SpellSlots{1: 1},
			slotLevel:    1,
			wantSlots:    spells.SpellSlots{1: 0},
			wantErr:      false,
		},
		{
			name:         "cannot expend when none available",
			initialSlots: spells.SpellSlots{1: 0},
			slotLevel:    1,
			wantSlots:    spells.SpellSlots{1: 0},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scm := &SpellcastingManager{
				currentSlots: tt.initialSlots,
			}

			err := scm.ExpendSpellSlot(tt.slotLevel)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExpendSpellSlot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Compare maps by checking each level
				for level, wantCount := range tt.wantSlots {
					if scm.currentSlots[level] != wantCount {
						t.Errorf("After ExpendSpellSlot(%d): level %d slots = %d, want %d",
							tt.slotLevel, level, scm.currentSlots[level], wantCount)
					}
				}
			}
		})
	}
}

func TestAddKnownSpell(t *testing.T) {
	tests := []struct {
		name            string
		spell           *spells.Spell
		wantDamageCount int
		wantHealCount   int
		wantErr         bool
	}{
		{
			name: "add damage spell",
			spell: &spells.Spell{
				ID:        1,
				Name:      "Fireball",
				Level:     3,
				SpellType: core.STDamage,
				Formulas: map[int]core.CastFormula{
					3: {NumberOfDice: 8, Die: core.D6},
				},
			},
			wantDamageCount: 1,
			wantHealCount:   0,
			wantErr:         false,
		},
		{
			name: "add healing spell",
			spell: &spells.Spell{
				ID:        2,
				Name:      "Cure Wounds",
				Level:     1,
				SpellType: core.STHealing,
				Formulas: map[int]core.CastFormula{
					1: {NumberOfDice: 1, Die: core.D8, AmountToAdd: 0, UseSpellmod: true},
				},
			},
			wantDamageCount: 0,
			wantHealCount:   1,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scm := &SpellcastingManager{
				healingSpells: map[int][]*spells.Spell{},
				damageSpells:  map[int][]*spells.Spell{},
			}

			err := scm.AddKnownSpell(tt.spell)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddKnownSpell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if scm.damageSpellCount != tt.wantDamageCount {
				t.Errorf("damageSpellCount = %d, want %d", scm.damageSpellCount, tt.wantDamageCount)
			}

			if scm.healingSpellCount != tt.wantHealCount {
				t.Errorf("healingSpellCount = %d, want %d", scm.healingSpellCount, tt.wantHealCount)
			}
		})
	}
}

func TestGetHighestAvailableSpellSlot(t *testing.T) {
	tests := []struct {
		name      string
		slots     spells.SpellSlots
		wantLevel int
		wantErr   bool
	}{
		{
			name:      "has level 9 slots",
			slots:     spells.SpellSlots{1: 4, 9: 1},
			wantLevel: 9,
			wantErr:   false,
		},
		{
			name:      "only has low level slots",
			slots:     spells.SpellSlots{1: 4, 2: 3},
			wantLevel: 2,
			wantErr:   false,
		},
		{
			name:      "no slots available",
			slots:     spells.SpellSlots{},
			wantLevel: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scm := &SpellcastingManager{
				currentSlots: tt.slots,
			}

			got, err := scm.getHighestAvailableSpellSlot()

			if (err != nil) != tt.wantErr {
				t.Errorf("getHighestAvailableSpellSlot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantLevel {
				t.Errorf("getHighestAvailableSpellSlot() = %d, want %d", got, tt.wantLevel)
			}
		})
	}
}

// Test calculateFormulaAverages modifies spell formulas correctly
func TestCalculateFormulaAverages(t *testing.T) {
	scm := &SpellcastingManager{}

	spell := &spells.Spell{
		Name:  "Fireball",
		Level: 3,
		Formulas: map[int]core.CastFormula{
			3: {NumberOfDice: 8, Die: core.D6, AmountToAdd: 0}, // Should be 28
			4: {NumberOfDice: 9, Die: core.D6, AmountToAdd: 0}, // Should be 31
			5: {NumberOfDice: 1, Die: core.D4, AmountToAdd: 1}, // Should be 3
		},
	}

	scm.calculateFormulaAverages(spell)

	tests := []struct {
		level       int
		wantAverage int
	}{
		{3, 28}, // 8 * 3.5 = 28
		{4, 31}, // 9 * 3.5 = 31.5 -> 31
		{5, 3},  // 1 * 2.5 + 1 = 3.5 -> 3
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			formula := spell.Formulas[tt.level]
			if formula.AverageValue != tt.wantAverage {
				t.Errorf("Formula level %d: average = %d, want %d",
					tt.level, formula.AverageValue, tt.wantAverage)
			}
		})
	}
}
