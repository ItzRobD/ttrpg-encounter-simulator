package main

import (
	"context"
	database "dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/monster"
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
	//
	//var result weapon.Weapon
	//params := weapon.WeaponQueryParams{Name: "Warhammer"}
	//result, err = weapon.QueryWeaponData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Print("Weapon: ", result)

	//var result class.Class
	//params := class.ClassQueryParams{Name: "Barbarian"}
	//result, err = class.QueryClassData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Print("Class: ", result)

	var result monster.MonsterBase
	params := monster.MonsterQueryParams{ID: 5}
	result, err = monster.QueryMonsterData(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print("MonsterBase: ", result)
}
