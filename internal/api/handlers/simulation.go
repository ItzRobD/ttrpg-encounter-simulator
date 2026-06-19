package handlers

import (
	"dnd5e-encounter-simulator-backend/pkg/simulation"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Simulation input bounds. These are deliberately constants (not magic numbers
// scattered through the handler) so tier-based limits can raise them later
// without hunting for the checks. See plan Phase 1.2 / Phase 2 (tiers).
const (
	maxNumberOfRuns    = 1000
	maxRoundsPerEnc    = 50
	maxEncounters      = 20
	maxTotalCombatants = 30
)

type SimulationPayload struct {
	Payload simulation.MultiSimulationRequest `json:"payload"`
}

func CreateSimulation(c *gin.Context) {
	var request SimulationPayload

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		slog.Warn("CreateSimulation: ShouldBindJSON failed", "err", err)
		return
	}

	req := request.Payload

	// Bounds validation. Upper caps protect the worker pool / DB from a
	// pathological payload; the constants can be relaxed per subscription tier.
	if req.NumberOfRuns <= 0 || req.NumberOfRuns > maxNumberOfRuns {
		c.JSON(http.StatusBadRequest, gin.H{"error": "number_of_runs must be between 1 and 1000"})
		return
	}
	if req.MaxRounds <= 0 || req.MaxRounds > maxRoundsPerEnc {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max_rounds must be between 1 and 50"})
		return
	}
	if len(req.Encounters) < 1 || len(req.Encounters) > maxEncounters {
		c.JSON(http.StatusBadRequest, gin.H{"error": "encounters must be between 1 and 20"})
		return
	}

	// Cap total combatants across the whole request. The most crowded single
	// encounter (characters + that encounter's monsters) drives the per-round
	// cost, so bound that rather than the sum of all encounters.
	for _, enc := range req.Encounters {
		combatants := len(req.CharacterConfigs) + len(enc.MonsterIDs) + len(enc.MonsterConfigs)
		if combatants > maxTotalCombatants {
			c.JSON(http.StatusBadRequest, gin.H{"error": "an encounter exceeds the maximum of 30 combatants"})
			return
		}
	}

	// For now, using a placeholder owner ID. In a real app, this would come from auth context.
	ownerID := "018f20a6-1234-7000-8000-000000000000"

	simID, err := simulation.InsertNewSimulation(c.Request.Context(), ownerID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create simulation: " + err.Error()})
		slog.Error("CreateSimulation: failed to create simulation", "err", err)
		return
	}

	// Enqueue the simulation for background processing
	simulation.EnqueueSimulation(simID, req)

	c.JSON(http.StatusAccepted, gin.H{
		"data": gin.H{
			"simulation_id": simID,
			"status":        "pending",
		},
	})
}

func GetSimulationStatusByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Simulation ID is required"})
		return
	}

	sim, err := simulation.GetSimulationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Simulation not found"})
		return
	}

	response := gin.H{
		"simulation_id": sim.ID.String(),
		"status":        sim.Status,
		"created_at":    sim.CreatedAt,
		"updated_at":    sim.UpdatedAt,
	}

	if sim.ErrorMessage != nil {
		response["error"] = *sim.ErrorMessage
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

func GetSimulationResultsByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Simulation ID is required"})
		return
	}

	sim, err := simulation.GetSimulationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Simulation not found"})
		return
	}

	response := gin.H{
		"simulation_id": sim.ID.String(),
		"status":        sim.Status,
		"created_at":    sim.CreatedAt,
		"updated_at":    sim.UpdatedAt,
	}

	if sim.ErrorMessage != nil {
		response["error"] = *sim.ErrorMessage
	}

	if sim.ResultData != nil {
		var result simulation.MultiSimulationResult
		if err := json.Unmarshal([]byte(*sim.ResultData), &result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal result data"})
			slog.Error("GetSimulationResultsByID: failed to unmarshal result data", "sim_id", id, "err", err)
			return
		}
		response["results"] = result
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}
