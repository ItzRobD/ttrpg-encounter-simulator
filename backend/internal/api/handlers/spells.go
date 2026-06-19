package handlers

import (
	"dnd5e-encounter-simulator-backend/internal/database/repo"
	"dnd5e-encounter-simulator-backend/pkg/spells"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetSpellByID(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Spell ID is required"})
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid s ID: %v", err),
		})
		return
	}

	params := spells.SpellQueryParams{ID: []int{id}}

	s, err := repo.QuerySpellData(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to query s data: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  s,
		"count": "1",
	})
	return
}

func GetSpellSummaries(c *gin.Context) {
	summaries, err := repo.GetSpellSummaries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get spell summaries: %v", err),
		})
		return
	}

	if len(summaries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No spells found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  summaries,
		"count": len(summaries),
	})
}

func GetSpellsByClassID(c *gin.Context) {
	classIDParam := c.Param("id")
	if classIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Class ID is required"})
		return
	}

	classID, err := strconv.Atoi(classIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid class ID: %v", err),
		})
		return
	}

	spellIDs, err := repo.GetUsableSpellIDsByClassID(c.Request.Context(), uint8(classID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get spell IDs by class ID: %v", err),
		})
		return
	}

	s, err := repo.QuerySpellData(c.Request.Context(), spells.SpellQueryParams{ID: spellIDs})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to query spell data: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  s,
		"count": len(s),
	})
}

func GetSpellsByClassIDSummaries(c *gin.Context) {
	classIDParam := c.Param("id")
	classID, err := strconv.Atoi(classIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid class ID: %v", err),
		})
		return
	}

	summaries, err := repo.GetSpellSummariesByClassID(c.Request.Context(), uint8(classID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get spell summaries by class ID: %v", err),
		})
		return
	}

	if len(summaries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No spells found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  summaries,
		"count": len(summaries),
	})
}
