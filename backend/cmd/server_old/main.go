package main

import (
	"dnd5e-encounter-simulator-backend/internal/api"
	"dnd5e-encounter-simulator-backend/internal/database"
	"dnd5e-encounter-simulator-backend/pkg/simulation"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func main() {
	// 0. Start pprof server_old
	go func() {
		log.Println("Starting pprof server_old on :6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Printf("Pprof server_old failed: %v", err)
		}
	}()

	// 1. Initialize Database
	err := database.InitDb(nil) // nil uses default .env loading
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDb()

	// 2. Initialize Worker Pool
	simulation.InitWorkerPool(4)
	defer simulation.ShutdownWorkerPool()

	// 3. Setup Gin Router
	r := gin.Default()
	r.Use(cors.Default())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	api.RegisterRoutes(r)

	// 4. Start Server
	log.Println("Starting server_old on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server_old: %v", err)
	}
}
