package handlers

import (
	"dnd5e-encounter-simulator-backend/internal/database/repo"
	"dnd5e-encounter-simulator-backend/pkg/simulation"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SimulationPayload struct {
	Payload simulation.MultiSimulationRequest `json:"payload"`
}

func CreateSimulation(c *gin.Context) {
	var request SimulationPayload

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		log.Printf("CreateSimulation: ShouldBindJSON failed: %v", err)
		return
	}

	req := request.Payload

	// Hydrate monsters from monster_ids
	for _, mID := range req.MonsterIDs {
		monsterCfg, err := repo.HydrateMonsterConfig(c.Request.Context(), fmt.Sprintf("%d", mID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to hydrate monster %d: %v", mID, err)})
			log.Printf("CreateSimulation: Failed to hydrate monster %d: %v", mID, err)
			return
		}
		req.ActorConfigs = append(req.ActorConfigs, *monsterCfg)
	}

	// Basic validation
	if req.NumberOfRuns <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid number of runs"})
		return
	}
	if req.MaxRounds <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max rounds"})
		return
	}
	if len(req.ActorConfigs) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid number of actors"})
		return
	}

	// For now, using a placeholder owner ID. In a real app, this would come from auth context.
	ownerID := "018f20a6-1234-7000-8000-000000000000"

	simID, err := simulation.InsertNewSimulation(c.Request.Context(), ownerID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create simulation: " + err.Error()})
		log.Printf("CreateSimulation: Failed to create simulation: %v", err)
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
			log.Printf("GetSimulationResultsByID: Failed to unmarshal result data for %s: %v", id, err)
			return
		}
		response["results"] = result
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}
