package main

import (
	"context"
	database "dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/spell"
	"fmt"
	"reflect"
)

func main() {
	dbErr := database.InitDb()

	if dbErr != nil {
		fmt.Println(dbErr)
		return
	}

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

	//var result monster.Monster
	//params := monster.MonsterQueryParams{ID: 5}
	////params := monster.MonsterQueryParams{Name: "Adult Brass Dragon"}
	////params := monster.MonsterQueryParams{Name: "barbed devil"}
	//result, err = monster.QueryMonsterData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println("Monster:")
	//printStructFields(result, "MonsterBase")

	//var sResult spell.Spell
	//params := spell.SpellQueryParams{Name: "Eldritch Blast", Level: 5}
	////params := spell.SpellQueryParams{ID: 40, Level: 4}
	//sResult, err = spell.QuerySpellData(ctx, params)
	////sResult, err = spell.GetSpellByID(ctx, 119, 6)
	//if err != nil {
	//	fmt.Printf("error %w", err)
	//}
	//fmt.Println("Spell:")
	//printStructFields(sResult, "")

	spells, err := spell.GetSpellsUsableByClassID(ctx, 12)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Spells:")
	fmt.Println(spells)
	for _, s := range spells {
		printStructFields(s, "")
	}
}

func printStructFields(v interface{}, prefix string) {
	val := reflect.ValueOf(v)
	typ := val.Type()

	// Handle pointer to struct
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	// Ensure the input is a struct
	if typ.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := typ.Field(i)
			fieldName := fieldType.Name

			// Check if the field is an embedded struct
			if fieldType.Anonymous {
				// Recursive call for embedded struct
				printStructFields(field.Interface(), prefix+fieldName+".")
			} else {
				fmt.Printf("%s%s: %v\n", prefix, fieldName, field.Interface())
			}
		}
	} else {
		fmt.Println("Provided value is not a struct")
	}
}
