package main

import (
	"dnd5e-encounter-simulator-backend/internal/api"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/simulation"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func main() {
	go func() {
		log.Println("Starting pprof server on :6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("Failed to start pprof server: %v", err)
		}
	}()

	err := database.InitDb(nil)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDb()

	log.Println("Server starting...")

	// 2. Initialize Worker Pool
	simulation.InitWorkerPool(4)
	defer simulation.ShutdownWorkerPool()

	r := gin.Default()
	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	api.RegisterRoutes(r)

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server_old: %v", err)
	}
}
