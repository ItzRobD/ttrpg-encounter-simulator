package races

import (
	"context"
	"database/sql"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg_old/core"
	"errors"
	"fmt"

	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	. "github.com/go-jet/jet/v2/postgres"
)

// getRaceIDByName retrieves the id of a race from the database based on its name.
// It accepts a context and the race name as inputs, returning the race id or an error if the operation fails.
func getRaceIDByName(ctx context.Context, name string) (uint8, error) {
	var id uint8
	stmt := SELECT(
		Races.ID,
	).FROM(
		Races,
	).WHERE(
		Races.RaceName.EQ(String(name)))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return id, fmt.Errorf("failed to query race id by name: %w", err)
	}
	err = row.Scan(&id)
	if err != nil {
		return id, fmt.Errorf("failed to scan race id by name: %w", err)
	}

	return id, nil
}

func getRaceByID(ctx context.Context, id uint8) (Race, error) {
	var raceResult Race
	stmt := SELECT(
		Races.ID,
		Races.RaceName,
	).FROM(
		Races,
	).WHERE(
		Races.ID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return raceResult, fmt.Errorf("failed to query race by id: %w", err)
	}
	err = row.Scan(&raceResult.ID, &raceResult.Name)
	if err != nil {
		return raceResult, fmt.Errorf("failed to scan race by id: %w", err)
	}

	return raceResult, nil
}

func QueryRaceData(ctx context.Context, params RaceQueryParams) (Race, error) {
	raceResult := Race{
		ID:                 0,
		Name:               "",
		DragonbornFeatures: nil,
		Resistances:        core.NewDamageResistances(),
		SavingThrowAdv:     NewRacialSavingThrowAdvantage(),
	}
	var err error

	if params.ID != 0 {
		raceResult, err = getRaceByID(ctx, params.ID.Int())
	} else if params.Name != "" {
		var id uint8
		id, err = getRaceIDByName(ctx, params.Name)
		if err != nil {
			return raceResult, err
		}
		raceResult, err = getRaceByID(ctx, id)
		if err != nil {
			return raceResult, err
		}
	} else {
		return raceResult, fmt.Errorf("no race id or name provided")
	}

	resistances, err := getResistancesByRace(ctx, raceResult.ID, params.DragonbornColor)
	if err != nil {
		return raceResult, err
	}
	raceResult.Resistances = resistances

	if raceResult.ID == Dragonborn {
		if params.DragonbornColor == nil {
			return raceResult, fmt.Errorf("dragonborn race id but no dragonborn color provided")
		}
		raceResult.DragonbornFeatures, err = getDragonbornFeatures(ctx, params)
		if err != nil {
			return raceResult, err
		}
		raceResult.DragonbornFeatures.AncestryColor = *params.DragonbornColor
	}

	adv, err := getSavingThrowAdvantagesByRace(ctx, raceResult.ID)
	if err != nil {
		return raceResult, err
	}
	raceResult.SavingThrowAdv = adv

	return raceResult, nil
}

func getSavingThrowAdvantagesByRace(ctx context.Context, id RaceID) (RacialSavingThrowAdvantage, error) {
	// Initialize with default maps so setters won't panic on nil maps
	rst := NewRacialSavingThrowAdvantage()
	var err error

	stmt := SELECT(
		TraitSavingThrowAdvantages.Strength,
		TraitSavingThrowAdvantages.Dexterity,
		TraitSavingThrowAdvantages.Constitution,
		TraitSavingThrowAdvantages.Intelligence,
		TraitSavingThrowAdvantages.Wisdom,
		TraitSavingThrowAdvantages.Charisma,
		TraitSavingThrowAdvantages.OnlySpell,
		DamageTypes.DamageType,
	).
		FROM(TraitSavingThrowAdvantages.LEFT_JOIN(DamageTypes, TraitSavingThrowAdvantages.VsDamageTypeID.EQ(DamageTypes.ID))).
		WHERE(TraitSavingThrowAdvantages.RaceID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return rst, fmt.Errorf("failed to query saving throw advantages by race id: %w", err)
	}

	defer rows.Close()
	for rows.Next() {
		var str, dex, con, intel, wis, cha int
		var onlySpell bool
		var d sql.NullString
		err = rows.Scan(&str, &dex, &con, &intel, &wis, &cha, &onlySpell, &d)
		if err != nil {
			return rst, fmt.Errorf("failed to scan saving throw advantages by race id: %w", err)
		}

		rst.SetAdvantageAbility(core.AbilityStrength, core.AdvantageType(str))
		rst.SetAdvantageAbility(core.AbilityDexterity, core.AdvantageType(dex))
		rst.SetAdvantageAbility(core.AbilityConstitution, core.AdvantageType(con))
		rst.SetAdvantageAbility(core.AbilityIntelligence, core.AdvantageType(intel))
		rst.SetAdvantageAbility(core.AbilityWisdom, core.AdvantageType(wis))
		rst.SetAdvantageAbility(core.AbilityCharisma, core.AdvantageType(cha))

		if d.Valid {
			dt, dtErr := core.MakeDamageType(d.String)
			if dtErr != nil {
				return rst, dtErr
			}
			rst.SetAdvantageDamageType(dt, core.RollAdvantage)
		}

		rst.SetAdvantageOnlyAgainstSpells(onlySpell)
	}

	return rst, nil
}

func getResistancesByRace(ctx context.Context, id RaceID, color *DragonbornColor) (core.DamageResistances, error) {
	stmt := SELECT(DamageTypes.DamageType, TraitResistances.Resistance).
		FROM(TraitResistances.INNER_JOIN(DamageTypes, TraitResistances.DamageTypeID.EQ(DamageTypes.ID))).
		WHERE(TraitResistances.RaceID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return core.DamageResistances{}, fmt.Errorf("failed to query resistances by race id: %w", err)
	}
	defer rows.Close()

	resistances := core.NewDamageResistances()
	for rows.Next() {
		var d string
		var resistance int
		err = rows.Scan(&d, &resistance)
		if err != nil {
			return resistances, fmt.Errorf("failed to scan resistances by race id: %w", err)
		}

		dt, dErr := core.MakeDamageType(d)
		if dErr != nil {
			return resistances, dErr
		}

		switch resistance {
		case 2:
			resistances.SetResistance(dt, core.ResistanceImmune, nil)
		case 1:
			resistances.SetResistance(dt, core.ResistanceResistant, nil)
		case 0:
			break
		case -1:
			resistances.SetResistance(dt, core.ResistanceVulnerable, nil)
		default:
			return resistances, fmt.Errorf("invalid resistance value: %d", resistance)
		}
	}

	// Dragonborn resistance to own damage type
	if id == Dragonborn {
		if color == nil {
			return resistances, fmt.Errorf("race id = dragonborn and no dragonborn color provided")
		}
		dt, dbErr := getDragonbornDamageByColor(ctx, color)
		if dbErr != nil {
			return resistances, dbErr
		}
		resistances.SetResistance(dt, core.ResistanceResistant, nil)
	}

	return resistances, nil
}

func getDragonbornFeatures(ctx context.Context, params RaceQueryParams) (*DragonbornFeatures, error) {
	if params.DragonbornColor == nil {
		return nil, fmt.Errorf("no dragonborn color provided")
	}

	features := DragonbornFeatures{}
	numDice, dieType, err := getDragonbornBreathFormula(ctx, params.Level)
	if err != nil {
		return nil, err
	}

	dt, err := getDragonbornDamageByColor(ctx, params.DragonbornColor)
	if err != nil {
		return nil, err
	}

	features.DamageType = dt
	features.NumberOfDice = numDice
	features.Die = dieType

	return &features, nil
}

func getDragonbornDamageByColor(ctx context.Context, color *DragonbornColor) (core.DamageType, error) {
	if color == nil {
		return "", fmt.Errorf("no dragonborn color provided")
	}

	var dt string
	stmt := SELECT(DamageTypes.DamageType).
		FROM(RacesDragonbornAncestries.INNER_JOIN(DamageTypes, RacesDragonbornAncestries.DamageTypeID.EQ(DamageTypes.ID))).
		WHERE(RacesDragonbornAncestries.Color.EQ(String(string(*color)))).
		LIMIT(1)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return core.DamageNone, fmt.Errorf("failed to query dragonborn damage type by color: %w", err)
	}
	if err = row.Scan(&dt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.DamageNone, fmt.Errorf("dragonborn color %s not found", string(*color))
		}
		return core.DamageNone, fmt.Errorf("failed to scan dragonborn damage type by color: %w", err)
	}

	damageType, err := core.MakeDamageType(dt)
	if err != nil {
		return core.DamageNone, err
	}

	return damageType, nil
}

func getDragonbornBreathFormula(ctx context.Context, level uint8) (int, core.DiceType, error) {
	stmt := SELECT(RacesDragonbornBreathDamageFormula.NumberOfDice, RacesDragonbornBreathDamageFormula.Die).
		FROM(RacesDragonbornBreathDamageFormula).
		WHERE(RacesDragonbornBreathDamageFormula.Level.LT_EQ(Int(int64(level)))).
		ORDER_BY(RacesDragonbornBreathDamageFormula.Level.DESC()).
		LIMIT(1)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query dragonborn breath damage formula: %w", err)
	}
	var numberOfDice int
	var d int

	err = row.Scan(&numberOfDice, &d)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to scan dragonborn breath damage formula: %w", err)
	}

	if numberOfDice <= 0 && d <= 0 {
		return 0, 0, fmt.Errorf("no dragonborn breath damage formula found for level %d", level)
	}

	return numberOfDice, core.DiceType(d), nil
}
