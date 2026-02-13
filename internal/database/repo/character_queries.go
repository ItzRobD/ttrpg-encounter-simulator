package repo

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"encoding/json"

	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/model"
	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

type CharSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RaceID   int    `json:"race_id"`
	ClassID  int    `json:"class_id"`
	Level    int    `json:"level"`
	IsCustom bool   `json:"is_custom"`
}

func InsertCustomCharacterConfig(ctx context.Context, cfg actor.ActorConfig) error {
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
		RaceID:  int32(cfg.Metadata.RaceID),
		ClassID: int32(cfg.Metadata.ClassID),
		Level:   int32(cfg.Metadata.Level),
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

func GetCustomCharacterByID(ctx context.Context, id string) (*actor.ActorConfig, error) {
	cID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	var cfg actor.ActorConfig

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

func GetCustomCharacterSummaries(ctx context.Context) ([]CharSummary, error) {
	summaries := make([]CharSummary, 0)

	stmt := table.CharacterSummary.SELECT(table.CharacterSummary.AllColumns)

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var summary CharSummary
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
