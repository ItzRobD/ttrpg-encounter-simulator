package repo

import (
	"context"
	"database/sql"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/internal/util"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"errors"
	"fmt"
	"strconv"
	"strings"

	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	. "github.com/go-jet/jet/v2/postgres"
)

func hydrateMonsterSpecialAbilityDataSRD(ctx context.Context, monsterID int) ([]core.Feature, error) {
	if monsterID <= 0 || monsterID > HighestSRDMonsterID {
		return nil, fmt.Errorf("invalid monster ID %d", monsterID)
	}

	featuresMap := make(map[int]*core.Feature)
	var featureIDs []int

	// Unified query using LEFT JOIN to fetch base feature and its associated data block
	stmt := SELECT(
		FeaturesMonsters.FeatureID,
		FeaturesMonsters.FeatureDataID,
		Features.Name,
		Features.Description,
		FeaturesData.Value,
		FeaturesData.NumberOfDice,
		FeaturesData.Die,
		FeaturesData.Modifier,
		FeaturesData.DamageType,
		FeaturesData.Scaler,
		FeaturesData.ScalerType,
		FeaturesData.Dc,
		FeaturesData.Ability,
		FeaturesData.DcOnsuccess,
	).FROM(
		FeaturesMonsters.
			INNER_JOIN(Features, Features.ID.EQ(FeaturesMonsters.FeatureID)).
			LEFT_JOIN(FeaturesData, FeaturesMonsters.FeatureDataID.EQ(FeaturesData.ID)),
	).WHERE(
		FeaturesMonsters.MonsterID.EQ(Int(int64(monsterID))),
	)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query special abilities for monster %d: %w", monsterID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var saID, saDataID sql.NullInt64
		var saName, saDescription sql.NullString
		var saValue, saNumberOfDice, saDie, saModifier, saScaler, saDC sql.NullInt64
		var saScalerType, saDmgType, saAbility, saOnSuccess sql.NullString

		err = rows.Scan(
			&saID, &saDataID, &saName, &saDescription,
			&saValue, &saNumberOfDice, &saDie, &saModifier, &saDmgType,
			&saScaler, &saScalerType, &saDC, &saAbility, &saOnSuccess,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan special ability data: %w", err)
		}

		if !saID.Valid {
			continue
		}

		var value, numDice, modifier, scaler, dc int
		var die core.DiceType
		var scalerType core.ScalerType
		var ability, onSuccess string
		dt := make([]core.DamageType, 0)

		if saValue.Valid {
			value = int(saValue.Int64)
		}
		if saNumberOfDice.Valid {
			numDice = int(saNumberOfDice.Int64)
		}
		if saDie.Valid {
			die = core.DiceType(saDie.Int64)
		}
		if saModifier.Valid {
			modifier = int(saModifier.Int64)
		}
		if saDmgType.Valid {
			s := saDmgType.String
			split := strings.Split(s, ",")
			for _, part := range split {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					dt = append(dt, core.DamageType(trimmed))
				}
			}
		}
		if saScaler.Valid {
			scaler = int(saScaler.Int64)
		}
		if saScalerType.Valid {
			scalerType = core.ScalerType(saScalerType.String)
		}
		if saDC.Valid {
			dc = int(saDC.Int64)
		}
		if saAbility.Valid {
			ability = saAbility.String
		}
		if saOnSuccess.Valid {
			onSuccess = saOnSuccess.String
		}

		data := core.FeatureData{
			Value:        value,
			NumberOfDice: numDice,
			Die:          die,
			Modifier:     modifier,
			DamageType:   dt,
			Scaler:       scaler,
			ScalerType:   scalerType,
			DC:           dc,
			Ability:      core.Ability(ability),
			DCOnSuccess:  core.DCOnSuccess(onSuccess),
		}

		f := core.NewFeatureFromSpecialAbility(strconv.Itoa(int(saID.Int64)), core.SpecialAbility(saName.String), saDescription.String, data)
		featuresMap[int(saID.Int64)] = &f
		featureIDs = append(featureIDs, int(saID.Int64))
	}
	rows.Close()

	if len(featureIDs) == 0 {
		return []core.Feature{}, nil
	}

	// Fetch all hooks for all identified features in a single query to avoid N+1
	hooksStmt := SELECT(
		FeaturesHooks.FeatureID,
		FeaturesHooks.HookType,
	).FROM(FeaturesHooks).WHERE(FeaturesHooks.FeatureID.IN(util.IntsToExpressions(featureIDs)...))

	hooksQuery, hooksArgs := hooksStmt.Sql()
	hooksRows, err := database.Query(ctx, hooksQuery, hooksArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query special ability hooks: %w", err)
	}
	defer hooksRows.Close()

	for hooksRows.Next() {
		var fID int
		var hookType string
		err = hooksRows.Scan(&fID, &hookType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan special ability hooks: %w", err)
		}
		if f, ok := featuresMap[fID]; ok {
			f.Hooks[core.MakeHookType(hookType)] = true
		}
	}

	// Assemble the final slice maintaining order if needed, or just from the map
	result := make([]core.Feature, 0, len(featuresMap))
	for _, id := range featureIDs {
		if f, ok := featuresMap[id]; ok {
			err = f.ValidateData()
			if err != nil {
				return nil, fmt.Errorf("failed to validate special ability data for %s: %w", f.Name, err)
			}
			result = append(result, *f)
			delete(featuresMap, id) // avoid duplicates if featureIDs has them (though unlikely with INNER JOIN)
		}
	}

	return result, nil
}

func HydrateClassFeaturesSRD(ctx context.Context, classID int, level int) ([]core.Feature, error) {
	if classID <= 0 || classID > HighestSRDClassID {
		return nil, fmt.Errorf("invalid class ID %d", classID)
	}

	featuresMap := make(map[int]*core.Feature)
	var featureIDs []int

	stmt := SELECT(
		FeaturesClasses.FeatureID,
		FeaturesClasses.FeatureDataID,
		Features.Name,
		Features.Description,
		FeaturesData.Value,
		FeaturesData.NumberOfDice,
		FeaturesData.Die,
		FeaturesData.Modifier,
		FeaturesData.DamageType,
		FeaturesData.Scaler,
		FeaturesData.ScalerType,
		FeaturesData.Dc,
		FeaturesData.Ability,
		FeaturesData.DcOnsuccess,
		FeaturesData.Effect,
		FeaturesData.BonusTargetTypes,
		FeaturesData.RerollType,
		FeaturesData.RerollThreshold,
	).FROM(
		FeaturesClasses.
			INNER_JOIN(Features, Features.ID.EQ(FeaturesClasses.FeatureID)).
			LEFT_JOIN(FeaturesData, FeaturesClasses.FeatureDataID.EQ(FeaturesData.ID)),
	).WHERE(
		FeaturesClasses.ClassID.EQ(Int(int64(classID))).
			AND(FeaturesClasses.Level.LT_EQ(Int(int64(level)))),
	)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query class features for class %d: %w", classID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var saID, saDataID sql.NullInt64
		var saName, saDescription, saEffect sql.NullString
		var saValue, saNumberOfDice, saDie, saModifier, saScaler, saDC sql.NullInt64
		var saScalerType, saDmgType, saAbility, saOnSuccess sql.NullString
		var saBonusTargetTypes sql.NullString
		var saRerollType sql.NullString
		var saRerollThreshold sql.NullInt64
		err = rows.Scan(
			&saID, &saDataID, &saName, &saDescription,
			&saValue, &saNumberOfDice, &saDie, &saModifier, &saDmgType,
			&saScaler, &saScalerType, &saDC, &saAbility, &saOnSuccess,
			&saEffect, &saBonusTargetTypes, &saRerollType, &saRerollThreshold,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan class feature data: %w", err)
		}
		if !saID.Valid {
			continue
		}
		var value, numDice, modifier, scaler, dc int
		var die core.DiceType
		var scalerType core.ScalerType
		var ability, onSuccess string
		dt := make([]core.DamageType, 0)
		if saValue.Valid {
			value = int(saValue.Int64)
		}
		if saNumberOfDice.Valid {
			numDice = int(saNumberOfDice.Int64)
		}
		if saDie.Valid {
			die = core.DiceType(saDie.Int64)
		}
		if saModifier.Valid {
			modifier = int(saModifier.Int64)
		}
		if saDmgType.Valid {
			s := saDmgType.String
			split := strings.Split(s, ",")
			for _, part := range split {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					dt = append(dt, core.DamageType(trimmed))
				}
			}
		}
		if saScaler.Valid {
			scaler = int(saScaler.Int64)
		}
		if saScalerType.Valid {
			scalerType = core.ScalerType(saScalerType.String)
		}
		if saDC.Valid {
			dc = int(saDC.Int64)
		}
		if saAbility.Valid {
			ability = saAbility.String
		}
		if saOnSuccess.Valid {
			onSuccess = saOnSuccess.String
		}
		bonusTargets := make([]core.MonsterType, 0)
		if saBonusTargetTypes.Valid {
			split := strings.Split(saBonusTargetTypes.String, ",")
			for _, part := range split {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					bonusTargets = append(bonusTargets, core.MonsterType(trimmed))
				}
			}
		}
		data := core.FeatureData{
			Value:            value,
			NumberOfDice:     numDice,
			Die:              die,
			Modifier:         modifier,
			DamageType:       dt,
			Scaler:           scaler,
			ScalerType:       scalerType,
			DC:               dc,
			Ability:          core.Ability(ability),
			DCOnSuccess:      core.DCOnSuccess(onSuccess),
			BonusTargetTypes: bonusTargets,
			RerollType:       saRerollType.String,
			RerollThreshold:  int(saRerollThreshold.Int64),
		}
		// Dynamic Scaling
		if data.ScalerType == core.ScalerLevel {
			data.Value = data.Value * level
		}
		f := core.NewFeatureFromSpecialAbility(strconv.Itoa(int(saID.Int64)), core.SpecialAbility(saName.String), saDescription.String, data)
		featuresMap[int(saID.Int64)] = &f
		featureIDs = append(featureIDs, int(saID.Int64))
	}
	rows.Close()
	if len(featureIDs) == 0 {
		return []core.Feature{}, nil
	}
	hooksStmt := SELECT(
		FeaturesHooks.FeatureID,
		FeaturesHooks.HookType,
	).FROM(FeaturesHooks).WHERE(FeaturesHooks.FeatureID.IN(util.IntsToExpressions(featureIDs)...))
	hooksQuery, hooksArgs := hooksStmt.Sql()
	hooksRows, err := database.Query(ctx, hooksQuery, hooksArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query class feature hooks: %w", err)
	}
	defer hooksRows.Close()
	for hooksRows.Next() {
		var fID int
		var hookType string
		err = hooksRows.Scan(&fID, &hookType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan class feature hooks: %w", err)
		}
		if f, ok := featuresMap[fID]; ok {
			f.Hooks[core.MakeHookType(hookType)] = true
		}
	}
	result := make([]core.Feature, 0, len(featuresMap))
	for _, id := range featureIDs {
		if f, ok := featuresMap[id]; ok {
			err = f.ValidateData()
			if err != nil {
				return nil, fmt.Errorf("failed to validate class feature data for %s: %w", f.Name, err)
			}
			result = append(result, *f)
			delete(featuresMap, id)
		}
	}
	return result, nil
}

func GetSpellSlotsByClassAndLevel(ctx context.Context, classID int, level int) (spells.SpellSlots, error) {
	stmt := SELECT(
		ClassesSpellSlots.Level1, ClassesSpellSlots.Level2, ClassesSpellSlots.Level3,
		ClassesSpellSlots.Level4, ClassesSpellSlots.Level5, ClassesSpellSlots.Level6,
		ClassesSpellSlots.Level7, ClassesSpellSlots.Level8, ClassesSpellSlots.Level9,
	).FROM(
		ClassesSpellSlots,
	).WHERE(
		ClassesSpellSlots.ClassID.EQ(Int(int64(classID))).AND(ClassesSpellSlots.Level.EQ(Int(int64(level)))),
	)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query spell slots by class id and level: %w", err)
	}

	levels := make([]int, 9)
	err = row.Scan(&levels[0], &levels[1], &levels[2], &levels[3], &levels[4], &levels[5], &levels[6], &levels[7], &levels[8])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make(spells.SpellSlots), nil
		}
		return nil, fmt.Errorf("failed to scan spell slots by class id and level: %w", err)
	}

	slots := make(spells.SpellSlots)
	for i, value := range levels {
		if value > 0 {
			slots[i+1] = value
		}
	}

	return slots, nil
}

func GetSpellcastingAbilityByClassID(ctx context.Context, classID int) (core.Ability, error) {
	stmt := SELECT(Classes.Spellmod).
		FROM(Classes).
		WHERE(Classes.ID.EQ(Int(int64(classID))))
	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return core.AbilityNone, fmt.Errorf("failed to query spellmod by class id: %w", err)
	}
	var spellmod sql.NullString
	err = row.Scan(&spellmod)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.AbilityNone, nil
		}
		return core.AbilityNone, fmt.Errorf("failed to scan spellmod by class id: %w", err)
	}

	if spellmod.Valid {
		return core.MakeAbility(spellmod.String), nil
	}
	return core.AbilityNone, nil
}

func GetAttackCountByClassAndLevel(ctx context.Context, classID int, level int) (int, error) {
	numberOfAttacks := 1

	stmt := SELECT(ClassesExtraAttack.NumberOfAttacks).
		FROM(ClassesExtraAttack).
		WHERE(ClassesExtraAttack.ClassID.EQ(Int(int64(classID))).AND(ClassesExtraAttack.ClassLevel.LT_EQ(Int(int64(level))))).
		ORDER_BY(ClassesExtraAttack.ClassLevel.DESC()).
		LIMIT(1)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return numberOfAttacks, fmt.Errorf("failed to query extra attacks by level and class: %w", err)
	}

	var numberOfAttacksNullable sql.NullInt64
	err = row.Scan(&numberOfAttacksNullable)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return numberOfAttacks, nil
		}
		return numberOfAttacks, fmt.Errorf("failed to scan extra attacks by level and class: %w", err)
	}

	if numberOfAttacksNullable.Valid {
		numberOfAttacks = int(numberOfAttacksNullable.Int64)
	}

	return numberOfAttacks, nil
}

func HydrateRaceFeaturesSRD(ctx context.Context, raceID int, level int, dragonbornColor string) ([]core.Feature, error) {
	if raceID <= 0 || raceID > HighestSRDRaceID {
		return nil, fmt.Errorf("invalid race ID %d", raceID)
	}

	featuresMap := make(map[int]*core.Feature)
	var featureIDs []int

	// Generic race features
	stmt := SELECT(
		FeaturesRaces.FeatureID,
		FeaturesRaces.FeatureDataID,
		Features.Name,
		Features.Description,
		FeaturesData.Value,
		FeaturesData.NumberOfDice,
		FeaturesData.Die,
		FeaturesData.Modifier,
		FeaturesData.DamageType,
		FeaturesData.Scaler,
		FeaturesData.ScalerType,
		FeaturesData.Dc,
		FeaturesData.Ability,
		FeaturesData.DcOnsuccess,
		FeaturesData.Effect,
		FeaturesData.BonusTargetTypes,
		FeaturesData.RerollType,
		FeaturesData.RerollThreshold,
	).FROM(
		FeaturesRaces.
			INNER_JOIN(Features, Features.ID.EQ(FeaturesRaces.FeatureID)).
			LEFT_JOIN(FeaturesData, FeaturesRaces.FeatureDataID.EQ(FeaturesData.ID)),
	).WHERE(
		FeaturesRaces.RaceID.EQ(Int(int64(raceID))),
	)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query race features for race %d: %w", raceID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var saID, saDataID sql.NullInt64
		var saName, saDescription, saEffect sql.NullString
		var saValue, saNumberOfDice, saDie, saModifier, saScaler, saDC sql.NullInt64
		var saScalerType, saDmgType, saAbility, saOnSuccess sql.NullString
		var saBonusTargetTypes sql.NullString
		var saRerollType sql.NullString
		var saRerollThreshold sql.NullInt64
		err = rows.Scan(
			&saID, &saDataID, &saName, &saDescription,
			&saValue, &saNumberOfDice, &saDie, &saModifier, &saDmgType,
			&saScaler, &saScalerType, &saDC, &saAbility, &saOnSuccess,
			&saEffect, &saBonusTargetTypes, &saRerollType, &saRerollThreshold,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan race feature data: %w", err)
		}
		if !saID.Valid {
			continue
		}

		var value, numDice, modifier, scaler, dc int
		var die core.DiceType
		var scalerType core.ScalerType
		var ability, onSuccess string
		dt := make([]core.DamageType, 0)

		if saValue.Valid {
			value = int(saValue.Int64)
		}
		if saNumberOfDice.Valid {
			numDice = int(saNumberOfDice.Int64)
		}
		if saDie.Valid {
			die = core.DiceType(saDie.Int64)
		}
		if saModifier.Valid {
			modifier = int(saModifier.Int64)
		}
		if saDmgType.Valid {
			s := saDmgType.String
			split := strings.Split(s, ",")
			for _, part := range split {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					dt = append(dt, core.DamageType(trimmed))
				}
			}
		}
		if saScaler.Valid {
			scaler = int(saScaler.Int64)
		}
		if saScalerType.Valid {
			scalerType = core.ScalerType(saScalerType.String)
		}
		if saDC.Valid {
			dc = int(saDC.Int64)
		}
		if saAbility.Valid {
			ability = saAbility.String
		}
		if saOnSuccess.Valid {
			onSuccess = saOnSuccess.String
		}
		bonusTargets := make([]core.MonsterType, 0)
		if saBonusTargetTypes.Valid {
			split := strings.Split(saBonusTargetTypes.String, ",")
			for _, part := range split {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					bonusTargets = append(bonusTargets, core.MonsterType(trimmed))
				}
			}
		}

		data := core.FeatureData{
			Value:            value,
			NumberOfDice:     numDice,
			Die:              die,
			Modifier:         modifier,
			DamageType:       dt,
			Scaler:           scaler,
			ScalerType:       scalerType,
			DC:               dc,
			Ability:          core.Ability(ability),
			DCOnSuccess:      core.DCOnSuccess(onSuccess),
			BonusTargetTypes: bonusTargets,
			RerollType:       saRerollType.String,
			RerollThreshold:  int(saRerollThreshold.Int64),
		}

		// Dragonborn Breath Weapon customization
		if core.RaceID(raceID) == core.Dragonborn && saName.String == "Breath Weapon" {
			// Fetch damage type and formula
			dbStmt := SELECT(
				DamageTypes.DamageType,
				RacesDragonbornBreathDamageFormula.NumberOfDice,
				RacesDragonbornBreathDamageFormula.Die,
			).FROM(
				RacesDragonbornAncestries.
					INNER_JOIN(DamageTypes, DamageTypes.ID.EQ(RacesDragonbornAncestries.DamageTypeID)).
					CROSS_JOIN(RacesDragonbornBreathDamageFormula),
			).WHERE(
				RacesDragonbornAncestries.Color.EQ(String(dragonbornColor)).
					AND(RacesDragonbornBreathDamageFormula.Level.LT_EQ(Int(int64(level)))),
			).ORDER_BY(
				RacesDragonbornBreathDamageFormula.Level.DESC(),
			).LIMIT(1)

			dbQuery, dbArgs := dbStmt.Sql()
			dbRow, dbErr := database.QueryRow(ctx, dbQuery, dbArgs...)
			if dbErr == nil {
				var dbDmgType string
				var dbNumDice, dbDie int
				err = dbRow.Scan(&dbDmgType, &dbNumDice, &dbDie)
				if err == nil {
					data.DamageType = []core.DamageType{core.DamageType(dbDmgType)}
					data.NumberOfDice = dbNumDice
					data.Die = core.DiceType(dbDie)
				}
			}
		}

		// Dragonborn Draconic Resistance customization
		if core.RaceID(raceID) == core.Dragonborn && saName.String == "Draconic Resistance" {
			dbStmt := SELECT(
				DamageTypes.DamageType,
			).FROM(
				RacesDragonbornAncestries.
					INNER_JOIN(DamageTypes, DamageTypes.ID.EQ(RacesDragonbornAncestries.DamageTypeID)),
			).WHERE(
				RacesDragonbornAncestries.Color.EQ(String(dragonbornColor)),
			).LIMIT(1)

			dbQuery, dbArgs := dbStmt.Sql()
			dbRow, dbErr := database.QueryRow(ctx, dbQuery, dbArgs...)
			if dbErr == nil {
				var dbDmgType string
				err = dbRow.Scan(&dbDmgType)
				if err == nil {
					data.DamageType = []core.DamageType{core.DamageType(dbDmgType)}
				}
			}
		}

		f := core.NewFeatureFromSpecialAbility(strconv.Itoa(int(saID.Int64)), core.SpecialAbility(saName.String), saDescription.String, data)
		featuresMap[int(saID.Int64)] = &f
		featureIDs = append(featureIDs, int(saID.Int64))
	}
	rows.Close()
	if len(featureIDs) == 0 {
		return []core.Feature{}, nil
	}
	hooksStmt := SELECT(
		FeaturesHooks.FeatureID,
		FeaturesHooks.HookType,
	).FROM(FeaturesHooks).WHERE(FeaturesHooks.FeatureID.IN(util.IntsToExpressions(featureIDs)...))
	hooksQuery, hooksArgs := hooksStmt.Sql()
	hooksRows, err := database.Query(ctx, hooksQuery, hooksArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query race feature hooks: %w", err)
	}
	defer hooksRows.Close()
	for hooksRows.Next() {
		var fID int
		var hookType string
		err = hooksRows.Scan(&fID, &hookType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan race feature hooks: %w", err)
		}
		if f, ok := featuresMap[fID]; ok {
			f.Hooks[core.MakeHookType(hookType)] = true
		}
	}
	result := make([]core.Feature, 0, len(featuresMap))
	for _, id := range featureIDs {
		if f, ok := featuresMap[id]; ok {
			err = f.ValidateData()
			if err != nil {
				return nil, fmt.Errorf("failed to validate race feature data for %s: %w", f.Name, err)
			}
			result = append(result, *f)
			delete(featuresMap, id)
		}
	}
	return result, nil
}
