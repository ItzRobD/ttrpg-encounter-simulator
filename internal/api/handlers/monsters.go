package handlers

import (
	"dnd5e-encounter-simulator-backend/internal/database/repo"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetMonsterByID(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Monster ID is required"})
		return
	}

	cfg, err := repo.HydrateMonsterConfig(c.Request.Context(), idParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to query monster data: %v", err),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  cfg,
		"count": "1",
	})
}

func GetMonsterSummaries(c *gin.Context) {
	summaries, err := repo.GetMonsterSummaries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get monster summaries: %v", err),
		})
		return
	}

	if len(summaries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No monsters found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  summaries,
		"count": len(summaries),
	})
}
