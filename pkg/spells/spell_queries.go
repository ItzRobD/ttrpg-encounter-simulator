package spells

import (
	"context"
	"database/sql"
	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/enum"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/internal/util"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"
	. "github.com/go-jet/jet/v2/postgres"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// QuerySpellData retrieves spells based on the provided IDs or names in the params.
// If IDs are provided in the params, it fetches spells directly by those IDs.
// If only names are provided, it resolves the IDs first by matching spell names, then retrieves the spells.
// Returns a map of spell IDs to Spell objects or an error if the query fails.
func QuerySpellData(ctx context.Context, params SpellQueryParams) (map[int]Spell, error) {
	var spells map[int]Spell
	var err error
	if len(params.ID) == 0 {
		ids, err := getSpellIDsByName(ctx, params.Name)
		if err != nil {
			return nil, err
		}
		spells, err = getSpellsByID(ctx, ids)
		if err != nil {
			return nil, err
		}
	} else {
		spells, err = getSpellsByID(ctx, params.ID)
		if err != nil {
			return nil, err
		}
	}

	return spells, nil
}

func GetUsableSpellIDsByClassID(ctx context.Context, classID uint8) ([]int, error) {
	var ids []int
	stmt := SELECT(
		Spells.ID,
	).FROM(
		Spells.
			INNER_JOIN(SpellUsers, Spells.ID.EQ(SpellUsers.SpellID)),
	).WHERE(
		SpellUsers.ClassID.EQ(Int(int64(classID))).
			AND(
				Spells.SpellType.EQ(enum.Stype.Damage).
					OR(Spells.SpellType.EQ(enum.Stype.Healing)),
			),
	)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query spells usable by class id: %d - %w", classID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to collect spells usable by class id: %d - %w", classID, err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func getSpellIDsByName(ctx context.Context, names []string) ([]int, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("no spell names provided")
	}
	var ids []int
	titlized := make([]string, len(names))
	caser := cases.Title(language.English)
	for i, name := range names {
		titlized[i] = caser.String(name)
	}

	stmt := SELECT(Spells.ID).
		FROM(Spells).
		WHERE(Spells.Name.IN(util.StringsToExpressions(titlized)...))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query spells id by name: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to collect spells id by name: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func getSpellsByID(ctx context.Context, ids []int) (map[int]Spell, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no spell ids provided")
	}
	spellMap := make(map[int]Spell)
	stmt := SELECT(
		Spells.ID, // Begin Spell
		Spells.Name,
		Spells.Description,
		Spells.Concentration,
		Spells.CastingTime,
		Spells.Ritual,
		Spells.Level,
		Spells.SpellType,
		Spells.IsAoe,
		Spells.HasDc,
		Spells.APIURL,
		SpellFormulas.LevelType,
		SpellFormulas.FormulaLevel,
		SpellDc.Ability, // Begin SpellDC
		SpellDc.OnSuccess,
		SpellDamage.NumberOfDice, // Begin CastFormula
		SpellDamage.Die,
		SpellDamage.AmountToAdd,
		SpellDamage.UseSpellmod,
		SpellDamage.DamageType,
		// Average value - we calculate per formula
		SpellHeal.NumberOfDice, // healing spell CastFormula
		SpellHeal.Die,
		SpellHeal.AmountToAdd,
		SpellHeal.UseSpellmod,
	).
		FROM(Spells.
			LEFT_JOIN(SpellFormulas, Spells.ID.EQ(SpellFormulas.SpellID)).
			LEFT_JOIN(SpellDc, Spells.ID.EQ(SpellDc.SpellID)).
			LEFT_JOIN(SpellDamage, SpellFormulas.FormulaID.EQ(SpellDamage.SpellFormulaID)).
			LEFT_JOIN(SpellHeal, SpellFormulas.FormulaID.EQ(SpellHeal.SpellFormulaID)),
		).WHERE(
		Spells.ID.IN(util.IntsToExpressions(ids)...),
	)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query spells: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s Spell
		var formula core.CastFormula
		var dcAbility, dcOnSuccess sql.NullString
		var levelType sql.NullString
		var fLevel sql.NullInt64
		var dNumDice, dDie, dAmountToAdd sql.NullInt64
		var hNumDice, hDie, hAmountToAdd sql.NullInt64
		var dUseSpellmod, hUseSpellmod sql.NullBool
		var damageType sql.NullString
		err = rows.Scan(
			&s.ID,
			&s.Name,
			&s.Description,
			&s.IsConcentration,
			&s.CastingTime,
			&s.IsRitual,
			&s.Level,
			&s.SpellType,
			&s.IsAOE,
			&s.HasDC,
			&s.ApiURL,
			&levelType,
			&fLevel,
			&dcAbility,
			&dcOnSuccess,
			&dNumDice,
			&dDie,
			&dAmountToAdd,
			&dUseSpellmod,
			&damageType,
			&hNumDice,
			&hDie,
			&hAmountToAdd,
			&hUseSpellmod)
		if err != nil {
			return nil, fmt.Errorf("failed to collect spells: %w", err)
		}

		// Formula Level
		// This should only be needed if spell type is other/support - features not implemented yet
		if fLevel.Valid {
			formula.CastLevel = int(fLevel.Int64)
		} else {
			formula.CastLevel = -1
		}

		// Level type
		if levelType.Valid {
			s.LevelType = levelType.String
		} else {
			s.LevelType = ""
		}

		// DC Ability
		if dcAbility.Valid {
			s.SpellDC.Ability = core.MakeAbility(dcAbility.String)
		} else {
			s.SpellDC.Ability = core.AbilityNone
		}

		// Dc OnSuccess
		if dcOnSuccess.Valid {
			s.SpellDC.OnSuccess, err = core.MakeDCOnSuccess(dcOnSuccess.String)
			if err != nil {
				return nil, fmt.Errorf("failed to collect spells: invalid dc on success returned: spellID %d", s.ID)
			}
		} else {
			s.SpellDC.OnSuccess = core.DCOnSuccessNone
		}

		// Cast level is set to -1 if it's a support or other spell
		// This is not yet a feature -> formula creations
		if formula.CastLevel != -1 {
			// Number of dice
			switch {
			case dNumDice.Valid:
				formula.NumberOfDice = int(dNumDice.Int64)
			case hNumDice.Valid:
				formula.NumberOfDice = int(hNumDice.Int64)
			default:
				return nil, fmt.Errorf("failed to collect spells: invalid number of dice returned: spellID %d", s.ID)
			}

			// Die type
			switch {
			case dDie.Valid:
				formula.Die = core.DiceType(dDie.Int64)
			case hDie.Valid:
				formula.Die = core.DiceType(hDie.Int64)
			default:
				formula.Die = core.DiceType(0)
			}

			// Amount to add
			switch {
			case dAmountToAdd.Valid:
				formula.AmountToAdd = int(dAmountToAdd.Int64)
			case hAmountToAdd.Valid:
				formula.AmountToAdd = int(hAmountToAdd.Int64)
			default:
				return nil, fmt.Errorf("failed to collect spells: invalid amount to add returned: spellID %d", s.ID)
			}

			// Use spellmod
			switch {
			case dUseSpellmod.Valid:
				formula.UseSpellmod = dUseSpellmod.Bool
			case hUseSpellmod.Valid:
				formula.UseSpellmod = hUseSpellmod.Bool
			default:
				return nil, fmt.Errorf("failed to collect spells: invalid use spellmod returned: spellID %d", s.ID)
			}

			// Damage type
			if damageType.Valid {
				formula.DamageType, err = core.MakeDamageType(damageType.String)
				if err != nil {
					return nil, fmt.Errorf("failed to collect spells: invalid damage type returned: spellID %d", s.ID)
				}
			}

			// Average Value
			avg, err := core.GetAverageRoll(formula.NumberOfDice, formula.Die, formula.AmountToAdd)
			if err != nil {
				return nil, fmt.Errorf("failed to collect spells: unable to calculate average value for spell id: %d - %w", s.ID, err)
			}
			formula.AverageValue = avg
		}

		spell, exists := spellMap[s.ID]
		if exists {
			if spell.Formulas == nil {
				spell.Formulas = make(map[int]core.CastFormula)
			}
			spell.Formulas[formula.CastLevel] = formula
			spellMap[s.ID] = spell
		} else {
			s.Formulas = make(map[int]core.CastFormula)
			s.Formulas[formula.CastLevel] = formula
			spellMap[s.ID] = s
		}
	}

	if errR := rows.Err(); errR != nil {
		return nil, errR
	}

	return spellMap, nil
}

//func getSpellCastLevelsByID(ctx context.Context, id int) ([]int, error) {
//	var levels []int
//	stmt := SELECT(
//		SpellFormulas.FormulaLevel,
//	).FROM(
//		SpellFormulas,
//	).WHERE(SpellFormulas.SpellID.EQ(Int(int64(id))))
//
//	query, args := stmt.Sql()
//	rows, err := database.Query(ctx, query, args...)
//	if err != nil {
//		return levels, fmt.Errorf("failed to query spells cast levels by id: %d - %w", id, err)
//	}
//	defer rows.Close()
//	for rows.Next() {
//		var level int
//		err = pgx.Row.Scan(rows,
//			&level,
//		)
//		if err != nil {
//			return levels, fmt.Errorf("failed to collect spells cast levels by id: %d - %w", id, err)
//		}
//		levels = append(levels, level)
//	}
//	return levels, nil
//}
//
//func getSpellTypeByID(ctx context.Context, id int) (string, error) {
//	var levelType string
//	stmt := SELECT(
//		SpellFormulas.LevelType,
//	).FROM(
//		SpellFormulas,
//	).WHERE(SpellFormulas.SpellID.EQ(Int(int64(id))))
//	query, args := stmt.Sql()
//	row, err := database.QueryRow(ctx, query, args...)
//	if err != nil {
//		return levelType, fmt.Errorf("failed to query spells level type by id: %d - %w", id, err)
//	}
//	err = pgx.Row.Scan(row,
//		&levelType,
//	)
//	if err != nil {
//		return levelType, fmt.Errorf("failed to collect spells level type by id: %d - %w", id, err)
//	}
//	return levelType, nil
//}
//
//func getMinimumSpellLevelByID(ctx context.Context, id int) (int, error) {
//	var level int
//	stmt := SELECT(
//		MIN(Spells.Level),
//	).
//		FROM(
//			Spells,
//		).
//		WHERE(
//			Spells.ID.EQ(Int(int64(id))),
//		)
//
//	query, args := stmt.Sql()
//	row, err := database.QueryRow(ctx, query, args...)
//	if err != nil {
//		return level, fmt.Errorf("failed to query minimum spells level by id: %d - %w", id, err)
//	}
//	err = pgx.Row.Scan(row,
//		&level,
//	)
//	if err != nil {
//		return level, fmt.Errorf("failed to collect minimum spells level by id: %d - %w", id, err)
//	}
//
//	return level, nil
//}
//
//func getMaxFormulaLevelBySpellID(ctx context.Context, spellID, formulaLevel int) (int, error) {
//	var maxFormulaLevel int
//	stmt := SELECT(
//		MAX(SpellFormulas.FormulaLevel),
//	).FROM(
//		SpellFormulas,
//	).WHERE(
//		SpellFormulas.SpellID.EQ(Int(int64(spellID))).
//			AND(SpellFormulas.FormulaLevel.LT_EQ(Int(int64(formulaLevel)))),
//	)
//
//	query, args := stmt.Sql()
//	row, err := database.QueryRow(ctx, query, args...)
//	if err != nil {
//		return 0, fmt.Errorf("failed to query max formula level for spellID: %d - %w", spellID, err)
//	}
//	err = row.Scan(&maxFormulaLevel)
//	if err != nil {
//		return 0, fmt.Errorf("failed to collect max formula level for spellID: %d - %w", spellID, err)
//	}
//
//	return maxFormulaLevel, nil
//}
//
//func getSpellByID(ctx context.Context, id int) (Spell, error) {
//	var spell Spell
//	stmt := SELECT(
//		Spells.AllColumns, SpellFormulas.LevelType,
//		CASE().
//			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
//			THEN(SpellDc.Ability).
//			ELSE(enum.Abilityscore.None),
//		CASE().
//			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
//			THEN(SpellDc.OnSuccess).
//			ELSE(enum.Dcsuccess.None),
//	).
//		FROM(
//			Spells.
//				LEFT_JOIN(SpellDc, Spells.ID.EQ(SpellDc.SpellID)).
//				LEFT_JOIN(SpellFormulas, Spells.ID.EQ(SpellFormulas.SpellID)),
//		).
//		WHERE(
//			Spells.ID.EQ(Int(int64(id))),
//		)
//
//	query, args := stmt.Sql()
//	row, err := database.QueryRow(ctx, query, args...)
//	if err != nil {
//		return spell, fmt.Errorf("failed to query spells: %w", err)
//	}
//	var ability sql.NullString
//	var onSuccess sql.NullString
//	var levelType sql.NullString
//	err = pgx.Row.Scan(row,
//		&spell.ID,
//		&spell.Name,
//		&spell.Description,
//		&spell.IsConcentration,
//		&spell.CastingTime,
//		&spell.IsRitual,
//		&spell.Level,
//		&spell.SpellType,
//		&spell.IsAOE,
//		&spell.HasDC,
//		&spell.ApiURL,
//		&ability,
//		&onSuccess,
//		&levelType,
//	)
//	if err != nil {
//		return spell, fmt.Errorf("failed to collect spells: %w", err)
//	}
//
//	if ability.Valid {
//		spell.SpellDC.Ability = core.Ability(ability.String)
//	} else {
//		spell.SpellDC.Ability = ""
//	}
//	if onSuccess.Valid {
//		spell.SpellDC.OnSuccess, err = core.MakeDCOnSuccess(onSuccess.String)
//		if err != nil {
//			return spell, err
//		}
//	} else {
//		spell.SpellDC.OnSuccess = core.DCOnSuccessNone
//	}
//	if levelType.Valid {
//		spell.LevelType = levelType.String
//	} else {
//		spell.LevelType = ""
//	}
//
//	if levelType.Valid {
//		formulas, err := getSpellFormulas(ctx, id)
//		if err != nil {
//			return spell, err
//		}
//
//		spell.Formulas = make(map[int]core.CastFormula)
//
//		for _, formula := range formulas {
//			spell.Formulas[formula.CastLevel] = formula
//		}
//	} else {
//		spell.Formulas = nil
//	}
//	return spell, nil
//}
//
//func getSpellFormulas(ctx context.Context, spellID int) ([]core.CastFormula, error) {
//	var formulas []core.CastFormula
//	stmt := SELECT(
//		SpellFormulas.FormulaLevel,
//		CASE().
//			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
//			THEN(SpellDamage.NumberOfDice).
//			ELSE(SpellHeal.NumberOfDice),
//		CASE().
//			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
//			THEN(SpellDamage.Die).
//			ELSE(SpellHeal.Die),
//		CASE().
//			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
//			THEN(SpellDamage.AmountToAdd).
//			ELSE(SpellHeal.AmountToAdd),
//		CASE().
//			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
//			THEN(SpellDamage.DamageType).
//			ELSE(enum.Dmg.None),
//		CASE().
//			WHEN(Spells.SpellType.EQ(enum.Stype.Damage)).
//			THEN(SpellDamage.UseSpellmod).
//			ELSE(SpellHeal.UseSpellmod),
//	).FROM(
//		Spells.
//			LEFT_JOIN(SpellFormulas, Spells.ID.EQ(SpellFormulas.SpellID)).
//			LEFT_JOIN(SpellDamage, SpellFormulas.FormulaID.EQ(SpellDamage.SpellFormulaID)).
//			LEFT_JOIN(SpellHeal, SpellFormulas.FormulaID.EQ(SpellHeal.SpellFormulaID)),
//	).WHERE(
//		Spells.ID.EQ(Int(int64(spellID))),
//	).ORDER_BY(
//		SpellFormulas.FormulaLevel.ASC(),
//	)
//
//	query, args := stmt.Sql()
//	rows, err := database.Query(ctx, query, args...)
//	if err != nil {
//		return formulas, fmt.Errorf("failed to query spell formulas: %w", err)
//	}
//	defer rows.Close()
//	for rows.Next() {
//		var formula core.CastFormula
//		err2 := rows.Scan(
//			&formula.CastLevel,
//			&formula.NumberOfDice,
//			&formula.Die,
//			&formula.AmountToAdd,
//			&formula.DamageType,
//			&formula.UseSpellmod,
//		)
//		if err2 != nil {
//			return formulas, fmt.Errorf("failed to collect spell formulas: %w", err)
//		}
//		formulas = append(formulas, formula)
//	}
//	return formulas, nil
//}
//
//func getSpellIDByName(ctx context.Context, name string) (int, error) {
//	var id int
//	stmt := SELECT(Spells.ID).
//		FROM(Spells).
//		WHERE(Spells.Name.EQ(String(name)))
//
//	query, args := stmt.Sql()
//	row, err := database.QueryRow(ctx, query, args...)
//	if err != nil {
//		return id, fmt.Errorf("failed to query spells id by name: %w", err)
//	}
//	err = pgx.Row.Scan(row, &id)
//	if err != nil {
//		return id, fmt.Errorf("failed to collect spells id by name: %w", err)
//	}
//	return id, nil
//}
//
//func getSpellQueryLevel(ctx context.Context, id int, paramLevel int) (int, error) {
//	var queryLevel int
//	var minLevel int
//	var sType string
//	var err error
//	minLevel, err = getMinimumSpellLevelByID(ctx, id)
//	if err != nil {
//		return 0, err
//	}
//	sType, err = getSpellTypeByID(ctx, id)
//	if err != nil {
//		return 0, err
//	}
//
//	switch {
//	case sType == "slot":
//		switch {
//		case paramLevel < minLevel:
//			queryLevel = minLevel
//		default:
//			queryLevel = paramLevel
//		}
//	case sType == "character":
//		castLevels := make([]int, 0, 9)
//		castLevels, err = getSpellCastLevelsByID(ctx, paramLevel)
//		if err != nil {
//			return 0, err
//		}
//		queryLevel = 1
//		for _, v := range castLevels {
//			if v == paramLevel {
//				queryLevel = paramLevel
//				break
//			}
//		}
//	}
//	return queryLevel, nil
//}
//
//func GetSpellFormulaByLevel(ctx context.Context, spellID int, formulaLevel int) (*core.CastFormula, error) {
//	minLevel, err := getMinimumSpellLevelByID(ctx, spellID)
//	if err != nil {
//		return nil, err
//	}
//	if formulaLevel < minLevel {
//		return nil, fmt.Errorf("spell formula level must be greater than or equal to minimum spell level")
//	}
//
//	maxLevel, err := getMaxFormulaLevelBySpellID(ctx, spellID, formulaLevel)
//	if err != nil {
//		return nil, err
//	}
//
//	var formula core.CastFormula
//
//	stmt := SELECT(
//		SpellFormulas.FormulaLevel,
//		SpellFormulas.LevelType,
//		SpellDamage.NumberOfDice,
//		SpellDamage.Die,
//		SpellDamage.AmountToAdd,
//		SpellDamage.UseSpellmod,
//		SpellDamage.DamageType,
//	).FROM(
//		Spells.INNER_JOIN(SpellFormulas, Spells.ID.EQ(SpellFormulas.SpellID)).
//			INNER_JOIN(SpellDamage, SpellFormulas.FormulaID.EQ(SpellDamage.SpellFormulaID)),
//	).WHERE(
//		Spells.ID.EQ(Int(int64(spellID))).
//			AND(SpellFormulas.FormulaLevel.EQ(Int(int64(maxLevel)))),
//	)
//
//	query, args := stmt.Sql()
//	row, err := database.QueryRow(ctx, query, args...)
//	if err != nil {
//		return nil, fmt.Errorf("failed to query spell formula by level: %w", err)
//	}
//	err = row.Scan(&formula.CastLevel, &formula.NumberOfDice, &formula.Die, &formula.AmountToAdd, &formula.UseSpellmod, &formula.DamageType)
//	if err != nil {
//		return nil, fmt.Errorf("failed to collect spell formula by level: %w", err)
//	}
//	return &formula, nil
//}
//
//
//
//func GetHealingAndDamageSpellsByClassID(ctx context.Context, classID uint8) ([]Spell, []Spell, error) {
//	var healingSpells []Spell
//	var damageSpells []Spell
//	var err error
//	var ids []int
//	ids, err = GetUsableSpellIDsByClassID(ctx, classID)
//	if err != nil {
//		return healingSpells, damageSpells, err
//	}
//	for _, id := range ids {
//		spell, err2 := QuerySpellData(ctx, SpellQueryParams{ID: id, Level: 0})
//		if err2 != nil {
//			return healingSpells, damageSpells, err2
//		}
//		if spell.SpellType == "healing" {
//			healingSpells = append(healingSpells, spell)
//		}
//		if spell.SpellType == "damage" {
//			damageSpells = append(damageSpells, spell)
//		}
//	}
//	return healingSpells, damageSpells, nil
//}

// TODO: Guardian of faith is unique that it has no roll. just flat damage
// Need to account for this in spell attacks
