package handlers

import (
	"dnd5e-encounter-simulator-backend/internal/database/repo"
	"dnd5e-encounter-simulator-backend/pkg/actor"
	"dnd5e-encounter-simulator-backend/pkg/core"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetCharacterSummaries(c *gin.Context) {
	summaries, err := repo.GetCustomCharacterSummaries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get character summaries: " + err.Error()})
		log.Printf("GetCharacterSummaries: Failed to get character summaries: %v", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  summaries,
		"count": len(summaries),
	})
}

func GetCharacterByID(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "character ID is required"})
		return
	}

	char, err := repo.GetCustomCharacterByID(c.Request.Context(), idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		log.Printf("GetCharacterByID: Failed to get character by ID %s: %v", idStr, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  char,
		"count": "1",
	})
}

func SaveCharacter(c *gin.Context) {
	var char actor.ActorConfig

	if err := c.ShouldBindJSON(&char); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		log.Printf("SaveCharacter: ShouldBindJSON failed: %v", err)
		return
	}

	//Validate Character(char)

	// Generate UUID for the character
	char.ID = core.NewUUIDv7()

	// Insert to the Database
	err := repo.InsertCustomCharacterConfig(c.Request.Context(), char)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save character: " + err.Error()})
		log.Printf("SaveCharacter: InsertCustomCharacterConfig failed: %v", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"id": char.ID,
		},
		"count": "1",
	})
}
