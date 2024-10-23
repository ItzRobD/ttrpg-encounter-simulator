package main

import (
	"context"
	database "dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/simulation"
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

	//var err error
	//
	//ctx := context.Background()

	//bonus, err := shared.GetMonsterProficiencyBonus(26)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(bonus)

	//modifier, err := shared.GetAbilityScoreModifier(12)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(modifier)

	//var result weapon.Weapon
	//params := weapon.WeaponQueryParams{Name: "Warhammer"}
	////params := weapon.WeaponQueryParams{ID: 7}
	//result, err = weapon.QueryWeaponData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Print("Weapon: ", result)

	//var result class.Class
	////params := class.ClassQueryParams{Name: "Artificer"}
	//params := class.ClassQueryParams{ID: 13}
	//result, err = class.QueryClassData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Print("Class: ", result)
	//for k, v := range result.Spellcasting.MaxSpellSlots {
	//	fmt.Printf("Level %d: Slots: %v\n", k, v)
	//}

	var err error
	ctx := context.Background()
	//var m monster.Monster
	params := monster.MonsterQueryParams{ID: 1}
	//params := monster.MonsterQueryParams{Name: "Archmage"}
	m, err := monster.QueryMonsterData(ctx, params)
	if err != nil {
		fmt.Println(err)
	}

	params = monster.MonsterQueryParams{Name: "Adult Brass Dragon"}
	m2, err := monster.QueryMonsterData(ctx, params)
	if err != nil {
		fmt.Println(err)
	}

	//fmt.Println("Monster:")
	//printStructFields(result, "MonsterBase")

	//var sResult spells.Spell
	////params := spells.SpellQueryParams{Name: "Eldritch Blast", Level: 5}
	//params := spells.SpellQueryParams{ID: 40, Level: 4}
	//sResult, err = spells.QuerySpellData(ctx, params)
	////sResult, err = spells.GetSpellByID(ctx, 119, 6)
	//if err != nil {
	//	fmt.Printf("error %w", err)
	//}
	//fmt.Println("Spell:")
	//printStructFields(sResult, "")

	//var sResult spells.Spell
	//params := spells.SpellQueryParams{ID: 30, Level: 5}
	//sResult, err = spells.QuerySpellData(ctx, params)
	//if err != nil {
	//	fmt.Printf("error %w", err)
	//}
	//printStructFields(sResult, "")

	//var cs []spells.Spell
	//cs, err = spells.GetUsableSpellSliceByClassID(ctx, 12)
	//if err != nil {
	//	fmt.Printf("error %w", err)
	//}
	//fmt.Println("Spells:")
	//for _, s := range cs {
	//	printStructFields(s, "")
	//}

	//ctx := context.Background()

	var as shared.AbilityScores
	as.Strength = 12
	as.Dexterity = 18
	as.Constitution = 16
	as.Intelligence = 10
	as.Wisdom = 14
	as.Charisma = 18

	var hp shared.PlayerHP
	hp.HP = 72
	hp.MaxHP = 84

	c, err := character.New(ctx, "Frank", 2, 5, as, hp)
	if err != nil {
		fmt.Println(err)
	}

	err = c.AddSRDArmor(ctx, 7)
	if err != nil {
		fmt.Println(err)
	}
	err = c.AddSRDWeapon(ctx, 7, "primary")
	if err != nil {
		fmt.Println(err)
	}
	err = c.AddSRDWeapon(ctx, 10, "secondary")
	if err != nil {
		fmt.Println(err)
	}
	err = c.AddSRDWeapon(ctx, 12, "ranged")
	if err != nil {
		fmt.Println(err)
	}
	//err = c.AddCustomWeapon("Bat", true, 1, 20, "bludgeoning", true, "ranged")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//err = c.AddCustomArmor("Dragonmail", 17, true, true, 12)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//err = c.AddKnownSpell(ctx, 30)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//fmt.Println(c)

	options := simulation.Options{
		Prioritization: simulation.NoPriority,
	}
	s := simulation.New(options)
	s.Encounter.AddPartyMember(&c)
	s.Encounter.AddMonster(&m)
	s.Encounter.AddMonster(&m2)

	err = s.Simulate()
	if err != nil {
		fmt.Println(err)
	}

	//s.PrintSimulationLog()

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
