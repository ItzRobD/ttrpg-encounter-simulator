package main

import (
	"context"
	database "dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/weapon"
	"fmt"
)

func main() {
	database.InitDb()
	defer database.CloseDb()

	//// Example query using the armorData package
	//queryParam := armor.ArmorQueryParams{Name: "Breastplate"}
	//armorData, err := database.QueryArmorDataByName(queryParam)
	//if err != nil {
	//	log.Fatalf("Error querying armorData data: %v", err)
	//}
	//fmt.Printf("Armor: %+v\n", armorData)

	var err error

	ctx := context.Background()

	//var result armor.Armor
	//params := armor.ArmorQueryParams{Name: "Breastplate"}
	//result, err = armor.QueryArmorData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Print("Armor: ", result, "")

	var result weapon.Weapon
	params := weapon.WeaponQueryParams{ID: 7}
	result, err = weapon.QueryWeaponData(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print("Weapon: ", result)
}
