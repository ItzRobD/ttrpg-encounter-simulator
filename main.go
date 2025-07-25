package main

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
)

func main() {
	dbErr := database.InitDb()

	if dbErr != nil {
		fmt.Println(dbErr)
		return
	}
	defer database.CloseDb()

	ctx := context.Background()
	var ids []int

	for i := 1; i <= 319; i++ {
		ids = append(ids, i)
	}

	params := spells.SpellQueryParams{Name: []string{"Fireball", "Acid Splash"}}
	s, err := spells.QuerySpellData(ctx, params)
	if err != nil {
		fmt.Println(err)

	}
	fmt.Println(s)
}
