package api

import (
	"dnd5e-encounter-simulator-backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	// Health check
	r.GET("/health", handlers.HandleHealth)

	api := r.Group("/api/v1")
	{
		api.GET("/monsters/:id", handlers.GetMonsterByID)
		api.GET("/monsters/summaries", handlers.GetMonsterSummaries)
		//
		//api.GET("/characters/:id", handlers.GetCharacterByID)
		//api.GET("/characters/summaries", handlers.GetCharacterSummaries)
		//api.POST("/characters/save", handlers.SaveCharacter)

		api.GET("/equipment/weapons/:id", handlers.GetWeaponByID)
		api.GET("/equipment/weapons", handlers.GetWeaponSummaries)

		api.GET("/equipment/armor/:id", handlers.GetArmorByID)
		api.GET("/equipment/armor", handlers.GetArmorData)

		api.GET("/spells/:id", handlers.GetSpellByID)
		api.GET("/spells/summaries", handlers.GetSpellSummaries)
		api.GET("/spells/summaries/class/:id", handlers.GetSpellsByClassIDSummaries)
		api.GET("/spells/class/:id", handlers.GetSpellsByClassID)

		api.GET("/users/limits", handlers.GetUserLimits)

		//api.POST("/simulation/create", handlers.CreateSimulation)
		//api.GET("/simulation/status/:id", handlers.GetSimulationStatusByID)
		//api.GET("/simulation/results/:id", handlers.GetSimulationResultsByID)
	}

}
