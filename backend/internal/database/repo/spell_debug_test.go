package repo

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"fmt"
	"testing"
)

func TestDebugSpells(t *testing.T) {
	err := database.InitDb(nil)
	if err != nil {
		t.Skip("DB not available")
	}
	defer database.CloseDb()

	ids := []int{116, 119}
	spellsMap, err := getSpellsByID(context.Background(), ids)
	if err != nil {
		t.Fatalf("Failed to get spells: %v", err)
	}

	fmt.Printf("Got %d spells from DB\n", len(spellsMap))
	for id, s := range spellsMap {
		fmt.Printf("Spell ID: %s, Name: %s, Formulas: %d levels\n", id, s.Name, len(s.Formulas))
		for lvl, formulas := range s.Formulas {
			fmt.Printf("  Level %d: %d formulas\n", lvl, len(formulas))
			for _, f := range formulas {
				fmt.Printf("    %dd%d %s (CastLvl: %d)\n", f.NumberOfDice, f.Die, f.DamageType, f.CastLevel)
			}
		}
	}
}
