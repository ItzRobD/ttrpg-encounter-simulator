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

	now := time.Now()
	status := model.Status_New
	m := model.Simulations{
		ID:             simID,
		OwnerID:        ownerUUID,
		Status:         &status,
		RequestPayload: &payloadStr,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	stmt := Simulations.INSERT(
		Simulations.ID,
		Simulations.OwnerID,
		Simulations.Status,
		Simulations.RequestPayload,
		Simulations.CreatedAt,
		Simulations.UpdatedAt,
	).MODEL(m)

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

	now := time.Now()
	status := model.Status_Completed
	stmt := Simulations.UPDATE(
		Simulations.Status,
		Simulations.ResultData,
		Simulations.UpdatedAt).
		SET(
			postgres.CAST(postgres.String(string(status))).AS("status"),
			postgres.CAST(postgres.String(resultStr)).AS("jsonb"),
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
