package character

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"encoding/json"

	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/model"
	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

func InsertCustomCharacterConfig(ctx context.Context, cfg CharacterConfig) error {
	id, err := uuid.Parse(cfg.ID)
	if err != nil {
		return err
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	cfgStr := string(cfgBytes)

	summary := model.CharacterSummary{
		ID:      id,
		Name:    cfg.Name,
		RaceID:  int32(cfg.RaceID),
		ClassID: int32(cfg.ClassID),
		Level:   int32(cfg.Level),
	}
	details := model.CharacterConfig{
		ID:     id,
		Config: cfgStr,
	}

	stmt := table.CharacterSummary.INSERT(table.CharacterSummary.AllColumns).
		MODEL(summary).ON_CONFLICT(table.CharacterSummary.ID).DO_NOTHING()

	query, args := stmt.Sql()
	_, err = database.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	stmt = table.CharacterConfig.INSERT(table.CharacterConfig.AllColumns).
		MODEL(details).ON_CONFLICT(table.CharacterConfig.ID).DO_NOTHING()

	query, args = stmt.Sql()
	_, err = database.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func GetCustomCharacterByID(ctx context.Context, id string) (*CharacterConfig, error) {
	cID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var cfg CharacterConfig

	stmt := table.CharacterConfig.SELECT(table.CharacterConfig.Config).
		WHERE(table.CharacterConfig.ID.EQ(postgres.UUID(cID)))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	err = row.Scan(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func GetCustomCharacterSummaries(ctx context.Context) ([]CharacterSummary, error) {
	summaries := make([]CharacterSummary, 0)

	stmt := table.CharacterSummary.SELECT(table.CharacterSummary.AllColumns)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var summary CharacterSummary
		err = rows.Scan(
			&summary.ID,
			&summary.Name,
			&summary.RaceID,
			&summary.ClassID,
			&summary.Level)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}
