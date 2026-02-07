package handlers

import (
	"dnd5e-encounter-simulator-backend/internal/database/repo"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetArmorByID(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Armor ID is required"})
		return
	}

	a, err := repo.HydrateArmorData(c.Request.Context(), idParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get armor data: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  a,
		"count": "1",
	})
	return
}

func GetArmorData(c *gin.Context) {
	armorData, err := repo.GetAllArmorData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get armor data: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  armorData,
		"count": len(armorData),
	})
}
