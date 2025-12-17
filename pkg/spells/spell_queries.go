package spells

import (
	"context"
	"database/sql"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/internal/util"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"fmt"

	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/enum"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
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

		// DC AbilityUsed
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
