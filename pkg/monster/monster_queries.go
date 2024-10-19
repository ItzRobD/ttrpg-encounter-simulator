package monster

import (
	"context"
	"database/sql"
	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/enum"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/shared"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func getMonsterIDByName(ctx context.Context, name string) (int, error) {
	var id int
	stmt := SELECT(
		Monsters.ID,
	).FROM(
		Monsters,
	).WHERE(Monsters.Name.EQ(String(name)))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return id, fmt.Errorf("failed to query monster base data by name: %w", err)
	}
	err = row.Scan(&id)
	if err != nil {
		return id, fmt.Errorf("failed to scan monster base data by name: %w", err)
	}

	return id, nil
}

func getMonsterBaseDataByID(ctx context.Context, id int) (MonsterBase, error) {
	var monsterResult MonsterBase
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
		WHERE(Monsters.ID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return monsterResult, fmt.Errorf("failed to query monster base data by id: %w", err)
	}
	err = row.Scan(
		&monsterResult.ID,
		&monsterResult.Name,
		&monsterResult.Size,
		&monsterResult.Type,
		&monsterResult.AC,
		&monsterResult.ProficiencyBonus,
		&monsterResult.CR,
		&monsterResult.ApiURL,
		&monsterResult.IsLegendary,
		&monsterResult.IsSpellcaster,
		&monsterResult.IsInnateSpellcaster,
		&monsterResult.AbilityScores.Strength,
		&monsterResult.AbilityScores.Dexterity,
		&monsterResult.AbilityScores.Constitution,
		&monsterResult.AbilityScores.Intelligence,
		&monsterResult.AbilityScores.Wisdom,
		&monsterResult.AbilityScores.Charisma,
		&monsterResult.HP.HPAverage,
		&monsterResult.HP.NumberOfDice,
		&monsterResult.HP.Die,
		&monsterResult.HP.AmountToAdd,
		&monsterResult.SaveProficiencies.Strength,
		&monsterResult.SaveProficiencies.Dexterity,
		&monsterResult.SaveProficiencies.Constitution,
		&monsterResult.SaveProficiencies.Intelligence,
		&monsterResult.SaveProficiencies.Wisdom,
		&monsterResult.SaveProficiencies.Charisma,
	)
	if err != nil {
		return monsterResult, fmt.Errorf("failed to scan monster base data by id: %w", err)
	}

	return monsterResult, nil
}

func getMonsterDamageModifiersByID(ctx context.Context, id int) ([]MonsterDamageModifier, error) {
	var monsterDamageModifiers []MonsterDamageModifier
	stmt := SELECT(
		MonsterDamageModifiers.DamageType,
		MonsterDamageModifiers.ModifierType,
	).
		FROM(
			MonsterDamageModifiers,
		).WHERE(
		MonsterDamageModifiers.MonsterID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return monsterDamageModifiers, fmt.Errorf("failed to query monster damage modifiers by id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var monsterDamageModifier MonsterDamageModifier
		err = rows.Scan(&monsterDamageModifier.DamageType, &monsterDamageModifier.ModifierType)
		if err != nil {
			return monsterDamageModifiers, fmt.Errorf("failed to scan monster damage modifiers by id: %w", err)
		}
		monsterDamageModifiers = append(monsterDamageModifiers, monsterDamageModifier)
	}

	if err := rows.Err(); err != nil {
		return monsterDamageModifiers, fmt.Errorf("failed to query monster damage modifiers by id: %w", err)
	}

	return monsterDamageModifiers, nil
}

func getMonsterResistBreakersByID(ctx context.Context, id int) ([]shared.DamageBreaker, error) {
	var monsterResistBreakers []shared.DamageBreaker
	stmt := SELECT(
		MonsterDamageModifiers.DamageType,
		MonsterResistBreakers.ResistBreakerType,
	).FROM(
		Monsters.
			INNER_JOIN(MonsterDamageModifiers, Monsters.ID.EQ(MonsterDamageModifiers.MonsterID)).
			LEFT_JOIN(MonsterDamageResistBreakers, MonsterDamageResistBreakers.ModifierID.EQ(MonsterDamageModifiers.ModifierID)).
			LEFT_JOIN(MonsterResistBreakers, MonsterResistBreakers.ResistBreakerID.EQ(MonsterDamageResistBreakers.ResistBreakerID))).
		WHERE(Monsters.ID.EQ(Int(int64(id))))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return monsterResistBreakers, fmt.Errorf("failed to query monster resist breakers by id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var damageBreaker shared.DamageBreaker
		var damageBreakerName pgtype.Text
		err = rows.Scan(&damageBreaker.DamageType, &damageBreakerName)
		if err != nil {
			return monsterResistBreakers, fmt.Errorf("failed to scan monster resist breakers by id: %w", err)
		}
		if damageBreakerName.Valid {
			damageBreaker.Breaker = shared.WeaponBreakerType(damageBreakerName.String)
			monsterResistBreakers = append(monsterResistBreakers, damageBreaker)
		}
	}

	if err := rows.Err(); err != nil {
		return monsterResistBreakers, fmt.Errorf("failed to query monster resist breakers by id: %w", err)
	}

	return monsterResistBreakers, nil
}

func getMonsterActionsByID(ctx context.Context, id int) ([]MonsterAction, error) {
	var monsterActions []MonsterAction
	stmt := SELECT(
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
		MonsterActions.MonsterID.EQ(Int(int64(id))),
	).ORDER_BY(
		MonsterActions.ActionID.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return monsterActions, fmt.Errorf("failed to query monster actions by id: %w", err)
	}
	defer rows.Close()
	monsterActions, err = pgx.CollectRows(rows, pgx.RowToStructByPos[MonsterAction])
	if err != nil {
		return monsterActions, fmt.Errorf("failed to scan monster actions by id: %w", err)
	}

	return monsterActions, nil
}

func getMonsterMultiattacksByID(ctx context.Context, id int) ([]MonsterMultiattack, error) {
	var monsterMultiattacks []MonsterMultiattack
	stmt := SELECT(
		MonsterMultiattacks.ActionID,
		MonsterMultiattacks.AttackCount,
		MonsterMultiattacks.IsOption,
		MonsterMultiattacks.OptionIndex,
	).FROM(
		MonsterMultiattacks,
	).WHERE(
		MonsterMultiattacks.MonsterID.EQ(Int(int64(id))),
	).ORDER_BY(MonsterMultiattacks.ActionID.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return monsterMultiattacks, fmt.Errorf("failed to query monster multiattacks by id: %w", err)
	}
	defer rows.Close()
	monsterMultiattacks, err = pgx.CollectRows(rows, pgx.RowToStructByPos[MonsterMultiattack])
	if err != nil {
		return monsterMultiattacks, fmt.Errorf("failed to query monster multiattacks by id collect rows: %w", err)
	}

	return monsterMultiattacks, nil
}

func getMonsterLegendaryActionsByID(ctx context.Context, id int) ([]LegendaryAction, error) {
	var monsterLegendaryActions []LegendaryAction
	stmt := SELECT(
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
		MonsterActionsLegendary.MonsterID.EQ(Int(int64(id))),
	).ORDER_BY(
		MonsterActionsLegendary.ActionID.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return monsterLegendaryActions, fmt.Errorf("failed to query monster legendary actions by id: %w", err)
	}
	defer rows.Close()
	monsterLegendaryActions, err = pgx.CollectRows(rows, pgx.RowToStructByPos[LegendaryAction])
	if err != nil {
		return monsterLegendaryActions, fmt.Errorf("failed to assign legendary actions by id: %w", err)
	}

	return monsterLegendaryActions, nil
}

func getMonsterSpecialAbilities(ctx context.Context, id int) ([]SpecialAbility, error) {
	var specialAbilities []SpecialAbility
	stmt := SELECT(
		MonsterSpecialAbilities.Name,
		MonsterSpecialAbilities.UsageCount,
		MonsterSpecialAbilities.Description,
	).FROM(
		MonsterSpecialAbilities,
	).WHERE(
		MonsterSpecialAbilities.MonsterID.EQ(Int(int64(id))),
	).ORDER_BY(
		MonsterSpecialAbilities.Name.ASC())

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return specialAbilities, fmt.Errorf("failed to query monster special abilities by id: %w")
	}
	defer rows.Close()
	for rows.Next() {
		var sa SpecialAbility
		var name string
		var usageCount sql.NullInt64
		var description string
		err = rows.Scan(&name, &usageCount, &description)
		if err != nil {
			return specialAbilities, fmt.Errorf("failed to scan monster special abilities by id: %w", err)
		}
		if usageCount.Valid {
			sa.UsageCount = int(usageCount.Int64)
		} else {
			sa.UsageCount = 0
		}
		sa.Name = name
		sa.Description = description

		specialAbilities = append(specialAbilities, sa)
	}

	return specialAbilities, nil
}

func getMonsterSpellcastingSlotsByID(ctx context.Context, id int) (map[int]int, error) {
	var spellcastingSlots map[int]int
	stmt := SELECT(
		MonsterSpellcastingSlots.SpellLevel,
		MonsterSpellcastingSlots.Slots,
	).FROM(
		MonsterSpellcastingSlots,
	).WHERE(
		MonsterSpellcastingSlots.MonsterID.EQ(Int(int64(id))),
	)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	for rows.Next() {
		var spellLevel int
		var slots int
		err = rows.Scan(&spellLevel, &slots)
		if err != nil {
			return spellcastingSlots, fmt.Errorf("failed to scan monster spellcasting slots by id: %w", err)
		}
		if spellcastingSlots == nil {
			spellcastingSlots = make(map[int]int)
		}
		spellcastingSlots[spellLevel] = slots
	}

	return spellcastingSlots, nil
}

func getMonsterSpellsByID(ctx context.Context, id int) ([]int, error) {
	var spellIDs []int
	stmt := SELECT(
		MonsterAvailableSpells.SpellID,
	).FROM(
		MonsterAvailableSpells.
			LEFT_JOIN(Spells, MonsterAvailableSpells.SpellID.EQ(Spells.ID)),
	).WHERE(
		MonsterAvailableSpells.MonsterID.EQ(Int(int64(id))).
			AND(
				Spells.SpellType.EQ(enum.Stype.Damage).
					OR(Spells.SpellType.EQ(enum.Stype.Healing)),
			),
	)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return spellIDs, fmt.Errorf("failed to query monster spell ids by monster id: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var spellID int
		err = rows.Scan(&spellID)
		if err != nil {
			return spellIDs, fmt.Errorf("failed to scan monster spell ids by monster id: %w", err)
		}
		spellIDs = append(spellIDs, spellID)
	}

	return spellIDs, nil
}

func getMonsterSpellcastingByID(ctx context.Context, id int) (MSpellcasting, error) {
	var spellcasting MSpellcasting

	var isInnateCaster bool

	stmt := SELECT(Monsters.IsInnateCaster).FROM(Monsters).WHERE(Monsters.ID.EQ(Int(int64(id))))
	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return spellcasting, fmt.Errorf("failed to query is innate caster by id: %w", err)
	}
	err = row.Scan(&isInnateCaster)
	if err != nil {
		return spellcasting, fmt.Errorf("failed to scan is innate caster by id: %w", err)
	}

	stmt = SELECT(
		MonsterSpellcasting.CastingLevel,
		MonsterSpellcasting.Ability,
		MonsterSpellcasting.AttackModifier,
		MonsterSpellcasting.SaveDc,
	).FROM(
		MonsterSpellcasting,
	).WHERE(
		MonsterSpellcasting.MonsterID.EQ(Int(int64(id))),
	)

	query, args = stmt.Sql()
	row, err = database.QueryRow(ctx, query, args...)
	if err != nil {
		return spellcasting, fmt.Errorf("failed to query monster spellcasting by id: %w", err)
	}
	err = row.Scan(&spellcasting.CastingLevel, &spellcasting.Ability, &spellcasting.AttackModifier, &spellcasting.SaveDC)
	if err != nil {
		return spellcasting, fmt.Errorf("failed to scan monster spellcasting by id: %w", err)
	}

	if isInnateCaster {
		stmt = SELECT(
			MonsterAvailableSpellsInnate.SpellID,
			MonsterAvailableSpellsInnate.TimesPerDay,
		).FROM(
			MonsterAvailableSpellsInnate.
				LEFT_JOIN(Spells, MonsterAvailableSpellsInnate.SpellID.EQ(Spells.ID)),
		).WHERE(
			MonsterAvailableSpellsInnate.MonsterID.EQ(Int(int64(id))).
				AND(
					Spells.SpellType.EQ(enum.Stype.Damage).
						OR(Spells.SpellType.EQ(enum.Stype.Healing)),
				),
		)

		query, args = stmt.Sql()
		rows, err2 := database.Query(ctx, query, args...)
		if err2 != nil {
			return spellcasting, fmt.Errorf("failed to query monster innate spells by id: %w", err2)
		}
		defer rows.Close()
		for rows.Next() {
			var spellID int
			var timesPerDay int
			err2 = rows.Scan(&spellID, &timesPerDay)
			if err2 != nil {
				return spellcasting, fmt.Errorf("failed to scan monster innate spell ids: %w", err2)
			}
			var s spells.Spell
			sQueryParams := spells.SpellQueryParams{ID: spellID, Level: 1}
			s, err2 = spells.QuerySpellData(ctx, sQueryParams)
			if err2 != nil {
				return spellcasting, fmt.Errorf("failed to query spell data by id: %w", err2)
			}
			var iSpell InnateSpell
			iSpell.Spell = s
			iSpell.TimePerDay = timesPerDay
			spellcasting.InnateSpells = append(spellcasting.InnateSpells, iSpell)
		}
	} else {
		// Get all spells
		spellIDs, err2 := getMonsterSpellsByID(ctx, id)
		for _, spellID := range spellIDs {
			var s spells.Spell
			sQueryParams := spells.SpellQueryParams{ID: spellID, Level: 0}
			s, err2 = spells.QuerySpellData(ctx, sQueryParams)
			if err2 != nil {
				return spellcasting, err2
			}
			spellcasting.SC.Spells = append(spellcasting.SC.Spells, s)
		}
		// Get spell slots
		scSlots, err2 := getMonsterSpellcastingSlotsByID(ctx, id)
		if err2 != nil {
			return MSpellcasting{}, err2
		}
		spellcasting.SC.SpellSlots = scSlots
		spellcasting.SC.MaxSpellSlots = scSlots

	}
	return spellcasting, nil
}

func QueryMonsterData(ctx context.Context, params MonsterQueryParams) (Monster, error) {
	var monsterResult Monster
	var monsterBaseResult MonsterBase
	var err error

	if params.ID != 0 {
		monsterBaseResult, err = getMonsterBaseDataByID(ctx, params.ID)
	} else if params.Name != "" {
		var id int
		id, err = getMonsterIDByName(ctx, params.Name)
		if err != nil {
			return monsterResult, err
		}
		monsterBaseResult, err = getMonsterBaseDataByID(ctx, id)
		if err != nil {
			return monsterResult, err
		}
	} else {
		err = fmt.Errorf("no name or id provided for monster data query")
		return monsterResult, err
	}

	monsterResult.MonsterBase = monsterBaseResult

	if monsterBaseResult.ID != 0 {
		var monsterDamageModifiers []MonsterDamageModifier
		monsterDamageModifiers, err = getMonsterDamageModifiersByID(ctx, monsterBaseResult.ID)
		if err != nil {
			return monsterResult, err
		}
		monsterResult.DamageModifiers = monsterDamageModifiers

		var monsterResistBreakers []shared.DamageBreaker
		monsterResistBreakers, err = getMonsterResistBreakersByID(ctx, monsterBaseResult.ID)
		if err != nil {
			return monsterResult, err
		}
		monsterResult.ResistBreakers = monsterResistBreakers

		var monsterActions []MonsterAction
		monsterActions, err = getMonsterActionsByID(ctx, monsterBaseResult.ID)
		if err != nil {
			return monsterResult, err
		}
		monsterResult.Actions = monsterActions

		var monsterMultiattacks []MonsterMultiattack
		monsterMultiattacks, err = getMonsterMultiattacksByID(ctx, monsterBaseResult.ID)
		if err != nil {
			return monsterResult, err
		}
		monsterResult.Multiattacks = monsterMultiattacks

		var mSpecialAbilities []SpecialAbility
		mSpecialAbilities, err = getMonsterSpecialAbilities(ctx, monsterBaseResult.ID)
		if err != nil {
			return monsterResult, err
		}
		monsterResult.SpecialAbilities = mSpecialAbilities

		if monsterResult.IsLegendary {
			var monsterLegendaryActions []LegendaryAction
			monsterLegendaryActions, err = getMonsterLegendaryActionsByID(ctx, monsterBaseResult.ID)
			if err != nil {
				return monsterResult, err
			}
			monsterResult.LegendaryActions = monsterLegendaryActions
		}

		if monsterResult.IsSpellcaster || monsterResult.IsInnateSpellcaster {
			var monsterSpellcasting MSpellcasting
			monsterSpellcasting, err = getMonsterSpellcastingByID(ctx, monsterBaseResult.ID)
			if err != nil {
				return monsterResult, err
			}
			monsterResult.Spellcasting = monsterSpellcasting
		}
	} else {
		err = fmt.Errorf("invalid monster id to query additional data")
		return monsterResult, err
	}

	return monsterResult, nil
}
