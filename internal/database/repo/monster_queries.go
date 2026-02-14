package repo

import (
	"context"
	"database/sql"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type MonsterSummary struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	CR             float64 `json:"cr"`
	Type           string  `json:"type"`
	Size           string  `json:"size"`
	AC             int     `json:"ac"`
	IsLegendary    bool    `json:"is_legendary"`
	IsSpellcaster  bool    `json:"is_spellcaster"`
	IsInnateCaster bool    `json:"is_innate_caster"`
	IsCustom       bool    `json:"is_custom"`
}

func GetMonsterSummaries(ctx context.Context) (map[core.ID]MonsterSummary, error) {
	// Get SRD Summaries
	srdSummaries, err := GetMonsterSummariesSRD(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SRD monster summaries: %w", err)
	}

	// Get Custom Summaries

	return srdSummaries, nil
}

func GetMonsterSummariesSRD(ctx context.Context) (map[core.ID]MonsterSummary, error) {
	res := make(map[core.ID]MonsterSummary)

	stmt := SELECT(
		Monsters.ID,
		Monsters.Name,
		Monsters.Cr,
		Monsters.Type,
		Monsters.Size,
		Monsters.ArmorClass,
		Monsters.IsLegendary,
		Monsters.IsSpellcaster,
		Monsters.IsInnateCaster,
	).FROM(Monsters)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query monster summaries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m MonsterSummary
		var srdID int
		err = rows.Scan(
			&srdID,
			&m.Name,
			&m.CR,
			&m.Type,
			&m.Size,
			&m.AC,
			&m.IsLegendary,
			&m.IsSpellcaster,
			&m.IsInnateCaster,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monster summary: %w", err)
		}

		id := core.MakeID(srdID)
		m.ID = id.String()
		res[id] = m

	}

	return res, nil
}

func HydrateMonsterConfig(ctx context.Context, monsterID string) (*actor.ActorConfig, error) {
	var err error
	_, err = uuid.Parse(monsterID)
	if err == nil { // monsterID is a valid uuid -> hydrate from custom table
		return nil, nil // TODO: Custom Monster lookup
	} // else: monsterID is not a valid uuid -> hydrate from SRD

	id, err := strconv.Atoi(monsterID)
	if err != nil {
		return nil, fmt.Errorf("unable to parse monster id: %w", err)
	}
	if id <= 0 || id > HighestSRDMonsterID {
		return nil, fmt.Errorf("invalid monster id: %d", id)
	}

	var cfg actor.ActorConfig
	// Base
	err = hydrateBaseDataSRD(ctx, &cfg, id)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate base data for SRD monster: %w", err)
	}
	// Actions
	err = hydrateActionDataSRD(ctx, &cfg, id)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate action data for SRD monster: %w", err)
	}

	// Multiattacks
	err = hydrateMultiattackDataSRD(ctx, &cfg, id)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate multiattack data for SRD monster: %w", err)
	}
	// Special Abilities
	features, err := hydrateMonsterSpecialAbilityDataSRD(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate special ability data for SRD monster: %w", err)
	}
	cfg.Features = features
	// Legendary Actions
	err = hydrateLegendaryActionDataSRD(ctx, &cfg, id)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate legendary action data for SRD monster: %w", err)
	}
	// Resistances
	err = hydrateResistancesSRD(ctx, &cfg, id)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate resistances for SRD monster: %w", err)
	}
	// Spellcasting
	err = hydrateSpellcastingDataSRD(ctx, &cfg, id)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate spellcasting data for SRD monster: %w", err)
	}

	return &cfg, nil
}

func hydrateBaseDataSRD(ctx context.Context, cfg *actor.ActorConfig, id int) error {
	stmt := SELECT(
		Monsters.ID,
		Monsters.Name,
		Monsters.Size,
		Monsters.Type,
		Monsters.ArmorClass,
		Monsters.Cr,
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
		WHERE(Monsters.ID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query monster data: %w", err)
	}

	var hitDie int
	var monsterType string
	var monsterSize string
	var strSave, dexSave, conSave, intSave, wisSave, chaSave sql.NullInt32
	err = row.Scan(
		&cfg.ID,
		&cfg.Name,
		&monsterSize,
		&monsterType,
		&cfg.AC,
		&cfg.Metadata.CR,
		&cfg.Metadata.IsLegendary,
		&cfg.Metadata.SpellcasterMetadata.IsSpellcaster,
		&cfg.Metadata.SpellcasterMetadata.IsInnateCaster,
		&cfg.Abilities.AbilityScores.Strength,
		&cfg.Abilities.AbilityScores.Dexterity,
		&cfg.Abilities.AbilityScores.Constitution,
		&cfg.Abilities.AbilityScores.Intelligence,
		&cfg.Abilities.AbilityScores.Wisdom,
		&cfg.Abilities.AbilityScores.Charisma,
		&cfg.HPConfig.HPAverage,
		&cfg.HPConfig.NumberOfDice,
		&hitDie,
		&cfg.HPConfig.AmountToAdd,
		&strSave, &dexSave, &conSave, &intSave, &wisSave, &chaSave,
	)
	if err != nil {
		return fmt.Errorf("failed to scan monster data: %w", err)
	}

	dt, err := core.MakeDiceType(hitDie)
	if err != nil {
		return fmt.Errorf("failed to make dice type: %w", err)
	}
	cfg.HPConfig.HitDice = map[core.DiceType]int{
		dt: cfg.HPConfig.NumberOfDice,
	}

	mt := core.MakeMonsterType(monsterType)
	ms := core.MakeMonsterSize(monsterSize)

	if cfg.Metadata.IsLegendary {
		cfg.Metadata.MaxLegendaryActions = 3
	}
	cfg.Metadata.MonsterType = mt
	cfg.Metadata.MonsterSize = ms
	cfg.ActorType = core.ActorTypeMonster
	cfg.Abilities.Proficiencies.Strength = strSave.Valid && strSave.Int32 > 0
	cfg.Abilities.Proficiencies.Dexterity = dexSave.Valid && dexSave.Int32 > 0
	cfg.Abilities.Proficiencies.Constitution = conSave.Valid && conSave.Int32 > 0
	cfg.Abilities.Proficiencies.Intelligence = intSave.Valid && intSave.Int32 > 0
	cfg.Abilities.Proficiencies.Wisdom = wisSave.Valid && wisSave.Int32 > 0
	cfg.Abilities.Proficiencies.Charisma = chaSave.Valid && chaSave.Int32 > 0

	return nil
}

func hydrateActionDataSRD(ctx context.Context, cfg *actor.ActorConfig, id int) error {
	// Query actions
	stmt := SELECT(
		MonsterActions.ActionID,
		MonsterActions.Name,
		MonsterActions.RechargeValue,
		MonsterActions.HasDc,
		MonsterActions.Index,
		MonsterActions.IsAoe,
		MonsterDcDamageBlocks.Ability,
		MonsterDcDamageBlocks.OnSuccess,
		MonsterDcDamageBlocks.DcValue,
	).FROM(
		MonsterActions.
			LEFT_JOIN(MonsterDcDamageBlocks, MonsterActions.ActionID.EQ(MonsterDcDamageBlocks.ActionID)),
	).WHERE(
		MonsterActions.MonsterID.EQ(Int(int64(id))),
	).ORDER_BY(
		MonsterActions.ActionID.ASC(), MonsterActions.Index.ASC(),
	)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query actions for monster %d: %w", id, err)
	}
	defer rows.Close()

	for rows.Next() {
		var action core.Action
		var index int
		var dcAbility, dcOnSuccess sql.NullString
		var dcValue sql.NullInt64
		var isAoe sql.NullBool

		err = rows.Scan(
			&action.ID,
			&action.Name,
			&action.RechargeValue,
			&action.HasDC,
			&index,
			&isAoe,
			&dcAbility,
			&dcOnSuccess,
			&dcValue,
		)
		if err != nil {
			return fmt.Errorf("failed to scan monster action: %w", err)
		}

		if dcAbility.Valid {
			action.DCAbility = core.MakeAbility(dcAbility.String)
		}
		if dcOnSuccess.Valid {
			action.DCOnSuccess = core.DCOnSuccess(dcOnSuccess.String)
		}
		if dcValue.Valid {
			action.DCSaveDC = int(dcValue.Int64)
		}
		if isAoe.Valid {
			action.IsAOE = isAoe.Bool
		}

		// Action Damage blocks
		action.DiceBlock = make([]core.DiceBlock, 0)

		// Attack Bonus Blocks
		stmtDmgBlocks := SELECT(
			MonsterAttackBonusBlocks.NumberOfDice,
			MonsterAttackBonusBlocks.Die,
			MonsterAttackBonusBlocks.AmountToAdd,
			MonsterAttackBonusBlocks.AttackBonus,
			MonsterAttackBonusBlocks.DmgType,
		).FROM(MonsterAttackBonusBlocks).
			WHERE(MonsterAttackBonusBlocks.ActionID.EQ(Int(int64(action.ID.Int()))))

		queryDmgBlocks, argsDmgBlocks := stmtDmgBlocks.Sql()
		err = func() error {
			rows, err := database.Query(ctx, queryDmgBlocks, argsDmgBlocks...)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var dmgBlock core.DiceBlock
				err = rows.Scan(
					&dmgBlock.NumberOfDice,
					&dmgBlock.Die,
					&dmgBlock.Modifier,
					&action.AttackBonus,
					&dmgBlock.DamageType,
				)
				if err != nil {
					return err
				}
				action.DiceBlock = append(action.DiceBlock, dmgBlock)
			}
			return nil
		}()
		if err != nil {
			return fmt.Errorf("failed to query monster damage blocks for action %s: %w", action.ID, err)
		}

		// DC Damage Blocks
		stmtDcDmgBlocks := SELECT(
			MonsterDcDamageBlocks.NumberOfDice,
			MonsterDcDamageBlocks.Die,
			MonsterDcDamageBlocks.AmountToAdd,
			MonsterDcDamageBlocks.DmgType,
		).FROM(MonsterDcDamageBlocks).
			WHERE(MonsterDcDamageBlocks.ActionID.EQ(Int(int64(action.ID.Int()))))

		queryDcDmgBlocks, argsDcDmgBlocks := stmtDcDmgBlocks.Sql()
		err = func() error {
			rows, err := database.Query(ctx, queryDcDmgBlocks, argsDcDmgBlocks...)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var dmgBlock core.DiceBlock
				err = rows.Scan(
					&dmgBlock.NumberOfDice,
					&dmgBlock.Die,
					&dmgBlock.Modifier,
					&dmgBlock.DamageType,
				)
				if err != nil {
					return err
				}
				action.DiceBlock = append(action.DiceBlock, dmgBlock)
			}
			return nil
		}()
		if err != nil {
			return fmt.Errorf("failed to query monster DC damage blocks for action %s: %w", action.ID, err)
		}

		// Calculate Average Damage for AI
		for _, db := range action.DiceBlock {
			avg, _ := core.GetAverageRoll(db.NumberOfDice, db.Die, db.Modifier)
			action.AverageDamage += avg
		}

		// Set default type and cost, might be overridden by multiattack or legendary hydration
		if index == -1 {
			action.ActionType = core.ATLegendary
		} else {
			action.ActionType = core.ATAction
			action.Cost = core.ActionCost{
				ActivationType: core.ActAction,
				Value:          1,
			}
		}

		cfg.Actions = append(cfg.Actions, action)
	}

	return nil
}

func hydrateMultiattackDataSRD(ctx context.Context, cfg *actor.ActorConfig, id int) error {
	stmt := SELECT(
		MonsterMultiattacks.ActionID,
		MonsterMultiattacks.AttackCount,
		MonsterMultiattacks.OptionIndex,
	).FROM(MonsterMultiattacks).
		WHERE(MonsterMultiattacks.MonsterID.EQ(Int(int64(id)))).
		ORDER_BY(MonsterMultiattacks.OptionIndex.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query multiattacks for monster %d: %w", id, err)
	}
	defer rows.Close()

	options := make(map[int][]core.Multiattack)
	var indices []int

	for rows.Next() {
		var ma core.Multiattack
		var index int
		err = rows.Scan(&ma.ActionID, &ma.Count, &index)
		if err != nil {
			return fmt.Errorf("failed to scan multiattack: %w", err)
		}

		if _, ok := options[index]; !ok {
			indices = append(indices, index)
		}
		options[index] = append(options[index], ma)
	}

	for _, index := range indices {
		maAction := core.Action{
			Name:       "Multiattack",
			ActionType: core.ATMultiAttack,
			Cost: core.ActionCost{
				ActivationType: core.ActAction,
				Value:          1,
			},
			Multiattack: options[index],
		}
		cfg.Actions = append(cfg.Actions, maAction)
	}

	return nil
}

func hydrateLegendaryActionDataSRD(ctx context.Context, cfg *actor.ActorConfig, id int) error {
	stmt := SELECT(
		MonsterActionsLegendary.ActionID,
		MonsterActionsLegendary.Name,
		MonsterActionsLegendary.ActionCost,
	).FROM(MonsterActionsLegendary).
		WHERE(MonsterActionsLegendary.MonsterID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query legendary actions for monster %d: %w", id, err)
	}
	defer rows.Close()

	for rows.Next() {
		var actionID int
		var cost sql.NullInt64
		var name sql.NullString
		err = rows.Scan(&actionID, &name, &cost)
		if err != nil {
			return fmt.Errorf("failed to scan legendary action: %w", err)
		}

		// Try to find the action in the already hydrated actions
		found := false
		for i := range cfg.Actions {
			if cfg.Actions[i].Name == name.String {
				cfg.Actions[i].ActionType = core.ATLegendary
				cfg.Actions[i].Cost = core.ActionCost{
					ActivationType: core.ActLegendary,
					Value:          int(cost.Int64),
				}
				found = true
				break
			}
		}

		// If not found, add it as a skeleton (unlikely for SRD, but safe)
		if !found {
			legAction := core.Action{
				ID:         core.MakeID(actionID),
				Name:       name.String,
				ActionType: core.ATLegendary,
				Cost: core.ActionCost{
					ActivationType: core.ActLegendary,
					Value:          int(cost.Int64),
				},
			}
			cfg.Actions = append(cfg.Actions, legAction)
		}
	}

	return nil
}

func hydrateResistancesSRD(ctx context.Context, cfg *actor.ActorConfig, id int) error {
	stmt := SELECT(
		MonsterResistances.ResistanceType,
		MonsterResistances.DamageType,
	).FROM(MonsterResistances).
		WHERE(MonsterResistances.MonsterID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query resistances for monster %d: %w", id, err)
	}
	defer rows.Close()

	cfg.Resistances = core.NewDamageResistances()

	for rows.Next() {
		var rType, dType sql.NullString
		err = rows.Scan(&rType, &dType)
		if err != nil {
			return fmt.Errorf("failed to scan resistance: %w", err)
		}

		if !rType.Valid || !dType.Valid {
			return fmt.Errorf("invalid resistance: %s, %s", rType.String, dType.String)
		}

		resType, err := core.MakeResistanceType(strings.ToLower(strings.TrimSpace(rType.String)))
		if err != nil {
			return fmt.Errorf("invalid resistance type %s: %w", rType.String, err)
		}

		dmgType, err := core.MakeDamageType(strings.ToLower(strings.TrimSpace(dType.String)))
		if err != nil {
			return fmt.Errorf("invalid damage type %s: %w", dType.String, err)
		}

		cfg.Resistances.SetResistanceType(dmgType, resType)
	}

	// Also check for breakers
	stmtBreakers := SELECT(
		MonsterResistances.DamageType,
		MonsterResistances.ResistanceType,
		ResistBreakers.ResistBreakerType,
	).FROM(
		MonsterResistances.
			INNER_JOIN(MonsterResistBreakers, MonsterResistances.ResistanceID.EQ(MonsterResistBreakers.ResistanceID)).
			INNER_JOIN(ResistBreakers, MonsterResistBreakers.ResistBreakerID.EQ(ResistBreakers.ResistBreakerID)),
	).WHERE(MonsterResistances.MonsterID.EQ(Int(int64(id))))

	queryB, argsB := stmtBreakers.Sql()
	rowsB, err := database.Query(ctx, queryB, argsB...)
	if err != nil {
		return fmt.Errorf("failed to query resist breakers for monster %d: %w", id, err)
	}
	defer rowsB.Close()

	for rowsB.Next() {
		var dmgTypeStr, resTypeStr, breakerStr string
		err = rowsB.Scan(&dmgTypeStr, &resTypeStr, &breakerStr)
		if err != nil {
			return fmt.Errorf("failed to scan resist breaker: %w", err)
		}

		dmgType, _ := core.MakeDamageType(dmgTypeStr)
		resType, _ := core.MakeResistanceType(resTypeStr)
		breaker, _ := core.MakeResistBreaker(breakerStr)

		cfg.Resistances.SetResistanceType(dmgType, resType)
		cfg.Resistances.AddBreaker(dmgType, breaker)
	}

	return nil
}

func hydrateSpellcastingDataSRD(ctx context.Context, cfg *actor.ActorConfig, id int) error {
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
	).WHERE(MonsterSpellcasting.MonsterID.EQ(Int(int64(id)))).
		GROUP_BY(MonsterSpellcasting.MonsterID, MonsterSpellcasting.CastingLevel, MonsterSpellcasting.Ability, MonsterSpellcasting.AttackModifier, MonsterSpellcasting.SaveDc)

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query monster spellcasting by id: %w", err)
	}

	var monsterID, castingLevel int
	var am, dc sql.NullInt64
	var a string
	var standardSpellIDs []int
	var innateSpellsJSON, spellSlotsJSON string

	err = row.Scan(&monsterID, &castingLevel, &a, &am, &dc, &standardSpellIDs, &innateSpellsJSON, &spellSlotsJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // No spellcasting for this monster
		}
		return nil // Assume no spellcasting if no row found
	}

	// Parse innate spells JSON
	innateSpellsMap := make(map[int]int)
	if innateSpellsJSON != "{}" && innateSpellsJSON != "" {
		err = json.Unmarshal([]byte(innateSpellsJSON), &innateSpellsMap)
		if err != nil {
			return fmt.Errorf("failed to unmarshal innate spells: %w", err)
		}
	}

	// Parse spell slots JSON
	spellSlots := make(map[int]int)
	if spellSlotsJSON != "{}" && spellSlotsJSON != "" {
		err = json.Unmarshal([]byte(spellSlotsJSON), &spellSlots)
		if err != nil {
			return fmt.Errorf("failed to unmarshal spell slots: %w", err)
		}
	}

	// Collect all spell IDs
	allSpellIDs := make(map[int]struct{})
	for _, spellID := range standardSpellIDs {
		allSpellIDs[spellID] = struct{}{}
	}
	for spellID := range innateSpellsMap {
		allSpellIDs[spellID] = struct{}{}
	}

	// Get all spells in one query
	spellIDSlice := make([]int, 0, len(allSpellIDs))
	for spellID := range allSpellIDs {
		spellIDSlice = append(spellIDSlice, spellID)
	}

	var allSpells map[core.ID]spells.Spell
	if len(spellIDSlice) > 0 {
		params := spells.SpellQueryParams{ID: spellIDSlice}
		allSpells, err = QuerySpellData(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to query spells by id: %w", err)
		}
	} else {
		allSpells = make(map[core.ID]spells.Spell)
	}

	// Build final config
	var leveledSpells []spells.Spell
	var innateSpells []spells.InnateSpell

	// Standard spells
	for _, spellID := range standardSpellIDs {
		if spell, exists := allSpells[core.MakeID(spellID)]; exists {
			leveledSpells = append(leveledSpells, spell)
		}
	}

	// Innate spells with usage
	innateSpellIDs := make([]int, 0, len(innateSpellsMap))
	for spellID := range innateSpellsMap {
		innateSpellIDs = append(innateSpellIDs, spellID)
	}
	sort.Ints(innateSpellIDs)

	for _, spellID := range innateSpellIDs {
		timesPerDay := innateSpellsMap[spellID]
		if spell, exists := allSpells[core.MakeID(spellID)]; exists {
			innateSpells = append(innateSpells, spells.InnateSpell{
				Spell:          spell,
				MaxCastsPerDay: timesPerDay,
			})
		}
	}

	cfg.Spellcasting = &actor.MonsterSpellcastingConfig{
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
		LeveledSpells: leveledSpells,
		InnateSpells:  innateSpells,
		SpellSlots:    spellSlots,
	}

	// Update metadata
	cfg.Metadata.SpellcasterMetadata.SpellcastingAbility = cfg.Spellcasting.Ability
	cfg.Metadata.SpellcasterMetadata.SpellcastingLevel = cfg.Spellcasting.CastingLevel

	return nil
}
