package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUserLimits(c *gin.Context) {
	_ = c.GetHeader("Authorization")

	c.JSON(http.StatusOK, gin.H{
		"tier": "premium",
		"characters": gin.H{
			"current": 0,
			"max":     10,
		},
		"monsters": gin.H{
			"current": 0,
			"max":     10,
		},
		"spells": gin.H{
			"current": 0,
			"max":     10,
		},
		"equipment": gin.H{
			"current": 0,
			"max":     10,
		},
	})
}
