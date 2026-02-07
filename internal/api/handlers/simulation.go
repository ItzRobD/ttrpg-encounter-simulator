package handlers

//
//type SimulationPayload struct {
//	Payload simulation.MultiSimulationRequest `json:"payload"`
//}
//
//func CreateSimulation(c *gin.Context) {
//	var request SimulationPayload
//
//	if err := c.ShouldBindJSON(&request); err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
//		// Log the error for debugging
//		log.Printf("CreateSimulation: ShouldBindJSON failed: %v", err)
//		return
//	}
//
//	req := request.Payload
//
//	// Basic validation
//	if req.NumberOfRuns <= 0 {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid number of runs"})
//		log.Printf("CreateSimulation: Invalid number of runs: %d", req.NumberOfRuns)
//		return
//	}
//	if req.MaxRounds <= 0 {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max rounds"})
//		log.Printf("CreateSimulation: Invalid max rounds: %d", req.MaxRounds)
//		return
//	}
//	if len(req.CharacterConfigs) < 1 && len(req.MonsterIDs) < 1 && len(req.MonsterConfigs) < 1 {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid number of characters or monsters"})
//		log.Printf("CreateSimulation: Invalid number of characters or monsters: %d", len(req.CharacterConfigs)+len(req.MonsterIDs)+len(req.MonsterConfigs))
//		return
//	}
//
//	// For now, using a placeholder owner ID. In a real app, this would come from auth context.
//	ownerID := "018f20a6-1234-7000-8000-000000000000"
//
//	simID, err := simulation.InsertNewSimulation(c.Request.Context(), ownerID, req)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create simulation: " + err.Error()})
//		log.Printf("CreateSimulation: Failed to create simulation: %v", err)
//		return
//	}
//
//	// Enqueue the simulation for background processing
//	simulation.EnqueueSimulation(simID, req)
//
//	c.JSON(http.StatusAccepted, gin.H{
//		"data": gin.H{
//			"simulation_id": simID,
//			"status":        "pending",
//		},
//	})
//}
//
//func GetSimulationStatusByID(c *gin.Context) {
//	id := c.Param("id")
//	if id == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Simulation ID is required"})
//		return
//	}
//
//	sim, err := simulation.GetSimulationByID(c.Request.Context(), id)
//	if err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "Simulation not found"})
//		return
//	}
//
//	response := gin.H{
//		"simulation_id": sim.ID.String(),
//		"status":        sim.Status,
//		"created_at":    sim.CreatedAt,
//		"updated_at":    sim.UpdatedAt,
//	}
//
//	if sim.ErrorMessage != nil {
//		response["error"] = *sim.ErrorMessage
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"data": response,
//	})
//}
//
//func GetSimulationResultsByID(c *gin.Context) {
//	id := c.Param("id")
//	if id == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "Simulation ID is required"})
//		return
//	}
//
//	sim, err := simulation.GetSimulationByID(c.Request.Context(), id)
//	if err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "Simulation not found"})
//		return
//	}
//
//	response := gin.H{
//		"simulation_id": sim.ID.String(),
//		"status":        sim.Status,
//		"created_at":    sim.CreatedAt,
//		"updated_at":    sim.UpdatedAt,
//	}
//
//	if sim.ErrorMessage != nil {
//		response["error"] = *sim.ErrorMessage
//	}
//
//	if sim.EntityConfigs != nil {
//		var cfgs simulation.SimulationEntityConfigs
//		if err := json.Unmarshal([]byte(*sim.EntityConfigs), &cfgs); err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal entity configs"})
//			log.Printf("GetSimulationResultsByID: Failed to unmarshal entity configs for %s: %v", id, err)
//			return
//		}
//		response["entity_configs"] = cfgs
//	}
//
//	if sim.ResultData != nil {
//		var result simulation.MultiSimulationResult
//		if err := json.Unmarshal([]byte(*sim.ResultData), &result); err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal result data"})
//			log.Printf("GetSimulationResultsByID: Failed to unmarshal result data for %s: %v", id, err)
//			return
//		}
//		response["results"] = result
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"data": response,
//	})
//}
