package monster

import (
	"context"
	"database/sql"
	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/enum"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/internal/util"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/core/monster_action_manager"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"encoding/json"
	"fmt"
	. "github.com/go-jet/jet/v2/postgres"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func QueryMonsterConfigData(ctx context.Context, params MonsterQueryParams) (map[int]MonsterConfig, error) {
	var config map[int]MonsterConfig
	var err error

	if len(params.ID) == 0 {
		ids, err := getMonsterIDsByName(ctx, params.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get monster ids by name: %w", err)
		}
		config, err = getMonsterConfigByID(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("failed to get monster config by id: %w", err)
		}
	} else {
		config, err = getMonsterConfigByID(ctx, params.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get monster config by id: %w", err)
		}
	}

	return config, nil
}
func getMonsterConfigByID(ctx context.Context, ids []int) (map[int]MonsterConfig, error) {
	config := make(map[int]MonsterConfig)

	bases, err := getMonsterBaseDataByID(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get monster base data by id: %w", err)
	}
	actions, err := getMonsterActionsByID(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get monster actions by id: %w", err)
	}
	multiattacks, err := getMonsterMultiattacksByID(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get monster multiattacks by id: %w", err)
	}
	legendaryActions, err := getMonsterLegendaryActionsByID(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get monster legendary actions by id: %w", err)
	}
	specialAbilities, err := getMonsterSpecialAbilitiesByID(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get monster special abilities: %w", err)
	}
	resistances, err := getMonsterResistancesByID(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get monster resistances by id: %w", err)
	}
	scConfig, err := GetMonsterSpellcastingConfigByID(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get monster spellcasting config by id: %w", err)
	}

	for _, i := range ids {
		monster := MonsterConfig{
			Base:               bases[i],
			Actions:            actions[i],
			Multiattacks:       multiattacks[i],
			LegendaryActions:   legendaryActions[i],
			SpecialAbilities:   specialAbilities[i],
			Resistances:        resistances[i],
			spellcastingConfig: scConfig[i],
		}
		config[i] = monster
	}

	return config, nil
}

func getMonsterIDsByName(ctx context.Context, names []string) ([]int, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("no names provided")
	}
	var ids []int

	titlized := make([]string, len(names))
	caser := cases.Title(language.English)
	for i, name := range names {
		titlized[i] = caser.String(name)
	}

	stmt := SELECT(Monsters.ID).
		FROM(Monsters).
		WHERE(Monsters.Name.IN(util.StringsToExpressions(titlized)...))

	query, args := stmt.Sql()
	row, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query monster id by name: %w", err)
	}
	defer row.Close()
	for row.Next() {
		var id int
		err = row.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monster id by name: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func getMonsterBaseDataByID(ctx context.Context, ids []int) (map[int]MonsterBase, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no ids provided")
	}
	bases := make(map[int]MonsterBase)
	var strSave, dexSave, conSave, intSave, wisSave, chaSave sql.NullInt32 // Used as placeholders for save profs
	stmt := SELECT(
		Monsters.ID,
		Monsters.Name,
		Monsters.Size,
		Monsters.Type,
		Monsters.ArmorClass,
		Monsters.ProficiencyBonus,
		Monsters.Cr,
		Monsters.APIURL,
		Monsters.IsLegendary,
		Monsters.IsSpellcaster,
		Monsters.IsInnateCaster,
		MonsterAbilityScoreBlock.Strength,
		MonsterAbilityScoreBlock.Dexterity,
		MonsterAbilityScoreBlock.Constitution,
		MonsterAbilityScoreBlock.Intelligence,
		MonsterAbilityScoreBlock.Wisdom,
		MonsterAbilityScoreBlock.Charisma,
		MonsterHitPointFormulas.HpAverage,
		MonsterHitPointFormulas.NumberOfDice,
		MonsterHitPointFormulas.Die,
		MonsterHitPointFormulas.AmountToAdd,
		MonsterSaveProficiencies.Strength,
		MonsterSaveProficiencies.Dexterity,
		MonsterSaveProficiencies.Constitution,
		MonsterSaveProficiencies.Intelligence,
		MonsterSaveProficiencies.Wisdom,
		MonsterSaveProficiencies.Charisma,
	).FROM(Monsters.
		LEFT_JOIN(MonsterAbilityScoreBlock, Monsters.ID.EQ(MonsterAbilityScoreBlock.MonsterID)).
		LEFT_JOIN(MonsterHitPointFormulas, Monsters.ID.EQ(MonsterHitPointFormulas.MonsterID)).
		LEFT_JOIN(MonsterSaveProficiencies, Monsters.ID.EQ(MonsterSaveProficiencies.MonsterID))).
		WHERE(Monsters.ID.IN(util.IntsToExpressions(ids)...))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query monster base data by id: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var base MonsterBase
		err = rows.Scan(
			&base.ID,
			&base.Name,
			&base.Size,
			&base.Type,
			&base.AC,
			&base.ProficiencyBonus,
			&base.CR,
			&base.ApiURL,
			&base.IsLegendary,
			&base.IsSpellcaster,
			&base.IsInnateSpellcaster,
			&base.AbilityScores.Strength,
			&base.AbilityScores.Dexterity,
			&base.AbilityScores.Constitution,
			&base.AbilityScores.Intelligence,
			&base.AbilityScores.Wisdom,
			&base.AbilityScores.Charisma,
			&base.HP.HPAverage,
			&base.HP.NumberOfDice,
			&base.HP.Die,
			&base.HP.AmountToAdd,
			&strSave, &dexSave, &conSave, &intSave, &wisSave, &chaSave,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monster base data by id: %w", err)
		}

		base.AbilityScoreProf = core.NewAbilityScoresProficiencies(
			strSave.Valid && strSave.Int32 != 0,
			dexSave.Valid && dexSave.Int32 != 0,
			conSave.Valid && conSave.Int32 != 0,
			intSave.Valid && intSave.Int32 != 0,
			wisSave.Valid && wisSave.Int32 != 0,
			chaSave.Valid && chaSave.Int32 != 0)

		bases[base.ID] = base
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return bases, nil
}

func getMonsterActionsByID(ctx context.Context, ids []int) (map[int]map[int]monster_action_manager.Action, error) {
	mActionsMap := make(map[int]map[int]monster_action_manager.Action)
	stmt := SELECT(
		MonsterActions.MonsterID,
		MonsterActions.ActionID,
		MonsterActions.Name,
		MonsterActions.RechargeValue,
		MonsterActions.HasDc,
		MonsterActions.Index,
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(false))).
			THEN(MonsterAttackBonusBlocks.NumberOfDice).
			ELSE(MonsterDcDamageBlocks.NumberOfDice),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(false))).
			THEN(MonsterAttackBonusBlocks.Die).
			ELSE(MonsterDcDamageBlocks.Die),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(false))).
			THEN(MonsterAttackBonusBlocks.AmountToAdd).
			ELSE(MonsterDcDamageBlocks.AmountToAdd),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(false))).
			THEN(MonsterAttackBonusBlocks.AttackBonus).
			ELSE(Int(0)),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(false))).
			THEN(MonsterAttackBonusBlocks.DmgType).
			ELSE(MonsterDcDamageBlocks.DmgType),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(true))).
			THEN(MonsterDcDamageBlocks.Ability).
			ELSE(enum.Abilityscore.None),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(true))).
			THEN(MonsterDcDamageBlocks.OnSuccess).
			ELSE(enum.Dcsuccess.None),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(true))).
			THEN(MonsterDcDamageBlocks.DcValue).
			ELSE(Int(0)),
	).
		FROM(
			MonsterActions.
				LEFT_JOIN(MonsterAttackBonusBlocks, MonsterActions.ActionID.EQ(MonsterAttackBonusBlocks.ActionID)).
				LEFT_JOIN(MonsterDcDamageBlocks, MonsterActions.ActionID.EQ(MonsterDcDamageBlocks.ActionID)),
		).WHERE(
		MonsterActions.MonsterID.IN(util.IntsToExpressions(ids)...),
	).ORDER_BY(
		MonsterActions.MonsterID.ASC(), MonsterActions.ActionID.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return mActionsMap, fmt.Errorf("failed to query monster actions by id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var action monster_action_manager.Action
		var monsterID int
		err = rows.Scan(
			&monsterID,
			&action.ActionID,
			&action.Name,
			&action.RechargeValue,
			&action.HasDC,
			&action.Index,
			&action.NumberOfDice,
			&action.Die,
			&action.AmountToAdd,
			&action.AttackBonus,
			&action.DamageType,
			&action.DCAbility,
			&action.DCOnSuccess,
			&action.DC)
		if err != nil {
			return mActionsMap, fmt.Errorf("failed to scan monster actions by monsterID (%d): %w", monsterID, err)
		}

		monster, exists := mActionsMap[monsterID]
		if exists {
			monster[action.ActionID] = action
		} else {
			monster = make(map[int]monster_action_manager.Action)
			monster[action.ActionID] = action
			mActionsMap[monsterID] = monster
		}
	}

	if errI := rows.Err(); errI != nil {
		return mActionsMap, fmt.Errorf("failed to query monster actions by id: %w", errI)
	}

	return mActionsMap, nil
}

func getMonsterMultiattacksByID(ctx context.Context, id []int) (map[int]map[int][]monster_action_manager.Multiattack, error) {
	mMAMap := make(map[int]map[int][]monster_action_manager.Multiattack)

	stmt := SELECT(
		MonsterMultiattacks.MonsterID,
		MonsterMultiattacks.ActionID,
		MonsterMultiattacks.AttackCount,
		MonsterMultiattacks.IsOption,
		MonsterMultiattacks.OptionIndex).
		FROM(MonsterMultiattacks).
		WHERE(MonsterMultiattacks.MonsterID.IN(util.IntsToExpressions(id)...)).
		ORDER_BY(MonsterMultiattacks.MonsterID.ASC(), MonsterMultiattacks.OptionIndex.ASC(), MonsterMultiattacks.ActionID.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying monster multiattacks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var monsterID int
		var aid, count, index int
		var isOption bool
		err = rows.Scan(&monsterID, &aid, &count, &isOption, &index)
		if err != nil {
			return nil, fmt.Errorf("error scanning monster multiattacks: %w", err)
		}

		if mMAMap[monsterID] == nil {
			mMAMap[monsterID] = make(map[int][]monster_action_manager.Multiattack)
		}

		multiattack := monster_action_manager.Multiattack{
			ActionID: aid,
			Count:    count,
		}

		mMAMap[monsterID][index] = append(mMAMap[monsterID][index], multiattack)

	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating monster multiattacks: %w", err)
	}

	return mMAMap, nil
}

func getMonsterLegendaryActionsByID(ctx context.Context, id []int) (map[int][]monster_action_manager.LegendaryAction, error) {
	mLAMap := make(map[int][]monster_action_manager.LegendaryAction)
	stmt := SELECT(
		MonsterActionsLegendary.MonsterID,
		MonsterActionsLegendary.ActionCost,
		MonsterActions.ActionID,
		MonsterActions.Name,
		MonsterActions.RechargeValue,
		MonsterActions.HasDc,
		MonsterActions.Index,
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(false))).
			THEN(MonsterAttackBonusBlocks.NumberOfDice).
			ELSE(MonsterDcDamageBlocks.NumberOfDice),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(false))).
			THEN(MonsterAttackBonusBlocks.Die).
			ELSE(MonsterDcDamageBlocks.Die),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(false))).
			THEN(MonsterAttackBonusBlocks.AmountToAdd).
			ELSE(MonsterDcDamageBlocks.AmountToAdd),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(false))).
			THEN(MonsterAttackBonusBlocks.AttackBonus).
			ELSE(Int(0)),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(false))).
			THEN(MonsterAttackBonusBlocks.DmgType).
			ELSE(MonsterDcDamageBlocks.DmgType),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(true))).
			THEN(MonsterDcDamageBlocks.Ability).
			ELSE(enum.Abilityscore.None),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(true))).
			THEN(MonsterDcDamageBlocks.OnSuccess).
			ELSE(enum.Dcsuccess.None),
		CASE().
			WHEN(MonsterActions.HasDc.EQ(Bool(true))).
			THEN(MonsterDcDamageBlocks.DcValue).
			ELSE(Int(0)),
	).FROM(
		MonsterActionsLegendary.
			LEFT_JOIN(MonsterActions, MonsterActionsLegendary.ActionID.EQ(MonsterActions.ActionID)).
			LEFT_JOIN(MonsterAttackBonusBlocks, MonsterActions.ActionID.EQ(MonsterAttackBonusBlocks.ActionID)).
			LEFT_JOIN(MonsterDcDamageBlocks, MonsterActions.ActionID.EQ(MonsterDcDamageBlocks.ActionID)),
	).WHERE(
		MonsterActionsLegendary.MonsterID.IN(util.IntsToExpressions(id)...),
	).ORDER_BY(
		MonsterActionsLegendary.MonsterID.ASC(),
		MonsterActionsLegendary.ActionID.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query monster legendary actions by id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var monsterID int
		var la monster_action_manager.LegendaryAction
		err = rows.Scan(
			&monsterID,
			&la.Cost,
			&la.Action.ActionID,
			&la.Action.Name,
			&la.Action.RechargeValue,
			&la.Action.HasDC,
			&la.Action.Index,
			&la.Action.NumberOfDice,
			&la.Action.Die,
			&la.Action.AmountToAdd,
			&la.Action.AttackBonus,
			&la.Action.DamageType,
			&la.Action.DCAbility,
			&la.Action.DCOnSuccess,
			&la.Action.DC)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monster legendary actions by monsterID: %w", err)
		}

		mLAMap[monsterID] = append(mLAMap[monsterID], la)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query monster legendary actions by id: %w", err)
	}

	return mLAMap, nil
}

func getMonsterSpecialAbilitiesByID(ctx context.Context, id []int) (map[int][]monster_action_manager.SpecialAbility, error) {
	mSAMap := make(map[int][]monster_action_manager.SpecialAbility)
	stmt := SELECT(
		MonsterSpecialAbilities.MonsterID,
		MonsterSpecialAbilities.Name,
		MonsterSpecialAbilities.UsageCount,
		MonsterSpecialAbilities.Description,
	).FROM(
		MonsterSpecialAbilities,
	).WHERE(
		MonsterSpecialAbilities.MonsterID.IN(util.IntsToExpressions(id)...),
	).ORDER_BY(
		MonsterSpecialAbilities.MonsterID.ASC(), MonsterSpecialAbilities.Name.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query monster special abilities by id: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sa monster_action_manager.SpecialAbility
		var monsterID int
		var usageCount sql.NullInt64
		err = rows.Scan(&monsterID, &sa.Name, &usageCount, &sa.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monster special abilities by id: %w", err)
		}
		if usageCount.Valid {
			sa.UsageCount = int(usageCount.Int64)
		} else {
			sa.UsageCount = 0
		}

		mSAMap[monsterID] = append(mSAMap[monsterID], sa)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query monster special abilities by id: %w", err)
	}

	return mSAMap, nil
}

func getMonsterResistancesByID(ctx context.Context, id []int) (map[int]core.DamageResistances, error) {
	mResistanceMap := make(map[int]core.DamageResistances)
	stmt := SELECT(
		MonsterResistances.MonsterID,
		MonsterResistances.ResistanceType,
		MonsterResistances.DamageType,
		ResistBreakers.ResistBreakerType,
	).FROM(MonsterResistances.
		LEFT_JOIN(MonsterResistBreakers, MonsterResistances.ResistanceID.EQ(MonsterResistBreakers.ResistanceID)).
		LEFT_JOIN(ResistBreakers, MonsterResistBreakers.ResistBreakerID.EQ(ResistBreakers.ResistBreakerID)),
	).WHERE(MonsterResistances.MonsterID.IN(util.IntsToExpressions(id)...)).
		ORDER_BY(MonsterResistances.MonsterID.ASC(), MonsterResistances.DamageType.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query monster resistances by id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var monsterID int
		var rType, dmgType, bType sql.NullString

		err = rows.Scan(&monsterID, &rType, &dmgType, &bType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monster resistances by id: %w", err)
		}

		if !dmgType.Valid {
			return nil, fmt.Errorf("damage type cannot be null")
		}
		damageType, err := core.MakeDamageType(dmgType.String)
		if err != nil {
			return nil, fmt.Errorf("failed to make damage type: %w", err)
		}

		// Initialize monster's resistance map if needed (but keep it empty initially)
		if mResistanceMap[monsterID] == nil {
			mResistanceMap[monsterID] = make(core.DamageResistances)
		}

		// Get existing resistance or create new one
		resistance := mResistanceMap[monsterID][damageType]

		// Set resistance type (should be same across rows for same monster+damageType)
		if rType.Valid {
			resistanceType, err := core.MakeResistanceType(rType.String)
			if err != nil {
				return nil, fmt.Errorf("failed to make resistance type: %w", err)
			}
			resistance.Resistance = resistanceType
		} else {
			resistance.Resistance = core.ResistanceNone
		}

		// Add breaker if present
		if bType.Valid {
			breakerType, err := core.MakeResistBreaker(bType.String)
			if err != nil {
				return nil, fmt.Errorf("failed to make resist breaker: %w", err)
			}
			resistance.Breakers = append(resistance.Breakers, breakerType)
		}

		// Put the updated resistance back
		mResistanceMap[monsterID][damageType] = resistance

	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to query monster resistances by id: %w", err)
	}

	return mResistanceMap, nil
}

func GetMonsterSpellcastingConfigByID(ctx context.Context, id []int) (map[int]MonsterSpellcastingConfig, error) {
	configMap := make(map[int]MonsterSpellcastingConfig)

	stmt := SELECT(
		MonsterSpellcasting.MonsterID,
		MonsterSpellcasting.CastingLevel,
		MonsterSpellcasting.Ability,
		MonsterSpellcasting.AttackModifier,
		MonsterSpellcasting.SaveDc,
		Raw("COALESCE(ARRAY_AGG(DISTINCT monster_available_spells.spell_id) FILTER (WHERE monster_available_spells.spell_id IS NOT NULL), '{}')").AS("standard_spell_ids"),
		Raw("COALESCE(JSON_OBJECT_AGG(monster_available_spells_innate.spell_id, monster_available_spells_innate.times_per_day) FILTER (WHERE monster_available_spells_innate.spell_id IS NOT NULL), '{}')").AS("innate_spells"),
		Raw("COALESCE(JSON_OBJECT_AGG(monster_spellcasting_slots.spell_level, monster_spellcasting_slots.slots) FILTER (WHERE monster_spellcasting_slots.spell_level IS NOT NULL), '{}')").AS("spell_slots"),
	).FROM(MonsterSpellcasting.
		LEFT_JOIN(MonsterAvailableSpells, MonsterSpellcasting.MonsterID.EQ(MonsterAvailableSpells.MonsterID)).
		LEFT_JOIN(MonsterAvailableSpellsInnate, MonsterSpellcasting.MonsterID.EQ(MonsterAvailableSpellsInnate.MonsterID)).
		LEFT_JOIN(MonsterSpellcastingSlots, MonsterSpellcasting.MonsterID.EQ(MonsterSpellcastingSlots.MonsterID)),
	).WHERE(MonsterSpellcasting.MonsterID.IN(util.IntsToExpressions(id)...)).
		GROUP_BY(MonsterSpellcasting.MonsterID, MonsterSpellcasting.CastingLevel, MonsterSpellcasting.Ability, MonsterSpellcasting.AttackModifier, MonsterSpellcasting.SaveDc).
		ORDER_BY(MonsterSpellcasting.MonsterID.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query monster spellcasting by id: %w", err)
	}
	defer rows.Close()

	type tempConfig struct {
		MonsterID      int
		CastingLevel   int
		Ability        core.Ability
		AttackModifier int
		SaveDC         int
		StandardSpells []int
		InnateSpells   map[int]int // spell_id -> times_per_day
		SpellSlots     map[int]int
	}

	// Collect all spell IDs from both standard and innate spells
	allSpellIDs := make(map[int]struct{})
	tempConfigs := make(map[int]tempConfig)

	for rows.Next() {
		var monsterID, castingLevel int
		var am, dc sql.NullInt64
		var a string
		var standardSpellIDs []int
		var innateSpellsJSON, spellSlotsJSON string

		err = rows.Scan(&monsterID, &castingLevel, &a, &am, &dc, &standardSpellIDs, &innateSpellsJSON, &spellSlotsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monster spellcasting by id: %w", err)
		}

		// Parse innate spells JSON
		innateSpells := make(map[int]int)
		if innateSpellsJSON != "{}" {
			err = json.Unmarshal([]byte(innateSpellsJSON), &innateSpells)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal innate spells: %w", err)
			}
		}

		// Parse spell slots JSON
		spellSlots := make(map[int]int)
		if spellSlotsJSON != "{}" {
			err = json.Unmarshal([]byte(spellSlotsJSON), &spellSlots)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal spell slots: %w", err)
			}
		}

		// Collect all spell IDs
		for _, spellID := range standardSpellIDs {
			allSpellIDs[spellID] = struct{}{}
		}
		for spellID := range innateSpells {
			allSpellIDs[spellID] = struct{}{}
		}

		tempConfigs[monsterID] = tempConfig{
			MonsterID:    monsterID,
			CastingLevel: castingLevel,
			Ability:      core.MakeAbility(a),
			AttackModifier: func() int {
				if am.Valid {
					return int(am.Int64)
				}
				return 0
			}(),
			SaveDC: func() int {
				if dc.Valid {
					return int(dc.Int64)
				}
				return 0
			}(),
			StandardSpells: standardSpellIDs,
			InnateSpells:   innateSpells,
			SpellSlots:     spellSlots,
		}
	}

	if errR := rows.Err(); errR != nil {
		return nil, fmt.Errorf("failed to query monster spellcasting by id: %w", errR)
	}

	// Get all spells in one query
	spellIDSlice := make([]int, 0, len(allSpellIDs))
	for spellID := range allSpellIDs {
		spellIDSlice = append(spellIDSlice, spellID)
	}

	var allSpells map[int]spells.Spell
	if len(spellIDSlice) > 0 {
		params := spells.SpellQueryParams{ID: spellIDSlice}
		allSpells, err = spells.QuerySpellData(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("failed to query spells by id: %w", err)
		}
	} else {
		allSpells = make(map[int]spells.Spell)
	}

	// Build final configs
	for monsterID, temp := range tempConfigs {
		// Combine standard and innate spells
		var standardSpells []spells.Spell
		var innateSpells []spells.InnateSpell

		// Standard spells
		for _, spellID := range temp.StandardSpells {
			if spell, exists := allSpells[spellID]; exists {
				standardSpells = append(standardSpells, spell)
			}
		}

		// Innate spells with usage
		for spellID, timesPerDay := range temp.InnateSpells {
			if spell, exists := allSpells[spellID]; exists {
				innateSpells = append(innateSpells, spells.InnateSpell{
					Spell:          spell,
					MaxCastsPerDay: timesPerDay,
				})
			}
		}

		configMap[monsterID] = MonsterSpellcastingConfig{
			MonsterID:      temp.MonsterID,
			CastingLevel:   temp.CastingLevel,
			Ability:        temp.Ability,
			AttackModifier: temp.AttackModifier,
			SaveDC:         temp.SaveDC,
			LeveledSpells:  standardSpells, // Regular spells
			InnateSpells:   innateSpells,   // Innate spells with usage
			SpellSlots:     temp.SpellSlots,
		}
	}

	return configMap, nil
}

//func getMonsterActionManagerConfig(ctx context.Context, id int) (*monster_action_manager.MAMConfig, error) {
//	var err error
//	var config monster_action_manager.MAMConfig
//
//	config.Actions, err = getMonsterActionsByID(ctx, id)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get monster actions by id: %w", err)
//	}
//	config.Multiattacks, err = getMonsterMultiattacksByID(ctx, id)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get monster multiattacks by id: %w", err)
//	}
//	config.LegendaryActions, err = getMonsterLegendaryActionsByID(ctx, id)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get monster legendary actions by id: %w", err)
//	}
//	config.SpecialAbilities, err = getMonsterSpecialAbilitiesByID(ctx, id)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get monster special abilities by id: %w", err)
//	}
//
//	return &config, nil
//}
