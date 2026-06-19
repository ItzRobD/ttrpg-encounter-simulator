package simulation

import (
	"context"
	"dnd5e-encounter-simulator-backend/internal/database"
	"encoding/json"
	"fmt"
	"time"

	"dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/model"
	. "dnd5e-encounter-simulator-backend/.gen/5e-encounter-simulator/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

func InsertNewSimulation(ctx context.Context, ownerID string, payload MultiSimulationRequest) (string, error) {
	simID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate uuid: %w", err)
	}

	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return "", fmt.Errorf("failed to parse owner id: %w", err)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	payloadStr := string(payloadBytes)

	// TODO: ValidatePayload()

	cfgs := NewSimulationEntityConfigs()
	cfgs.MonsterConfigs = payload.MonsterConfigs
	cfgs.CharacterConfigs = payload.CharacterConfigs
	cfgs.MonsterIDs = payload.MonsterIDs
	//cfgs.LairConfig = *payload.LairConfig // TODO: Add LairConfig

	cfgBytes, err := json.Marshal(cfgs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal simulation entity configs: %w", err)
	}
	cfgStr := string(cfgBytes)

	now := time.Now()
	status := model.Status_New
	m := model.Simulations{
		ID:             simID,
		OwnerID:        ownerUUID,
		Status:         &status,
		RequestPayload: &payloadStr,
		EntityConfigs:  &cfgStr,
		ResultData:     nil,
		ErrorMessage:   nil,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	stmt := Simulations.INSERT(Simulations.AllColumns).
		MODEL(m).
		ON_CONFLICT(Simulations.ID).
		DO_NOTHING()

	query, args := stmt.Sql()
	_, err = database.Exec(ctx, query, args...)
	if err != nil {
		return "", err
	}

	return simID.String(), nil
}

func UpdateSimulationStatus(ctx context.Context, id string, status model.Status, errorMsg *string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("failed to parse simulation id: %w", err)
	}

	now := time.Now()
	var errExpr postgres.Expression
	if errorMsg != nil {
		errExpr = postgres.String(*errorMsg)
	} else {
		errExpr = postgres.NULL
	}

	stmt := Simulations.UPDATE(
		Simulations.Status,
		Simulations.ErrorMessage,
		Simulations.UpdatedAt).
		SET(postgres.CAST(postgres.String(string(status))).AS("status"), errExpr, postgres.TimestampT(now)).
		WHERE(Simulations.ID.EQ(postgres.UUID(uid)))

	query, args := stmt.Sql()
	_, err = database.Exec(ctx, query, args...)
	return err
}

func UpdateSimulationResult(ctx context.Context, id string, result *MultiSimulationResult) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("failed to parse simulation id: %w", err)
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	resultStr := string(resultBytes)

	initialStates, err := json.Marshal(result.InitialState)
	if err != nil {
		return fmt.Errorf("failed to marshal initial state: %w", err)
	}
	iStateStr := string(initialStates)

	now := time.Now()
	status := model.Status_Completed
	stmt := Simulations.UPDATE(
		Simulations.Status,
		Simulations.ResultData,
		Simulations.InitialState,
		Simulations.UpdatedAt).
		SET(
			postgres.CAST(postgres.String(string(status))).AS("status"),
			postgres.CAST(postgres.String(resultStr)).AS("jsonb"),
			postgres.CAST(postgres.String(iStateStr)).AS("jsonb"),
			postgres.TimestampT(now),
		).
		WHERE(Simulations.ID.EQ(postgres.UUID(uid)))

	query, args := stmt.Sql()
	_, err = database.Exec(ctx, query, args...)
	return err
}

func GetSimulationByID(ctx context.Context, id string) (*model.Simulations, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse simulation id: %w", err)
	}

	var sim model.Simulations

	stmt := Simulations.SELECT(Simulations.AllColumns).
		FROM(Simulations).
		WHERE(Simulations.ID.EQ(postgres.UUID(uid)))

	query, args := stmt.Sql()
	row, err := database.QueryRow(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	err = row.Scan(
		&sim.ID,
		&sim.Status,
		&sim.RequestPayload,
		&sim.ResultData,
		&sim.ErrorMessage,
		&sim.CreatedAt,
		&sim.UpdatedAt,
		&sim.OwnerID,
		&sim.EntityConfigs,
		&sim.InitialState,
	)
	if err != nil {
		return nil, err
	}

	return &sim, nil
}

func GetSimulationIDsByUserID(ctx context.Context, userID string) ([]string, error) {
	ids := make([]string, 0)

	stmt := Simulations.SELECT(Simulations.ID).
		FROM(Simulations).
		WHERE(Simulations.OwnerID.EQ(postgres.String(userID)))

	query, args := stmt.Sql()
	rows, err := database.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id.String())
	}

	return ids, nil
}
