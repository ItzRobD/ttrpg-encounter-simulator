package main

import (
	"context"
	database "dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/character"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/monster"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/simulation"
	"fmt"
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
	//m, err := monster.QueryMonsterData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//}

	m, err := monster.NewSRDMonster(ctx, params, core.EntityModifiers{InitiativeAdvantage: shared.RollNormal, InitiativeBonus: 0})
	if err != nil {
		fmt.Println(err)
	}

	//params = monster.MonsterQueryParams{Name: "Adult Brass Dragon"}
	//m2, err := monster.QueryMonsterData(ctx, params)
	//if err != nil {
	//	fmt.Println(err)
	//}

	//fmt.Println("Monster:")
	//helpers.PrintStructFields(m2, "MonsterBase")

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
	hp.HP = 20
	hp.MaxHP = 84

	weaponProf := character.WeaponProficiencies{
		Primary:   true,
		Secondary: true,
		Ranged:    true,
	}

	entityModifiers := core.EntityModifiers{
		InitiativeAdvantage: shared.RollNormal,
		InitiativeBonus:     0,
		UseVersatileAttacks: true,
	}

	c, err := character.New(ctx, "Frank", 13, 10, as, hp, shared.APPreferMelee, shared.SPNoPreference, entityModifiers)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(c)
	c.SetWeaponProficiencies(weaponProf)

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

	err = c.AddKnownSpell(ctx, 116) // fire bolt
	if err != nil {
		fmt.Println(err)
	}
	err = c.AddKnownSpell(ctx, 35) // burning hands
	if err != nil {
		fmt.Println(err)
	}
	//err = c.AddKnownSpell(ctx, 70) // cure wounds
	//if err != nil {
	//	fmt.Println(err)
	//}
	err = c.AddKnownSpell(ctx, 119) // fireball
	if err != nil {
		fmt.Println(err)
	}
	//var f *spells.CastFormula
	//f, err2 := spells.GetSpellFormulaByLevel(ctx, 119, 3)
	//if err2 != nil {
	//	fmt.Println(err2)
	//}
	//c.KnownSpells[2].CastFormula = *f
	//fmt.Println(c.KnownSpells[2].CastFormula)

	//for _, s := range c.KnownSpells {
	//	fmt.Println(s.CastFormula)
	//}

	//
	//fmt.Println(c)

	//fmt.Println("Spell slots:")
	//fmt.Println(c.GetSpellSlots())
	//fmt.Println("Max spell slots:")
	//fmt.Println(c.Class.Spellcasting.MaxSpellSlots)

	options := simulation.Options{
		CanMonstersCrit:         true,
		CanPlayersCrit:          true,
		HasIncreasedCrits:       false,
		AllowPlayerHeals:        true,
		AllowMonsterHeals:       true,
		TargetPriority:          shared.NoPriority,
		HealPriority:            shared.PrioritizeMostDamaged,
		ActionPreference:        shared.APNoPreference,
		AOEHitsAllEnemies:       false,
		PlayerHealThresholdPct:  50,
		MonsterHealThresholdPct: 50,
	}
	s := simulation.New(options)
	s.Encounter.AddPartyMember(&c)
	s.Encounter.AddMonster(m)
	//s.Encounter.AddMonster(&m2)

	err = s.Simulate(10)
	if err != nil {
		fmt.Println(err)
	}

	//s.PrintSimulationLog()

}
