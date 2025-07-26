package main

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/monster"
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

	//params := spells.SpellQueryParams{Name: []string{"Fireball", "Acid Splash"}}
	//s, err := spells.QuerySpellData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//
	//}
	//fmt.Println(s)

	params := monster.MonsterQueryParams{ID: []int{1, 5}}
	c, err := monster.QueryMonsterConfigData(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(c)
}
