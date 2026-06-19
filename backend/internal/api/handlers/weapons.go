package handlers

import (
	"dnd5e-encounter-simulator-backend/internal/database/repo"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetWeaponByID(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Weapon ID is required"})
		return
	}

	w, err := repo.HydrateWeaponData(c.Request.Context(), idParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get weapon data: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  w,
		"count": "1",
	})
}

func GetWeaponSummaries(c *gin.Context) {
	summaries, err := repo.GetAllWeaponData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get weapon summaries: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  summaries,
		"count": len(summaries),
	})
}
